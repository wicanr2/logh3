# SPRITE.SPR：60 張 16×20 的 planar 圖塊，未壓縮。
# 版面照原版 2376:0D25 的 blit：每一列依序寫 plane0..3 各一個 word。
import json, os, struct, sys, zlib
sys.path.insert(0, '/work/tools/scripts')
from amg_export import write_png  # 共用 PNG 輸出

W, H = 16, int(sys.argv[1]) if len(sys.argv) > 1 else 20
def decode_tile(d):
    px = bytearray(W * H)
    for y in range(H):
        words = struct.unpack_from('<4H', d, y * 8)
        for x in range(W):
            bit = 15 - x if x < 16 else 0
            v = 0
            for p in range(4):
                # word 是小端，但 VGA 的位元順序是高位在左：先換成大端再取位元
                be = ((words[p] & 0xFF) << 8) | (words[p] >> 8)
                if be & (0x8000 >> x):
                    v |= 1 << p
            px[y*W + x] = v
    return bytes(px)

name = sys.argv[2] if len(sys.argv) > 2 else 'SPRITE.SPR'
f = open('/orig/' + name, 'rb').read()
n = len(f) // (H * 8)
print(f'{name} {len(f)} bytes → {n} 張 {W}×{H}')
png = open('/work/workplace/oracle/key1/screen.png', 'rb').read()
i, pal = 8, None
while i < len(png):
    ln = struct.unpack_from('>I', png, i)[0]
    if png[i+4:i+8] == b'PLTE':
        pal = list(png[i+8:i+8+ln]); break
    i += 12 + ln
pal = (pal + [0]*768)[:48]
out = '/work/workplace/assets/' + name.split('.')[0].lower()
os.makedirs(out, exist_ok=True)
for k in range(n):
    px = decode_tile(f[k*(H*8):(k+1)*(H*8)])
    write_png(f'{out}/spr-{k:02d}.png', W, H, px, pal)
print('導出', n, '張 →', out)
