# 下一輪 goal — 艦隊組成編輯器與副官變更（#42 後半），再接右下角分頁（#43）與星系情報視窗（#44）

日期：2026-09-16。這一輪（同日）把作戰五項全部接上並關 #41（五十八批 `fcmd-*`
收據整張逐位元組，規格 051 CONFORMED），「變更」只接到小選單與兩張子視窗的
開關（#42 前半），詳見 `docs/re/313`–`316`、`docs/playtest/021`、`WORKLOG.md`
最後三條。

下面這段可以直接當下一個 session 的開場提示詞。

---

## 提示詞

你在 `~/cht/logh3`（銀河英雄傳說III SP 的 DOS remake）。先照 `CLAUDE.md`
的開工順序：`tools/issues_snapshot.sh` → `tools/go.sh run ./cmd/worklist verify`
（現況：0 條斷言不成立，open 的是 #42、#43、#44、#16）→ 讀 `CONTEXT.md`
現況一覽與「玩家還碰不到的功能」→ `git status --short`。再讀直接相關的四份：
`docs/re/316`（§三、§四、§六）、`docs/spec/015`、`docs/spec/051`「變更」一節、
`docs/playtest/021`。

**本輪目標：讓玩家編得出旗艦以外的部隊。** 現在玩家「出擊命令」編的艦隊只有
旗艦一組（`fcmd-win/fleet0.bin`：`+0x0F` ＝ 1），艦隊組成那張畫面開得起來、
點不動——「變更 → 編組」與出擊那條路開的是同一張（`0FF4:000A`），做好一次兩邊
都通。使用者 2026-09-15 已裁決：先做艦隊命令這一族、每張新畫面維持 same-state
契約（RE → spec → dosgolem 收據 → READY → 實作 → 整張逐位元組）、關 issue ＋
commit ＋ push 不必再問。不要重問這三件事。

### 交付順序（便宜的先做，每一項做完就對拍、進版控）

**A. #42 艦隊組成編輯器（`0FF4:000A`）——這一輪一定要做完**

1. 反組譯 `0FF4:0909`（事件迴圈）、`0FF4:077D`／`1083`（回傳值怎麼定）、
   `0FF4:10C1`（決定：寫回編制）、`0FF4:1127`（中止：退回）、`0FF4:0094`／`00F5`
   （進畫面：是不是就是「先扣旗艦那一組」的那兩支，`docs/re/316` §四）。
   切片 `workplace/ida/fcx0654.com`（`0FF4:0000` ＝ 切片偏移 `0x9A00`）。
   熱區清單 `2A00`（七個艦種列 ＋ 陣形／中止／決定，規格 015）。待驗的假設：
   點艦種列只移選取條（已接）、點格盤空格放一組選中的艦種（`2AD2:0029` →
   `2703:1433` 從存量扣 `min(滿編, 存量)`）、點有部隊的格拆掉（`2AD2:018C` 還回
   存量、艦種 7 不還）、「陣形」套樣板（`screen-formation` 那八列，收據
   `form1`–`form3` 只看過清單）、部隊數上限 ＝ `+0x0E`（提督等級）。
2. 收據從 `fcmd-change-org.state`（既有艦隊、旗艦在中央）展開：點格盤一格 → 點
   另一艦種列再點一格 → 點已放的格 → 陣形 → 選一個 → 決定；每一步倒
   `fleet0.bin`（`+0x0F`、`+0x10` 起每組 4 bytes）與星域 0 的存量
   （`-peek lin:5F2FC:28`，帝國七個艦種）。出擊那條路另存一個檢查點
   `fcmd-fok.state`（`fcmd-base` 的序列跑到 350M）再錄同一組操作，確認兩條路
   同一套。決定之後再開艦隊視窗看「部隊組成」那一列與費用（`0DEA:05DD`／`0A80`）。
3. remake：`internal/game/fleetchange.go` 的 `fleetChangeOrgClick` 與 `game.go`
   `StateFleetComposition` 那段共用一個編輯模型（`Game.provisional` 已是「進畫面
   扣、中止退、決定照用」）；編制改動走 `rules.TakeRegionShips`／`ReturnRegionShips`；
   決定寫回 `Fleet.Composition`／`Groups`／費用（`rules.FleetUpkeep`／`FleetOtherCost`）。
   規格 015 補「編輯」一節並改 CONFORMED；規格 051「變更」V3''' 那列改成規則。

**B. #42 副官 變更／卸任（`07F9:043F` 的「副官」鈕）**

從 `fcmd-change-adj.state` 點 (430,300)（滑鼠）開 變更／卸任 兩列小選單（探索批次
`fcx-change-adj-menu`，與出擊那條路的 `fadj` 同一張）；再各點一項錄下去：
「變更」多半開一張候選提督清單（同陣營、沒職務、沒帶艦隊——AI 挑副官的過濾
在 `docs/re/208`），「卸任」把提督記錄 `+0x07` 清 0xFF。回來之後 `297B:04EF`
把副官抄進艦隊 `+0x05`、`297B:0546` 重算能力與費用。remake：`StateFleetAdjutant`
的兩項目前只畫不做，接上就兩條路都通。

**C. #43 右下角清單的四個分頁與捲動** — 線索都在 issue 本文（分頁鈕座標、
`073F:0522` 重畫入口、敵艦隊分頁要 poke 局面、捲動要六支以上）。

**D. #44 星系情報視窗** — 左鍵點星圖；「區域小圖與行星名」與指定待命行星同一套
（`screen.FleetStandbySpec.drawMinimap`／`drawLeaders` 可直接搬），線索在 issue 本文。

做到哪算哪：**A 這一輪一定要做完**；B、C、D 依序推進，做不完的把收據與反組譯
進版控、issue 本文補「卡在」。

### 每一項的完成定義

- 收據腳本化進 `tools/fleetcmd_receipts.sh`（同一條檢查點鏈；要 poke 的局面把
  poke 寫進那一行）；批次名 `fcmd-*`，探索用 `fcx-*`；一批 25M 步約 3 秒、
  冷啟動約 55 秒，不用省。
- 規格：015（編輯）、051（變更）改 CONFORMED 才關 #42；#43／#44 各自一份新 spec。
- 測試整張 640×480 比（`matches`），GUI 每個新畫面至少一例進 `tools/gui_check.sh`。
- `docs/re/317` 起逐次紀錄（不改舊的，只加指標）；`docs/internals/56`（艦隊組成
  面板）／`58` 更新事實；`docs/playtest/021` 追加表列（雜湊全寫）；`CONTEXT.md`／
  `WORKLOG.md`。
- 關 issue 前把 verify 換成 `go_test`，重抓快照、`verify` 0 條斷言不成立、
  `render -w`／`render-matrix -w`。commit 訊息尾巴 `Co-Authored-By: Claude …`，
  不放 session 連結；`git config user.email` 要是 `wicanr2@gmail.com`。

### 已知的坑（這一輪踩過的，別再踩）

- **`-poke` 要落在畫面重畫之前**（例如 170M 進戰略命令前）：原版畫面會殘留舊值，
  remake 逐幀從狀態算，poke 在重畫之後就永遠對不上（`docs/re/315` §三）。
- **訊息框有逾時**：`3765:006C` 把「訊息速度 × 2」交給計時器倒數，這條收據鏈
  （`-tick 20000`）量到一頁 36M 步就自動翻；點了之後 23M 步截來得及，別拉長。
  找翻頁時機用 `-watch-screen 100000 -screen-delta 50`。remake 沒做逾時（已知差異）。
- **改了幾格顏色之後原版只重畫變過的格子**（`341B:02E5`）：直接塗在畫面上的
  東西（小圖、引線）留著，點過的鈕蓋在上面（`FleetStandbyView.Repainted`）。
  差一個像素先想這個。
- 首都戰艦的待命艘數開局是 **78000**，進艦隊組成才扣旗艦那一組；量存量用
  `-peek lin:5F2FC:28`（星域 0 `+0x30` 起）。舊文件寫 75000 的都是艦隊組成開著時量的。
- 選單寬度：`3667:0004` 的寬度參數原版有時傳標籤位元組長度（變更那張 4 格寬），
  `fleet-command-menu.json` 每張選單可各自給 `width`。
- 疊在活畫面上的舊式視窗：`spec.Windows["overlay"].Draw(fb, font, 3, faces)` 可以
  單獨畫一個 `GameWindow`（`drawFleetChangeAdj`／`drawFleetChangeOrg` 的做法），
  不要再走整張 `screen-*` 靜態底板。
- 收據旗標是滑鼠座標，畫面 y ＝ 滑鼠 y ＋ 40；`-save-state` 放在 `-steps` 的前一步；
  多行 `rec`（反斜線續行）的單批重跑腳本要照 awk 把續行接起來。
- 對拍收斂三步：`tools/scripts/celldiff.py` → `planerows.py` → `crop_png.py`；
  `LOGH3_DUMP_REMAKE=1` 讓失敗的測試把 remake 畫面倒到 `workplace/remake-<批次>.bin`。
- `workplace/ida/*.i64` 連續 `idat` 會壞：一次多目標；壞了刪掉重建。
- Docker：只走 `tools/*.sh` 包裝器；收尾 `docker ps -a | grep logh3` 與
  `find . -user root`；**禁止任何 prune／rmi**。

### 不在這一輪

#16（提督資料表的紙本來源）、占領命令在回合裡的效果（走到敵方星域之後那一路，
規格 051 排除範圍）、音樂三項（交戰演出＋戰鬥曲、自產曲庫管線、MIDI 音色對
MPU-401）——排在 #42–#44 之後。不改 dosgolem，除非收據需要；改了就整條
`fcmd-*` 重跑。
