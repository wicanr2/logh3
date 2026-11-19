# 451 — 計劃畫面的參謀鈕：預算、生產、建造三個入口

日期：2026-11-19。第173輪。

起因：使用者轉來玩家回報。菜單頁面與月結頁面裡，生產計劃的參謀和生產額度分配點擊無效；
徵稅計劃的參謀在報告頁面也換不了，在戰略頁面可以。

本輪**沒有跑原版**：dosgolem 專用副本 `~/cht/dosgolem-logh3/` 已不在本機（第170輪記錄），
使用者裁決先依推論實作。下面每一條都標等級。

## 一、三個入口各自的真因（proven，remake 程式碼）

| 入口 | 玩家看到的 | 真因 | 位置（修正前，main `0d6ad6aa`）|
|---|---|---|---|
| 月初報告 → 次月份預算計劃 → 參謀 | 點了沒反應，之後點哪裡都沒用 | 點擊確實開了參謀清單（`beginStaffPick`），但預算計劃在 `Frame` 走自己的分支，那一支沒有呼叫 `drawStaffPick`。清單是隱形的，後續點擊全被它吃掉 | `internal/game/game.go` `case StateBudgetPlan` |
| 戰略命令 → 生產計劃 → 參謀、七列箭頭 | 點了沒反應 | `planScreenClick` 在 `StateProductionPlan` 只認「結束」 | `internal/game/planscreens.go` |
| 月初 → 船艦訂造報告 → 建造計劃 → 參謀 | 清單開得出來，選不了人 | 參謀鈕開的 `StateAdmiralList` 只處理捲動，點列沒有動作 | `internal/game/game.go` `case StateAdmiralList` |

徵稅計劃（戰略命令那一條）走預設分支，有畫疊層，所以「戰略頁面可以換」。

建造計劃的七列箭頭本來就接著。各列合計 100% 時「加」不動是原版行為
（`rules.BuildPlan.Inc` 的註解、`docs/re/11`）：要先減出餘額才加得上去。

## 二、熱區來源（proven，位元組，沿用 `docs/re/343`）

- 預算計劃與徵稅計劃共用熱區表 `ds:14A8`（六筆，第 2 筆是參謀），繪圖與事件碼也共用。
  所以預算計劃那一顆換的與徵稅計劃相同：內政專任參謀 `ds:8F4F[座位]`。
- 生產計劃與建造計劃共用熱區表 `ds:26E2`（7 列 × 2 箭頭、參謀、結束）。remake 的
  `screen-build-plan.json` hotspots 就是這一份（滑鼠座標，參謀 `(448,320,32,20)`），
  生產計劃的命中測試直接借它。

## 三、生產與建造那一顆換哪一格（strong inference）

推論為**軍事專任參謀** `ds:8F4B[座位]`（`rules.PostMilitary`），依據：

1. 這兩張畫面的署名與頭像都是軍事專任參謀（`screen-build-plan`、`screen-production-plan`
   的 `staff_post`）。另外兩個已證實的入口都是「哪一張畫面的署名，那一顆就換哪一位」：
   情報收集換情報參謀，徵稅計劃換內政參謀（`docs/re/348` §三、§六）。
2. 船艦單價由軍事專任參謀的營運決定（`docs/spec/037`、`docs/re/265`），換這一位才會影響這兩張的數字。
3. 情報、內政兩格已經各有入口；三個幕僚職裡只剩軍事這一格沒有。

沒有收據的部分：

- 原版寫入的位元組（預期 `-peek` 的 `ds:8F4B` 那兩格之一改變，`8F4D`、`8F4F` 不動）。
- 換完之後單價與船艦數是否**當場**重算（unknown）。remake 取「立刻反映」，
  與內政那一顆換完當場重算稅收（`docs/re/348` §七）一致。
- 建造計劃清單點一列之後的畫面：remake 疊上同一張詳細資料（第三顆鈕「決定」），
  「中止」回清單、「決定」回建造計劃，比照另外兩條路（規格 064 規則 5）。

## 四、修正

- `planscreens.go`：生產計劃的七列箭頭動兩張共用的百分比 `g.pct`（`docs/spec/004`），
  參謀開 `beginStaffPick(PostMilitary, StateProductionPlan)`。
- `staffpick.go`：`staffPickDecide` 回到 `back` 那一張；換軍事參謀時更新已開著的
  建造計劃 `g.plan.StaffOps`；在預算計劃換內政參謀時重算 `monthReport`（同 `stepTaxLevel`）。
  新增 `beginBuildPlanStaffDetail`、`buildPlanAdmiralAt`。
- `game.go`：預算計劃與建造計劃清單的分支補畫參謀疊層；建造計劃清單點列開詳細資料。
- `controlhover.go`：生產計劃的游標熱區改用 `ds:26E2` 那一份。

## 五、要補原版收據時

起點用 `tools/fleetcmd_receipts.sh` 的 `fcmd-plan-prod`（生產計劃）與月初那條路的
`admirals1`。生產計劃參謀的滑鼠座標 (463,329)，清單第一列 (450,52)，決定 (497,290)。
`-peek lin:5DA7B:6` 一次倒 `ds:8F4B`–`8F50` 三個幕僚職。收據綁執行器版本，
先重建 `~/cht/dosgolem-logh3/`（`logh3-oracle` 分支）並重跑整條檢查點鏈。
