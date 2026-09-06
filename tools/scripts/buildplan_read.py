# 從原版 dump 讀出建造計劃的百分比／金額欄位。
# 半形字形取自 GRAPH.IMG，用字色 7 的位元圖比對。
import json, sys
W = 640
GH, HW = 15, 8

spec = json.load(open('/work/docs/internals/data/screen-build-plan.json'))
F = spec['fields']
INK = spec['windows']['report']['ink']

fnt = open('/orig/GRAPH.IMG', 'rb').read()
meta = json.load(open('/work/docs/internals/data/font-graph-img.json'))
half_off = 0  # 半形字區從檔案開頭起，256 個字 × 15 bytes（font-graph-img.json）
if False:
    print(json.dumps(meta)[:400]); sys.exit(1)

def glyph(code):
    o = half_off + code * GH
    return list(fnt[o:o + GH])

pats = {}
for ch in "0123456789%":
    pats[tuple(glyph(ord(ch)))] = ch

def read(img, x, y, w):
    """把 x..x+w 這一段照 8 px 切格，逐格比對字形。"""
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
    img = open(name, "rb").read()
    got = []
    for r in F['rows']:
        p = read(img, 160, r['y'], 32).strip()
        a = read(img, 280, r['y'], 40).strip()
        got.append((p, a))
    rp = read(img, 160, F['rest']['y'], 32).strip()
    ra = read(img, 280, F['rest']['y'], 40).strip()
    cost = read(img, 280, F['cost']['y'], 40).strip()
    print(name.split('/')[-1], "預算", cost)
    print("   ", got, "餘", rp, ra)
