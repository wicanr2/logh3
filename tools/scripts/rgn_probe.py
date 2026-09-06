"""`STRMAP.RGN` 是不是「每個星系的區域遮罩」？

檔頭看起來是 6 bytes 一筆的表：`格x, 格y, 寬(格), 高(格), 資料位移(u16)`。
如果真是這樣，每一筆的位移應該正好等於前一筆的位移加上
`寬 × 高 × 40`（一格 16×20 的 1bpp 遮罩是 40 bytes）——**這個關係是
自我驗證的**，猜錯的話第三筆就對不上了。

用法：tools/py.sh tools/scripts/rgn_probe.py
"""
import pathlib, struct

d = pathlib.Path('/orig/STRMAP.RGN').read_bytes()
print('檔案 %d bytes' % len(d))

rows, off = [], 0
while True:
    x, y, w, h, o = struct.unpack_from('<BBBBH', d, off)
    if w == 0 or h == 0 or o != sum(r[2] * r[3] * 40 for r in rows):
        break
    rows.append((x, y, w, h, o))
    off += 6

print('連續吻合 %d 筆，表長 %d bytes' % (len(rows), off))
data = sum(r[2] * r[3] * 40 for r in rows)
print('遮罩資料合計 %d bytes；表 + 資料 = %d（檔案 %d，差 %d）'
      % (data, off + data, len(d), len(d) - off - data))
for i, (x, y, w, h, o) in enumerate(rows[:8]):
    print('  %2d. 格(%2d,%2d) %d×%d 格 → 像素 (%d,%d) %d×%d，位移 %04X'
          % (i, x, y, w, h, x * 16, y * 20, w * 16, h * 20, o))
print('  …')
for i, (x, y, w, h, o) in enumerate(rows[-3:], len(rows) - 3):
    print('  %2d. 格(%2d,%2d) %d×%d 格，位移 %04X' % (i, x, y, w, h, o))
