"""這個畫面上的頭像是 `FACE.AMG` 的第幾張？

規格裡的 `face.index` 以前是人工比對出來的。這支直接拿每一張去跟畫面
指定位置比色號，回報最相符的幾張與差異數——**0 差異才算找到**，
「最接近」不算，因為 83 張裡有好幾張構圖相近。

用法：tools/py.sh tools/scripts/face_find.py <批次> <x> <y>
"""
import struct, sys
sys.path.insert(0, '/work/tools/scripts')
from amg_decode import decode

batch = sys.argv[1]
X, Y = int(sys.argv[2]), int(sys.argv[3])
scr = open(f'/work/workplace/oracle/{batch}/screen.bin', 'rb').read()
f = open('/orig/FACE.AMG', 'rb').read()
idx = [struct.unpack_from('<HI', f, k * 6) for k in range(83)]

out = []
for k, (s, o) in enumerate(idx):
    w, h, px, _ = decode(f[o:o + s])
    bad = sum(1 for y in range(h) for x in range(w)
              if Y + y < 480 and X + x < 640
              and px[y * w + x] != scr[(Y + y) * 640 + X + x])
    out.append((bad, k, w, h))
out.sort()
for bad, k, w, h in out[:5]:
    print('第 %2d 張 %d×%d 差 %d 像素%s' % (k, w, h, bad, '  ← 完全相同' if bad == 0 else ''))
