"""拿畫面上一段像素的 planar 位元組去素材檔裡搜。

    tools/py.sh tools/scripts/find_in_asset.py <批次> <x> <y> <寬(8的倍數)> <檔...>

畫面是 4 位元色號，VGA planar 一列每個 plane 各 W/8 bytes
（`docs/internals/10`）。這裡把那一段轉成四個 plane 的位元組，
再逐一到素材檔裡找——**比反組譯快得多**，命中就知道那塊圖出自哪裡。
"""
import pathlib
import sys

W = 640
batch, x0, y, w = sys.argv[1], int(sys.argv[2]), int(sys.argv[3]), int(sys.argv[4])
scr = pathlib.Path('/work/workplace/oracle/%s/screen.bin' % batch).read_bytes()

row = [scr[y*W + x0 + i] for i in range(w)]
planes = []
for p in range(4):
    bs = bytearray()
    for i in range(0, w, 8):
        b = 0
        for k in range(8):
            if row[i+k] >> p & 1:
                b |= 0x80 >> k
        bs.append(b)
    planes.append(bytes(bs))

print('(%d,%d) 寬 %d 的色號：%s' % (x0, y, w, ' '.join('%X' % c for c in row[:32])))
for p, bs in enumerate(planes):
    print('  plane %d: %s' % (p, bs.hex()))

for name in sys.argv[5:]:
    data = pathlib.Path('/orig/' + name).read_bytes()
    hits = []
    for p, bs in enumerate(planes):
        if len(set(bs)) < 2:
            continue  # 整段同一個值，搜了也沒意義
        i = data.find(bs)
        if i >= 0:
            hits.append('plane %d @ %d(0x%X)' % (p, i, i))
    print('%-14s %s' % (name, '、'.join(hits) if hits else '沒有'))
