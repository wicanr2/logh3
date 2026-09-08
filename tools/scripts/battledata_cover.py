"""「資料」畫面蓋掉戰場的哪一塊：與同狀態的前一格（s3data0）逐像素比。"""
import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
a, _ = S.load_screen('/work/workplace/oracle/s3data/screen.bin')
b, _ = S.load_screen('/work/workplace/oracle/s3data0/screen.bin')
xs, ys, n = [], [], 0
per_row = []
for y in range(S.H):
    c = sum(1 for x in range(S.W) if a[y][x] != b[y][x])
    per_row.append(c)
    if c:
        ys.append(y)
        for x in range(S.W):
            if a[y][x] != b[y][x]:
                xs.append(x)
        n += c
print('不同 %d 像素，x %d–%d，y %d–%d' % (n, min(xs), max(xs), min(ys), max(ys)))
print('每一列的差異數（只印非零）：')
run = None
for y, c in enumerate(per_row):
    if c and run is None: run = (y, c)
    if (not c or c != run[1]) and run is not None:
        if y-1 > run[0]: print('  y %3d–%3d 各 %d' % (run[0], y-1, run[1]))
        run = (y, c) if c else None
if run: print('  y %3d–%3d 各 %d' % (run[0], S.H-1, run[1]))
# 右側兩塊有沒有動
def region(x0,y0,x1,y1):
    return sum(1 for y in range(y0,y1+1) for x in range(x0,x1+1) if a[y][x]!=b[y][x])
print('右上小地圖 (465,59)-(623,219)：%d' % region(465,59,623,219))
print('右下狀態欄 (465,220)-(623,419)：%d' % region(465,220,623,419))
print('左側大表   (16,56)-(462,419)：%d' % region(16,56,462,419))
