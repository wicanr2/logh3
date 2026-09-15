"""比兩張 screen.bin，按 16×20 格印差異格數與每格差異數（格座標＝字元平面）。"""
import sys, pathlib
W, H = 640, 480
a = pathlib.Path(sys.argv[1]).read_bytes()
b = pathlib.Path(sys.argv[2]).read_bytes()
cells = {}
for i in range(W*H):
    if a[i] != b[i]:
        x, y = i % W, i // W
        cells.setdefault((x//16, (y-40)//20 if y >= 40 else -1), []).append(i)
print('差異格數', len(cells), '差異像素', sum(len(v) for v in cells.values()))
rows = {}
for (cx, cy), v in sorted(cells.items(), key=lambda kv: (kv[0][1], kv[0][0])):
    rows.setdefault(cy, []).append((cx, len(v)))
for cy in sorted(rows):
    print('列 %3d:' % cy, ' '.join('%d(%d)' % (cx, n) for cx, n in rows[cy]))
