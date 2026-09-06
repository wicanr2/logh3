#!/usr/bin/env python3
"""把 25 份熱區清單各自對到「登錄它的那個視窗」。

登錄動作長這樣（位元組簽章，段值不可靠）：

    1E BB <清單off> 53          push ds; mov bx,清單; push bx
    <推視窗指標的兩個 push>       FF 76 xx FF 76 xx   或   FF 36 xx xx FF 36 xx xx
                                 或 1E BB <視窗off> 53
    9A 0F 00 A9 37              call far（登錄常式）

掃整份記憶體找 `9A 0F 00 A9 37`，往回讀 push，就知道哪一份清單登錄給哪個視窗。
"""
import json, re, struct, sys

MEM = sys.argv[1] if len(sys.argv) > 1 else '/work/workplace/explore/mem640k.bin'
CALL = bytes.fromhex('9a0f00a937')
d = open(MEM, 'rb').read()
lists = {L['ds_offset'] for L in json.load(
    open('/work/docs/internals/data/hotspot-lists.json', encoding='utf-8'))['lists']}

hits, i = [], 0
while True:
    i = d.find(CALL, i)
    if i < 0:
        break
    back = d[max(0, i - 24):i]
    # 找 push ds; mov bx,imm16; push bx
    pushes = [struct.unpack_from('<H', back, m.start() + 2)[0]
              for m in re.finditer(rb'\x1e\xbb..\x53', back, re.S)]
    globs = [struct.unpack_from('<H', back, m.start() + 2)[0]
             for m in re.finditer(rb'\xff\x36..', back, re.S)]
    locals_ = [back[m.start() + 2] for m in re.finditer(rb'\xff\x76.', back, re.S)]
    hits.append({'at': i, 'imm_ptrs': pushes, 'globals': globs, 'locals': locals_})
    i += 1

print('登錄呼叫 %d 處' % len(hits))
mapped = {}
for h in hits:
    lst = next((p for p in h['imm_ptrs'] if p in lists), None)
    if lst is None:
        continue
    win = [p for p in h['imm_ptrs'] if p != lst]
    mapped.setdefault(lst, []).append(
        {'code_at': '%05X' % h['at'],
         'window_static': ['%04X' % w for w in win],
         'window_global': ['%04X' % g for g in h['globals']],
         'window_local': ['bp+%d' % l for l in h['locals']]})
print('對到清單的有 %d 份：' % len(mapped))
for k in sorted(mapped):
    for m in mapped[k]:
        print('  清單 %04X ← 視窗 靜態%s 全域%s 區域%s（碼在 %s）'
              % (k, m['window_static'] or '－', m['window_global'] or '－',
                 m['window_local'] or '－', m['code_at']))
json.dump({'note': '熱區清單 ↔ 登錄它的視窗；簽章 9A 0F 00 A9 37',
           'calls': len(hits), 'map': {('%04X' % k): v for k, v in mapped.items()}},
          open('/work/docs/internals/data/hotspot-windows.json', 'w', encoding='utf-8'),
          ensure_ascii=False, indent=1)
