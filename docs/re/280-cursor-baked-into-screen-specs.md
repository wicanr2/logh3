# 280 — 四份畫面規格裡烘著量測當時的游標

日期：2026-09-11
執行器：`logh3-oracle` `6841e68`
狀態：**confirmed**（刮掉之後四張畫面整張 640×480 逐位元組相同）

## 怎麼冒出來的

執行器修好之後（[`276`](276-what-the-executor-fix-changed-on-screen.md)
形狀一），收據末尾那次 `-mouse-at` 真的把游標搬走了：`info1e`／`lst1e`／
`lst2e`／`lst3e`／`settings1e` 與同盟的五張，十份都是「游標從剛剛點的那一格
搬到 `(600,340)`」。

remake 跟著搬之後，原來被游標壓住的地方露了出來，而那裡多了一小塊字：

| 規格 | 多出來的東西 | 位置 | 差 |
|---|---|---|---:|
| `screen-info` | `overlay.boxes[6]` ＋ `overlay.rules[0]` | `(360,145)` 14×15 | 274 px |
| `screen-fleet-list` | `overlay.rules[0..2]` | `(468,158)` 3×3 | 7 px |
| `screen-admiral-list-view` | 同上 | `(468,178)` 3×3 | 7 px |
| `screen-enemy-list` | 同上 | `(468,198)` 3×3 | 7 px |

三份清單畫面的那三條 `rules` 位置一模一樣、只差 20 像素（一列），而 20 就是
選單一列的高度——三份規格各自量的時候，游標停在選單的第 1／2／3 列，
量框的程式把游標的**尖端**當成畫面的一部分記了下來。`screen-info` 更完整，
連整隻游標都記成一個 14×15 的 box。

同一類事故在 [`../spec/001`](../spec/001-region-summary-panel.md)
「底板規格的游標污染」已經有一次：量 `main1e` 的時候游標壓在清單條上，
19 個游標像素被當成底板記了下來。**只要量測用的收據裡游標壓在那張畫面上，
就會發生**，而它在游標不動的年代完全看不出來。

## 處置

1. 四份規格裡那幾條刮掉；`screen-info` 另外補一個 `mouse`
   （原本是 `null`，游標被記成 box 了，沒有 `mouse` 就沒地方放活的游標）。
2. `internal/game.liveCursor` 的白名單補上六個狀態：`StateInfo`、
   `StateFleetList`、`StateAdmiralListView`、`StateEnemyList`、
   `StateSettings`、`StateFleetOrg`。前五個的證據是上面那十份收據，
   `StateFleetOrg` 是 `fnum2`／`fadj2`（游標停在剛剛點下去的框緣）。

3. `vs3` 的比對時機挪到那次點擊**之後**：收據的第三次點擊
   （`-clicks 170000000:400:222` ＝ 畫面 `(400,262)`）會把游標搬過去，
   先比就比到點下去之前那一張。

驗收是整張 640×480 逐位元組：`TestCommandPhaseNavigation` 與
`TestAllianceCommandPhaseScreens` 各十二張、`TestFleetOrgHotspotsAreNarrowerThanTheBoxes`
兩張、`TestVersusHandover` 的 `vs3`，全部 0 個不同的像素。**刮錯的話這個數字不會是 0**——它同時是刮掉的
那幾點原本就不該在、以及底下露出來的內容本來就對的證據。

## 規則（寫給下一次量規格的人）

量畫面規格的收據，**游標必須不在要量的那一塊上**；做不到就拿另一份游標
在別處的收據把那一塊補乾淨（`cmdscreen_spec.py` 已經這樣做）。
量完之後檢查一次：規格裡有沒有落單的小方塊或 1×3／3×1 的 `rules`，
座標剛好落在 `mouse` 的 13×16 外接矩形裡——那就是游標。
