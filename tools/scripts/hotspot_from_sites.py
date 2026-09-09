# -*- coding: utf-8 -*-
"""從**呼叫點**去解熱區清單，補進 `data/hotspot-lists.json`（`docs/re/156`）。

    tools/py.sh tools/scripts/hotspot_from_sites.py <記憶體傾印> [--write]

`hotspot-lists.json` 原本是拿「長得像熱區清單」的形狀去掃資料段掃出來的，
**那個掃法有洞**：27 個命中測試呼叫點寫死的清單位移裡有八個掃不到。
這支改成用 `hittest-sites.json` 的位移當種子逐一解開——呼叫點是程式自己
寫死的位址，不會漏。

只**新增**缺的清單、只**補**既有條目缺的欄位（`kind`），不改既有的標籤與
`window`（那些是人工標過的，`docs/re/19` 踩過一次覆蓋事故）。
"""
import json
import pathlib
import struct
import sys

DS = 0x54B30                      # 資料段的線性位址（同一次執行內固定）
REC = 12                          # 一筆熱區 12 bytes（docs/internals/23）
CELL_W, CELL_H = 16, 20

root = pathlib.Path('/work')
dump = sys.argv[1] if len(sys.argv) > 1 else 'workplace/s3fight.bin'
write = '--write' in sys.argv
mem = (root / dump).read_bytes()
sites = json.loads((root / 'docs/internals/data/hittest-sites.json').read_text(encoding='utf-8'))
path = root / 'docs/internals/data/hotspot-lists.json'
doc = json.loads(path.read_text(encoding='utf-8'))


def big5(off):
    end = mem.index(b'\0', DS + off, DS + off + 128)
    return mem[DS + off:end].decode('big5', 'replace')


def sane(e):
    """看起來是不是真的一筆熱區（畫面是 40×24 格）。"""
    return (e['enabled'] in (0, 1) and 1 <= e['size'][0] <= 40 and 1 <= e['size'][1] <= 24
            and e['grid'][0] <= 40 and e['grid'][1] <= 24)


def decode(off, cap=64):
    out, p = [], DS + off
    for _ in range(cap):
        if mem[p] == 0xFF and mem[p + 1] == 0xFF:
            break
        po, ps = struct.unpack_from('<HH', mem, p + 8)
        e = {
            'grid': [mem[p], mem[p + 1]],
            'size': [mem[p + 2], mem[p + 3]],
            'enabled': mem[p + 4],
            'index': mem[p + 5],
            'kind': mem[p + 6],          # 37A9:004C 依它分派（docs/re/156）
            'ptr': '%04X:%04X' % (ps, po),
        }
        if ps == DS >> 4:
            e['label'] = big5(po)
        else:
            e['action'] = mem[p + 8:p + 11].hex(' ').upper()
        out.append(e)
        p += REC
    return out


refs = sorted({int(x['list'], 16)
               for v in sites.values() if isinstance(v, list)
               for x in v if x.get('list')})
known = {l['ds_offset']: l for l in doc['lists']}
missing = [o for o in refs if o not in known]
print('呼叫點引用的清單 %d 份，既有 JSON %d 份，缺 %d 份：%s'
      % (len(refs), len(doc['lists']), len(missing), [hex(o) for o in missing]))

empty = []
for o in missing:
    ent = decode(o)
    if not ent or not all(sane(e) for e in ent):
        # 這份傾印裡沒有內容——那份清單是**執行期才填**的。
        empty.append(o)
        continue
    doc['lists'].append({
        'ds_offset': o,
        'count': len(ent),
        'entries': ent,
        'pixels': [{'x': e['grid'][0] * CELL_W, 'y': e['grid'][1] * CELL_H,
                    'w': e['size'][0] * CELL_W, 'h': e['size'][1] * CELL_H} for e in ent],
        'labels': [e.get('label') for e in ent],
        'from_call_site': True,
    })
    print('\nds:%04X（%d 筆）' % (o, len(ent)))
    for e in ent:
        print('   格(%2d,%2d) %d×%d 啟用%d 索引%2d 種類%d  %s'
              % (e['grid'][0], e['grid'][1], e['size'][0], e['size'][1],
                 e['enabled'], e['index'], e['kind'], e.get('label') or e.get('action')))

if empty:
    print('\n這份傾印裡沒有內容的（執行期才填）：%s' % [hex(o) for o in empty])

# 既有條目補 kind（只加欄位，不動標籤）
patched = 0
for l in doc['lists']:
    if l.get('from_call_site'):
        continue
    fresh = decode(l['ds_offset'])
    for e, f in zip(l['entries'], fresh):
        if 'kind' not in e and e['grid'] == f['grid'] and e['size'] == f['size']:
            e['kind'] = f['kind']
            patched += 1
print('\n既有條目補上 kind 的筆數：%d' % patched)

doc['lists'].sort(key=lambda l: l['ds_offset'])
doc['count'] = len(doc['lists'])
if write:
    path.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + '\n', encoding='utf-8')
    print('已寫回 %s，共 %d 份' % (path, doc['count']))
else:
    print('（沒有 --write，沒有寫檔）')
