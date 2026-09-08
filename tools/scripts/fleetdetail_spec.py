#!/usr/bin/env python3
"""把艦隊詳細資料那一層量成規格（`docs/re/112`）。

    tools/py.sh tools/scripts/fleetdetail_spec.py [--emit]
"""
import json, sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import load_screen, measure_frame, frame_patches

rows, _ = load_screen('/work/workplace/oracle/s3detail/screen.bin')
emit = '--emit' in sys.argv


def frame(x0, y0, x1, y1, fill, sample, corner=10, exclude=(), label='', layers_max=16):
    lay, alt = measure_frame(rows, x0, y0, x1, y1, fill=fill, sample=sample,
                             layers_max=layers_max)
    pat = frame_patches(rows, x0, y0, x1, y1, lay, alt, fill=fill,
                        corner=corner, exclude=exclude)
    print('%-8s (%d,%d) %dx%d fill=%d 層=%d 修正=%d' %
          (label, x0, y0, x1-x0+1, y1-y0+1, fill, len(lay), len(pat)))
    for i, (l, a) in enumerate(zip(lay, alt)):
        print('    %2d 上%2d 右%2d 下%2d 左%2d  alt %s' % (i, l[0], l[1], l[2], l[3], a))
    if pat:
        print('    修正', pat)
    return {'layers': [list(map(int, l)) for l in lay],
            'alt': [list(map(int, a)) for a in alt], 'fill': fill,
            'patches': [{'dx': p['dx'], 'dy': p['dy'], 'color': int(p['color'])} for p in pat]}


spec = {'window': {'x': 16, 'y': 60, 'w': 447, 'h': 360},
        'panel': {'x': 32, 'y': 80, 'w': 416, 'h': 320,
                  'frame': frame(32, 80, 447, 399, 9, (200, 90), corner=10,
                                 exclude=[(352, 100, 80, 100)], label='內層面板')},
        'face': {'x': 352, 'y': 100, 'w': 80, 'h': 100},
        'cell': {'x': 48, 'y': 102, 'col': 16, 'row': 20, 'half': 8, 'rows': 14}}


# ── 原版的格式字串（逐位元組抄自資料段，`docs/re/112`）───────────────
# 每一條都是印在字元格上的：一格 16 像素，全形佔一格、半形佔半格。
# 這裡把它走一遍，算出靜態字的 x 與每個欄位的 x／寬度，Go 那邊就不必
# 再實作一次 printf。
# `%W` 不是原版的格式——原版那一格印的是**一個全形數字**（部隊編號），
# 這裡借一個記號把它表達成「寬一格的全形欄位」。
FIELD = {'%W': ('wide', 2), '%2d': ('level', 2), '%3u': ('v3', 3), '%4d': ('n4', 4),
         '%6ld': ('n6', 6), '%6lu': ('n6', 6), '%-8s': ('s8', 8),
         '%-14s': ('s14', 14), '%s': ('s', 0)}

def walk(fmt, col, row):
    """走一條格式字串，回 (靜態段, 欄位段)。x 用像素、寬度用半形格數。"""
    x = spec['cell']['x'] + col * spec['cell']['col']
    y = spec['cell']['y'] + row * spec['cell']['row']
    statics, fields, cur, curx = [], [], [], x
    i = 0
    b = fmt.encode('big5')
    while i < len(b):
        if b[i] == 0x25:  # '%'
            j = i + 1
            while j < len(b) and chr(b[j]) in '-0123456789':
                j += 1
            while j < len(b) and chr(b[j]) in 'ldu':
                j += 1
            if j < len(b) and chr(b[j]) in 'sduW':
                j += 1
            tok = b[i:j].decode()
            if cur:
                statics.append({'x': curx, 'codes': cur}); cur = []
            name, w = FIELD[tok]
            fields.append({'x': x, 'w': w, 'fmt': tok})
            x += w * spec['cell']['half']
            curx = x
            i = j
            continue
        if b[i] >= 0x80:
            code = (b[i] << 8) | b[i+1]
            if code == 0xA241:      # 原版的全形斜線是 A1FE，不是 A241
                code = 0xA1FE
            if code != 0xA140:      # 全形空白不畫
                if not cur:
                    curx = x
                cur.append(code)
            elif cur:
                statics.append({'x': curx, 'codes': cur}); cur = []
            x += 16; i += 2
        else:
            if b[i] != 0x20:
                if not cur:
                    curx = x
                cur.append(b[i])
            elif cur:
                statics.append({'x': curx, 'codes': cur}); cur = []
            x += 8; i += 1
    if cur:
        statics.append({'x': curx, 'codes': cur})
    return {'y': y, 'static': statics, 'fields': fields}


# 一行一條（col, row, 格式）。格式逐位元組抄自資料段。
LINES = [
    ('faction', 0, 0, '%s'),
    ('fleet_no', 4, 0, '第%2d艦隊'),
    ('commander', 0, 1, '指揮官  %s'),
    ('cmd_level', 4, 2, '等級    %2d    經驗值  %6lu'),
    ('adjutant', 0, 3, '副官    %s'),
    ('adj_level', 4, 4, '等級    %2d    經驗值  %6lu'),
    ('grid0', 0, 5, '艦隊士氣值      統率力          光線砲        '),
    ('grid1', 0, 6, '回轉速度        運用            飛彈          '),
    ('grid2', 0, 7, '巡航速度        營運            戰鬥機        '),
    ('grid3', 0, 8, '戰鬥速度        占領            防禦            '),
    ('unit_no', 0, 9, '第%W部隊'),
    ('unit_ship', 4, 9, '  %-8s                %4d／  %4d隻'),
    ('u_morale', 0, 10, '部隊士氣值 %3u'),
    ('w_beam', 7, 10, '  %-14s  %6ld／%6ld門'),
    ('u_top', 0, 11, '最高速度   %3u'),
    ('w_missile', 7, 11, '  %-14s  %6ld／%6ld發'),
    ('u_battle', 0, 12, '戰鬥速度   %3u'),
    ('w_launch', 7, 12, '        發射裝置  %6ld／%6ld座'),
    ('u_dur', 0, 13, '耐久力     %3u'),
    ('w_fighter', 7, 13, '  %-14s  %6ld／%6ld架'),
]
spec['lines'] = {}
for name, col, row, fmt in LINES:
    spec['lines'][name] = walk(fmt, col, row)

# 十二個數值各自的格點（`docs/re/112`）。
VALUES = [
    ('fleet_morale', 5, 5, None), ('turn_speed', 5, 6, 13), ('cruise_speed', 5, 7, 12),
    ('battle_speed', 5, 8, 12),
    ('command', 13, 5, None), ('operation', 13, 6, 23), ('management', 13, 7, 14),
    ('occupy', 13, 8, 19),
    ('beam', 21, 5, 15), ('missile', 21, 6, 16), ('fighter', 21, 7, 17),
    ('defence', 21, 8, 18),
]
spec['values'] = [{'name': n, 'x': spec['cell']['x'] + c * spec['cell']['col'],
                   'y': spec['cell']['y'] + r * spec['cell']['row'], 'w': 4,
                   'source': src}
                  for n, c, r, src in VALUES]


def codes(t):
    b = t.encode('big5')
    out, i = [], 0
    while i < len(b):
        if b[i] >= 0x80:
            out.append((b[i] << 8) | b[i+1]); i += 2
        else:
            out.append(b[i]); i += 1
    return out

# 上下兩段之間那條線：色 15，一列高，正好 24 格寬。
spec['rules'] = [{'x': 48, 'y': 279, 'w': 384, 'h': 1, 'colour': 15}]

spec['ship_kinds'] = [{'text': t, 'codes': codes(t)} for t in
                      ['戰艦', '高速戰艦', '巡航艦', '航空母艦', '驅逐艦',
                       '運補艦', '修護艦', '偵察艦']]
spec['weapons'] = [{'text': t, 'codes': codes(t)} for t in
                   ['中子光線砲', '光子砲', '中子砲', '軌道砲', '防空砲',
                    '雷射水雷', '中子飛彈', '光子飛彈', '中子彈', '光子魚雷',
                    '王爾古雷', '斯巴達尼恩']]
spec['not_equipped'] = [{'row': 10, 'text': '  光線砲                    非搭載'},
                        {'row': 11, 'text': '  飛彈                      非搭載'},
                        {'row': 13, 'text': '  艦載機                    非搭載'}]
for e in spec['not_equipped']:
    e['line'] = walk(e['text'], 7, e['row'])

print()
for k, v in spec['lines'].items():
    print('%-10s y=%3d 靜態 %s 欄位 %s' % (k, v['y'],
          [(s['x'], len(s['codes'])) for s in v['static']],
          [(f['x'], f['w'], f['fmt']) for f in v['fields']]))

print()
print('頭像四周（x 350–432, y 98–200 的角落）：')
for y in (100, 101, 102, 103, 193, 194, 195):
    print('  y=%3d %s' % (y, ''.join('%X' % rows[y][x] for x in range(350, 434))))

if emit:
    json.dump(spec, open('/work/docs/internals/data/fleet-detail.json', 'w'),
              ensure_ascii=False, indent=1)
    print('\n已寫出 fleet-detail.json')
