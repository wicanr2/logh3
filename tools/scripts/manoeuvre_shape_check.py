# -*- coding: utf-8 -*-
"""驗艦隊選點的外形（`165F:004D`）逐像素對上收據（`docs/re/155`）。

    tools/py.sh tools/scripts/manoeuvre_shape_check.py [批次]

預設收據 `s3cmdrange`：第 28 艦隊，回轉 15、巡航 30 → 距離 240、
朝向 190，圓心（旗艦圖塊中心）(298,409)。

照原版走三步：`165F:004D` 產生點（積分那一圈是 **do-while**、`rem`
是 16 位元無號）→ `3870:08FF` 的 Cohen–Sutherland 裁切端點 →
`39D7:0936` 的四支畫線，每個像素 **XOR**（共用端點因此互相抵消）。

外形的墨色是 1，畫面上的顏色隨底色而定：底色 3 → 2、格線 6 → 7、
黑 0 → 1。所以不能拿單一顏色當篩選條件。
"""
import json
import pathlib
import sys
from collections import Counter

W, H = 640, 480
CX, CY = 298, 409
TURN, DIST, FACING = 15, 240, 190
X0, Y0, X1, Y1 = 16, 60, 463, 419          # 戰場，兩端都含
INK = 1

root = pathlib.Path('/work')
batch = sys.argv[1] if len(sys.argv) > 1 else 's3cmdrange'
scr = (root / ('workplace/oracle/%s/screen.bin' % batch)).read_bytes()
q = json.loads((root / 'docs/internals/data/sin-table.json').read_text(encoding='utf-8'))['quarter']


def sin(a):
    a &= 255
    if a <= 64:
        return q[a]
    if a <= 128:
        return q[128 - a]
    if a <= 192:
        return -q[a - 128]
    return -q[256 - a]


def cos(a):
    return sin(a + 64)


def sdiv(a, b):
    return (-1 if (a < 0) != (b < 0) else 1) * (abs(a) // abs(b))


def boundary(d, r, facing):
    """165F:004D"""
    half, i = [], 0
    while True:
        if sdiv(i * 16, r) > d:
            half.append((0, 0))          # 預算走完：累加值清成 0 再存一點
            break
        sx = sy = 0
        a, rem = 0, i & 0xFFFF
        while True:                      # do-while，rem 是 16 位元無號
            sx += cos(a) * 8
            sy += sin(a) * 8
            a += r
            rem = (rem - r) & 0xFFFF
            if rem < r:
                break
        frac = sdiv(rem * 8, r)
        sx += frac * cos(i)
        sy += frac * sin(i)
        st = d - sdiv(i * 16, r)
        sx += st * cos(i)
        sy += st * sin(i)
        if sy < 0:
            break                        # y 的累加值變負就停，這一點不存
        half.append((sx >> 10, sy >> 10))
        i += 8
    full = half + [(x, -y) for x, y in reversed(half)]      # 鏡射
    return [(((cos(facing) * x - sin(facing) * y + 512) >> 10),
             ((sin(facing) * x + cos(facing) * y + 512) >> 10)) for x, y in full]


def outcode(x, y):
    c = 0
    if x < X0:
        c |= 1
    elif x > X1:
        c |= 2
    if y < Y0:
        c |= 8
    elif y > Y1:
        c |= 4
    return c


def move(code, x, y, ox, oy):
    """3870:0A30，除法往零截斷"""
    if code & 8:
        x += sdiv((Y0 - y) * (x - ox), y - oy); y = Y0
    elif code & 4:
        x += sdiv((Y1 - y) * (x - ox), y - oy); y = Y1
    elif code & 1:
        y += sdiv((X0 - x) * (y - oy), x - ox); x = X0
    elif code & 2:
        y += sdiv((X1 - x) * (y - oy), x - ox); x = X1
    return x, y


def clip(x0, y0, x1, y1):
    """3870:08FF"""
    c0, c1 = outcode(x0, y0), outcode(x1, y1)
    for _ in range(8):
        if c0 & c1:
            return None
        if not c0 and not c1:
            return x0, y0, x1, y1
        if c0:
            x0, y0 = move(c0, x0, y0, x1, y1)
            c0 = outcode(x0, y0)
            if not c0 and not c1:
                return x0, y0, x1, y1
        if c0 & c1:
            return None
        if c1:
            x1, y1 = move(c1, x1, y1, x0, y0)
            c1 = outcode(x1, y1)
    return None


def draw(x0, y0, x1, y1):
    """39D7:0936 的四支；呼叫端先按 x 排好"""
    if x1 < x0:
        x0, y0, x1, y1 = x1, y1, x0, y0
    dx, dy = x1 - x0, y1 - y0
    ady, sy = abs(dy), (1 if dy >= 0 else -1)
    if dx > ady:
        step, acc, y = (ady << 16) // dx, 0x8000, y0
        for k in range(dx + 1):
            yield x0 + k, y
            acc += step
            if acc >= 0x10000:
                acc -= 0x10000; y += sy
    elif dx == ady:
        for k in range(dx + 1):
            yield x0 + k, y0 + k * sy
    elif dx == 0:
        for k in range(ady + 1):
            yield x0, y0 + k * sy
    else:
        step, acc, x = (dx << 16) // ady, 0x8000, x0
        for k in range(ady + 1):
            yield x, y0 + k * sy
            acc += step
            if acc >= 0x10000:
                acc -= 0x10000; x += 1


pts = boundary(DIST, TURN, FACING)
cnt = Counter()
for k in range(len(pts) - 1):
    seg = clip(CX + pts[k][0], CY + pts[k][1], CX + pts[k+1][0], CY + pts[k+1][1])
    if seg:
        for p in draw(*seg):
            cnt[p] += 1
lit = {p for p, c in cnt.items() if c % 2}       # XOR：畫偶數次等於沒畫
hist = Counter(scr[y * W + x] for x, y in lit)
bad = [(x, y, scr[y*W+x]) for x, y in sorted(lit)
       if scr[y*W+x] ^ INK not in (3, 6, 0)]
print('外形 %d 點，折線在戰場裡亮 %d 個像素' % (len(pts), len(lit)))
print('落點的顏色：', dict(sorted(hist.items())),
      '（底色 3→2、格線 6→7、黑 0→1）')
print('解釋不了的落點：%d %s' % (len(bad), bad))
real = {(x, y) for y in range(Y0, Y1+1) for x in range(X0, X1+1) if scr[y*W+x] == 2}
rest = real - lit
cells = Counter((x // 16, y // 16) for x, y in rest)
print('畫面上其餘 %d 個色 2 落在 %d 個 16×16 格裡（分隊的圖塊邊）'
      % (len(rest), len(cells)))
