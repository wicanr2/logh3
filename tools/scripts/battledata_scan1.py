"""逐像素掃一條帶，把每個對得上字型的位置列出來（不假設格點）。"""
import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
name, y, x0, x1, ink = sys.argv[1], int(sys.argv[2]), int(sys.argv[3]), int(sys.argv[4]), int(sys.argv[5])
rows, _ = S.load_screen('/work/workplace/oracle/%s/screen.bin' % name)
for bold in (False, True):
    full, half = S.font_lookup('/orig/GRAPH.IMG', bold=bold)
    hits = []
    for x in range(x0, x1-7):
        b16 = S.cell_bits(rows, x, y, ink, 16)
        if any(b16) and b16 in full:
            c = full[b16]
            try: t = bytes([c>>8, c&0xFF]).decode('big5')
            except Exception: t = '?%04X'%c
            hits.append((x, '全', t))
        b8 = S.cell_bits(rows, x, y, ink, 8)
        if any(b8) and b8 in half:
            hits.append((x, '半', chr(half[b8])))
    print('%s bold=%s: %s' % (y, bold, ' '.join('%d%s%s'%h for h in hits)))
