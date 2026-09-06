import sys, pathlib
W = 640
a = pathlib.Path(sys.argv[1]).read_bytes(); b = pathlib.Path(sys.argv[2]).read_bytes()
x0, y0, x1, y1 = (int(v) for v in sys.argv[3].split(','))
for y in range(y0, y1):
    xs = [x for x in range(x0, x1) if a[y*W+x] != b[y*W+x]]
    if xs:
        print('y=%3d n=%3d x %d..%d' % (y, len(xs), min(xs), max(xs)))
