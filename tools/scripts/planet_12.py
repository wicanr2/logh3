# -*- coding: utf-8 -*-
"""列出行星記錄 +0x12 非 0 的那些（要塞？），順帶看 +0x45。"""
BASE = 0x53132
N, REC = 81, 82
mem = open('workplace/s3datamem.bin', 'rb').read()
n = 0
for i in range(N):
    r = BASE + i * REC
    name = mem[r:r+20].decode('big5', 'replace').replace('　', '').strip()
    if mem[r+0x12]:
        n += 1
        print('%2d %-10s 擁有者 %d  +0x12 %02X  +0x45 %02X  防衛 %d'
              % (i, name, mem[r+0x11], mem[r+0x12], mem[r+0x45],
                 mem[r+0x14] | mem[r+0x15] << 8))
print('+0x12 非 0 的有 %d 顆／%d' % (n, N))
from collections import Counter
print('+0x45 的分布：', Counter(mem[BASE+i*REC+0x45] for i in range(N)).most_common())
