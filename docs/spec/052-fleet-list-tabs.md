# 052 — 戰略地圖介面右下角清單：四個分頁與捲動

狀態：**CONFORMED（艦隊／行星／提督三頁與捲動）**：二十六批 `fcmd-tab-*`／
`fcmd-forg-rclick` 收據整張 640×480 逐位元組相同（都是正常玩家路徑，從 `fcmd-base`
接下去），四個分頁的筆數與位移逐批對上原版的 `ds:9592`／`95B4`／`95D8`／`CFFC`
那幾格；顯示層（Ebiten）`fcmd-tab-planet-row`／`fcmd-tab-admiral-down`／
`fcmd-tab-fleet2` 同樣相同（`tools/gui_check.sh`）。
「敵艦隊」那一頁只對過空清單（新遊戲第一回合），有索敵等級的列與它的視窗是
**DRAFT**（讀自程式碼，見「還沒做的」）。
證據：[`docs/re/319`](../re/319-fleet-list-tabs.md)（分派、四頁的列印與點擊、行星情報
視窗、出擊那條路的右鍵）、[`docs/re/312`](../re/312-strategic-fleet-command-flow.md) §一
（艦隊分頁的點列）、[`docs/re/313`](../re/313-fleet-command-round-2.md) §一（艦隊分頁的列）、
[`docs/re/273`](../re/273-region-supply-pool-refill.md)（星域層存量分回行星）
資料：[`docs/internals/data/fleet-list-panel.json`](../internals/data/fleet-list-panel.json)
（`tabs`）、[`docs/internals/data/planet-info-strategic.json`](../internals/data/planet-info-strategic.json)、
[`docs/internals/data/capital-panel.json`](../internals/data/capital-panel.json)（`tabs`）
執行器：dosgolem-logh3 `b5485eb`（`tools/fleetcmd_receipts.sh` 第 17 段）
收據表：[`docs/playtest/021`](../playtest/021-strategic-fleet-command.md)

## 怎麼到得了

戰略地圖介面右下角那張清單的上緣有四顆分頁鈕 艦隊／行星／敵艦隊／提督
（畫面 y 280–299，x 336／400／464／544 起）。內容區是 18 × 5 格（x 336–623、
y 300–399），欄 17（x 608–623）的列 7 與列 11（y 300–319、380–399）是 ▲▼。

## 規則（typed input → 狀態）

| # | 輸入 | 規則 | 收據 |
|---|---|---|---|
| T1 | 點分頁鈕 | 選中的分頁換成那一顆（核取方塊填 2，其餘 9）；**每個陣營各記一份**（`ds:8F29[座位]`），對戰模式換手時換回自己的；那一頁的捲動位移歸零、內容照清單看的星域（右鍵選過的，沒有就自己的首都）重建 | `fcmd-tab-planet`／`-enemy`／`-admiral`、`fcmd-tab-fleet2-list`、`fcmd-tab-fleet6-list` |
| T2 | 每一頁 | 筆數／捲動位移／五列一頁；印之前 `位移 ＋ 5 > 筆數 → 位移 ＝ max(筆數 − 5, 0)`；筆數 > 5 才印 ▲▼，否則那兩格印空白圖塊 | 同上 |
| T3 | ▲／▼ | 位移 −／＋，夾回範圍。原版按住連捲一次一列；remake 一下一頁（已知差異）。行星那一頁的捲動熱區在列 6 與 11（不是 7 與 11）而且方向相反，行星最多五顆所以永遠不動——remake 不接它 | `fcmd-tab-admiral-down`／`-down2`／`-up`／`-end`、`fcmd-tab-fleet6-down`／`-up`、`fcmd-tab-planet-scroll` |
| T4 | 「艦隊」列 | 第 (位移＋列) 支；點開艦隊視窗（規格 051），那一列反白；關窗**不**重建、位移留著 | `fcmd-tab-fleet6-row`／`-close` |
| T5 | 「行星」列 | 一列 ＝ 擁有者名（陣營色）＋ 行星／要塞（陣營色）＋ 名字 ＋ 防衛力（與上限相同 7、剩一半以上 5、剩不到一半 2）；點了那一列反白、開**行星情報視窗**（下表）；「確認」或右鍵關、反白還原 | `fcmd-tab-planet`／`-row`／`-ok`／`-rclick` |
| T6 | 「提督」列 | 清單 ＝ 那一格裡同陣營、沒有職務、沒帶艦隊的提督（`28E0:0103(星域, 0x25)`，`rules.IdleAdmiralsInRegion`），一列 ＝ 姓名 ＋ 軍階；點了那一列反白、開提督詳細資料（與出擊命令那張同一支）：有「出擊」（首都還是自己的且艦隊不到 16 支）是 經歷／中止／出擊，否則 經歷／結束 兩顆往右挪；中止／結束／右鍵關、反白還原；「出擊」＝ 出擊命令那條路（`07F9:00D5`）：進艦隊編成就配槽，確認 → 艦隊組成 → 決定建艦隊回地圖，**選著的分頁不變**、艦隊與提督兩頁位移歸零、那一位與他的副官從清單消失 | `fcmd-tab-admiral`／`-row`／`-cancel`／`-rclick`／`-sortie`、`fcmd-tab-fleet2`／`-list`、`fcmd-tab-fleet6` |
| T7 | 「敵艦隊」列 | 別的陣營、在那一格、索敵等級（各組艦艇 `+0x17` 取最大）> 0 的艦隊：等級 3 印 第Ｎ艦隊／提督／士氣（同艦隊列的版面），否則「第？艦隊」＋ 九個問號。新遊戲第一回合空頁、點空列沒事 | `fcmd-tab-enemy`／`-row` |
| T8 | 出擊那條路的艦隊編成按右鍵 | ＝ 確認：直接進艦隊組成（`07F9:00D5` 不看 `043F` 的結果就建艦隊）。要作廢得在艦隊組成按中止 | `fcmd-tab-sortie-rclick`、`fcmd-forg-rclick` |
| T9 | 左下角那塊 | 只有清單那幾張（提督資料、待命中提督資料）自己寫首都；從右下角走到詳細資料、艦隊編成、艦隊組成一個字都不動（出擊五次仍是「治安 82%／艦隊 0：0」） | `fcmd-tab-fleet6` |

### 行星情報視窗（`2AF5:0226`）

格 (9,5) 起 23 × 11 格（內容 21 × 9），色 0x97。內容與戰場那張（規格 019 那一支，
`docs/re/144`）同一家族：星系（擁有者色）、行星／要塞＋名字、空行、經濟力與成長率、
防衛力／上限與回復率、七個艦種的字母與**補充部隊數 ＝ 補充存量 ÷ 滿編**、飛彈、
戰鬥機各一行；不是自己的行星時後面幾段印 `----`。右下角一顆「確認」。
開窗那一刻先把這一格星域層的存量按各顆行星經濟力的整數百分比分回行星
（`2703:0540`，`rules.DistributeRegionSupply`：七種艦艇進 `+0x20`、飛彈／艦載機進
`+0x3C`／`+0x40`，餘額給經濟力最大那顆）。開新局時同一條分配跟著開機那次回充跑
一遍（星域層飛彈／艦載機夾成「載入時從行星加總的經濟力 ＋ 防衛力」）。

## 已知差異

| 差異 | 類別 | 到期條件 |
|---|---|---|
| ▲▼ 原版按住連捲一次一列（`3765:0255`），remake 一下一頁；收據用 0.3M 步的按住剛好五列 | 輸入模型（沒有「按住多久」） | 與副官清單同一件事（規格 013）：做按住的事件模型時一起換 |
| 提督詳細資料沒有「出擊」的那一張（艦隊 16 支或首都易主）沒有收據，鈕的位置照熱區表 `1FAE` | 局面難造 | 有 16 支艦隊的局面時補一批 |

## 還沒做的

- 「敵艦隊」分頁有索敵等級的列、它的視窗（`0B20:0336`：第Ｎ艦隊／提督／部隊種類、
  「確認」）與等級不到 3 的「情報資料來源不足」訊息：remake 的戰略層艦隊沒帶
  艦艇的索敵等級（`+0x17` 是戰場上累積的，`docs/internals/38`），要先把它接到
  `rules.FleetGroup`、再造一個打過仗的局面（讀 S3 存檔或 `-poke`）取收據。
  `docs/re/319` §七有列印格式與視窗版面。

## 驗證

- `TestListTabPlanetMatchesOracle`：`fcmd-tab-planet`／`-row`／`-ok`／`-rclick`／`-scroll` 整張；筆數 4、位移 0。
- `TestListTabEnemyMatchesOracle`：`fcmd-tab-enemy`／`-row` 整張。
- `TestListTabAdmiralMatchesOracle`：`fcmd-tab-admiral`／`-down`／`-down2`／`-up`／`-end`／`-row`／`-cancel`／`-rclick`／`-sortie`／`fcmd-tab-sortie-rclick`／`fcmd-tab-fleet2`／`-list` 整張；筆數 35 → 33、位移 5／10／0／30。
- `TestFleetOrgRightClickIsConfirm`：`fcmd-forg-rclick` 整張、存量 75000。
- `TestListTabFleetScrollMatchesOracle`：`fcmd-tab-fleet6`／`-list`／`-down`／`-up`／`-row`／`-close` 整張；艦隊頁 6 筆、位移 1、關窗仍 1；提督頁 25 筆。
- `TestListTabPerFaction`：分頁每個陣營各一份。
- `internal/rules`：`TestIdleAdmiralsInRegionAndListPage`。
- GUI：`tools/gui_check.sh` 的 `fcmd-tab-planet-row`、`fcmd-tab-admiral-down`、`fcmd-tab-fleet2`。

以上全綠（2026-09-16；`go test ./...` 與 `tools/gui_check.sh`）。
