# 每一筆熱區的 +8 是 far 指標。看它指到什麼——同一份清單的指標多半
# 落在清單附近，內容是短短幾個 byte 的動作碼。
import json, struct
d = open('/work/workplace/explore/ds54b3.bin', 'rb').read()
lists = json.load(open('/work/docs/internals/data/hotspot-lists.json', encoding='utf-8'))['lists']
for L in lists:
    offs = []
    for e in L['entries']:
        seg, off = e['ptr'].split(':')
        offs.append((int(seg, 16), int(off, 16)))
    segs = {s for s, _ in offs}
    same = segs == {0x54B3}
    print('DS:%04X（%d 筆）指標段 %s%s' % (
        L['ds_offset'], L['count'],
        ' '.join('%04X' % s for s in sorted(segs)),
        '' if same else '  ← 有跨段'))
    for (s, o), e in zip(offs, L['entries']):
        if s == 0x54B3 and o + 6 <= len(d):
            print('    (%2d,%2d) → %04X: %s' % (e['grid'][0], e['grid'][1], o,
                                                d[o:o + 6].hex(' ')))
        else:
            print('    (%2d,%2d) → %04X:%04X（不在這一段）' % (e['grid'][0], e['grid'][1], s, o))
