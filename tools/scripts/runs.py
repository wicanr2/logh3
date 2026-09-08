"""把一份收據的某一列或某一行切成「同色的連續段」。

    tools/py.sh tools/scripts/runs.py <批次> row|col <座標> [最短長度]

量框線、面板邊界、格線間距用。比逐點看快得多。
"""
import sys

W, H = 640, 480
d = open('/work/workplace/oracle/%s/screen.bin' % sys.argv[1], 'rb').read()
mode, at = sys.argv[2], int(sys.argv[3])
least = int(sys.argv[4]) if len(sys.argv) > 4 else 2

if mode == 'row':
    line = [d[at*W + x] for x in range(W)]
else:
    line = [d[y*W + at] for y in range(H)]

i = 0
while i < len(line):
    c = line[i]
    j = i
    while j < len(line) and line[j] == c:
        j += 1
    if j - i >= least:
        print('%s%d  %3d..%3d  色%2d（%d）' % (mode, at, i, j-1, c, j-i))
    i = j
