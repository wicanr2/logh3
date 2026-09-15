#!/usr/bin/env python3
"""把戰略地圖上的艦隊視窗（`0DEA:0040`，`docs/re/312`、`docs/spec/051`）
用到的格式字串與版面抄成規格。

    tools/py.sh tools/scripts/fleetwindow_spec.py <1MB 傾印> <輸出.json>

格式字串逐位元組抄自資料段（DS＝54B3），順序就是 `0DEA:0040` 與它的
五支子常式呼叫 `3617:003A` 的順序；Go 那邊照同一個順序把值填進去
（`internal/screen/fleetwindow.go`）。視窗幾何來自建構參數
`3603:0006(0, 0x97, 7, 0, 25, 17, 1)`；鈕、頭像與作戰選單量自收據 `fcmd-win`／
`fcmd-menu`（字元平面 `plane.bin`）。
"""
import json, sys

DS = 0x54B3 * 16


def main(src, dst):
    m = open(src, 'rb').read()

    def cstr(off):
        p = DS + off
        return m[p:m.index(b'\0', p)]

    def hexs(off):
        return cstr(off).hex()

    def far(off):
        o = m[DS + off] | m[DS + off + 1] << 8
        sg = m[DS + off + 2] | m[DS + off + 3] << 8
        p = sg * 16 + o
        return m[p:m.index(b'\0', p)].hex()

    # 0DEA:0040 印的第一段（列 1–3）與分隔列；名字對應 docs/re/312 §二。
    formats = {
        'clear': hexs(0x22E2),          # \x0c
        'fleet_no': hexs(0x22E4),       # %s    第   （%s ＝ 陣營名，色 0x96）
        'fleet_no_tail': hexs(0x22ED),  # 艦隊\n
        'commander': hexs(0x22F3),      # 提督      %s%s\n
        'cmd_level': hexs(0x2303),      # 等級 %2d 經驗值
        'exp': hexs(0x2321),            # %6lu      \n
        'rule': hexs(0x232D),           # 整列空白（style 1 ＝ 底線）
        'ops_label': hexs(0x2357),      # 作戰
        'goto': hexs(0x235C),           # \x01[%d,%d,
        'position': hexs(0x2365),       # 現在位置
        'space': hexs(0x2372),          # ' '
        'nl_pad50': hexs(0x2374),       # \n%50s
        'status_head': hexs(0x237A),    # \n狀態      部隊組成\n
        'pad50': hexs(0x2391),          # %50s
        'adjutant_label': hexs(0x2396), # 副官\n
        'costs': hexs(0x23A3),          # 營運經費  維持費          移動經費
        'abil_head': hexs(0x23D6),      # 艦隊能力  艦隊士氣值      光線砲\n
        'abil_2': hexs(0x23F8),         # 統率力 飛彈
        'abil_3': hexs(0x241C),         # 回轉速度 戰鬥機
        'abil_4': hexs(0x243E),         # 巡航速度 占領
        'abil_5': hexs(0x245E),         # 戰鬥速度 防禦
        # 0DEA:03F9：作戰那兩列
        'ops_fight': hexs(0x2486),      # %s戰鬥
        'ops_in_fight': hexs(0x248D),   # 戰鬥中
        'ops_occupying': hexs(0x2496),  # 占領中
        # 0DEA:04C0：現在位置那一列右邊
        'move_target': hexs(0x24A8),    # 移動目標
        'standby_planet': hexs(0x24B5), # 待命行星
        'standby': hexs(0x24C2),        # 待命中（整列）
        # 0DEA:05DD：部隊組成
        'colour_str': hexs(0x24E8),     # \x03[%d,%s
        'empty_slot': hexs(0x24F0),     # □
        'two_spaces': hexs(0x24F3),     # 兩個空白
        # 0DEA:06FE：飛彈／戰鬥機
        'missile_label': hexs(0x24FF),  # 飛彈
        'pct': hexs(0x250C),            # %3ld%%
        'dashes': hexs(0x2513),         # ----
        'fighter_label': hexs(0x2518),  # 戰鬥機
        # 0DEA:0918：副官
        'name_rank': hexs(0x253E),      # %s%s
        'adj_name': hexs(0x2543),       # %-26s
        'adj_level': hexs(0x2552),      # 等級 %2d 經驗值
        'adj_exp': hexs(0x2566),        # %6lu
        'none': hexs(0x256B),           # 沒有
        'adj_none_level': hexs(0x257F), # 等級 -- 經驗值 -----
        # 0DEA:0A80：費用與十二格
        'colour_num6': hexs(0x25A1),    # \x03[%d,%6d
        'num6': hexs(0x25B3),           # %6d
        'num3': hexs(0x25C0),           # ' %3u'
        'colour_num3': hexs(0x25DC),    # \x03[%d, %3u
    }
    spec = {
        'note': '戰略地圖上的艦隊視窗（0DEA:0040；docs/spec/051、re/312）。'
                '格式字串逐位元組抄自資料段，順序就是原版呼叫 3617:003A 的順序。',
        'cell_x': 7, 'cell_y': 0, 'cols': 25, 'rows': 17, 'style': 1, 'colour': 0x97,
        'face': {'x': 448, 'y': 60, 'w': 80, 'h': 100,
                 'note': '3958:0215(提督索引, 0x38, 5)；矩形量自 fcmd-win'},
        'buttons': [
            {'name': 'ops', 'text': '作戰', 'cell_x': 22, 'cell_y': 14, 'cols': 2,
             'codes': [cstr(0x21B0).hex()]},
            {'name': 'change', 'text': '變更', 'cell_x': 22, 'cell_y': 15, 'cols': 2,
             'codes': [cstr(0x21B5).hex()]},
            {'name': 'close', 'text': '結束', 'cell_x': 22, 'cell_y': 16, 'cols': 2,
             'codes': [cstr(0x21BA).hex()]},
        ],
        'button_colour': 0x87,
        # ds:60F4：帶顏色跳脫碼的陣營名（\x03[150,帝國軍\x03[151,）；
        # ds:6099：純名字（確認框用）。
        'faction_names_coloured': [far(0x60F4 + 4 * i) for i in range(3)],
        'faction_names': [hexs(0x6099), hexs(0x60A0), hexs(0x60A7)],
        'faction_colour': [0x96, 0x95, 0x97],
        # ds:6172：艦種代號字母 Ｂ Ｈ Ｃ Ａ Ｄ Ｔ Ｅ Ｐ（0DEA:05DD 逐組印）。
        'kind_letters': [far(0x6172 + 4 * k) for k in range(8)],
        # ds:68EA：+0x08 那六種移動狀態的名字（0DEA:03F9 印「%s戰鬥」）。
        'move_names': [far(0x68EA + 4 * k) for k in range(6)],
        # 297B:0E8B：艦艇速度 ＝ 艦種 × 艦隊能力，再依組數打折（docs/re/312 §七）。
        'speed_shrink_table': list(m[DS + 0x4E2E:DS + 0x4E3A]),
        'formats': formats,
        'action_menu': {
            'note': '「作戰」開的五項選單（3667:0004 那一套，版面與戰場命令選單相同，'
                    '但夾在戰略地圖的板子裡：clamp_rows 17；量自 fcmd-menu）',
            'items': [cstr(o).decode('big5') for o in (0x21F0, 0x21F5, 0x21FA, 0x21FF, 0x2204)],
            'item_hex': [hexs(o) for o in (0x21F0, 0x21F5, 0x21FA, 0x21FF, 0x2204)],
            'width': 2, 'clamp_rows': 17,
        },
        'verbs': [hexs(o) for o in (0x22A8, 0x22AF, 0x22B4, 0x22BB, 0x22C6)],
        'verb_standby': hexs(0x2DF1),
        'confirm_format': hexs(0x2DF6),
        'unreachable': hexs(0x2DDC),
        'unreachable_kind': 0,
        'confirm_kind': 1,
    }
    with open(dst, 'w', encoding='utf-8') as fh:
        json.dump(spec, fh, ensure_ascii=False, indent=1)
    print('寫出', dst, len(formats), '條格式字串')


if __name__ == '__main__':
    main(sys.argv[1], sys.argv[2])
