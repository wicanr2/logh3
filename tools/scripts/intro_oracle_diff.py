"""把原版倒出來的動畫格與 remake 解出來的同一格逐位元組比。

    tools/py.sh tools/scripts/intro_oracle_diff.py <收據名> <輸出目錄> <格,格,…> [要跳過的格]

收據要有 `frames/introfull-f<格號>.bin`（`-dump-at` 順手寫的 640×350 色號
陣列，與調色盤無關——`docs/re/261`）。remake 這側用
`tools/scripts/intro_decode.py`（與 `internal/asset.IntroDecoder` 同一套
語法，`docs/internals/19`）累加到第 N 格，照 `screen.IntroPage` 放在
第 61 列，組成同樣的 640×350 頁面。

輸出：每一格三張 PNG（原版／remake／差分）＋ 一張逐格統計表。
調色盤直接沿用原版 PNG 的 PLTE，兩邊才看得出是不是同一張圖
（動畫播到一半原版會把 DAC 換成 `TITSPT.RGB`，硬寫一份會對不上）。
"""
import os
import struct
import sys
import zlib

sys.path.insert(0, '/work/tools/scripts')
import intro_decode as I

W, H, PL = 640, 227, 18160
PAGE_W, PAGE_H, TOP = 640, 350, 61


def read_plte(path):
    b = open(path, 'rb').read()
    i = 8
    while i < len(b):
        n = struct.unpack('>I', b[i:i + 4])[0]
        tag = b[i + 4:i + 8]
        if tag == b'PLTE':
            return b[i + 8:i + 8 + n]
        i += 12 + n
    return bytes(range(256)) * 3


def png(path, pix, w, h, plte):
    raw = b''.join(b'\x00' + bytes(pix[y * w:(y + 1) * w]) for y in range(h))

    def chunk(tag, data):
        c = tag + data
        return struct.pack('>I', len(data)) + c + struct.pack('>I', zlib.crc32(c))
    with open(path, 'wb') as f:
        f.write(b'\x89PNG\r\n\x1a\n')
        f.write(chunk(b'IHDR', struct.pack('>IIBBBBB', w, h, 8, 3, 0, 0, 0)))
        f.write(chunk(b'PLTE', (plte + bytes(768))[:768]))
        f.write(chunk(b'IDAT', zlib.compress(raw, 9)))
        f.write(chunk(b'IEND', b''))


def main():
    batch, out_dir = sys.argv[1], sys.argv[2]
    want = sorted({int(v) for v in sys.argv[3].split(',')})
    # 第四個參數是「解出來但不要 XOR 進去」的格（原版跳過的格）。
    skip = {int(v) for v in sys.argv[4].split(',')} if len(sys.argv) > 4 and sys.argv[4] else set()
    src = '/work/workplace/oracle/%s/frames' % batch
    os.makedirs(out_dir, exist_ok=True)

    data = open(I.DATA, 'rb').read()
    _, lens = I.frame_bounds()
    planes = [bytearray(PL) for _ in range(4)]
    pos = 0
    print('  格  原版非零  remake非零   不同   最大差  差異框')
    for f in range(max(want) + 1):
        used = 0
        if f in skip:
            # 跳過：不 XOR，只把檔案位移往前推（原版就是這樣漏掉那一格的）。
            pos += lens[f]
        else:
            for p in range(4):
                used += I.decode_plane(data, pos + used, planes[p])
                used += 1
            pos += lens[f]
        if f not in want:
            continue
        page = bytearray(PAGE_W * PAGE_H)
        for y in range(H):
            for bx in range(80):
                o = y * 80 + bx
                b0, b1, b2, b3 = planes[0][o], planes[1][o], planes[2][o], planes[3][o]
                for k in range(8):
                    m = 0x80 >> k
                    page[(TOP + y) * PAGE_W + bx * 8 + k] = (
                        (1 if b0 & m else 0) | (2 if b1 & m else 0)
                        | (4 if b2 & m else 0) | (8 if b3 & m else 0))
        name = 'introfull-f%03d' % f
        got = open(os.path.join(src, name + '.bin'), 'rb').read()
        plte = read_plte(os.path.join(src, name + '.png'))
        diff = bytearray(PAGE_W * PAGE_H)
        nd, mx = 0, 0
        x0, y0, x1, y1 = PAGE_W, PAGE_H, -1, -1
        for i in range(PAGE_W * PAGE_H):
            if got[i] != page[i]:
                nd += 1
                mx = max(mx, abs(got[i] - page[i]))
                diff[i] = 15
                x, y = i % PAGE_W, i // PAGE_W
                x0, y0, x1, y1 = min(x0, x), min(y0, y), max(x1, x), max(y1, y)
        png(os.path.join(out_dir, 'orig-f%03d.png' % f), got, PAGE_W, PAGE_H, plte)
        png(os.path.join(out_dir, 'remake-f%03d.png' % f), page, PAGE_W, PAGE_H, plte)
        png(os.path.join(out_dir, 'diff-f%03d.png' % f), diff, PAGE_W, PAGE_H,
            bytes([0, 0, 0] + [0] * 42 + [255, 0, 0]))
        box = '—' if x1 < 0 else '(%d,%d)-(%d,%d)' % (x0, y0, x1, y1)
        print('%4d %9d %10d %7d %8d  %s'
              % (f, sum(1 for v in got if v), sum(1 for v in page if v), nd, mx, box))


main()
