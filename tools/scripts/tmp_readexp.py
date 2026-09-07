import sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import load_screen, font_lookup, read_line
full, half = font_lookup('/orig/GRAPH.IMG')
rows, _ = load_screen('/work/workplace/oracle/expSortieRow0/screen.bin')
for y in (322,):
    cells = read_line(rows, 80, y, 636, 7, full, half)
    for c in cells:
        ch = bytes.fromhex(c['hex']).decode('big5','replace') if c['w']==16 else chr(c['code'])
        print(y, c['x'], c['w'], repr(ch))
