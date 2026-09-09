# -*- coding: utf-8 -*-
"""在程式碼區找某個 16 位元字面值出現的位置，連前後文一起印。

    tools/py.sh tools/scripts/find_word.py 42EA 4318 4396
"""
import sys
mem = open('workplace/s3datamem.bin', 'rb').read()
LO, HI = 0x10000, 0x54B30
for a in sys.argv[1:]:
    v = int(a, 16)
    pat = bytes([v & 0xFF, v >> 8])
    print('== %04X ==' % v)
    i = mem.find(pat, LO)
    n = 0
    while 0 <= i < HI:
        n += 1
        if n <= 10:
            print('   線性 %05X  前後: %s' %
                  (i, ' '.join('%02X' % b for b in mem[i - 6:i + 4])))
        i = mem.find(pat, i + 1)
    print('   共 %d 處' % n)
