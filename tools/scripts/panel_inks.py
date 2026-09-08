"""列右側面板每一列用到的色號（依字元格分段）。

    tools/py.sh tools/scripts/panel_inks.py <批次> [x0 y0 列數]
"""
import collections
import pathlib
import sys

W = 640
batch = sys.argv[1]
x0 = int(sys.argv[2]) if len(sys.argv) > 2 else 480
y0 = int(sys.argv[3]) if len(sys.argv) > 3 else 242
rows = int(sys.argv[4]) if len(sys.argv) > 4 else 8
scr = pathlib.Path("/work/workplace/oracle/%s/screen.bin" % batch).read_bytes()

for r in range(rows):
    y = y0 + r * 20
    out = []
    for c in range(9):
        x = x0 + c * 16
        h = collections.Counter(scr[(y + j) * W + x + i]
                                for j in range(16) for i in range(16))
        h.pop(9, None)          # 底色
        h.pop(3, None)          # 立體邊
        out.append("-" if not h else "/".join(str(k) for k, _ in h.most_common(2)))
    print("列%d y=%3d  %s" % (r, y, "  ".join("%-5s" % v for v in out)))
