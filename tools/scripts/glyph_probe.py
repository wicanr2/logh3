"""從 dosgolem 的 framebuffer 量出中文字形的尺寸與點陣。

畫面是 640×480 的色號陣列（每 byte 一個像素，值 0–15）。主選單的文字視窗
是唯一有內容的區域，字是亮色、底是暗色。這支腳本只做一件事：把文字區域
的色號分佈與逐列輪廓印出來，讓尺寸由資料決定，不由猜測決定。
"""
import sys, collections

W, H = 640, 480
idx = open(sys.argv[1], 'rb').read()
assert len(idx) == W * H, len(idx)

rows = [idx[y*W:(y+1)*W] for y in range(H)]
bg = collections.Counter(idx).most_common(1)[0][0]
print(f"背景色號 = {bg}")

ys = [y for y in range(H) if any(b != bg for b in rows[y])]
xs = [x for y in ys for x in range(W) if rows[y][x] != bg]
y0, y1, x0, x1 = min(ys), max(ys), min(xs), max(xs)
print(f"內容 bbox = ({x0},{y0})-({x1},{y1})  {x1-x0+1}×{y1-y0+1}")

sub = [rows[y][x0:x1+1] for y in range(y0, y1+1)]
print("區域色號分佈:", collections.Counter(b for r in sub for b in r).most_common())

# 文字色 = 區域內最亮/最少見的那個（先全部列出來，由人判讀）
for c in sorted({b for r in sub for b in r}):
    cnt = sum(r.count(c) for r in sub)
    print(f"  色號 {c:2d}: {cnt} 像素")

txt = int(sys.argv[2]) if len(sys.argv) > 2 else None
if txt is None:
    sys.exit(0)
if len(sys.argv) > 6:
    x0, y0, x1, y1 = (int(v) for v in sys.argv[3:7])
    sub = [rows[y][x0:x1+1] for y in range(y0, y1+1)]
    print(f"改用指定範圍 ({x0},{y0})-({x1},{y1})")

# 只留文字色，印出逐列 / 逐行的投影，字形邊界會自己浮出來
print(f"\n只看色號 {txt} 的投影：")
colproj = [sum(1 for r in sub if r[x] == txt) for x in range(len(sub[0]))]
rowproj = [sum(1 for b in r if b == txt) for r in sub]

def runs(proj):
    out, start = [], None
    for i, v in enumerate(proj):
        if v and start is None: start = i
        elif not v and start is not None:
            out.append((start, i-1)); start = None
    if start is not None: out.append((start, len(proj)-1))
    return out

cr, rr = runs(colproj), runs(rowproj)
print("  有字的行段（相對 bbox）:", rr, "→ 高度", [b-a+1 for a,b in rr])
print("  有字的列段（相對 bbox）:", cr, "→ 寬度", [b-a+1 for a,b in cr])

print("\n第一個字的點陣（. = 底、# = 字）:")
if rr and cr:
    ra, rb = rr[0]; ca, cb = cr[0]
    for y in range(ra, rb+1):
        print('   ' + ''.join('#' if sub[y][x] == txt else '.' for x in range(ca, cb+1)))
    print(f"   → {cb-ca+1} × {rb-ra+1}")

# 用法二：glyph_probe.py <bin> <文字色號> <x0> <y0> <x1> <y1>
