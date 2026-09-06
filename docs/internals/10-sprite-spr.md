# SPRITE.SPR — 艦隊編號圖塊

證據等級：**強證據**（版面來自原版 blit 常式；圖形可辨識且完整對齊）
素材：`workplace/assets/sprite/spr-NN.png`（75 張，不進 Git；
以 `tools/py.sh tools/scripts/spr_export.py 16 SPRITE.SPR` 重建）
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

- 尚未與原版畫面逐像素對上：這些圖示要地圖上有艦隊才會畫出來，
  目前的收據（796 年 1 月開局）畫面上還沒有。
- `3DBLOCK.SPR`（1536 bytes）與 `COM_FLAG.SPR`（1440 bytes）**不是**這個版面，
  照 16 × 16 解出來是雜訊。
