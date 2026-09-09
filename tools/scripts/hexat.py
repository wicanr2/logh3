"""從線性位址傾印裡印一段 hex。

    hexat.py <傾印> <起始線性位址hex> <位址hex[:長度hex]> ...
"""
import sys

d = open(sys.argv[1], "rb").read()
base = int(sys.argv[2], 16)
for t in sys.argv[3:]:
    a, _, n = t.partition(":")
    a = int(a, 16)
    n = int(n, 16) if n else 0x60
    o = a - base
    for r in range(0, n, 16):
        print("%05X: %s" % (a + r, d[o + r:o + r + 16].hex(" ")))
    print()
