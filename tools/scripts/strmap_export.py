# STRMAP.BIT：星域地圖，300 個 16×20 的 planar 圖塊（row-major）。
# 版面來自原版 2376:0D3C 的 blit：每列 4 個 plane 各一個 word，SI 連續走完
# 一塊再走下一塊；第一塊畫在畫面 (0,40)。
import struct, sys, os
sys.path.insert(0, '/work/tools/scripts')
from amg_export import write_png

TW, TH, COLS = 16, 20, 20
f = open('/orig/STRMAP.BIT', 'rb').read()
n = len(f) // (TH * 8)
rows_of_tiles = n // COLS
W, H = COLS * TW, rows_of_tiles * TH
px = bytearray(W * H)
for k in range(n):
    tx, ty = (k % COLS) * TW, (k // COLS) * TH
    d = f[k*TH*8:(k+1)*TH*8]
    for y in range(TH):
        w4 = struct.unpack_from('<4H', d, y*8)
        for x in range(TW):
            v = 0
            for p in range(4):
                be = ((w4[p] & 0xFF) << 8) | (w4[p] >> 8)
                if be & (0x8000 >> x):
                    v |= 1 << p
            px[(ty+y)*W + tx+x] = v
print(f'{len(f)} bytes → {n} 塊 → {W}×{H}')

# 判準：貼到畫面 (0,40) 要能對上原版。地圖上另外畫了星球與艦隊圖示，
# 所以不會 0 差異；看的是**底圖對不對**。
fb = open('/work/workplace/oracle/key1/screen.bin', 'rb').read()
SW, X0, Y0 = 640, 0, 40
# 會計報告視窗蓋住地圖右半，所以只比對沒被遮住的左邊那一條。
for x1 in (140, 150, 160):
    tot = bad = 0
    for y in range(H):
        for x in range(x1):
            if Y0+y >= 480:
                continue
            tot += 1
            if fb[(Y0+y)*SW + X0+x] != px[y*W+x]:
                bad += 1
    print(f'x < {x1}：{bad} / {tot} 不同')

png = open('/work/workplace/oracle/key1/screen.png', 'rb').read()
i, pal = 8, None
while i < len(png):
    ln = struct.unpack_from('>I', png, i)[0]
    if png[i+4:i+8] == b'PLTE':
        pal = list(png[i+8:i+8+ln]); break
    i += 12 + ln
os.makedirs('/work/workplace/assets', exist_ok=True)
write_png('/work/workplace/assets/strmap.png', W, H, bytes(px), (pal+[0]*768)[:48])
print('→ workplace/assets/strmap.png')
