import sys
sys.path.insert(0, '/work/tools/scripts')
from screenspec import load_screen, font_lookup, read_line
full, half = font_lookup('/orig/GRAPH.IMG')
batch = sys.argv[1]
x0 = int(sys.argv[2]) if len(sys.argv) > 2 else 80
x1 = int(sys.argv[3]) if len(sys.argv) > 3 else 636
rows, _ = load_screen(f'/work/workplace/oracle/{batch}/screen.bin')
for y in range(62, 420, 20):
    for ink in (7, 6, 5, 15):
        cells = read_line(rows, x0, y, x1, ink, full, half)
        if not cells:
            continue
        parts, prev = [], None
        for c in cells:
            ch = bytes.fromhex(c['hex']).decode('big5', 'replace') if c['w'] == 16 else chr(c['code'])
            if prev is None or c['x'] != prev:
                parts.append([c['x'], ''])
            parts[-1][1] += ch
            prev = c['x'] + c['w']
        print(y, ink, ' '.join(f'{x}:{t}' for x, t in parts))
