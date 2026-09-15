"""把 screen.bin（640×480 色號）配 pal.bin（256×3，6-bit）裁一塊放大成 PNG，
兩張並排（上：第一張，下：第二張）。
用法：crop_png.py pal.bin a.bin b.bin x0 y0 x1 y1 scale out.png"""
import sys, zlib, struct, pathlib
pal = pathlib.Path(sys.argv[1]).read_bytes()
a = pathlib.Path(sys.argv[2]).read_bytes()
b = pathlib.Path(sys.argv[3]).read_bytes()
x0, y0, x1, y1, sc = (int(v) for v in sys.argv[4:9])
out = sys.argv[9]
W = 640
w, h = (x1 - x0) * sc, (y1 - y0) * sc
rows = []
def rgb(i):
    return bytes(min(255, v*255//63) for v in pal[i*3:i*3+3])
for img in (a, b):
    for y in range(y0, y1):
        line = b''.join(rgb(img[y*W+x]) * sc for x in range(x0, x1))
        for _ in range(sc):
            rows.append(b'\0' + line)
    rows.append(b'\0' + b'\xff\0\0' * w)
raw = b''.join(rows)
H = len(rows)
def chunk(t, d):
    return struct.pack('>I', len(d)) + t + d + struct.pack('>I', zlib.crc32(t + d) & 0xffffffff)
png = b'\x89PNG\r\n\x1a\n' + chunk(b'IHDR', struct.pack('>IIBBBBB', w, H, 8, 2, 0, 0, 0)) + chunk(b'IDAT', zlib.compress(raw)) + chunk(b'IEND', b'')
open(out, 'wb').write(png)
print('寫出', out, w, H)
