"""把星域路徑表（EMS 裡那兩張 32×32）匯出成 JSON。

輸入是 `-dump-mem lin:E1000:2048` 在路徑計算執行當下倒出來的 2 KB
（`docs/re/28`）。表是**中點矩陣**：`T[i][j] == 255` 代表 i 與 j 直接相連，
否則 T[i][j] 是中繼點，`path(i,j) = path(i,T) + path(T,j)`。
"""
import json, pathlib, sys

src = pathlib.Path(sys.argv[1] if len(sys.argv) > 1 else '/work/workplace/explore/path.bin')
out = pathlib.Path(sys.argv[2] if len(sys.argv) > 2 else
                   '/work/docs/internals/data/star-graph.json')
t = src.read_bytes()
tab = [[list(t[base + i*32:base + (i+1)*32]) for i in range(32)] for base in (0, 1024)]
adj = [sorted(j for j in range(32) if j != i and tab[0][i][j] == 255) for i in range(32)]
adj1 = [sorted(j for j in range(32) if j != i and tab[1][i][j] == 255) for i in range(32)]

def path(i, j, k, depth=0):
    if i == j:
        return [i]
    if depth > 40:
        raise RuntimeError('太深 %d→%d' % (i, j))
    v = tab[k][i][j]
    if v == 255:
        return [i, j]
    return path(i, v, k, depth+1)[:-1] + path(v, j, k, depth+1)

longest = 0
for k in (0, 1):
    a = adj if k == 0 else adj1
    for i in range(32):
        for j in range(32):
            p = path(i, j, k)
            for x, y in zip(p, p[1:]):
                assert y in a[x], '第 %d 張表 %d→%d 走到不相鄰的 %d-%d' % (k, i, j, x, y)
            longest = max(longest, len(p) - 1)

doc = {
 'note': ('星域之間的相鄰關係與路徑表。原版把兩張 32×32 的中點矩陣放在 EMS 裡'
          '（每個陣營一張），far ptr 在 ds:D06A／ds:D06E（docs/re/28）。'
          'T[i][j] == 255 代表 i 與 j 直接相連；否則是中繼點，'
          'path(i,j) = path(i,T) + path(T,j)。'),
 'regions': 32,
 'adjacency_same_for_both': adj == adj1,
 'longest_path_steps': longest,
 'adjacency': adj,
 'midpoint_empire': tab[0],
 'midpoint_alliance': tab[1],
}
out.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + '\n')
print('相鄰表兩張一致：%s；最長 %d 步 → %s' % (adj == adj1, longest, out))
