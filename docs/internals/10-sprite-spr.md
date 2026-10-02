# SPRITE.SPR — 16 × 16 圖塊庫（艦隊組成的格盤用它）

證據等級：**強證據**（版面來自原版 blit 常式；圖形可辨識且完整對齊）
素材：`assets/original/sprite/spr-NN.png`（75 張，進版控；
以 `tools/py.sh tools/scripts/spr_export.py 16 SPRITE.SPR` 重建；清單與雜湊見 [`assets/original/README.md`](../../assets/original/README.md)）
實作：`internal/asset/spr.go`

## 版面

9600 bytes ＝ **75 張 16 × 16**，未壓縮，沒有檔頭也沒有索引。

每張 128 bytes：**每一列依序放 plane 0–3 各一個 word**（小端存放，
畫面上的位元高位在左），走完一列再往下 16 列。

版面來自原版 `GIN3PS.EXE` `2376:0D25` 的 blit：

```
0D35  B0 02 EE        mov al,02 / out dx,al   ; 序列器 index 2 ＝ map mask
0D39  B0 01 EE        mov al,01 / out dx,al   ; plane 0
0D3C  AD              lodsw
0D3D  26 89 05        mov es:[di],ax
0D40  B0 02 EE AD     …plane 1
0D47  B0 04 EE AD     …plane 2
0D4E  B0 08 EE AD     …plane 3
0D55  83 C7 50        add di,0050             ; 下一列（80 ＝ mode 12h 每列位元組）
0D58  E2 DF           loop
```

## 內容

同盟（藍底）與帝國（紅底）各一組：數字 1–16 與字母 B、H、C、A、D、T、E、P，
另有金色三角形指標與幾個小圖示。這是地圖上標艦隊用的。

## 尺寸怎麼定的

**整除證明不了尺寸**：9600 同樣是 16 × 20（160 bytes）的整數倍，
按 20 列解出來是 60 張錯位的圖——數字方塊被切在張與張的邊界上，
但每一張仍然「看起來像有東西」。判準是圖形本身要完整對齊，
加上 blit 常式那一列 80 bytes 的步進。

## 未確認

- 早期796年1月開局收據沒有艦隊圖示；後續戰場用途由研究85及376補證。
  指定方向圖示已驗，完整圖塊庫與所有畫面不外推。
- `COM_FLAG.SPR`（1440 bytes）**不是**這個版面，照 16 × 16 解出來是雜訊。
  `3DBLOCK.SPR` 則是同一套，只是尺寸不同（48 × 64，見 `docs/internals/21`）。

## 戰場方向圖示，第36輪續證

第0..7塊為八向方向圖示，proven。3982:0001以索引左移7取得128-byte圖形，
原版入口RAM的1024 bytes與SPRITE.SPR前八塊相同。25DB:04E7控制Fleet+6A的方向／隱藏9，
0670讀方向位移表。來源、原始bytes、工具與分級在[方向索引](data/battle-directions.json)及[研究376](../re/376-second-battle-render.md)。
正式renderer依[084](../spec/084-battle-fleet-directions.md)保存透明與裁切，S1的43幀共18920個
方向像素相同；完整畫面仍有其他差異，不能當成整個圖塊庫或所有畫面驗收。

## 已知用途：艦隊組成的格盤（confirmed）

艦隊組成畫面中間那塊 9 × 9 的格盤（(136,168)–(279,311)，一格 16 × 16）
就是從這裡取圖塊：**空位是第 44 張、中央那一格是第 36 張**，兩張都與
收據 `fok` 逐像素相同（`docs/re/67`）。

格盤是棋盤狀的：`(欄 + 列)` 為偶數的 41 格畫圖塊，其餘留底色 3。

比對的時候要注意**位元組順序**：檔案裡一列一個 plane 是「第一個 byte
是最左邊 8 個像素」，也就是把 word 當大端看。照字面讀成小端會完全對不上
（`internal/asset.DecodeTiles` 讀出 word 之後有一次 `v<<8 | v>>8`，
做的就是這件事）。
