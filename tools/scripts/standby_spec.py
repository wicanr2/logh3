#!/usr/bin/env python3
"""把「指定待命行星」那張視窗（`0DEA:1320`，`docs/re/314`）的格式字串、熱區與
小圖參數抄成規格 `docs/internals/data/fleet-standby-panel.json`。

    tools/py.sh tools/scripts/standby_spec.py workplace/fcx-fleetwin-mem.bin \
        docs/internals/data/fleet-standby-panel.json

熱區表 `ds:2248`（7 筆 × 12 bytes：格x、格y、寬、高、啟用、動作碼、種類、
標籤遠指標；`FF FF` 結尾）直接倒；小圖與引線的常數讀自 `0DEA:1320`／`1594`
與 `1727:05C6`（都是立即值，這裡照抄並註明出處）。
"""
import json, sys

DS = 0x54B3 * 16


def main(src, dst):
    m = open(src, 'rb').read()

    def hexs(off):
        p = DS + off
        return m[p:m.index(b'\0', p)].hex()

    hot = []
    for i in range(7):
        e = m[DS + 0x2248 + 12 * i:DS + 0x2248 + 12 * (i + 1)]
        hot.append({'cell_x': e[0], 'cell_y': e[1], 'cols': e[2], 'rows': e[3], 'code': e[5]})
    spec = {
        'note': '「指定待命行星」：作戰選單第 4 項開的視窗（0DEA:1320；docs/re/314、'
                'docs/spec/051 §待命）。視窗 3603:0006(0, 0x97, 8, 4, 18, 11, 1)；文字照 '
                '0DEA:14D2 的順序：標題、3982:00B6 的「%s第Ｎ艦隊  提督」、星系那一列、'
                '37A9:000F 畫熱區表 ds:2248 的鈕（前五筆是行星，只畫 i < 行星數的；'
                '第 6 筆決定、第 7 筆中止）、艦隊 +0x08==5 且 +0x07!=FF 時那一顆鈕用 0x27。'
                '之後 2376:0B44 把 MAP<nn>.CMP 縮成的 128×128 單色圖貼在 (152,220)，'
                '0DEA:1594 畫外框與每顆行星的引線。',
        'cell_x': 8, 'cell_y': 4, 'cols': 18, 'rows': 11, 'style': 1, 'colour': 0x97,
        'formats': {
            'title': hexs(0x2689),          # \x0c指定待命行星\n
            'fleet_prefix': hexs(0x8E98),   # %s第（3982:00B6，%s ＝ ds:60F4[陣營] 帶色陣營名）
            'fleet_suffix': hexs(0x8E9D),   # 艦隊  %s（%s ＝ 提督姓名）
            'region': hexs(0x2698),         # \n      星系      %s\n（%s ＝ 星域記錄，名字在開頭）
            'planet_label': hexs(0x2683),   # %-16s（0DEA:1433 把行星名排進 17 bytes 的標籤）
        },
        'hotspots': hot,
        'button_colour': 0x87, 'pressed_colour': 0x27,
        'decide': hexs(0x223E), 'cancel': hexs(0x2243),
        'minimap': {
            'x': 0x98, 'y': 0xB4 + 40, 'w': 128, 'h': 128,
            'frame': {'x0': 0x97, 'y0': 0xB3 + 40, 'x1': 0x118, 'y1': 0x134 + 40, 'ink': 7},
            'ink': 1, 'bg': 3,
            'grid_row_every': 12, 'grid_col_every': 15,
            'empty_tiles': [0, 1, 5],
            'row_repeat_every': 5,
            'note': '1727:05C6：來源列每輸出 5 列重用一列（102 列拉成 128 列；'
                    '第 127 列讀到地圖緩衝區外——那 128 個像素是垃圾，規格列為已知差異）。'
                    '1588:073C／078C 先把格線塞進圖塊：列 %12==0 的空格 → 3、'
                    '欄 %15==0 的空格 → 4（3 → 2）；ds:3C8C 的表：圖塊 0／1／5 不算，其餘都算。',
        },
        'leader': {
            'x0': 0x120, 'y0_base': 190 + 40, 'y0_pitch': 20, 'shadow_ink': 3,
            'note': '0DEA:1594：每顆行星從 (288, 230+20i) 拉到 (152+X>>4, 220+Y>>4)，'
                    '顏色 ds:6092[擁有者]；先畫一條 (−1,−1) 的色 3 影子。'
                    '行星位置是載入戰場地圖時算的（rules.PlanetSites）。',
        },
    }
    with open(dst, 'w', encoding='utf-8') as fh:
        json.dump(spec, fh, ensure_ascii=False, indent=1)
    print('寫出', dst)


if __name__ == '__main__':
    main(sys.argv[1], sys.argv[2])
