"""把 `STRMAP.RGN` 的 32 個星區遮罩畫出來。

表是 6 bytes 一筆（`格x, 格y, 寬(格), 高(格), 資料位移 u16`），
遮罩以 16×20 的格子為單位，一格 40 bytes ＝ 每列 2 bytes、20 列，
word 的高位在左（與 `STRMAP.BIT` 相同）。

用法：tools/py.sh tools/scripts/rgn_export.py
"""
import collections, pathlib, struct, sys
sys.path.insert(0, '/work/tools/scripts')
from pngout import write_png, game_palette

W, H, CW, CH = 320, 300, 16, 20
d = pathlib.Path('/orig/STRMAP.RGN').read_bytes()
pal = game_palette('/work/workplace/oracle/main1/screen.png')

rows, off = [], 0
while True:
    x, y, w, h, o = struct.unpack_from('<BBBBH', d, off)
    if w == 0 or h == 0 or o != sum(r[2] * r[3] * 40 for r in rows):
        break
    rows.append((x, y, w, h, o)); off += 6
base = off

px = bytearray(W * H)
owner = bytearray(W * H)          # 每個像素屬於哪一區（0 ＝ 沒有）
overlap = 0
for idx, (gx, gy, gw, gh, o) in enumerate(rows, 1):
    for cy in range(gh):
        for cx in range(gw):
            cell = base + o + (cy * gw + cx) * 40
            for y in range(CH):
                word = d[cell + y * 2] << 8 | d[cell + y * 2 + 1]
                py = (gy + cy) * CH + y
                if py >= H:
                    continue
                for k in range(CW):
                    if word >> (15 - k) & 1:
                        pxx = (gx + cx) * CW + k
                        if pxx >= W:
                            continue
                        i = py * W + pxx
                        if owner[i]:
                            overlap += 1
                        owner[i] = idx
                        px[i] = 1 + (idx - 1) % 15

covered = sum(1 for v in owner if v)
print('32 個星區的遮罩：覆蓋 %d / %d 像素（%.1f%%），重疊 %d'
      % (covered, W * H, covered * 100 / (W * H), overlap))
cnt = collections.Counter(owner)
print('每區像素數（前 6）：', cnt.most_common(7)[1:])
write_png('/work/workplace/assets/strmap-regions.png', W, H, px, pal)
