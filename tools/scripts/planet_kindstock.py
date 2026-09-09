# -*- coding: utf-8 -*-
"""行星記錄 +0x20 起八個 32 位元（每個艦種一份的補充存量？）與 +0x40。"""
import struct
BASE, N, REC = 0x53132, 81, 82
mem = open('workplace/s3datamem.bin', 'rb').read()
def u32(o): return struct.unpack_from('<I', mem, o)[0]
nz = [0]*8
for i in range(N):
    r = BASE + i*REC
    for k in range(8):
        if u32(r + 0x20 + k*4):
            nz[k] += 1
print('八個槽各有幾顆非 0：', nz)
for i in (0, 1, 33, 50, 80):
    r = BASE + i*REC
    name = mem[r:r+20].decode('big5','replace').replace('　','').strip()[:8]
    print('%2d %-9s %s  +0x40=%d 經濟=%d'
          % (i, name, [u32(r+0x20+k*4) for k in range(8)], u32(r+0x40), u32(r+0x1A)))
