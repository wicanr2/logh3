# 解對了嗎？拿音級分佈做調性擬合。解錯的話音級會接近均勻分佈，
# 相關係數上不去；解對的話會貼合大調或小調的樣板。
import glob, json, math, os
MAJ = [6.35,2.23,3.48,2.33,4.38,4.09,2.52,5.19,2.39,3.66,2.29,2.88]
MIN = [6.33,2.68,3.52,5.38,2.60,3.53,2.54,4.75,3.98,2.69,3.34,3.17]
NAMES = 'C C# D D# E F F# G G# A A# B'.split()

def corr(a, b):
    ma, mb = sum(a)/12, sum(b)/12
    da, db = [x-ma for x in a], [x-mb for x in b]
    n = sum(x*y for x, y in zip(da, db))
    d = math.sqrt(sum(x*x for x in da) * sum(y*y for y in db))
    return n/d if d else 0

for p in sorted(glob.glob('/work/docs/internals/data/mds/*.json')):
    if p.endswith('index.json'):
        continue
    s = json.load(open(p, encoding='utf-8'))
    h = [0]*12
    for t in s['tracks']:
        for e in t['events']:
            if e[0] == 'note' and e[4]:
                h[e[1] % 12] += e[3]      # 用音長加權
    if not sum(h):
        continue
    best = max(((corr(h[k:]+h[:k], tpl), NAMES[k]+m)
                for k in range(12) for tpl, m in ((MAJ, ' 大調'), (MIN, ' 小調'))))
    print('%-10s %-30s 最佳調性 %-8s 相關 %.3f'
          % (os.path.basename(p)[:-5], s['title'][:30], best[1], best[0]))
