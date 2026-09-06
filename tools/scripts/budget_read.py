# 讀預算計劃畫面上的四個數字（半形字形比對）。
import json, sys
W, GH, HW = 640, 15, 8
spec = json.load(open('/work/docs/internals/data/screen-budget-plan.json'))
INK = spec['windows']['report']['ink']
fnt = open('/orig/GRAPH.IMG', 'rb').read()
pats = {}
for ch in "0123456789%":
    o = ord(ch) * GH
    pats[tuple(fnt[o:o + GH])] = ch
def read(img, x, y, w):
    out = ""
    for cx in range(x, x + w, HW):
        rows = []
        for gy in range(GH):
            v = 0
            for gx in range(HW):
                if img[(y + gy) * W + cx + gx] == INK:
                    v |= 1 << (HW - 1 - gx)
            rows.append(v)
        out += pats.get(tuple(rows), "?" if any(rows) else " ")
    return out
for name in sys.argv[1:]:
    img = open(name, 'rb').read()
    print(name.split('/')[-1])
    for y, label in ((282, '預估徵稅額'), (302, '預估軍事費'),
                     (322, '船艦建造費'), (342, '剩餘軍事費')):
        print('   %s |%s|' % (label, read(img, 300, y, 200)))
