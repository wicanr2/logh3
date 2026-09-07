"""把主畫面規格裡「跟局面走」的內容拿掉，改由程式現算。

    tools/py.sh tools/scripts/maingame_computed.py [規格名]

`mainscreen_spec.py` 是照一張收據（帝國、796 年 1 月）量的，量到的東西
裡有三類**不是版面而是內容**，換一個陣營或換一個月份就不對：

1. 底板右側那五行——首都摘要的行星數、艦隊數、成長率、稅率。
   命令階段那一批畫面的規格裡本來就沒有這幾行（`internal/screen.CapitalPanel`
   會畫），主畫面也該一致。
2. 會計報告的署名（y=142）與五個金額（半形數字）。署名跟著內政專任參謀
   走，金額是 `rules.Budget` 算出來的。
3. 報告視窗的頭像——那是署名那位參謀本人，索引跟著提督表走。

留在規格裡的是版面：框、固定的字、分頁鈕、按鈕。
"""
import json
import sys

name = sys.argv[1] if len(sys.argv) > 1 else 'screen-main-game'
path = f'/work/docs/internals/data/{name}.json'
spec = json.load(open(path))

MONEY_ROWS = (222, 242, 262, 282, 302)
STAFF_ROW = 142

board = spec['windows'].get('board')
if board is not None:
    n = len(board.get('lines', []))
    board['lines'] = []
    print(f'底板：拿掉 {n} 行（首都摘要改由 CapitalPanel 畫）')

report = spec['windows'].get('report')
if report is not None:
    lines, dropped, stripped = [], 0, 0
    for l in report.get('lines', []):
        if l['y'] == STAFF_ROW:
            dropped += 1
            continue
        if l['y'] in MONEY_ROWS:
            cells = [c for c in l['cells'] if c['w'] != 8]
            stripped += len(l['cells']) - len(cells)
            if not cells:
                continue
            l = dict(l, cells=cells)
        lines.append(l)
    report['lines'] = lines
    if report.pop('face', None) is not None:
        print('報告：頭像改由程式依幕僚索引畫')
    print(f'報告：拿掉 {dropped} 行署名、{stripped} 個金額字元')

json.dump(spec, open(path, 'w'), ensure_ascii=False, indent=1)
