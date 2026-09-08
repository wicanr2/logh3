"""把資料段裡的艦種表倒成 JSON。

    tools/py.sh tools/scripts/shiptype_export.py <記憶體 dump> [dump 的起始線性位址]

表在 `ds:4E3A`，**8 筆 × 26 bytes × 2 陣營**（帝國在前）。第 8 筆不是
畫面上那七種艦艇，語意未解，一併倒出來。

目前只確認 `+0x00` ＝ 一單位幾艘（與 `SHIP.DT?` 的 `+0x08` 相同，也就是
艦隊組成畫面上的「數量」欄，`docs/re/69`）。其餘欄位倒原始位元組，
不猜語意——**猜的東西不進 JSON**。

dump 用 `tools/oracle.sh <批次> … -dump-mem lin:0:1048576:/src/workplace/<檔>.bin`
產生；起始位址預設 0（整份記憶體）。
"""
import json
import struct
import sys

DS = 0x54B30
BASE = 0x4E3A
RECORD = 26
NAMES = ['戰艦', '高速戰艦', '巡航艦', '航空母艦', '驅逐艦', '運補艦', '修護艦', None]

path = sys.argv[1]
origin = int(sys.argv[2], 0) if len(sys.argv) > 2 else 0
d = open(path, 'rb').read()

out = {
    'note': ('艦種表，`ds:4E3A` 起 8 筆 × 26 bytes × 2 陣營（帝國在前）。'
             '`+0x00` ＝ 一單位幾艘（confirmed，`docs/re/69`）；其餘欄位語意未解，'
             '倒的是原始位元組。第 8 筆不是畫面上那七種艦艇。'),
    'ds_offset': '0x%04X' % BASE,
    'record_bytes': RECORD,
    'rows': [n for n in NAMES if n],
    'ships_per_unit': [],
    'records': [],
}
for f in range(2):
    per = []
    for t in range(8):
        o = DS + BASE + (f * 8 + t) * RECORD - origin
        rec = d[o:o + RECORD]
        if len(rec) != RECORD:
            sys.exit(f'dump 太短：讀不到陣營 {f} 第 {t} 筆')
        per.append(struct.unpack_from('<H', rec, 0)[0])
        out['records'].append({
            'faction': f, 'type': t, 'name': NAMES[t],
            'ships_per_unit': struct.unpack_from('<H', rec, 0)[0],
            'bytes': rec.hex(' '),
        })
    out['ships_per_unit'].append(per[:7])

json.dump(out, open('/work/docs/internals/data/ship-type-table.json', 'w'),
          ensure_ascii=False, indent=1)
print('一單位幾艘：帝國', out['ships_per_unit'][0], '同盟', out['ships_per_unit'][1])
print('倒出 16 筆到 docs/internals/data/ship-type-table.json')
