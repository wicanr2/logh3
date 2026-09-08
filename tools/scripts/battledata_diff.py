import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
from collections import Counter
a, _ = S.load_screen('/work/workplace/battledata-remake.bin')
b, _ = S.load_screen('/work/workplace/oracle/s3data/screen.bin')
pts = [(x, y) for y in range(60, 420) for x in range(16, 463) if a[y][x] != b[y][x]]
print('共 %d 點' % len(pts))
print('色對（remake→原版）前 8:', Counter((a[y][x], b[y][x]) for x, y in pts).most_common(8))
# 依 20 列分組
rowsum = Counter()
for x, y in pts:
    if y < 120: rowsum['標籤/表頭 y%d' % y] += 1
    elif y < 400: rowsum['列%d' % ((y-120)//20)] += 1
    else: rowsum['底 y%d' % y] += 1
for k, v in sorted(rowsum.items(), key=lambda kv: -kv[1])[:14]:
    print(' %-14s %d' % (k, v))
xs = Counter(x//16*16 for x, y in pts)
print('x 分佈（每 16 一格）前 12:', xs.most_common(12))
