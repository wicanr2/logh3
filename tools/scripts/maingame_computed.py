"""把「疊在戰略底板上」那幾張規格裡跟局面走的內容拿掉，改由程式現算。

    tools/py.sh tools/scripts/maingame_computed.py <規格名>

`mainscreen_spec.py` 是照一張收據（帝國、796 年 1 月）量的，量到的東西
裡有幾類**不是版面而是內容**，換一個陣營或換一個月份就不對：

1. 底板右側那幾行——首都摘要的行星數、艦隊數、成長率、稅率。
   命令階段那一批畫面的規格裡本來就沒有這幾行（`internal/screen.CapitalPanel`
   會畫），這幾張也該一致。
2. 署名那一行與那位參謀的頭像。改成規格只記位置與職位
   （`staff`／`face.staff_post`），姓名、軍階與頭像索引由遊戲填。
3. 半形數字（金額、造艦數、部隊數）與標題的全形年月。

留在規格裡的是版面：框、固定的字、分頁鈕、按鈕。對應規格 `docs/spec/011`。
"""
import json
import sys

# 每一張的：署名那一行的 y、職位（0 軍事 1 情報 2 內政）、
# 要清掉半形數字的列、標題那一行的 y（清全形數字）。
RECIPES = {
    'screen-main-game': {
        'staff_y': 142, 'post': 2, 'staff_x': 160,
        'money_rows': (222, 242, 262, 282, 302), 'header_y': None,
        # 月初的會計報告連五個金額一起交給 internal/screen.ReportPanel，
        # 署名與頭像也在那裡，所以規格不留 staff 區塊。
        'drop_face': True, 'staff_block': False,
    },
    'screen-ship-order': {
        'staff_y': 142, 'post': 0, 'staff_x': 160,
        'money_rows': (242, 262, 282, 302, 322, 342, 362), 'header_y': 102,
        'drop_face': False, 'staff_block': True,
    },
    'screen-build-plan': {
        'staff_y': 122, 'post': 0, 'staff_x': 96,
        # y=162 的「船艦建造費」是 rules.Budget 算出來的（ReportPanel）。
        'money_rows': (162,), 'header_y': None,
        'drop_face': False, 'staff_block': True,
    },
}

name = sys.argv[1] if len(sys.argv) > 1 else 'screen-main-game'
r = RECIPES[name]
path = f'/work/docs/internals/data/{name}.json'
spec = json.load(open(path))

board = spec['windows'].get('board')
if board is not None and board.get('lines'):
    print(f'底板：拿掉 {len(board["lines"])} 行（首都摘要改由 CapitalPanel 畫）')
    board['lines'] = []

report = spec['windows'].get('report')
if report is not None:
    lines, dropped, stripped = [], 0, 0
    for l in report.get('lines', []):
        if l['y'] == r['staff_y']:
            dropped += 1
            continue
        if l['y'] in r['money_rows']:
            cells = [c for c in l['cells'] if c['w'] != 8]
            stripped += len(l['cells']) - len(cells)
            if not cells:
                continue
            l = dict(l, cells=cells)
        if r['header_y'] is not None and l['y'] == r['header_y']:
            cells = [c for c in l['cells']
                     if not 0xA2AF <= c['code'] <= 0xA2B8]
            stripped += len(l['cells']) - len(cells)
            l = dict(l, cells=cells)
        lines.append(l)
    report['lines'] = lines
    print(f'報告：拿掉 {dropped} 行署名、{stripped} 個數字')

    if r['drop_face']:
        if report.pop('face', None) is not None:
            print('報告：頭像改由程式依幕僚索引畫')
    elif report.get('face') is not None:
        report['face']['staff_post'] = r['post']
        print(f'報告：頭像改成跟著第 {r["post"]} 個職位的幕僚走')

    if r['staff_block']:
        report['staff'] = {'post': r['post'], 'x': r['staff_x'],
                           'y': r['staff_y'], 'ink': report.get('ink', 7)}
        print(f'報告：署名改成 staff 區塊 @({r["staff_x"]},{r["staff_y"]})')

json.dump(spec, open(path, 'w'), ensure_ascii=False, indent=1)
