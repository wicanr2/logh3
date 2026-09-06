"""在字型表裡找「畫面上這一格」最像哪個字形＋哪種變換。

用法：glyph_match.py <screen.bin> <font> <x0> <x1> <y0> <y1> <ink...>
對 (x,y) 各一格做全表搜尋，印出每種變換的最小差異。
"""
import sys, pathlib
W, GW, GH, STRIDE = 640, 16, 15, 30
M = 0xFFFF

def bold_r(g):  return [(v | (v >> 1)) & M for v in g]
def bold_l(g):  return [(v | (v << 1)) & M for v in g]
def dbl(g):
    out = []
    for v in g:
        w = 0
        for i in range(8):          # 只有左半 8 欄放大兩倍
            b = v >> (15 - i) & 1
            w = (w << 2) | (b * 3)
        out.append(w & M)
    return out
def sh_ul(g):   # ∪ 左上一格（下一列左移一格疊上來）
    return [(g[r] | ((g[r+1] << 1) & M if r + 1 < GH else 0)) & M for r in range(GH)]
def sh_dr(g):
    return [(g[r] | ((g[r-1] >> 1) if r else 0)) & M for r in range(GH)]

BASE = {'原樣': lambda g: g, '粗右': bold_r, '粗左': bold_l, '橫向2倍': dbl}
POST = {'': lambda g: g, '+左上': sh_ul, '+右下': sh_dr}

def cell(scr, x, y, inks):
    return [sum((1 if scr[(y+r)*W+x+c] in inks else 0) << (15-c) for c in range(GW))
            for r in range(GH)]

def main():
    scr = pathlib.Path(sys.argv[1]).read_bytes()
    fnt = pathlib.Path(sys.argv[2]).read_bytes()
    x0, x1, y0, y1 = (int(a) for a in sys.argv[3:7])
    inks = set(int(a) for a in sys.argv[7:])
    glyphs = []
    for off in range(0, len(fnt) - STRIDE + 1, STRIDE):
        g = [fnt[off+2*r] << 8 | fnt[off+2*r+1] for r in range(GH)]
        if any(g):
            glyphs.append((off, g))
    print('候選字形 %d 個' % len(glyphs))
    best = None
    for y in range(y0, y1):
        for x in range(x0, x1):
            t = cell(scr, x, y, inks)
            for off, g in glyphs:
                for bn, bf in BASE.items():
                    b = bf(g)
                    for pn, pf in POST.items():
                        v = pf(b)
                        d = sum(bin((v[r] ^ t[r]) & M).count('1') for r in range(GH))
                        if best is None or d < best[0]:
                            best = (d, x, y, off, bn + pn, v, t)
    d, x, y, off, name, v, t = best
    print('最小差 %d 位元 @ (%d,%d) 位移 %d(0x%X) 變換「%s」' % (d, x, y, off, off, name))
    for r in range(GH):
        f = lambda w: ''.join('#' if w >> (15-i) & 1 else '.' for i in range(16))
        print('  %s   %s' % (f(t[r]), f(v[r])))

main()
