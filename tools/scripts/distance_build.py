"""照原版重算兩張距離矩陣與兩張中點矩陣，跟 EMS 傾印逐格對。

    tools/py.sh tools/scripts/distance_build.py <8KB頁框傾印> <1MB傾印> [--write]

`--write` 會把邊權表併進 `docs/internals/data/star-graph.json`。

原版分兩支（`docs/re/243`）：

    2B99:007D(陣營)          用 ds:533A 的 32×32 位元組邊權表建初始矩陣
    2B99:01B1(距離表, 中點)  Floyd–Warshall，中點矩陣是它的見證矩陣

初始矩陣：

    邊權 0xFF                                     → 0x7FFF（不相鄰）
    否則 w ＝ 邊權；擁有者(i) 或 擁有者(j) 不是自己 → w += 0x400

擁有者從表本身反解：相鄰對的兩段路最少 20、任何單邊最多 18，所以
`距離[i][j] == 邊權[i][j]` 就代表那條直接邊沒被罰，也就是兩端都是自己的。
剩下的用約束傳播 ＋ 窮舉，再拿完整重建驗證。
"""
import itertools
import json
import pathlib
import sys

INF = 0x7FFF
DS = 0x54B30
ADJ = DS + 0x533A

frame = pathlib.Path(sys.argv[1]).read_bytes()
mem = pathlib.Path(sys.argv[2]).read_bytes()
write = "--write" in sys.argv

weight = [[mem[ADJ + i * 32 + j] for j in range(32)] for i in range(32)]
SIDES = (("帝國", 0, 0x0000, 0x1000, "midpoint_empire"),
         ("同盟", 1, 0x0800, 0x1400, "midpoint_alliance"))


def words(base):
    return [[int.from_bytes(frame[base + (r * 32 + c) * 2:][:2], "little")
             for c in range(32)] for r in range(32)]


def bytes_(base):
    return [list(frame[base + r * 32:base + r * 32 + 32]) for r in range(32)]


def build(owner, faction):
    """2B99:007D ＋ 2B99:01B1。比較是**無號**的，所以 0x7FFF + 0x7FFF 溢成
    0xFFFE 也不會誤鬆弛。"""
    d = [[INF if weight[i][j] == 0xFF else weight[i][j] +
          (0x400 if owner[i] != faction or owner[j] != faction else 0)
          for j in range(32)] for i in range(32)]
    mid = [[0xFF] * 32 for _ in range(32)]
    for i in range(32):
        d[i][i] = 0
    for k in range(32):
        for i in range(32):
            for j in range(32):
                v = (d[i][k] + d[k][j]) & 0xFFFF
                if v < d[i][j]:
                    d[i][j] = v
                    mid[i][j] = k
    return d, mid


def solve_owners():
    want = {f: words(b) for _, f, b, _, _ in SIDES}
    cand = [{0, 1, 2} for _ in range(32)]
    edges = [(i, j) for i in range(32) for j in range(32)
             if i != j and weight[i][j] != 0xFF]
    changed = True
    while changed:
        changed = False
        for i, j in edges:
            for f, tab in want.items():
                if tab[i][j] == weight[i][j]:      # 沒被罰 → 兩端都是 f
                    for r in (i, j):
                        if cand[r] != {f}:
                            cand[r] = {f}
                            changed = True
                else:                              # 被罰了 → 至少一端不是 f
                    if cand[i] == {f} and f in cand[j]:
                        cand[j].discard(f)
                        changed = True
                    if cand[j] == {f} and f in cand[i]:
                        cand[i].discard(f)
                        changed = True
    for combo in itertools.product(*[sorted(c) for c in cand]):
        if all(build(combo, f)[0] == t for f, t in want.items()):
            return list(combo)
    raise SystemExit("沒有一組擁有者能重現這兩張表")


def path(mid, i, j):
    m = mid[i][j]
    return [i, j] if m == 0xFF else path(mid, i, m)[:-1] + path(mid, m, j)


owner = solve_owners()
print("反解出的星域擁有者：%s" % "".join(map(str, owner)))
ok = True
graph = json.load(open("docs/internals/data/star-graph.json", encoding="utf-8"))
for name, faction, dbase, mbase, gkey in SIDES:
    d, mid = build(owner, faction)
    wd, wm = words(dbase), bytes_(mbase)
    nd = sum(d[r][c] == wd[r][c] for r in range(32) for c in range(32))
    nm = sum(mid[r][c] == wm[r][c] for r in range(32) for c in range(32))
    print("%s：距離 %d/1024、中點 %d/1024 與傾印相同" % (name, nd, nm))
    ok &= nd == 1024 and nm == 1024
    if mid != graph[gkey]:
        print("  ⚠ 與 star-graph.json 的 %s 不同" % gkey)
        ok = False
    # 值的分解：1024 × k ＋ 沿途邊權和，k ＝ 路上兩端不全是自己的邊數。
    bad = 0
    for i in range(32):
        for j in range(32):
            if i == j or d[i][j] >= INF:
                continue
            p = path(mid, i, j)
            k = sum(1 for a, b in zip(p, p[1:])
                    if owner[a] != faction or owner[b] != faction)
            cost = sum(weight[a][b] for a, b in zip(p, p[1:]))
            if 1024 * k + cost != d[i][j]:
                bad += 1
    print("  值 ≠ 1024×k ＋ 邊權和 的格數：%d" % bad)
    ok &= bad == 0

asym = [(i, j) for i in range(32) for j in range(i + 1, 32)
        if weight[i][j] != weight[j][i]]
print("邊權表不對稱的對：%s" % (asym or "無"))

if write:
    graph["edge_cost_note"] = (
        "ds:533A 的 32×32 位元組邊權表（255 ＝ 不相鄰）。`2B99:007D` 拿它建"
        "初始距離矩陣，兩端不全是自己的邊再加 0x400。見 docs/re/243。")
    graph["edge_cost"] = weight
    p = pathlib.Path("docs/internals/data/star-graph.json")
    p.write_text(json.dumps(graph, ensure_ascii=False, indent=1) + "\n",
                 encoding="utf-8")
    print("已寫回 %s" % p)

sys.exit(0 if ok else 1)
