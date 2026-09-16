# 052 — 戰略地圖介面右下角清單：四個分頁與捲動

狀態：**CONFORMED**：二十六批 `fcmd-tab-*`／`fcmd-forg-rclick` 收據整張 640×480
逐位元組相同（都是正常玩家路徑，從 `fcmd-base` 接下去），四個分頁的筆數與位移逐批
對上原版的 `ds:9592`／`95B4`／`95D8`／`CFFC` 那幾格；顯示層（Ebiten）
`fcmd-tab-planet-row`／`fcmd-tab-admiral-down`／`fcmd-tab-fleet2` 同樣相同
（`tools/gui_check.sh`）。「敵艦隊」那一頁有列的局面十三批 `enm-*`（S3 帝國第一回合，
用改記憶體造出「同盟艦隊在瓦爾哈拉、艦艇有索敵等級」，規格 014）整張逐位元組相同。
證據：[`docs/re/319`](../re/319-fleet-list-tabs.md)（分派、四頁的列印與點擊、行星情報
視窗、出擊那條路的右鍵）、[`docs/re/325`](../re/325-enemy-tab-rows-and-window.md)
（敵艦隊那一頁有列、視窗、訊息框）、[`docs/re/312`](../re/312-strategic-fleet-command-flow.md) §一
（艦隊分頁的點列）、[`docs/re/313`](../re/313-fleet-command-round-2.md) §一（艦隊分頁的列）、
[`docs/re/273`](../re/273-region-supply-pool-refill.md)（星域層存量分回行星）
資料：[`docs/internals/data/fleet-list-panel.json`](../internals/data/fleet-list-panel.json)
（`tabs`）、[`docs/internals/data/planet-info-strategic.json`](../internals/data/planet-info-strategic.json)、
[`docs/internals/data/capital-panel.json`](../internals/data/capital-panel.json)（`tabs`）、
[`docs/internals/data/enemy-fleet-window.json`](../internals/data/enemy-fleet-window.json)、
[`docs/internals/data/message-box.json`](../internals/data/message-box.json)（`enemy_no_intel`）
執行器：dosgolem-logh3 `b5485eb`（`tools/fleetcmd_receipts.sh` 第 17 段、`tools/enemytab_receipts.sh`）
收據表：[`docs/playtest/021`](../playtest/021-strategic-fleet-command.md)、
[`docs/playtest/023`](../playtest/023-enemy-tab.md)

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
| T7 | 「敵艦隊」列 | 別的陣營、在那一格、索敵等級（各組艦艇 `+0x17` 取最大）> 0 的艦隊，照艦隊表的順序：等級 3 印 第Ｎ艦隊（陣營色）／提督／士氣（<50 色 2；同艦隊列的版面，記號那格空白），否則「第？艦隊」（陣營色）＋「？？？？？？？？？ ?? 」。新遊戲第一回合空頁、點空列沒事；等級 0 的不列 | `fcmd-tab-enemy`／`-row`、`enm-l3`／`enm-l1`／`enm-mix`／`enm-l0` |
| T7a | 「敵艦隊」列，等級 3 | 那一列反白（格 0–16 的底色換 2，鈕的立體邊不動）、開**敵艦隊視窗**（下表）；「結束」或右鍵關、反白還原；點在視窗別處不動 | `enm-l3-row`／`-ok`／`-rclick`、`enm-mix-row1` |
| T7b | 「敵艦隊」列，等級 1–2 | 那一列反白、跳訊息框「情報資料來源不足」（樣板 0）：**模態**，點框內別處與右鍵都不動，只有「確認」關；關了反白還原 | `enm-l1-row`／`-miss`／`-rclk`／`-ok`、`enm-mix-row0` |
| T7c | 清單的視窗或訊息框開著 | 游標經過星圖，左下角那塊不動（各自的事件迴圈不跑 hover） | `enm-l1-row-miss` |
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

### 敵艦隊視窗（`0B20:0336` → `0B20:0411`）

內容格 (8,6) 起 24 × 6（外面一格框），色 0x97。第 0 列「軍名（陣營色）  第Ｎ艦隊」、
第 1 列「提督      姓名軍階」、第 4 列「部隊種類  」＋ 各組（艦隊 `+0x0F` 組）的艦艇：
索敵等級 ≥ 2 印艦種字母（ＢＨＣＡＤＴＥＰ，`ds:6172`）、1 印「？」、0 不印。
「結束」在內容格 (21,5) 寬 2（色 0x87）；提督頭像 80×100 貼在 (448,180)。

## 已知差異

| 差異 | 類別 | 到期條件 |
|---|---|---|
| ▲▼ 原版按住連捲一次一列（`3765:0255`），remake 一下一頁；收據用 0.3M 步的按住剛好五列 | 輸入模型（沒有「按住多久」） | 與副官清單同一件事（規格 013）：做按住的事件模型時一起換 |
| 提督詳細資料沒有「出擊」的那一張（艦隊 16 支或首都易主）沒有收據，鈕的位置照熱區表 `1FAE` | 局面難造 | 有 16 支艦隊的局面時補一批 |

## 還沒做的

- 「敵艦隊」列的收據是改記憶體造的（S3 ＋ `-poke`），正常玩家路徑（打過一仗回到
  戰略地圖、敵艦隊留在同一格）還沒有收據——要等戰術層與原版對齊到同一個回合。
  索敵等級在 remake 是 `rules.FleetGroup.Detail`：載存檔時讀艦艇 `+0x17`、戰後
  `ApplyBattleResult` 從戰場單位寫回；新編的艦隊沒有艦艇池，永遠 0（與原版新遊戲
  一樣要打過仗才有）。
- 帝國艦隊當敵方（同盟座位）的陣營色只從 `ds:6092`／`\x03[150,` 推，沒有收據。

## 驗證

- `TestListTabPlanetMatchesOracle`：`fcmd-tab-planet`／`-row`／`-ok`／`-rclick`／`-scroll` 整張；筆數 4、位移 0。
- `TestListTabEnemyMatchesOracle`：`fcmd-tab-enemy`／`-row` 整張。
- `TestEnemyTabClearMatchesOracle`／`TestEnemyTabDimMatchesOracle`／`TestEnemyTabMixedMatchesOracle`／
  `TestEnemyTabUnseenIsEmpty`：十二批 `enm-*` 整張（`internal/game/enemytab_test.go`）。
- `TestListTabAdmiralMatchesOracle`：`fcmd-tab-admiral`／`-down`／`-down2`／`-up`／`-end`／`-row`／`-cancel`／`-rclick`／`-sortie`／`fcmd-tab-sortie-rclick`／`fcmd-tab-fleet2`／`-list` 整張；筆數 35 → 33、位移 5／10／0／30。
- `TestFleetOrgRightClickIsConfirm`：`fcmd-forg-rclick` 整張、存量 75000。
- `TestListTabFleetScrollMatchesOracle`：`fcmd-tab-fleet6`／`-list`／`-down`／`-up`／`-row`／`-close` 整張；艦隊頁 6 筆、位移 1、關窗仍 1；提督頁 25 筆。
- `TestListTabPerFaction`：分頁每個陣營各一份。
- `internal/rules`：`TestIdleAdmiralsInRegionAndListPage`。
- GUI：`tools/gui_check.sh` 的 `fcmd-tab-planet-row`、`fcmd-tab-admiral-down`、`fcmd-tab-fleet2`。

以上全綠（2026-09-17；`go test ./...` 與 `tools/gui_check.sh`；`enm-*` 是改記憶體的局面，GUI 那條沒有對應案例）。
