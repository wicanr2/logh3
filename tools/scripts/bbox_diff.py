"""比兩張 screen.bin，印差異數與外接矩形。可限定矩形：--rect x0,y0,x1,y1"""
import sys, pathlib
W = 640
args = sys.argv[1:]
rect = None
if '--rect' in args:
    i = args.index('--rect')
    rect = tuple(int(v) for v in args[i+1].split(','))
    del args[i:i+2]
a = pathlib.Path(args[0]).read_bytes()
for p in args[1:]:
    b = pathlib.Path(p).read_bytes()
    if rect:
        x0, y0, x1, y1 = rect
        idx = [y*W+x for y in range(y0, y1) for x in range(x0, x1)]
    else:
        idx = range(len(a))
    d = [i for i in idx if a[i] != b[i]]
    name = p.split('/')[-2]
    if not d:
        print('%-28s 完全相同' % name); continue
    xs = [i % W for i in d]; ys = [i // W for i in d]
    print('%-28s 差 %6d  bbox x%d..%d y%d..%d' % (name, len(d), min(xs), max(xs), min(ys), max(ys)))
