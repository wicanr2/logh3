#!/usr/bin/env python3
"""把「疊在既有畫面上的一層視窗」量成規格。

    tools/py.sh tools/scripts/overlay_spec.py <批次> <底層規格> <輸出名> \
        <x0> <y0> <x1> <y1> <游標x> <游標y> [小框 x0:y0:x1:y1,...] \
        [純色塊 x:y:w:h:色,...] [頭像 index:x:y] [額外文字區 x0:y0:x1:y1]

底層原封不動沿用（那一層已經對拍過），只量最上面那一層——重量一次
會因為文字被切在視窗邊界而讀錯（`docs/re/12`）。

游標座標給 `-1` ＝ 沿用底層那一份（含它的可見範圍）。疊一層上去通常
不會動到游標：滑鼠沒移動，畫面上那隻還是原來被夾成幾列的樣子。

小框（清單列、按鈕）要明列——它們的框線點陣剛好等於某些製表符，
不列出來就會被讀成一行「文字」。位置用 `diff_parts.py` 的連通分量抄。
"""
import json, sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import (W, H, load_screen, font_lookup, measure_frame,
                        frame_patches, read_line)

batch, base_name, out_name = sys.argv[1], sys.argv[2], sys.argv[3]
X0, Y0, X1, Y1 = (int(v) for v in sys.argv[4:8])
MX, MY = int(sys.argv[8]), int(sys.argv[9])
BOXES = [tuple(int(v) for v in b.split(':'))
         for b in (sys.argv[10].split(',') if len(sys.argv) > 10 and sys.argv[10] else [])]
# 純色塊：圖例的色票那種，沒有立體邊，用框去量會量出一堆假的層次。
RULES = [tuple(int(v) for v in r.split(':'))
         for r in (sys.argv[11].split(',') if len(sys.argv) > 11 and sys.argv[11] else [])]
# 頭像：哪一張用 tools/scripts/face_find.py 問，別用肉眼比。
FACE = ([int(v) for v in sys.argv[12].split(':')]
        if len(sys.argv) > 12 and sys.argv[12] else None)
# 額外文字區：有些畫面除了疊上去那一層，還會改到底層別處的字
# （提督資料會重寫左下角那塊）。座標是絕對的，掃出來的行併進同一層。
EXTRA = ([int(v) for v in sys.argv[13].split(':')]
         if len(sys.argv) > 13 and sys.argv[13] else None)

# 額外文字來源：**游標蓋住的那一格讀不出來**，而有些畫面的迴圈不理會純移動
# 事件（設定、生產計劃…），游標只會停在最後一次點擊的地方。同一張畫面點在
# 不同位置各抓一份收據，兩份的文字聯集就補得回來。幾何仍然只看主收據。
ALSO = sys.argv[14] if len(sys.argv) > 14 and sys.argv[14] else None

rows, _ = load_screen(f'/work/workplace/oracle/{batch}/screen.bin')
full, half = font_lookup('/orig/GRAPH.IMG')
# 原版有兩種粗細：細的直接查字型檔，粗的要拿加粗後的點陣查
# （`docs/internals/26`）。兩種都掃，重疊的行由 _merge 取字多的那個。
fullb, halfb = font_lookup('/orig/GRAPH.IMG', bold=True)
spec = json.load(open(f'/work/docs/internals/data/{base_name}.json'))

layers, alts = measure_frame(rows, X0, Y0, X1, Y1)
patches = frame_patches(rows, X0, Y0, X1, Y1, layers, alts, corner=12)

# 文字：行距 20，相位用全域最佳化找（一行誤判會讓後面全部錯位）。
def scan(phase, tabs):
    out = []
    for y in range(Y0 + 10 + phase, Y1 - 14, 20):
        skip = [(bx0, by0, bx1 - bx0 + 1, by1 - by0 + 1)
                for bx0, by0, bx1, by1 in BOXES]
        # 頭像也要跳過：人像的線條會被讀成製表符或底線，而且畫的時候
        # 頭像在文字之前，那幾格會被字的陰影咬掉一列。
        if FACE:
            skip.append((FACE[1], FACE[2], 80, 100))
        # 起點要掃寬一點：這一層裡不只一欄文字，左邊那欄的縮排
        # 各畫面不同，範圍太窄就整欄讀不到。
        ln = max((read_line(rows, sx, y, X1 - 12, 7, tabs[0], tabs[1], skip)
                  for sx in range(X0 + 4, X0 + 80)), key=len)
        # 短短一兩格的製表符不是文字，是被讀成字的框線——而畫出來的
        # 陰影會把那條線蓋掉一個像素。
        if ln and len(ln) <= 4 and all(0xA140 <= c['code'] <= 0xA1FE for c in ln):
            continue
        if ln and (len(ln) < 4 or len({c['code'] for c in ln}) > 1):
            out.append({'y': y, 'ink': 7, 'bold': tabs[2], 'cells': ln})
    return out

def _span(line):
    xs = [c['x'] for c in line['cells']]
    return min(xs), max(c['x'] + c['w'] for c in line['cells'])

def _merge(acc, cand):
    """一層裡的行不只一種相位（標籤欄與清單欄各排各的），所以把所有相位
    的結果併起來，再把佔到同一塊位置的行去掉，只留字數多的那一個。
    只取單一最佳相位會整欄漏掉。"""
    lo, hi = _span(cand)
    for i, cur in enumerate(acc):
        # 顏色不同的兩行可以疊在同一列（標籤是色 7、數字是色 6），
        # 只有同色才算佔到同一塊位置。
        if cur.get('ink') != cand.get('ink'):
            continue
        if abs(cur['y'] - cand['y']) >= 15:
            continue
        clo, chi = _span(cur)
        if lo < chi and clo < hi:
            if len(cand['cells']) > len(cur['cells']):
                acc[i] = cand
            return
    acc.append(cand)

def scan_area(x0, y0, x1, y1, phase, ink=7, tabs=None):
    """額外文字區。**顏色不只一種**——左下那塊的數字是色 6 與 5，
    只掃 7 會整片漏掉。"""
    out = []
    for y in range(y0 + phase, y1 - 14, 20):
        skip = [(bx0, by0, bx1 - bx0 + 1, by1 - by0 + 1)
                for bx0, by0, bx1, by1 in BOXES]
        ln = max((read_line(rows, sx, y, x1, ink, tabs[0], tabs[1], skip)
                  for sx in range(x0, x0 + 40)), key=len)
        if ln and len(ln) <= 4 and all(0xA140 <= c['code'] <= 0xA1FE for c in ln):
            continue
        if ln and (len(ln) < 4 or len({c['code'] for c in ln}) > 1):
            out.append({'y': y, 'ink': ink, 'bold': tabs[2], 'cells': ln})
    return out

TABS = [(full, half, False), (fullb, halfb, True)]
lines = []
for _t in TABS:
    for _p in range(20):
        for _l in scan(_p, _t):
            _merge(lines, _l)
if ALSO:
    _main_rows = rows
    rows, _ = load_screen(f'/work/workplace/oracle/{ALSO}/screen.bin')
    _extra = []
    for _t in TABS:
        for _p in range(20):
            for _l in scan(_p, _t):
                _merge(_extra, _l)
    rows = _main_rows

    def _union(acc, cand):
        """逐格聯集。兩份各自被游標咬掉不同的一格，誰也不是誰的超集，
        所以不能用「留字多的那一份」——要把格子併起來。"""
        for cur in acc:
            if cur.get('ink') != cand.get('ink') or cur['y'] != cand['y']:
                continue
            have = {c['x'] for c in cur['cells']}
            add = [c for c in cand['cells']
                   if c['x'] not in have
                   and not any(o['x'] < c['x'] + c['w'] and c['x'] < o['x'] + o['w']
                               for o in cur['cells'])]
            if add:
                cur['cells'] = sorted(cur['cells'] + add, key=lambda c: c['x'])
            return True
        return False

    # 只往既有的行補格子，**不新增行**——第二份收據萬一點歪了、開的是別的
    # 畫面，多出來的行就會整片變成假資料。少一格看得出來，多一行看不出來。
    for _l in _extra:
        _union(lines, _l)
if EXTRA:
    for _t in TABS:
        for _ink in (7, 6, 5):
            for _p in range(20):
                for _l in scan_area(*EXTRA, _p, _ink, _t):
                    _merge(lines, _l)
lines.sort(key=lambda l: (l['y'], l['cells'][0]['x']))

def _refit_under_base_cursor(spec):
    """底層規格裡「其實是游標」的修正點，拿這一張畫面重量一次。

    底層那一份是連著游標一起量的：游標壓在框上的那幾格會被當成框的
    角落修正記進 `patches`。在底層那張畫面上看不出來——游標最後會蓋回
    同一批像素；但這一層把游標移走之後，那幾格就露出來了。

    不能整批丟掉：游標蓋住的像素裡有一部分本來就是框的修正。所以改成
    **在這一張畫面上重量那一塊**——這裡游標已經移開，量到的才是框本身。
    """
    m = spec.get('mouse')
    if not m:
        return 0
    cur = json.load(open('/work/docs/internals/data/cursor.json'))
    ox, oy = cur['offset_from_mouse']
    cx0, cy0 = m['x'] + ox, m['y'] + oy
    cx1, cy1 = cx0 + cur['width'], cy0 + cur['height']
    n = 0

    def abs_xy(x, y, w, h, pt):
        return (x + pt['dx'] if pt['dx'] >= 0 else x + w + pt['dx'],
                y + pt['dy'] if pt['dy'] >= 0 else y + h + pt['dy'])

    def refit(x, y, w, h, layers, alts, fill, patches):
        nonlocal n
        if not (x < cx1 and cx0 < x + w and y < cy1 and cy0 < y + h):
            return patches
        fresh = frame_patches(rows, x, y, x + w - 1, y + h - 1, layers, alts,
                              fill=fill, corner=12,
                              exclude=[(X0, Y0, X1 - X0 + 1, Y1 - Y0 + 1)])
        inside = lambda pt: (lambda a: cx0 <= a[0] < cx1 and cy0 <= a[1] < cy1)(
            abs_xy(x, y, w, h, pt))
        out = [pt for pt in patches if not inside(pt)] + [pt for pt in fresh if inside(pt)]
        n += len(patches) + len(out) - 2 * sum(1 for pt in patches if pt in out)
        return out

    for win in spec['windows'].values():
        f = win.get('frame')
        if f:
            f['patches'] = refit(win['x'], win['y'], win['w'], win['h'],
                                 f['layers'], f['alt'], f.get('fill', 9),
                                 f.get('patches') or [])
        for b in win.get('boxes') or []:
            b['patches'] = refit(b['x'], b['y'], b['w'], b['h'],
                                 b['layers'], b['alt'], b.get('fill', 9),
                                 b.get('patches') or [])
    return n


if MX >= 0:
    _n = _refit_under_base_cursor(spec)
    if _n:
        print('底層游標壓住那一塊重量：修正點動了 %d 個' % _n)

spec['windows']['overlay'] = {
    'x': X0, 'y': Y0, 'w': X1 - X0 + 1, 'h': Y1 - Y0 + 1,
    'frame': {'layers': layers, 'alt': alts, 'fill': 9, 'patches': patches},
    'ink': 7, 'lines': lines,
    'rules': [{'x': rx, 'y': ry, 'w': rw, 'h': rh, 'colour': rc}
              for rx, ry, rw, rh, rc in RULES],
    'boxes': [{'x': bx0, 'y': by0, 'w': bx1 - bx0 + 1, 'h': by1 - by0 + 1,
               'layers': (_l := measure_frame(rows, bx0, by0, bx1, by1,
                                              fill=rows[(by0 + by1) // 2][(bx0 + bx1) // 2]))[0],
               'alt': _l[1],
               'fill': rows[(by0 + by1) // 2][(bx0 + bx1) // 2],
               'patches': frame_patches(rows, bx0, by0, bx1, by1, *_l,
                                        fill=rows[(by0 + by1) // 2][(bx0 + bx1) // 2],
                                        corner=12)}
              for bx0, by0, bx1, by1 in BOXES],
}
# 底層不一定有 report——命令階段主畫面只有 board 一個視窗。
if FACE:
    spec['windows']['overlay']['face'] = {
        'source': 'FACE.AMG', 'index': FACE[0], 'x': FACE[1], 'y': FACE[2]}
spec['order'] = [n for n in ('report', 'overlay') if n in spec['windows']]
# 游標座標給 -1 ＝ **沿用底層那一份**（含它的可見範圍）。
# 疊一層上去通常不會動到游標：滑鼠沒移動，畫面上那隻還是原來被夾成
# 幾列的樣子，重新算一次反而會畫成完整的一隻。
if MX >= 0:
    spec['mouse'] = {'x': MX, 'y': MY}
elif MX == -2:
    # 這一層把游標整個蓋掉了（覆蓋範圍含游標位置），不要畫。
    spec.pop('mouse', None)
spec['hotspots'] = []
json.dump(spec, open(f'/work/docs/internals/data/{out_name}.json', 'w'),
          ensure_ascii=False, indent=1)
print('overlay (%d,%d) %d×%d 層 %d 角落 %d 文字 %d 行'
      % (X0, Y0, X1 - X0 + 1, Y1 - Y0 + 1, len(layers), len(patches), len(lines)))
for l in lines:
    print('   y=%3d %s' % (l['y'], ''.join(c.get('ascii', '[%s]' % c['hex']) for c in l['cells'])))
