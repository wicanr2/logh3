# 共用的 8-bit 索引 PNG 輸出，以及從 oracle 畫面 PNG 撈調色盤。
#
# 調色盤要從畫面 PNG 撈，不能直接用 pal.bin：色號 → 屬性暫存器 → DAC
# 這條鏈由執行器走完了，pal.bin 只有最後一段（DAC）。
import struct, zlib


def write_png(path, w, h, px, pal):
    raw = b''.join(b'\x00' + bytes(px[y * w:(y + 1) * w]) for y in range(h))

    def chunk(t, d):
        c = t + d
        return struct.pack('>I', len(d)) + c + struct.pack('>I', zlib.crc32(c))

    with open(path, 'wb') as fp:
        fp.write(b'\x89PNG\r\n\x1a\n')
        fp.write(chunk(b'IHDR', struct.pack('>IIBBBBB', w, h, 8, 3, 0, 0, 0)))
        fp.write(chunk(b'PLTE', bytes(pal)))
        fp.write(chunk(b'IDAT', zlib.compress(raw, 9)))
        fp.write(chunk(b'IEND', b''))


def game_palette(screen_png='/work/workplace/oracle/key1/screen.png'):
    png = open(screen_png, 'rb').read()
    i = 8
    while i < len(png):
        n = struct.unpack_from('>I', png, i)[0]
        if png[i + 4:i + 8] == b'PLTE':
            return (list(png[i + 8:i + 8 + n]) + [0] * 768)[:48]
        i += 12 + n
    return [0] * 48
