# -*- coding: utf-8 -*-
"""把選單表與跳躍表從執行期傾印倒出來。

    tools/py.sh tools/scripts/menu_dump.py menu 42EA 5      # 5 項的選單表
    tools/py.sh tools/scripts/menu_dump.py jump 4396 4      # 4 項的跳躍表
"""
import struct, sys

DS = 0x54B30
mem = open('workplace/s3datamem.bin', 'rb').read()

def u16(o): return struct.unpack_from('<H', mem, o)[0]

def s(off):
    p = DS + off
    e = mem.find(b'\x00', p)
    return mem[p:e].decode('big5', 'replace')

mode, off, n = sys.argv[1], int(sys.argv[2], 16), int(sys.argv[3])
base = DS + off
if mode == 'menu':
    # 一項 6 bytes：旗標、pad、名稱遠指標（位移:段）
    for i in range(n):
        o = base + i * 6
        flag = mem[o]
        noff, nseg = u16(o + 2), u16(o + 4)
        print('%2d 旗標 %02X  名稱 %04X:%04X → %s'
              % (i, flag, nseg, noff, s(noff) if nseg == 0x54B3 else '(非資料段)'))
else:
    for i in range(n):
        o = base + i * 4
        print('%2d → %04X:%04X' % (i, u16(o + 2), u16(o)))
