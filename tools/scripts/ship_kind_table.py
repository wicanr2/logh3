"""艦種屬性表：`ds:4E3A`，每個陣營一筆 208 位元組，內含 8 個艦種各 26。

    tools/py.sh tools/scripts/ship_kind_table.py <記憶體dump> [輸出.json]

索引是「陣營 × 208 ＋ 艦種 × 26」——陣營來自艦艇記錄的 `+0x00`、
艦種來自 `+0x04`（`docs/re/86`）。每筆 13 個 word，已解出的是
第 0 個（滿編艘數）與第 11、12 個（攻擊係數的分母用的兩個值）。
"""
import json
import pathlib
import struct
import sys

DS = 0x60BA4 - 0xC074
TABLE = 0x4E3A
FACTIONS, KINDS, WORDS = 2, 8, 13

mem = pathlib.Path(sys.argv[1]).read_bytes()
out = []
for f in range(FACTIONS):
    kinds = []
    for k in range(KINDS):
        off = DS + TABLE + f*KINDS*WORDS*2 + k*WORDS*2
        w = list(struct.unpack_from('<%dH' % WORDS, mem, off))
        kinds.append({'kind': k, 'full_ships': w[0], 'coef16': w[11],
                      'coef18': w[12], 'words': w})
        print('陣營 %d 艦種 %d  滿編 %5d  係數 %3d／%3d  %s'
              % (f, k, w[0], w[11], w[12], w))
    out.append({'faction': f, 'kinds': kinds})

if len(sys.argv) > 2:
    pathlib.Path(sys.argv[2]).write_text(json.dumps({
        '_note': '艦種屬性表，來源 ds:4E3A（docs/re/86、docs/internals/40）。'
                 '每個陣營一筆 208 位元組，內含 8 個艦種各 26 位元組 ＝ 13 個 word。'
                 '已解出的欄位：words[0] 滿編艘數、words[11]／words[12] 是'
                 '攻擊係數分母用的兩個值；其餘還沒解。',
        'factions': out,
    }, ensure_ascii=False, indent=1) + '\n', encoding='utf-8')
    print('寫出 %s' % sys.argv[2])
