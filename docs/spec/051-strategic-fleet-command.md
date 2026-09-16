# 051 — 戰略地圖上的艦隊命令（艦隊視窗 → 作戰 → 目的地／待命行星 → 確認）

狀態：**CONFORMED（移動／駐留／輸送／待命）**：清單點列、艦隊視窗、作戰選單、
移動／駐留／輸送的目的地選取、確認框、寫入命令、取消命令、關窗與各層右鍵、
下了命令之後再開視窗、按「實行」之後的一回合、待命的「指定待命行星」
——三十批 `fcmd-*` 收據在正常玩家路徑下整張 640×480 逐位元組相同
（指定待命行星那兩張各排除小圖最後一列 128 個像素，見已知差異），顯示層
（Ebiten）`fcmd-win`／`fcmd-done`／`fcmd-exec`／`fcmd-standby-pick` 同樣相同
（`tools/gui_check.sh`）。占領、「變更」那兩項、選單的停用旗標**還不在**
（見「排除範圍」）。
證據：[`docs/re/312`](../re/312-strategic-fleet-command-flow.md)（§七是實作
對拍時補讀的）、[`docs/re/313`](../re/313-fleet-command-round-2.md)（第二輪：
記號、取消規則、駐留／輸送、實行之後的錢與地圖）、
[`docs/re/314`](../re/314-fleet-standby-planet.md)（待命：指定待命行星）。
收據：`fcmd-*` 三十批（`tools/fleetcmd_receipts.sh`，dosgolem `b5485eb`），
摘要在 [`docs/playtest/021`](../playtest/021-strategic-fleet-command.md)。
相關：[`043`](043-ai-capital-reach.md)（可達表，本片從艦隊所在星域起算）、
[`022`](022-fleet-detail.md)（十二個能力值與色 5 的規則同源）、
[`docs/internals/33`](../internals/33-fleet-table.md)（`+0x07`／`+0x08`）。

## 為什麼是這一片

remake 到目前為止玩家建得了艦隊、打得了仗，卻**派不出去**：艦隊編成完只能
停在首都，戰鬥只有 AI 或讀檔才觸發得了。原版派艦隊的入口在戰略地圖介面
右下角的「艦隊」分頁（不在戰略命令選單裡），這一片把那條路接上。

## 玩家路徑（typed input → state）

座標是**畫面座標**（收據旗標是滑鼠座標，y 少 40）。

| # | 狀態 | 輸入 | 下一個狀態 | 收據 |
|---|---|---|---|---|
| 1 | `StateCommandPhase` | 左鍵點右下角清單第 k 列（`x` 336–623、`y` 300＋20k … ＋19，k＝0–4），且第 `捲動位移＋k` 筆存在 | `StateFleetWindow`（那支艦隊）| `fcmd-win` |
| 1' | 同上 | 點在清單上但那一列沒有艦隊 | 不變 | （`fcx-row0` 的第一次：點 (450,350) 什麼都沒變）|
| 2 | `StateFleetWindow` | 左鍵「作戰」(480,340) 32×20 | `StateFleetActionMenu`（反白 ＝ 游標所在那一項）| `fcmd-menu` |
| 2' | 同上 | 左鍵「結束」(480,380) 32×20；或右鍵 | `StateCommandPhase` | `fcmd-close`、`fcmd-win-rclick` |
| 2'' | 同上 | 左鍵「變更」(480,360) | **本片不接**（原版開 副官／編組 小選單）| — |
| 3 | `StateFleetActionMenu` | 游標移動 | 反白跟著游標：第 n 項 `y` 300＋20n … ＋19（n＝0–4），`x` 464–527 | `fcmd-menu-hover` |
| 3' | 同上 | 右鍵；左鍵點在五項外 | `StateFleetWindow`（選單關掉，「作戰」鈕不反白）| `fcmd-menu-rclick`、`fcx-pop-255` |
| 3'' | 同上 | 左鍵第 0 項「移動」、第 2 項「駐留」、第 3 項「輸送」 | `StateFleetTargetPick`（動作 0／2／3；動詞 移動至／駐留在／輸送物資至）| `fcmd-pick`、`fcmd-garrison-pick`、`fcmd-transport-pick` |
| 3''' | 同上 | 左鍵第 4 項「待命」 | `StateFleetStandby`：「指定待命行星」（下一節）| `fcmd-standby` |
| 3'''' | 同上 | 左鍵第 1 項「占領」 | **還沒接**（見排除範圍；先照 3' 關選單）| — |
| 4 | `StateFleetTargetPick` | 左鍵點地圖上的星域 r，`允許[r] == 1`，r ≠ 現在位置 | `StateFleetConfirm(r)`：「要移動至<陣營色><陣營名>◇<星域名>星系嗎？」中止／決定 | `fcmd-confirm` |
| 4' | 同上 | 左鍵點星域 r，`允許[r] == 0` | `StateFleetMessage`：「這是無法前往的星系。」只有「確認」；關掉回 `StateFleetTargetPick` | `fcmd-far` |
| 4'' | 同上 | 左鍵點現在位置 | `StateFleetConfirm(現在位置)`，動詞換成「待命」；決定 ＝ 取消既有命令（`+0x08 ← 5`、`+0x07 ← 0xFF`），**不算有改，視窗留著**（回 `StateFleetWindow`）；駐留例外——一律寫 `+0x08 ← 2`、`+0x07 ← 現在位置` 並關窗（`docs/re/313` §二，讀自程式碼）| `fcmd-self`、`fcmd-self-done` |
| 4''' | 同上 | 左鍵點地圖外（星域 0xFF、或最上層不是地圖視窗）| 不變 | （反組譯 `113A:01AB`）|
| 4'''' | 同上 | 右鍵 | `StateFleetWindow`（視窗整張回來）| `fcmd-pick-rclick` |
| 5 | `StateFleetConfirm(r)` | 左鍵「決定」(416,300) 32×20 | 寫命令（`+0x07 ← r`、`+0x08 ← 動作碼`）→ `StateCommandPhase`；清單那一列「第」前面印記號（移動 ☆、駐留 ＊、輸送 ◎，`fleet-list-panel.json` 的 `marker`）| `fcmd-done`、`fcmd-garrison-done`、`fcmd-transport-done` |
| 5' | 同上 | 左鍵「中止」(368,300) 32×20 | `StateFleetTargetPick` | `fcmd-cancel` |
| 6 | `StateCommandPhase` | 再點清單那一列 | `StateFleetWindow`：第 5–7 列印「移動戰鬥／現在位置／移動目標 ◇星域名」（`+0x08 != 5` 的分支）| `fcmd-reopen` |
| 7 | `StateCommandPhase` | 系統指令 → 實行 | 艦隊走一格（`rules.Advance`）；到了就 `+0x07 ← 0xFF`、`+0x08 ← 5`；**沒有**出發／到達報告；維持費的一半與移動經費流進行星經濟力、星域經濟力重加總、有自軍艦隊的星域在地圖上換色表（[`docs/re/313`](../re/313-fleet-command-round-2.md) §三）| `fcmd-exec` |

決定之後寫的東西（`fcmd-done/fleet0.bin` 與 `fcmd-base/fleet0.bin` 的差）：
`+0x07 ← r`、`+0x08 ← 0`（移動）；其餘 120 個位元組不動。之後
`090A:0E6D`（未安定度重算，remake ＝ `recomputeUnrest(g.turn)`，與
`docs/spec/042` 的「玩家編成」同一個接線）。

## 待命：「指定待命行星」（`0DEA:1320`，[`docs/re/314`](../re/314-fleet-standby-planet.md)）

18×11 格的視窗疊在艦隊視窗上（格 (8,4)，畫面 x 128–447、y 120–379，色 0x97，
「作戰」鈕反白）：「指定待命行星」、「<陣營色陣營名>第Ｎ艦隊  提督姓名」、
「星系 <所在星域名>」（不換色）；所在星域的行星一顆一顆鈕（格 (9,4+i) 起 8 寬，
標籤 `%-16s` 行星名，色 0x87；只畫存在的那幾顆）；「中止」(12,10)、「決定」(15,10)
各 2 寬。左邊 (152,220) 起一張 128×128 的單色小圖（那個星域的戰場地圖，色 1 有東西、
色 3 沒有），外框 (151,219)-(280,348) 色 7；每顆行星從鈕的左緣 (288, 230＋20i) 拉一條
擁有者陣營色的線到小圖上那顆行星 (152＋X／16, 220＋Y／16)，先畫往左上一個像素的
色 3 影子。版面與格式字串：`docs/internals/data/fleet-standby-panel.json`。

小圖：欄不縮、列 `來源 ＝ py − ⌊py／5⌋`（102 列拉成 128 列）；一格有東西 ＝ 圖塊
∉ {0, 1, 5}，或它是格線（列 %12 ＝ 0 或欄 %15 ＝ 0 的圖塊 0）。行星位置 ＝
`rules.PlanetSites`（`docs/re/89`）。

| # | 狀態 | 輸入 | 下一個狀態 | 收據 |
|---|---|---|---|---|
| S1 | `StateFleetStandby` | 開窗 | 艦隊 `+0x08 == 5` 且 `+0x07 != 0xFF` 的那一顆一開就反白（0x27）| `fcmd-standby`（沒有）|
| S2 | 同上 | 左鍵第 i 顆行星鈕（x 288–415、y 220＋20i … ＋19，i < 顆數）| 舊反白改回 0x87、第 i 顆 0x27；**被點過的鈕重畫在引線上面**，沒點過的引線仍壓在字上 | `fcmd-standby-pick` |
| S3 | 同上 | 左鍵「決定」(384,340) 32×20；或**右鍵** | 有反白：`+0x08 ← 5`、`+0x07 ← i`（星域內序號），`recomputeUnrest`，關艦隊視窗 → `StateCommandPhase`；清單那一列沒有記號。沒反白：只 `recomputeUnrest`，回 `StateFleetWindow` | `fcmd-standby-done`、`fcmd-standby-rclick` |
| S4 | 同上 | 左鍵「中止」(336,340) 32×20 | 清掉反白，回 `StateFleetWindow`；記錄不動 | `fcmd-standby-cancel` |
| S5 | 同上 | 左鍵在鈕外（含小圖、艦隊視窗露出的部分）；游標移動 | 不變 | （反組譯 `0DEA:16AB`）|
| S6 | `StateCommandPhase` | 再點清單那一列 | `StateFleetWindow`：第 7 列印「待命行星 <擁有者色行星名>」 | `fcmd-standby-reopen` |

## 允許表與地圖著色（`113A:00BE`）

```
可達 ＝ ai.ReachFrom(星圖, 擁有者, 陣營, 艦隊.Region)     ; 2B99:0392 從艦隊所在星域起算
允許[i] ＝ 可達[i] 且（擁有者[i] == 陣營 或 i == 艦隊.Region）   ; 移動：A ＝ 自己、B ＝ 0xFF
```

目的地模式下地圖整張重新著色（畫面上地圖視窗蓋住艦隊視窗左半，右半照舊）：

| 格子 | 顏色 |
|---|---|
| 艦隊所在 | `{帝國 6, 同盟 5, 中立 7}[擁有者]` |
| 不在允許表 | 3 |
| 其餘 | `{帝國 1, 同盟 12, 中立 7}[擁有者]` |

離開目的地模式（決定、中止之後再右鍵、右鍵）地圖畫回原樣。

## 艦隊視窗的內容

27×19 格：外框佔字元平面欄 7–33、列 0–18（畫面 x 112–543、y 40–419），
內容 25×17 在欄 8–32、列 1–17；熱區原點 ＝ 內容原點 ＝ 格 (8,1) ＝ 畫面 (128,60)。
建構參數 `3603:0006(0, 0x97, 7, 0, 25, 17, 1)`：cell_x 7、cell_y 0、25×17、樣式 1、色 0x97。
文字照字元平面（`docs/re/312` §二那張表），來源：

| 列 | 內容 | 來源 |
|---|---|---|
| 1 | 陣營名（色 6）、第 N 艦隊 | `Fleet.Faction`、`Number` |
| 2 | 提督 姓名＋軍階 | 提督記錄（同 `docs/spec/012`）|
| 3 | 等級 %2d、經驗值 %6d（色 6）| 提督 `+0x06`、經驗值（同 022）|
| 5–7 | 作戰 待命戰鬥／現在位置 星域名（色 6）／待命中 | `+0x08 == 5`：第 7 列「待命中」，`+0x07 != 0xFF` 時改印「待命行星 <行星名>」（擁有者色）；`+0x08 != 5`：「移動戰鬥／現在位置／移動目標 <星域名>」（`fcmd-reopen`、`fcmd-standby-reopen`）|
| 8 | 部隊組成 艦種字母 ＋ □ | 同 `docs/spec/021` 的字母表；空槽印 □（`0xA1BC`？以收據平面的字碼為準）|
| 9 | 飛彈 %3d% 戰鬥機 %3d% | `0DEA:06FE`（`docs/re/312` §7.5）：逐組 上限 += 艦種列 `+0x10`／`+0x14` × 艘數、現量 += 艦艇記錄 `+0x0E`／`+0x12`（艦種列 `+0x0E`／`+0x12` 是 0xFF 的艦種不算）；上限 0 印 `----`，否則 現量×100÷上限。remake 的編制還沒帶每組現量，新編艦隊現量 ＝ 上限（見已知差異）|
| 10–11 | 副官 姓名軍階／等級／經驗值 | 副官 ＝ 提督 `+0x07` |
| 12 | 維持費（色 5）、移動經費（色 5）| 艦隊 `+0x0A`（`rules.FleetUpkeep`）與 `+0x0C`（`rules.FleetOtherCost`，`docs/internals/33`）；收據 22／26 ＝ `fleet0.bin` 的 `+0x0A`／`+0x0C` |
| 13–17 | 艦隊士氣值、統率力、回轉／巡航／戰鬥速度、光線砲、飛彈、戰鬥機、占領、防禦 | 與 022 相同的十二格，色 5 ＝ 副官比提督高 |
| 15–17 右 | 作戰／變更／結束 三顆鈕 | 版面照收據 |

整列底線（樣式 `0x01`，畫在格子最後一列、色 15）在列 7、9、11、12；列 4
那條只到欄 27（`rule` 字串是 20 格）。頭像 `FACE.AMG[提督索引]`，矩形
量自收據（`docs/re/312` §二：約 x 448–527、y 58–157）。

視窗開著時右下角清單被點的那一列反白：欄 21–37 那 17 格底色 8 → 2、字不動
（`docs/re/312` §7.2；`fleet-list-panel.json` 的 `row`）。

## 左下角與右側面板（與 [`001`](001-region-summary-panel.md) 同一套規則）

- 目的地選取是另一個會看游標的迴圈：進去時「上一次量到的那一格」從 −1
  起算，游標底下那一格變了才重寫左下角（`fcmd-pick` 進場不動、
  `fcmd-confirm`／`fcmd-far`／`fcmd-self` 寫點到的那格、`fcmd-cancel`
  回到選取時清空）；按「決定」回地圖用回地圖介面自己的記憶
  （`fcmd-done` 留著亞爾提那）。
- 兩塊面板都是寫進去那一刻的文字：編成之後右側首都摘要與 `dec1` 留下的
  左下角仍寫「艦隊 0：0」，游標移到首都重寫的那一次才是「1：0」
  （`fcmd-hover-home`；星域記錄 `+0x28` 建艦隊時就已經是 1）。
- 選取模式的畫面：視窗 x < 320 那一半回到底板（著色地圖 ＋ 左下角連框），
  x ≥ 320 那一半留著（`docs/re/312` §7.3）。

## 玩家編的艦隊

「決定」建出來的艦隊與 AI 走同一支 `297B:0066`：`+0x09` ＝ 0xFF（沒有類別鍵，
`rules.NoClassKey`）、只有旗艦一組（艦種 ＝ 提督 `+0x08` 低四位、隊形槽 20、
從首都待命部隊扣滿編艘數）、`+0x0E` 由提督等級決定、營運 ＝ max(提督, 副官)
的 `+14`、維持費／移動經費照 `rules.FleetUpkeep`／`FleetOtherCost`
（`fcmd-win/fleet0.bin`：22／26）。星域記錄的艦隊數（`+0x28`／`+0x29`）當場重數
（`rules.RecountRegionFleets`）。

## 排除範圍（還沒做，另開 issue）

- 占領（動作 1：允許表換成 敵方＋中立、費沙警告、「你確定要占領費沙嗎」）。
  選單上點這一項**只關選單**。
- 「變更」→ 副官／編組。
- 清單的捲動（欄 17 的上下鈕）與四個分頁（行星／敵艦隊／提督）。
- `1213:0003(星域)` 那道檢查（三個收據都過，語意沒讀）；回合裡的三張報告
  「無法前往敵方的星系」「由於移動經費不足無法移動」與占領那一路。

## 已知差異

| 差異 | 類別 | 到期條件 |
|---|---|---|
| 飛彈／戰鬥機現量：remake 的 `FleetGroup` 沒有每組的 `+0x0E`／`+0x12`，一律印上限（新編艦隊本來就是 100%）| 資料缺口 | 編制帶艦艇記錄的現量（打過仗、讀檔的艦隊）就照 §7.5 算 |
| 作戰選單五項的啟用旗標（不是自己的星域／有戰鬥只剩「移動」；艦隊 `+0x00` bit0 關掉占領與輸送）本片沒接，五項一律可點 | 排除範圍（#41）| 接其餘四項時一起做，收據要含停用的樣子 |
| 指定待命行星的小圖最後一列（畫面 y 347、x 152–279）：原版那一列的來源列是 102，超出 102 列的地圖，讀到的是緩衝區後面的東西；remake 畫底色（色 3）。對拍時跳過那 128 個像素 | 原版讀越界 | 不到期——那 128 個 byte 是堆積區的鄰居，不是地圖資料 |
| 待命寫入時原版另把行星位置抄進艦隊 `+0x5C`／`+0x5E`（`0DEA:1759`）；remake 的艦隊記錄沒有這兩格 | 資料缺口 | 解出這兩格在戰場上的用途（`docs/re/314` §八）再補 |

## 驗證

- `TestFleetWindowMatchesOracle`：`fcmd-win` 整張 0／307,200。
- `TestFleetActionMenuMatchesOracle`：`fcmd-menu`（反白第三項）、`fcmd-menu-hover`（反白第一項）。
- `TestFleetTargetPickMatchesOracle`：`fcmd-pick` 整張（含重新著色的地圖）。
- `TestFleetConfirmMatchesOracle`：`fcmd-confirm`、`fcmd-self`（動詞「待命」）、`fcmd-far`（訊息框）。
- `TestFleetMoveOrderWritesFleet`：走完 1→5，艦隊的 `Target ＝ 2`、動作 ＝ 移動；`fcmd-done` 整張 0 差異（含清單那一列的 ☆）。
- `TestFleetReopenShowsTarget`（`fcmd-reopen`）、`TestFleetCancelOrderKeepsWindow`（`fcmd-self-done`）、`TestFleetGarrisonAndTransportOrders`（六張）、`TestFleetOrderThenExecuteMatchesOracle`（`fcmd-exec`：走到星域 2、命令清掉、整張相同）。
- 待命：`TestFleetStandbyMatchesOracle`（`fcmd-standby`、`fcmd-standby-pick`，各排除小圖最後一列）、`TestFleetStandbyDecideWritesPlanet`（`Target ＝ 1`、`Move ＝ 5`；`fcmd-standby-done`、`fcmd-standby-reopen` 整張）、`TestFleetStandbyCancelAndRightClick`（`fcmd-standby-cancel`、`fcmd-standby-rclick` 整張）。
- 右鍵三層與中止：`fcmd-win-rclick`／`fcmd-menu-rclick`／`fcmd-pick-rclick`／`fcmd-cancel`／`fcmd-close` 各一張整張對拍。
- `TestFleetHoverHomeCountsNewFleet`：決定之後首都記錄的艦隊數是 1，游標移到首都那塊重寫成 1：0（`fcmd-hover-home`）。
- GUI：`tools/gui_check.sh` 的 `fcmd-win`、`fcmd-done`、`fcmd-exec`（都整張）、`fcmd-standby-pick`（排除那 128 個像素）。

以上全綠（2026-09-16；`go test ./...` 與 `tools/gui_check.sh`）。
