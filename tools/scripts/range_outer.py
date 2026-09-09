# -*- coding: utf-8 -*-
"""把 s3cmdrange 上的外形與虛線格線分開，再拿積分模型去對。

格線每 240 個世界像素一條（docs/internals/34），視野原點 (1040,540)。
旗艦圖塊中心 (298,409) ＝ 世界 (1338,949)。
"""
import math
W, H = 640, 480
scr = open('workplace/oracle/s3cmdrange/screen.bin', 'rb').read()
VX, VY = 1040, 540
CX, CY = 298, 409                      # 畫面上的圓心
WX, WY = CX + VX, CY + VY              # 世界座標的圓心
TURN, DIST = 15, 240                   # 艦隊 28 的回轉 15、巡航 30 × 8

pts = []
for y in range(40, 440):
    for x in range(16, 464):
        if scr[y*W+x] != 6:
            continue
        wx, wy = x + VX, y + VY
        if wx % 240 == 0 or wy % 240 == 0:
            continue                    # 格線
        pts.append((x, y))
print('扣掉格線之後的色 6 像素：%d 個' % len(pts))

SIN = [round(math.sin(i / 256 * 2 * math.pi) * 1024) for i in range(256)]

def boundary(turn, dist, facing):
    """照 165F:004D 積分出來的邊界點。"""
    out = []
    i = 0
    while (i * 16) // turn <= dist:
        sx = sy = 0
        a, rem = 0, i
        while rem >= turn:
            sx += SIN[(a + facing + 0x40) & 0xFF] << 3
            sy += SIN[(a + facing) & 0xFF] << 3
            a += turn
            rem -= turn
        frac = (rem << 3) // turn
        h = (i + facing) & 0xFF
        sx += frac * SIN[(h + 0x40) & 0xFF]
        sy += frac * SIN[h]
        straight = dist - (i * 16) // turn
        sx += straight * SIN[(h + 0x40) & 0xFF]
        sy += straight * SIN[h]
        out.append((CX + sx // 1024, CY + sy // 1024))
        i += 1
    return out

for facing in (190,):
    b = boundary(TURN, DIST, facing)
    print('模型算出 %d 個點，前五個：%s' % (len(b), b[:5]))
    near = 0
    for p in b:
        if any(abs(p[0]-q[0]) <= 2 and abs(p[1]-q[1]) <= 2 for q in pts):
            near += 1
    print('其中 %d 個落在收據的色 6 像素 2 格內（%d/%d）' % (near, near, len(b)))

# 收據上的點離圓心多遠、在哪個方向
rs = sorted(((math.hypot(x-CX, y-CY),
              int(((math.degrees(math.atan2(y-CY, x-CX)) + 360) % 360)/360*256))
             for x, y in pts), reverse=True)
print('最遠的十個（半徑, 256 制方位）：', [(round(r), a) for r, a in rs[:10]])
