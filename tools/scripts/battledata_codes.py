"""印出某一批收據表上每一列的字碼（給對拍測試當固定輸入）。"""
import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
name = sys.argv[1]
rows, _ = S.load_screen('/work/workplace/oracle/%s/screen.bin' % name)
tabs = {b: S.font_lookup('/orig/GRAPH.IMG', bold=b) for b in (False, True)}

def read(y, ink, x0=32, x1=447):
    out, x = [], x0
    while x + 8 <= x1:
        hit = None
        for bold in (False, True):
            full, half = tabs[bold]
            b = S.cell_bits(rows, x, y, ink, 16)
            if any(b) and b in full:
                hit = (full[b], 16); break
        if not hit:
            for bold in (False, True):
                full, half = tabs[bold]
                b = S.cell_bits(rows, x, y, ink, 8)
                if any(b) and b in half:
                    hit = (half[b], 8); break
        if hit:
            out.append((x, hit[0])); x += hit[1]
        else:
            x += 8
    return out

for r in range(14):
    y = 120 + r*20 + 2
    for ink in (5, 6, 7):
        got = [(x, c) for x, c in read(y, ink) if c not in (0x2502,)]
        got = [(x, c) for x, c in got if not (x == 32)]
        if got:
            print('列%2d 色%d %s' % (r, ink,
                  ' '.join('%d:0x%04X' % (x, c) for x, c in got)))
