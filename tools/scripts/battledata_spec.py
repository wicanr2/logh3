#!/usr/bin/env python3
"""把戰術選單「資料」那張畫面量成規格（`docs/re/111`）。

    tools/py.sh tools/scripts/battledata_spec.py [--emit]

不帶 --emit 只印量到的層；帶了才寫 docs/internals/data/battle-data.json。
"""
import json, sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import W, H, load_screen, measure_frame, frame_patches

rows, _ = load_screen('/work/workplace/oracle/s3data/screen.bin')
emit = '--emit' in sys.argv


def frame(x0, y0, x1, y1, fill, sample, corner=10, exclude=(), label='', layers_max=16):
    lay, alt = measure_frame(rows, x0, y0, x1, y1, fill=fill, sample=sample,
                             layers_max=layers_max)
    pat = frame_patches(rows, x0, y0, x1, y1, lay, alt, fill=fill,
                        corner=corner, exclude=exclude)
    print('%-10s (%d,%d) %dx%d fill=%d 層=%d 修正=%d' %
          (label, x0, y0, x1-x0+1, y1-y0+1, fill, len(lay), len(pat)))
    for i, (l, a) in enumerate(zip(lay, alt)):
        print('    %2d 上%2d 右%2d 下%2d 左%2d   alt %s' % (i, l[0], l[1], l[2], l[3], a))
    if pat:
        print('    修正', pat)
    return {'layers': [list(map(int, l)) for l in lay],
            'alt': [list(map(int, a)) for a in alt],
            'fill': fill,
            'patches': [{'dx': p['dx'], 'dy': p['dy'], 'color': int(p['color'])} for p in pat]}


spec = {}
spec['window'] = {'x': 16, 'y': 60, 'w': 447, 'h': 360,
                  'frame': frame(16, 60, 462, 419, 9, (300, 110), label='視窗')}
spec['tab'] = {'y': 80, 'w': 66, 'h': 20, 'xs': [32, 112, 192],
               'frame': frame(32, 80, 97, 99, 8, (60, 84), corner=6,
                              exclude=[(36, 85, 11, 11), (48, 82, 66, 15)], label='標籤框')}
spec['row'] = {'x': 32, 'y': 120, 'w': 402, 'h': 20, 'count': 14,
               'frame': frame(32, 300, 433, 319, 8, (200, 305), corner=6, label='表列')}
# 捲軸那一欄是**三塊**：上下各 20（與表列同高），中間 240。
spec['scroll'] = {'x': 434, 'w': 14,
                  'boxes': [{'y': 120, 'h': 20}, {'y': 140, 'h': 240}, {'y': 380, 'h': 20}],
                  'frame': frame(434, 140, 447, 379, 8, (440, 200), corner=6, label='捲軸')}

# 標籤裡的小方塊：11 × 11。外凸一層（上左 7、下右 3）→ 一圈色 15 →
# 內凹一層（上左 3、下右 7）→ 中間 5 × 5：被選中的填色 2，沒選的填 8。
spec['tab']['mark'] = {'dx': 4, 'dy': 5, 'w': 11, 'h': 11,
                       # 三層之後就是中間那 5 × 5：沒選填 9（視窗底色），選中填 2。
                       'frame': frame(116, 85, 126, 95, 9, (120, 88), corner=5,
                                      label='小方塊', layers_max=3),
                       'on': 2}


def cells(text, x):
    """把一條字串攤成 16 像素的格。

    原版的文字引擎一格 16 像素：位元組 ≥ 0x80 是一個全形字，< 0x80 是
    **兩個半形字塞在同一格**（`docs/internals/31`）。加粗是整格一起
    `v | v>>1`，所以左半那個字的最右一行會溢進右半的地盤——拆成兩個
    8 像素的格去畫，那一行就不見了（`BHCADTE` 的 E 右邊少一豎）。
    """
    b, out = text.encode('big5'), []
    i = 0
    while i < len(b):
        if b[i] >= 0x80:
            out.append({'x': x, 'w': 16, 'code': (b[i] << 8) | b[i+1]}); i += 2
        else:
            lo = b[i+1] if i + 1 < len(b) and b[i+1] < 0x80 else 0x20
            out.append({'x': x, 'w': 16, 'code': (b[i] << 8) | lo})
            i += 2 if i + 1 < len(b) and b[i+1] < 0x80 else 1
        x += 16
    return out


# 三條表頭字串逐位元組抄自記憶體（線性 0x580D5 一帶，`docs/re/111`）。
HEADERS = [
    '艦隊資料                        艦數  最大  士氣',
    '行星資料          經濟力  防衛力  補給  BHCADTE ',
    '敵艦隊資料                      艦數  最大  士氣',
]
LABELS = ['艦隊  ', '行星  ', '敵艦隊']
for h in HEADERS:
    assert len(h.encode('big5')) == 48, (h, len(h.encode('big5')))

spec['text'] = {'x': 48, 'header_y': 102, 'row_dy': 2,
                'ink': 7, 'lead_ink': 6, 'shadow': 3}
spec['headers'] = [{'text': h, 'cells': [c for c in cells(h, 48)
                                         if c['code'] != 0x2020]} for h in HEADERS]
spec['tab']['labels'] = [{'text': t, 'cells': [c for c in cells(t, spec['tab']['xs'][i] + 16)
                                               if c['code'] != 0x2020]}
                         for i, t in enumerate(LABELS)]
print()
print('表頭字的 x：')
for i, h in enumerate(spec['headers']):
    print('  %d %s' % (i, [(c['x'], c['w']) for c in h['cells']][:12]))

print()
print('標籤小方塊 (37,86)–(45,94)：')
for y in range(84, 97):
    print('  y=%3d %s' % (y, ''.join('%X' % rows[y][x] for x in range(34, 49))))

if emit:
    json.dump(spec, open('/work/docs/internals/data/battle-data.json', 'w'),
              ensure_ascii=False, indent=1)
    print('\n已寫出 battle-data.json')
