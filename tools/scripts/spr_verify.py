# 判準：解出的圖塊要能在原版畫面上找到逐像素相同的位置。
import struct, sys, os
W = H = 16
f = open('/orig/SPRITE.SPR', 'rb').read()
def tile(k):
    d = f[k*128:(k+1)*128]
    px = bytearray(W*H)
    for y in range(H):
        w4 = struct.unpack_from('<4H', d, y*8)
        for x in range(W):
            v = 0
            for p in range(4):
                be = ((w4[p] & 0xFF) << 8) | (w4[p] >> 8)
                if be & (0x8000 >> x):
                    v |= 1 << p
            px[y*W+x] = v
    return bytes(px)
fb = open('/work/workplace/oracle/key1/screen.bin', 'rb').read()
SW = 640
hits = []
for k in range(len(f)//128):
    t = tile(k)
    if len(set(t)) < 3:
        continue
    for y0 in range(0, 480-H):
        for x0 in range(0, SW-W):
            if fb[y0*SW+x0] != t[0] or fb[y0*SW+x0+8] != t[8]:
                continue
            if all(fb[(y0+y)*SW+x0:(y0+y)*SW+x0+W] == t[y*W:(y+1)*W] for y in range(H)):
                hits.append((k, x0, y0))
                break
        else:
            continue
        break
print('在原版畫面上逐像素命中的圖塊:', hits[:20])
print('命中張數:', len(hits))
