# 文字繪製：粗體、陰影、底線與網點

畫面上的中文有好幾種樣子——細的、粗的、帶淺色描邊的、灰掉的。
字型檔只有一份（`GRAPH.IMG`，見 [`06`](06-font-graph-img.md)），
差別全部是畫的時候疊出來的。

機器可讀版本在 [`data/text-rendering.json`](data/text-rendering.json)。
證據等級：繪製流程與加粗／陰影／網點／底線 **confirmed**（反組譯 ＋ 十三張
畫面逐位元組對拍）；樣式旗標 bit 3／4／5／7 的效果 **未解**。
調查過程見 [`docs/re/18`](../re/18-settings-screen-font.md)。

## 一個字怎麼畫出來（confirmed）

遊戲先把 16 × 15 的字形讀進一個 **20 列的緩衝區**（`ds:829C`，每列一個
word，高位在前、bit15 在最左，與字型檔同一種排法），照樣式旗標就地改那
份點陣，再分兩趟送進 VGA：

| 步驟 | 做什麼 | 顏色 |
|---|---|---|
| 1 | 底色填滿整格（20 列 × 16 像素）| 屬性位元組的高四位 |
| 2 | 樣式變換（加粗／網點）就地改緩衝區 | — |
| 3 | 陰影趟：**第 r 列畫緩衝區第 r + 1 列、整列往左移一格**，共 19 列 | 3 或 7 |
| 4 | 本體趟：第 r 列畫緩衝區第 r 列，共 20 列 | 屬性位元組的低四位 |
| 5 | 底線（可選）| 15 |

第 3 步的位移是**左上各一格**：畫面上 (x, y) 這一點只要字形的
(x + 1, y + 1) 有筆畫就上陰影色。兩趟疊起來的效果，是筆畫的左上緣多一道
淺色邊。

第 3 步只有 19 列而第 4 步有 20 列，因為陰影列往上挪了一格——
緩衝區第 0 列的陰影會落到格子外面，直接不畫。

### 顏色由一個屬性位元組決定（confirmed）

`cl` 是屬性位元組：**低四位是字身色、高四位是底色**。
陰影色不在裡面，而是看樣式旗標 `ch` 的 bit 6：設了用 **7**，沒設用 **3**。

## 樣式旗標 `ch`（confirmed 三項、其餘未解）

| 位元 | 效果 | 怎麼做的 |
|---|---|---|
| 0 | 底線 | 格子最後一列整條填色 15（`mov word es:[di-0x50], FFFF`）|
| 1 | 網點（灰掉）| 緩衝區逐列 `and` `5555`／`AAAA` 交錯 |
| 2 | **加粗** | 緩衝區逐列 `v |= v >> 1`，往右加粗一格 |
| 3, 4, 5, 7 | 描邊之類的疊畫 | 另一組疊畫常式，**未解** |
| 6 | 陰影用色 7（否則 3）| 見上 |

加粗要在陰影趟之前做——陰影與本體都用加粗後的點陣。順序顛倒會多出一排
像素。

## 半形字兩個一組（confirmed）

緩衝區是 16 像素寬，半形字**兩個一組**塞進同一格。陰影那一趟是整格
16 位元一起往左移一格，所以**右半那個字最左邊那一行的陰影會落到左半的
地盤上**（畫面座標 x − 1），左半那個字移出格外的那一行才會被丟掉。

判準是這個字的 x 在不在 16 的邊界上：`x % 16 == 8` 的那個要多畫一行。
全形字佔滿整格，最左邊那一行的陰影一律丟掉。

> 量到過一次：取進度畫面的「──/──/── ──:──:──」少畫 48 個像素，
> 前面十六張對過的畫面剛好都沒有踩到。

## 硬體那一側只負責上色（confirmed）

兩趟都是在 **VGA write mode 3** 下寫的，所以 CPU 寫出去的位元組就是
位元遮罩，遮到的像素填 Set/Reset 的顏色、其餘保留 latch。字形的形狀
不會被硬體改。

底色那一趟用的是 **write mode 2**（CPU 位元組的低四位直接是顏色）。

兩段設定圖形控制器的常式：

```
底色（write mode 2）
BA CE 03  mov dx,3CE
B8 05 0A  mov ax,0A05     ; index 5 Mode ＝ 0x0A（write mode 2、read mode 1）
EF        out dx,ax
B8 03 00  mov ax,0003     ; index 3 Function ＝ replace
EF
B8 07 00  mov ax,0007     ; index 7 Color Don't Care ＝ 0
EF
CB

字身（write mode 3）
BA CE 03  mov dx,3CE
B8 05 0B  mov ax,0B05     ; index 5 Mode ＝ 0x0B（write mode 3、read mode 1）
EF
B8 01 0F  mov ax,0F01     ; index 1 Enable Set/Reset ＝ 0x0F
EF
8A E1     mov ah,cl       ; index 0 Set/Reset ＝ 顏色
32 C0     xor al,al
EF
B8 03 00  mov ax,0003
EF
B8 08 FF  mov ax,FF08     ; index 8 Bit Mask ＝ FF
EF
B8 07 00  mov ax,0007
EF
CB
```

**Read mode 1 加上 Color Don't Care ＝ 0，讀 VRAM 一律回 `FF`**。
所以字身那兩趟寫的是 `and es:[di], al`——讀回來的 `FF` 跟 `al` 相與
還是 `al`，遮罩就等於那一列的點陣。`and` 只是順手把 latch 讀進來，
沒有把已經畫好的東西當成輸入。

## 位元組簽章

`docs/internals/00-README.md` 說明為什麼這裡不寫線性位址。

| 片段 | 簽章 |
|---|---|
| 陰影色選 3 或 7 | `F6 C5 40 75 04 B1 03 EB 02 B1 07` |
| 陰影趟（左上一格）| `BE 9E 82 51 B9 13 00 AD 86 E0 D1 E0 86 E0 26 20 05 26 20 65 01 83 C7 50 E2 ED` |
| 本體趟 | `BE 9C 82 81 EF F0 05 51 B9 14 00 AD 26 20 05 26 20 65 01 83 C7 50 E2 F3` |
| 底色趟 | `B8 08 FF EF 8A C1 C0 E8 04 B9 14 00 26 88 05 26 88 45 01 83 C7 50 E2 F4` |
| 樣式分派（bit 2、bit 1）| `F6 C5 04 74 03 E8 09 00 F6 C5 02 74 03 E8 18 00 C3` |
| 加粗 | `BE 9C 82 B9 14 00 8B 04 86 E0 D1 E8 86 E0 09 04 46 46 E2 F2` |
| 網點 | `BE 9C 82 B9 14 00 B8 55 55 21 04 46 46 F7 D0 E2 F8` |
| 底線 | `F6 C5 01 74 0D B1 0F 9A ?? ?? ?? ?? 26 C7 45 B0 FF FF` |
| 設 write mode 2 | `BA CE 03 B8 05 0A EF B8 03 00 EF B8 07 00 EF CB` |
| 設 write mode 3 | `BA CE 03 B8 05 0B EF B8 01 0F EF 8A E1 32 C0 EF` |

`ds:829C` 是那個 20 列的字形緩衝區；`829E` 是它的第 1 列。

## remake 怎麼用

`internal/ui.DrawGlyph`／`DrawHalfGlyph` 的 `bold` 參數就是旗標 bit 2，
`shadow`／`shadowDY` 就是第 3 步。畫面規格（`docs/internals/data/screen-*.json`）
的每一行文字帶一個 `bold` 欄位。

反查字碼的時候要記得：粗體的畫面點陣**不等於**字型檔裡的點陣，
要拿加粗後的點陣當鍵（`tools/scripts/screenspec.py` 的 `font_lookup(bold=True)`）。
