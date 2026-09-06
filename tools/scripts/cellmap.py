"""印出 screen.bin 指定矩形的色號（每格一個十六進位字）。"""
import sys, pathlib
W = 640
scr = pathlib.Path(sys.argv[1]).read_bytes()
x0, y0, x1, y1 = (int(a) for a in sys.argv[2:6])
print('    ' + ''.join('%x' % (x % 16) for x in range(x0, x1)))
for y in range(y0, y1):
    print('%3d ' % y + ''.join('%X' % scr[y*W+x] for x in range(x0, x1)))
