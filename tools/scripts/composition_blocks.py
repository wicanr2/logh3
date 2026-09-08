"""把艦隊組成規格裡「量成框、其實不是框」的兩塊改成該有的樣子。

    tools/py.sh tools/scripts/composition_blocks.py

`overlay_spec.py` 要把格盤那一塊排除在文字掃描之外，只能把它當成小框傳
進去；但它其實是**一塊沒有立體邊的純色矩形**（(112,140) 192×200、色 3），
量成框會得到 16 層假的層次。這支把它換成 `rules` 裡的一塊純色。

七個艦種列的框裡，第 0 列的內部是色 2 ——那是**選取**不是底色
（滑鼠停在別的地方）。七個框的 `layers`／`alt`／`patches` 逐項相同，
只有 `fill` 不同，所以把第 0 列的 `fill` 改回 8，選取條改成 `select`
區塊由程式畫（`docs/re/67`）。

格盤本身寫成 `grid`：9 欄 × 9 列、一格 16 × 16、圖塊來自 `SPRITE.SPR`，
`(欄 + 列)` 偶數的格子畫「空位」那一張。
"""
import json

PATH = '/work/docs/internals/data/screen-fleet-composition.json'
spec = json.load(open(PATH))
w = spec['windows']['overlay']

# 1. 格盤那一塊：框 → 純色矩形
boxes = w['boxes']
grid_box = next((b for b in boxes if b['x'] == 112 and b['y'] == 140), None)
if grid_box is not None:
    boxes.remove(grid_box)
    w.setdefault('rules', [])
    w['rules'] = [r for r in w['rules'] if not (r['x'] == 112 and r['y'] == 140)]
    w['rules'].append({'x': 112, 'y': 140, 'w': 192, 'h': 200, 'colour': 3})
    print('格盤：框 → 純色矩形 (112,140) 192×200 色 3')

# 2. 七個艦種列：第 0 列的 fill 改回 8，選取條交給程式
rows = [b for b in boxes if b['w'] == 68]
if len(rows) != 7:
    raise SystemExit(f'艦種列應該有 7 個框，找到 {len(rows)} 個')
ref = rows[1]
for k in ('layers', 'alt', 'patches'):
    if any(b[k] != ref[k] for b in rows):
        raise SystemExit(f'七個艦種列的 {k} 不一致，不能只改第 0 列的 fill')
if rows[0]['fill'] != ref['fill']:
    print(f"艦種第 0 列：fill {rows[0]['fill']} → {ref['fill']}（那是選取條，不是底色）")
    rows[0]['fill'] = ref['fill']
w['select'] = {'first': boxes.index(rows[0]), 'count': 7, 'colour': 2}

# 3. 格盤內容
w['grid'] = {'x': 136, 'y': 168, 'cols': 9, 'rows': 9, 'tile': 16,
             'source': 'SPRITE.SPR', 'empty': 44, 'unit': 36,
             'note': '(欄 + 列) 偶數的 41 格才畫；中央 (4,4) 是編隊的第一格。'}

json.dump(spec, open(PATH, 'w'), ensure_ascii=False, indent=1)
print('選取條 select', w['select'], '；格盤 grid', {k: v for k, v in w['grid'].items() if k != 'note'})
