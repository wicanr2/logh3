"""把「滑鼠停哪一列就刷哪一列」的反白條版面寫進規格。

    tools/py.sh tools/scripts/hover_block.py <規格名> <視窗> <x> <y> <w> <h> <行距> <列數> <色>

反白條**不能量進 `lines`／`rules`**：它跟著滑鼠走，量進去就綁死在收據
那一列上了。規格只記矩形與行距，哪一列由遊戲依滑鼠位置決定
（`internal/screen.GameWindow.Hover`，`docs/re/67`）。

第一列的 `y` 是條子的上緣，比那一列文字的 `y` 少 2。
"""
import json
import sys

name, win = sys.argv[1], sys.argv[2]
x, y, w, h, pitch, rows, colour = (int(v) for v in sys.argv[3:10])
path = f'/work/docs/internals/data/{name}.json'
spec = json.load(open(path))
spec['windows'][win]['hover'] = {'x': x, 'y': y, 'w': w, 'h': h,
                                 'pitch': pitch, 'rows': rows, 'colour': colour}
json.dump(spec, open(path, 'w'), ensure_ascii=False, indent=1)
print(f'{name}／{win}：反白條 ({x},{y}) {w}×{h}，行距 {pitch}，{rows} 列，色 {colour}')
