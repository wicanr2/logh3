import sys
f = sys.argv[1]; n = int(sys.argv[2]) if len(sys.argv) > 2 else 60
start = int(sys.argv[3], 16) if len(sys.argv) > 3 else 0x100
d = open('/orig/' + f, 'rb').read()
NAMES = {0xE0: 'PROG', 0xE2: 'CC', 0xE4: '?E4', 0xFA: 'TEMPO', 0xFB: '?FB',
         0xFC: '?FC', 0xFD: 'BAR', 0xFE: 'SEP'}
p, bar, tick, shown = start, 0, 0, 0
while p + 4 <= len(d) and shown < n:
    b, a = d[p], list(d[p+1:p+4])
    if b < 0x80:
        print('%04X 小節%3d tick%4d 音符 %3d 步%3d 長%3d 力%3d' % (p, bar, tick, b, a[0], a[1], a[2]))
        tick += a[0]
    else:
        print('%04X 小節%3d tick%4d %-5s %s' % (p, bar, tick, NAMES.get(b, hex(b)), a))
        if b == 0xFD:
            bar += a[0]; tick = 0
        if b == 0xFE:
            bar, tick = 0, 0
    p += 4
    shown += 1
