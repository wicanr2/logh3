# COMANI.BMG（交戰演出的動畫庫）→ 每段每一格的 PNG ＋ 索引 JSON。
#
#   tools/py.sh tools/scripts/bmg_export.py
#
# 格式與證據：docs/internals/20-comani-bmg.md、docs/re/328-comani-bmg-codec.md；Go 版
# 解碼器 internal/asset/bmg.go（TestBMGEngagementMatchesOracle 用 eng-* 收據逐格對拍）。
#   - 目錄的位移**從檔案 +2 算起**（原版 seek 到「位移＋2」）。
#   - 每段 ＝ u16 格數 ＋ 一串格；每格 10 bytes 格頭 ＋ 壓縮資料。
#   - 格種 0：`.AMG`／`ANI_PNL` 那一套（348A:02F9），**覆寫**。
#   - 格種 1：一 byte 一列、0 後面接「跳過幾列」（可跨直行／plane），非 0 的 byte **XOR**（348A:03A3）。
#   - 格頭的位移是 mode 12h 的 VRAM 位元組位移（列×80＋位元組欄），相對於視窗原點。
#
# 輸出：assets/original/comani/<段名>/fNN.png ＝「畫完第 NN 格之後」視窗 256×102 的樣子
# （段名第二個字元不全是 ASCII，寫成前兩字元的十六進位 ＋ 後兩字元）；索引
# docs/internals/data/comani-bmg.json。開頭是 XOR 格的段，底圖用同一個單位的 D0 第 0 格
# ——PPAB 疊 PPD0 有收據，其餘是推論，JSON 的 base 欄寫明。
import hashlib, json, os, struct, zlib

SRC = '/orig/COMANI.BMG'
OUT = '/work/assets/original/comani'
INDEX = '/work/docs/internals/data/comani-bmg.json'
PAL_PNG = '/work/workplace/oracle/eng-text/screen.png'   # 色號 → RGB 由執行器走完屬性暫存器
COLS, ROWS = 32, 102                                       # 視窗：32 位元組欄（256 像素）× 102 列


def write_png(path, w, h, px, pal):
    raw = b''.join(b'\x00' + bytes(px[y*w:(y+1)*w]) for y in range(h))
    def chunk(t, d):
        c = t + d
        return struct.pack('>I', len(d)) + c + struct.pack('>I', zlib.crc32(c))
    body = b'\x89PNG\r\n\x1a\n'
    body += chunk(b'IHDR', struct.pack('>IIBBBBB', w, h, 8, 3, 0, 0, 0))
    body += chunk(b'PLTE', bytes(pal))
    body += chunk(b'IDAT', zlib.compress(raw, 9))
    body += chunk(b'IEND', b'')
    with open(path, 'wb') as fp:
        fp.write(body)
    return body


def read_palette(path):
    png = open(path, 'rb').read()
    i = 8
    while i < len(png):
        n = struct.unpack_from('>I', png, i)[0]
        if png[i+4:i+8] == b'PLTE':
            return (list(png[i+8:i+8+n]) + [0]*48)[:48]
        i += 12 + n
    raise SystemExit('收據 PNG 沒有 PLTE：' + path)


def directory(d):
    """u16 目錄長度（10 × 筆數）；每筆 char[4] 名、u16 長度、u32 位移（從 +2 算）。"""
    size = struct.unpack_from('<H', d, 0)[0]
    ents = []
    for p in range(2, 2 + size, 10):
        name = d[p:p+4]
        length, off = struct.unpack_from('<HI', d, p+4)
        ents.append((name, length, off))
    return ents


def frames(seg):
    """段 → [(格種, 旗標byte, 下一格相對距離, VRAM位移, 欄數, 高, 資料起點)]。"""
    count = struct.unpack_from('<H', seg, 0)[0]
    p, out = 2, []
    for _ in range(count):
        kind, flag = seg[p], seg[p+1]
        nxt, off, w, h = struct.unpack_from('<HHHH', seg, p+2)
        out.append((kind, flag, nxt, off, w & 0xFF, h & 0xFF, p+10, p+4+nxt))
        p += 4 + nxt
    return out, p


class Window:
    """四個 plane，各 COLS × ROWS 位元組（每位元組 8 像素）。"""
    def __init__(self):
        self.p = [bytearray(COLS*ROWS) for _ in range(4)]

    def copy(self):
        w = Window()
        w.p = [bytearray(x) for x in self.p]
        return w

    def put(self, plane, di, value, xor):
        row, col = divmod(di, 80)
        if not (0 <= row < ROWS and 0 <= col < COLS):
            raise ValueError('寫到視窗外：列 %d 欄 %d' % (row, col))
        i = row*COLS + col
        if xor:
            self.p[plane][i] ^= value
        else:
            self.p[plane][i] = value

    def pixels(self):
        W = COLS*8
        px = bytearray(W*ROWS)
        for i in range(COLS*ROWS):
            y, xb = divmod(i, COLS)
            b0, b1, b2, b3 = (self.p[k][i] for k in range(4))
            for bit in range(8):
                m = 0x80 >> bit
                px[y*W + xb*8 + bit] = ((b0 & m) and 1) | ((b1 & m) and 2) | ((b2 & m) and 4) | ((b3 & m) and 8)
        return px


def blit(win, seg, fr):
    """照 348A:02F9（格種 0）／348A:03A3（格種 1）畫一格；回傳吃到哪個位移。"""
    kind, _flag, _nxt, off, w, h, q, _end = fr
    base = off
    plane, col = 0, 0
    di = base
    if kind == 0:
        cnt = 0
        while plane < 4:
            a, c = seg[q], seg[q+1]
            q += 2
            n = 1
            if a == c:                       # 重複組：a a c n → (a, c) × n
                c, n = seg[q], seg[q+1] or 0x10000
                q += 2
            for _ in range(n):
                win.put(plane, di, a, False)
                win.put(plane, di+80, c, False)
                di += 160
                cnt += 1
                if cnt == h:                 # 一直行是 h 組（2h 列）
                    cnt = 0
                    col += 1
                    di = base + col
                    if col == w:
                        col, plane, di = 0, plane + 1, base
                        if plane == 4:
                            break
        return q
    row = 0
    while plane < 4:
        a = seg[q]
        q += 1
        if a:
            win.put(plane, di, a, True)      # and es:[di],al 在 read mode 1／XOR 下 ＝ 畫面 ^= a
            di += 80
            row += 1
            if row == h:
                row, col = 0, col + 1
                di = base + col
                if col == w:
                    col, plane, di = 0, plane + 1, base
        else:
            n = seg[q]                       # 0 n：跳過 n 列，可以跨直行、跨 plane
            q += 1
            while True:
                if n >= h - row:
                    n -= h - row
                    row, col = 0, col + 1
                    di = base + col
                    if col == w:
                        col, plane, di = 0, plane + 1, base
                        if plane == 4:
                            break
                    if n == 0:
                        break
                else:
                    row += n
                    di += 80*n
                    break
    return q


def main():
    d = open(SRC, 'rb').read()
    pal = read_palette(PAL_PNG)
    ents = directory(d)
    segs = {name: d[off+2:off+2+length] for name, length, off in ents}
    os.makedirs(OUT, exist_ok=True)

    # 先把每段第 0 格是覆寫格的 D0 解出來，給 XOR 開頭的段當底圖
    d0 = {}
    for name, seg in segs.items():
        if name[2:] == b'D0':
            fr, _ = frames(seg)
            if fr[0][0] == 0:
                w = Window()
                blit(w, seg, fr[0])
                d0[name[:2]] = w

    meta, exact_all = [], True
    for name, length, off in ents:
        seg = segs[name]
        fr, end = frames(seg)
        tag = name[:2].hex().upper() + name[2:].decode('ascii')   # 第二個字元不全是 ASCII
        folder = os.path.join(OUT, tag)
        os.makedirs(folder, exist_ok=True)
        if fr[0][0] == 0:
            win, base = Window(), 'none（第 0 格是覆寫格）'
        elif name[:2] in d0:
            win = d0[name[:2]].copy()
            base = name[:2].hex().upper() + 'D0 第 0 格' + \
                   ('（收據證實）' if name == b'PPAB' else '（推論）')
        else:
            win, base = Window(), 'black（同單位沒有 D0，XOR 疊在黑底上）'
        start = win.copy()
        rows = []
        for k, f in enumerate(fr):
            q = blit(win, seg, f)
            exact = q == f[7]
            exact_all &= exact
            px = win.pixels()
            write_png(os.path.join(folder, 'f%02d.png' % k), COLS*8, ROWS, px, pal)
            rows.append({'frame': k, 'kind': f[0], 'flag': f[1], 'vram_off': f[3],
                         'row': f[3] // 80, 'col_byte': f[3] % 80, 'w_bytes': f[4],
                         'h': f[5], 'h_rows': f[5]*2 if f[0] == 0 else f[5],
                         'data_bytes': f[7] - f[6], 'exact': exact,
                         'sha256_window_pixels': hashlib.sha256(px).hexdigest()})
        meta.append({'name_hex': name.hex(), 'tag': tag, 'offset_plus2': off + 2,
                     'length': length, 'frames': len(fr), 'segment_exact': end == length,
                     'base': base,
                     'returns_to_start': win.p == start.p,
                     'kinds': ''.join(str(f[0]) for f in fr), 'frame_list': rows})

    json.dump({'source': 'COMANI.BMG', 'sha256': hashlib.sha256(d).hexdigest(),
               'segments': len(ents), 'window': {'w': COLS*8, 'h': ROWS},
               'palette_from': 'workplace/oracle/eng-text/screen.png 的 PLTE（收據）',
               'hash_note': 'sha256_window_pixels 是 256×102 色號位元組的雜湊，與 PNG 編碼器無關',
               'items': meta},
              open(INDEX, 'w'), ensure_ascii=False, indent=1)
    print('段數', len(ents), '格數', sum(m['frames'] for m in meta), '→', OUT, '；索引', INDEX)
    print('每格都剛好吃到下一格:', exact_all, '；每段都剛好吃完:', all(m['segment_exact'] for m in meta))
    print('畫完回到起點的段:', sum(m['returns_to_start'] for m in meta), '/', len(meta))
    print('XOR 開頭但沒有 D0 底圖:', [m['tag'] for m in meta if m['base'].startswith('black')])
    for t in ('5050AB', '5050D0', '46C0D0'):
        m = next(x for x in meta if x['tag'] == t)
        print(t, [r['sha256_window_pixels'][:16] for r in m['frame_list']])


if __name__ == '__main__':
    main()
