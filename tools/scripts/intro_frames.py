"""把開頭動畫解成 PNG（`docs/re/36`、`docs/re/71`、`docs/re/72`）。

    tools/py.sh tools/scripts/intro_frames.py <輸出目錄> [要輸出的格,格,…]

一格一格 XOR 累加，指定的那幾格寫成 PNG。色盤用收據 `intro40` 的
`pal.bin`（16 色，每個分量 0–63）。輸出是一次性產物，不進 Git。
"""
import os
import sys
import zlib
import struct

sys.path.insert(0, '/work/tools/scripts')
import intro_decode as I

W, H, PL = 640, 227, 18160
out_dir = sys.argv[1]
want = {int(v) for v in sys.argv[2].split(',')} if len(sys.argv) > 2 else {0}
os.makedirs(out_dir, exist_ok=True)

# 收據的 `pal.bin` 已經是 0–255（灰階漸層 0 → 243），不要再乘 255/63。
pal = open('/work/workplace/oracle/intro40/pal.bin', 'rb').read()
rgb = bytes(pal[:48])


def png(path, pix):
    raw = b''.join(b'\x00' + bytes(pix[y * W:(y + 1) * W]) for y in range(H))
    def chunk(tag, data):
        c = tag + data
        return struct.pack('>I', len(data)) + c + struct.pack('>I', zlib.crc32(c))
    with open(path, 'wb') as f:
        f.write(b'\x89PNG\r\n\x1a\n')
        f.write(chunk(b'IHDR', struct.pack('>IIBBBBB', W, H, 8, 3, 0, 0, 0)))
        f.write(chunk(b'PLTE', rgb))
        f.write(chunk(b'IDAT', zlib.compress(raw, 9)))
        f.write(chunk(b'IEND', b''))


data = open(I.DATA, 'rb').read()
_, lens = I.frame_bounds()
planes = [bytearray(PL) for _ in range(4)]
pos = 0
for f in range(374):
    used = 0
    for p in range(4):
        used += I.decode_plane(data, pos + used, planes[p])
        used += 1
    pos += lens[f]
    if f not in want:
        continue
    pix = bytearray(W * H)
    for y in range(H):
        for bx in range(80):
            o = y * 80 + bx
            b0, b1, b2, b3 = planes[0][o], planes[1][o], planes[2][o], planes[3][o]
            for k in range(8):
                m = 0x80 >> k
                pix[y * W + bx * 8 + k] = ((b0 & m and 1) | (b1 & m and 2)
                                           | (b2 & m and 4) | (b3 & m and 8))
    png(os.path.join(out_dir, 'frame-%03d.png' % f), pix)
    print('寫出 frame-%03d.png' % f)
