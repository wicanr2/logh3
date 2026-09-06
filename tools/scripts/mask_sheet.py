# 把單色（1 位元）遮罩檔畫成 PNG。
#   tools/py.sh tools/scripts/mask_sheet.py <檔名> <一列幾 byte> [放大]
import os, sys
sys.path.insert(0, '/work/tools/scripts')
from pngout import write_png

name, BPR = sys.argv[1], int(sys.argv[2])
SC = int(sys.argv[3]) if len(sys.argv) > 3 else 2
# 交錯：N ＝ 幾張圖輪流佔用連續的列（第 k 張拿第 k, k+N, k+2N… 列）。
IL = int(sys.argv[4]) if len(sys.argv) > 4 else 1
d = open('/orig/' + name, 'rb').read()
W, H = BPR * 8, len(d) // BPR
px = bytearray(W * H)
rows = list(range(H))
if IL > 1:
    rows = [k * (H // IL) + i for i in range(H // IL) for k in range(IL)]
    rows = [rows.index(y) for y in range(H)] if False else None
    # 反交錯：輸出第 k 張的第 i 列 ← 原檔的第 i*IL+k 列
    rows = [(i * IL + k) for k in range(IL) for i in range(H // IL)]
for out_y, src_y in enumerate(rows):
    for xb in range(BPR):
        b = d[src_y * BPR + xb]
        for i in range(8):
            if b & (0x80 >> i):
                px[out_y * W + xb * 8 + i] = 15
if SC > 1:
    big = bytearray(W * SC * H * SC)
    for y in range(H):
        row = bytes(v for p in px[y*W:(y+1)*W] for v in [p]*SC)
        for k in range(SC):
            o = (y*SC+k)*W*SC
            big[o:o+len(row)] = row
    px, W, H = big, W*SC, H*SC
pal = [0]*45 + [255, 255, 255]
out = '/work/workplace/assets/mask-%s-%d-%d.png' % (name.split('.')[0].lower(), BPR, IL)
os.makedirs('/work/workplace/assets', exist_ok=True)
write_png(out, W, H, px, pal)
print(name, len(d), 'bytes →', W//SC, '×', H//SC, '→', out)
