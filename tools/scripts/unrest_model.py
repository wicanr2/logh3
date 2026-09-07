"""照 `docs/re/43` 反組譯出來的算式跑一遍，與量到的未安定度對。

    unrest_model.py <unrest-at-load.json> <槽位副檔名>

輸出每個星域的 f1–f4、算到一半的治安度、量到的治安度與差額。
差額就是還沒實作的 f5（提督）＋ f6（艦隊）＋尾巴那兩段修正。
"""
import json
import sys

REC = 126
ADM = 28

data = json.load(open(sys.argv[1]))
ext = sys.argv[2]
slot = [s for s in data["slots"] if s["slot"] == ext][0]

stellar = open("/orig/STELLAR.DT" + ext, "rb").read()
planet = open("/orig/PLANET.DT" + ext, "rb").read()
fleet = open("/orig/FLEET.DT" + ext, "rb").read()
admiral = open("/orig/ADMIRAL.DT" + ext, "rb").read()
pladat = open("/orig/PLADAT.DAT", "rb").read()
graph = json.load(open("/work/docs/internals/data/star-graph.json"))
adj = graph["adjacency"]   # 每個星域一串相鄰星域

# 執行期參數（S3 那一份，`docs/re/43`）
RULER = [16, 60]
A_TAB = [1, 2]        # ds:8FD4
B_TAB = [2, 2]        # ds:8FF8
DI = [50, 75, 100, 150, 200]   # ds:147E
T = [170, 150, 130, 90, 60]    # ds:4B28
TURN = slot["turn"]

counts = [int.from_bytes(pladat[i * 2:i * 2 + 2], "little") for i in range(32)]
start = [sum(counts[:i]) for i in range(32)]


def idiv(a, b):
    """x86 的 `idiv` 往零截斷，Python 的 `//` 往下取整——負數會差 1。

    `d = 100 + (v0 × 100 / di − 100) / 4` 這一步括號內是負的，
    差這 1 會讓 f1 差 2，整支算下來就對不上（`docs/re/43`）。
    """
    q = abs(a) // abs(b)
    return q if (a < 0) == (b < 0) else -q


def rec(i):
    return stellar[i * REC:(i + 1) * REC]


def f1(i, owner):
    r = rec(i)
    a, b = A_TAB[owner], B_TAB[owner]
    di = DI[b]
    growth = int.from_bytes(r[0x1E:0x20], "little")
    v0 = r[0x27] * 100 // growth if growth > 0 else di
    if v0 <= 25:
        v0 = di
    d = 100 + idiv(v0 * 100 // di - 100, 4)
    m = TURN % 10
    blend = ((m + 5) * T[a] + (14 - m) * T[b]) // 19
    return idiv(blend * 100, d)


def f2(i, owner):
    r = rec(i)
    return (r[0x7C] if owner == 0 else 200 - r[0x7C]) // 10


def f3(i, owner):
    out = 0
    for j in adj[i]:
        n = rec(j)
        if n[0x12] == owner:
            if n[0x13] & 0x14:
                out += 7
        else:
            out += 3 if n[0x12] == 2 else 5
            if n[0x13] & 0x14:
                out += 2
    return out


def f4(i, owner):
    n = 0
    for k in range(counts[i]):
        p = planet[(start[i] + k) * 82:(start[i] + k + 1) * 82]
        if p[0x12] != 0:
            n += 1
    return n * 10


# 尾巴那一段：`v -= 2 × 2B99:0450(...)`。拿相鄰圖跑 BFS 出來的
# 「離自己首都幾步」代進去，殘差就從 0…−14 收成一個常數（`docs/re/43`）。
import collections


# 尾巴那一段是 `v -= 2 × 距離`。兩種距離都算出來比：相鄰圖跑 BFS，
# 以及原版自己那張中點矩陣（`docs/internals/26`）遞迴展開的步數。
MID = [graph["midpoint_empire"], graph["midpoint_alliance"]]


def steps(tab, i, j):
    if i == j:
        return 0
    t = tab[i][j]
    return 1 if t == 255 else steps(tab, i, t) + steps(tab, t, j)


def bfs(src):
    d = {src: 0}
    q = collections.deque([src])
    while q:
        u = q.popleft()
        for w in adj[u]:
            if w not in d:
                d[w] = d[u] + 1
                q.append(w)
    return d


CAPITAL = [0, 31]
BFS = [bfs(0), bfs(31)]
MIDD = [{j: steps(MID[f], CAPITAL[f], j) for j in range(32)} for f in (0, 1)]


def f6(i, owner):
    """艦隊那一項：同陣營、`+0x08 == 5` 的艦隊，Σ+0x0F × maxP × maxQ / 10000。"""
    tot, ps, qs = 0, [0], [0]
    for k in range(32):
        r = fleet[k * 122:(k + 1) * 122]
        if r[0x04] == 0xFF or r[0x06] != i or r[0x01] != owner or r[0x08] != 5:
            continue
        tot += r[0x0F]
        a = admiral[r[0x04] * ADM:(r[0x04] + 1) * ADM]
        ps.append(a[0x0B])
        qs.append(a[0x13])
    return tot * max(ps) * max(qs) // 10000

print("星域 own  f1  f2  f3  f4  base  ×能力  f6  加f6後 減步數 量到治安  差")
for e in slot["regions"]:
    i, owner = e["index"], e["owner"]
    if owner == 2:
        print("%2d 無主  未安定度 %d（程式直接寫 100）" % (i, e["at_load"]))
        continue
    a1, a2, a3, a4 = f1(i, owner), f2(i, owner), f3(i, owner), f4(i, owner)
    base = a1 - a2 - a3 + a4
    if rec(i)[0x13] & 4:
        base -= 30
    base = max(base, 0)
    v = idiv(base * admiral[RULER[owner] * ADM + 0x0E], 100)
    v = min(v, 200)
    want = 200 - e["at_load"]
    a6 = f6(i, owner)
    v2 = min(v + a6, 200)
    # 尾巴：`v -= 2 × 步數`。載入的當下中點矩陣還是整片 FF（「都直接相連」），
    # 所以非首都一律 1 步（`docs/re/43`）。
    # 尾巴的步數就是「離自己首都幾步」——對 `2703:0ED6` 設 -regs-at
    # 量到的 31 個回值與相鄰圖跑 BFS 完全相同（`docs/re/43`）。
    hops = BFS[owner][i]
    v3 = max(v2 - 2 * hops, 0)
    want = 200 - e["at_load"]
    print("%2d  %d   %3d %3d %3d %3d  %4d  %4d %4d  %5d  %5d  %5d  %+4d"
          % (i, owner, a1, a2, a3, a4, base, v, a6, v2, v3, want, want - v3))
