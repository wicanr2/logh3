"""把畫面 dump 的一塊區域用色號畫成粗略的地圖。

    tools/py.sh tools/scripts/screen_map.py <批次> x0 y0 x1 y1 [取樣間隔]
"""
import collections
import pathlib
import sys

W = 640
batch = sys.argv[1]
x0, y0, x1, y1 = (int(v) for v in sys.argv[2:6])
step = int(sys.argv[6]) if len(sys.argv) > 6 else 4
scr = pathlib.Path("/work/workplace/oracle/%s/screen.bin" % batch).read_bytes()

hist = collections.Counter(scr[y * W + x] for y in range(y0, y1) for x in range(x0, x1))
print("色號分佈:", " ".join("%d:%d" % kv for kv in sorted(hist.items())))
print("   " + "".join("%d" % ((x // step) % 10) for x in range(x0, x1, step)))
for y in range(y0, y1, step):
    row = "".join("0123456789ABCDEF"[scr[y * W + x] & 15] if scr[y * W + x] else "."
                  for x in range(x0, x1, step))
    print("%3d %s" % (y, row))
