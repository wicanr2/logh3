"""把對戰交棒那一刻底板上多出來的兩隻箭頭量進規格。

    tools/py.sh tools/scripts/versus_board.py <批次> <規格名>

命令階段的底板（`screen-command-phase`）量自 `main1e`，右下角清單那兩顆
上下鈕裡面是空的；交棒的那一刻鈕裡各有一隻箭頭（收據 `vs3`／`vs5`，
`docs/spec/016`）。整塊底板只差這兩顆，所以照交棒那張把它們重量一次，
其餘沿用底層。

**箭頭不是字**：它是色 9 的三角形配色 3 的暗邊與色 7 的亮邊，查不到
字型裡的碼，所以照框的修正點（`frame_patches`，角落半徑放大到整顆鈕）
記下來。
"""
import json
import sys

sys.path.insert(0, '/work/tools/scripts')
from screenspec import load_screen, measure_frame, frame_patches

batch, name = sys.argv[1], sys.argv[2]
rows, _ = load_screen(f'/work/workplace/oracle/{batch}/screen.bin')
path = f'/work/docs/internals/data/{name}.json'
spec = json.load(open(path))
board = spec['windows']['board']

n = 0
for b in board['boxes']:
    # 右緣那一欄的兩顆小鈕（14×20）。中間那顆 14×60 是滑塊軌道，不動。
    if b['x'] < 600 or b['h'] != 20:
        continue
    x0, y0 = b['x'], b['y']
    x1, y1 = x0 + b['w'] - 1, y0 + b['h'] - 1
    fill = rows[(y0 + y1) // 2][(x0 + x1) // 2]
    layers, alts = measure_frame(rows, x0, y0, x1, y1, fill=fill)
    b['layers'], b['alt'], b['fill'] = layers, alts, fill
    b['patches'] = frame_patches(rows, x0, y0, x1, y1, layers, alts,
                                 fill=fill, corner=max(b['w'], b['h']))
    n += 1
    print('鈕 (%d,%d)：層 %d 修正 %d 點' % (x0, y0, len(layers), len(b['patches'])))
if n != 2:
    sys.exit(f'預期兩顆鈕，量到 {n} 顆')
json.dump(spec, open(path, 'w'), ensure_ascii=False, indent=1)
