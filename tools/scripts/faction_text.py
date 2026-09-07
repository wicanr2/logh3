"""把「兩個陣營印不同字」的那一行從規格搬進 `faction_text`。

    tools/py.sh tools/scripts/faction_text.py <規格名> <視窗> <y> <帝國批次> <同盟批次>

規格是照帝國那張收據量的，所以那一行原本只有帝國的版本。這支從兩張
收據各讀一次同一行，寫成 `faction_text`（`variants` 依陣營索引），
並把原本那一行從 `lines` 拿掉。

目前只用在貨幣單位：帝國「單位＝億帝國馬克」、同盟「單位＝億帝納」。
"""
import json
import sys

sys.path.insert(0, '/work/tools/scripts')
from screenspec import load_screen, font_lookup, read_line

name, win, y = sys.argv[1], sys.argv[2], int(sys.argv[3])
batches = sys.argv[4:6]
path = f'/work/docs/internals/data/{name}.json'
spec = json.load(open(path))
w = spec['windows'][win]

old = [l for l in w.get('lines', []) if l['y'] == y]
if not old:
    sys.exit(f'{name} 的 {win} 沒有 y={y} 這一行')
line = old[0]
x, ink = line['cells'][0]['x'], line.get('ink', 7)

full, half = font_lookup('/orig/GRAPH.IMG')
variants, texts = [], []
for b in batches:
    rows, _ = load_screen(f'/work/workplace/oracle/{b}/screen.bin')
    cells = read_line(rows, x, y, x + 400, ink, full, half)
    if not cells:
        sys.exit(f'收據 {b} 的 ({x},{y}) 讀不到字')
    variants.append([c['code'] for c in cells])
    texts.append(''.join(bytes.fromhex(c['hex']).decode('big5', 'replace')
                         if c['w'] == 16 else chr(c['code']) for c in cells))

w['lines'] = [l for l in w['lines'] if l['y'] != y]
w.setdefault('faction_text', [])
w['faction_text'] = [t for t in w['faction_text'] if t['y'] != y]
w['faction_text'].append({'x': x, 'y': y, 'ink': ink,
                          'text': texts, 'variants': variants})
json.dump(spec, open(path, 'w'), ensure_ascii=False, indent=1)
print(f'{name}／{win} y={y}：{texts}')
