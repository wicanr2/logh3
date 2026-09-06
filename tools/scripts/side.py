"""並排印出兩張 640×480 色號圖的同一塊。"""
import sys, pathlib
W = 640
a = pathlib.Path(sys.argv[1]).read_bytes(); b = pathlib.Path(sys.argv[2]).read_bytes()
x0, y0, x1, y1 = (int(v) for v in sys.argv[3].split(','))
hdr = ''.join('%x' % (x % 16) for x in range(x0, x1))
print('    %s   %s' % (hdr, hdr))
for y in range(y0, y1):
    ra = ''.join('%X' % a[y*W+x] for x in range(x0, x1))
    rb = ''.join('%X' % b[y*W+x] for x in range(x0, x1))
    print('%3d %s   %s %s' % (y, ra, rb, '<' if ra != rb else ''))
