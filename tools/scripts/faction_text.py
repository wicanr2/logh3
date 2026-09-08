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
done = [t for t in w.get('faction_text', []) if t['y'] == y]
if not old and done:
    print(f'{name}／{win} y={y}：已經搬過了（{done[0]["text"]}）')
    sys.exit(0)
if not old:
    sys.exit(f'{name} 的 {win} 沒有 y={y} 這一行')
line = old[0]
x, ink = line['cells'][0]['x'], line.get('ink', 7)

# **兩個陣營的字不一定同色。** 貨幣單位那一行兩邊都是色 7，但陣營名
# （「帝國軍」／「同盟軍」）是用陣營色印的——帝國 6、同盟 5
# （`rules.FactionColour`）。所以每一份收據各自試一遍顏色，
# 只用帝國那一個去讀同盟那張會整行讀不到。
INKS = [ink] + [i for i in (6, 5, 7) if i != ink]

full, half = font_lookup('/orig/GRAPH.IMG')
variants, texts, inks = [], [], []
for b in batches:
    rows, _ = load_screen(f'/work/workplace/oracle/{b}/screen.bin')
    cells, hit = [], ink
    for cand in INKS:
        cells = read_line(rows, x, y, x + 400, cand, full, half)
        if cells:
            hit = cand
            break
    inks.append(hit)
    if not cells:
        sys.exit(f'收據 {b} 的 ({x},{y}) 讀不到字')
    variants.append([c['code'] for c in cells])
    texts.append(''.join(bytes.fromhex(c['hex']).decode('big5', 'replace')
                         if c['w'] == 16 else chr(c['code']) for c in cells))

# **只搬走那一行，不是那一列。** 同一個 y 上可以有兩行：陣營名（陣營色）
# 與後面接的固定字（色 7）——對戰的交棒視窗就是這樣。整列刪掉的話，
# 固定字那半邊會從規格裡消失，畫出來只剩「帝國軍」。
w['lines'] = [l for l in w['lines']
              if not (l['y'] == y and l['cells'][0]['x'] == x
                      and l.get('ink', 7) == ink)]
w.setdefault('faction_text', [])
w['faction_text'] = [t for t in w['faction_text'] if t['y'] != y]
entry = {'x': x, 'y': y, 'ink': ink, 'text': texts, 'variants': variants}
if len(set(inks)) > 1:
    entry['inks'] = inks
w['faction_text'].append(entry)
json.dump(spec, open(path, 'w'), ensure_ascii=False, indent=1)
print(f'{name}／{win} y={y}：{texts}')
