"""測「距離表高位元組 k」的候選定義（`docs/re/242`）。

    tools/py.sh tools/scripts/distance_k_probe.py [swap]

拿 `star-distance.json` 的 k（＝ 高位元組 ÷ 4）比對「沿著中點矩陣
展開的路徑上，滿足某條件的星域有幾格」。**十個候選都不成立**，
最好的 `not_mine_all` 也只有 743／1024——結果記在 `docs/re/242`，
留著這支是為了讓下一個人不要重跑同一批。

`swap` 參數換中點矩陣的索引順序（也試過，沒有變好）。
"""
import json, sys
SWAP = len(sys.argv) > 1
g = json.load(open('docs/internals/data/star-graph.json'))
d = json.load(open('docs/internals/data/star-distance.json'))
roll = json.load(open('docs/internals/data/month-rollover-796-01.json'))
owner = {r['index']: r['owner'] for r in roll['regions']}
adj = g['adjacency']

def path(mid, r, c):
    """遞迴展開中點矩陣（與 rules.StarGraph.Route 同一支）。"""
    def walk(i, j, depth):
        if i == j:
            return [i]
        if depth > 40:
            return None
        v = mid[i][j]
        if v == 0xFF:
            return [i, j]
        a, b = walk(i, v, depth + 1), walk(v, j, depth + 1)
        if a is None or b is None:
            return None
        return a[:-1] + b
    return walk(r, c, 0)

for side, key in (('empire', 'midpoint_empire'), ('alliance', 'midpoint_alliance')):
    mid = g[key]
    fac = 0 if side == 'empire' else 1
    ok = bad = nopath = 0
    guesses = {}
    for r in range(32):
        for c in range(32):
            p = path(mid, r, c)
            if p is None:
                nopath += 1
                continue
            k = d[side][r][c] >> 8
            # 幾種候選定義
            cand = {
                'not_mine_all': sum(1 for x in p if owner[x] != fac),
                'not_mine_mid': sum(1 for x in p[1:-1] if owner[x] != fac),
                'not_mine_tail': sum(1 for x in p[1:] if owner[x] != fac),
                'not_mine_head': sum(1 for x in p[:-1] if owner[x] != fac),
                'enemy_all': sum(1 for x in p if owner[x] == 1 - fac),
                'enemy_mid': sum(1 for x in p[1:-1] if owner[x] == 1 - fac),
                'enemy_tail': sum(1 for x in p[1:] if owner[x] == 1 - fac),
                'neutral_all': sum(1 for x in p if owner[x] == 2),
                'hops': len(p) - 1,
                'endpoint_not_mine': int(owner[c] != fac),
            }
            for name, v in cand.items():
                guesses.setdefault(name, [0, 0])
                if v * 4 == k:
                    guesses[name][0] += 1
                else:
                    guesses[name][1] += 1
            ok += 1
    print(side, '可走的對', ok, '走不到', nopath)
    for name, (hit, miss) in sorted(guesses.items()):
        print('   %-14s 中 %4d 不中 %4d' % (name, hit, miss))
