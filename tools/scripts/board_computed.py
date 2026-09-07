"""把命令階段那一批畫面規格裡「底板右側」的字拿掉。

    tools/py.sh tools/scripts/board_computed.py <規格名> [規格名...]

底板右側是首都摘要，`internal/screen.CapitalPanel` 會照局面現算並畫在
底板之後（`docs/spec/002`）。規格裡再留一份量死的，等於同一塊地方畫兩次
——**而文字是疊上去的，不是蓋上去的**：同一格畫兩個不同的字，兩份點陣
會混在一起，看起來像雜訊。玩帝國時兩份剛好一樣所以看不出來，換成同盟
才顯露（`docs/re/62`）。

`screen-command-phase.json` 本來就沒有這幾行，這一步是讓其餘幾張跟它一致。
"""
import json
import sys

for name in sys.argv[1:]:
    path = f'/work/docs/internals/data/{name}.json'
    spec = json.load(open(path))
    board = spec.get('windows', {}).get('board')
    if board is None:
        print(f'{name}：沒有 board 視窗')
        continue
    n = len(board.get('lines', []))
    board['lines'] = []
    json.dump(spec, open(path, 'w'), ensure_ascii=False, indent=1)
    print(f'{name}：拿掉 {n} 行')
