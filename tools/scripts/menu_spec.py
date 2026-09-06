"""從 oracle 收據量出主選單的精確版面，輸出成 remake 要對齊的規格。

不目測、不猜：每一個數字都從色號陣列算出來。
"""
import json, collections, sys
sys.path.insert(0, '/work/tools/scripts')

W, H = 640, 480
idx = open('/work/workplace/re06/screen.bin', 'rb').read()
pal = open('/work/workplace/re06/pal.bin', 'rb').read()
rows = [idx[y*W:(y+1)*W] for y in range(H)]

cnt = collections.Counter(idx)
bg = cnt.most_common(1)[0][0]
print("色號分佈:", cnt.most_common())

ys = [y for y in range(H) if any(b != bg for b in rows[y])]
xs = sorted({x for y in ys for x in range(W) if rows[y][x] != bg})
x0, x1, y0, y1 = xs[0], xs[-1], ys[0], ys[-1]
print(f"視窗 bbox = ({x0},{y0})-({x1},{y1})  {x1-x0+1}×{y1-y0+1}")

# 逐列印出視窗上緣的色號，看外框是怎麼疊的
print("\n上緣前 8 列（相對 x0，只印前 12 個像素）:")
for y in range(y0, y0+8):
    print(f"  y={y-y0:2d}: " + ' '.join(f"{rows[y][x]:2d}" for x in range(x0, x0+12)))
print("下緣後 6 列:")
for y in range(y1-5, y1+1):
    print(f"  y={y-y0:2d}: " + ' '.join(f"{rows[y][x]:2d}" for x in range(x0, x0+12)))
print("\n左緣前 8 行（相對 y0，取 y0+20 那一列）:")
yy = y0 + 20
print("  " + ' '.join(f"{rows[yy][x]:2d}" for x in range(x0, x0+12)))
print("右緣後 8 行:")
print("  " + ' '.join(f"{rows[yy][x]:2d}" for x in range(x1-11, x1+1)))

# 內部填色區
inner = None
for dy in range(0, 20):
    r = rows[y0+dy]
    run = [x for x in range(x0, x1+1) if r[x] == 9]
    if len(run) > 40:
        inner = dy
        break
print(f"\n第一列出現大片色號 9 的相對 y = {inner}")

spec = dict(
    canvas=dict(width=W, height=H, mode='VGA 12h 640x480x16'),
    background=bg,
    window=dict(x=x0, y=y0, w=x1-x0+1, h=y1-y0+1),
    text=dict(color=7, glyph_w=16, glyph_h=15, col_pitch=16, row_pitch=20,
              cols_x=[288, 304, 320], rows_y=[202, 222, 242, 262]),
    items=[['帝', '　', '國'], ['同', '　', '盟'], ['對', '　', '戰'], ['取', '進', '度']],
    palette_6bit=[[pal[i*3]>>2, pal[i*3+1]>>2, pal[i*3+2]>>2] for i in range(16)],
    source='workplace/re06/screen.bin + pal.bin（dosgolem oracle）')
with open('/work/docs/internals/data/screen-main-menu.json', 'w') as f:
    json.dump(spec, f, ensure_ascii=False, indent=2)
print("\n寫出 docs/internals/data/screen-main-menu.json")
