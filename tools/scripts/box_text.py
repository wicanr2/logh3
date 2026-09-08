"""把小框裡的字讀成規格的 `text`／`codes`，不要留在 `patches` 裡。

    tools/py.sh tools/scripts/box_text.py <批次> <規格名> <視窗名> <框索引> [層數] [ink] [shadow]

`frame_patches` 只記**角落**的修正（`corner=12`），框中央那幾行不記——
所以按鈕的字如果落進 `patches`，中間會缺一塊，而缺的地方剛好是字的中間，
畫出來像「字被咬掉一口」。字要當字存，不要當像素修正存。

**框的深度也要一起壓回去。** `measure_frame` 是一圈一圈往內量的，框裡有字
的時候那幾圈量到的是字的像素，於是一個兩像素的立體邊會被量成十層，
畫出來是一片交錯的雜訊。按鈕的立體邊是兩層（與「中止」「決定」那兩顆
量出來的一樣），層數用參數給，預設 2。
"""
import json, os, sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import (load_screen, font_lookup, frame_patches, read_line,
                        GLYPH_H)

batch, spec_name, win_name = sys.argv[1], sys.argv[2], sys.argv[3]
idx = int(sys.argv[4])
DEPTH = int(sys.argv[5]) if len(sys.argv) > 5 else 2
INK = int(sys.argv[6]) if len(sys.argv) > 6 else 7
SHADOW = int(sys.argv[7]) if len(sys.argv) > 7 else 3

rows, _ = load_screen(f'/work/workplace/oracle/{batch}/screen.bin')
full, half = font_lookup('/orig/GRAPH.IMG')
path = f'/work/docs/internals/data/{spec_name}.json'
spec = json.load(open(path))
box = spec['windows'][win_name]['boxes'][idx]

tx, ty = box['x'] + 2, box['y'] + 2
cells = read_line(rows, tx, ty, box['x'] + box['w'] - 2, INK, full, half)
if not cells:
    sys.exit(f'框 {idx} 在 ({tx},{ty}) 讀不到字')

# **游標蓋住的那一格讀不出來。** 這張畫面的迴圈不理會純移動事件
# （另抓一張只移動滑鼠的收據，畫面逐位元組相同），所以補不回來；
# 那時用 `LOGH3_BOX_CODES` 把整顆按鈕的字碼直接給進來，讀到的那幾格
# 必須是它的前綴——對不上就是給錯了。字碼是否正確由畫面對拍判定。
if os.environ.get('LOGH3_BOX_CODES'):
    want = [int(c, 16) for c in os.environ['LOGH3_BOX_CODES'].split(',')]
    got = [c['code'] for c in cells]
    if want[:len(got)] != got:
        sys.exit(f'框 {idx} 讀到 {got}，與給定的 {want} 對不上')
    x0 = cells[0]['x']
    cells = [{'x': x0 + i * 16, 'w': 16, 'code': c, 'hex': f'{c:04X}'}
             for i, c in enumerate(want)]

box['text_x'], box['text_y'] = cells[0]['x'], ty
box['codes'] = [c['code'] for c in cells]
box['text'] = ''.join(bytes.fromhex(c['hex']).decode('big5', 'replace')
                      if c['w'] == 16 else chr(c['code']) for c in cells)
box['ink'], box['shadow'] = INK, SHADOW

box['layers'] = box['layers'][:DEPTH]
box['alt'] = box['alt'][:DEPTH]

# 字佔掉的那一塊不算框的修正——陰影往左上各多一格。
x0 = cells[0]['x'] - 1
x1 = cells[-1]['x'] + cells[-1]['w']
box['patches'] = frame_patches(
    rows, box['x'], box['y'], box['x'] + box['w'] - 1, box['y'] + box['h'] - 1,
    box['layers'], box['alt'], fill=box['fill'], corner=12,
    exclude=[(x0, ty - 1, x1 - x0, GLYPH_H + 1)])

json.dump(spec, open(path, 'w'), ensure_ascii=False, indent=1)
print(f"框 {idx}：{box['text']} @({box['text_x']},{ty})，修正 {len(box['patches'])} 點")
