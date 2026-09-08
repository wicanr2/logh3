# -*- coding: utf-8 -*-
"""在倒出來的 .COM 段裡找對某個 seg:off 的 far call（9A off_lo off_hi seg_lo seg_hi）。

    tools/py.sh tools/scripts/find_farcall.py 2BE1 06C5
"""
import os, sys, struct

seg, off = int(sys.argv[1], 16), int(sys.argv[2], 16)
pat = b'\x9a' + struct.pack('<HH', off, seg)
d = 'workplace/ida'
for fn in sorted(os.listdir(d)):
    if not fn.endswith('.com'):
        continue
    b = open(os.path.join(d, fn), 'rb').read()
    # 檔名 segXXXX.com / s3segXXXX.com 就是段基底；前面補了 0x100
    base = fn.replace('s3seg', '').replace('seg', '').split('.')[0]
    try:
        sb = int(base[:4], 16)
    except ValueError:
        sb = None
    i = b.find(pat)
    while i >= 0:
        where = '%s +%04X' % (fn, i - 0x100)
        if sb is not None:
            where += '  ＝ %04X:%04X' % (sb, i - 0x100)
        print(where)
        i = b.find(pat, i + 1)
