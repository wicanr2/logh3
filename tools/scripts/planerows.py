"""印字元平面（plane.bin，40×20 格 × 6 bytes）指定列的每格 (+4 屬性,+5 樣式,字碼)。
用法：planerows.py plane.bin 列 [列…]；不給列就印全部有字的列。"""
import sys, pathlib
m = pathlib.Path(sys.argv[1]).read_bytes()
rows = [int(r) for r in sys.argv[2:]] or range(20)
for r in rows:
    out = []
    for c in range(40):
        o = (r*40+c)*6
        own, mode, lo, hi, attr, style = m[o:o+6]
        if hi == 0 and lo == 0 and attr == 0 and style == 0:
            continue
        ch = ''
        if hi >= 0x80 and hi != 0xFF:
            try: ch = bytes([hi, lo]).decode('big5')
            except Exception: ch = '??'
        elif hi == 0xFF: ch = 'G%02x' % lo
        else: ch = repr(chr(lo))+ (repr(chr(hi)) if hi else '')
        out.append('%2d:%s a%02x s%02x o%02x' % (c, ch, attr, style, own))
    print('列 %2d: ' % r + ' | '.join(out))
