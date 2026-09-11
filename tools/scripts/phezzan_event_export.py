# -*- coding: utf-8 -*-
"""把費沙事件用到的常數從執行期傾印倒成 JSON（docs/internals/57）。

    tools/py.sh tools/scripts/phezzan_event_export.py > docs/internals/data/phezzan-event.json

讀 `workplace/s3datamem.bin`（1 MB，線性 0 起；DS ＝ 0x54B30）：
兩張鄰接清單（0xFF 結尾）、懲罰比例表（含越界讀到的第 6 格）、六組兩頁訊息的原文。
規則本身（誰呼叫誰、閘門）寫在 docs/internals/57，這裡只放資料。
"""
import json

DS = 0x54B30
mem = open('/work/workplace/s3datamem.bin', 'rb').read()


def w(off):
    return mem[DS + off] | mem[DS + off + 1] << 8


def cstr(seg, off):
    p = seg * 16 + off
    return mem[p:mem.index(b'\x00', p)].decode('big5')


def ff_list(off):
    out = []
    p = DS + off
    while mem[p] != 0xFF:
        out.append(mem[p])
        p += 1
    return out


def pages(off):
    return [cstr(w(off + 4 * k + 2), w(off + 4 * k)) for k in range(2)]


MSG = [
    ('3016', 'pick_target', 0, '選費沙當移動目標、ds:8F1D＝帝國（1213:0003）'),
    ('301E', 'pick_target', 1, '選費沙當移動目標、ds:8F1D＝同盟（1213:0003）'),
    ('3260', 'captured_owner', 0, '帝國拿下、給帝國（人類座位）看（1213:0096）'),
    ('3268', 'captured_owner', 1, '同盟拿下、給同盟（人類座位）看（1213:0096）'),
    ('3278', 'captured_opponent', 0, '帝國拿下、給同盟（人類座位）看（1213:0096）'),
    ('3270', 'captured_opponent', 1, '同盟拿下、給帝國（人類座位）看（1213:0096）'),
]

pct = [mem[DS + 0x328E + i] for i in range(6)]
data = {
    'source': 'workplace/s3datamem.bin（DS = 0x54B30）；規則見 docs/internals/57-phezzan-event.md',
    'evidence': ['docs/re/192-phezzan-event.md', 'docs/re/292-rebellion-decision-and-effect.md',
                 'docs/re/295-region-morale-writers.md', 'docs/re/296-phezzan-event-trigger.md'],
    'phezzan_region': 14,
    'phezzan_planet': 34,
    'transfer_percent': 80,
    'by_owner': {
        '0': {'target_region': 31, 'target_slice_index': 4, 'target_planet': 80,
              'rebel_list_ds': '3280', 'rebel_list': ff_list(0x3280)},
        'other': {'target_region': 0, 'target_slice_index': 0, 'target_planet': 0,
                  'rebel_list_ds': '3284', 'rebel_list': ff_list(0x3284)},
    },
    'penalty': {
        'only_if_owner_is_human_seat': True,
        'turn_limit': 600,
        'percent_by_turn_div_100': pct,
        'note': '索引 0–4 是 ds:328E 的五格；索引 5（回合 500–599）越界讀到 ds:3293 的第一個 byte',
        'morale_after': 100,
        'skips_region': 14,
    },
    'messages': [
        {'ds': m[0], 'kind': m[1], 'faction': m[2], 'when': m[3], 'pages': pages(int(m[0], 16))}
        for m in MSG
    ],
    'rebellion_line': {'ds': ['329B', '32A8'],
                       'text': [cstr(0x54B3, 0x329B), cstr(0x54B3, 0x32A8)]},
}
print(json.dumps(data, ensure_ascii=False, indent=2))
