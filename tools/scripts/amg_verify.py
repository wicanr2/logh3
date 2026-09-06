# 決定性驗證：解碼結果要能在原版 framebuffer 上找到逐像素相同的位置。
import sys, struct
sys.path.insert(0, '/work/tools/scripts')
from amg_decode import decode
f = open('/orig/FACE.AMG', 'rb').read()
s, o = struct.unpack_from('<HI', f, 0)
W, H, px, _ = decode(f[o:o+s])
fb = open('/work/workplace/oracle/face-fresh/screen.bin', 'rb').read()
SW = 640
best = None
for y0 in range(180, 210):
    for x0 in range(135, 165):
        if x0 + W > SW or y0 + H > 480:
            continue
        bad = sum(1 for y in range(H) for x in range(W)
                  if fb[(y0+y)*SW + x0+x] != px[y*W+x])
        if best is None or bad < best[0]:
            best = (bad, x0, y0)
print(f'最佳對齊：({best[1]},{best[2]}) 不同像素 {best[0]} / {W*H}')
