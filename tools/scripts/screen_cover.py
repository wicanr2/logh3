"""比兩份收據，印出差異的範圍與逐列統計。"""
import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
a,_=S.load_screen('/work/workplace/oracle/%s/screen.bin'%sys.argv[1])
b,_=S.load_screen('/work/workplace/oracle/%s/screen.bin'%sys.argv[2])
pts=[(x,y) for y in range(S.H) for x in range(S.W) if a[y][x]!=b[y][x]]
print('%s vs %s：%d 點' % (sys.argv[1], sys.argv[2], len(pts)))
if pts:
    print('  x %d–%d，y %d–%d' % (min(p[0] for p in pts), max(p[0] for p in pts),
                                  min(p[1] for p in pts), max(p[1] for p in pts)))
