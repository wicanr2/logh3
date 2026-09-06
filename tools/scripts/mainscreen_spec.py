"""量出遊戲主畫面（796 年 1 月）的版面。

視窗的 bbox 取自最外層框線（色號 15 的長線段），不是從填色往外剝——
三個區塊的填色都是 9 而且互相連通，flood fill 會一路跑到整個畫面。
"""
import collections, json, sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import (W, H, load_screen, font_lookup, measure_frame,
                        frame_patches, read_line, ink_rows)

batch = sys.argv[1] if len(sys.argv) > 1 else 'key1'
outname = sys.argv[2] if len(sys.argv) > 2 else 'screen-main-game'
rows, flat = load_screen(f'/work/workplace/oracle/{batch}/screen.bin')
full, half = font_lookup('/orig/GRAPH.IMG')
# 背景取畫面角落，**不是**出現最多的色號——主畫面上視窗填色比背景多。
bg = rows[0][0]

# 主畫面只有**兩個**視窗：一塊蓋滿下方的大底板，加上疊在上面的會計報告。
# 星域地圖畫在底板上，正好蓋住它的左上角——所以看起來像三、四個區塊。
# bbox 要含最外那一層（背景色 3），少算一層的話整圈框都會差一個像素。
WINDOWS = {
    'board': (0, 40, 639, 418),
    'report': (144, 80, 495, 399),
}
# 疊上去的那個視窗每個畫面大小不同，可以從命令列蓋掉。
if len(sys.argv) > 8:
    WINDOWS['report'] = tuple(int(v) for v in sys.argv[5:9])
# 第三層：有些畫面在報告視窗之上再疊一個（待命中提督資料）。
if len(sys.argv) > 12:
    WINDOWS['overlay'] = tuple(int(v) for v in sys.argv[9:13])
# 底板的四邊要在沒被遮住的地方取樣；左上角整個被地圖蓋住，不能拿來
# 量角落修正。
SAMPLE = {'board': (600, 380)}
EXCLUDE = {'board': ((0, 40, 320, 300),)}
if 'overlay' in WINDOWS:
    # 底板與報告視窗都被最上層蓋掉右半邊，量框與掃字都要避開。
    _ox0 = WINDOWS['overlay'][0]
    EXCLUDE['board'] = EXCLUDE['board'] + ((_ox0, 40, 640, 480),)
    EXCLUDE['report'] = ((_ox0, 40, 640, 480),)
    SAMPLE['report'] = (WINDOWS['report'][0] + 20, WINDOWS['report'][3] - 20)

# 視窗裡自成一格的小框（輸入框、面板上的橫條、地圖旁的裝飾條）。
# 位置是從「remake 與 oracle 的差異連通分量」量出來的
# （tools/scripts/diff_parts.py）。
BOXES = {
    'board': [(496, 109, 631, 152), (1, 340, 143, 348), (320, 41, 327, 79),
              (311, 400, 327, 417)],
    'report': [],
}
# 疊上去那個視窗裡的小框，各畫面不同。位置量自「remake 與 oracle 的差異
# 連通分量」（tools/scripts/diff_parts.py）——那支工具直接回答「還缺什麼」。
EXTRA_BOXES = {
    # 右下的輸入框是**上下兩格**，不是一個 36×40 的框。
    'screen-main-game': [(430, 340, 465, 359), (430, 360, 465, 379)],
    'screen-ship-order': [(430, 340, 465, 359), (430, 360, 465, 379)],
    # 預算分配的上下箭頭：2 欄 × 7 列，每個 16×20；外加兩個按鈕。
    'screen-build-plan': [(208 + c*16, 200 + r*20, 208 + c*16 + 15, 200 + r*20 + 19)
                          for r in range(7) for c in range(2)]
                         + [(446, 360, 481, 379), (494, 360, 529, 379)],
    # 預算計劃：兩組上下箭頭（徵稅檔位、建造費百分比）與兩個按鈕。
    'screen-budget-plan': [(464, 280, 479, 299), (480, 280, 495, 299),
                           (464, 320, 479, 339), (480, 320, 495, 339),
                           (398, 360, 433, 379), (446, 360, 481, 379)],
}
# 純色的線（報告視窗裡的表格分隔線）。
# 各畫面自己的補充：游標的可見範圍（原版的游標是疊上去的，視窗一重畫
# 就把它蓋掉一部分）與熱區（座標見 docs/re/09）。
EXTRA_SPEC = {
    'screen-main-game': {
        'window_origin_grid': [10, 3],
        'hotspots': [
            {'x': 432, 'y': 300, 'w': 32, 'h': 20, 'grid': [17, 12],
             'action': 'budget-plan', 'note': '次月份預算計劃（docs/re/12）'},
            {'x': 432, 'y': 320, 'w': 32, 'h': 20, 'grid': [17, 13],
             'action': 'ship-order', 'note': '796 年 1 月份船艦訂造報告'},
        ],
    },
    'screen-ship-order': {'mouse_clip': {'x': 0, 'y': 378, 'w': 640, 'h': 102}},
    'screen-build-plan': {'mouse_clip': {'x': 0, 'y': 0, 'w': 640, 'h': 362}},
    'screen-budget-plan': {
        'hotspots': [],
        'note': '次月份預算計劃；由主畫面熱區 (432,300) 點進去（docs/re/12）',
    },
}

# 表格分隔線也是各畫面自己的：套到別的畫面會蓋掉那裡的框。
RULES_BY_SCREEN = {
    'screen-main-game': [(160, 259, 192, 1), (160, 279, 192, 1), (160, 299, 192, 1)],
    'screen-budget-plan': [(144, 239, 256, 1), (144, 279, 256, 1)],
}
# 同一個視窗裡不是每一行都用同一個顏色：面板的艦隊數字是 6 與 5。
TEXT_INKS = {'board': [7, 6, 5], 'report': [7], 'overlay': [7]}

# 文字掃描區：底板被地圖與報告視窗遮住的部分讀不到，分成兩塊來讀。
_rx0, _ry0, _rx1, _ry1 = WINDOWS['report']
TEXT_AREAS = {
    # 底板被地圖與那個視窗遮住的部分讀不到；掃描區跟著視窗走。
    'board': ([] if 'overlay' in WINDOWS
              else [(max(496, _rx1 + 1), 49, 638, 410)] if _rx1 + 40 < 638 else []),
    # 右邊界停在 420：再往右是輸入框，它的框線點陣剛好等於某個製表符，
    # 會被讀成一行「文字」，而那行的**陰影**會把框的左緣蓋掉一個像素。
    # 掃描區要避開小框：框線的點陣剛好等於某些製表符，會被讀成一行
    # 「文字」，而那行的陰影會把框的邊緣蓋掉一個像素。
    'report': {
        'screen-main-game': [(158, 90, 483, 390)],
        'screen-ship-order': [(158, 90, 483, 390)],
        'screen-build-plan': [(94, 70, 545, 392)],
        'screen-budget-plan': [(142, 70, 497, 392)],
    }.get(outname, [(WINDOWS['report'][0]+14, WINDOWS['report'][1]+10,
                     WINDOWS['report'][2]-12, WINDOWS['report'][3]-10)]),
}
if 'overlay' in WINDOWS:
    _v0, _w0, _v1, _w1 = WINDOWS['overlay']
    TEXT_AREAS['overlay'] = [(_v0 + 14, _w0 + 10, _v1 - 12, _w1 - 10)]
    # 報告視窗只剩左半邊看得到。
    TEXT_AREAS['report'] = [(WINDOWS['report'][0] + 14, WINDOWS['report'][1] + 10,
                             _v0 - 2, WINDOWS['report'][3] - 10)]
spec = {'canvas': {'width': W, 'height': H, 'mode': 'VGA 12h 640x480x16'},
        'background': bg,
        'map': {'source': 'STRMAP.BIT', 'regions': 'STRMAP.RGN', 'x': 0, 'y': 40},
        'windows': {}}

for name, (x0, y0, x1, y1) in WINDOWS.items():
    layers, alts = measure_frame(rows, x0, y0, x1, y1, sample=SAMPLE.get(name))
    # 角落範圍要比層數寬：斜切的最後一個像素落在第 9、10 格上，
    # corner=8 會剛好差一個像素而看不出來。
    patches = frame_patches(rows, x0, y0, x1, y1, layers, alts,
                            corner=12, exclude=EXCLUDE.get(name, ()))
    # 文字色固定 7、陰影 3（屬性 0x97，與主選單、防拷問答同一套）。
    # 自動判斷會被頭像帶偏——那塊的像素比文字多得多。
    ink = 7
    # 文字行距固定 20（與主選單、防拷問答相同），**相位用全域最佳化找**：
    # 一行一行往下找的話，某一行誤判（框線的點陣剛好等於某個半形字）
    # 之後的行全部跟著錯位。
    def ink_count(y, xs, xe, ink):
        # 算整個字高的總和。只看頂端那一列的話，上緣留白的字（多數中文字）
        # 會被當成空行整行跳過。
        return sum(1 for yy in range(y, min(y+15, H))
                   for x in range(xs, xe) if rows[yy][x] == ink)

    def scan(phase, ink):
        out = []
        for xs, ay0, xe, ay1 in TEXT_AREAS[name]:
            for y in range(ay0 + phase, ay1 - 14, 20):
                if ink_count(y, xs, xe, ink) < 30:
                    continue
                first = next((x for x in range(xs, xe)
                              if any(rows[yy][x] == ink for yy in range(y, min(y+15, H)))), None)
                if first is None:
                    continue
                skip = [(bx0, by0, bx1-bx0+1, by1-by0+1)
                        for bx0, by0, bx1, by1 in EXTRA_BOXES.get(outname, [])] \
                    if name == 'report' else []
                best_line = max((read_line(rows, sx, y, xe, ink, full, half, skip)
                                 for sx in range(max(xs, first-6), first+1)), key=len)
                # 整行同一個字碼 ＝ 框線（點陣剛好等於某個製表符），不是文字。
                # 整行同一個字碼且夠長 ＝ 框線（點陣剛好等於某個製表符）。
                # 條件不能只寫「全同碼」——面板上那個孤零零的數字就會被吃掉。
                if best_line and (len(best_line) < 4
                                  or len({c['code'] for c in best_line}) > 1):
                    out.append({'y': y, 'cells': best_line})
        return out

    lines = []
    for ink in TEXT_INKS[name]:
        got = max((scan(p, ink) for p in range(20)),
                  key=lambda ls: sum(len(l['cells']) for l in ls))
        for l in got:
            l['ink'] = ink
        lines += got
    lines.sort(key=lambda l: (l['y'], l['cells'][0]['x']))
    spec['windows'][name] = {
        'x': x0, 'y': y0, 'w': x1-x0+1, 'h': y1-y0+1,
        'frame': {'layers': layers, 'alt': alts, 'fill': 9, 'patches': patches},
        'ink': 7, 'lines': lines,
        'boxes': [{'x': bx0, 'y': by0, 'w': bx1-bx0+1, 'h': by1-by0+1,
                   'layers': (lay := measure_frame(rows, bx0, by0, bx1, by1,
                                                   fill=rows[(by0+by1)//2][(bx0+bx1)//2])[0]),
                   'alt': measure_frame(rows, bx0, by0, bx1, by1,
                                        fill=rows[(by0+by1)//2][(bx0+bx1)//2])[1],
                   'fill': rows[(by0+by1)//2][(bx0+bx1)//2],
                   'patches': frame_patches(
                       rows, bx0, by0, bx1, by1,
                       *measure_frame(rows, bx0, by0, bx1, by1,
                                      fill=rows[(by0+by1)//2][(bx0+bx1)//2]),
                       # 這些小框整塊比對：它們是純色填充，沒有「內容」
                       # 會被誤當成修正，而角落的斜切在小框上佔的比例很高。
                       fill=rows[(by0+by1)//2][(bx0+bx1)//2], corner=999)}
                  for bx0, by0, bx1, by1 in (BOXES.get(name, []) if name == 'board'
                                             else EXTRA_BOXES.get(outname + ':' + name,
                                                                  EXTRA_BOXES.get(outname, [])
                                                                  if name == 'report' else []))],
        'rules': [{'x': rx, 'y': ry, 'w': rw, 'h': rh, 'colour': rows[ry][rx]}
                  for rx, ry, rw, rh in (RULES_BY_SCREEN.get(
                      outname if name == 'report' else outname + ':' + name, []))],
    }
    print(f'{name}: ({x0},{y0}) {x1-x0+1}×{y1-y0+1}  層 {len(layers)}  '
          f'角落修正 {len(patches)}  文字色 {ink}  {len(lines)} 行')
    for ln in lines:
        txt = ''.join(c.get('ascii', f"[{c['hex']}]") for c in ln['cells'])
        print(f"    y={ln['y']:3d} {txt}")

# 不是框也不是文字的元件：頭像、輸入框、面板上的橫條。
# 每個都從一個內部點往外掃到填色改變為止。
def solid(cx, cy):
    c = rows[cy][cx]
    x0 = cx
    while x0 > 0 and rows[cy][x0-1] == c: x0 -= 1
    x1 = cx
    while x1 < W-1 and rows[cy][x1+1] == c: x1 += 1
    y0 = cy
    while y0 > 0 and rows[y0-1][cx] == c: y0 -= 1
    y1 = cy
    while y1 < H-1 and rows[y1+1][cx] == c: y1 += 1
    return {'x': x0, 'y': y0, 'w': x1-x0+1, 'h': y1-y0+1, 'colour': c}

# 游標位置是那一份收據的輸入序列決定的，不能沿用別的畫面。
if 'overlay' in WINDOWS:
    spec['order'] = ['report', 'overlay']
spec['mouse'] = {'x': int(sys.argv[3]) if len(sys.argv) > 3 else 300,
                 'y': int(sys.argv[4]) if len(sys.argv) > 4 else 170}
# 頭像編號不寫死：在報告視窗裡把 83 張逐一比對，找逐像素相同的那張。
import struct as _st
_f = open('/orig/FACE.AMG', 'rb').read()
_idx = [_st.unpack_from('<HI', _f, k*6) for k in range(83)]
sys.path.insert(0, '/work/tools/scripts')
from amg_decode import decode as _dec
_best = None
for _k, (_s, _o) in enumerate(_idx):
    _W, _H, _px, _ = _dec(_f[_o:_o+_s])
    for _y0 in range(_ry0 + 5, _ry0 + 60):
        for _x0 in range(_rx1 - 130, _rx1 - 40):
            _bad = 0
            for _y in range(0, _H, 4):
                _r = [rows[_y0+_y][_x0+_i] for _i in range(_W)]
                _bad += sum(1 for _a, _b in zip(_r, _px[_y*_W:(_y+1)*_W]) if _a != _b)
                if _best and _bad > _best[0]:
                    break
            if _best is None or _bad < _best[0]:
                _best = (_bad, _k, _x0, _y0)
print(f'頭像：第 {_best[1]} 張 於 ({_best[2]},{_best[3]})  抽樣差異 {_best[0]}')
spec['windows']['report']['face'] = {'source': 'FACE.AMG', 'index': _best[1],
                                     'x': _best[2], 'y': _best[3]}
spec['windows']['report']['input_box'] = solid(460, 350)
spec['windows']['board']['bar'] = solid(560, 125)
print('輸入框', spec['windows']['report']['input_box'])
print('面板橫條', spec['windows']['board']['bar'])

for _k, _v in EXTRA_SPEC.get(outname, {}).items():
    if _k == 'mouse_clip':
        spec['mouse']['clip'] = _v
    else:
        spec[_k] = _v

json.dump(spec, open(f'/work/docs/internals/data/{outname}.json', 'w'),
          ensure_ascii=False, indent=1)
print(f'→ docs/internals/data/{outname}.json')
