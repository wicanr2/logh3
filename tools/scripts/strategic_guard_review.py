"""核第100輪清單、正式原池與保存；歷史來源以當輪提交核實，不改舊索引。"""
from pathlib import Path
import ast
import hashlib
import json
import subprocess
import sys

root = Path('/src')
parent = '1ff96c1e292b766a93af5c18f28083404407e7a7'
index = root / 'docs/internals/data/strategic-guard.json'


def read(name):
    return (root / name).read_bytes()


def identity(b):
    return {'bytes': len(b), 'sha256': hashlib.sha256(b).hexdigest()}


def pair(a, b):
    assert read(a) == read(b), (a, b)
    return json.loads(read(a))


life = pair('workplace/assault56-r100-lifecycle-v6-a.json', 'workplace/assault56-r100-lifecycle-v6-b.json')
assert identity(read('workplace/assault56-r100-lifecycle-v6-a.json'))['sha256'] == 'd5689d017f511288e74080ca91f25e22d20112a54d47d048b1c88f38d9e1afd0'
assert life['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
assert life['dosgolem_dirty'] is False and life['tick'] == 20000 and life['uid'] == 1000
for row in life['rows']:
    assert not row['error']
    assert len(row['writes']) == 56
    assert {w['location'] for w in row['writes']} == {'12AA:03DC', '12AA:0441', '12AA:045D'}
    assert len({e['list']['bytes'] for e in row['events']}) == 1
    assert all(e['list']['count'] == 55 and len(bytes.fromhex(e['list']['bytes'])) == 164 for e in row['events'])
    assert all(len(bytes.fromhex(e['ship_hex'])) == 27456 for e in row['events'])
    if row['next_assault']:
        entries = [e for e in row['events'] if e['name'] == '20B4:1497']
        assert len(entries) == 1
        stack = bytes.fromhex(entries[0]['stack_hex'])
        assert stack[:6] == bytes.fromhex('ab0574031b00')
        assert [bytes.fromhex(e['stack_hex'])[4] for e in row['events'] if e['name'] == '20B4:1512'] == [26, 27]
        assert row['events'][-1]['cs'] == 0x374 and row['events'][-1]['ip'] == 0x5ab

probe = pair('workplace/assault56-r100-raw-probe-a.json', 'workplace/assault56-r100-raw-probe-b.json')
assert len(probe['rows']) == 5
assert all(not any(r['differences'].values()) and r['seed'] == r['wanted_seed'] for r in probe['rows'])

# 第99輪反例與完整來源固定在當輪HEAD。後續功能修正不能要求舊索引等於新程式。
old = json.loads(read('docs/internals/data/strategic-assault-edges.json'))
assert identity(read('docs/internals/data/strategic-assault-edges.json'))['sha256'] == '07d60572fb700512f99f560366398dfcbdad59daa107f5714d8e92276f1de7d5'
historical = []
for name, wanted in old['sources'].items():
    if identity(read(name)) != wanted:
        p = subprocess.run(['git', '-C', str(root), 'show', parent + ':' + name], capture_output=True, check=True)
        assert identity(p.stdout) == wanted, name
        historical.append(name)

for log in ['workplace/round100-standard-raw-gui.log', 'workplace/round100-strategic-guard-verified.log']:
    s = read(log).decode()
    assert 'FAIL' not in s and 'SKIP' not in s
standard = read('workplace/round100-standard-raw-gui.log').decode().splitlines()
assert sum(l.startswith('ok ') for l in standard) == 13
assert sum('[no test files]' in l for l in standard) == 5
target = read('workplace/round100-strategic-guard-verified.log').decode()
for name in ['TestStrategicGuardRawMatchesOracle', 'TestStrategicGuardMatchesOracle', 'TestStrategicGuardGameOriginal',
             'TestStrategicGuardGameInvalidPoolStops', 'TestStrategicGuardEndBattleRetainsPool',
             'TestBattleFinishGameOriginal', 'TestBattleFinishSaveReload', 'TestOccupyAstarteRunsStrategicAssault']:
    assert '--- PASS: ' + name + ' ' in target, name

if sys.argv[1:] == ['--freeze']:
    assert not index.exists() and index.parent.stat().st_uid == 1000
    names = {'go.mod', 'go.sum', 'tools/go.sh', 'tools/gui.sh',
             'docs/spec/115-strategic-assault-guard.md', 'docs/spec/058-strategic-planet-assault.md',
             'docs/spec/110-battle-finish.md', 'docs/re/434-strategic-guard-lifecycle.md',
             'docs/playtest/088-strategic-guard.md', 'docs/internals/53-strategic-planet-assault.md',
             'docs/goal/2026-10-05-strategic-guard-lifecycle-round-100.md',
             'docs/internals/data/strategic-assault-edges.json',
             'docs/internals/data/battle-terrain.json', 'docs/internals/data/battle-formation-slots.json',
             'docs/internals/data/ship-kind-table.json',
             'workplace/assault56-r99-v3-a.json', 'workplace/assault56-r99-v3-b.json',
             'workplace/oracle/cleanup79-r88-entry-a/entry.state',
             'workplace/round100-standard-raw-gui.log', 'workplace/round100-strategic-guard-verified.log',
             'tools/strategic_assault_edge_probe.sh'}
    for directory in ['internal/rules', 'internal/game', 'internal/asset', 'internal/gfx']:
        names.update(str(p.relative_to(root)) for p in (root / directory).glob('*.go'))
    for pattern in ['tools/scripts/strategic_guard*', 'tools/strategic_guard*', 'workplace/assault56-r100-*']:
        names.update(str(p.relative_to(root)) for p in root.glob(pattern) if p.is_file())
    original = {name: identity((Path('/orig') / name).read_bytes())
                for name in ['GIN3.COM'] + [f'MAP{i:02d}.CMP' for i in range(32)]}
    result = {'schema': 'logh3.strategic-guard.v1', 'issue': 56, 'date': '2026-10-05',
              'parent_commit': parent, 'remake_dirty': True, 'dosgolem_head': life['dosgolem_head'],
              'dosgolem_dirty': False, 'gui_image': 'logh3-ebiten:2', 'go_image': 'golang:1.24-bookworm',
              'spec115': 'CONFORMED-limited', 'production_changed': True,
              'original_lifecycle_cases': 2, 'full_raw_cases': 5, 'full_bytes_per_side': 223940,
              'standard_pass_packages': 13, 'standard_no_test_packages': 5,
              'historical_source_commit': parent, 'historical_changed_bindings': historical,
              'scope': '明示局面五案正式規則／Game六表清單seed、保存及清單契約；正常占領回歸，不含正常原版整回合或新增GUI對拍',
              'sources': {n: identity(read(n)) for n in sorted(names)}, 'original_inputs': original}
    with index.open('x') as f:
        json.dump(result, f, ensure_ascii=False, indent=2)
        f.write('\n')
elif sys.argv[1:] != ['--check']:
    raise SystemExit('使用 --freeze 或 --check')

result = json.loads(index.read_bytes())
for name, wanted in result['sources'].items():
    assert identity(read(name)) == wanted, name
for name, wanted in result['original_inputs'].items():
    assert identity((Path('/orig') / name).read_bytes()) == wanted, name
assert result['historical_changed_bindings'] == historical
for name in ['tools/scripts/strategic_guard_review.py', 'tools/scripts/strategic_guard_lifecycle_prepare.py']:
    ast.parse(read(name), filename=name)
print('PASS 第100輪現行來源', len(result['sources']), '歷史變更', len(historical), identity(index.read_bytes())['sha256'])
