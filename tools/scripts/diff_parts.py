# 把 remake 與 oracle 的差異切成連通分量，報出每塊的 bbox 與原版的顏色。
# 「還缺什麼」用這個問，比逐處目視快。
import collections, sys
a = open('/work/workplace/remake-maingame.bin', 'rb').read()
b = open('/work/workplace/oracle/key1/screen.bin', 'rb').read()
W, H = 640, 480
diff = {(i % W, i//W) for i in range(len(a)) if a[i] != b[i]}
seen = set()
parts = []
for p in sorted(diff):
    if p in seen:
        continue
    stack, comp = [p], []
    seen.add(p)
    while stack:
        x, y = stack.pop()
        comp.append((x, y))
        for nx, ny in ((x-1, y), (x+1, y), (x, y-1), (x, y+1),
                       (x-1, y-1), (x+1, y-1), (x-1, y+1), (x+1, y+1)):
            if (nx, ny) in diff and (nx, ny) not in seen:
                seen.add((nx, ny))
                stack.append((nx, ny))
    xs = [x for x, _ in comp]
    ys = [y for _, y in comp]
    cols = collections.Counter(b[y*W+x] for x, y in comp)
    mine = collections.Counter(a[y*W+x] for x, y in comp)
    parts.append((len(comp), min(xs), min(ys), max(xs)-min(xs)+1, max(ys)-min(ys)+1,
                  cols.most_common(3), mine.most_common(2)))
parts.sort(reverse=True)
print(f'{len(parts)} 塊，總計 {sum(p[0] for p in parts)} 像素')
for n, x, y, w, h, cols, mine in parts[:14]:
    print(f'  {n:5d} px  ({x},{y}) {w}×{h}  原版 {cols}  remake {mine}')
