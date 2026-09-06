# STRMAP.BIT — 星域地圖底圖

證據等級：**confirmed**（版面來自原版 blit 常式；非透明像素與原版畫面完全相同）
素材：`workplace/assets/strmap.png`（不進 Git；
以 `tools/py.sh tools/scripts/strmap_export.py` 從自備原版重建）
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

## 透明色

地圖裡的色號 **0 是留白**，底下還有一層著色：把地圖貼到原版畫面的
`(0,40)`，**地圖非 0 的像素與原版畫面完全相同**，所有差異都是
「地圖是 0、畫面是 1 / 3 / 6」。那層著色多半是各勢力的星域範圍，
來源推測是 `MAP*.CMP`（32 份、大小各異）——**未確認**。

回歸測試 `TestStrmapSheet` 釘的就是這件事：一旦出現「地圖非 0 卻和畫面
不同」，就表示圖塊排列或位元序解錯了。

## 怎麼找到的

`-vram-sites` 差分找到畫地圖的 blit（`2376:0D3C`），再用 `-regs-at` 記下
它執行時的暫存器：`DS:SI = 6D8C:0004`、每列 `SI += 8`、`CX` 從 20 遞減、
`DI` 每列 `+80`——圖塊尺寸與來源緩衝同時確定。把那段記憶體的位元組
拿去比對原版檔案，一次命中 `STRMAP.BIT` 的 offset 0。

**監看寫入只看得到目的地，看不到它從哪裡搬**；先前直接拿畫面像素去
`.BLK`／`.CMP` 裡搜，全部落空。
