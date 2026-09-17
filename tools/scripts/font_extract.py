"""解 GRAPH.IMG 的中文字型：釘住格式與索引，輸出 JSON 與 PNG。

⚠ 會整份覆寫 docs/internals/data/font-graph-img.json：那份後來人工補過欄位，這支還是舊格式，
重跑之前先處理（2026-09-17 踩過，git checkout 還原）。

驗證用畫面當 oracle（九個已知字逐位元吻合），再用倚天 stdfont.15 做外部
對照，確認遊戲是不是直接內嵌了倚天字型。
"""
import json, os, sys
sys.path.insert(0, '/work/tools/scripts')
import png

W, H = 640, 480
ROWS, WIDTH, STRIDE = 15, 16, 30

img = open('/orig/GRAPH.IMG', 'rb').read()
scr = open('/work/workplace/re06/screen.bin', 'rb').read()
rows = [scr[y*W:(y+1)*W] for y in range(H)]
INK = 7
COL_X, ROW_Y = [288, 304, 320], [202, 222, 242, 262]

def big5_index(code):
    """Big5 字碼 → 序號（0xA140 起算）。"""
    lead, trail = code >> 8, code & 0xFF
    t = trail - 0x40 if trail <= 0x7E else trail - 0xA1 + 63
    return (lead - 0xA1) * 157 + t

def glyph_at(off):
    return [img[off + 2*r] << 8 | img[off + 2*r + 1] for r in range(ROWS)]

def screen_glyph(col, row):
    out = []
    for y in range(ROWS):
        v = 0
        for x in range(WIDTH):
            if rows[ROW_Y[row] + y][COL_X[col] + x] == INK:
                v |= 1 << (WIDTH - 1 - x)
        out.append(v)
    return out

CHARS = [(0,0,'帝'), (0,2,'國'), (1,0,'同'), (1,2,'盟'),
         (2,0,'對'), (2,2,'戰'), (3,0,'取'), (3,1,'進'), (3,2,'度')]

# 先用一個字定 base，再用其餘八個驗證
first = CHARS[2]                                   # 同
code0 = int.from_bytes(first[2].encode('big5'), 'big')
want0 = screen_glyph(first[1], first[0])
pat0 = b''.join(bytes([v >> 8, v & 0xFF]) for v in want0)
off0 = img.find(pat0)
assert off0 != -1, '在 GRAPH.IMG 找不到基準字'
BASE = off0 // STRIDE - big5_index(code0)
print(f"基準：{first[2]} Big5 {code0:04X} 偏移 0x{off0:06X}"
      f" → 字碼序號前面還有 {BASE} 個字形\n")

def offset_of(code):
    return (big5_index(code) + BASE) * STRIDE

ok, table = 0, []
for row, col, ch in CHARS:
    code = int.from_bytes(ch.encode('big5'), 'big')
    o = offset_of(code)
    match = glyph_at(o) == screen_glyph(col, row)
    ok += match
    print(f"  {ch} {code:04X} → 0x{o:06X}  {'吻合' if match else '**不吻合**'}")
    table.append(dict(char=ch, big5=f"{code:04X}", offset=o, verified=match))
print(f"\n{ok}/{len(CHARS)} 逐位元吻合")

total = len(img) // STRIDE
print(f"GRAPH.IMG {len(img)} bytes ÷ {STRIDE} = {total} 個字形"
      f"（{'整除' if len(img) % STRIDE == 0 else '有餘數 %d' % (len(img) % STRIDE)}）")

# 與倚天 stdfont.15 外部對照
ext = None
if os.path.exists('/ref/stdfont.15'):
    std = open('/ref/stdfont.15', 'rb').read()
    # stdfont 從 Big5 A440 起；GRAPH.IMG 的對應偏移
    a440 = (big5_index(0xA440) + BASE) * STRIDE
    n = min(len(std), len(img) - a440)
    same = sum(1 for i in range(0, n, STRIDE) if std[i:i+STRIDE] == img[a440+i:a440+i+STRIDE])
    ext = dict(reference='倚天 stdfont.15', compared=n // STRIDE, identical=same)
    print(f"\n與倚天 stdfont.15 對照：{n//STRIDE} 個字形中 {same} 個逐位元相同"
          f"（{same*100.0/(n//STRIDE):.2f}%）")

# 前 BASE 個字形是什麼？印出前幾個看看
print(f"\n字碼序號 0 之前的 {BASE} 個字形，取樣（. 空 # 有）：")
for g in range(0, min(BASE, 3)):
    bits = glyph_at(g * STRIDE)
    for r in bits[:15]:
        print('   ' + ''.join('#' if r >> (15 - x) & 1 else '.' for x in range(16)))
    print('   ' + '-' * 16)

os.makedirs('/work/workplace/assets/font', exist_ok=True)
os.makedirs('/work/docs/internals/data', exist_ok=True)

# PNG：整份字型做成圖集，每列 96 個字
per_row = 96
nrow = (total + per_row - 1) // per_row
sheet = [[0] * (per_row * WIDTH) for _ in range(nrow * ROWS)]
for g in range(total):
    gx, gy = (g % per_row) * WIDTH, (g // per_row) * ROWS
    bits = glyph_at(g * STRIDE)
    for y in range(ROWS):
        for x in range(WIDTH):
            sheet[gy + y][gx + x] = (bits[y] >> (WIDTH - 1 - x)) & 1
png.write_mono('/work/workplace/assets/font/graph-img-sheet.png',
               per_row * WIDTH, nrow * ROWS, sheet)
print(f"\n寫出 workplace/assets/font/graph-img-sheet.png"
      f"（{per_row*WIDTH}×{nrow*ROWS}，{total} 個字形）")

meta = dict(
    file='GRAPH.IMG', bytes=len(img), glyphs=total,
    format=dict(width=WIDTH, rows=ROWS, stride=STRIDE,
                bit_order='MSB 在左', row_bytes='每列 2 bytes，高位在前'),
    index=dict(
        formula='offset = (big5_index(code) + base) * 30',
        big5_index='(lead-0xA1)*157 + (trail<=0x7E ? trail-0x40 : trail-0xA1+63)',
        base=BASE,
        note=f'字碼序號 0 對應 Big5 0xA140；它前面還有 {BASE} 個字形，內容未確認'),
    verified_against='workplace/re06/screen.bin（主選單，dosgolem oracle）',
    verified=table,
    external_reference=ext)
with open('/work/docs/internals/data/font-graph-img.json', 'w') as f:
    json.dump(meta, f, ensure_ascii=False, indent=2)
print("寫出 docs/internals/data/font-graph-img.json")
