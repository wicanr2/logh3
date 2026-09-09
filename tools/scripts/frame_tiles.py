"""把視窗框線的圖塊從資料段解出來。

    frame_tiles.py <ds 傾印> <調色盤.bin> <輸出.png> <輸出.json> [張數]

`2376:0E4A` 畫框線格：`si = ds:6BF4 ＋ 0xA0 × (字碼低位元組 − 1)`，
一格 16 × 20、四平面交錯（每列 4 個字組，對應 map mask 1／2／4／8），
共 160 bytes。色號 0 的像素會被視窗顏色的高四位填掉，所以這裡解出來
的 0 是「透明」而不是黑。
"""
import json
import sys

sys.path.insert(0, __file__.rsplit("/", 1)[0])
import png as pngmod  # noqa: E402

BASE = 0x6BF4
TILE = 0xA0
W, H = 16, 20

ds = open(sys.argv[1], "rb").read()
pal = open(sys.argv[2], "rb").read()
out_png, out_json = sys.argv[3], sys.argv[4]
n = int(sys.argv[5]) if len(sys.argv) > 5 else 8

tiles = []
for t in range(n):
    off = BASE + TILE * t
    rows = []
    for y in range(H):
        r = off + y * 8
        planes = [int.from_bytes(ds[r + 2 * p:r + 2 * p + 2], "little") for p in range(4)]
        row = []
        for x in range(W):
            byte = 0 if x < 8 else 1
            bit = 7 - (x & 7)
            v = 0
            for p in range(4):
                b = (planes[p] >> (8 * byte)) & 0xFF
                if (b >> bit) & 1:
                    v |= 1 << p
            row.append(v)
        rows.append(row)
    tiles.append(rows)

rgb = []
for y in range(H):
    line = bytearray()
    for t in range(n):
        for x in range(W):
            v = tiles[t][y][x]
            line += pal[v * 3:v * 3 + 3]
    rgb.append(line)
pngmod.write_rgb(out_png, W * n, H, rgb)
json.dump({"base": BASE, "tileBytes": TILE, "width": W, "height": H,
           "count": n, "tiles": tiles}, open(out_json, "w"), separators=(",", ":"))
print("%d 張 → %s / %s" % (n, out_png, out_json))
for t, rows in enumerate(tiles):
    hist = {}
    for row in rows:
        for v in row:
            hist[v] = hist.get(v, 0) + 1
    print("  圖塊 %d 色號分佈 %s" % (t, sorted(hist.items())))
