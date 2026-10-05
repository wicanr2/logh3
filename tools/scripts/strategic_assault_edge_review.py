"""核對#56固定原版、正式規則反例與DRAFT修法；來源只凍結一次。"""
from pathlib import Path
import ast
import base64
import hashlib
import json
import sys

root = Path('/src')


def read(name):
    return (root / name).read_bytes()


def digest(b):
    return hashlib.sha256(b).hexdigest()


def pair(a, b):
    assert read(a) == read(b), (a, b)
    return json.loads(read(a))


native = pair('workplace/assault56-r99-v3-a.json', 'workplace/assault56-r99-v3-b.json')
baseline = pair('workplace/assault56-r99-model-v2-a.json', 'workplace/assault56-r99-model-v2-b.json')
probe = pair('workplace/assault56-r99-probe-a.json', 'workplace/assault56-r99-probe-b.json')
assert [r['name'] for r in native['rows']] == ['empty', 'near', 'out', 'far', 'terrain']
assert len(baseline['rows']) == len(probe['rows']) == 10
assert native['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
assert native['dosgolem_dirty'] is False and native['tick'] == 20000
assert native['seed'] == '0x12345678' and native['uid'] == 1000
assert native['program_sha256'] == digest(Path('/orig/GIN3.COM').read_bytes())
guard_calls, terrain_calls, blocked_calls = 0, 0, 0
for row in native['rows']:
    assert row['region'] == 26 and row['entry_cs'] == 0x20b4 and row['entry_ip'] == 0x1497
    assert row['stop'] == 'FFFF:0000'
    assert row['injections'] and all(x['linear'].startswith('lin:') for x in row['injections'])
    for key, size in {'admiral': 2296, 'fleet': 3904, 'ship': 27456, 'work': 230,
                      'planet': 6706, 'stellar': 4032, 'lists': 164}.items():
        assert len(bytes.fromhex(row['before'][key])) == len(bytes.fromhex(row['after'][key])) == size
    calls = [c for c in row['calls'] if c['location'] == '203A:04D1']
    assert calls and all(c['return_ax'] == (30 if row['name'] == 'near' else 0) for c in calls)
    guard_calls += len(calls)
    calls = [c for c in row['calls'] if c['location'] == '1588:04EA']
    terrain_calls += len(calls)
    rejected = [c for c in calls if c['return_ax'] == 0]
    assert len(rejected) == (6 if row['name'] == 'terrain' else 0)
    assert all(c['args'][:2] == [936, 1452] for c in rejected)
    blocked_calls += len(rejected)
for row in baseline['rows']:
    expected = 14 if row['name'] == 'near' else 36 if row['name'] == 'terrain' and not row['with_terrain'] else 0
    assert len(row['differences']) == expected, row
for row in probe['rows']:
    expected = 36 if row['name'] == 'terrain' and not row['with_terrain'] else 0
    assert len(row['differences']) == expected, row
near, out = [next(r for r in native['rows'] if r['name'] == n)['after'] for n in ['near', 'out']]
effect = {k: sum(a != b for a, b in zip(bytes.fromhex(near[k]), bytes.fromhex(out[k]))) for k in near}
assert effect == {'admiral': 2, 'fleet': 2, 'lists': 0, 'planet': 1, 'ship': 26, 'stellar': 0, 'work': 3}

index = root / 'docs/internals/data/strategic-assault-edges.json'
if sys.argv[1:] == ['--freeze']:
    assert not index.exists()
    assert index.parent.stat().st_uid == 1000
    frames = root / 'workplace/assault56-r99-frames'
    assert not frames.exists() and frames.parent.stat().st_uid == 1000
    frames.mkdir()
    frame_meta = []
    for row in native['rows']:
        assert len(row['frames']) == 1
        frame = row['frames'][0]
        for key, ext in [('png', '.png'), ('screen', '.bin'), ('palette_rgb', '-rgb16.bin')]:
            b = base64.b64decode(frame[key])
            if key == 'screen':
                assert len(b) == 307200
            if key == 'palette_rgb':
                assert len(b) == 48
            p = frames / (row['name'] + ext)
            p.write_bytes(b)
            frame_meta.append({'name': row['name'], 'kind': key, 'path': str(p.relative_to(root)), 'sha256': digest(b)})
    names = {
        'go.mod', 'go.sum', 'tools/go.sh', 'tools/gui.sh', 'tools/py.sh',
        'docs/goal/2026-10-05-strategic-assault-edges-round-99.md',
        'docs/re/433-strategic-assault-edges.md', 'docs/spec/115-strategic-assault-guard.md',
        'docs/spec/058-strategic-planet-assault.md', 'docs/internals/53-strategic-planet-assault.md',
        'docs/internals/data/battle-finish-attack.json', 'docs/internals/data/settings-interface.json',
        'docs/internals/data/battle-terrain.json', 'docs/internals/data/battle-formation-slots.json',
        'workplace/finish79-r89-native.go',
        'workplace/oracle/cleanup79-r88-entry-a/entry.state',
        'workplace/oracle/cleanup79-r88-entry-a/mem-in.bin',
    }
    for directory in ['internal/rules', 'internal/asset', 'internal/gfx']:
        names.update(str(p.relative_to(root)) for p in (root / directory).glob('*.go') if not p.name.endswith('_test.go'))
    names.update(str(p.relative_to(root)) for p in (root / 'tools').glob('strategic_assault_edge*.sh'))
    names.update(str(p.relative_to(root)) for p in (root / 'tools/scripts').glob('strategic_assault_*.py'))
    names.update(str(p.relative_to(root)) for p in (root / 'tools/scripts').glob('strategic_assault_*.go'))
    names.update(str(p.relative_to(root)) for p in (root / 'workplace').glob('assault56-r99-*') if p.is_file())
    names.update(str(p.relative_to(root)) for p in frames.iterdir())
    sources = {name: {'bytes': len(read(name)), 'sha256': digest(read(name))} for name in sorted(names)}
    original = {name: {'bytes': len((Path('/orig') / name).read_bytes()),
                       'sha256': digest((Path('/orig') / name).read_bytes())}
                for name in ['GIN3.COM'] + [f'MAP{i:02d}.CMP' for i in range(32)]}
    result = {'schema': 'logh3.strategic-assault-edges.v1', 'issue': 56, 'date': '2026-10-05',
              'parent_commit': '46dabcb9aa5e7a38225c1281a8c12425a55de6c4', 'remake_dirty': True,
              'dosgolem_head': native['dosgolem_head'], 'dosgolem_dirty': False,
              'go_image': 'golang:1.24-bookworm', 'gui_image': 'logh3-ebiten:2',
              'spec115': 'DRAFT', 'production_changed': False,
              'guard_queries': guard_calls, 'terrain_queries': terrain_calls, 'blocked_queries': blocked_calls,
              'effect_bytes_near_vs_out': effect, 'model_rows': 10, 'probe_with_terrain_zero_difference_rows': 5,
              'scope': '原版明示戰略診斷；正式typed反例及/tmp守軍修法，不含完整原池／正式Game／正常戰後清單生命週期',
              'frames': frame_meta, 'original_inputs': original, 'sources': sources}
    with index.open('x') as f:
        json.dump(result, f, ensure_ascii=False, indent=2)
        f.write('\n')
elif sys.argv[1:] != ['--check']:
    raise SystemExit('使用 --freeze 或 --check')
result = json.loads(index.read_bytes())
for name, wanted in result['sources'].items():
    b = read(name)
    assert {'bytes': len(b), 'sha256': digest(b)} == wanted, name
for name, wanted in result['original_inputs'].items():
    b = (Path('/orig') / name).read_bytes()
    assert {'bytes': len(b), 'sha256': digest(b)} == wanted, name
for name in ['tools/scripts/strategic_assault_edge_prepare.py', 'tools/scripts/strategic_assault_edge_review.py']:
    ast.parse(read(name), filename=name)
print('五案原版雙跑、正式反例、DRAFT修法及', len(result['sources']), '來源符合；SHA-256', digest(index.read_bytes()))
print('守軍查詢', guard_calls, '地形查詢', terrain_calls, '拒絕', blocked_calls)
