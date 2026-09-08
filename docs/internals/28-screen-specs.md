# 畫面規格 JSON：`data/screen-*.json`

證據等級：**confirmed**——每一份都對應一張逐位元組對拍過的畫面
（`VERIFICATION-MATRIX.md`），量測命令在 `tools/specs.sh`。

remake 的畫面不是寫死的程式碼，是**資料**：一份 JSON 描述一張畫面，
`internal/screen.MainGame` 照著畫。所以「多一張畫面」多半是多一份 JSON，
不是多一段繪圖程式。

## 怎麼產生

```
tools/specs.sh              # 全部重量一次（十四張）
tools/specs.sh cmd-menu     # 只做一張
```

量測靠 `tools/scripts/`：`cmdscreen_spec.py` 量整張底板，
`overlay_spec.py` 量「疊在既有畫面上的一層」（底層原封不動沿用，
重量一次會因為文字被切在視窗邊界而讀錯，見 `docs/re/12`）。

## 欄位

| 欄位 | 內容 |
|---|---|
| `canvas` | 640 × 480，VGA mode 12h |
| `background` | 底色色號 |
| `map` | 星圖：`source`／`regions` 是原版檔名，`x`／`y` 是畫在哪，`colours` 指定用 `strmap-regions.json` 的哪一組星域色（空的話用勢力色）|
| `windows` | 一個名字對一個視窗（`board`、`report`、`overlay`）|
| `order` | 疊在地圖之上的視窗畫的順序，先畫的在下面 |
| `hotspots` | 可點區域：像素矩形 ＋ 原始格子座標 ＋ 動作名 |
| `mouse` | 游標位置與可見範圍，見下 |

### 視窗

| 欄位 | 內容 |
|---|---|
| `x`／`y`／`w`／`h` | 位置與大小 |
| `frame.layers` | 立體邊由外往內每一層的四邊色號 `[上, 右, 下, 左]` |
| `frame.alt` | 同一層的交錯色（網點邊）；0 表示不交錯 |
| `frame.fill` | 內部底色 |
| `frame.patches` | 四個角對不上模型的個別像素修正 |
| `boxes` | 視窗裡自成一格的小框（清單列、按鈕），欄位同 `frame` |
| `rules` | 純色塊（表格分隔線、反白條）|
| `lines` | 文字：每一行有 `y`、`ink`、`bold`，`cells` 是 `{x, w, code}` |
| `face` | 頭像：`FACE.AMG` 的第幾張、畫在哪 |

`lines[].bold` 是原版的粗體（字形每一列往右加粗一格，`docs/internals/25`）。
反查字碼時**粗體要拿加粗後的點陣當鍵**，否則整行讀不到。

### 游標

| 欄位 | 內容 |
|---|---|
| `mouse.x`／`y` | 滑鼠座標；游標畫在 `(x, y + 40)`（`cursor.json` 的 `offset_from_mouse`）|
| `mouse.clip` | 游標的可見矩形。原版的游標是疊上去的，視窗一重畫就蓋掉一部分 |
| `mouse.after_boxes` | 這個視窗的小框在游標之後再畫一次 |
| `mouse.after_box_fills` | 同上，但**只填游標壓到的那幾個框的內部**，不重畫立體邊 |
| `mouse.after_rules` | 這個視窗的純色塊在游標之後再畫一次（滑鼠移過去才刷的反白條）|

那三個 `after_*` 是同一件事的三種粒度：**原版先畫游標、再刷游標底下那一塊**，
所以游標會被蓋掉一部分。粒度挑錯就會多擦或少擦幾個像素
（`docs/playtest/006`）。

## 目前有哪些

| 規格 | 畫面 | 收據 |
|---|---|---|
| `screen-main-menu` | 主選單 | `r0` |
| `screen-copy-protection` | 防拷問答 | `cb1` |
| `screen-main-game` | 遊戲主畫面（月初報告）| `key1` |
| `screen-ship-order` | 船艦訂造報告 | `ship1` |
| `screen-build-plan` | 次月份船艦建造計劃 | `plan1` |
| `screen-budget-plan` | 次月份預算計劃 | `budget1` |
| `screen-admiral-list` | 待命中提督資料（疊在建造計劃上）| `admirals1` |
| `screen-account` | 月份會計報告 | `account1` |
| `screen-command-phase` | 戰略地圖介面 | `main1e` |
| `screen-cmd-menu`／`screen-list-menu`／`screen-sys-menu` | 三張下拉選單 | `cmd1e`／`lstmenu1e`／`sysmenu1e` |
| `screen-standby-admirals` | 出擊命令 → 待命中提督資料 | `sortie1e` |
| `screen-info` | 戰略情報報告 | `info1e` |
| `screen-tax-plan`／`screen-production-plan` | 徵稅／生產計劃 | `tax1e`／`prod1e` |
| `screen-star-list`／`screen-fleet-list`／`screen-admiral-list-view`／`screen-enemy-list` | 一覽表四張 | `star1e`／`lst1e`／`lst2e`／`lst3e` |
| `screen-load-game`／`screen-save-game` | 取進度／存進度 | `load1e`／`save1e` |
| `screen-settings` | 遊戲設定 | `settings1e` |
| `screen-quit-confirm` | 結束確認 | `quit1e` |
| `screen-ship-delivery` | 月底的船艦接收報告 | `turnmonth2` |
| `screen-admiral-detail` | 提督詳細資料 | `adm1e` |
| `screen-admiral-bio` | 提督經歷（人物傳）| `keireki1e` |
| `screen-fleet-org` | 艦隊編成 | `fleet1e` |
| `screen-versus-input` | 對戰的交棒視窗 | `vs3`（帝國）／`vs5`（同盟）|
| `screen-versus-account` | 對戰的月初會計報告 | `vs4` |

## 現在的限制

**文字是量出來畫死的**，不是從遊戲狀態算出來的。左下角的星域摘要、
右側面板的數字、日期那一行都寫在 `lines` 裡，所以換一個局面就不對了。
把它們改成從規則算出來是 `docs/spec/` 的工作。
