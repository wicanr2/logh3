"""星圖上的勢力色是哪個欄位決定的？

`STRMAP.RGN` 給出每個星域在畫面上的像素（`docs/re/17`）。星圖底色是黑
（色 3），被占領的星域填上勢力色：帝國是深藍（色 1）、同盟是棕（色 12）。
數每個星域裡這兩種顏色各有多少，得到歸屬，再回頭問 `STELLAR.DT*`：
**哪一個位移的值，和這 32 個歸屬是一對一的？**

用法：tools/py.sh tools/scripts/stellar_owner.py [畫面收據名]
"""
import collections, pathlib, struct, sys

W, H, CW, CH, SCRW, TOP = 320, 300, 16, 20, 640, 40
name = sys.argv[1] if len(sys.argv) > 1 else 'main1'
rgn = pathlib.Path('/orig/STRMAP.RGN').read_bytes()
scr = pathlib.Path('/work/workplace/oracle/%s/screen.bin' % name).read_bytes()

rows, off = [], 0
while True:
    x, y, w, h, o = struct.unpack_from('<BBBBH', rgn, off)
    if w == 0 or h == 0 or o != sum(r[2] * r[3] * 40 for r in rows):
        break
    rows.append((x, y, w, h, o)); off += 6
base = off

EMPIRE, ALLIANCE = 1, 12
owners = []
for gx, gy, gw, gh, o in rows:
    c = collections.Counter()
    for cy in range(gh):
        for cx in range(gw):
            cell = base + o + (cy * gw + cx) * 40
            for y in range(CH):
                word = rgn[cell + y * 2] << 8 | rgn[cell + y * 2 + 1]
                py = (gy + cy) * CH + y
                for k in range(CW):
                    if word >> (15 - k) & 1 and py < H:
                        c[scr[(TOP + py) * SCRW + (gx + cx) * CW + k]] += 1
    e, a = c[EMPIRE], c[ALLIANCE]
    owners.append(0 if max(e, a) * 20 < sum(c.values()) else (1 if e > a else 2))
print('每個星域的歸屬（0 ＝ 沒填色、1 ＝ 帝國、2 ＝ 同盟）：')
print(' ', owners)
print('  帝國 %d 個、同盟 %d 個、沒填色 %d 個'
      % (owners.count(1), owners.count(2), owners.count(0)))
lowbits = owners

for ver in ('DT8', 'DT9', 'DTA', 'DTX'):
    d = pathlib.Path('/orig/STELLAR.%s' % ver).read_bytes()
    rec = 126
    hits = []
    for pos in range(rec):
        vals = [d[i * rec + pos] for i in range(32)]
        m = collections.defaultdict(set)
        for v, c in zip(vals, lowbits):
            m[v].add(c)
        if all(len(s) == 1 for s in m.values()) and len(set(vals)) > 1:
            hits.append((pos, sorted(set(vals)), {k: list(v)[0] for k, v in sorted(m.items())}))
    print('\n%s：%d 個位移可以一對一決定顏色' % (ver, len(hits)))
    for pos, vals, mapping in hits[:6]:
        print('  +%3d  值 %s → 色 %s' % (pos, vals, mapping))
