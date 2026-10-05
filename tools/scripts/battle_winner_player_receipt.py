"""#79第112輪來源核對；限定目前remake回歸，不升格勝後占領。"""
import argparse
import ast
import hashlib
import json
import os
from pathlib import Path
import re

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('--check', action='store_true')
a = p.parse_args()
def info(path):
    b = Path(path).read_bytes()
    return {'bytes':len(b),'sha256':hashlib.sha256(b).hexdigest()}
def read(path):
    return json.loads(Path(path).read_text())
parent_path = 'docs/internals/data/planet-eligibility.json'
parent = read(parent_path)
original_parent = read('docs/internals/data/battle-occupation-player.json')
for path, want in original_parent['originals'].items():
    assert info(path) == want, path
bundle = original_parent['original_bundle']
assert info(bundle['path'])['sha256'] == bundle['sha256']
for group in ['sources','implementation_sources']:
    for path, want in parent[group].items():
        assert info(path) == want, path
base = Path('workplace/winner79-r112-replay-a-plan')
other = Path('workplace/winner79-r112-replay-b-plan')
files = sorted(x.relative_to(base) for x in base.rglob('*') if x.is_file())
assert files == sorted(x.relative_to(other) for x in other.rglob('*') if x.is_file())
assert all((base/x).read_bytes() == (other/x).read_bytes() for x in files)
plan = read(base/'plan.json')
reports = plan['lifecycle_reports']
assert [(x['kind'],x['battle_turn']) for x in reports] == [('annihilation',9),('victory',9)]
probe = read('workplace/winner79-r112-latea/normal-1/observations.json')
assert probe['capture_reports'] == 0 and probe['victory_reports'] == 1
assert reports[1]['tables'] == probe['observations'][0]['tables']
assert reports[1]['seed'] == probe['observations'][0]['seed']
gui_dir = Path('workplace/winner79-r112-replay-a-gui')
gui = read(gui_dir/'receipt.json')
assert gui['test_tps'] == 60 and gui['public_save_load_equal']
assert not gui['battle_planet_capture'] and not gui['ownership_report']
assert all(x['pixels']==307200 and x['x11_different']==0 and x['game_index_equal'] for x in gui['comparisons'])
updates = 0
for line in (gui_dir/'updates.jsonl').open():
    assert json.loads(line)['tps'] == 60
    updates += 1
load_diff = {}
for path in sorted((base/'player-save').glob('*.DT0')):
    assert path.read_bytes() == (base/path.name).read_bytes()
    assert path.read_bytes() == (gui_dir/'player-save'/path.name).read_bytes()
    loaded = (base/'loaded-save'/path.name).read_bytes()
    assert loaded == (gui_dir/path.name).read_bytes()
    offsets = [i for i,(x,y) in enumerate(zip(path.read_bytes(),loaded)) if x!=y]
    assert len(path.read_bytes()) == len(loaded)
    if path.name != 'STELLAR.DT0':
        assert not offsets, path.name
    load_diff[path.name] = offsets
logs = Path('workplace/round112-standard.log').read_text()
assert '\nFAIL' not in logs and 'panic:' not in logs
passed = re.findall(r'^ok\s+github.com/wicanr2/logh3/',logs,re.M)
untested = re.findall(r'^\?\s+github.com/wicanr2/logh3/',logs,re.M)
assert len(passed)==13 and len(untested)==5
cases = []
for tag in ['reservea','s1a','s3beam','s2beam','s2flag','s2flagb','s3wait','s3pursue','latea']:
    d = read('workplace/winner79-r112-'+tag+'/normal-1/observations.json')
    cases.append({'revision':tag,'scenario':d['scenario'],'end_state':d['state'],
                  'victory_reports':d['victory_reports'],'capture_reports':d['capture_reports'],
                  'reports':[{k:o[k] for k in ['kind','battle_turn','region','seed']} for o in d['observations'] or []],
                  'classification':'generator failed before flagships applied; not that experiment' if tag=='s2flag' else 'normal Game helper exploration; bootstrap coverage only latea'})
sources = set()
for pattern in ['workplace/winner79-r112*','workplace/round112-*.log']:
    for path in Path('.').glob(pattern):
        if path.is_dir():
            sources.update(str(x) for x in path.rglob('*') if x.is_file())
        elif path.is_file():
            sources.add(str(path))
sources.update(['docs/goal/2026-10-06-winner-occupation-player-round-112.md',
                'docs/playtest/100-battle-winner-player-search.md',
                'docs/spec/110-battle-finish.md',parent_path,
                'docs/internals/data/battle-occupation-player.json',
                'docs/internals/data/battle-lifecycle-player.json',
                'tools/scripts/battle_occupation_player_probe.py',
                'tools/scripts/battle_occupation_player_replay.py'])
for path in Path('tools/scripts').glob('battle_winner_player_*.py'):
    ast.parse(path.read_text())
    sources.add(str(path))
result = {'schema':1,'issue':79,'date':'2026-10-06',
          'remake_parent_commit':'d84dd871f41b13cb777ae7fc3bc38a0e80f354c5',
          'classification':'remake normal-player regression; no new original same-state claim',
          'original_program_sha256':parent['original_program_sha256'],
          'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
          'originals':original_parent['originals'],'original_bundle':bundle,
          'winner_occupation_complete':False,'production_changed':False,'cases':cases,
          'bootstrap':{'identical_files_each':len(files),'reports':[x['kind'] for x in reports],
                       'victory_seed':reports[1]['seed'],'save_load_different_offsets':load_diff},
          'gui':{'actions':len(gui['actions']),'full_frames':len(gui['comparisons']),
                 'pixels':sum(x['pixels'] for x in gui['comparisons']),
                 'different':0,'updates':updates,'tps':60,'camera_gestures':len(gui['camera_gestures'])},
          'standard':{'passed':len(passed),'no_tests':len(untested)},
          'parent_validation':{'sources':len(parent['sources']),'implementation_sources':len(parent['implementation_sources'])},
          'tools':{'image':'logh3-ebiten:2',
                   'image_id':'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
                   'go':'go1.24.13','python':'3.11','network':'none','uid_gid':'1000:1000'},
          'sources':{x:info(x) for x in sorted(sources)}}
output = Path('docs/internals/data/battle-winner-player-search.json')
encoded = (json.dumps(result,ensure_ascii=False,indent=2)+'\n').encode()
if a.check:
    assert output.read_bytes() == encoded, '本輪來源或結果已改變'
else:
    with output.open('xb') as f:
        f.write(encoded)
print('sources',len(sources),'bootstrap',len(files),'GUI',result['gui'],'index',info(output))
