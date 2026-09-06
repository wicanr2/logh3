"""確認遊戲用的是哪一套字型、以及字碼怎麼對到字形索引。

oracle 是畫面：主選單的九個字我們知道字碼也知道長相。把每個字的點陣抽出來，
在候選字型檔裡找同一串 bytes，找到就同時得到格式、索引與字型來源。

列數與位元組序都不預設，全部試過一輪；「找不到」只有在所有組合都試過
之後才成立。
"""
import sys, itertools, json, os

W, H = 640, 480
d = open('/work/workplace/re06/screen.bin', 'rb').read()
rows = [d[y*W:(y+1)*W] for y in range(H)]
INK = 7
COL_X = [288, 304, 320]
ROW_Y = [202, 222, 242, 262]
CHARS = [
    (0, 0, '帝', 0xAB63), (0, 2, '國', 0xB0EA),
    (1, 0, '同', 0xA650), (1, 2, '盟', 0xB7F9),
    (2, 0, '對', 0xB9EF), (2, 2, '戰', 0xBED4),
    (3, 0, '取', 0xA8FA), (3, 1, '進', 0xB669), (3, 2, '度', 0xABD7),
]

def bits(cx, cy, dx, dy, n):
    out = []
    for y in range(n):
        v = 0
        for x in range(16):
            px, py = cx + dx + x, cy + dy + y
            if 0 <= px < W and 0 <= py < H and rows[py][px] == INK:
                v |= 1 << (15 - x)
        out.append(v)
    return out

def pack(vals, hi_first):
    b = bytearray()
    for v in vals:
        b += bytes([(v >> 8) & 0xFF, v & 0xFF] if hi_first else [v & 0xFF, (v >> 8) & 0xFF])
    return bytes(b)

targets = {}
for t in sys.argv[1:]:
    if os.path.exists(t):
        targets[t] = open(t, 'rb').read()
for t, b in targets.items():
    print(f"{t}: {len(b)} bytes")
print()

results = []
for r, c, ch, code in CHARS:
    cx, cy = COL_X[c], ROW_Y[r]
    best = None
    for nrows, dy, dx, hi in itertools.product((15, 16, 14), range(-4, 2), range(-4, 2), (True, False)):
        pat = pack(bits(cx, cy, dx, dy, nrows), hi)
        if pat.count(0) > nrows * 2 - 6:
            continue
        for t, b in targets.items():
            off = b.find(pat)
            if off != -1:
                best = dict(file=t, off=off, nrows=nrows, dx=dx, dy=dy,
                            hi_first=hi, count=b.count(pat))
                break
        if best:
            break
    if best:
        print(f"{ch} {code:04X}  {os.path.basename(best['file'])} "
              f"@0x{best['off']:06X}({best['off']})  {best['nrows']} 列 "
              f"dx={best['dx']} dy={best['dy']} "
              f"{'高位在前' if best['hi_first'] else '低位在前'} ×{best['count']}")
        best.update(char=ch, big5=code)
        results.append(best)
    else:
        print(f"{ch} {code:04X}  找不到")

def eten_index(code):
    """倚天／Big5 的標準字形序號。"""
    lead, trail = code >> 8, code & 0xFF
    t = trail - 0x40 if trail <= 0x7E else trail - 0xA1 + 63
    return (lead - 0xA1) * 157 + t

if len(results) >= 2:
    print("\n索引檢驗（倚天標準公式 idx = (lead-0xA1)*157 + trail 序）：")
    for r in results:
        i = eten_index(r['big5'])
        stride = r['nrows'] * 2
        print(f"  {r['char']} idx={i:5d}  idx*{stride}=0x{i*stride:06X}  "
              f"實際 0x{r['off']:06X}  差 {r['off'] - i*stride}")

if results:
    os.makedirs('/work/docs/internals/data', exist_ok=True)
    with open('/work/docs/internals/data/font-probe.json', 'w') as f:
        json.dump(dict(source='workplace/re06/screen.bin', ink_color=INK,
                       matches=results), f, ensure_ascii=False, indent=2)
    print("\n寫出 docs/internals/data/font-probe.json")
