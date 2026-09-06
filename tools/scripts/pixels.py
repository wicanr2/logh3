"""印一塊區域的色號，兩份收據可以並排比。

量框線與立體邊的時候，連通分量只說「這裡不一樣」，這支說「每個像素是什麼」。

用法：tools/py.sh tools/scripts/pixels.py <批次>[,<批次>] <x> <y> <寬> <高>
"""
import pathlib, sys

batches = sys.argv[1].split(',')
x0, y0, w, h = (int(v) for v in sys.argv[2:6])
scr = [pathlib.Path('/work/workplace/oracle/%s/screen.bin' % b).read_bytes()
       for b in batches]
head = '    ' + ' '.join('%X' % (x % 16) for x in range(x0, x0 + w))
print('    '.join(['%-*s' % (len(head), '%s' % b) for b in batches]))
print('    '.join([head] * len(batches)))
for y in range(y0, y0 + h):
    row = []
    for s in scr:
        row.append('%3d ' % y + ' '.join('%X' % s[y * 640 + x]
                                         for x in range(x0, x0 + w)))
    print('    '.join(row))
