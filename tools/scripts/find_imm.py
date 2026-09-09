# -*- coding: utf-8 -*-
"""在 1 MB 的執行期傾印裡找「把某個立即值載進暫存器」的位置。

    tools/py.sh tools/scripts/find_imm.py 42EA 4318 4396

只掃 `mov r16, imm16`（B8–BF）與 `push imm16`（68）兩種，
並把命中位置換算成幾個常見段的 seg:off。
"""
import struct, sys

mem = open('workplace/s3datamem.bin', 'rb').read()
for a in sys.argv[1:]:
    v = int(a, 16)
    lo, hi = v & 0xFF, v >> 8
    print('== %04X ==' % v)
    n = 0
    for i in range(1, len(mem) - 2):
        op = mem[i - 1]
        if not (0xB8 <= op <= 0xBF or op == 0x68):
            continue
        if mem[i] != lo or mem[i + 1] != hi:
            continue
        # 只看程式碼區（線性 0x10000–0x54B30）
        if not (0x10000 <= i < 0x54B30):
            continue
        n += 1
        if n <= 12:
            print('   線性 %05X（op %02X）' % (i - 1, op))
    print('   共 %d 處' % n)
