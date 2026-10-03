# 055 — 一覽表四張的列（星系資料／艦隊資料／提督資料／敵將資料）

狀態：**CONFORMED（列、右鍵、▲▼、游標）**：點列開的那一張視窗、關回清單或地圖、
▲▼ 一列、三種游標規則（地圖反白與左下角摘要）——三十九批 `fcmd-lst-*` 裡三十四批整張
逐位元組（星系情報視窗那三批排除小圖最後一列，規格 053 的同一塊）；顯示層
`fcmd-lst-adm-row`／`fcmd-lst-fleet-row`（`tools/gui_check.sh`）。**分頁 CONFORMED**：
五批整張逐位元組（`docs/re/327`），顯示層 `fcmd-lst-star-tab-tax`。
證據：[`docs/re/323`](../re/323-list-screen-rows.md)；四張的內容本身見 [`docs/re/37`](../re/37-empty-lists.md)、
[`docs/re/50`](../re/50-admiral-lists.md)、[`docs/re/64`](../re/64-command-phase-follows-the-player.md)。
收據摘要：[`docs/playtest/021`](../playtest/021-strategic-fleet-command.md)。
相關：[`053`](053-region-info-window.md)（星系資料點列開的視窗與換選）、[`051`](051-strategic-fleet-command.md)
（艦隊資料點列開的艦隊視窗）、[`012`](012-admiral-detail.md)（提督／敵將資料點列開的詳細資料；
現在位置後面的狀態與姓名前那一格在這一片量到）、[`052`](052-fleet-list-tabs.md)（右下角清單，
同一種列與反白）。

## 版面（`docs/internals/data/list-screen-panel.json`）

四張同一個視窗：畫面 (320,40) 320 × 340，內容原點 (336,60)。內容區 (336,80) 272 × 320 ＝ 十六列，
列 k 底板 y 80＋20k、字 82；▲ (608,80) 16 × 20、▼ (608,380)；分頁鈕 (432|496|560, 60) 48 × 20。
點開子視窗時那一列刷成底色 2（`3603:00DB` 的 0x20），關了還原。

## 規則

| # | 狀態 | 輸入 | 下一個狀態 | 收據 |
|---|---|---|---|---|
| 1 | `StateListMenu` | 星系資料／艦隊資料／提督資料／敵將資料 | 那一張；艦隊與提督／敵將的位移歸零，星系的位移與分頁每座位記住（`ds:1DF2`／`8F2D`） | `fcmd-lst-star`／`-fleet`／`-adm`／`-enemy`、`fcmd-lst-star-reopen` |
| 2 | `StateStarList` | 左鍵列 k（`位移＋k < 32`） | 換選那一格（規格 053 的 `selectRegion`）→ 關清單 → `StateRegionInfo`；結束回地圖 | `fcmd-lst-star-row`／`-row2`／`-down-row`、`-row-close`／`-row2-close` |
| 3 | `StateFleetList` | 左鍵有艦隊的列 | 那一列反白、`StateFleetWindow`（規格 051 那一張，底下仍是清單）；結束或右鍵回 `StateFleetList`（反白還原、地圖沒反白、左下角空白）；空列沒事 | `fcmd-lst-fleet-row`、`-row-close`、`-row-rclick`、`-row1` |
| 4 | `StateAdmiralListView`／`StateEnemyList` | 左鍵有人的列 | 那一列反白、提督詳細資料疊上來（規格 012；旗標 0：經歷／結束）；結束或右鍵回清單；經歷疊在上面、只認右鍵 | `fcmd-lst-adm-row`、`-row-close`、`-row-rclick`、`-row-bio`、`-row-bio-close`、`fcmd-lst-enemy-row`、`-row-close` |
| 5 | 四張 | ▲／▼ | 位移 ∓1，夾在 `0 … 筆數 − 16`（原版按住連捲，一列一次） | `fcmd-lst-star-down`／`-up`、`fcmd-lst-adm-down` |
| 6 | 四張 | 右鍵 | 子視窗開著只關子視窗；否則回 `StateCommandPhase`（地圖反白回右側面板看的那一格） | `fcmd-lst-*-rclick` |
| 7 | 星系／艦隊資料 | 分頁鈕（熱區碼 3／4／5） | 換那一頁、整頁重印，每座位記住（`ds:8F2D`／`ds:8F31`）；提督／敵將沒有分頁熱區 | `fcmd-lst-star-tab-assign`／`-tab-tax`／`-reopen`、`fcmd-lst-fleet-tab-unit`／`-tab-ops` |

游標（`3705:0591`；每一種都是「與現在反白的同一格就不動」）：

| 清單 | 停在有東西的列 | 停在空列／▲▼ | 清單外 |
|---|---|---|---|
| 星系 | 反白那一格、左下角寫它 | 不動 | 不動 |
| 艦隊 | 反白那一支所在的格、左下角寫它（▲ 算列 0、▼ 算列 15） | 取消反白、左下角清空 | 不動 |
| 提督 | 同艦隊（那位提督的 `+0x0A`） | 取消反白、清空 | 取消反白、清空 |
| 敵將 | 不看游標 | | |

開清單那一刻只把反白從緩衝區拿掉、不重畫（`3705:0585`）：畫面上仍是開清單前的反白，
到第一次換反白才變（`Game.lst.hlDrawn`）。

## 提督詳細資料多量到的（規格 012 的「還沒做的」到期）

姓名前那一格：帶艦隊「第Ｎ艦隊」（`+0x09` 低七位是艦隊槽）、幕僚職「參謀」（`0x7F`）、
待命空白（bit7）；現在位置後面「：待命中／執勤中／出擊中」同一個分法；敵將那張現在位置
印 ＊＊＊＊＊＊＊＊、沒有狀態、「同盟軍」用陣營色 5。清單的記號同一個分法：○ 幕僚職、
★ 帶艦隊。

## 分頁（`docs/re/327`）

選中那顆核取方塊填 2、其餘 9。名字後面那一段：

| 清單 | 分頁 | 自己的 | 別家 |
|---|---|---|---|
| 星系 | 艦隊 | 帝國／同盟艦隊數「%2d﹕%2d」（`+0x13` bit2 時「﹕」用色 2） | 同左 |
| 星系 | 配屬 | 八格空白、一格空白、七種艦艇各一字：0 `-`、不到一個滿編單位小寫、否則大寫（`BHCADTE`） | 「 -------」 |
| 星系 | 稅金 | 「%8lu (%2d.%1d%%)」＝ 經濟力 × 稅率 ÷ 1000、稅率（千分之一） | 「     --- (-----)」 |
| 艦隊 | 士氣 | `+0x3C`，小於 50 用色 2 | — |
| 艦隊 | 部隊 | 「%2d/%2d」＝ `+0x0F`／`+0x0E` | — |
| 艦隊 | 作戰 | 所在星域 bit2「戰鬥」、bit4「占領」，否則命令名（移動／占領／駐留／輸送／返航／待命） | — |

## 已知差異

| 差異 | 類別 | 到期條件 |
|---|---|---|
| 星系資料「艦隊」分頁：原版讀星域 `+0x28`–`+0x2B`（依座位挑兩格、`0xFF` 印 `--`），remake 照艦隊表數，敵方的數原版可能是 `--` | 來源不同（hypothesis） | 同盟座位或敵方艦隊在場時錄一批（`docs/re/327` §四） |
| 「﹕」戰鬥中換色 2、作戰分頁的「戰鬥」「占領」、士氣低於 50 換色：照反組譯接的，沒有收據 | 待收據 | 有那種局面時各對一次 |
| 原版按住連捲速度隨重畫成本變動；remake短按一列、保持350／100ms | 第58輪使用者核准的刻意速度差異 | 保留差異；正式入口與驗收見[089](089-left-button-repeat.md) |
| 原版把「游標搬到會換反白的列的同一瞬間按下」的那一下丟掉（`docs/re/323` §六）；remake 照收 | 原版的輸入時序 | 不到期——不是規則 |

## 驗證

- `internal/game/listscreen_test.go`：`TestListStarRowsMatchOracle`、`TestListFleetRowsMatchOracle`、
  `TestListAdmiralRowsMatchOracle`、`TestListEnemyRowsMatchOracle`（三十四批整張、三批排除小圖
  最後一列；另比位移、開的視窗與詳細資料的提督）；`TestListScreenTabsMatchOracle`（分頁五批）。
- GUI：`tools/gui_check.sh` 的 `fcmd-lst-adm-row`、`fcmd-lst-fleet-row`、`fcmd-lst-star-tab-tax`。
