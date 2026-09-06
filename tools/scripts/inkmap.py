"""把 screen.bin 的矩形畫成 ASCII：指定色號為 #，其餘為 .。"""
import sys, pathlib
W = 640
scr = pathlib.Path(sys.argv[1]).read_bytes()
x0, y0, x1, y1 = (int(a) for a in sys.argv[2:6])
inks = set(int(a) for a in sys.argv[6:])
for y in range(y0, y1):
    row = ''.join('#' if (scr[y*W+x] in inks if inks else scr[y*W+x] != scr[y0*W+x0]) else '.'
                  for x in range(x0, x1))
    print('%3d %s' % (y, row))
