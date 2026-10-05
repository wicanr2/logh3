"""第107輪正常行星換手收據核對器。只在Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import struct

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
choice = p.add_mutually_exclusive_group(required=True)
choice.add_argument('--freeze', action='store_true')
choice.add_argument('--check', action='store_true')
mode = '--freeze' if p.parse_args().freeze else '--check'
target = Path('docs/internals/data/battle-occupation-player.json')


def digest(path):
    data = Path(path).read_bytes()
    return {'bytes': len(data), 'sha256': hashlib.sha256(data).hexdigest()}


def equal_folders(left, right):
    a = sorted(str(v.relative_to(left)) for v in left.rglob('*') if v.is_file())
    b = sorted(str(v.relative_to(right)) for v in right.rglob('*') if v.is_file())
    assert a == b
    for name in a:
        assert (left / name).read_bytes() == (right / name).read_bytes(), name
    return len(a)


root = Path('workplace')
probe = root / 'occupy79-r107-s3planetb/normal-1'
observed = json.loads((probe / 'observations.json').read_text())
assert observed['state'] == 'StateBattle' and observed['capture_reports'] == 0
assert len(observed['planet_captures']) == len(observed['ownership_reports']) == 1
capture = observed['planet_captures'][0]
assert (capture['region'], capture['battle_turn']) == (26, 6)
assert [(v['fleet'], v['damage'], v['counter'], v['captured']) for v in capture['assaults']] == [
    (16, 558, 189, False), (17, 279, 218, False), (28, 943, 0, True)]
planet = capture['planets'][0]
assert (planet['index'], planet['owner'], planet['x'], planet['y']) == (63, 1, 1320, 812)
assert planet['stock'] == {'missiles': 20000, 'fighters': 0, 'ships': [0] * 7}
input_data = json.loads((probe / 'inputs.json').read_text())
seeds = [v for v in input_data['inputs'] if v['kind'] == 'seed']
assert len(seeds) == 1 and seeds[0]['seed'] == 0x12345678
plans = [root / ('occupy79-r107-replay-' + rev + '-plan') for rev in 'abcd']
repeats = [equal_folders(plans[0], folder) for folder in plans[1:]]
assert repeats == [32, 32, 32]
for suffix in ['-game.go', '-main.go']:
    files = [root / ('occupy79-r107-replay-' + rev + suffix) for rev in 'abcd']
    assert len({digest(v)['sha256'] for v in files}) == 1
plan = json.loads((plans[2] / 'plan.json').read_text())
assert len(plan['actions']) == 472
assert len(plan['lifecycle_reports']) == 1
report = plan['lifecycle_reports'][0]
assert (report['region'], report['battle_turn'], report['face']) == (26, 6, 75)
region = bytes.fromhex(report['tables']['stellar'])[26 * 126:27 * 126]
assert region[18] == 2 and list(region[22:25]) == [0, 3, 0]
gui = root / 'occupy79-r107-replay-c-gui'
receipt = json.loads((gui / 'receipt.json').read_text())
assert receipt['test_tps'] == 60 and len(receipt['actions']) == 472
assert receipt['battle_planet_capture'] and receipt['ownership_report']
assert not receipt['natural_finish'] and not receipt['natural_annihilation']
assert len(receipt['comparisons']) == 4
camera_count = len(receipt['camera_gestures'])
assert all(set(['before_view', 'wanted_view', 'after', 'at_seconds']) <= set(v)
           for v in receipt['camera_gestures'])
for row in receipt['comparisons']:
    assert row['pixels'] == 307200 and row['x11_different'] == 0 and row['game_index_equal']
    name = row['name']
    assert (gui / (name + '.bin')).read_bytes() == (plans[2] / (name + '.bin')).read_bytes()
updates = 0
with (gui / 'updates.jsonl').open() as f:
    for line in f:
        st = json.loads(line)
        assert st['tps'] == 60
        updates += 1
names = ['ADMIRAL', 'FLEET', 'PLANET', 'SHIP', 'STELLAR', 'WORK']
saved_bytes = 0
load_diff = {}
for name in names:
    file = name + '.DT0'
    before = (plans[2] / file).read_bytes()
    public = (plans[2] / 'player-save' / file).read_bytes()
    loaded = (plans[2] / 'loaded-save' / file).read_bytes()
    assert before == public == (probe / file).read_bytes()
    assert public == (gui / 'player-save' / file).read_bytes()
    assert loaded == (gui / file).read_bytes()
    saved_bytes += len(before)
    if name != 'STELLAR':
        assert before == loaded
    else:
        diff = [{'offset': i, 'before': x, 'after': y}
                for i, (x, y) in enumerate(zip(before, loaded)) if x != y]
        assert [v['offset'] for v in diff] == [3319, 3401, 3527, 3653]
        assert before[26 * 126 + 18] == loaded[26 * 126 + 18] == 1
        load_diff[name] = diff
    if name == 'PLANET':
        assert before[63 * 82 + 17] == 1
        assert struct.unpack_from('<II', before, 63 * 82 + 60) == (20000, 0)
assert saved_bytes == 44560
standard = (root / 'round107-standard-gui.log').read_text()
assert standard.count('ok  \t') == 13 and standard.count('[no test files]') == 5
assert 'FAIL' not in standard

if mode == '--freeze':
    assert not target.exists()
    paths = {
        'tools/scripts/battle_occupation_player_probe.py',
        'tools/scripts/battle_occupation_player_replay.py',
        'tools/scripts/battle_occupation_player_receipt.py',
        'docs/playtest/095-battle-occupation-player.md',
        'docs/goal/2026-10-05-battle-occupation-player-round-107.md',
        'docs/spec/019-planet-assault.md', 'docs/spec/078-battle-rule-corrections.md',
        'docs/spec/106-battle-region-ownership.md', 'docs/spec/110-battle-finish.md',
        'docs/internals/data/battle-lifecycle-player.json',
        'docs/internals/data/battle-finish-natural.json',
        'docs/internals/data/windows-save-load.json',
        'tools/gui.sh', 'go.mod', 'go.sum',
        'tools/scripts/battle_finish_victory_probe.py',
        'workplace/round79-capture', 'workplace/order79-r93-input',
    }
    for name in ['workplace/lifecycle79-r95-gui-game.go',
                 'workplace/lifecycle79-r95-gui-main.go',
                 'workplace/lifecycle79-r95-gui-plan.go',
                 'workplace/lifecycle79-r95-gui-driver.py']:
        paths.add(name)
    for folder in ['cmd', 'internal']:
        paths.update(str(v) for v in Path(folder).rglob('*.go'))
    for pattern in ['occupy79-r107-*', 'round107-*.log']:
        for v in root.glob(pattern):
            if v.is_file():
                if v.name in ['round107-freeze.log', 'round107-check.log']:
                    continue
                paths.add(str(v))
            else:
                paths.update(str(w) for w in v.rglob('*') if w.is_file())
    sources = {name: digest(name) for name in sorted(paths)}
    old = json.loads(Path('docs/internals/data/windows-save-load.json').read_text())
    originals = {'org_game/extracted/' + name: value for name, value in old['original_files'].items()}
    for name, want in originals.items():
        assert digest(name) == want, name
    result = {
        'schema': 'logh3.battle-occupation-player.v1', 'date': '2026-10-05', 'issue': 79,
        'parent_commit': '9b5c4b451022286c30b9731c17dba4cf71fbe257',
        'classification': 'remake正常第六步行星換手與8-2報告；沒有新原版整回合同態',
        'build_image': 'logh3-ebiten:2', 'engine_changed': False,
        'seed': 0x12345678, 'seed_once': True, 'state_injection': False,
        'bootstrap_repeats': 4, 'repeat_files_equal': 32,
        'gui': {'revision': 'c', 'tps': 60, 'actions': 472, 'frames': 4,
                'pixels': 1228800, 'different': 0, 'updates': updates,
                'extra_camera_gestures': camera_count, 'save_bytes': saved_bytes},
        'loaded_stellar_differences': load_diff['STELLAR'],
        'sources': sources, 'originals': originals, 'original_bundle': old['original_bundle'],
        'limitations': ['110勝後占領GUI仍缺', '非旗艦自然死亡／空組與自然升級仍缺',
                        '062 DRAFT及#79 OPEN；沒有新dosgolem收據',
                        '未新增#48登陸／要塞砲演出callback對拍',
                        '8-2等待時PLANET快取未同步，不稱完整六表same-state'],
    }
    with target.open('x') as f:
        f.write(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
index = json.loads(target.read_text())
assert digest(index['original_bundle']['path'])['sha256'] == index['original_bundle']['sha256']
for group in ['sources', 'originals']:
    for name, want in index[group].items():
        assert digest(name) == want, name
print(len(index['sources']), '來源、', len(index['originals']), '原始輸入符合；SHA-256', digest(target)['sha256'])
