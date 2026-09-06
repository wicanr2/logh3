#!/usr/bin/env python3
"""在資料段的傾印裡掃出所有「熱區清單」。

一筆 12 bytes（`docs/re/09`）：

    +0 格子 x   +1 格子 y   +2 寬（格）  +3 高（格）
    +4 啟用旗標 +5 項目索引 +6/+7 ?      +8 far 指標

清單以 `FF FF` 結束。判準：連續 N 筆（N≥1）都符合
「格子座標 < 40／< 30、寬高 1–8、啟用旗標 0 或 1、索引遞增」，
而且緊接著 `FF FF`。
"""
import struct, sys

SRC = sys.argv[1] if len(sys.argv) > 1 else '/work/workplace/explore/ds54b3.bin'
d = open(SRC, 'rb').read()


def rec_ok(r, want_index):
    gx, gy, w, h, en, idx = r[0], r[1], r[2], r[3], r[4], r[5]
    return (gx < 40 and gy < 30 and 1 <= w <= 12 and 1 <= h <= 12
            and en in (0, 1) and idx == want_index)


found = []
p = 0
while p + 14 <= len(d):
    n = 0
    q = p
    while q + 12 <= len(d) and rec_ok(d[q:q + 12], n):
        n += 1
        q += 12
    if n >= 1 and q + 2 <= len(d) and d[q] == 0xFF and d[q + 1] == 0xFF:
        recs = []
        for k in range(n):
            r = d[p + k * 12:p + (k + 1) * 12]
            off, seg = struct.unpack_from('<HH', r, 8)
            recs.append({'grid': [r[0], r[1]], 'size': [r[2], r[3]],
                         'enabled': r[4], 'index': r[5],
                         'ptr': '%04X:%04X' % (seg, off)})
        found.append((p, recs))
        p = q + 2
    else:
        p += 2

import json, os
out = [{'ds_offset': off, 'count': len(recs), 'entries': recs,
        'pixels': [{'x': (r['grid'][0]) * 16, 'y': (r['grid'][1]) * 20,
                    'w': r['size'][0] * 16, 'h': r['size'][1] * 20}
                   for r in recs]} for off, recs in found]
os.makedirs('/work/docs/internals/data', exist_ok=True)
json.dump({'note': '資料段裡所有「熱區清單」形狀的資料；像素座標是相對視窗原點的，'
                   '要加上該畫面視窗原點的格子座標才是畫面座標（docs/re/09）',
           'segment': 'DS=54B3（796 年 1 月，走到會計報告時）',
           'count': len(found), 'lists': out},
          open('/work/docs/internals/data/hotspot-lists.json', 'w', encoding='utf-8'),
          ensure_ascii=False, indent=1)
print('找到 %d 份清單 → docs/internals/data/hotspot-lists.json' % len(found))
for off, recs in found:
    px = ['(%d,%d) %dx%d' % (r['grid'][0], r['grid'][1], r['size'][0], r['size'][1])
          for r in recs]
    print('  DS:%04X  %d 筆  %s' % (off, len(recs), '  '.join(px)))
