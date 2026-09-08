"""讀「資料」畫面上的文字與小框（`docs/re/109`）。"""
import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S

rows, _ = S.load_screen('/work/workplace/oracle/s3data/screen.bin')
full, half = S.font_lookup('/orig/GRAPH.IMG')
fullb, halfb = S.font_lookup('/orig/GRAPH.IMG', bold=True)

def big5(code):
    try:
        return bytes([code >> 8, code & 0xFF]).decode('big5')
    except Exception:
        return '?%04X' % code

# 上排三個標籤框：找 y 70–100 之間的短水平色 7 線段
print('小框（y 70–115、長 20–120 的色 7 水平線段）：')
for y in range(70, 118):
    x = 0
    while x < 470:
        if rows[y][x] == 7:
            x0 = x
            while x < 470 and rows[y][x] == 7:
                x += 1
            if 20 <= x - x0 <= 120:
                print('  y=%3d x %3d–%3d 長 %d' % (y, x0, x-1, x-x0))
        else:
            x += 1

# 逐列掃文字：每 20 像素一列，字高 15
print('\n文字（掃 y 72–400，每列試 15 高的格）：')
for y in range(72, 400):
    out, x = [], 24
    while x + 16 <= 460:
        hit = None
        for tab, wid, conv in ((full, 16, big5), (fullb, 16, big5)):
            b = S.cell_bits(rows, x, y, 7 if tab is full else 7, 16)
            if b in tab and any(b):
                hit = (conv(tab[b]), 16); break
        if not hit:
            for tab, conv in ((half, chr), (halfb, chr)):
                b = S.cell_bits(rows, x, y, 7, 8)
                if b in tab and any(b):
                    hit = (conv(tab[b]), 8); break
        if hit:
            out.append((x, hit[0])); x += hit[1]
        else:
            x += 8
    if out:
        print('  y=%3d %s' % (y, ' '.join('%d:%s' % (a, b) for a, b in out)))
