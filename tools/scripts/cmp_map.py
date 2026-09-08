"""解 `MAP<nn>.CMP`（一個星域一張戰場地圖）並到畫面上對位。

    tools/py.sh tools/scripts/cmp_map.py <MAPnn.CMP> [批次] [圖塊檔]

編碼（`docs/re/81`）：

    >= 0x80  一格，圖塊編號 ＝ byte & 0x7F
    <  0x80  兩個 byte：(圖塊, 次數)，重複那麼多格

32 張全部展開成 **13,056 格 ＝ 128 欄 × 102 列**；圖塊 16×20，
所以整張戰場是 2048 × 2040——與艦隊記錄裡量到的戰場座標對得上
（`docs/internals/34`）。

給了批次就把畫面上的戰場 (16,60) 28×18 那一塊拿去比，找出視野的位置。
"""
import pathlib
import sys

COLS, ROWS = 128, 102
TW, TH = 16, 20
VIEW_C, VIEW_R = 28, 18
FIELD_X, FIELD_Y = 16, 60
W = 640


def unpack(data):
    out = []
    i = 0
    while i < len(data):
        if data[i] >= 0x80:
            out.append(data[i] & 0x7F)
            i += 1
        elif i + 1 < len(data):
            out.extend([data[i]] * data[i+1])
            i += 2
        else:
            out.append(data[i])
            i += 1
    return out


def tiles(path, n=128):
    d = pathlib.Path(path).read_bytes()
    size = 2 * 4 * TH
    out = []
    for k in range(n):
        base = k * size
        img = []
        for y in range(TH):
            row = [0] * TW
            for p in range(4):
                for b in range(2):
                    v = d[base + y*8 + p*2 + b]
                    for j in range(8):
                        if v >> (7-j) & 1:
                            row[b*8 + j] |= 1 << p
            img.append(row)
        out.append(img)
    return out


cells = unpack(pathlib.Path(sys.argv[1]).read_bytes())
print('%s → %d 格（%d 欄 × %d 列）' % (sys.argv[1], len(cells), COLS, ROWS))
print('  用到的圖塊編號 %d 種，最小 %d、最大 %d'
      % (len(set(cells)), min(cells), max(cells)))
if len(sys.argv) < 3:
    sys.exit(0)

scr = pathlib.Path('/work/workplace/oracle/%s/screen.bin' % sys.argv[2]).read_bytes()
tl = tiles('/orig/' + (sys.argv[3] if len(sys.argv) > 3 else 'BLOCK_2.BLK'))

best, bestBad = None, 1 << 30
for r0 in range(ROWS - VIEW_R + 1):
    for c0 in range(COLS - VIEW_C + 1):
        bad = 0
        for r in range(VIEW_R):
            for c in range(VIEW_C):
                t = tl[cells[(r0+r)*COLS + c0+c] % len(tl)]
                for y in range(TH):
                    sy = FIELD_Y + r*TH + y
                    sx = FIELD_X + c*TW
                    if scr[sy*W + sx: sy*W + sx + TW] != bytes(t[y]):
                        bad += 1
            if bad >= bestBad:
                break
        if bad < bestBad:
            best, bestBad = (c0, r0), bad
print('  最合的視野起點 %s，不合的列 %d／%d' % (best, bestBad, VIEW_R*VIEW_C*TH))

# 哪幾格對不上——**列出來才知道差異有沒有規律**（是不是都在某一欄／某一列）。
c0, r0 = best
bad = []
for r in range(VIEW_R):
    for c in range(VIEW_C):
        idx = cells[(r0+r)*COLS + c0+c]
        t = tl[idx % len(tl)]
        n = 0
        for y in range(TH):
            sy = FIELD_Y + r*TH + y
            sx = FIELD_X + c*TW
            if scr[sy*W + sx: sy*W + sx + TW] != bytes(t[y]):
                n += 1
        if n:
            bad.append((r, c, idx, n))
print('  對不上的格子 %d／%d' % (len(bad), VIEW_R*VIEW_C))
rows = sorted({r for r, _, _, _ in bad})
colsb = sorted({c for _, c, _, _ in bad})
print('  落在列 %s' % rows)
print('  落在欄 %s' % colsb)
import collections
print('  對不上的圖塊編號 %s' % collections.Counter(i for _, _, i, _ in bad).most_common(8))
