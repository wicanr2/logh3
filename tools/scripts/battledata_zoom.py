import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
name, x0, y0, x1, y1 = sys.argv[1], *map(int, sys.argv[2:6])
rows, _ = S.load_screen('/work/workplace/oracle/%s/screen.bin' % name)
print('   ' + ''.join('%X' % ((x//10)%16) for x in range(x0, x1+1)))
print('   ' + ''.join('%X' % (x%10) for x in range(x0, x1+1)))
for y in range(y0, y1+1):
    print('%3d %s' % (y, ''.join('.' if rows[y][x] in (8,9) else '%X'%rows[y][x] for x in range(x0, x1+1))))
