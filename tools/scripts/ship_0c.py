# -*- coding: utf-8 -*-
"""比對艦艇記錄的 +0x08（現有艘數）、+0x0A（滿編）與 +0x0C。"""
import struct
POOL = 0x668B2
N = 148
REC = 78
mem = open('workplace/s3datamem.bin', 'rb').read()
def u16(o): return struct.unpack_from('<H', mem, o)[0]
print('筆  陣營 艦種  +0x08  +0x0A  +0x0C   0C−08')
diff = {}
for i in range(N):
    r = POOL + i * REC
    f, k = mem[r], mem[r + 4]
    a, b, c = u16(r + 8), u16(r + 10), u16(r + 12)
    diff.setdefault(c - a, 0)
    diff[c - a] += 1
    if i < 12 or a != b:
        print('%3d %4d %4d %6d %6d %6d %7d' % (i, f, k, a, b, c, c - a))
print('0C−08 的分布：', sorted(diff.items()))
