"""第115輪：核對自然升級、正常軍階UI、公開存讀及完整凍結來源。"""
import argparse
import hashlib
import json
import os
from pathlib import Path

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('--write', action='store_true')
p.add_argument('--check-only', action='store_true')
a = p.parse_args()
target = Path('docs/internals/data/battle-natural-promotion.json')
parent_path = Path('docs/internals/data/battle-reinforcement.json')


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda: f.read(1024*1024), b''):
            h.update(chunk)
    return h.hexdigest()


def source(path):
    st = path.stat()
    assert st.st_uid == st.st_gid == 1000, str(path)
    return {'sha256': sha(path), 'bytes': st.st_size}


def load(path):
    return json.loads(Path(path).read_text())


def verify_sources(sources):
    for name, expected in sources.items():
        assert source(Path(name)) == expected, name


def equal_trees(left, right):
    x, y = Path(left), Path(right)
    xs = {str(p.relative_to(x)): sha(p) for p in x.rglob('*') if p.is_file()}
    ys = {str(p.relative_to(y)): sha(p) for p in y.rglob('*') if p.is_file()}
    assert xs and xs == ys, (left, right)
    return len(xs)


assert sha(parent_path) == 'ba00445ab34f4f28665122098e690b15fc52db318e36ec19d72b0898611f34cb'
parent = load(parent_path)
verify_sources(parent['sources'])
verify_sources(load('docs/internals/data/battle-flagship-approach.json')['sources'])
verify_sources(load('docs/internals/data/battle-winner-player-search.json')['originals'])
verify_sources(load('docs/internals/data/planet-eligibility.json')['implementation_sources'])
if not a.write and not a.check_only:
    receipt = load(target)
    assert source(parent_path) == receipt['parent']
    verify_sources(receipt['sources'])
    print('第115輪來源保持', len(receipt['sources']), '父來源', len(parent['sources']))
    raise SystemExit

assert not target.exists()
helpers = equal_trees('workplace/promotion79-r115-h', 'workplace/promotion79-r115-i')
bootstrap = equal_trees('workplace/promotion79-r115-replay-b-plan', 'workplace/promotion79-r115-replay-c-plan')
d = load('workplace/promotion79-r115-h/normal-0/observations.json')
assert d['state'] == 'StateCommandPhase' and d['save_error'] == '<nil>'
assert d['promotion_reports'] == 1 and d['victory_reports'] == d['capture_reports'] == 0
assert sum(v['kind'] == 'seed' for v in d['inputs']) == 1
assert [(v['after']['index'], v['after']['target'], v['after']['move']) for v in d['strategic_orders']] == [
    (7,9,0),(16,13,0),(7,11,0),(7,16,1)]
assert [(v['region'],v['battle_turn']) for v in d['battle_entries']] == [(16,1)]
c = d['promotions'][0]
assert (c['turn'],c['region'],c['battle_turn'],c['face'],c['wait']) == (98,16,1,23,180)
record = bytes.fromhex(c['admiral'])
before = Path('workplace/reinforce79-r114-replay-d-plan/player-save/ADMIRAL.DT0').read_bytes()[23*28:24*28]
assert before[6] == 4 and int.from_bytes(before[2:6],'little') == 2652
assert record[6] == 5 and int.from_bytes(record[2:6],'little') == 152
assert bytes.fromhex(c['tables']['admiral'])[23*28:24*28] == record
plan = load('workplace/promotion79-r115-replay-b-plan/plan.json')
assert len(plan['lifecycle_reports']) == 1
r = plan['lifecycle_reports'][0]
assert r['kind'] == 'promotion'
assert all(r[k] == c[k] for k in ['tables','seed','face','text_hex','region','battle_turn','wait'])
rank_actions = [v for v in plan['actions'] if v.get('snapshot') and v.get('after',{}).get('state') == 'StateFleetWindow']
assert rank_actions
rank = rank_actions[-1]['snapshot']
assert (Path('workplace/promotion79-r115-h/normal-0/rank.bin').read_bytes() ==
        Path('workplace/promotion79-r115-replay-b-plan',rank+'.bin').read_bytes())
assert (Path('workplace/promotion79-r115-h/normal-0/promotion-01.bin').read_bytes() ==
        Path('workplace/promotion79-r115-replay-b-plan/lifecycle-report-01.bin').read_bytes())
gui_path = Path('workplace/promotion79-r115-replay-b-gui')
gui = load(gui_path/'receipt.json')
assert gui['test_tps'] == 60 and gui['natural_promotion'] is True
assert gui['natural_finish'] is False and gui['natural_annihilation'] is False
assert gui['battle_planet_capture'] is False and gui['versus_loaded'] is True
assert gui['six_diagnostic_saves_equal'] and gui['public_save_load_equal']
assert all(v['x11_different'] == 0 and v['game_index_equal'] for v in gui['comparisons'])
assert [(v['kind'],v['seed']) for v in gui['actions'] if v['kind']=='seed'] == [('seed',0x12345678)]
assert any(v['kind']=='click' and v['x']==300 and v['y']==122 and
           v['before']['state']=='StateLoadGame' and v['after']['state']=='StateLoadFaction'
           for v in gui['actions'])
assert any(v['kind']=='wait_report' and v['state']['lifecycle_report']['kind']=='promotion' for v in gui['actions'])
assert any(v['name']==rank for v in gui['comparisons'])
updates = 0
with (gui_path/'updates.jsonl').open() as f:
    for line in f:
        assert json.loads(line)['tps'] == 60
        updates += 1
assert updates > 0
for name in ['FLEET','SHIP','ADMIRAL','WORK','PLANET','STELLAR']:
    f = name+'.DT0'
    helper = Path('workplace/promotion79-r115-h/normal-0')/f
    base = Path('workplace/promotion79-r115-replay-b-plan')
    assert helper.read_bytes() == (base/f).read_bytes() == (base/'player-save'/f).read_bytes() == (gui_path/'player-save'/f).read_bytes()
    assert (base/'loaded-save'/f).read_bytes() == (gui_path/f).read_bytes() == helper.read_bytes()
final = Path('workplace/promotion79-r115-h/normal-0/ADMIRAL.DT0').read_bytes()[23*28:24*28]
assert final[6] == 5 and int.from_bytes(final[2:6],'little') == 1128
standard = Path('workplace/round115-standard.log').read_text()
assert sum(v.startswith('ok  ') for v in standard.splitlines()) == 13 and 'FAIL' not in standard
if a.check_only:
    print('第115輪核對通過',helpers,bootstrap,len(gui['actions']),len(gui['comparisons']),updates)
    raise SystemExit
paths = [Path('tools/scripts/battle_natural_promotion_'+v+'.py') for v in ['probe','replay','receipt']]
paths += [Path('docs/goal/2026-10-06-battle-natural-promotion-round-115.md'), Path('docs/playtest/103-battle-natural-promotion.md')]
paths += [v for v in Path('workplace').glob('promotion79-r115*') if v.is_file()]
paths += [q for v in Path('workplace').glob('promotion79-r115*') if v.is_dir() for q in v.rglob('*') if q.is_file()]
paths += [v for v in Path('workplace').glob('round115-*.log') if v.is_file()]
receipt = {
    'scope':'remake normal saved-game natural promotion; original growth oracle retained; no new original whole-turn parity',
    'head':'501ae687ad2b6973b117a99e5865244e4d277d6d', 'worktree_dirty_at_test':True,
    'image':'logh3-ebiten:2', 'image_id':'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
    'oracle_head_unchanged':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
    'scenario':'public S0 save from normal S1 game; Empire; two human seats',
    'seed_first_tactical_execute':0x12345678, 'seed_set_count':1, 'state_injection':False,
    'cross_save_rng_continuity_claimed':False, 'test_tps':60, 'parent':source(parent_path),
    'helper_files_equal':helpers, 'bootstrap_files_equal':bootstrap, 'rank_snapshot':rank,
    'promotion':{k:c[k] for k in ['turn','region','battle_turn','face','wait','seed','admiral']},
    'level_before_after':[4,5], 'merit_before_after_growth':[2652,152], 'merit_final':1128,
    'rank_changed':False, 'gui_actions':len(gui['actions']), 'gui_frames':len(gui['comparisons']),
    'gui_pixels':sum(v['pixels'] for v in gui['comparisons']), 'gui_updates_tps60':updates,
    'gui_camera_gestures':len(gui['camera_gestures']),
    'gui_last_input_seconds':max(v.get('at_seconds',0) for v in gui['actions']),
    'original_whole_turn_verified':False, 'spec062':'DRAFT', 'issue79':'OPEN',
    'sources':{str(v):source(v) for v in sorted(set(paths))}}
with target.open('x') as f:
    f.write(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
print('第115輪核對及凍結',len(receipt['sources']),source(target))
