"""逐列讀「資料」畫面大表的內容，色 6 與色 7 分開讀。"""
import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
rows, _ = S.load_screen('/work/workplace/oracle/s3data/screen.bin')
tabs = {}
for bold in (False, True):
    tabs[bold] = S.font_lookup('/orig/GRAPH.IMG', bold=bold)

def big5(c):
    try: return bytes([c>>8, c&0xFF]).decode('big5')
    except Exception: return '?%04X'%c

def read(y, ink, x0=32, x1=433):
    out, x = [], x0
    while x + 8 <= x1:
        hit = None
        for bold in (False, True):
            full, half = tabs[bold]
            b = S.cell_bits(rows, x, y, ink, 16)
            if any(b) and b in full:
                hit = (big5(full[b]), 16, bold); break
        if not hit:
            for bold in (False, True):
                full, half = tabs[bold]
                b = S.cell_bits(rows, x, y, ink, 8)
                if any(b) and b in half:
                    hit = (chr(half[b]), 8, bold); break
        if hit:
            out.append((x, hit[0], hit[2])); x += hit[1]
        else:
            x += 8
    return out

for r in range(14):
    y = 120 + r*20 + 2
    for ink in (6, 7):
        got = read(y, ink)
        if got:
            print('列%2d y=%3d 色%d  %s' % (r, y, ink,
                  ' '.join('%d:%s%s' % (x, ch, '*' if b else '') for x, ch, b in got)))
print()
for ink in (6, 7):
    print('表頭 y=102 色%d %s' % (ink, read(102, ink, 32, 459)))
    print('標籤 y=82  色%d %s' % (ink, read(82, ink, 32, 300)))
