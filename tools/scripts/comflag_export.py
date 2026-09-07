"""把 `COM_FLAG.SPR` 解成五張標籤 PNG 與一份 JSON。

    tools/py.sh tools/scripts/comflag_export.py

格式見 `docs/internals/21`：64 × 45，四個 plane 以「一列一個 plane」交錯，
一個影像列 32 bytes。內容是戰鬥指令的英文字樣，一個 9 列高。
調色盤與 `spr_export.py` 同一個來源（收據的畫面 PNG）。
"""
import json, os, struct, sys
sys.path.insert(0, '/work/tools/scripts')
from amg_export import write_png

W, PLANE, LABEL_H = 64, 8, 9
LABELS = ['BEEM', 'MISSILE', 'FIGHTER', 'ATTACK', 'DEFENCE']

d = open('/orig/COM_FLAG.SPR', 'rb').read()
rows = len(d) // (PLANE * 4)
px = bytearray(W * rows)
for r in range(rows):
    base = r * PLANE * 4
    for x in range(W):
        by, m = x // 8, 0x80 >> (x % 8)
        v = 0
        for p in range(4):
            if d[base + p * PLANE + by] & m:
                v |= 1 << p
        px[r * W + x] = v

png = open('/work/workplace/oracle/key1/screen.png', 'rb').read()
i, pal = 8, None
while i < len(png):
    ln = struct.unpack_from('>I', png, i)[0]
    if png[i + 4:i + 8] == b'PLTE':
        pal = list(png[i + 8:i + 8 + ln])
        break
    i += 12 + ln
pal = (pal + [0] * 768)[:48]

out = '/work/workplace/assets/comflag'
os.makedirs(out, exist_ok=True)
for k, name in enumerate(LABELS):
    sub = px[k * LABEL_H * W:(k + 1) * LABEL_H * W]
    write_png(f'{out}/{k}-{name}.png', W, LABEL_H, sub, pal)
write_png(f'{out}/sheet.png', W, rows, px, pal)

json.dump({
    'source': 'COM_FLAG.SPR',
    'bytes': len(d),
    'width': W, 'height': rows,
    'layout': '一個影像列 32 bytes ＝ plane 0..3 各 8 bytes（64 像素）；色號 = p0|p1<<1|p2<<2|p3<<3',
    'label_height': LABEL_H,
    'labels': LABELS,
    'ink': 4, 'outline': 11,
    'note': 'BEEM 是原版自帶的錯字（應為 BEAM），保留原樣。用途推測是戰鬥畫面的指令字樣（COM ＝ command），還沒在畫面上對過。',
}, open('/work/docs/internals/data/com-flag.json', 'w'), ensure_ascii=False, indent=1)
print('五張標籤 ＋ sheet 寫進', out)
