"""量出命令階段主畫面的版面（右鍵兩下關掉月初報告之後的那張）。

這張畫面只有**一個**視窗：蓋滿下方的大底板，星域地圖畫在它的左上角。
其餘看起來像視窗的東西（命令列三個框、右上的橫條、右側資訊面板、
下方清單）都是畫在底板上的小框——視窗串列雖然有六個視窗
（`docs/internals/24`），但框線是底板一次畫完的。

小框的位置從「remake 與 oracle 的差異連通分量」量出來
（`tools/scripts/diff_parts.py`）：先產一版沒有小框的規格，
對拍之後把差異連通分量抄回下面的 BOXES。

用法：tools/py.sh tools/scripts/cmdscreen_spec.py [批次] [輸出名]
"""
import json, sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import (W, H, load_screen, font_lookup, measure_frame,
                        frame_patches, read_line)

batch = sys.argv[1] if len(sys.argv) > 1 else 'main1'
outname = sys.argv[2] if len(sys.argv) > 2 else 'screen-command-phase'
rows, flat = load_screen(f'/work/workplace/oracle/{batch}/screen.bin')
full, half = font_lookup('/orig/GRAPH.IMG')
bg = rows[0][0]

BOARD = (0, 40, 639, 418)
# 底板的框**直接沿用已經對拍過的那份**（`screen-main-game`）——同一塊底板，
# 重量一次只會因為這張畫面上多出來的小框而讀錯層次。
BASE = json.load(open('/work/docs/internals/data/screen-main-game.json',
                      encoding='utf-8'))['windows']['board']

# 畫在底板上的小框。位置對得上視窗串列（`docs/internals/24`）的五個面板：
# 右上、命令列、右側資訊、左下——星圖那一個被地圖整個蓋掉，不用畫。
BOXES = [
    (320, 40, 639, 119),    # 右上：日期與軍事費
    # 命令列面板：角落修正要抓到 24 px——命令列那條框的上下緣（y 114、152）
    # 離面板邊 14 px，corner 只有 12 的話那四段線修不到。
    (320, 100, 639, 159, None, 40),
    (320, 140, 639, 419),   # 右側資訊：星系狀態與清單
    (0, 340, 319, 419),     # 左下：對方星系
]
# 面板裡的小框（命令列那三個灰框、清單框…）。掃文字時要避開它們的框線，
# 位置從差異連通分量補（`tools/scripts/diff_parts.py`）。
SMALL_BOXES = [
    # 清單標題列的四個小框（同樣是高 20、含立體邊）
    (336, 280, 385, 299), (400, 280, 449, 299),
    (464, 280, 529, 299), (544, 280, 593, 299),
] + [
    # 每個標題框裡靠左有一個 7×12 的凹格（左緣 ＋8），四個都一樣。
    (x + 8, 285, x + 14, 296) for x in (336, 400, 464, 544)
] + [
    # 命令列：先畫整條（填色要在按鈕之間取樣），再畫三個按鈕框
    (328, 115, 631, 145, (430, 130)),
    (350, 120, 417, 139), (446, 120, 513, 139), (542, 120, 609, 139),
    # 右側捲軸：上箭頭、滑軌、下箭頭
    (610, 300, 623, 319), (610, 320, 623, 379), (610, 380, 623, 399),
    # 清單第一條沒有自己的上緣暗邊——y 298–299 那兩列是標題框的底邊，
    # 而標題框之間有縫，整條拉過去會把縫塗掉。
    (336, 300, 609, 319),
] + [
    # 其餘四條：間距 20、每條高 22，上緣兩列暗、再兩列亮，
    # 下一條的上緣壓在前一條的最後兩列上。
    (336, 318 + 20 * k, 609, 339 + 20 * k) for k in range(4)
]

# 文字掃描區：地圖蓋住左上角，所以右側與左下分開掃。
# 掃描區要**閃開框線**：底板左緣的框線也是色 7，掃描起點會停在那裡，
# 整行就從框線開始讀，讀出一格垃圾之後結束。
TEXT_AREAS = [(330, 46, 632, 412), (16, 344, 312, 414)]
TEXT_INKS = [7, 6, 5]

x0, y0, x1, y1 = BOARD

def ink_count(y, xs, xe, ink):
    return sum(1 for yy in range(y, min(y + 15, H))
               for x in range(xs, xe) if rows[yy][x] == ink)

def scan(area, phase, ink):
    """掃一個區塊的一種相位。

    **相位要每個區塊各找各的**：這張畫面上半的面板從 y ≡ 1 起排，
    下半的資訊欄從 y ≡ 4 起排，用同一個相位掃就會漏掉另一半。"""
    out = []
    for xs, ay0, xe, ay1 in (area,):
        for y in range(ay0 + phase, ay1 - 14, 20):
            if ink_count(y, xs, xe, ink) < 30:
                continue
            first = next((x for x in range(xs, xe)
                          if any(rows[yy][x] == ink
                                 for yy in range(y, min(y + 15, H)))), None)
            if first is None:
                continue
            # 面板框不進 skip：那是**整片**區域，跳過就等於不掃文字。
            # skip 只用來避開小按鈕框（它們的框線點陣剛好等於某些製表符）。
            skip = [(b[0], b[1], b[2] - b[0] + 1, b[3] - b[1] + 1)
                    for b in SMALL_BOXES]
            best = max((read_line(rows, sx, y, xe, ink, full, half, skip)
                        for sx in range(max(xs, first - 6), first + 1)), key=len)
            # 整行同一個字碼且夠長 ＝ 框線（點陣剛好等於某個製表符），不是文字。
            if best and (len(best) < 4 or len({c['code'] for c in best}) > 1):
                out.append({'y': y, 'cells': best})
    return out

def span(line):
    xs = [c['x'] for c in line['cells']]
    ws = [c['w'] for c in line['cells']]
    return min(xs), max(x + w for x, w in zip(xs, ws))

def merge(acc, cand):
    """同一塊區域裡的行不只一種相位（上半 y ≡ 1、下半 y ≡ 4），
    所以要把所有相位的結果併起來，再把**佔到同一塊位置**的行去掉，
    只留字數多的那一個。只取單一最佳相位會整片漏掉另一種。"""
    lo, hi = span(cand)
    for i, cur in enumerate(acc):
        if abs(cur['y'] - cand['y']) >= 15:
            continue
        clo, chi = span(cur)
        if lo < chi and clo < hi:
            if len(cand['cells']) > len(cur['cells']):
                acc[i] = cand
            return
    acc.append(cand)

lines = []
for area in TEXT_AREAS:
    for ink in TEXT_INKS:
        got = []
        for ph in range(20):
            for l in scan(area, ph, ink):
                l['ink'] = ink
                merge(got, l)
        lines += got
lines.sort(key=lambda l: (l['y'], l['cells'][0]['x']))

def box(bx0, by0, bx1, by1, sample=None, corner=12):
    """量一個小框。

    `sample` 是取填色的位置：預設是框的正中央，但命令列那一條的正中央
    落在嵌進去的按鈕裡，量到的會是按鈕的顏色而不是框自己的。"""
    sx, sy = sample or ((bx0 + bx1) // 2, (by0 + by1) // 2)
    fill = rows[sy][sx]
    lay, alt = measure_frame(rows, bx0, by0, bx1, by1, fill=fill)
    return {'x': bx0, 'y': by0, 'w': bx1 - bx0 + 1, 'h': by1 - by0 + 1,
            'layers': lay, 'alt': alt, 'fill': fill,
            'patches': frame_patches(rows, bx0, by0, bx1, by1, lay, alt,
                                     fill=fill, corner=corner)}

spec = {
    'canvas': {'width': W, 'height': H, 'mode': 'VGA 12h 640x480x16'},
    'background': bg,
    'map': {'source': 'STRMAP.BIT', 'regions': 'STRMAP.RGN', 'x': 0, 'y': 40},
    'windows': {'board': {
        'x': x0, 'y': y0, 'w': x1 - x0 + 1, 'h': y1 - y0 + 1,
        'frame': BASE['frame'],
        'ink': 7, 'lines': lines,
        'boxes': [box(*b[:4], b[4] if len(b) > 4 else None,
                      b[5] if len(b) > 5 else 12)
                  for b in BOXES + SMALL_BOXES],
        # 命令列框下緣兩端的斜切：框模型只認四條邊與角落修正，
        # 這兩段落在框外一列，補成純色線。
        'rules': [{'x': 328, 'y': 152, 'w': 4, 'h': 1, 'colour': 3},
                  {'x': 627, 'y': 152, 'w': 4, 'h': 1, 'colour': 3}],
    }},
    'order': [],
    # 游標畫在 (300,200)（遊戲座標，畫面上是 y+40），但右鍵關掉報告視窗
    # 之後整片重畫，只補回了左半邊——原版的游標是疊上去的，
    # 被蓋掉的部分要等下次移動才回來（`docs/re/09`）。
    'mouse': {'x': 300, 'y': 200, 'clip': {'x': 302, 'y': 0, 'w': 6, 'h': 480}},
    'window_origin_grid': [21, 4],
    'hotspots': [
        {'x': 352, 'y': 80, 'w': 32, 'h': 20, 'grid': [1, 0],
         'action': 'strategy-menu', 'note': '戰略命令（docs/re/16）'},
        {'x': 448, 'y': 80, 'w': 32, 'h': 20, 'grid': [7, 0],
         'action': 'list-menu', 'note': '一覽表'},
        {'x': 544, 'y': 80, 'w': 32, 'h': 20, 'grid': [13, 0],
         'action': 'system-menu', 'note': '系統指令'},
    ],
}
out = f'/work/docs/internals/data/{outname}.json'
json.dump(spec, open(out, 'w'), ensure_ascii=False, indent=1)
print('寫出', out, '｜文字', sum(len(l['cells']) for l in lines), '格、小框', len(BOXES), '個')
