"""把兩張旗艦名指標表解成 JSON。

    tools/py.sh tools/scripts/flagship_export.py

輸入是執行期資料段的傾印（`workplace/dsFull.bin`，線性 `54B30` 起 64 KB，
用 `tools/oracle.sh ... -dump-mem lin:54B30:65536:...` 取得）。

資料段 `0x66A8` 與 `0x6748` 各是 **16 個 far 指標**，指向 `0x6570` 起那批
旗艦名：前者帝國、後者同盟。提督詳細資料那張畫面印的「旗艦」就是
`表[陣營][該陣營內的序號]`（`docs/re/65`）。
"""
import json

BASE = 0x54B30
SEG = 0x54B3
TABLES = {0: 0x66A8, 1: 0x6748}
NAMES_AT = 0x6570

b = open('/work/workplace/dsFull.bin', 'rb').read()


def cstr(off):
    j = b.find(b'\0', off)
    return b[off:j]


out = {
    'note': ('提督詳細資料那張畫面的「旗艦」。資料段 0x66A8（帝國）與 0x6748'
             '（同盟）各 16 個 far 指標，指向 0x6570 起的旗艦名；'
             '索引是該提督在自己陣營裡的序號（帝國 = 提督索引、'
             '同盟 = 提督索引 − 40）。見 docs/re/65。'),
    'source': {'segment': 'GIN3PS.EXE 資料段', 'dump': 'lin:54B30:65536',
               'tables': {'empire': '0x66A8', 'alliance': '0x6748'},
               'names': '0x6570'},
    'tables': [],
}
for f in (0, 1):
    rows = []
    for k in range(16):
        o = TABLES[f] + k * 4
        off = int.from_bytes(b[o:o+2], 'little')
        seg = int.from_bytes(b[o+2:o+4], 'little')
        if seg != SEG or off == 0:
            rows.append(None)
            continue
        raw = cstr(off)
        codes = [raw[i] << 8 | raw[i+1] for i in range(0, len(raw) - 1, 2)]
        rows.append({'slot': k, 'offset': '0x%04X' % off,
                     'text': raw.decode('big5', 'replace'), 'codes': codes})
    out['tables'].append(rows)

json.dump(out, open('/work/docs/internals/data/flagships.json', 'w'),
          ensure_ascii=False, indent=1)
print('帝國', [r['text'] for r in out['tables'][0] if r])
print('同盟', [r['text'] for r in out['tables'][1] if r])
