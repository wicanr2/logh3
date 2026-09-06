# 把解出來的 .VO5 跟原版畫的畫面對一次。ENDT40 在 #19.4M 讀進來、
# ENDT30 在 #20.24M，所以 20M 那一張只含第一張圖。
import sys
sys.path.insert(0, '/work/tools/scripts')
from vo5_decode import decode_file
W, H = 640, 480
for shot, name in (('e20', 'ENDT40'), ('e28', 'ENDT30')):
    screen = open('/work/workplace/explore/%s.bin' % shot, 'rb').read()
    w, h, px = decode_file('/orig/%s.VO5' % name)
    best = None
    for oy in range(0, H - h + 1):
        for ox in range(0, W - w + 1, 8):
            bad = 0
            for y in range(0, h, 13):
                row = screen[(oy + y) * W + ox:(oy + y) * W + ox + w]
                bad += sum(1 for a, b in zip(row, px[y * w:(y + 1) * w]) if a != b)
                if best and bad > best[0]:
                    break
            if best is None or bad < best[0]:
                best = (bad, ox, oy)
    bad, ox, oy = best
    full = sum(1 for y in range(h) for x in range(w)
               if screen[(oy + y) * W + ox + x] != px[y * w + x])
    print('%s ↔ %s：位置 (%d,%d) 全圖差異 %d / %d' % (name, shot, ox, oy, full, w * h))
