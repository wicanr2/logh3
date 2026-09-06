"""`COM_FLAG.SPR` 的版面：每 8 bytes 是一列 8 像素？

零 byte 全部落在每 8 bytes 的前四個位置（1,016 個零裡沒有一個在後四個），
所以版面是 **4 bytes ＋ 4 bytes** ——很像「遮罩 ＋ 四個平面的資料」，
遮罩全 0 表示那一列完全不透明。1,440 / 8 ＝ 180 列。

這支把它按幾種高度切開輸出，用來判斷一張圖幾列。

用法：tools/py.sh tools/scripts/com_flag_probe.py
"""
import pathlib, sys
sys.path.insert(0, '/work/tools/scripts')
from pngout import write_png, game_palette

d = pathlib.Path('/orig/COM_FLAG.SPR').read_bytes()
pal = game_palette('/work/workplace/oracle/main1e/screen.png')
rows = len(d) // 8
px = bytearray(8 * rows)
for r in range(rows):
    m = d[r * 8:r * 8 + 4]
    v = d[r * 8 + 4:r * 8 + 8]
    for k in range(8):
        c = 0
        for p in range(4):
            if v[p] >> (7 - k) & 1:
                c |= 1 << p
        px[r * 8 + k] = c
import os
os.makedirs('/work/workplace/assets', exist_ok=True)
write_png('/work/workplace/assets/com_flag-8x%d.png' % rows, 8, rows, px, pal)
print('8 × %d 寫出' % rows)
# 一張圖寬 8×g、高 h，資料按列優先：一列 g 組，走完 h 列換下一張。
for g, h in ((2, 15), (3, 15), (4, 15), (2, 20), (3, 20), (2, 45), (3, 12)):
    per = g * h
    if rows % per:
        continue
    n = rows // per
    w, W = 8 * g, 8 * g * n
    out = bytearray(W * h)
    for i in range(n):
        for y in range(h):
            for gx in range(g):
                src = (i * per + y * g + gx) * 8
                for k in range(8):
                    out[y * W + i * w + gx * 8 + k] = px[src + k]
    # 放大四倍再輸出：8×15 那種尺寸縮圖看不出是不是旗幟。
    Z = 4
    big = bytearray(W * Z * h * Z)
    for y in range(h * Z):
        for x in range(W * Z):
            big[y * W * Z + x] = out[(y // Z) * W + x // Z]
    name = '/work/workplace/assets/com_flag-%dx%d-x%d.png' % (w, h, n)
    write_png(name, W * Z, h * Z, big, pal)
    print('  %d 張 %d×%d' % (n, w, h))
