# .VO5 解碼的共用實作（版面與演算法見 docs/internals/18-vo5.md）。
import struct


def unpack_plane(d, p, w, h):
    buf = bytearray(w * h)
    di, col, left, prev = 0, 0, h, -1
    while True:
        v = d[p]; p += 1
        while True:
            buf[di] = v; di += w; prev = v; left -= 1
            if left == 0:
                left = h; di -= w * h - 1; col += 1
                if col == w:
                    return bytes(buf), p
            v2 = d[p]; p += 1
            if v2 != prev:
                v = v2
                continue
            n = d[p]; p += 1
            if n == 0:
                n = struct.unpack_from('<H', d, p)[0]; p += 2
            for _ in range(n - 1):
                buf[di] = prev; di += w; left -= 1
                if left == 0:
                    left = h; di -= w * h - 1; col += 1
                    if col == w:
                        return bytes(buf), p
            break


def decode(d):
    w = d[0]
    h = struct.unpack_from('<H', d, 1)[0]
    p, planes = 3, []
    for _ in range(4):
        pl, p = unpack_plane(d, p, w, h)
        planes.append(pl)
    px = bytearray(w * 8 * h)
    for y in range(h):
        base = y * w
        for xb in range(w):
            b0, b1, b2, b3 = (planes[k][base + xb] for k in range(4))
            o = y * w * 8 + xb * 8
            for i in range(8):
                m = 0x80 >> i
                px[o + i] = ((1 if b0 & m else 0) | (2 if b1 & m else 0) |
                             (4 if b2 & m else 0) | (8 if b3 & m else 0))
    return w * 8, h, bytes(px), p


def decode_file(path):
    w, h, px, _ = decode(open(path, 'rb').read())
    return w, h, px
