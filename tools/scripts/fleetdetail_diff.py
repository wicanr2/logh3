import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
from collections import Counter
a,_=S.load_screen('/work/workplace/fleetdetail-remake.bin')
b,_=S.load_screen('/work/workplace/oracle/s3detail/screen.bin')
pts=[(x,y) for y in range(60,420) for x in range(16,463) if a[y][x]!=b[y][x]]
print('共 %d 點' % len(pts))
print('色對（remake→原版）:', Counter((a[y][x],b[y][x]) for x,y in pts).most_common(6))
rows=Counter()
for x,y in pts:
    rows[(y-102)//20 if 102<=y<382 else ('上' if y<102 else '下')] += 1
print('依列:', sorted(rows.items(), key=lambda kv: str(kv[0])))
xs=Counter(x//32*32 for x,y in pts)
print('x 分佈:', sorted(xs.items())[:16])
