# -*- coding: utf-8 -*-
"""從 1 MB 的執行期記憶體傾印（線性位址 0 起）讀資料段裡的字串。

    tools/py.sh tools/scripts/read_ds_string.py 3F14 [4A24 ...]
"""
import sys

DS = 0x54B30
mem = open('workplace/s3datamem.bin', 'rb').read()

for a in sys.argv[1:]:
    off = int(a, 16)
    p = DS + off
    end = mem.find(b'\x00', p)
    raw = mem[p:end]
    try:
        txt = raw.decode('big5')
    except UnicodeDecodeError:
        txt = raw.decode('big5', 'replace')
    print('ds:%s（線性 %05X）%r' % (a.upper(), p, raw))
    print('   → %s' % txt)
