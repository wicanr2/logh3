# 讀一張畫面上的文字（給探索用：比較兩個狀態的數字差別）。
import sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import W, H, font_lookup, read_line

path = sys.argv[1]
x0, y0, x1, y1 = (int(v) for v in sys.argv[2:6])
ink = int(sys.argv[6]) if len(sys.argv) > 6 else 7
d = open(path, 'rb').read()
rows = [d[y*W:(y+1)*W] for y in range(H)]
full, half = font_lookup('/orig/GRAPH.IMG')
for y in range(y0, y1):
    first = next((x for x in range(x0, x1)
                  if any(rows[yy][x] == ink for yy in range(y, min(y+15, H)))), None)
    if first is None:
        continue
    best = max((read_line(rows, sx, y, x1, ink, full, half)
                for sx in range(max(x0, first-6), first+1)), key=len)
    if len(best) >= 3:
        txt = ''.join(c.get('ascii', chr(c['code']) if c['code'] < 128 else
                            bytes([c['code'] >> 8, c['code'] & 0xFF]).decode('big5hkscs', 'replace'))
                      for c in best)
        print(f'  y={y:3d} {txt}')
