"""兩份**原版收據**之間的差異切成連通分量。

    tools/py.sh tools/scripts/oracle_parts.py <批次A> <批次B> [遮罩 x0:y0:x1:y1 ...]

`diff_parts.py` 比的是 remake 對原版；這一支比的是原版對原版——問
「這兩個時刻差在哪」。差異切成連通分量之後，一句結論常常只要一次比對：
`vs5` 對 `vs3` 扣掉游標只差六個字（兩處「帝國軍」變「同盟軍」），
`vs6` 對 `vs3` 只差日期那一格（`docs/re/76`）。

遮罩用來把已知會不同的那幾塊排除掉（疊上去的視窗、游標），
問「除了這些以外還差什麼」。
"""
import sys

W, H = 640, 480
a = open('/work/workplace/oracle/%s/screen.bin' % sys.argv[1], 'rb').read()
b = open('/work/workplace/oracle/%s/screen.bin' % sys.argv[2], 'rb').read()
rects = [tuple(int(v) for v in r.split(':')) for r in sys.argv[3:]]


def masked(x, y):
    return any(x0 <= x <= x1 and y0 <= y <= y1 for x0, y0, x1, y1 in rects)


diff = {(i % W, i // W) for i in range(len(a))
        if a[i] != b[i] and not masked(i % W, i // W)}
seen, parts = set(), []
for p in sorted(diff):
    if p in seen:
        continue
    stack, comp = [p], []
    seen.add(p)
    while stack:
        x, y = stack.pop()
        comp.append((x, y))
        for n in ((x-1, y), (x+1, y), (x, y-1), (x, y+1),
                  (x-1, y-1), (x+1, y-1), (x-1, y+1), (x+1, y+1)):
            if n in diff and n not in seen:
                seen.add(n)
                stack.append(n)
    xs = [x for x, _ in comp]
    ys = [y for _, y in comp]
    parts.append((len(comp), min(xs), min(ys),
                  max(xs) - min(xs) + 1, max(ys) - min(ys) + 1))
parts.sort(reverse=True)
print('%d 塊，總計 %d 像素' % (len(parts), sum(p[0] for p in parts)))
for n, x, y, w, h in parts[:25]:
    print('  %6d px  (%d,%d) %dx%d' % (n, x, y, w, h))
