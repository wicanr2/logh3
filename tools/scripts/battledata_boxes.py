"""量「資料」畫面上三個標籤框與表頭的細部。"""
import sys, collections
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
rows, _ = S.load_screen('/work/workplace/oracle/s3data/screen.bin')

def rect_of(x0,y0,x1,y1,c):
    xs=[x for y in range(y0,y1+1) for x in range(x0,x1+1) if rows[y][x]==c]
    ys=[y for y in range(y0,y1+1) for x in range(x0,x1+1) if rows[y][x]==c]
    return (min(xs),min(ys),max(xs),max(ys)) if xs else None

print('三個標籤框的輪廓（色 7 的外接矩形，各自的範圍內）：')
for i,(a,b) in enumerate(((30,110),(110,190),(190,270))):
    print('  框%d 色7 %s' % (i, rect_of(a,76,b,100,7)))
    for c in (15,2,8,3):
        r = rect_of(a,76,b,100,c)
        if r: print('       色%-2d %s' % (c, r))

print('\n框 0 的逐列樣貌（x 32–100，y 78–100）：')
for y in range(78,101):
    print(' y=%3d %s' % (y, ''.join('%X'%rows[y][x] for x in range(32,101))))

print('\n表頭列 y=100–119 的逐列色（x 288–433）：')
for y in range(98,121):
    c=collections.Counter(rows[y][288:434])
    print(' y=%3d %s' % (y, sorted(c.items(), key=lambda kv:-kv[1])[:4]))

print('\n大表視窗四角（x 20–459, y 64–414）往內剝：')
lay, alt = S.measure_frame(rows, 20, 64, 459, 414, fill=9, sample=(300, 70))
for i,(a,b) in enumerate(zip(lay,alt)):
    print('  第%d 層 %s alt %s' % (i,a,b))
