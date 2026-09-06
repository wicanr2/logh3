# 把一個未知的 planar 圖檔照指定尺寸拼成一張聯絡表，用來目視挑對的版面。
#   tools/py.sh tools/scripts/spr_sheet.py <檔名> <寬> <高> [每列幾張]
import os, struct, sys
sys.path.insert(0, '/work/tools/scripts')
from pngout import write_png, game_palette

name, W, H = sys.argv[1], int(sys.argv[2]), int(sys.argv[3])
COLS = int(sys.argv[4]) if len(sys.argv) > 4 else 12
SCALE = int(sys.argv[5]) if len(sys.argv) > 5 else 4
# MODE：row ＝ 每列 plane0..3（SPRITE.SPR 的版面）、plane ＝ 一張裡四個 plane
# 各自連續、pack4 ＝ 線性 4 位元（一個 byte 兩個像素）、pack8 ＝ 線性一像素一 byte。
MODE = sys.argv[6] if len(sys.argv) > 6 else 'row'
BPR = (W + 7) // 8          # 每個 plane 一列幾 bytes
TILE = {'row': BPR * 4 * H, 'plane': BPR * 4 * H,
        'pack4': (W // 2) * H, 'pack8': W * H}[MODE]

d = open('/orig/' + name, 'rb').read()
n = len(d) // TILE
print(f'{name} {len(d)} bytes，{W}×{H} → 每張 {TILE} bytes，{n} 張，剩 {len(d)%TILE}')
if n == 0:
    sys.exit(0)

def tile(off):
    px = bytearray(W * H)
    for y in range(H):
        for x in range(W):
            if MODE == 'pack8':
                px[y * W + x] = d[off + y * W + x]
                continue
            if MODE == 'pack4':
                b = d[off + y * (W // 2) + x // 2]
                px[y * W + x] = (b >> 4) if x % 2 == 0 else (b & 15)
                continue
            v = 0
            for p in range(4):
                if MODE == 'row':
                    b = d[off + y * BPR * 4 + p * BPR + x // 8]
                else:  # plane：一張裡 plane0 整張、plane1 整張…
                    b = d[off + p * BPR * H + y * BPR + x // 8]
                if b & (0x80 >> (x % 8)):
                    v |= 1 << p
            px[y * W + x] = v
    return px

rows = (n + COLS - 1) // COLS
SW, SH = COLS * (W + 1) + 1, rows * (H + 1) + 1
sheet = bytearray([1]) * (SW * SH)
for k in range(n):
    px = tile(k * TILE)
    ox, oy = (k % COLS) * (W + 1) + 1, (k // COLS) * (H + 1) + 1
    for y in range(H):
        sheet[(oy + y) * SW + ox:(oy + y) * SW + ox + W] = px[y * W:(y + 1) * W]

pal = game_palette()
if SCALE > 1:
    big = bytearray(SW * SCALE * SH * SCALE)
    for y in range(SH):
        row = bytes(b for v in sheet[y * SW:(y + 1) * SW] for b in [v] * SCALE)
        for k in range(SCALE):
            o = ((y * SCALE) + k) * SW * SCALE
            big[o:o + len(row)] = row
    sheet, SW, SH = big, SW * SCALE, SH * SCALE
out = f'/work/workplace/assets/sheet-{name.split(".")[0].lower()}-{W}x{H}.png'
os.makedirs('/work/workplace/assets', exist_ok=True)
write_png(out, SW, SH, bytes(sheet), pal)
print('→', out)
