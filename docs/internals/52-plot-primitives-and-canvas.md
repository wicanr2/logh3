# 52：畫點的三支常式與那 40 列的畫布偏移

證據等級：**confirmed**（反組譯位元組 ＋ 收據 `s3fight` 逐點驗算，
`docs/re/154`）。

## 邏輯畫布是 640×400，實體畫面是 640×480

所有經過 `3C5B:00B2` 的繪圖都吃**邏輯座標**，實體位置要再往下 40 列：

```
3C5B:00B2(bx = x, cx = y)
  ax = cx << 4          ; y × 16
  cx = ax
  ax = ax << 2 + cx     ; y × 80
  bx = (bx >> 3) + ax + 0C80h
  retf                  ; bx ＝ framebuffer 位移（一列 80 位元組）
```

`0x0C80` ＝ 3200 ＝ 40 × 80 ＝ 第 40 條掃描線。畫面圖塊表（
`docs/internals/42`）用同一個常數當起點，底圖 `ANI_PNL0.A98` 也畫在
y ＝ 40（`docs/internals/37`），滑鼠是 `screen_y = mouse_y + 40`
（`docs/re/52`）。

換算時要小心 `docs/` 裡的座標寫的是哪一套：

| 講法 | 例子 |
|---|---|
| 邏輯（`3C5B:00B2` 的參數）| 戰場左上角 (16, 20)（`3870:0296` 算出來的，`docs/re/135`）|
| 實體（收據 PNG 上量的）| 戰場左上角 (16, 60)（`docs/internals/34`）|

## `39D7:08F4(x, y, 色)`：畫一個點

```
36AD:013D()                 ; 進臨界區
di = 3C5B:00B2(x, y)
2376:1810(ah = 色)          ; 設 EGA 暫存器
out 3CE, (80h >> (x & 7)):08 ; index 8 Bit Mask ← 只開這一位
es = 0A000h
or es:[di], al              ; al ＝ 0：讀取載入 latch，寫入走 Set/Reset
2376:1852()                 ; 還原 EGA 暫存器
36AD:0117()                 ; 離開臨界區
```

`or es:[di], al` 的 `al` 是 0——真正決定顏色的是 Set/Reset，寫入的資料
本身無關。這是 EGA 寫入模式 0 的標準寫法。

## `2376:1810(ah = 色)`：把寫入設成 XOR

| index | 值 | 意思 |
|---|---|---|
| 5 | `0x00` | Mode ← 寫入模式 0、讀取模式 0 |
| 0 | 色 | Set/Reset ← 顏色 |
| 1 | `0x0F` | Enable Set/Reset ← 四個平面都取 Set/Reset |
| 3 | `0x18` | Data Rotate/Function：rotate 0、function ＝ 11b ＝ **XOR** |

所以每一個經過這一支的像素都是「底下的值 XOR 顏色」，**同一點同一個
顏色寫兩次會還原**。原版用這個性質做可擦除的疊圖：小地圖上的分隊點
（`docs/internals/38`）、要塞砲的射程橢圓（`docs/internals/50`）都是。

`2376:1852` 是收尾：sequencer Map Mask ← `0x0F`，graphics index 0 ← 0
等，把暫存器放回預設。

## 直線走同一條路

折線／虛線的 `39D7:0936` 也是 `di = 3C5B:00B2(x0, y0)` 起手，走
Bresenham，另外用 `ds:8F0E` 當 8 位元的虛線圖樣，每畫一點 `ror` 一位
（`docs/re/135`）。
