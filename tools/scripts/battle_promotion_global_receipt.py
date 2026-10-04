"""凍結第86輪來源及全局組合驗算，拒絕覆寫。"""
from pathlib import Path
import collections
import hashlib
import json

def sha(b):
    return hashlib.sha256(b).hexdigest()

a = Path('workplace/promotion79-r86-global-a.json').read_bytes()
assert a == Path('workplace/promotion79-r86-global-b.json').read_bytes()
native = json.loads(a)
assert native['schema'] == 'logh3.battle-promotion-global-native.v1'
assert native['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
assert native['dosgolem_dirty'] is False and native['uid'] == 1000
old = json.loads(Path('workplace/promotion79-r85-recalc-a.json').read_bytes())
v2bytes = Path('docs/internals/data/battle-promotion-recalc-v2.json').read_bytes()
assert sha(v2bytes) == 'cda14b5b122e4e875a825948c61194c96b534a07c7a563c770b26570b35291e6'
v2 = json.loads(v2bytes)
probe = json.loads(Path('workplace/promotion79-r86-global-probe-v3.json').read_bytes())
assert probe['schema'] == 'logh3.battle-promotion-global-probe.v3'
assert probe['checkpoints'] == 53 and probe['different_bytes'] == 0
assert len(native['rows']) == len(old['rows']) == len(v2['cases']) == len(probe['finals']) == 66
sizes = {'admiral': 2296, 'fleet': 3904, 'ship': 27456, 'work': 230, 'stellar': 4032, 'planet': 6706, 'paths': 6144}
checks = collections.Counter()
rows = []
cursor = 0
for new, prev, modeled, final in zip(native['rows'], old['rows'], v2['cases'], probe['finals']):
    assert new['name'] == prev['name'] == modeled['name'] == final['name']
    for k in new.keys() - {'before', 'after', 'recalculations'}:
        assert new[k] == prev[k], (new['name'], k)
    for stage in ('before', 'after'):
        for n, width in sizes.items():
            assert len(bytes.fromhex(new[stage][n])) == width
        for n in ('admiral', 'fleet', 'ship', 'work'):
            assert new[stage][n] == prev[stage][n]
        # 舊觀測器從正確基底後32 bytes開始；前4000 bytes可逐byte回查。
        assert new[stage]['stellar'][64:] == prev[stage]['stellar'][:-64]
    for newcp, prevcp, modelcp in zip(new['recalculations'] or [], prev['recalculations'] or [], modeled['checkpoints']):
        current = probe['cases'][cursor]
        assert newcp['location'] == prevcp['location'] == modelcp['location'] == current['location']
        assert newcp['step'] == prevcp['step'] == modelcp['step'] == current['step']
        assert new['name'] == current['name']
        for n in ('admiral', 'fleet', 'ship'):
            assert newcp['tables'][n] == prevcp['tables'][n]
            actual = sha(bytes.fromhex(newcp['tables'][n]))
            assert actual == modelcp['model_sha256'][n] == current[n+'_input_sha256']
        for n in ('stellar', 'work'):
            assert sha(bytes.fromhex(newcp['tables'][n])) == current[n+'_sha256']
            assert current[n+'_differences'] == []
        for n in ('planet', 'paths'):
            expected_key = 'planet_sha256' if n == 'planet' else 'path_sha256'
            assert sha(bytes.fromhex(newcp['tables'][n])) == current[expected_key]
        cursor += 1
    for n in ('stellar', 'work', 'planet'):
        assert sha(bytes.fromhex(new['after'][n])) == final[n+'_sha256']
    tabledata = {}
    for n in ('admiral', 'fleet', 'ship', 'stellar', 'work', 'planet'):
        before, after = (bytes.fromhex(new[stage][n]) for stage in ('before', 'after'))
        delta = [{'offset_hex': hex(i), 'before': v, 'after': after[i]} for i, v in enumerate(before) if v != after[i]]
        checks[n+'_changed_bytes'] += len(delta)
        tabledata[n] = {'bytes': len(before), 'before_sha256': sha(before), 'after_sha256': sha(after), 'differences': delta}
    rows.append({k: new[k] for k in ('name', 'case', 'injections', 'seed_after_hex', 'steps')} | {'tables': tabledata})
assert cursor == 53
checks.update({'cases': 66, 'recalculations': 53, 'random_outputs': v2['summary']['random_outputs'], 'upgrades': v2['summary']['upgrades']})
sources = []
for p in [
    'workplace/promotion79-r86-global-a.json', 'workplace/promotion79-r86-global-b.json',
    'workplace/promotion79-r86-global.go', 'workplace/promotion79-r86-global-probe-v3.json',
    'workplace/promotion79-r85-recalc-a.json', 'workplace/promotion79-r85-recalc.go',
    'workplace/oracle/ownership79-r82-entry-a/mem-in.bin',
    'workplace/oracle/ownership79-r82-entry-a/entry.state',
    'docs/internals/data/battle-promotion-recalc-v2.json', 'docs/internals/data/star-graph.json',
    'tools/scripts/battle_promotion_global_prepare.py', 'tools/scripts/battle_promotion_global_probe.go',
    'tools/scripts/battle_promotion_global_receipt.py', 'tools/battle_promotion_global_oracle.sh',
    'tools/battle_promotion_global_native_run.sh',
    'internal/rules/growth.go', 'internal/rules/region.go', 'internal/rules/unrest.go',
    'internal/rules/stargraph.go', 'internal/rules/stardistance.go', 'internal/rules/save.go',
]:
    p = Path(p)
    assert p.is_file() and p.stat().st_uid == 1000
    b = p.read_bytes()
    sources.append({'path': str(p), 'bytes': len(b), 'sha256': sha(b)})
result = {
    'schema': 'logh3.battle-promotion-global.v1', 'grade': 'proven',
    'scope': '66個固定8-4入口診斷；53次完整三表沿已驗v2模型，全局經濟／稅／騷動／費用／造艦預估組合驗算；沒有正式Game或正常玩家驗收',
    'errata': '第85輪STELLAR讀取DS:A7BC，正確DS:A79C；舊三表及費用證據保持。原始paths欄位為EMS frame，只有重算返回後獨立建表逐byte核過的資料才能當路徑矩陣。',
    'oracle_identity': {k: v for k, v in native.items() if k != 'rows'},
    'original_json_sha256': sha(a), 'summary': dict(checks), 'global_model': probe,
    'cases': rows, 'sources': sources,
}
dest = Path('docs/internals/data/battle-promotion-global.json')
with dest.open('x', encoding='utf-8') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write('\n')
print('全局來源通過', dict(checks))
print(dest, len(dest.read_bytes()), sha(dest.read_bytes()), 'sources', len(sources))
print('原版JSON', len(a), sha(a))
