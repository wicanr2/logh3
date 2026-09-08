"""戰場入口表：每個星域列出「從哪個鄰域進來、佈在哪一格」。

    tools/py.sh tools/scripts/deploy_table.py <記憶體dump> [輸出.json]

原版在 `113A:0D2C`（`docs/re/84`）：拿星域編號去 `ds:4B2E` 起的 far
指標陣列取一張表，表裡一筆 4 個位元組——鍵（來源星域）、格欄、格列、
一個永遠是 0 的位元組，`0xFF` 結尾。找到之後艦隊的座標是
（欄 × 16, 列 × 20）。
"""
import json
import pathlib
import struct
import sys

FLEET = 0x60BA4
DS = FLEET - 0xC074            # ds 段基底
TABLES = 0x4B2E                # ds:4B2E 起的 far ptr 陣列
REGIONS = 32
TILE_W, TILE_H = 16, 20

mem = pathlib.Path(sys.argv[1]).read_bytes()
out = []
for i in range(REGIONS):
    off, seg = struct.unpack_from('<HH', mem, DS + TABLES + i * 4)
    lin = seg * 16 + off
    entries = []
    p = lin
    while p + 4 <= len(mem) and mem[p] != 0xFF:
        entries.append({
            'from': mem[p],
            'col': mem[p + 1],
            'row': mem[p + 2],
            'x': mem[p + 1] * TILE_W,
            'y': mem[p + 2] * TILE_H,
        })
        p += 4
    out.append({'region': i, 'entries': entries})
    print('星域 %2d：%s' % (i, '  '.join(
        '來自 %2d → 格 (%3d,%3d) ＝ (%4d,%4d)' % (e['from'], e['col'], e['row'], e['x'], e['y'])
        for e in entries)))

if len(sys.argv) > 2:
    doc = {
        '_note': '戰場入口表，來源 113A:0D2C 的表（docs/re/84、docs/internals/39）。'
                 'x/y ＝ 欄×16、列×20，就是艦隊記錄 +0x5C／+0x5E 的值。',
        'tile_w': TILE_W,
        'tile_h': TILE_H,
        'regions': out,
    }
    pathlib.Path(sys.argv[2]).write_text(
        json.dumps(doc, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    print('寫出 %s' % sys.argv[2])
