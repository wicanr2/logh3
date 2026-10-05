"""第113輪來源核對；編組完成與勝後占領探索分開驗證。"""
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

def read(path):
    return json.loads(Path(path).read_text())

def info(path):
    b = Path(path).read_bytes()
    return {'bytes':len(b),'sha256':hashlib.sha256(b).hexdigest()}

parent_path = 'docs/internals/data/battle-winner-player-search.json'
parent = read(parent_path)
assert info(parent_path)['sha256'] == '66cc520e0e8afc73a34940d0ad4c7d64031ee765053b9c58fae0ea8b095450e8'
for path,want in parent['sources'].items():
    assert info(path) == want, path
for path,want in parent['originals'].items():
    assert info(path) == want, path
bundle = parent['original_bundle']
assert info(bundle['path'])['sha256'] == bundle['sha256']
implementation = read('docs/internals/data/planet-eligibility.json')['implementation_sources']
for path,want in implementation.items():
    assert info(path) == want, path

m = read('workplace/winner79-r113-m/normal-0/observations.json')
n = read('workplace/winner79-r113-n/normal-0/observations.json')
assert m['inputs'] == n['inputs'] and m['observations'] == n['observations']
assert [x['seed'] for x in n['inputs'] if x['kind']=='seed'] == [0x12345678]
assert [(x['region'],x['battle_turn'],x['kind']) for x in n['observations']] == [(13,4,'victory')]*2
assert n['capture_reports'] == 0 and len(n['occupation_results']) == 2
core = []
for x in n['occupation_results']:
    assert x['event'] is None and x['done'] and x['result']['Hits']==3
    assert x['result']['Captured'] is None
    b = bytes.fromhex(x['tables']['planet'])
    start = sum(int.from_bytes(b[i:i+2],'little') for i in range(0,26,2))
    assert start == 33
    rec = b[64+start*82:64+(start+1)*82]
    assert rec[17] == 1
    core.append({'region':x['region'],'battle_turn':x['battle_turn'],'hits':x['result']['Hits'],
                 'owner':rec[17],'defence':int.from_bytes(rec[20:22],'little'),'seed':x['seed']})
assert [x['defence'] for x in core] == [10473,11103]

composition = read('workplace/winner79-r113-p/normal-1/observations.json')
steps = composition['composition_steps']
assert steps[0]['fleet']['groups']==7 and steps[-1]['fleet']['groups']==4
assert [x['slot'] for x in steps if 'slot' in x] == [28,29,38]
assert steps[0]['region_stock']==[0,0,4000,0,3000,0,0]
assert steps[-1]['region_stock']==[0,0,12000,0,3000,0,1000]
assert steps[0]['fleet']['composition'][0]['pool']==135
assert steps[-1]['fleet']['composition'][0]['pool']==135
inputs = read('workplace/winner79-r113-p/normal-1/inputs.json')['inputs']
assert len(inputs)==15 and all(x['at_ns']==0 and x['kind'] in ['click','right'] for x in inputs)

base = Path('workplace/flagship79-r113-replay-b-plan')
other = Path('workplace/flagship79-r113-replay-c-plan')
files = sorted(x.relative_to(base) for x in base.rglob('*') if x.is_file())
assert files == sorted(x.relative_to(other) for x in other.rglob('*') if x.is_file())
assert all((base/x).read_bytes() == (other/x).read_bytes() for x in files)
plan = read(base/'plan.json')
assert not plan['lifecycle_reports'] and plan['final']['state']=='StateCommandPhase'
gui_dir = Path('workplace/flagship79-r113-replay-c-gui')
gui = read(gui_dir/'receipt.json')
assert gui['test_tps']==60 and gui['public_save_load_equal']
assert not gui['natural_finish'] and not gui['natural_annihilation'] and not gui['battle_planet_capture']
assert all(x['pixels']==307200 and x['x11_different']==0 and x['game_index_equal'] for x in gui['comparisons'])
updates = 0
for line in (gui_dir/'updates.jsonl').open():
    assert json.loads(line)['tps']==60
    updates += 1
save_bytes = 0
for path in sorted((base/'player-save').glob('*.DT0')):
    data = path.read_bytes()
    assert data == (base/path.name).read_bytes()
    assert data == (base/'loaded-save'/path.name).read_bytes()
    assert data == (gui_dir/'player-save'/path.name).read_bytes()
    assert data == (gui_dir/path.name).read_bytes()
    assert data == (Path('workplace/winner79-r113-p/normal-1')/path.name).read_bytes()
    save_bytes += len(data)
assert save_bytes == 44560
logs = Path('workplace/round113-standard.log').read_text()
assert '\nFAIL' not in logs and 'panic:' not in logs
passed = re.findall(r'^ok\s+github.com/wicanr2/logh3/',logs,re.M)
untested = re.findall(r'^\?\s+github.com/wicanr2/logh3/',logs,re.M)
assert len(passed)==13 and len(untested)==5

cases = []
for tag in ['b','c','d','e','f','g','h','i','j','k','l','m','n','p']:
    faction = 0 if tag in ['k','l','m','n'] else 1
    d = read(f'workplace/winner79-r113-{tag}/normal-{faction}/observations.json')
    seed_count = sum(x['kind']=='seed' for x in d['inputs'])
    assert seed_count==(0 if tag=='p' else 1)
    cases.append({'revision':tag,'scenario':d['scenario'],'faction':faction,'state':d['state'],
                  'weapon':d['weapon_preference'],'standoff_pixels':d.get('standoff_pixels',0),
                  'victory_reports':d['victory_reports'],'capture_reports':d['capture_reports'],
                  'reports':[{k:x[k] for k in ['region','kind','battle_turn','seed']} for x in d['observations'] or []],
                  'executed_test_command':f'bash tools/gui.sh xrun bash -c \'LOGH3_FLEET_CHAIN_OUTPUT=/src/workplace/winner79-r113-{tag} go test -overlay=workplace/winner79-r113-{tag}-overlay.json ./internal/game -run "^TestBattleFinishNormalProbe$" -count=1 -v > workplace/round113-{tag}.log 2>&1\'',
                  'classification':'composition normal bootstrap and GUI verified' if tag=='p' else
                  'versus harness selection failure; no battle completion claim' if tag in ['h','i'] else
                  'normal Game helper exploration; no occupation completion'})
sources = set()
for pattern in ['workplace/winner79-r113*','workplace/flagship79-r113*','workplace/round113-*.log']:
    for path in Path('.').glob(pattern):
        if path.is_dir():
            sources.update(str(x) for x in path.rglob('*') if x.is_file())
        elif path.is_file():
            sources.add(str(path))
sources.update([parent_path,'docs/spec/015-fleet-composition.md','docs/spec/051-strategic-fleet-command.md',
                'docs/spec/110-battle-finish.md','docs/internals/data/star-graph.json',
                'docs/goal/2026-10-06-flagship-approach-round-113.md',
                'docs/playtest/101-battle-flagship-approach.md'])
for path in Path('tools/scripts').glob('battle_flagship_*.py'):
    ast.parse(path.read_text())
    sources.add(str(path))
result = {'schema':1,'issue':79,'date':'2026-10-06',
          'remake_parent_commit':'bbeb149e9a55417665a04cb6152e26774fe29092',
          'remake_worktree_dirty':True,
          'classification':'remake normal composition verification and occupation exploration; no new original same-state claim',
          'production_changed':False,'winner_occupation_complete':False,
          'original_program_sha256':parent['original_program_sha256'],
          'dosgolem_head':parent['dosgolem_head'],'original_bundle':bundle,
          'cases':cases,'winner_core_observations':core,
          'composition':{'before_groups':7,'after_groups':4,'removed_slots':[28,29,38],
                         'returned_ships':9000,'public_save_bytes':save_bytes,'save_load_different':0},
          'bootstrap':{'identical_files_each':len(files),'public_input_steps':15,'battle_reports':0},
          'gui':{'actions':len(gui['actions']),'full_frames':len(gui['comparisons']),
                 'pixels':sum(x['pixels'] for x in gui['comparisons']),'different':0,'updates':updates,'tps':60},
          'standard':{'passed':len(passed),'no_tests':len(untested)},
          'commands':[
              'bash tools/gui.sh xrun python3 tools/scripts/battle_flagship_composition_replay.py b',
              'bash tools/gui.sh xrun python3 tools/scripts/battle_flagship_composition_replay.py c',
              'bash tools/gui.sh xrun python3 tools/scripts/battle_flagship_composition_replay.py c --gui',
              "bash tools/gui.sh xrun bash -c 'go test ./internal/... ./cmd/... -count=1 > workplace/round113-standard.log 2>&1'"],
          'limits':['all occupation searches failed; helper exploration only',
                    'early victory is not sufficient to capture high-defence planet',
                    'versus selection assumptions unresolved; no product bug claim',
                    'no new original full-turn, audio, macOS, release or #48 completion'],
          'preserved_tool_failures':['revision a generator filename collision; no test execution',
                                     'composition replay a used wrong parent save directory; fixed in b and c',
                                     'GUI b inherited incorrect natural-finish flags; generator fixed and GUI c rerun'],
          'parent_validation':{'sources':len(parent['sources']),'originals':len(parent['originals']),
                               'implementation_sources':len(implementation)},
          'tools':parent['tools'],'sources':{x:info(x) for x in sorted(sources)}}
output = Path('docs/internals/data/battle-flagship-approach.json')
encoded = (json.dumps(result,ensure_ascii=False,indent=2)+'\n').encode()
if a.check:
    assert output.read_bytes()==encoded,'本輪來源或結果已改變'
else:
    with output.open('xb') as f:
        f.write(encoded)
print('sources',len(sources),'bootstrap',len(files),'GUI',result['gui'],'index',info(output))
