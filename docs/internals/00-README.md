# 原版內部構造參考

《銀河英雄傳說III SP》PC DOS 中文版（1994，亨和企業．微波軟體發行）的
結構化逆向資料。這裡放**目前為止查證出來的事實本身**；查證過程與逐次
調查紀錄在 `docs/re/`，執行器規格在 dosgolem 的 `docs/spec/`。

這個專案除了 remake 之外，還有一個目的是保存這款中文版的技術細節——
1994 年台灣中文化作品的封裝方式、文字編碼取捨、資料格式與製作團隊，
在原始碼與開發資料都不存在的情況下，只剩執行檔本身可以問。所以逆向出來
的東西一律整理成可讀的 markdown，不停留在對話紀錄或一次性的輸出檔裡。

## 這裡的東西怎麼寫

音色庫來源與完整授權索引：[`player-music-library.json`](data/player-music-library.json)，
用途與原版素材邊界見 [`17-mds.md`](17-mds.md)。
原生 GM 移植與真實轉檔驗收索引：[`native-gm-import.json`](data/native-gm-import.json)，
方法及限制見[驗證 030](../playtest/030-native-gm-import.md)。
原生 FM 的固定版本／曲目／封裝索引：[`fm-import-lock.json`](data/fm-import-lock.json)，
原版觸發與重生限制見[規格 071](../spec/071-native-fm-import.md)、[研究 359](../re/359-native-fm-import.md)。
完整原生曲庫匯入與三路徑重跑收據：[`native-fm-import.json`](data/native-fm-import.json)，
正式解碼、重生命令與平台限制見[驗證 031](../playtest/031-native-fm-import.md)。
remake 增強策略的正式入口驗證索引：[`enhanced-ai-player-path.json`](data/enhanced-ai-player-path.json)，
範圍及原版差異見[規格 068](../spec/068-enhanced-ai-player-path.md)。

**每一項都要標推論等級**：`confirmed`（有實測或反組譯位元組證據）、
`強證據`、`假說`、`未知`。推測不得因為「看起來很合理」就升格，
升格要補證據並註明來源。

**什麼是穩定的、什麼不是**——這決定資料該怎麼記：

| 穩定 | 不穩定 |
|---|---|
| 結構欄位的**位移**（`rec[+0x0B]`）| **段值**（`6B93:0068`）|
| 資料段內的**變數位移**（`ds:94E0`）| 執行期的**線性位址** |
| **位元組簽章**（一段沒有絕對位址的機器碼）| 步數（`#25757331`）|
| 演算法、編碼、檔案格式 | 暫存器快照 |

段值取決於程式被載到哪裡，而那取決於 DOS 的記憶體佈局。實例：修好
dosgolem 的 EXEC 記憶體回收之後，`GIN3PS.EXE` 的 PSP 從 `0x3922` 降到
`0x0339`，所有模組跟著搬家，`docs/re/05` 之前記下的段值全部失效
（結構位移與演算法則完全不受影響）。

**所以定位程式碼一律用位元組簽章。** dosgolem 的 probe 可以一次驗一批：

```
tools/go.sh run ./cmd/probe -exe /orig/GIN3.COM -root /orig ... \
  -find "26 8A 47 0B 30 E4 26 8A 4F 0C 30 ED F7 E1, 83 3E E6 94 FF 75 1A"
```

## 產出格式

**每一項解出來的東西都要有機器可讀版本。** markdown 給人看，
`data/*.json` 給後續的工具與 remake 用；兩邊由同一份證據產生，不得只更新
一邊。素材（字型、圖庫、地圖）解出來之後匯出成 **PNG** 放
`workplace/assets/`——那是原版素材，**不進版控**，但要讓下一輪能直接拿來用。

挑簽章時避開含絕對位址的位元組（`9A xx xx ss ss` 的 far call 目標、
`A1 xx xx` 之類指到別段的存取），那些會隨載入位址改變。

## 工具

| 包裝器 | 用途 |
|---|---|
| `tools/py.sh` | Python 分析（`python:3.12-slim`）。原版素材唯讀掛在 `/orig`，倚天字型之類的外部參照用 `LOGH3_EXTRA_RO` 掛在 `/ref`。|
| `tools/ida.sh` | IDA Pro 9.4 headless。`analyze <exe>` 建 `.i64`，`script <exe> <py>` 跑 IDAPython。**exit code 不能當證據，只看輸出檔**；腳本一律寫檔，一支腳本一次處理多個目標（對同一個 `.i64` 連續開關會弄壞它）。|

`GIN3PS.EXE` 是六層封裝（GLUE／TINYPROG／PKLITE 系，見 `02`），IDA 只看得到最外層的解壓殼——本體要在執行期逆向
（dosgolem 的 `-peek`／`-find`／trace／檔案存取記錄）。`OPEN.EXE` 與
`ENDING.EXE` 沒有封裝，適合 IDA。

## 目錄

| 檔 | 內容 |
|---|---|
| [`01-assets-inventory.md`](01-assets-inventory.md) | 原版素材清單 |
| [`02-executables-and-credits.md`](02-executables-and-credits.md) | 執行檔、封裝器與製作團隊 |
| [`03-startup-and-memory.md`](03-startup-and-memory.md) | 啟動鏈、記憶體佈局與堆積配置器 |
| [`04-text-windows-and-big5.md`](04-text-windows-and-big5.md) | 文字視窗、儲存格格式與中文編碼 |
| [`05-palette-and-fade.md`](05-palette-and-fade.md) | 調色盤與淡入淡出 |
| [`06-font-graph-img.md`](06-font-graph-img.md) | 中文字型：`GRAPH.IMG` |
| [`07-window-frame.md`](07-window-frame.md) | 視窗外框 |
| [`08-admiral-dt.md`](08-admiral-dt.md) | ADMIRAL.DT[89AX] — 提督能力表 |
| [`09-face-amg.md`](09-face-amg.md) | FACE.AMG — 人物頭像（.AMG 圖庫格式） |
| [`10-sprite-spr.md`](10-sprite-spr.md) | SPRITE.SPR — 艦隊編號圖塊 |
| [`11-copy-protection-screen.md`](11-copy-protection-screen.md) | 防拷問答畫面 |
| [`12-strmap-bit.md`](12-strmap-bit.md) | STRMAP.BIT — 星域地圖底圖 |
| [`13-strmap-rgn.md`](13-strmap-rgn.md) | STRMAP.RGN — 星域區域遮罩 |
| [`14-screen-strings.md`](14-screen-strings.md) | 畫面文字表與熱區 |
| [`15-keireki-dat.md`](15-keireki-dat.md) | KEIREKI.DAT — 人物傳 |
| [`16-ending-exe.md`](16-ending-exe.md) | ENDING.EXE — 結局文字與職員表 |
| [`17-mds.md`](17-mds.md) | `.MDS` — 樂曲 |
| [`18-vo5.md`](18-vo5.md) | `.VO5` — 結局圖 |
| [`19-ibm16-350.md`](19-ibm16-350.md) | `IBM16.350` ＋ `IBMFLS16.350` — 開頭動畫 |
| [`20-comani-bmg.md`](20-comani-bmg.md) | `COMANI.BMG` — 戰鬥動畫封存檔 |
| [`21-3dblock-spr.md`](21-3dblock-spr.md) | `3DBLOCK.SPR` — 等距方塊 |
| [`22-stellar-dt.md`](22-stellar-dt.md) | `STELLAR.DT[89AX]` — 星域資料 |
| [`23-hotspot-lists.md`](23-hotspot-lists.md) | 熱區清單與按鈕標籤 |
| [`24-window-list-and-right-button.md`](24-window-list-and-right-button.md) | 視窗串列、命中測試與右鍵 |
| [`25-text-rendering.md`](25-text-rendering.md) | 文字繪製：粗體、陰影、底線與網點 |
| [`26-star-region-graph.md`](26-star-region-graph.md) | 星域之間的相鄰關係與路徑表 |
| [`27-treasury-globals.md`](27-treasury-globals.md) | 財政的全域變數 |
| [`28-screen-specs.md`](28-screen-specs.md) | 畫面規格 JSON：`data/screen-*.json` |
| [`29-planets-and-economy.md`](29-planets-and-economy.md) | `PLADAT.DAT` 與經濟力 |
| [`30-save-format.md`](30-save-format.md) | 存檔：`*.DT[89A]` 就是執行期的表 |
| [`31-text-plane.md`](31-text-plane.md) | 全螢幕字元平面（`ds:D81E`）|
| [`32-ship-build-costs.md`](32-ship-build-costs.md) | 船艦建造：單價（兩個陣營不同）、一單位幾艘與捨去規則 |
| [`data/`](data/) | 機器可讀版本（JSON），給後續工具與 remake 用 |

- [`34-battle-screen.md`](34-battle-screen.md)：戰術（戰鬥）畫面的版面、命令選單與怎麼走到。
- [`35-block3-blk.md`](35-block3-blk.md)：戰場的圖塊集（128 塊 16×20）。
- [`36-map-cmp.md`](36-map-cmp.md)：每個星域的戰場地圖（128×102 格，兩條規則的編碼）。
- [`37-ani-pnl-a98.md`](37-ani-pnl-a98.md)：戰鬥畫面的底圖（`.AMG` 編碼、畫在 y=40）。
- [`38-battle-units.md`](38-battle-units.md)：戰場上的單位是「分隊」，不是艦隊。
- [`39-battle-entry-points.md`](39-battle-entry-points.md)：戰場入口表。
- [`40-ship-kind-table.md`](40-ship-kind-table.md)：艦種屬性表。
- [`data/live-fleet-save.json`](data/live-fleet-save.json)：現況艦隊存讀、速度寫入的分級語意與原版／GUI 雜湊；入口為[規格 067](../spec/067-live-fleet-save.md)。
- [`41-sin-table.md`](41-sin-table.md)：原版的正弦表。
- [`42-screen-tile-table.md`](42-screen-tile-table.md)：畫面的圖塊表 `ds:D81E`。
- [`43-battle-terrain-and-rally.md`](43-battle-terrain-and-rally.md)：戰場地形的通行規則與集結點。
- [`44-battle-command-menus.md`](44-battle-command-menus.md)：戰場的命令選單（五張表、可選旗標、版面公式）。
- [`45-battle-unit-panel.md`](45-battle-unit-panel.md)：戰場右側的分隊面板（八列、值的來源、三個顏色規則）。
- [`46-battle-formation-chooser.md`](46-battle-formation-chooser.md)：戰場「編隊」的選格子畫面。
- [`47-window-frame-tiles.md`](47-window-frame-tiles.md)：視窗外框的圖塊（`FRAME.BLK`）。
- [`48-message-inventory.md`](48-message-inventory.md)：訊息框說了什麼。
- [`49-text-engine.md`](49-text-engine.md)：印字進視窗。
- [`50-fortress-gun.md`](50-fortress-gun.md)：要塞砲的射程、瞄準與開火。
- [`51-borland-long-helpers.md`](51-borland-long-helpers.md)：Borland 的 32 位元輔助函式（`3CC9` 那一段）。
- [`52-plot-primitives-and-canvas.md`](52-plot-primitives-and-canvas.md)：畫點的三支常式與那 40 列的畫布偏移。
- [`53-strategic-planet-assault.md`](53-strategic-planet-assault.md)：戰略回合裡的打行星（`20B4:1512` → `20B4:1662`）。
- [`54-strategic-ai-decision.md`](54-strategic-ai-decision.md)：戰略 AI 決定一支艦隊要不要動。
- [`56-fleet-composition-panel.md`](56-fleet-composition-panel.md)：艦隊組成畫面的工作區（編隊那一行的字母與格盤的圖塊）。
- [`57-phezzan-event.md`](57-phezzan-event.md)：費沙事件——第一次易主時觸發、經濟力流向對手首都、邊境叛離、人類座位受罰。
- [`58-strategic-fleet-window.md`](58-strategic-fleet-window.md)：戰略地圖上的艦隊視窗與命令（清單點列 → 作戰 → 目的地 → 確認）。
- [`59-list-screens.md`](59-list-screens.md)：一覽表四張（星系／艦隊／提督／敵將資料）：點列開什麼、游標反白、位移與分頁的變數。
- [`60-mouse-roles.md`](60-mouse-roles.md)：滑鼠在每張畫面的作用總表（左鍵、右鍵、按住、游標經過），逐格掛證據；原版有、remake 沒有的列在最後。
- [`61-turn-transition-windows.md`](61-turn-transition-windows.md)：回合處理中的畫面（「戰略模式處理中」、清單清空、地圖不看座位）與每支艦隊走一格的過場視窗（出發前往、被擋、經費不足、問話）、等待怎麼結束；報告框（拿下行星、士氣、已全軍覆沒）。
- [`62-strategic-attrition.md`](62-strategic-attrition.md)：戰略階段士氣低落停在別人星域的艦隊每回合扣兵、第 0 組 0 艘就全滅（報告、艦艇作廢、提督失蹤、星域與陣營艦隊數減一），沒全滅就移掉 0 艘的組。
- [`63-info-screen-tabs.md`](63-info-screen-tabs.md)：情報收集的三個分頁（治安／勢力／敵艦）與「確認」。
- [`64-ship-pool.md`](64-ship-pool.md)：艦艇池一筆 78 bytes 的欄位、配置與作廢（分隊就是它）。
- [回合後距離／座標語意索引](data/post-turn-runtime-fields.json)：研究 354 的原始定位與證據分級，由 IDA 匯出自動合併；[全表收據摘要](data/post-turn-runtime-receipts.json) 的驗收入口是[對拍 027](../playtest/027-post-turn-runtime-fields.md)。
