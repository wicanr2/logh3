# 從建造計劃的規格裡找出「會變的數字」在哪：各列的百分比與金額、餘額。
# remake 要能改預算，就得知道這些數字畫在哪幾格。
import json

p = '/work/docs/internals/data/screen-build-plan.json'
spec = json.load(open(p))
lines = spec['windows']['report']['lines']

def runs(cells):
    """把一行切成一段一段的半形數字（含 % 號）。"""
    out, cur = [], []
    for c in cells:
        # 相鄰兩格中間有空白就分段——不然「3037 3790 1 0 0」會黏成一串。
        if cur and c['x'] != cur[-1]['x'] + cur[-1]['w']:
            out.append(cur)
            cur = []
        if c['w'] == 8 and 0x30 <= c['code'] <= 0x39:
            cur.append(c)
        elif c['w'] == 8 and c['code'] == 0x25 and cur:   # '%'
            cur.append(c)
            out.append(cur)
            cur = []
        else:
            if cur:
                out.append(cur)
                cur = []
    if cur:
        out.append(cur)
    return out

def box(seg):
    return {'x': seg[0]['x'], 'w': seg[-1]['x'] + seg[-1]['w'] - seg[0]['x'],
            'text': ''.join(chr(c['code']) for c in seg)}

rows = []
for l in lines:
    segs = runs(l['cells'])
    if l['y'] == 162 and segs:
        spec.setdefault('fields', {})['cost'] = dict(box(segs[0]), y=l['y'])
    elif 200 <= l['y'] <= 330 and len(segs) >= 2:
        r = {'y': l['y'], 'pct': box(segs[0]), 'amount': box(segs[1])}
        # 金額右邊還有「船艦數」與「部隊」「建造中部隊」三欄，也會跟著預算變。
        for name, k in (('ships', 2), ('units', 3), ('building', 4)):
            if len(segs) > k:
                r[name] = box(segs[k])
        rows.append(r)
    elif l['y'] == 342 and len(segs) >= 2:
        spec.setdefault('fields', {})['rest'] = {
            'y': l['y'], 'pct': box(segs[0]), 'amount': box(segs[1])}
spec['fields']['rows'] = rows
spec['fields']['note'] = ('會變的數字：各列的百分比、金額、船艦數、部隊、'
                          '建造中部隊，以及餘額。每一欄都是靠右對齊，'
                          '右邊界固定。')
json.dump(spec, open(p, 'w'), ensure_ascii=False, indent=1)
print('cost:', spec['fields'].get('cost'))
for r in rows:
    print(f"  y={r['y']} pct={r['pct']} amount={r['amount']}")
print('rest:', spec['fields'].get('rest'))
