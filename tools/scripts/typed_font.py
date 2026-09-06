# 畫面上那三個數字用的不是 GRAPH.IMG 的半形字型。把點陣拿去所有原版檔裡找。
import os
b = open('/work/workplace/oracle/typed138/screen.bin', 'rb').read()
W = 640
files = {n: open('/orig/'+n, 'rb').read() for n in sorted(os.listdir('/orig'))
         if os.path.isfile('/orig/'+n)}
for x0, y0, h, w in ((243, 267, 15, 8), (243, 267, 16, 8), (244, 268, 14, 8),
                     (243, 267, 15, 16), (243, 266, 16, 16)):
    rows = []
    for y in range(h):
        v = 0
        for x in range(w):
            if b[(y0+y)*W + x0+x] == 15:
                v |= (0x8000 if w == 16 else 0x80) >> x
        rows.append(v)
    pat = b''.join(bytes([r >> 8, r & 0xFF]) if w == 16 else bytes([r]) for r in rows)
    hits = [(n, d.find(pat)) for n, d in files.items() if pat in d]
    print(f'({x0},{y0}) {w}×{h} → {hits[:4] if hits else "沒中"}')
