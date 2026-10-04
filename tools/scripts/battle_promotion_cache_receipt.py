"""第87輪固定歸屬變更／控制對照與既有矩陣的原版證據。"""
from pathlib import Path
import hashlib
import json

def sha(b):
    return hashlib.sha256(b).hexdigest()

a = Path('workplace/promotion79-r87-cache-a.json').read_bytes()
assert a == Path('workplace/promotion79-r87-cache-b.json').read_bytes()
native = json.loads(a)
assert native['schema'] == 'logh3.battle-promotion-cache-native.v1'
assert native['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
assert native['dosgolem_dirty'] is False and native['uid'] == 1000
model = json.loads(Path('workplace/promotion79-r87-cache-model-v2.json').read_bytes())
assert len(native['rows']) == model['checkpoints'] == 2 and model['different_bytes'] == 0
assert [c['premature_rebuilt_different_entries'] for c in model['cases']] == [684, 0]
baseline = json.loads(Path('workplace/promotion79-r86-global-a.json').read_bytes())
for row in native['rows']:
    original = next(r for r in baseline['rows'] if r['case'][:5] == row['case'][:5] and r['case'][5] == 0)
    for stage in ('before', 'after'):
        for table in ('admiral', 'fleet', 'ship', 'planet'):
            assert row[stage][table] == original[stage][table]
    returns = {'28E0:098D', '28E0:09A0', '28E0:070D', '28E0:078A', '28E0:0857', '28E0:08E2'}
    assert [(c['location'], c['ax']) for c in row['calls'] if c['location'] in returns] == [(c['location'], c['ax']) for c in original['calls'] if c['location'] in returns]
    assert row['seed_after_hex'] == original['seed_after_hex']
sources = []
for path in [
    'workplace/promotion79-r87-cache-a.json', 'workplace/promotion79-r87-cache-b.json',
    'workplace/promotion79-r87-cache.go', 'workplace/promotion79-r87-cache-model-v2.go',
    'workplace/promotion79-r87-cache-model-v2.json', 'workplace/promotion79-r86-global-a.json',
    'tools/scripts/battle_promotion_global_probe.go', 'tools/scripts/battle_promotion_cache_prepare.py',
    'tools/scripts/battle_promotion_cache_receipt.py', 'tools/battle_promotion_cache_oracle.sh',
    'tools/battle_promotion_cache_native_run.sh', 'docs/internals/data/star-graph.json',
    'workplace/oracle/ownership79-r82-entry-a/mem-in.bin',
    'workplace/oracle/ownership79-r82-entry-a/entry.state',
]:
    p = Path(path)
    assert p.stat().st_uid == 1000
    b = p.read_bytes()
    sources.append({'path': path, 'bytes': len(b), 'sha256': sha(b)})
dest = Path('docs/internals/data/battle-promotion-cache.json')
result = {'schema': 'logh3.battle-promotion-cache.v1', 'grade': 'proven',
          'scope': '兩個固定8-4入口診斷，一個星域歸屬變更、一個同值控制對照；非正常占領流程',
          'oracle_identity': {k: v for k, v in native.items() if k != 'rows'},
          'original_json_sha256': sha(a), 'model': model,
          'injections': [{k: r[k] for k in ('name', 'case', 'injections', 'seed_after_hex', 'steps')} for r in native['rows']],
          'sources': sources}
with dest.open('x', encoding='utf-8') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write('\n')
print(dest, len(dest.read_bytes()), sha(dest.read_bytes()), 'sources', len(sources))
print('原版', len(a), sha(a))
