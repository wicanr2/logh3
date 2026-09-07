# 原版內部構造參考

《銀河英雄傳說III SP》PC DOS 中文版（1994，亨和企業．微波軟體發行）的
結構化逆向資料。這裡放**目前為止查證出來的事實本身**；查證過程與逐次
調查紀錄在 `docs/re/`，執行器規格在 dosgolem 的 `docs/spec/`。

這個專案除了 remake 之外，還有一個目的是保存這款中文版的技術細節——
1994 年台灣中文化作品的封裝方式、文字編碼取捨、資料格式與製作團隊，
在原始碼與開發資料都不存在的情況下，只剩執行檔本身可以問。所以逆向出來
的東西一律整理成可讀的 markdown，不停留在對話紀錄或一次性的輸出檔裡。

## 這裡的東西怎麼寫

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

`GIN3PS.EXE` 是 GLUE 封裝的，IDA 只看得到解壓殼——本體要在執行期逆向
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
| [`data/`](data/) | 機器可讀版本（JSON），給後續工具與 remake 用 |
