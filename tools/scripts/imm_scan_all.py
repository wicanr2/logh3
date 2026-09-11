# -*- coding: utf-8 -*-
"""在所有 1 MB 執行期傾印裡找「把某個立即值載進暫存器」的位置（含 push imm16）。

    tools/py.sh tools/scripts/imm_scan_all.py 85D 86F ...

`find_imm.py` 只掃 `s3datamem.bin` 的線性 0x10000–0x54B30；低於 0x10000 的
程式段（`0349`／`0374`／`056A`……）整段在它的過濾器外，會回假零
（docs/re/292 就是這樣撞到的）。這一支從線性 0 掃起，而且掃全部 1 MB 傾印——
程式分段常駐與否會隨階段變，一份掃不到不等於不存在（docs/re/288）。
"""
import glob
import os
import sys

vals = [int(a, 16) for a in sys.argv[1:]]
files = sorted(glob.glob('/work/workplace/*.bin'))
for fn in files:
    if os.path.getsize(fn) != 1 << 20:
        continue
    mem = open(fn, 'rb').read()
    for v in vals:
        lo, hi = v & 0xFF, v >> 8
        hits = []
        for i in range(1, len(mem) - 2):
            if mem[i] != lo or mem[i + 1] != hi:
                continue
            op = mem[i - 1]
            if 0xB8 <= op <= 0xBF or op == 0x68:
                hits.append((i - 1, op))
        if hits:
            print('%-28s %04X: %s' % (os.path.basename(fn), v,
                                      ' '.join('%05X(op %02X)' % h for h in hits)))
