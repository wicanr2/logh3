"""待命中提督資料畫面的規格。

這個畫面是三層：底板 → 次月份船艦建造計劃 → 提督清單。下面兩層與
`screen-build-plan` 完全相同（只是右半被蓋住），所以直接沿用那份規格，
這裡只量最上層——重量一次反而會因為文字被切在視窗邊界而量錯。
"""
import json, sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import (W, H, load_screen, font_lookup, measure_frame,
                        frame_patches, read_line)

BATCH = 'admirals1'
BASE = '/work/docs/internals/data/screen-build-plan.json'
OUT = '/work/docs/internals/data/screen-admiral-list.json'
# 最上層視窗（量自 admirals1 的框線剖面）。
X0, Y0, X1, Y1 = 320, 40, 639, 418
# 清單列：16 條凹framed，x 336..615、y 80 起、高 20、間距 20。
BAR_X, BAR_W, BAR_H, BAR_N, BAR_Y0, BAR_PITCH = 336, 280, 20, 16, 80, 20
MOUSE = (464, 330)

rows, _ = load_screen(f'/work/workplace/oracle/{BATCH}/screen.bin')
full, half = font_lookup('/orig/GRAPH.IMG')
spec = json.load(open(BASE))

layers, alts = measure_frame(rows, X0, Y0, X1, Y1)
patches = frame_patches(rows, X0, Y0, X1, Y1, layers, alts, corner=12)

# 右側的捲軸：與清單列同高、同樣的凹框，x 616..623。
SB_X, SB_W = 616, 8

boxes = []
for k in range(BAR_N):
    bx0, by0 = BAR_X, BAR_Y0 + k * BAR_PITCH
    bx1, by1 = bx0 + BAR_W - 1, by0 + BAR_H - 1
    fill = rows[(by0 + by1) // 2][(bx0 + bx1) // 2]
    lay, alt = measure_frame(rows, bx0, by0, bx1, by1, fill=fill)
    boxes.append({'x': bx0, 'y': by0, 'w': BAR_W, 'h': BAR_H,
                  'layers': lay, 'alt': alt, 'fill': fill,
                  'patches': frame_patches(rows, bx0, by0, bx1, by1, lay, alt,
                                           fill=fill, corner=999)})
    sx0, sy0 = SB_X, BAR_Y0 + k * BAR_PITCH
    sx1, sy1 = sx0 + SB_W - 1, sy0 + BAR_H - 1
    sfill = rows[(sy0 + sy1) // 2][(sx0 + sx1) // 2]
    slay, salt = measure_frame(rows, sx0, sy0, sx1, sy1, fill=sfill)
    boxes.append({'x': sx0, 'y': sy0, 'w': SB_W, 'h': BAR_H,
                  'layers': slay, 'alt': salt, 'fill': sfill,
                  'patches': frame_patches(rows, sx0, sy0, sx1, sy1, slay, salt,
                                           fill=sfill, corner=999)})

# 底板左下的狀態列（帝國軍星／治安）只在這個畫面出現，補掃一次。
board_extra = []
for ink in (6, 7):
    for y in range(355, 400):
        ln = max((read_line(rows, sx, y, 140, ink, full, half, [])
                  for sx in range(10, 30)), key=len)
        if len(ln) >= 2:
            board_extra.append({'y': y, 'ink': ink, 'cells': ln})
spec['windows']['board']['lines'] = sorted(
    spec['windows']['board']['lines'] + board_extra,
    key=lambda l: (l['y'], l['cells'][0]['x']))

# 只有標題一行文字，在清單列上方。
lines = []
for y in range(Y0 + 10, BAR_Y0):
    ln = max((read_line(rows, sx, y, X1 - 12, 7, full, half, [])
              for sx in range(X0 + 10, X0 + 30)), key=len)
    if len(ln) >= 2:
        lines.append({'y': y, 'ink': 7, 'cells': ln})
        break

spec['windows']['overlay'] = {
    'x': X0, 'y': Y0, 'w': X1 - X0 + 1, 'h': Y1 - Y0 + 1,
    'frame': {'layers': layers, 'alt': alts, 'fill': 9, 'patches': patches},
    'ink': 7, 'lines': lines, 'boxes': boxes, 'rules': [],
}
spec['order'] = ['report', 'overlay']
# 原版是先畫游標、再刷清單列，所以游標落在清單列上的那幾段會被蓋掉。
spec['mouse'] = {'x': MOUSE[0], 'y': MOUSE[1], 'after_boxes': 'overlay'}
spec['hotspots'] = []
spec['note'] = '待命中提督資料；由建造計劃左按鈕點進去（docs/re/12）'
json.dump(spec, open(OUT, 'w'), ensure_ascii=False, indent=1)
txt = ''.join(c.get('ascii', '[%s]' % c['hex']) for c in (lines[0]['cells'] if lines else []))
print(f'overlay ({X0},{Y0}) {X1-X0+1}×{Y1-Y0+1} 層 {len(layers)} 角落 {len(patches)} '
      f'清單列 {len(boxes)} 標題 y={lines[0]["y"] if lines else "-"} {txt}')
print('→', OUT)
