# ADMIRAL.DT[89AX] → JSON。82 筆 × 28 bytes。
import json, collections
out = {}
for ext in ('DT8', 'DT9', 'DTA', 'DTX'):
    d = open(f'/orig/ADMIRAL.{ext}', 'rb').read()
    recs = []
    for i in range(len(d) // 28):
        r = d[i*28:(i+1)*28]
        recs.append({
            'index': i,
            'raw': r.hex(),
            'f0': r[0], 'f1_5': list(r[1:6]),
            'f6': r[6], 'f7': r[7], 'f8': r[8], 'f9': r[9], 'f10': r[10],
            'abilities': list(r[11:25]),   # 14 個能力值；+11 ＝ 占領能力（已驗證）
            'tail': list(r[25:28]),
        })
    out[ext] = recs

# 四個版本的差異：哪些欄位會變
diff = collections.Counter()
base = out['DTX']
for ext in ('DT8', 'DT9', 'DTA'):
    for a, b in zip(base, out[ext]):
        for k in range(28):
            if bytes.fromhex(a['raw'])[k] != bytes.fromhex(b['raw'])[k]:
                diff[k] += 1
print('與 DTX 相異的位元組位置（累計次數）:', sorted(diff.items()))
print('筆數', len(base))
print('第 0 筆 abilities:', base[0]['abilities'])
json.dump({'source': 'ADMIRAL.DT8/DT9/DTA/DTX', 'record_size': 28,
           'count': len(base), 'variants': out},
          open('/work/docs/internals/data/admiral.json', 'w'),
          ensure_ascii=False, indent=1)
print('寫出 docs/internals/data/admiral.json')
