# -*- coding: utf-8 -*-
"""量收據 s3cmdrange 上色 6 的像素，看外形對稱軸在哪。

旗艦在世界 (1330,941)、畫面 (290,401)，圖塊中心 (298,409)。
色 6 同時是虛線格線，所以先用四鄰接連通把弧與格線分開。
"""
import math, collections
W, H = 640, 480
scr = open('workplace/oracle/s3cmdrange/screen.bin', 'rb').read()
CX, CY = 298, 409

pts = [(x, y) for y in range(40, 440) for x in range(16, 464)
       if scr[y*W+x] == 6]
# 連通塊
seen, blocks = set(), []
for p in pts:
    if p in seen:
        continue
    stack, blk = [p], []
    seen.add(p)
    while stack:
        x, y = stack.pop()
        blk.append((x, y))
        for dx, dy in ((1,0),(-1,0),(0,1),(0,-1)):
            q = (x+dx, y+dy)
            if q not in seen and scr[q[1]*W+q[0]] == 6 and 16 <= q[0] < 464 and 40 <= q[1] < 440:
                seen.add(q); stack.append(q)
    blocks.append(blk)
blocks.sort(key=len, reverse=True)
print('色 6 的像素 %d 個，連通塊 %d 個，最大的幾塊大小：%s'
      % (len(pts), len(blocks), [len(b) for b in blocks[:6]]))

for n, blk in enumerate(blocks[:3]):
    r = [math.hypot(x-CX, y-CY) for x, y in blk]
    a = [(math.degrees(math.atan2(y-CY, x-CX)) + 360) % 360 for x, y in blk]
    print('塊 %d：%d 點  半徑 %.1f–%.1f（平均 %.1f）  角度 %.0f–%.0f'
          % (n, len(blk), min(r), max(r), sum(r)/len(r), min(a), max(a)))
    # 半徑最大的方向 ＝ 對稱軸的候選
    far = max(zip(r, a))
    print('     最遠點 半徑 %.1f 在 %.0f°（256 制 %.0f）' % (far[0], far[1], far[1]/360*256))

print()
print('== 依角度看色 6 的最大半徑（每 15° 一格）==')
import math
buck = {}
for x, y in pts:
    r = math.hypot(x-CX, y-CY)
    if r < 20:
        continue
    a = int(((math.degrees(math.atan2(y-CY, x-CX)) + 360) % 360) // 15)
    buck.setdefault(a, []).append(r)
for a in sorted(buck):
    rs = sorted(buck[a])
    print('%3d°–%3d°  n=%3d  最大 %5.1f  次大 %5.1f'
          % (a*15, a*15+15, len(rs), rs[-1], rs[-2] if len(rs) > 1 else -1))
