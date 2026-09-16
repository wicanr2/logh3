# 055 — 一覽表四張的列（星系資料／艦隊資料／提督資料／敵將資料）

狀態：**CONFORMED（列、右鍵、▲▼、游標）**：點列開的那一張視窗、關回清單或地圖、
▲▼ 一列、三種游標規則（地圖反白與左下角摘要）——三十九批 `fcmd-lst-*` 裡三十四批整張
逐位元組（星系情報視窗那三批排除小圖最後一列，規格 053 的同一塊）；顯示層
`fcmd-lst-adm-row`／`fcmd-lst-fleet-row`（`tools/gui_check.sh`）。**分頁 DRAFT**（畫面上是
量死的第一頁，點了沒事；五批收據錄了沒比，#51）。
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
| 7 | 四張 | 分頁鈕 | 不動（已知差異） | |

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

## 已知差異

| 差異 | 類別 | 到期條件 |
|---|---|---|
| 星系資料的「配屬」「稅金」與艦隊資料的「部隊」「作戰」分頁：原版點了換內容、每座位記住；remake 只畫第一頁、點了沒事 | 未實作 | #51：讀 `0D05:02FC`／`0C7F:0339` 的三種列，五批收據已錄 |
| ▲▼ 原版按住連捲（`3765:0255`）；remake 一下一列 | 輸入模型 | 做「按住」時一起接（規格 013／052 同一件事） |
| 艦隊資料士氣低於 50 換色是照右下角清單的規則假設的（`0C7F:0339` 沒讀） | 假設待驗 | 有一支士氣低於 50 的艦隊時對一次 |
| 原版把「游標搬到會換反白的列的同一瞬間按下」的那一下丟掉（`docs/re/323` §六）；remake 照收 | 原版的輸入時序 | 不到期——不是規則 |

## 驗證

- `internal/game/listscreen_test.go`：`TestListStarRowsMatchOracle`、`TestListFleetRowsMatchOracle`、
  `TestListAdmiralRowsMatchOracle`、`TestListEnemyRowsMatchOracle`（三十四批整張、三批排除小圖
  最後一列；另比位移、開的視窗與詳細資料的提督）。
- GUI：`tools/gui_check.sh` 的 `fcmd-lst-adm-row`、`fcmd-lst-fleet-row`。
