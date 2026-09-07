"""替每個星域找一個確定落在它遮罩內的點（地圖內座標）。"""
import sys, pathlib
RGN = pathlib.Path(sys.argv[1]).read_bytes()
TW, TH, HDR = 16, 20, 192
regs = []
for i in range(0, HDR, 6):
    bx, by, bw, bh = RGN[i], RGN[i+1], RGN[i+2], RGN[i+3]
    off = RGN[i+4] | RGN[i+5] << 8
    regs.append((bx, by, bw, bh, HDR + off))
per = TW * TH // 8

def covers(r, x, y):
    bx, by, bw, bh, off = r
    x -= bx * TW; y -= by * TH
    if x < 0 or y < 0 or x >= bw * TW or y >= bh * TH:
        return False
    bi = (y // TH) * bw + x // TW
    k = off + bi * per + (y % TH) * (TW // 8) + (x % TW) // 8
    return RGN[k] & (0x80 >> (x % 8)) != 0

def at(x, y):
    for i, r in enumerate(regs):
        if covers(r, x, y):
            return i
    return -1

want = [int(v) for v in sys.argv[2].split(',')] if len(sys.argv) > 2 else range(32)
for k in want:
    bx, by, bw, bh, _ = regs[k]
    best = None
    for y in range(by * TH, (by + bh) * TH):
        for x in range(bx * TW, (bx + bw) * TW):
            if at(x, y) == k:
                # 取離邊界遠一點的：四周八格都還是同一個星域
                if all(at(x + dx, y + dy) == k for dx in (-6, 0, 6) for dy in (-6, 0, 6)):
                    best = (x, y); break
        if best: break
    print('星域 %2d → 地圖內 (%s)' % (k, best))
