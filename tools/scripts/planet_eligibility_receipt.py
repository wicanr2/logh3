"""#91 原版向量與結算證據。所有預期值來自新執行的原版 RAM。"""
import argparse
import hashlib
import json
import os
import subprocess
from pathlib import Path

assert Path.cwd() == Path('/src') and os.getuid() == 1000
parser = argparse.ArgumentParser()
parser.add_argument('--prepare', action='store_true')
parser.add_argument('--freeze', action='store_true')
args = parser.parse_args()
ds = 0x54b30
def identity(p):
    b = Path(p).read_bytes()
    return dict(bytes=len(b), sha256=hashlib.sha256(b).hexdigest())
def decode(b):
    assert len(b) == 1048576
    def word(a): return int.from_bytes(b[a:a+2], 'little')
    def dword(a): return int.from_bytes(b[a:a+4], 'little')
    def far(a): return word(a+2)*16+word(a)
    ps, fs, us, targets, bits = [], [], [], [], []
    for i in range(b[ds+0xcffc]):
        p = far(ds+0xcff8)+i*82
        ps.append(dict(index=i, owner=b[p+0x11], x=word(p+0x46), y=word(p+0x48),
            defence=word(p+0x14), defence_max=word(p+0x16), fortress=b[p+0x12]!=0,
            gun_state=b[p+0x45], gun_aim=word(p+0x50),
            stock=dict(ships=[dword(p+0x20+k*4) for k in range(7)], missiles=dword(p+0x3c), fighters=dword(p+0x40))))
    for i in range(b[ds+0x9894]):
        n = b[ds+0x9898+i]
        f = ds+0xc074+n*122
        fs.append(dict(index=n, faction=b[f+1], number=b[f+2], flag0=b[f]&1!=0,
            action=b[f+0x48], x=word(f+0x5c), y=word(f+0x5e), facing=b[f+0x4a],
            integrity=b[f+0x3c], ability0=b[f+0x3d], ability3=b[f+0x40],
            defence=b[f+0x40]*b[f+0x41]//100,
            admiral_loss_factor=b[ds+0xb77c+b[f+4]*28+0x0b]))
        targets.append(-1 if b[f+0x5a]==255 else b[f+0x5a])
        bits.append(word(f+0x70))
        for k in range(b[f+0x0f]):
            si = word(f+0x10+k*4)
            s = far(ds+0xcfb4)+si*78
            us.append(dict(fleet=n, squadron=b[s+3], faction=b[s], kind=b[s+4],
                x=word(s+0x34), y=word(s+0x36), ships=word(s+8), full_ships=word(s+0xa),
                repair_cap=word(s+0xc), missiles=dword(s+0xe), fighters=dword(s+0x12)))
    return dict(planets=ps, fleets=fs, units=us, targets=targets, bits=bits, seed=dword(ds+0x94dc))

cases = []
sources = {}
for mode in ('matrix2', 'pipeline'):
    a = Path('workplace/oracle/planet91-r111-'+mode+'-a')
    b = Path('workplace/oracle/planet91-r111-'+mode+'-b')
    for p in sorted(a.iterdir()):
        assert p.is_file()
        peer = b/p.name
        if p.name != 'batch.json':
            assert p.read_bytes() == peer.read_bytes(), p
        sources[str(p)] = identity(p)
        sources[str(peer)] = identity(peer)
    batch = json.loads((a/'batch.json').read_text())
    assert batch['cases'] == (40 if mode == 'matrix2' else 1)
    if mode == 'matrix2':
        for p in sorted(a.glob('*-report.json')):
            r = json.loads(p.read_text())
            name = r['scenario']['Name']
            before = decode((a/(name+'-input-mem.bin')).read_bytes())
            after = decode((a/(name+'-end-mem.bin')).read_bytes())
            assert before['seed'] == after['seed'] == 0x12345678
            assert before['planets'] == after['planets'] and before['units'] == after['units']
            assert r['cpu_return']['IP'] == 65534 and r['end_step']-r['start_step'] <= 1000000
            assert len(before['fleets']) == len(r['targets'])
            assert after['targets'] == [-1 if x['target']==255 else x['target'] for x in r['targets']]
            assert after['bits'] == [x['bits'] for x in r['targets']]
            cases.append(dict(name=name, input=before, expected=after, report=str(p),
                input_ram=str(a/(name+'-input-mem.bin')), end_ram=str(a/(name+'-end-mem.bin'))))

a = Path('workplace/oracle/planet91-r111-pipeline-a')
before = decode((a/'phase6-input-mem.bin').read_bytes())
after = decode((a/'phase6-end-mem.bin').read_bytes())
calls = json.loads((a/'continuation-report.json').read_text())
assert before['seed'] == 0x12345678 and after['seed'] == calls['final_seed']
assert [c['original_location'] for c in calls['calls']] == ['20B4:0486','20B4:0503','20B4:0681','20B4:08B1']
assert before['planets'] == after['planets']
assert before['targets'] == [-1,-1,-1,-1,-1,-1,0] and before['bits'] == [0,0,0,0,0,0,2]
loss = sum(u['ships']-v['ships'] for u,v in zip(before['units'],after['units'],strict=True))
assert loss > 0
result = dict(schema='logh3.planet-eligibility.v1', issue=91, date='2026-10-06',
    remake_parent_commit='726f479b51857dd4989f19e1527f166fba099e6f', remake_worktree_dirty=True,
    classification='fresh original direct-entry Targets and counter-only continuation; not normal player full-turn same-state',
    original_program_sha256='770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918',
    parent_sha256='c6ce5af3a8d52acc415ac1029a29a0905e9c95e02ac0e0ba547e71300fd947c1',
    dosgolem_head='f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', dosgolem_dirty=False,
    cases=cases, pipeline=dict(input=before, expected=after, calls=calls, total_loss=loss),
    sources=sources)
ida_path = Path('workplace/ida/landing48-r108-driver.json')
ida = json.loads(ida_path.read_text())
raw = Path('workplace/oracle/planet91-r111-matrix2-a/baseline-parent-mem.bin').read_bytes()
rows = [r for r in ida['instructions'] if int(r['original_location'].split(':')[1],16) < 0x503]
for r in rows:
    address = int(r['runtime_linear_hex'],16)
    code = bytes.fromhex(r['original_bytes_hex'])
    assert raw[address:address+len(code)] == code
assert len(rows) == 485 and ida['kernel_version'] == '9.4'
result['ida_control_flow'] = dict(source=str(ida_path), identity=identity(ida_path), verified_instructions=len(rows),
    grade='proven for listed branches and observed matrix outputs',
    branches=[dict(r, semantic_grade='proven', evidence='40 fresh original cases and counter-only continuation')
              for r in rows if r['original_location'] in ['20B4:0153','20B4:015C','20B4:0169','20B4:016E','20B4:01A5','20B4:03D3','20B4:03D8','20B4:03E7']])
target = Path('docs/internals/data/planet-eligibility.json')
if not args.prepare:
    parent_path = Path('docs/internals/data/landing-caller.json')
    parent = json.loads(parent_path.read_text())
    allowed = {'internal/rules/planetassault.go', 'internal/game/battle.go'}
    historical = []
    for name, expected in parent['sources'].items():
        if identity(name) != expected:
            assert name in allowed, name
            data = subprocess.check_output(['git','show','726f479:'+name])
            assert dict(bytes=len(data), sha256=hashlib.sha256(data).hexdigest()) == expected
            historical.append(name)
    assert set(historical) == allowed
    for name, expected in parent['originals'].items():
        assert identity('/orig/'+name.removeprefix('org_game/extracted/')) == expected
    assert identity(parent['original_bundle']['path'])['sha256'] == parent['original_bundle']['sha256']
    result['parent_validation'] = dict(path=str(parent_path), identity=identity(parent_path),
        source_count=len(parent['sources']), original_count=len(parent['originals']),
        historical_commit='726f479', historical_sources=sorted(historical))
    paths = set(str(p) for directory in ('internal','cmd') for p in Path(directory).rglob('*.go'))
    paths.update(['docs/spec/119-planet-assault-eligibility.md','docs/spec/019-planet-assault.md',
        'docs/re/439-planet-eligibility.md','docs/playtest/099-planet-eligibility.md',
        'docs/goal/2026-10-06-planet-eligibility-round-111.md',
        'tools/planet_eligibility_oracle.sh','tools/planet_eligibility_native_run.sh',
        'tools/scripts/planet_eligibility_native.go','tools/scripts/planet_eligibility_receipt.py',
        'workplace/round111-standard-tests.log','workplace/round111-counter-game-tests.log',
        'workplace/round111-player-policy-tests.log','go.mod','go.sum','tools/gui.sh','tools/go.sh'])
    result['implementation_sources'] = {name:identity(name) for name in sorted(paths)}
    assert '狀態：CONFORMED。' in Path('docs/spec/119-planet-assault-eligibility.md').read_text()
    result['tools'] = dict(go_image='golang:1.24-bookworm',
        go_image_id='sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac',
        gui_image='logh3-ebiten:2',
        gui_image_id='sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
        go_version='go1.24.13', timer_contract='20000-step pinned IRQ0; CycleClock=false')
    standard = Path('workplace/round111-standard-tests.log').read_text()
    assert standard.count('ok  \t') == 13 and standard.count('[no test files]') == 5 and 'FAIL' not in standard
    player = Path('workplace/round111-player-policy-tests.log').read_text()
    assert 'FAIL' not in player and '\nPASS\n' in player
    assert all('--- PASS: '+name in player for name in ['TestBattleComputerDealsNoDamage','TestBattleShippedSavesNormalExecuteAndSave','TestPlanetCounterOnlyPassivePolicy','TestPlanetCounterOnlyAnimationModes'])
if args.prepare or args.freeze:
    if target.exists():
        assert json.loads(target.read_text())['schema'] == result['schema']
    with target.open('w') as f:
        json.dump(result,f,ensure_ascii=False,indent=2)
        f.write('\n')
else:
    assert json.loads(target.read_text()) == result
print(json.dumps(dict(cases=len(cases), raw_sources=len(sources), loss=loss, final_seed=after['seed'], index=identity(target)),ensure_ascii=False))
