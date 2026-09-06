#!/usr/bin/env python3
"""把「疊在既有畫面上的一層視窗」量成規格。

    tools/py.sh tools/scripts/overlay_spec.py <批次> <底層規格> <輸出名> \
        <x0> <y0> <x1> <y1> <游標x> <游標y>

底層原封不動沿用（那一層已經對拍過），只量最上面那一層——重量一次
會因為文字被切在視窗邊界而讀錯（`docs/re/12`）。
"""
import json, sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import (W, H, load_screen, font_lookup, measure_frame,
                        frame_patches, read_line)

batch, base_name, out_name = sys.argv[1], sys.argv[2], sys.argv[3]
X0, Y0, X1, Y1 = (int(v) for v in sys.argv[4:8])
MX, MY = int(sys.argv[8]), int(sys.argv[9])

rows, _ = load_screen(f'/work/workplace/oracle/{batch}/screen.bin')
full, half = font_lookup('/orig/GRAPH.IMG')
spec = json.load(open(f'/work/docs/internals/data/{base_name}.json'))

layers, alts = measure_frame(rows, X0, Y0, X1, Y1)
patches = frame_patches(rows, X0, Y0, X1, Y1, layers, alts, corner=12)

# 文字：行距 20，相位用全域最佳化找（一行誤判會讓後面全部錯位）。
def scan(phase):
    out = []
    for y in range(Y0 + 10 + phase, Y1 - 14, 20):
        ln = max((read_line(rows, sx, y, X1 - 12, 7, full, half, [])
                  for sx in range(X0 + 8, X0 + 40)), key=len)
        if len(ln) >= 2 and len({c['code'] for c in ln}) > 1:
            out.append({'y': y, 'ink': 7, 'cells': ln})
    return out

lines = max((scan(p) for p in range(20)),
            key=lambda ls: sum(len(l['cells']) for l in ls))

spec['windows']['overlay'] = {
    'x': X0, 'y': Y0, 'w': X1 - X0 + 1, 'h': Y1 - Y0 + 1,
    'frame': {'layers': layers, 'alt': alts, 'fill': 9, 'patches': patches},
    'ink': 7, 'lines': lines, 'boxes': [], 'rules': [],
}
spec['order'] = ['report', 'overlay']
spec['mouse'] = {'x': MX, 'y': MY}
spec['hotspots'] = []
json.dump(spec, open(f'/work/docs/internals/data/{out_name}.json', 'w'),
          ensure_ascii=False, indent=1)
print('overlay (%d,%d) %d×%d 層 %d 角落 %d 文字 %d 行'
      % (X0, Y0, X1 - X0 + 1, Y1 - Y0 + 1, len(layers), len(patches), len(lines)))
for l in lines:
    print('   y=%3d %s' % (l['y'], ''.join(c.get('ascii', '[%s]' % c['hex']) for c in l['cells'])))
