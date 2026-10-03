"""第70輪可丟棄畫面重建探針，不接正式 renderer。"""
import json
from pathlib import Path

root = Path('/work/workplace/oracle/blocks48-r70-a')
report = json.loads((root/'report.json').read_text())
events = report['events']
cursor = 0
screen = bytearray()
results = []
map_mask = 15


def blit(screen, ram, a):
    x, y, wb, h, count, off, seg = a[2:9]
    addr = seg*16+off
    for n in range(count):
        for sy in range(h):
            dy = y+40-4*n+sy
            if not 208 <= dy <= 429:
                continue
            for sx in range(wb*8):
                dx = x+sx
                if not 8 <= dx <= 615:
                    continue
                v = sum(((ram[addr+sy*wb*4+plane*wb+sx//8] >> (7-sx%8)) & 1) << plane
                        for plane in range(4))
                if v:
                    screen[dy*640+dx] = v


def line(screen, a, map_mask):
    x0, y0, x1, y1, color, mask = a[2:8]
    y0 += 40
    y1 += 40
    color &= 15
    if x0 > x1:
        x0, y0, x1, y1 = x1, y1, x0, y0
    dx, dy = x1-x0, abs(y1-y0)
    sy = 1 if y0 < y1 else -1
    major, minor = max(dx, dy), min(dx, dy)
    ratio = minor*65536//major if major else 0
    frac = 32768
    for _ in range(major+1):
        if 0 <= x0 < 640 and 0 <= y0 < 480:
            screen[y0*640+x0] = (screen[y0*640+x0] & (15 ^ map_mask)) | (color & map_mask)
        if dx >= dy:
            x0 += 1
        else:
            y0 += sy
        frac += ratio
        if frac >= 65536:
            frac -= 65536
            if dx >= dy:
                y0 += sy
            else:
                x0 += 1


for frame in report['frames']:
    raw = (root/(frame['name']+'-screen.bin')).read_bytes()
    if frame['stage'] == 'entry':
        screen = bytearray(raw)
    ram = (root/(frame['name']+'-mem.bin')).read_bytes()
    while cursor < len(events) and events[cursor]['step'] < frame['step']:
        ev = events[cursor]
        cursor += 1
        if ev['original_location'] == '2376:0AA5':
            map_mask = 15
            for y in range(208, 432):
                for x in range(8, 632):
                    screen[y*640+x] = 14 if (y-208)%12 == 0 or (x-8+y-208)%24 == 0 else 0
        elif ev['original_location'] == '2376:07FA':
            blit(screen, ram, ev['stack_words'])
            map_mask = 8
        elif ev['original_location'] == '3870:07C9':
            line(screen, ev['stack_words'], map_mask)
            map_mask = 15
    changed = [i for i, (a, b) in enumerate(zip(screen, raw)) if a != b]
    results.append({'name': frame['name'], 'different': len(changed),
                    'first': [{'x': i%640, 'y': i//640, 'probe': screen[i], 'original': raw[i]}
                              for i in changed[:10]]})
edges = []
ep = root.parent/'blocks48-r70-edges-valid-a'
if (ep/'report.json').exists():
    er = json.loads((ep/'report.json').read_text())
    bg = bytearray((root/'scene-01-02-background-screen.bin').read_bytes())
    ram = (root/'scene-01-01-entry-mem.bin').read_bytes()
    for c in er['cases']:
        got = bytearray(bg)
        blit(got, ram, c['stack_words'])
        want = (ep/(c['name']+'-screen.bin')).read_bytes()
        diff = [i for i, (a, b) in enumerate(zip(got, want)) if a != b]
        edges.append({'name': c['name'], 'different': len(diff),
                      'first': [{'x': i%640, 'y': i//640, 'probe': got[i], 'original': want[i]}
                                for i in diff[:10]]})
print(json.dumps({'frames': len(results), 'different': sum(x['different'] for x in results),
                  'results': results, 'edges': edges}, ensure_ascii=False, indent=2))
