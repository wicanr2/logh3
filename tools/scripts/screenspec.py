"""量測畫面版面的共用工具：外框、文字格點、字碼反查。

不目測、不猜：每一個數字都從色號陣列算出來，字碼是拿點陣回 GRAPH.IMG
反查的。
"""
import collections, struct

W, H = 640, 480
GLYPH_W, GLYPH_H, HALF_W = 16, 15, 8
FULL_STRIDE, HALF_STRIDE = GLYPH_H * 2, GLYPH_H
SYMBOL_BASE, HANZI_BASE, HANZI_INDEX, HALF_BASE = 3840, 27030, 471, 0


def load_screen(path):
    d = open(path, 'rb').read()
    return [d[y*W:(y+1)*W] for y in range(H)], d


def font_lookup(path):
    """回 (全形點陣→Big5 碼, 半形點陣→ASCII 碼) 兩張反查表。"""
    f = open(path, 'rb').read()

    def off(code):
        lead, trail = code >> 8, code & 0xFF
        t = trail - 0x40 if trail <= 0x7E else trail - 0xA1 + 63
        i = (lead - 0xA1) * 157 + t
        return SYMBOL_BASE + i*FULL_STRIDE if i < HANZI_INDEX else HANZI_BASE + (i-HANZI_INDEX)*FULL_STRIDE

    full = {}
    for lead in range(0xA1, 0xFA):
        for trail in list(range(0x40, 0x7F)) + list(range(0xA1, 0xFF)):
            code = lead << 8 | trail
            o = off(code)
            if 0 <= o and o + FULL_STRIDE <= len(f):
                full.setdefault(f[o:o+FULL_STRIDE], code)
    half = {}
    for code in range(0x20, 0x80):
        o = HALF_BASE + code*HALF_STRIDE
        if o + HALF_STRIDE <= len(f):
            half.setdefault(f[o:o+HALF_STRIDE], code)
    return full, half


def measure_frame(rows, x0, y0, x1, y1, fill=9, layers_max=16, sample=None):
    """一層層往內剝外框。每一邊取兩個相鄰樣本，認得出 1:1 交錯的網點。

    sample ＝ (掃上下緣用的 x, 掃左右緣用的 y)。**視窗被別的東西蓋住時
    一定要指定**——預設取中線，而中線很可能正落在疊上去的視窗或地圖上，
    量出來的層數會一路撞到上限。
    """
    mid_x, mid_y = sample if sample else ((x0+x1)//2, (y0+y1)//2)
    # 層數不能超過短邊的一半——剝過頭會讀到對面那一側的像素，
    # 而且不會報錯：層數一路長到上限，畫出來整塊都不對。
    layers_max = min(layers_max, min(x1-x0+1, y1-y0+1) // 2)
    layers, alts = [], []
    for i in range(layers_max):
        side = [
            (rows[y0+i][mid_x], rows[y0+i][mid_x+1]),
            (rows[mid_y][x1-i], rows[mid_y+1][x1-i]),
            (rows[y1-i][mid_x], rows[y1-i][mid_x+1]),
            (rows[mid_y][x0+i], rows[mid_y+1][x0+i]),
        ]
        if i > 0 and all(a == fill and b == fill for a, b in side):
            break
        layers.append([a for a, _ in side])
        alts.append([0 if a == b else b for a, b in side])
    coords_of = lambda i: [(mid_x, y0+i), (x1-i, mid_y), (mid_x, y1-i), (x0+i, mid_y)]
    for i, (lay, alt) in enumerate(zip(layers, alts)):
        for k, (cx, cy) in enumerate(coords_of(i)):
            if alt[k] and (cx + cy) % 2:
                lay[k], alt[k] = alt[k], lay[k]
    return layers, alts


def frame_patches(rows, x0, y0, x1, y1, layers, alts, fill=9, corner=8, exclude=()):
    """照 layers 模型自己畫一遍，把四個角對不上的像素列出來。

    exclude 是一串「不要比」的矩形（畫面座標）：被別的東西蓋住的角落
    比出來的差異是那個東西的內容，不是框的修正。
    """
    w, h = x1-x0+1, y1-y0+1
    canvas = [[fill]*w for _ in range(h)]

    def put(cx, cy, c, alt):
        canvas[cy][cx] = c if not alt or (cx+x0 + cy+y0) % 2 == 0 else alt

    for i, (lay, alt) in enumerate(zip(layers, alts)):
        lw, lh = w - 2*i, h - 2*i
        if lw <= 0 or lh <= 0:
            break
        for k in range(lh):
            put(i, i+k, lay[3], alt[3])
        for k in range(lw):
            put(i+k, i+lh-1, lay[2], alt[2])
        for k in range(lw):
            put(i+k, i, lay[0], alt[0])
        for k in range(lh):
            put(i+lw-1, i+k, lay[1], alt[1])
    out = []
    for dy in range(h):
        for dx in range(w):
            gx, gy = x0+dx, y0+dy
            if any(ex <= gx < ex+ew and ey <= gy < ey+eh for ex, ey, ew, eh in exclude):
                continue
            if (dx < corner or dx >= w-corner) and (dy < corner or dy >= h-corner) \
                    and canvas[dy][dx] != rows[gy][gx]:
                out.append({'dx': dx if dx < corner*2 else dx-w,
                            'dy': dy if dy < corner*2 else dy-h,
                            'color': rows[y0+dy][x0+dx]})
    return out


def cell_bits(rows, cx, cy, ink, width):
    """抽一格的點陣（只取 ink 色），回字型檔裡那種位元組。"""
    out = bytearray()
    for y in range(GLYPH_H):
        v = 0
        for x in range(width):
            px, py = cx + x, cy + y
            if 0 <= px < W and 0 <= py < H and rows[py][px] == ink:
                v |= (0x8000 if width == 16 else 0x80) >> x
        out += bytes([v >> 8, v & 0xFF]) if width == 16 else bytes([v])
    return bytes(out)


def read_line(rows, x, y, x_end, ink, full, half, exclude=()):
    """從 (x,y) 往右讀一行，全形與半形混排都認。

    先試全形（16 寬），不中再試半形（8 寬）——反過來的話全形字的左半邊
    常常剛好是某個半形字，會讀出一堆看似合理的垃圾。
    """
    cells = []
    while x + HALF_W <= x_end:
        # 小框的框線點陣剛好等於某些製表符，會被讀成「文字」，
        # 而那一行的陰影會把框的邊緣蓋掉一個像素。跳過框佔的格子。
        # 判斷用「這一格有沒有碰到框」，不是「格子左上角在不在框裡」——
        # 起點在框左邊的那一格照樣會蓋到框。
        if any(ex < x + GLYPH_W and x < ex + ew and ey < y + GLYPH_H and y < ey + eh
               for ex, ey, ew, eh in exclude):
            x += HALF_W
            continue
        f = cell_bits(rows, x, y, ink, GLYPH_W)
        if any(f) and f in full:
            cells.append({'x': x, 'w': GLYPH_W, 'code': full[f], 'hex': f'{full[f]:04X}'})
            x += GLYPH_W
            continue
        h = cell_bits(rows, x, y, ink, HALF_W)
        if any(h) and h in half:
            cells.append({'x': x, 'w': HALF_W, 'code': half[h], 'hex': f'{half[h]:02X}',
                          'ascii': chr(half[h])})
            x += HALF_W
            continue
        x += HALF_W
    return cells


def ink_rows(rows, x0, x1, y0, y1, ink, least=2, gap=3, min_h=8):
    """哪些列有文字。"""
    ys = [y for y in range(y0, y1) if sum(1 for x in range(x0, x1) if rows[y][x] == ink) > least]
    out, s, p = [], None, None
    for y in ys:
        if s is None:
            s = y
        elif y - p > gap:
            out.append((s, p))
            s = y
        p = y
    if s is not None:
        out.append((s, p))
    return [g for g in out if g[1] - g[0] >= min_h]
