import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
name, y, x0, n = sys.argv[1], int(sys.argv[2]), int(sys.argv[3]), int(sys.argv[4])
rows, _ = S.load_screen('/work/workplace/oracle/%s/screen.bin' % name)
cells = [[''.join('#' if rows[y+r][x0+c*8+k] == 7 else '.' for k in range(8))
          for r in range(S.GLYPH_H)] for c in range(n)]
print('   ' + '  '.join('x=%-6d' % (x0+c*8) for c in range(n)))
for r in range(S.GLYPH_H):
    print('%2d ' % r + '  '.join(c[r] for c in cells))
