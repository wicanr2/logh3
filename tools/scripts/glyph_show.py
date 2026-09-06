"""印出 GRAPH.IMG 指定位移的 16×15 字形。"""
import sys, pathlib
f = pathlib.Path(sys.argv[1]).read_bytes()
for a in sys.argv[2:]:
    o = int(a, 0)
    print('位移 %d (0x%X)' % (o, o))
    for r in range(15):
        v = f[o+2*r] << 8 | f[o+2*r+1]
        print('  ' + ''.join('#' if v >> (15-i) & 1 else '.' for i in range(16)))
