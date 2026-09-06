#!/usr/bin/env python3
"""把熱區清單與按鈕標籤對起來。

每一筆熱區的 `+8` 是 far 指標，指到一段 `\\0` 結尾的 Big5 字串——
那就是**畫在那個按鈕上的字**。少數指標是 `FF 17 00`／`FF 18 00`
這種三個位元組的動作碼（建造計劃的上下箭頭），或是 null。
"""
import json, os

SRC = '/work/workplace/explore/ds54b3.bin'
IN = '/work/docs/internals/data/hotspot-lists.json'
d = open(SRC, 'rb').read()
data = json.load(open(IN, encoding='utf-8'))


def label(off):
    if off + 2 > len(d):
        return None
    if d[off] == 0xFF:                       # FF 17 00 / FF 18 00：動作碼
        return '動作碼 %02X' % d[off + 1]
    e = off
    while e < len(d) and d[e] != 0 and e - off < 24:
        e += 1
    s = d[off:e]
    if not s:
        return None
    try:
        t = s.decode('big5hkscs')
    except UnicodeDecodeError:
        return None
    return t if t.strip() else None


for L in data['lists']:
    for e in L['entries']:
        seg, off = e['ptr'].split(':')
        e['label'] = label(int(off, 16)) if seg.upper() == '54B3' else None
    L['labels'] = [e['label'] for e in L['entries']]

json.dump(data, open(IN, 'w', encoding='utf-8'), ensure_ascii=False, indent=1)
for L in data['lists']:
    names = [x if x else '—' for x in L['labels']]
    print('DS:%04X（%d 筆）%s' % (L['ds_offset'], L['count'], '、'.join(names)))
