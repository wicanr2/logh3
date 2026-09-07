"""讀畫面上某一格是哪個字（半形先試，再試全形）。
用法：readcell.py <screen.bin> <x> <y> <ink>"""
import sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import load_screen, font_lookup, cell_bits, GLYPH_W, HALF_W
rows, _ = load_screen(sys.argv[1])
x, y, ink = int(sys.argv[2]), int(sys.argv[3]), int(sys.argv[4])
for bold in (False, True):
    full, half = font_lookup('/orig/GRAPH.IMG', bold=bold)
    h = cell_bits(rows, x, y, ink, HALF_W)
    f = cell_bits(rows, x, y, ink, GLYPH_W)
    if any(h) and h in half:
        print('%s半形 %02X %r' % ('粗體' if bold else '', half[h], chr(half[h]))); break
    if any(f) and f in full:
        print('%s全形 %04X' % ('粗體' if bold else '', full[f])); break
else:
    print('查不到')
