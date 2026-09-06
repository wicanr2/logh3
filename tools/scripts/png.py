"""最小 PNG 輸出（不依賴任何套件）。索引色與 RGB 兩種。"""
import struct, zlib

def _chunk(t, d):
    c = t + d
    return struct.pack('>I', len(d)) + c + struct.pack('>I', zlib.crc32(c) & 0xFFFFFFFF)

def write_rgb(path, w, h, rgb_rows):
    raw = b''.join(b'\x00' + bytes(r) for r in rgb_rows)
    png = (b'\x89PNG\r\n\x1a\n'
           + _chunk(b'IHDR', struct.pack('>IIBBBBB', w, h, 8, 2, 0, 0, 0))
           + _chunk(b'IDAT', zlib.compress(raw, 9)) + _chunk(b'IEND', b''))
    open(path, 'wb').write(png)

def write_mono(path, w, h, bit_rows, fg=(255, 255, 255), bg=(0, 0, 0)):
    """bit_rows：每列一個 list[0/1]。"""
    rows = []
    for r in bit_rows:
        line = bytearray()
        for v in r:
            line += bytes(fg if v else bg)
        rows.append(line)
    write_rgb(path, w, h, rows)
