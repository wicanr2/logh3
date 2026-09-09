# -*- coding: utf-8 -*-
"""驗艦艇 `+0x19` 的小地圖標記狀態機（`docs/re/154`）。

    tools/py.sh tools/scripts/minimap_mark_check.py [記憶體傾印] [批次]

預設拿 `workplace/s3fight.bin` 與收據 `s3fight` 的 `screen.bin`。查三件事：

1. `+0x19` bit6（目前點著）立起來的那幾筆，是不是就是
   `+0x1A != 0xFFFF`（有小地圖座標）的那幾筆；
2. 標記的實體像素 ＝ (480 ＋ x ÷ 16, 76 ＋ y ÷ 16) —— 小地圖原點
   (480, 75) 加上 `3C5B:00B2` 那 40 列的畫布偏移；
3. 那個像素的顏色是不是「底色 XOR 陣營墨色」（帝國 5、同盟 9）。
"""
import json
import pathlib
import struct
import sys

POOL, N, REC = 0x668B2, 148, 78   # 艦艇池：docs/internals/38
CANVAS_BIAS = 40                  # 3C5B:00B2 的 +0x0C80：docs/internals/52
PLOT_X, PLOT_Y = 0x1E0, 0x24      # 1727:0324 加的兩個常數
W = 640

root = pathlib.Path('/work')
dump = sys.argv[1] if len(sys.argv) > 1 else 'workplace/s3fight.bin'
batch = sys.argv[2] if len(sys.argv) > 2 else 's3fight'
mem = (root / dump).read_bytes()
scr = (root / ('workplace/oracle/%s/screen.bin' % batch)).read_bytes()
spec = json.loads((root / 'docs/internals/data/screen-battle.json').read_text(encoding='utf-8'))
m = spec['minimap']
ink = m['unit_ink']


def u16(o):
    return struct.unpack_from('<H', mem, o)[0]


bit6, has_pos, rows = [], [], []
for i in range(N):
    r = POOL + i * REC
    b19, x, y = mem[r + 0x19], u16(r + 0x1A), u16(r + 0x1C)
    if b19 & 0x40:
        bit6.append(i)
    if x != 0xFFFF:
        has_pos.append(i)
    if (b19 & 0x40) or x != 0xFFFF:
        px = (x >> 4) + PLOT_X
        py = (y >> 4) + PLOT_Y + CANVAS_BIAS
        rows.append((i, mem[r], b19, x, y, px, py,
                     scr[py * W + px] if 0 <= px < W and 0 <= py * W + px < len(scr) else -1))

print('bit6 立起來的：   ', bit6)
print('+0x1A != FFFF 的：', has_pos)
print('兩組相同：', bit6 == has_pos)
print('小地圖原點 (%d,%d)：實體像素 = (%d + x/16, %d + y/16)'
      % (m['x'], m['y'], PLOT_X, PLOT_Y + CANVAS_BIAS))
print()
print('  # 陣營 +19  +1A   +1C   實體像素  色  底色(色^墨)')
bad = 0
for i, fac, b19, x, y, px, py, c in rows:
    k = ink[fac] if fac < len(ink) else ink[0]
    if px != m['x'] + x // m['unit_scale'] or py != m['y'] + y // m['unit_scale'] + m['unit_row_bias']:
        bad += 1
    print('%3d %3d  %02X %5d %5d  (%3d,%3d) %3d   %2d'
          % (i, fac, b19, x, y, px, py, c, c ^ k))
print()
print('與規格（unit_scale %d、unit_row_bias %d）算出來不同的：%d'
      % (m['unit_scale'], m['unit_row_bias'], bad))
