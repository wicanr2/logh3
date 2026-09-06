#!/usr/bin/env python3
"""`STELLAR.DT[89AX]` — 星域資料。

4,032 bytes ＝ **32 筆 × 126 bytes**，沒有檔頭。32 筆正好對上
`STRMAP.RGN` 的 32 個星域，名字也對得上地圖上印的字。

每一筆開頭是 Big5（倚天）名稱，`\\0` 結束、後面補 0。其餘欄位的語意
逐欄比對四個版本（`DT8`／`DT9`／`DTA`／`DTX`，應該是四個劇本）
才有得談——這裡先把資料原樣倒出來。
"""
import json, os, struct

STRIDE, COUNT = 126, 32
OUT = '/work/docs/internals/data/stellar.json'
vers = {}
for ext in ('DT8', 'DT9', 'DTA', 'DTX'):
    p = '/orig/STELLAR.' + ext
    if not os.path.exists(p):
        continue
    d = open(p, 'rb').read()
    recs = []
    for i in range(COUNT):
        r = d[i * STRIDE:(i + 1) * STRIDE]
        e = 0
        while e < len(r) and r[e] not in (0, 0xFF):
            e += 1
        recs.append({
            'index': i,
            'name': r[:e].decode('big5hkscs', errors='replace'),
            'name_len': e,
            'raw': r.hex(),
            'words': list(struct.unpack_from('<%dH' % ((STRIDE - 24) // 2), r, 24)),
        })
    vers[ext] = recs

base = vers.get('DTX') or next(iter(vers.values()))
# 哪些位移在四個版本之間會變？不變的比較像固定屬性。
varies = set()
if len(vers) > 1:
    for i in range(COUNT):
        raws = [bytes.fromhex(v[i]['raw']) for v in vers.values()]
        for k in range(STRIDE):
            if len({r[k] for r in raws}) > 1:
                varies.add(k)

json.dump({'source': 'STELLAR.DT[89AX]', 'stride': STRIDE, 'count': COUNT,
           'note': '32 筆星域資料；名稱 confirmed，其餘欄位語意未確認',
           'varies_between_versions': sorted(varies),
           'versions': vers},
          open(OUT, 'w', encoding='utf-8'), ensure_ascii=False, indent=1)
print('%d 個版本 × %d 筆 → %s' % (len(vers), COUNT, OUT))
print('星域名（DTX）：')
for r in base:
    print('  %2d %s' % (r['index'], r['name']))
print('四個版本之間會變的位移：', sorted(varies))
