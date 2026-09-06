import sys, os
sys.path.insert(0, '/work/tools/scripts')
import png
img = open('/orig/GRAPH.IMG', 'rb').read()
os.makedirs('/work/workplace/assets/font', exist_ok=True)

def sheet(name, lo, hi, w, rows, per):
    stride = (w // 8) * rows
    n = (hi - lo) // stride
    nrow = (n + per - 1) // per
    s = [[0]*(per*w) for _ in range(nrow*rows)]
    for g in range(n):
        gx, gy = (g % per)*w, (g // per)*rows
        for y in range(rows):
            for xb in range(w // 8):
                b = img[lo + g*stride + y*(w//8) + xb]
                for x in range(8):
                    s[gy+y][gx+xb*8+x] = (b >> (7-x)) & 1
    png.write_mono(f'/work/workplace/assets/font/{name}.png', per*w, nrow*rows, s)
    print(f"{name}: 0x{lo:X}–0x{hi:X}，{n} 個 {w}×{rows}")
    return n

sheet('region-b-16x15', 3840, 12900, 16, 15, 48)
sheet('region-c-16x15', 12900, 27030, 16, 15, 64)
