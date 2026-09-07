import struct, zlib, sys

def read_png(path):
    d = open(path, 'rb').read()
    assert d[:8] == b'\x89PNG\r\n\x1a\n'
    i, idat, pal, w, h, bd, ct = 8, b'', None, 0, 0, 0, 0
    while i < len(d):
        n = struct.unpack('>I', d[i:i+4])[0]; t = d[i+4:i+8]; c = d[i+8:i+8+n]
        if t == b'IHDR': w, h, bd, ct = struct.unpack('>IIBB', c[:10])
        elif t == b'PLTE': pal = c
        elif t == b'IDAT': idat += c
        i += 12 + n
    raw = zlib.decompress(idat)
    bpp = {0: 1, 2: 3, 3: 1, 6: 4}[ct]
    stride = (w * bd * (3 if ct == 2 else 1) + 7) // 8 if ct != 3 else (w * bd + 7)//8
    out, prev = [], bytearray(stride)
    p = 0
    for y in range(h):
        f = raw[p]; p += 1
        line = bytearray(raw[p:p+stride]); p += stride
        for x in range(stride):
            a = line[x-bpp] if x >= bpp else 0
            b = prev[x]
            c = prev[x-bpp] if x >= bpp else 0
            if f == 1: line[x] = (line[x] + a) & 255
            elif f == 2: line[x] = (line[x] + b) & 255
            elif f == 3: line[x] = (line[x] + (a + b)//2) & 255
            elif f == 4:
                pp = a + b - c
                pa, pb, pc = abs(pp-a), abs(pp-b), abs(pp-c)
                pr = a if (pa <= pb and pa <= pc) else (b if pb <= pc else c)
                line[x] = (line[x] + pr) & 255
        out.append(bytes(line)); prev = line
    return w, h, bd, ct, pal, out

def rgb(path):
    w, h, bd, ct, pal, rows = read_png(path)
    px = []
    for r in rows:
        row = []
        if ct == 3:
            if bd == 8:
                idx = list(r)
            else:
                idx = []
                per = 8 // bd
                for b in r:
                    for k in range(per):
                        idx.append((b >> (8 - bd*(k+1))) & ((1 << bd)-1))
                idx = idx[:w]
            for v in idx:
                row.append((pal[v*3], pal[v*3+1], pal[v*3+2]))
        else:
            for x in range(w):
                row.append((r[x*3], r[x*3+1], r[x*3+2]))
        px.append(row)
    return w, h, px

a = rgb(sys.argv[1]); b = rgb(sys.argv[2])
n = sum(1 for y in range(0, a[1], 4) for x in range(0, a[0], 4) if a[2][y][x] != b[2][y][x])
print("抽樣不同", n, "左上角", a[2][0][0], b[2][0][0])
