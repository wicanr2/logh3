import sys, collections
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
rows, _ = S.load_screen('/work/workplace/oracle/s3data/screen.bin')

def census(x0,y0,x1,y1,label):
    c = collections.Counter()
    for y in range(y0,y1+1):
        for x in range(x0,x1+1):
            c[rows[y][x]] += 1
    print('%-22s %s' % (label, sorted(c.items(), key=lambda kv:-kv[1])))

census(33,65,458,79,'標籤列上方')
census(33,82,110,96,'標籤1 內')
census(33,120,433,139,'第 0 列')
census(33,140,433,159,'第 1 列')
census(33,100,433,119,'表頭列')
census(21,65,458,413,'整個大表視窗')

# 第 0 列的左半（第１艦隊）用什麼色畫
print('\n第 0 列 y=118–140 的逐列色分布（x 32–130）：')
for y in range(118,141):
    c = collections.Counter(rows[y][32:131])
    print(' y=%3d %s' % (y, sorted(c.items(), key=lambda kv:-kv[1])[:5]))
