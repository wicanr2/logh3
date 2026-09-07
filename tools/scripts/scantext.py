"""在畫面的一塊矩形裡掃文字（細體與粗體都試，色號可指定多個）。
用法：scantext.py <screen.bin> x0,y0,x1,y1 [色號,...]"""
import sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import load_screen, font_lookup, read_line
rows, _ = load_screen(sys.argv[1])
x0, y0, x1, y1 = (int(v) for v in sys.argv[2].split(','))
inks = [int(v) for v in (sys.argv[3].split(',') if len(sys.argv) > 3 else ['7'])]
tabs = [font_lookup('/orig/GRAPH.IMG'), font_lookup('/orig/GRAPH.IMG', bold=True)]
seen = {}
for ink in inks:
    for full, half in tabs:
        for y in range(y0, y1 - 14):
            best = max((read_line(rows, sx, y, x1, ink, full, half)
                        for sx in range(x0, x0 + 24)), key=len)
            if len(best) >= 2:
                key = (y, best[0]['x'])
                if key not in seen or len(best) > len(seen[key][1]):
                    seen[key] = (ink, best)
for (y, _), (ink, cells) in sorted(seen.items()):
    print('y=%3d ink=%d %s' % (y, ink,
          ''.join(c.get('ascii', '[%s]' % c['hex']) for c in cells)))
