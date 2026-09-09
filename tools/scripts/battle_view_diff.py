# -*- coding: utf-8 -*-
"""比對兩份戰場傾印的鏡頭與艦艇記錄（`docs/re/160`）。

    tools/py.sh tools/scripts/battle_view_diff.py <前.bin> <後.bin>

兩份都要是 `-dump-mem lin:0:1048576` 的整份記憶體。回答三件事：

1. 艦艇記錄有沒有變（有變的逐位移列出來）；
2. 鏡頭 `ds:9BF4`／`ds:9BF6` 有沒有變；
3. 每個視野裡有哪些分隊、塊號 `+0x2F` 是多少（255 ＝ 不畫）。

**「畫面不一樣」有三種可能：畫法變了、資料變了、鏡頭變了。** 這支專門
用來排掉後兩種。
"""
import pathlib
import struct
import sys
from collections import Counter

REC = 78
FIELD_W, FIELD_H = 448, 360
# 「實行  \0取進度\0」——拿它反推資料段的線性位址（docs/re/157 的選單表）
DS_PAT = b'\xb9\xea\xa6\xe6\x20\x20\x00\xa8\xfa\xb6\x69\xab\xd7\x00'
DS_OFF = 0x0CFA


def u16(m, o):
    return struct.unpack_from('<H', m, o)[0]


def load(path):
    m = pathlib.Path(path).read_bytes()
    i = m.find(DS_PAT)
    if i < 0:
        raise SystemExit('%s：找不到資料段（那份傾印可能不在遊戲本體裡）' % path)
    ds = i - DS_OFF
    off, seg = struct.unpack_from('<HH', m, ds + 0xCFB4)   # 艦艇池的遠指標
    return m, ds, seg * 16 + off


def units(m, pool, n=220):
    out = []
    for k in range(n):
        r = pool + k * REC
        if r + REC > len(m) or m[r + 4] == 0xFF:
            continue
        out.append((k, m[r + 1], m[r + 4], m[r + 0x17], m[r + 0x2F],
                    u16(m, r + 0x34), u16(m, r + 0x36)))
    return out


a, b = sys.argv[1], sys.argv[2]
ma, dsa, pa = load(a)
mb, dsb, pb = load(b)
print('%s：資料段 %05X 艦艇池 %05X' % (a, dsa, pa))
print('%s：資料段 %05X 艦艇池 %05X' % (b, dsb, pb))

ca = (u16(ma, dsa + 0x9BF4), u16(ma, dsa + 0x9BF6))
cb = (u16(mb, dsb + 0x9BF4), u16(mb, dsb + 0x9BF6))
print('\n鏡頭：%s → %s%s' % (ca, cb, '' if ca != cb else '（沒變）'))

print('\n艦艇記錄有變的：')
changed = 0
for k in range(220):
    ra, rb = pa + k * REC, pb + k * REC
    if ra + REC > len(ma) or ma[ra + 4] == 0xFF:
        continue
    d = [o for o in range(REC) if ma[ra + o] != mb[rb + o]]
    if d:
        changed += 1
        print('  #%3d %s' % (k, ' '.join('+%02X:%02X→%02X'
                                         % (o, ma[ra + o], mb[rb + o]) for o in d)))
if changed == 0:
    print('  （一筆都沒有）')

for nm, m, c, pool in ((a, ma, ca, pa), (b, mb, cb, pb)):
    vis = [u for u in units(m, pool)
           if c[0] <= u[5] < c[0] + FIELD_W and c[1] <= u[6] < c[1] + FIELD_H]
    print('\n%s 鏡頭 %s：視野內 %d 筆，塊號分佈 %s'
          % (nm, c, len(vis), sorted(Counter(u[4] for u in vis).items())))
    for u in vis[:12]:
        print('   #%3d 艦隊%3d 艦種%d +0x17=%d +0x2F=%3d 世界(%4d,%4d) 螢幕(%3d,%3d)'
              % (u[0], u[1], u[2], u[3], u[4], u[5], u[6],
                 u[5] - c[0] + 16, u[6] - c[1] + 60))
