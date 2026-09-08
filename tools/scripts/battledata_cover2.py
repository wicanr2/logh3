import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
a, _ = S.load_screen('/work/workplace/oracle/s3data/screen.bin')
b, _ = S.load_screen('/work/workplace/oracle/s3data0/screen.bin')
right = [(x,y) for y in range(S.H) for x in range(463, S.W) if a[y][x]!=b[y][x]]
print('右側差異 %d 點，x %d–%d，y %d–%d' % (len(right), min(p[0] for p in right),
      max(p[0] for p in right), min(p[1] for p in right), max(p[1] for p in right)))
from collections import Counter
print('右側差異的色（前→後）：', Counter((b[y][x], a[y][x]) for x,y in right).most_common(6))
left = [(x,y) for y in range(S.H) for x in range(0, 463) if a[y][x]!=b[y][x]]
print('左側差異 x %d–%d，y %d–%d' % (min(p[0] for p in left), max(p[0] for p in left),
      min(p[1] for p in left), max(p[1] for p in left)))
print('\nx 14–24 在 y 60–70 的兩張畫面：')
for y in range(58, 68):
    print(' y=%3d 前 %s 後 %s' % (y, ''.join('%X'%b[y][x] for x in range(14,26)),
                                   ''.join('%X'%a[y][x] for x in range(14,26))))
