"""第116輪：自然死亡、引用搬移、下一回合GUI與公開存讀來源核對。"""
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
target = Path('docs/internals/data/battle-natural-cleanup.json')
parent_path = Path('docs/internals/data/battle-natural-promotion.json')


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda: f.read(1024*1024), b''):
            h.update(block)
    return h.hexdigest()


def source(path):
    st = path.stat()
    assert st.st_uid == st.st_gid == 1000, str(path)
    return {'sha256':sha(path),'bytes':st.st_size}


def load(path):
    return json.loads(Path(path).read_text())


def verify(sources):
    for name, expected in sources.items():
        assert source(Path(name)) == expected, name


def equal_trees(left, right):
    x,y = Path(left),Path(right)
    xs = {str(p.relative_to(x)):sha(p) for p in x.rglob('*') if p.is_file()}
    ys = {str(p.relative_to(y)):sha(p) for p in y.rglob('*') if p.is_file()}
    assert xs and xs == ys, (left,right)
    return len(xs)


assert sha(parent_path) == '0790049faea28eb077b82c05ce8cf3a947f65e62d6b9bad564762e972e2e0824'
parent = load(parent_path)
verify(parent['sources'])
verify(load('docs/internals/data/battle-reinforcement.json')['sources'])
verify(load('docs/internals/data/battle-flagship-approach.json')['sources'])
verify(load('docs/internals/data/battle-winner-player-search.json')['originals'])
verify(load('docs/internals/data/planet-eligibility.json')['implementation_sources'])
original_path = Path('docs/internals/data/battle-cleanup.json')
original = load(original_path)
assert len(original['cases']) == 35 and original['removed_count'] == 77
for item in original['sources']:
    path = Path(item['path'])
    assert path.stat().st_size == item['size'] and sha(path) == item['sha256'], str(path)
if not a.write and not a.check_only:
    receipt = load(target)
    assert receipt['parent'] == source(parent_path)
    assert receipt['original_cleanup'] == source(original_path)
    verify(receipt['sources'])
    print('第116輪來源保持',len(receipt['sources']),'父來源',len(parent['sources']))
    raise SystemExit

assert not target.exists()
helpers = equal_trees('workplace/cleanup79-r116-c','workplace/cleanup79-r116-d')
bootstrap = equal_trees('workplace/cleanup79-r116-replay-b-plan','workplace/cleanup79-r116-replay-c-plan')
d = load('workplace/cleanup79-r116-c/normal-0/observations.json')
assert d['state'] == 'StateCommandPhase' and d['save_error'] == '<nil>'
assert d['cleanup_events'] == 1 and d['promotion_reports'] == 1
assert d['victory_reports'] == d['capture_reports'] == 0
assert sum(v['kind']=='seed' for v in d['inputs']) == 1
before,after = d['cleanup_results']
assert before['key'] == after['key'] == '99:16:3:16'
assert before['stage']=='before' and after['stage']=='after'
assert before['seed'] == after['seed'] and after['mapping'] == [0,-1,1,2,3,4,5]
# 按109的已證實byte契約獨立驗算，禁止呼叫產品清理程式產生期望。
fleet,ship = bytearray.fromhex(before['tables']['fleet']),bytearray.fromhex(before['tables']['ship'])
rec = 16*122
count = fleet[rec+15]
assert count == 7
refs = [int.from_bytes(fleet[rec+16+4*i:rec+18+4*i],'little') for i in range(count)]
assert refs == [93,95,96,97,98,99,100]
assert int.from_bytes(ship[93*78+8:93*78+10],'little') > 0
assert int.from_bytes(ship[95*78+8:95*78+10],'little') == 0
fleet[rec+15] -= 1
fleet[rec+20:rec+40] = fleet[rec+24:rec+44]
ship[95*78+1] = 255
ship[95*78+0x19] &= ~0x40
ship[95*78+0x38:95*78+0x3c] = bytes(4)
assert fleet.hex() == after['tables']['fleet'] and ship.hex() == after['tables']['ship']
for name in ['admiral','work','stellar','planet']:
    assert before['tables'][name] == after['tables'][name], name
units = []
for u in before['units']:
    v = dict(u)
    if v['fleet'] == 16 and v['squadron'] >= 0:
        mapped = after['mapping'][v['squadron']]
        if mapped < 0:
            continue
        v['squadron'] = mapped
    units.append(v)
assert units == after['units']
plan_path = Path('workplace/cleanup79-r116-replay-b-plan')
events = load(plan_path/'cleanup-events.json')
assert events == [{k:v for k,v in row.items() if k!='at_ns'} for row in d['cleanup_results']]
next_frame = 'cleanup-next-input-01'
plan = load(plan_path/'plan.json')
assert any(v.get('snapshot') == next_frame and v['after']['battle_turn']==4 for v in plan['actions'])
# 助手observe在右鍵後才執行；正式入口在Tick後立即取圖。
# 用同一公開重播補取首個右鍵後的全圖，不遮罩或改寫原快照。
timing_path = Path('workplace/cleanup79-r116-replay-timingb-snapshot-plan')
for file in plan_path.rglob('*'):
    if file.is_file():
        assert file.read_bytes() == (timing_path/file.relative_to(plan_path)).read_bytes(), str(file)
helper_frame = Path('workplace/cleanup79-r116-c/normal-0',next_frame+'.bin').read_bytes()
formal_frame = (plan_path/(next_frame+'.bin')).read_bytes()
assert helper_frame == (timing_path/'cleanup-first-right.bin').read_bytes()
differences = [i for i,(x,y) in enumerate(zip(helper_frame,formal_frame)) if x!=y]
assert len(helper_frame)==len(formal_frame)==640*480 and len(differences)==5535
snapshot_at = next(i for i,v in enumerate(plan['actions']) if v.get('snapshot')==next_frame)
first_right = plan['actions'][snapshot_at+1]
assert first_right['kind']=='right' and [first_right['x'],first_right['y']]==[17,61]
assert first_right['before']['view']==[41,48] and first_right['after']['view']==[40,47]
gui_path = Path('workplace/cleanup79-r116-replay-c-gui')
gui = load(gui_path/'receipt.json')
assert gui['natural_nonflagship_cleanup'] and gui['natural_promotion'] and gui['test_tps'] == 60
assert gui['natural_finish'] is False and gui['natural_annihilation'] is False
assert gui['battle_planet_capture'] is False and gui['public_save_load_equal']
assert gui['six_diagnostic_saves_equal'] and gui['versus_loaded']
assert load(gui_path/'cleanup-events.json') == events
assert all(v['x11_different']==0 and v['game_index_equal'] for v in gui['comparisons'])
assert any(v['name'] == next_frame for v in gui['comparisons'])
cycles = [v for v in gui['actions'] if v['kind']=='right' and v['before']['stage']==1]
assert len(cycles) > 0
assert [(v['kind'],v['seed']) for v in gui['actions'] if v['kind']=='seed'] == [('seed',0x12345678)]
updates = 0
with (gui_path/'updates.jsonl').open() as f:
    for line in f:
        assert json.loads(line)['tps'] == 60
        updates += 1
assert updates > 0
for name in ['FLEET','SHIP','ADMIRAL','WORK','PLANET','STELLAR']:
    f = name+'.DT0'
    raw = Path('workplace/cleanup79-r116-c/normal-0',f).read_bytes()
    assert raw == (plan_path/f).read_bytes() == (plan_path/'player-save'/f).read_bytes()
    assert raw == (plan_path/'loaded-save'/f).read_bytes() == (gui_path/'player-save'/f).read_bytes() == (gui_path/f).read_bytes()
saved_fleet = (plan_path/'player-save/FLEET.DT0').read_bytes()
assert saved_fleet[rec+15] == 6
assert [int.from_bytes(saved_fleet[rec+16+i*4:rec+18+i*4],'little') for i in range(6)] == [93,96,97,98,99,100]
standard = Path('workplace/round116-standard.log').read_text()
assert sum(v.startswith('ok  ') for v in standard.splitlines()) == 13 and 'FAIL' not in standard
if a.check_only:
    print('第116輪核對通過',helpers,bootstrap,len(gui['actions']),len(gui['comparisons']),updates,len(cycles))
    raise SystemExit

paths = [Path('tools/scripts/battle_natural_cleanup_'+v+'.py') for v in ['probe','replay','receipt']]
paths += [Path('tools/scripts/battle_natural_cleanup_observer.go')]
paths += [Path('tools/scripts/sound_tail_input.go')]
paths += [Path('docs/goal/2026-10-06-battle-natural-cleanup-round-116.md'),Path('docs/playtest/104-battle-natural-cleanup.md')]
paths += [v for v in Path('workplace').glob('cleanup79-r116*') if v.is_file()]
paths += [q for v in Path('workplace').glob('cleanup79-r116*') if v.is_dir() for q in v.rglob('*') if q.is_file()]
paths += [v for v in Path('workplace').glob('round116-*.log') if v.is_file()]
receipt = {'scope':'normal remake nonflagship death and middle-reference compaction; existing original35 retained; no new original whole-turn parity',
    'head':'e0db0df50673b1972d238573604fa41ab3c14a7b','worktree_dirty_at_test':True,
    'image':'logh3-ebiten:2','image_id':'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
    'oracle_head_unchanged':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
    'parent':source(parent_path),'original_cleanup':source(original_path),
    'scenario':'public S0 save from normal S1 game; Empire; two human seats',
    'seed_first_tactical_execute':0x12345678,'seed_set_count':1,'state_injection':False,
    'natural_death':{'turn':99,'region':16,'battle_turn':3,'fleet':16,'pool':95,'old_squadron':1,'groups':[7,6],'mapping':after['mapping'],'seed':after['seed']},
    'helper_files_equal':helpers,'bootstrap_files_equal':bootstrap,'gui_actions':len(gui['actions']),
    'gui_frames':len(gui['comparisons']),'gui_pixels':sum(v['pixels'] for v in gui['comparisons']),
    'gui_updates_tps60':updates,'gui_target_cycle_gestures':len(cycles),'gui_camera_gestures':len(gui['camera_gestures']),
    'gui_last_input_seconds':max(v.get('at_seconds',0) for v in gui['actions']),
    'helper_next_frame':'after first public right scroll; full frame equals snapshot-timing replay',
    'formal_gui_next_frame':'before first public right scroll; complete GUI frame exact',
    'snapshot_timing_difference':{'pixels':len(differences),'before_view':[41,48],'after_view':[40,47]},
    'original_whole_turn_verified':False,'spec062':'DRAFT','issue79':'OPEN',
    'sources':{str(v):source(v) for v in sorted(set(paths))}}
with target.open('x') as f:
    f.write(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
print('第116輪核對及凍結',len(receipt['sources']),source(target))
