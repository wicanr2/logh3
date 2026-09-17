# STRMAP.BIT — 星域地圖底圖

證據等級：**confirmed**（版面來自原版 blit 常式；非透明像素與原版畫面完全相同）
素材：`assets/original/strmap/strmap.png`（進版控；
以 `tools/py.sh tools/scripts/strmap_export.py` 從自備原版重建；清單與雜湊見 [`assets/original/README.md`](../../assets/original/README.md)）
實作：`internal/asset/spr.go` 的 `DecodeSheet`

## 版面

48,000 bytes ＝ **300 塊 16 × 20**，row-major 排成 20 × 15，
拼起來是 **320 × 300 像素**，畫在畫面的 `(0,40)`。

每塊 160 bytes：每一列 4 個 plane 各一個 word（小端存放，位元高位在左），
與 `SPRITE.SPR` 同一種 planar 交錯，只是塊的尺寸不同。

## 內容

伊謝爾倫回廊兩側的完整星域圖：星系名（VALHALLA、ODIN、KASTROP、
GANDHARVA、ISERLOHN、TIAMAT、DORIA、BALLAT…）、航路連線、星球圖示、
左上角的帝國鷲徽。

## 勢力著色是就地改資料

檔案裡的色號 **0 是留白**；遊戲把整個檔案載進記憶體之後，**就地把那些
0 塗成勢力色**（實測改了 13,420 / 48,000 bytes），再整份畫上去。
地圖在同一個畫面裡被畫了兩趟，兩趟的來源與目的地都相同——
中間差的就是這次改寫。

證據：把記憶體裡那 48,000 bytes 倒出來（`-dump-mem lin:6D8C4:48000:…`）
解碼，貼到畫面 `(0,40)`，**0 / 42,000 不同**（比對未被視窗遮住的
`x < 140`）。直接用檔案原始內容則有 9,980 個像素不同，而且**全部**是
「檔案是 0、畫面是 1 / 3 / 6」。

### 著色怎麼算的（confirmed）

原版 `GIN3PS.EXE` `2376:0BDC`：

```
0BDC  mov ax,[bp+00]      ; bp ＝ 星域形狀遮罩
0BDF  not ax
0BE1  and [si],ax         ; 先把工作副本上遮罩涵蓋的位元清掉（四個 plane）
0BEC  mov ax,es:[di]      ; ES ＝ EMS 頁框，放的是**原始的 STRMAP.BIT**
0BEF  or  ax,es:[di+2]    ; 四個 plane 相 OR ＝「這裡原本有沒有東西」
0BF3  or  ax,es:[di+4]
0BF7  or  ax,es:[di+6]
0BFB  not ax
0BFD  and ax,[bp+00]      ; ＝ 遮罩內、原圖是空白的那些位元
0C02  mov bx,ax
0C0B  mov ax,es:[di]      ; 再取一次原圖的 plane
0C10  test dl,1           ; 顏色的第 0 個位元
0C15  or  ax,bx           ; 有就把要塗的位元併進這個 plane
0C17  or  [si],ax         ; 寫回工作副本
```

也就是**在遮罩範圍內，只塗原圖留白的地方**；航路、星系名、星球圖示
原封不動。原始地圖留在 EMS 頁框當底本，塗色的結果寫進主記憶體的工作
副本——所以工作副本可以反覆重塗而不會越塗越髒。

remake 端是 `asset.PaintRegion`（在色號陣列上就是同一個條件）。

星域形狀遮罩（`[bp+00]`）來自 `STRMAP.RGN`；顏色（`[bp+12]`）取自
`ds:F162` 的 32 個 bytes，那份表由 `3705:037F` 逐格填：`STELLAR.DT*` 每一筆的
`+0x12`（擁有者）查「擁有者 → 色號」表——**沒有自軍艦隊**（`+0x28 ＋ 看地圖
的陣營` 是 0）查 `ds:8BD2` {1, 12, 3}，**有**就查 `ds:8BD6` {0, 2, 15}
（帝國自己的星域駐了艦隊就變黑）；反白的那一格（`ds:F182`，開局是自己的
首都，`3705:0284`）查 `ds:8BDA` {6, 5, 7}（`docs/re/17`、`docs/re/313` §3.2；
remake `rules.StrategicMapColour`）。
開局讀的是 **`DTX`**；`MAP*.CMP` 走到主畫面為止一份都沒被讀。

搬運本身是逐星域的：`3CA9:0002` 的 `memmove` 被呼叫 32 次，
呼叫端 frame 的前四個 word 正好是 `STRMAP.RGN` 表的
`格x, 格y, 寬(格), 高(格)`，來源段是 EMS 頁框、目的是工作副本。

回歸測試 `TestStrmapSheet` 釘的是「檔案非 0 的像素必須與畫面相同」——
一旦出現「地圖非 0 卻和畫面不同」，就是圖塊排列或位元序解錯了。

## 怎麼找到的

`-vram-sites` 差分找到畫地圖的 blit（`2376:0D3C`），再用 `-regs-at` 記下
它執行時的暫存器：`DS:SI = 6D8C:0004`、每列 `SI += 8`、`CX` 從 20 遞減、
`DI` 每列 `+80`——圖塊尺寸與來源緩衝同時確定。把那段記憶體的位元組
拿去比對原版檔案，一次命中 `STRMAP.BIT` 的 offset 0。

**監看寫入只看得到目的地，看不到它從哪裡搬**；先前直接拿畫面像素去
`.BLK`／`.CMP` 裡搜，全部落空。

「地圖被畫了兩趟、兩趟資料卻不同」是靠 `-vram-at`（只統計寫到某一個
VRAM 位移的指令）發現的：盯住一個「檔案說 0、畫面是 1」的像素，
看到的最後一個寫入者就是地圖自己的 blit——所以不是別人蓋上去的，
是地圖資料本身在兩趟之間變了。
