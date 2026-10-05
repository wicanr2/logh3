"""第114輪：核對正常增援、占領報告、完整存讀與凍結來源。"""
import argparse
import hashlib
import json
import os
from pathlib import Path

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('--write', action='store_true')
p.add_argument('--gui', default='d')
p.add_argument('--gui-output', default='workplace/reinforce79-r114-replay-d-gui-e')
a = p.parse_args()
target = Path('docs/internals/data/battle-reinforcement.json')
parent_path = Path('docs/internals/data/battle-flagship-approach.json')


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda: f.read(1024*1024), b''):
            h.update(chunk)
    return h.hexdigest()


def source(path):
    st = path.stat()
    assert st.st_uid == 1000 and st.st_gid == 1000, str(path)
    return {'sha256': sha(path), 'bytes': st.st_size}


def load(path):
    return json.loads(Path(path).read_text())


def equal_trees(left, right):
    x, y = Path(left), Path(right)
    xs = {str(p.relative_to(x)): sha(p) for p in x.rglob('*') if p.is_file()}
    ys = {str(p.relative_to(y)): sha(p) for p in y.rglob('*') if p.is_file()}
    assert xs and xs == ys, (left, right)
    return len(xs)


def verify_sources(sources):
    for name, expected in sources.items():
        actual = source(Path(name))
        assert actual == expected, (name, actual, expected)


if not a.write:
    receipt = load(target)
    assert source(parent_path) == receipt['parent']
    verify_sources(load(parent_path)['sources'])
    verify_sources(load('docs/internals/data/battle-winner-player-search.json')['originals'])
    verify_sources(load('docs/internals/data/planet-eligibility.json')['implementation_sources'])
    verify_sources(receipt['sources'])
    print('第114輪來源保持', len(receipt['sources']), '父來源', len(load(parent_path)['sources']))
    raise SystemExit

assert not target.exists()
assert sha(parent_path) == '6136fb277077f49dc2a9656b3c0ce202d17902346d6dbc64052421bdfea7d013'
parent = load(parent_path)
verify_sources(parent['sources'])
verify_sources(load('docs/internals/data/battle-winner-player-search.json')['originals'])
verify_sources(load('docs/internals/data/planet-eligibility.json')['implementation_sources'])
helpers = equal_trees('workplace/reinforce79-r114-m', 'workplace/reinforce79-r114-n')
bootstrap = equal_trees('workplace/reinforce79-r114-replay-c-plan', 'workplace/reinforce79-r114-replay-d-plan')
d = load('workplace/reinforce79-r114-m/normal-0/observations.json')
assert d['state'] == 'StateCommandPhase' and d['capture_reports'] == 1 and d['victory_reports'] == 2
assert sum(x['kind'] == 'seed' for x in d['inputs']) == 1
assert len(d['strategic_orders']) == 10
assert [(x['phase'], x['after']['index'], x['after']['target'], x['after']['move']) for x in d['strategic_orders']] == [
    (0,1,-1,5),(0,4,11,0),(0,5,11,0),(0,10,11,0),(0,12,11,0),
    (1,4,13,1),(1,5,13,1),(1,10,13,1),(1,12,13,1),(1,1,13,1)]
captures = [x for x in d['occupation_results'] if x['event'] is not None]
assert len(captures) == 1
c = captures[0]
assert c['region'] == 8 and c['battle_turn'] == 4 and c['event'] == {'Fleet':7,'Planet':0}
assert c['done'] is False and c['result']['Hits'] == 4
end = next(x for x in d['occupation_results'] if x['region']==8 and x['done'])
assert end['seed'] == c['seed'] and end['result']['Hits'] == 4
assert end['result']['Merit']['7']-c['result']['Merit']['7'] == 4800
assert end['result']['Merit']['23']-c['result']['Merit']['23'] == 2400
plan = load('workplace/reinforce79-r114-replay-'+a.gui+'-plan/plan.json')
for expected in d['observations']:
    got = [x for x in plan['lifecycle_reports'] if all(x[k]==expected[k] for k in ['kind','region','battle_turn','text_hex','face'])]
    assert len(got)==1 and got[0]['tables']==expected['tables'] and got[0]['seed']==expected['seed']
gui_path = Path(a.gui_output)
assert str(gui_path).startswith('workplace/reinforce79-r114-')
gui = load(gui_path/'receipt.json')
assert gui['test_tps']==60 and gui['battle_planet_capture'] is True and gui['versus_loaded'] is True
assert gui['six_diagnostic_saves_equal'] and gui['public_save_load_equal']
assert all(x['x11_different']==0 and x['game_index_equal'] for x in gui['comparisons'])
assert [(x['kind'],x['seed']) for x in gui['actions'] if x['kind']=='seed']==[('seed',0x12345678)]
assert any(x['kind']=='click' and x['x']==300 and x['y']==282 and
           x['before']['state']=='StateLoadGame' and x['after']['state']=='StateLoadFaction'
           for x in gui['actions'])
updates = 0
with (gui_path/'updates.jsonl').open() as f:
    for line in f:
        assert json.loads(line)['tps']==60
        updates += 1
assert updates > 0
for name in ['FLEET','SHIP','ADMIRAL','WORK','PLANET','STELLAR']:
    f = name+'.DT0'
    helper = Path('workplace/reinforce79-r114-m/normal-0')/f
    base = Path('workplace/reinforce79-r114-replay-'+a.gui+'-plan')
    assert helper.read_bytes()==(base/f).read_bytes()==(base/'player-save'/f).read_bytes()==(gui_path/'player-save'/f).read_bytes()
    assert (base/'loaded-save'/f).read_bytes()==(gui_path/f).read_bytes()
standard = Path('workplace/round114-standard.log').read_text()
assert sum(x.startswith('ok  ') for x in standard.splitlines())==13 and 'FAIL' not in standard
paths = [Path('tools/scripts/battle_reinforcement_'+x+'.py') for x in ['probe','replay','receipt']]
paths += [Path('docs/goal/2026-10-06-battle-reinforcement-round-114.md'), Path('docs/playtest/102-battle-reinforcement.md')]
paths += [p for p in Path('workplace').glob('reinforce79-r114*') if p.is_file()]
paths += [q for p in Path('workplace').glob('reinforce79-r114*') if p.is_dir() for q in p.rglob('*') if q.is_file()]
paths += [p for p in Path('workplace').glob('round114-*.log') if p.is_file()]
receipt = {'scope':'remake normal player validation; no new original whole-turn parity',
    'head':'05689e8495054d8e6d6bc3296952e33c17815238',
    'worktree_dirty_at_test':True,
    'image':'logh3-ebiten:2',
    'image_id':'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
    'oracle_head_unchanged':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
    'scenario':'public S1 load; Empire; two human seats',
    'seed_first_tactical_execute':0x12345678,'seed_set_count':1,
    'state_injection':False,'test_tps':60,'parent':source(parent_path),
    'gui_raw_classification':gui['classification'],
    'classification_erratum':'Raw GUI classification inherited S3; actual public S1 load is verified by row y=282 and recorded control states. Raw receipt remains unchanged.',
    'gui_camera_gestures':len(gui['camera_gestures']),
    'gui_last_input_seconds':max(x.get('at_seconds',0) for x in gui['actions']),
    'helper_files_equal':helpers,'bootstrap_files_equal':bootstrap,
    'gui_revision':a.gui,'gui_output':str(gui_path),'gui_actions':len(gui['actions']),'gui_frames':len(gui['comparisons']),
    'gui_pixels':sum(x['pixels'] for x in gui['comparisons']),'gui_updates_tps60':updates,
    'winner_capture':{k:c[k] for k in ['region','battle_turn','seed','event','done']},
    'capture_hits':4,'merit_after_dismiss':[4800,2400],
    'original_whole_turn_verified':False,'spec062':'DRAFT','issue79':'OPEN',
    'sources':{str(x):source(x) for x in sorted(set(paths))}}
with target.open('x') as f:
    f.write(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
print('第114輪核對及凍結',len(receipt['sources']),source(target))
