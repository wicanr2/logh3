"""把視窗框擬合成「一層層矩形」的參數：每層四邊各一個色號。

角落會被兩邊搶，所以量邊的時候把角排除；角落另外單獨列出來，
由 diff 決定誰贏。
"""
import json, os, collections
W, H = 640, 480
idx = open('/work/workplace/re06/screen.bin', 'rb').read()
rows = [idx[y*W:(y+1)*W] for y in range(H)]
X0, Y0, WD, HT = (int(v) for v in os.environ.get('RECT', '275,181,76,116').split(','))

def uniq(v):
    c = collections.Counter(v)
    return c.most_common(1)[0][0], len(c), c.most_common()

print(" 層   上          右          下          左")
layers = []
for r in range(9):
    x1, y1 = X0 + WD - 1 - r, Y0 + HT - 1 - r
    xa, xb = X0 + r + 1, X0 + WD - r - 2      # 排除角
    ya, yb = Y0 + r + 1, Y0 + HT - r - 2
    if xa > xb or ya > yb:
        break
    t = uniq([rows[Y0+r][x] for x in range(xa, xb+1)])
    b = uniq([rows[y1][x] for x in range(xa, xb+1)])
    l = uniq([rows[y][X0+r] for y in range(ya, yb+1)])
    rr = uniq([rows[y][x1] for y in range(ya, yb+1)])
    print(f" {r:2d}  {t[0]:2d}({t[1]})      {rr[0]:2d}({rr[1]})      {b[0]:2d}({b[1]})      {l[0]:2d}({l[1]})")
    layers.append(dict(ring=r, top=t[0], right=rr[0], bottom=b[0], left=l[0],
                       pure=all(x[1] == 1 for x in (t, b, l, rr))))
print("\n四角（層 0）:", rows[Y0][X0], rows[Y0][X0+WD-1], rows[Y0+HT-1][X0], rows[Y0+HT-1][X0+WD-1])
print("內部填色:", uniq([rows[Y0+HT//2][x] for x in range(X0+8, X0+WD-8)])[0])
with open('/work/docs/internals/data/frame-profile.json', 'w') as f:
    json.dump(dict(rect=[X0, Y0, WD, HT], layers=layers), f, indent=2)
