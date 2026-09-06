"""從 oracle 收據量出防拷問答畫面的版面，輸出 remake 要對齊的規格。

不目測、不猜：框層、格點、字碼全部從色號陣列算出來；字碼是拿畫面上的
點陣回 GRAPH.IMG 反查的，不是照著螢幕認字。
"""
import collections, json

W, H = 640, 480
INK, SHADOW = 7, 3
GLYPH_W, GLYPH_H = 16, 15
STRIDE = GLYPH_H * 2
SYMBOL_BASE, HANZI_BASE, HANZI_INDEX = 3840, 27030, 471

d = open('/work/workplace/oracle/cb1/screen.bin', 'rb').read()
pal = open('/work/workplace/oracle/cb1/pal.bin', 'rb').read()
rows = [d[y*W:(y+1)*W] for y in range(H)]
bg = collections.Counter(d).most_common(1)[0][0]

ys = [y for y in range(H) if any(v != bg for v in rows[y])]
xs = sorted({x for y in ys for x in range(W) if rows[y][x] != bg})
x0, x1, y0, y1 = xs[0], xs[-1], ys[0], ys[-1]

# 外框：一層一層往內剝，直到四邊都變成填色。
# **每一邊都取兩個相鄰樣本**：原版有的邊是 1:1 交錯的網點，
# 只看一個點會得到「一半的像素對」而框看起來完全正常。
mid_x, mid_y = (x0+x1)//2, (y0+y1)//2
fill = rows[mid_y][mid_x]
layers, alts = [], []
for i in range(16):
    side = [
        (rows[y0+i][mid_x], rows[y0+i][mid_x+1]),   # 上
        (rows[mid_y][x1-i], rows[mid_y+1][x1-i]),   # 右
        (rows[y1-i][mid_x], rows[y1-i][mid_x+1]),   # 下
        (rows[mid_y][x0+i], rows[mid_y+1][x0+i]),   # 左
    ]
    if i > 0 and all(a == fill and b == fill for a, b in side):
        break
    layers.append([a for a, _ in side])
    alts.append([0 if a == b else b for a, b in side])
# 交錯色要對齊 (x+y) 的奇偶：主色放在 (x+y) 為偶數的位置
for i, (lay, alt) in enumerate(zip(layers, alts)):
    coords = [(mid_x, y0+i), (x1-i, mid_y), (mid_x, y1-i), (x0+i, mid_y)]
    for k in range(4):
        if alt[k] and (coords[k][0] + coords[k][1]) % 2:
            lay[k], alt[k] = alt[k], lay[k]

# 文字：色號 7 的像素在哪些列／行成群
def runs(v, gap=1):
    out, s, p = [], None, None
    for i in v:
        if s is None:
            s = i
        elif i - p > gap:
            out.append((s, p)); s = i
        p = i
    if s is not None:
        out.append((s, p))
    return out

tx0, tx1 = 230, x1
ink_rows = [y for y in range(y0, y1) if sum(1 for x in range(tx0, tx1) if rows[y][x] == INK) > 2]
row_groups = [g for g in runs(ink_rows, 3) if g[1] - g[0] >= 8]
# 輸入框的邊也是色號 7，會被認成一行「文字」——那三個「製表符」是
# 框線的形狀，不是字。用**內部填色 8 的連續區段**把它認出來並排除。
def long_runs(y, colour, least):
    out, n, s0 = [], 0, None
    for x in range(x0, x1+1):
        if rows[y][x] == colour:
            if s0 is None:
                s0 = x
            n += 1
        else:
            if s0 is not None and n >= least:
                out.append((s0, x-1))
            s0, n = None, 0
    return out

box = None
cand = {y: long_runs(y, 8, 20) for y in range(y0+40, y1-6)}
cand = {y: r for y, r in cand.items() if r}
if cand:
    by0, by1 = min(cand), max(cand)
    bx0 = min(r[0][0] for r in cand.values())
    bx1 = max(r[-1][1] for r in cand.values())
    box = {'x': bx0-2, 'y': by0-2, 'w': bx1-bx0+1+4, 'h': by1-by0+1+4,
           'layers': [[rows[by0-2+i][bx0+2], rows[by0+2][bx1+2-i],
                       rows[by1+2-i][bx0+2], rows[by0+2][bx0-2+i]] for i in range(2)],
           'fill': 8, 'order': 'TRBL'}
    row_groups = [g for g in row_groups if g[0] < by0-2]
ink_cols = [x for x in range(tx0, tx1) if any(rows[y][x] == INK for y in range(y0, y1))]
col_groups = runs(ink_cols, 3)

# 字型：建一份「點陣 → Big5 碼」的反查表
font = open('/orig/GRAPH.IMG', 'rb').read()
def offset_of(code):
    lead, trail = code >> 8, code & 0xFF
    t = trail - 0x40 if trail <= 0x7E else trail - 0xA1 + 63
    i = (lead - 0xA1) * 157 + t
    return SYMBOL_BASE + i*STRIDE if i < HANZI_INDEX else HANZI_BASE + (i-HANZI_INDEX)*STRIDE
lookup = {}
for lead in range(0xA1, 0xFA):
    for trail in list(range(0x40, 0x7F)) + list(range(0xA1, 0xFF)):
        code = lead << 8 | trail
        off = offset_of(code)
        if 0 <= off and off + STRIDE <= len(font):
            lookup.setdefault(font[off:off+STRIDE], code)

def cell_bits(cx, cy):
    out = bytearray()
    for y in range(GLYPH_H):
        v = 0
        for x in range(GLYPH_W):
            px, py = cx + x, cy + y
            if 0 <= px < W and 0 <= py < H and rows[py][px] == INK:
                v |= 0x8000 >> x
        out += bytes([v >> 8, v & 0xFF])
    return bytes(out)

# 格點的原點要搜出來：色號成群的邊界不等於字框的頂端／左緣。
# **判準是反查命中數**，不是看起來對齊。
def try_origin(ry0, dx, dy):
    cells, x = [], col_groups[0][0] + dx
    while x + GLYPH_W <= tx1:
        bits = cell_bits(x, ry0 + dy)
        if any(bits):
            cells.append((x, lookup.get(bits)))
        x += GLYPH_W
    hit = sum(1 for _, c in cells if c)
    return hit, cells

lines = []
for ry0, ry1 in row_groups:
    best = (-1, None, 0, 0)
    for dy in range(-6, 7):
        for dx in range(-8, 9):
            hit, cells = try_origin(ry0, dx, dy)
            if hit > best[0]:
                best = (hit, cells, dx, dy)
    hit, cells, dx, dy = best
    lines.append({'y': ry0 + dy, 'dx': dx, 'dy': dy, 'hits': hit,
                  'cells': [{'x': x, 'code': c, 'hex': f'{c:04X}' if c else None}
                            for x, c in cells]})
    print(f'列 {ry0}: dx={dx} dy={dy} 命中 {hit}/{len(cells)}')

# 角落修正：照 layers 模型自己畫一遍，跟 oracle 比，把對不上的像素列出來。
# **角落規則還沒逆向出來**（`docs/internals/07`），兩個視窗尺寸的樣本仍不足以
# 歸納；這些像素是量出來的證據，不是規則。
canvas = [[9]*(x1-x0+1) for _ in range(y1-y0+1)]
def putv(cx, cy, n, c, alt):
    for i in range(n):
        canvas[cy+i][cx] = c if (cx+x0 + cy+i+y0) % 2 == 0 or not alt else alt
def puth(cx, cy, n, c, alt):
    for i in range(n):
        canvas[cy][cx+i] = c if (cx+i+x0 + cy+y0) % 2 == 0 or not alt else alt
for i, (lay, alt) in enumerate(zip(layers, alts)):
    lw, lh = (x1-x0+1) - 2*i, (y1-y0+1) - 2*i
    if lw <= 0 or lh <= 0:
        break
    putv(i, i, lh, lay[3], alt[3])              # 左
    puth(i, i+lh-1, lw, lay[2], alt[2])         # 下
    puth(i, i, lw, lay[0], alt[0])              # 上
    putv(i+lw-1, i, lh, lay[1], alt[1])         # 右
patches = []
for dy in range(y1-y0+1):
    for dx in range(x1-x0+1):
        want = rows[y0+dy][x0+dx]
        w_, h_ = x1-x0+1, y1-y0+1
        corner = (dx < 8 or dx >= w_-8) and (dy < 8 or dy >= h_-8)
        if corner and canvas[dy][dx] != want:
            patches.append({'dx': dx if dx < 20 else dx-(x1-x0+1),
                            'dy': dy if dy < 20 else dy-(y1-y0+1),
                            'color': want})

spec = {
    'canvas': {'width': W, 'height': H, 'mode': 'VGA 12h 640x480x16'},
    'background': bg,
    'window': {'x': x0, 'y': y0, 'w': x1-x0+1, 'h': y1-y0+1},
    'frame': {'layers': layers, 'alt': alts, 'fill': fill, 'patches': patches},
    'face': {'source': 'FACE.AMG', 'index': 0, 'x': 144, 'y': 192},
    'input_box': box,
    'palette_6bit': [[pal[i*3], pal[i*3+1], pal[i*3+2]] for i in range(16)],
    'text': {'ink': INK, 'shadow': SHADOW, 'glyph_w': GLYPH_W, 'glyph_h': GLYPH_H,
             'row_groups': row_groups, 'col_start': col_groups[0][0] if col_groups else None,
             'lines': lines},
}

json.dump(spec, open('/work/docs/internals/data/screen-copy-protection.json', 'w'),
          ensure_ascii=False, indent=1)
