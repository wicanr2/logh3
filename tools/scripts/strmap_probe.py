"""`STRMAP.BIT` 與主畫面星圖是什麼關係？

48,000 bytes 同時符合 packed 4bpp 與 planar 兩種 16 色版面，
所以「大小對得上」證明不了任何事。而且畫面上的星圖會**跟著治安換色**
（戰略情報報告那張），那表示底圖存的很可能不是顏色而是**索引**，
顏色是查表出來的。

所以這裡問的不是「解出來像不像」，而是「同一個索引值在畫面上是不是
同一個顏色」——一致性高就是查表關係，低就是解讀錯了。

用法：tools/py.sh tools/scripts/strmap_probe.py [畫面收據名]
"""
import collections, pathlib, sys
sys.path.insert(0, '/work/tools/scripts')
from pngout import write_png, game_palette

W, H = 320, 300
SCREEN_W = 640
name = sys.argv[1] if len(sys.argv) > 1 else 'main1'

bit = pathlib.Path('/orig/STRMAP.BIT').read_bytes()
scr = pathlib.Path('/work/workplace/oracle/%s/screen.bin' % name).read_bytes()
pal = game_palette('/work/workplace/oracle/%s/screen.png' % name)

def packed():
    px = bytearray(W * H)
    for i in range(W * H // 2):
        px[i * 2], px[i * 2 + 1] = bit[i] >> 4, bit[i] & 15
    return px

def packed_lo_first():
    px = bytearray(W * H)
    for i in range(W * H // 2):
        px[i * 2], px[i * 2 + 1] = bit[i] & 15, bit[i] >> 4
    return px

def planar():
    px = bytearray(W * H)
    plane = W * H // 8
    for p in range(4):
        for i in range(plane):
            b = bit[p * plane + i]
            for k in range(8):
                if b >> (7 - k) & 1:
                    px[i * 8 + k] |= 1 << p
    return px

def consistency(px, ox, oy):
    """同一個索引在畫面上是不是同一個顏色。回傳加權一致率。"""
    hist = collections.defaultdict(collections.Counter)
    for y in range(0, H, 2):
        sy = oy + y
        if not 0 <= sy < 480:
            continue
        for x in range(0, W, 2):
            sx = ox + x
            if not 0 <= sx < SCREEN_W:
                continue
            hist[px[y * W + x]][scr[sy * SCREEN_W + sx]] += 1
    total = sum(sum(c.values()) for c in hist.values())
    if not total:
        return 0.0, 0
    best = sum(c.most_common(1)[0][1] for c in hist.values())
    return best / total, len(hist)

for label, f in (('packed(高位在前)', packed), ('packed(低位在前)', packed_lo_first),
                 ('planar', planar)):
    px = f()
    top = max(((consistency(px, ox, oy)[0], ox, oy)
               for oy in range(36, 45) for ox in range(0, 9)), default=(0, 0, 0))
    rate, ox, oy = top
    _, nidx = consistency(px, ox, oy)
    print('%-16s 最佳對齊 (%d,%d)  索引→顏色一致率 %.1f%%（%d 種索引）'
          % (label, ox, oy, rate * 100, nidx))
    write_png('/work/workplace/strmap-%s.png' % label.split('(')[0], W, H, px, pal)
