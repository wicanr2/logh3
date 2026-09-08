"""把 dosgolem `-dump-at` 產生的 PNG 解回色號陣列。

    tools/py.sh tools/scripts/png_indices.py <png> <out.bin>

`-dump-at` 只寫 PNG（沒有 `.bin` 那一份），但版面要量的是色號不是 RGB。
這支只處理 8-bit 調色盤、非交錯的 PNG——dosgolem 產生的就是那種。
"""
import pathlib
import struct
import sys
import zlib

src, dst = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
d = src.read_bytes()
assert d[:8] == b'\x89PNG\r\n\x1a\n', '不是 PNG'
i, idat, w, h, depth, ctype = 8, b'', 0, 0, 0, 0
while i < len(d):
    ln = struct.unpack('>I', d[i:i + 4])[0]
    typ = d[i + 4:i + 8]
    body = d[i + 8:i + 8 + ln]
    if typ == b'IHDR':
        w, h, depth, ctype = struct.unpack('>IIBB', body[:10])
    elif typ == b'IDAT':
        idat += body
    i += 12 + ln
assert depth == 8 and ctype == 3, '只處理 8-bit 調色盤：depth=%d ctype=%d' % (depth, ctype)
raw = zlib.decompress(idat)
out = bytearray(w * h)
prev = bytearray(w)
p = 0
for y in range(h):
    f = raw[p]; p += 1
    line = bytearray(raw[p:p + w]); p += w
    if f == 1:
        for x in range(1, w):
            line[x] = (line[x] + line[x - 1]) & 0xFF
    elif f == 2:
        for x in range(w):
            line[x] = (line[x] + prev[x]) & 0xFF
    elif f == 3:
        for x in range(w):
            a = line[x - 1] if x else 0
            line[x] = (line[x] + ((a + prev[x]) >> 1)) & 0xFF
    elif f == 4:
        for x in range(w):
            a = line[x - 1] if x else 0
            c = prev[x - 1] if x else 0
            pa, pb, pc = abs(prev[x] - c), abs(a - c), abs(a + prev[x] - 2 * c)
            pr = a if (pa <= pb and pa <= pc) else (prev[x] if pb <= pc else c)
            line[x] = (line[x] + pr) & 0xFF
    out[y * w:(y + 1) * w] = line
    prev = line
dst.write_bytes(bytes(out))
print('%s → %s（%d×%d）' % (src.name, dst.name, w, h))
