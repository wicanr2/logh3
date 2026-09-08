# -*- coding: utf-8 -*-
"""在 s3form 的畫面上找隊形標記：拿 41 格的世界座標減視野原點去取 16×16。"""
import json, struct, collections

W, H = 640, 480
scr = open('workplace/oracle/s3form/screen.bin', 'rb').read()
tab = open('workplace/oracle/s3form/form_tab.bin', 'rb').read()

def u16(b, o): return struct.unpack_from('<H', b, o)[0]

# 視野原點：艦隊 31 旗艦世界 (1320,812)，收據上畫在 (280,272) → (1040,540)
VX, VY = 1040, 540

def block(x, y):
    if x < 0 or y < 0 or x + 16 > W or y + 16 > H:
        return None
    return b''.join(scr[(y + r) * W + x:(y + r) * W + x + 16] for r in range(16))

for dy in (0, 40):
    sigs = collections.Counter()
    free = []
    for i in range(41):
        x, y, flag = u16(tab, i * 10), u16(tab, i * 10 + 2), tab[i * 10 + 4]
        if flag:
            continue
        b = block(x - VX, y - VY + dy)
        if b is None:
            continue
        free.append((i, x - VX, y - VY + dy, b))
    for i, sx, sy, b in free:
        sigs[b] += 1
    print('畫面 y 位移 %d：可選格 %d 個，出現 %d 種 16×16 圖樣'
          % (dy, len(free), len(sigs)))
    for b, n in sigs.most_common(3):
        colours = sorted(set(b))
        print('   出現 %2d 次，用到色號 %s' % (n, colours[:10]))

print()
print('== 位移 0 的細節 ==')
sig = collections.Counter()
blocks = {}
for i in range(41):
    x, y, flag = u16(tab, i * 10), u16(tab, i * 10 + 2), tab[i * 10 + 4]
    if flag:
        continue
    b = block(x - VX, y - VY)
    if b is None:
        print('槽 %2d 在畫面外' % i); continue
    blocks[i] = (x - VX, y - VY, b)
    sig[b] += 1
common = sig.most_common(1)[0][0]
taken = [i for i in range(41) if tab[i * 10 + 4]]
tk = [(u16(tab, i * 10) - VX, u16(tab, i * 10 + 2) - VY) for i in taken]
for i, (sx, sy, b) in sorted(blocks.items()):
    if b == common:
        continue
    over = [t for t in tk if abs(t[0] - sx) < 16 and abs(t[1] - sy) < 16]
    print('槽 %2d 在 (%3d,%3d) 不同；與有人的格子重疊：%s' % (i, sx, sy, over or '無'))
print('標記本身（色號分布）：', collections.Counter(common).most_common())

print()
def show(b, label):
    print(label)
    for r in range(16):
        print('   ' + ''.join('%X' % (c & 0xF) if c else '.' for c in b[r*16:(r+1)*16]))
show(common, '標記（30 格共通）')
for i in (15, 16, 19, 24, 25):
    show(blocks[i][2], '槽 %d 在 (%d,%d)' % (i, blocks[i][0], blocks[i][1]))
