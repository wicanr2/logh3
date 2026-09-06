"""找一塊區域裡的框線：色 15／8 連續夠長的水平段與垂直段。

量面板版面時用得到——畫面上看起來像「三個按鈕」的東西，
實際上可能是一條大框加上嵌進去的三個小框，中間那幾塊 15 是框自己的填色。
連通分量（`diff_parts.py`）只說「這裡不一樣」，這支說「線在哪裡」。

用法：tools/py.sh tools/scripts/panel_edges.py <批次> <x0> <y0> <x1> <y1>
"""
import pathlib, sys

batch = sys.argv[1]
x0, y0, x1, y1 = (int(v) for v in sys.argv[2:6])
scr = pathlib.Path(f'/work/workplace/oracle/{batch}/screen.bin').read_bytes()
W = 640

def runs(seq, lo, minlen):
    out, s = [], None
    for i, c in seq:
        if c in (15, 8):
            s = i if s is None else s
        else:
            if s is not None and i - s >= minlen:
                out.append((s, i - 1))
            s = None
    if s is not None and lo - s >= minlen:
        out.append((s, lo))
    return out

print('水平線：')
for y in range(y0, y1 + 1):
    r = runs(((x, scr[y * W + x]) for x in range(x0, x1 + 1)), x1, 20)
    if r:
        print('  y=%3d %s' % (y, r))
print('垂直線：')
for x in range(x0, x1 + 1):
    r = runs(((y, scr[y * W + x]) for y in range(y0, y1 + 1)), y1, 10)
    if r:
        print('  x=%3d %s' % (x, r))
