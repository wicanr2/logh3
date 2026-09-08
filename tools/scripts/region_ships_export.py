"""把開局各星域的艦艇艘數倒成 JSON。

    tools/py.sh tools/scripts/region_ships_export.py <記憶體 dump> [dump 起始線性位址]

星域記錄 `+0x30` 起是 14 個 32 位元的艘數（帝國七個、同盟七個，
`docs/re/69`）。**這一段在 `STELLAR.DT*` 檔案裡是全 0**，載入時才填進去，
填的那一步還沒解——所以這份 JSON 是**量到的開局狀態**，不是從檔案算出來的。

同一個位置在四個 `STELLAR.DT*` 檔案裡都是 0，所以換劇本要另外量。
"""
import json
import struct
import sys

DS = 0x54B30
TABLE = 0xA79C          # ds 位移
RECORD = 126
COUNT = 32
SHIPS = 0x30
NAMES = ['戰艦', '高速戰艦', '巡航艦', '航空母艦', '驅逐艦', '運補艦', '修護艦']

path = sys.argv[1]
origin = int(sys.argv[2], 0) if len(sys.argv) > 2 else 0
d = open(path, 'rb').read()

base = DS + TABLE - origin
regions = []
for r in range(COUNT):
    o = base + r * RECORD
    dw = [struct.unpack_from('<I', d, o + SHIPS + 4 * i)[0] for i in range(14)]
    regions.append({'region': r, 'owner': d[o + 0x12],
                    'empire': dw[:7], 'alliance': dw[7:]})

out = {
    'note': ('開局（796 年 1 月）各星域的艦艇艘數，量自 dosgolem 的記憶體 dump。'
             '星域記錄 `+0x30` 起 14 個 32 位元：帝國七個、同盟七個（`docs/re/69`）。'
             '`STELLAR.DT*` 檔案裡這一段是全 0，載入時才填——填的那一步未解，'
             '所以這是**量到的狀態**，不是算出來的。'),
    'scenario': 'S1 796/1（ADMIRAL.DTX 那一組）',
    'rows': NAMES,
    'ships_per_unit': [3000, 3000, 4000, 2000, 3000, 3000, 1000],
    'regions': regions,
}
json.dump(out, open('/work/docs/internals/data/region-ships-796-01.json', 'w'),
          ensure_ascii=False, indent=1)
nz = sum(1 for r in regions if any(r['empire']) or any(r['alliance']))
print(f'倒出 {COUNT} 個星域（{nz} 個有艦艇）到 docs/internals/data/region-ships-796-01.json')
