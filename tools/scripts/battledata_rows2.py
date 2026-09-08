"""逐列讀「資料」畫面（批次由 argv 指定），色 6 與色 7 分開讀。"""
import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
name = sys.argv[1]
rows, _ = S.load_screen('/work/workplace/oracle/%s/screen.bin' % name)
tabs = {b: S.font_lookup('/orig/GRAPH.IMG', bold=b) for b in (False, True)}

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

print('批次', name)
for ink in (6, 7):
    print(' 標籤 y= 82 色%d %s' % (ink, [(x,c) for x,c,_ in read(82, ink, 32, 300)]))
    print(' 表頭 y=102 色%d %s' % (ink, [(x,c,'*' if b else '') for x,c,b in read(102, ink, 32, 459)]))
for r in range(14):
    y = 120 + r*20 + 2
    for ink in (6, 7):
        got = [(x,c) for x,c,_ in read(y, ink) if c != '▏']
        if got:
            print(' 列%2d 色%d %s' % (r, ink, ' '.join('%d:%s'%g for g in got)))
