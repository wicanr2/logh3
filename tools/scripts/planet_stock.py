# -*- coding: utf-8 -*-
"""看行星記錄 +0x3C 與 +0x40 兩個 32 位元量（補給／補充的存量？）。"""
import struct
BASE, N, REC = 0x53132, 81, 82
mem = open('workplace/s3datamem.bin', 'rb').read()
def u32(o): return struct.unpack_from('<I', mem, o)[0]
nz = 0
for i in range(N):
    r = BASE + i * REC
    a, b, c = u32(r+0x3C), u32(r+0x40), u32(r+0x44)
    if a or b:
        nz += 1
        if nz <= 12:
            name = mem[r:r+20].decode('big5','replace').replace('　','').strip()
            print('%2d %-12s 擁有者 %d 要塞 %d  +0x3C %8d  +0x40 %8d  +0x44 %08X  經濟 %d'
                  % (i, name[:10], mem[r+0x11], mem[r+0x12], a, b, c, u32(r+0x1A)))
print('+0x3C 或 +0x40 非 0 的有 %d 顆／%d' % (nz, N))
