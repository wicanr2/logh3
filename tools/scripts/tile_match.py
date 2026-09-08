"""把素材當成圖塊解開，逐塊到畫面上找。

    tools/py.sh tools/scripts/tile_match.py <批次> <檔名> <寬> <高>

`SPRITE.SPR` 那一套版面（`docs/internals/10`）：一列依序放 plane 0–3
各 `寬/8` bytes，一塊 `高` 列。

**為什麼要這樣搜**：如果框線是圖塊拼出來的，畫面上一段連續的 64 像素
在檔案裡本來就分散在四塊裡——直接拿畫面位元組去搜一定落空
（`find_in_asset.py` 就是這樣沒中的）。
"""
import pathlib
import sys

W = 640
batch, name, tw, th = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4])
scr = pathlib.Path('/work/workplace/oracle/%s/screen.bin' % batch).read_bytes()
data = pathlib.Path('/orig/' + name).read_bytes()

bpr = tw // 8                 # 一個 plane 一列幾 bytes
tile_bytes = bpr * 4 * th
n = len(data) // tile_bytes
print('%s：%d bytes ÷（%d×%d 一塊 %d bytes）＝ %d 塊，餘 %d'
      % (name, len(data), tw, th, tile_bytes, n, len(data) % tile_bytes))


def decode(i):
    out = []
    base = i * tile_bytes
    for y in range(th):
        row = [0] * tw
        for p in range(4):
            for b in range(bpr):
                v = data[base + y*bpr*4 + p*bpr + b]
                for k in range(8):
                    if v >> (7-k) & 1:
                        row[b*8 + k] |= 1 << p
        out.append(row)
    return out


# 畫面上所有 tw×th 的位置都試（步進 4，夠找到就好）
hits = 0
for i in range(n):
    t = decode(i)
    if len({c for r in t for c in r}) < 2:
        continue                      # 整塊同色，比中了也沒意義
    found = None
    for y in range(0, 480 - th, 4):
        for x in range(0, W - tw, 8):
            ok = True
            for ty in range(th):
                if scr[(y+ty)*W + x: (y+ty)*W + x + tw] != bytes(t[ty]):
                    ok = False
                    break
            if ok:
                found = (x, y)
                break
        if found:
            break
    if found:
        hits += 1
        if hits <= 12:
            print('  第 %3d 塊 → 畫面 (%d,%d)' % (i, found[0], found[1]))
print('命中 %d／%d 塊' % (hits, n))
