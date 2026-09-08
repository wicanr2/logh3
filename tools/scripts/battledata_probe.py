"""量戰術選單「資料」那張畫面的骨架（`docs/re/109`）。

先與戰場畫面差分定出被蓋掉的範圍，再掃色 7 的水平／垂直線段，
把大表的列距、分隔線與上面三個標籤框的位置都算出來。
"""
import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S

rows, _ = S.load_screen('/work/workplace/oracle/s3data/screen.bin')
base, _ = S.load_screen('/work/workplace/oracle/s3ai6/screen.bin')

# 1. 差分範圍
xs, ys = [], []
for y in range(S.H):
    for x in range(S.W):
        if rows[y][x] != base[y][x]:
            xs.append(x); ys.append(y)
print('差分 %d 像素，x %d–%d，y %d–%d' % (len(xs), min(xs), max(xs), min(ys), max(ys)))

# 2. 色 7 的長水平線段（表格的分隔線）
print('\n色 7 的水平線段（長度 ≥ 100）：')
for y in range(S.H):
    x = 0
    while x < S.W:
        if rows[y][x] == 7:
            x0 = x
            while x < S.W and rows[y][x] == 7:
                x += 1
            if x - x0 >= 100:
                print('  y=%3d  x %3d–%3d  長 %d' % (y, x0, x-1, x-x0))
        else:
            x += 1

# 3. 色 7 的長垂直線段
print('\n色 7 的垂直線段（長度 ≥ 60）：')
for x in range(S.W):
    y = 0
    while y < S.H:
        if rows[y][x] == 7:
            y0 = y
            while y < S.H and rows[y][x] == 7:
                y += 1
            if y - y0 >= 60:
                print('  x=%3d  y %3d–%3d  長 %d' % (x, y0, y-1, y-y0))
        else:
            y += 1
