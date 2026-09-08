# -*- coding: utf-8 -*-
"""在打過的狀態下比對艦艇記錄 +0x08／+0x0A／+0x0C（收據 s3form 的傾印）。"""
import struct
BASE = 0x60000
POOL = 0x668B2 - BASE
N, REC = 148, 78
mem = open('workplace/oracle/s3form/form_mem.bin', 'rb').read()
def u16(o): return struct.unpack_from('<H', mem, o)[0]
def u32(o): return struct.unpack_from('<I', mem, o)[0]
print('筆 艦隊 陣營 艦種  +0x08  +0x0A  +0x0C   0A−08  0C−08   +0x0E   +0x12')
n = 0
for i in range(N):
    r = POOL + i * REC
    fl, f, k = mem[r + 1], mem[r], mem[r + 4]
    a, b, c = u16(r + 8), u16(r + 10), u16(r + 12)
    if a == b == c:
        continue
    n += 1
    print('%3d %4d %4d %4d %6d %6d %6d %7d %6d %7d %7d'
          % (i, fl, f, k, a, b, c, b - a, c - a, u32(r + 14), u32(r + 18)))
print('三個不全等的有 %d 筆／%d' % (n, N))
