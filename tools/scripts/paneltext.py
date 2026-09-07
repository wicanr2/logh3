"""掃一塊矩形裡的文字，連格子座標一起印出來（給量面板版面用）。

    paneltext.py <screen.bin> x0,y0,x1,y1 [色號,...]

輸出每一行：y、ink、以及每一格的 x/w/字碼。與 scantext.py 的差別是
它會把格子座標吐出來，量版面要的就是那個。
"""
import sys, json
sys.path.insert(0, '/work/tools/scripts')
from screenspec import load_screen, font_lookup, read_line

rows, _ = load_screen(sys.argv[1])
x0, y0, x1, y1 = (int(v) for v in sys.argv[2].split(','))
inks = [int(v) for v in (sys.argv[3].split(',') if len(sys.argv) > 3 else ['7'])]
tabs = [(False, font_lookup('/orig/GRAPH.IMG')), (True, font_lookup('/orig/GRAPH.IMG', bold=True))]

out = []
for ink in inks:
    for bold, (full, half) in tabs:
        for y in range(y0, y1 - 14):
            best = max((read_line(rows, sx, y, x1, ink, full, half)
                        for sx in range(x0, x0 + 24)), key=len)
            if len(best) >= 2:
                out.append({'y': y, 'ink': ink, 'bold': bold,
                            'cells': [{'x': c['x'], 'w': c['w'], 'code': c['code']} for c in best],
                            'text': ''.join(c.get('ascii', '[%s]' % c['hex']) for c in best)})
# 同一個 (y, ink) 只留最長的一條
best = {}
for l in out:
    k = (l['y'], l['ink'])
    if k not in best or len(l['cells']) > len(best[k]['cells']):
        best[k] = l
print(json.dumps([best[k] for k in sorted(best)], ensure_ascii=False, indent=1))
