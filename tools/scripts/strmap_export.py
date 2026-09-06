"""把 `STRMAP.BIT` 解成主畫面的星圖底圖。

版面是**用改記憶體問出來的**，不是猜的：把載入後某個位移起的 64 bytes
改掉，看畫面哪裡變色。

| 位移 | 變化的畫面區域 |
|---|---|
| 0 | x 0–15, y 40–47 |
| 160 | x 16–31 |
| 320 | x 32–47 |
| 2400 | x 240–255 |

每 160 bytes 前進 16 個像素，而 16×20 的格子正好是 160 bytes（4bpp），
20×15 格 × 160 ＝ 48,000 ＝ 檔案大小，而星圖視窗也正好是 20×15 格。
格內每一列 8 bytes ＝ 四個平面各 2 bytes 交錯（改位移 4–7 會把色 3 變成
色 F，也就是加上 bit2 與 bit3）。

用法：tools/py.sh tools/scripts/strmap_export.py [畫面收據名]
"""
import pathlib, sys
sys.path.insert(0, '/work/tools/scripts')
from pngout import write_png, game_palette

CW, CH = 16, 20            # 一格
COLS, ROWS = 20, 15        # 格數
W, H = CW * COLS, CH * ROWS
CELL = CW * CH // 2        # 160 bytes
name = sys.argv[1] if len(sys.argv) > 1 else 'main1'

bit = pathlib.Path('/orig/STRMAP.BIT').read_bytes()
scr = pathlib.Path('/work/workplace/oracle/%s/screen.bin' % name).read_bytes()
pal = game_palette('/work/workplace/oracle/%s/screen.png' % name)

def decode():
    px = bytearray(W * H)
    for cy in range(ROWS):
        for cx in range(COLS):
            base = (cy * COLS + cx) * CELL
            for y in range(CH):
                row = base + y * 8
                for p in range(4):
                    hi, lo = bit[row + p * 2], bit[row + p * 2 + 1]
                    word = hi << 8 | lo
                    for k in range(16):
                        if word >> (15 - k) & 1:
                            px[(cy * CH + y) * W + cx * CW + k] |= 1 << p
    return px

px = decode()
same = sum(1 for y in range(H) for x in range(W)
           if px[y * W + x] == scr[(40 + y) * 640 + x])
print('與 %s 畫面同色 %d / %d (%.1f%%)' % (name, same, W * H, same * 100 / (W * H)))
import os
os.makedirs('/work/workplace/assets', exist_ok=True)
write_png('/work/workplace/assets/strmap.png', W, H, px, pal)

# 只比 bit2／bit3（檔案裡的那兩個平面；bit0／bit1 是執行期算出來的勢力色）
hi_same = sum(1 for y in range(H) for x in range(W)
              if (px[y * W + x] >> 2) == (scr[(40 + y) * 640 + x] >> 2))
print('只比高兩個 bit：%d / %d (%.1f%%)' % (hi_same, W * H, hi_same * 100 / (W * H)))
