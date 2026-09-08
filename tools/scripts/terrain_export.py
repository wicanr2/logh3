"""把戰場地形的通行規則從執行期記憶體倒成 JSON（`docs/re/100`）。

    tools/py.sh tools/scripts/terrain_export.py <記憶體傾印.bin> <DS(hex)> <輸出.json>

三張表：

  `ds:9BB4[圖塊]`  每個圖塊的四位元阻擋遮罩（實測只有 00 或 0F）
  `ds:3C3E[象限]`  象限 → 位元，象限 ＝ (半格列&1)*2 + (半格欄&1)
  圖塊 1           集結點（戰術層的候選目標），一張圖七個左右

順便把 32 張 `MAP<nn>.CMP` 的集結點掃出來，一起寫進去。
"""
import json
import pathlib
import sys

MASK_OFF, QUAD_OFF = 0x9BB4, 0x3C3E
RALLY_TILE, MAP_COLS, MAP_ROWS = 1, 128, 102
TILE_W, TILE_H = 16, 20


def decode_cmp(b):
    """`docs/re/81` 的 RLE：>=0x80 是一格，否則 (圖塊, 次數)。"""
    out, i = [], 0
    while i < len(b) and len(out) < MAP_COLS * MAP_ROWS:
        v = b[i]
        if v >= 0x80:
            out.append(v & 0x7F)
            i += 1
        else:
            out.extend([v] * b[i + 1])
            i += 2
    return out


mem = pathlib.Path(sys.argv[1]).read_bytes()
ds = int(sys.argv[2], 16) * 16
mask = [mem[ds + MASK_OFF + i] for i in range(64)]
quad = [mem[ds + QUAD_OFF + i] for i in range(4)]

rally = {}
for n in range(32):
    p = pathlib.Path('/orig/MAP%02d.CMP' % n)
    if not p.exists():
        continue
    cells = decode_cmp(p.read_bytes())
    rally[str(n)] = [[i % MAP_COLS * TILE_W, i // MAP_COLS * TILE_H]
                     for i, c in enumerate(cells) if c == RALLY_TILE]

out = {
    "_note": "戰場地形的通行規則與集結點（docs/re/100）。遮罩與象限表倒自"
             "執行期記憶體，集結點掃自 MAP<nn>.CMP 的圖塊 1。",
    "block_mask": mask,
    "quadrant_bits": quad,
    "half_cell_w": 8,
    "half_cell_h": 10,
    "footprint": 16,
    "step": 8,
    "rally_tile": RALLY_TILE,
    "rally_points": rally,
}
pathlib.Path(sys.argv[3]).write_text(
    json.dumps(out, ensure_ascii=False, indent=1), encoding='utf-8')
print('圖塊遮罩：', mask)
print('象限位元：', quad)
print('每張圖的集結點數：',
      {k: len(v) for k, v in sorted(rally.items(), key=lambda kv: int(kv[0]))})
