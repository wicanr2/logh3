"""第110輪唯一來源索引；候選與正式Game分開，歷史來源按父提交核對。"""
import argparse
import hashlib
import json
import os
from pathlib import Path

assert Path.cwd() == Path('/src') and os.getuid() == 1000
parser = argparse.ArgumentParser()
mode = parser.add_mutually_exclusive_group(required=True)
mode.add_argument('--freeze', action='store_true')
mode.add_argument('--check', action='store_true')
args = parser.parse_args()
target = Path('docs/internals/data/landing-caller.json')


def digest(path):
    b = Path(path).read_bytes()
    return dict(bytes=len(b), sha256=hashlib.sha256(b).hexdigest())


parentpath = Path('docs/internals/data/landing-frame.json')
assert digest(parentpath)['sha256'] == 'a6f8f5fc53543285f8cbd9c24442a6730788875c0541d821ec2dcdde706d2ff8'
parent = json.loads(parentpath.read_text())
historical = {'docs/spec/118-landing-scene.md': 'workplace/round110-parent-spec118.md',
              'internal/screen/engagement.go': 'workplace/round110-parent-screen-engagement.go'}
for name, identity in parent['sources'].items():
    assert digest(historical.get(name,name)) == identity, name
for name, identity in parent['originals'].items():
    assert digest('/orig/'+name.removeprefix('org_game/extracted/')) == identity, name
assert digest(parent['original_bundle']['path'])['sha256'] == parent['original_bundle']['sha256']

typedpath = Path('workplace/landing48-r110-typed-r3.json')
typed = json.loads(typedpath.read_text())
assert typed['Schema'] == 'logh3.landing-caller-probe.v1' and len(typed['Cases']) == 297
for name, identity in typed['Sources'].items():
    assert digest(name) == identity, name
palette = json.loads(Path('docs/internals/data/fleet-presentation-spec.json').read_text())['palette']
rgb = [[(x<<2)|(x>>4) for x in colour] for colour in palette]
comparisons = {}
for revision in ('r1','r2','r3'):
    directory = Path('workplace/landing48-r110-probe-'+revision)
    stats = json.loads((directory/'comparison.json').read_text())
    assert len(stats) == 297 and sum(s['pixels'] for s in stats) == 91238400
    if revision == 'r1':
        assert sum(s['different'] > 0 for s in stats) == 185
    else:
        for c,s in zip(typed['Cases'],stats,strict=True):
            assert s['case'] == c['Name'] and s['different'] == s['max_channel_difference'] == 0
            assert s['difference_bbox'] is None and len(s['regions']) == 8
            assert all(r['different'] == 0 and r['same'] == r['pixels'] for r in s['regions'])
            assert Path(c['NativePath']).read_bytes() == (directory/(c['Name']+'-remake.bin')).read_bytes()
            p = Path(c['NativePath'].removesuffix('-screen.bin')+'-palette.json')
            assert json.loads(p.read_text())[:16] == rgb
    comparisons[revision] = stats

candidatepath = Path('workplace/landing48-r110-panel-candidate.json')
assert digest(candidatepath)['sha256'] == '1b647169e5a010eabb8f3cdefdac8c3ae8899798ba263476eb3426d0ddbd6580'
candidate = json.loads(candidatepath.read_text())
without = dict(candidate)
assert without.pop('planet_frame')
assert without == json.loads(Path('docs/internals/data/engagement-panel.json').read_text())
exportpath = Path('workplace/ida/landing48-r110-348A.json')
export = json.loads(exportpath.read_text())
assert export['kernel_version'] == '9.4' and export['uid'] == 1000
assert digest(exportpath.with_suffix('.com'))['sha256'] == export['input_sha256']
ram = Path('workplace/oracle/landing48-r108-r5-failure-a/frame-20b4-078f-01-mem.bin').read_bytes()
for row in export['instructions']:
    a = int(row['runtime_linear_hex'],16)
    raw = bytes.fromhex(row['original_bytes_hex'])
    assert ram[a:a+len(raw)] == raw
rows = {row['original_location']:row['original_bytes_hex'] for row in export['instructions']}
assert rows['348A:0273'] == 'e80b00' and rows['348A:0276'] == 'e84100'
assert rows['348A:027A'] == 'e80400' and rows['348A:027D'] == 'e8ec00'
standard = Path('workplace/round110-standard-tests.log').read_text()
assert standard.count('ok  \t') == 13 and standard.count('[no test files]') == 5 and 'FAIL' not in standard
assert '狀態：READY。' in Path('docs/spec/118-landing-scene.md').read_text()
eligibility = []
for run in ('a','b'):
    p = Path(f'workplace/planet-eligibility-r110-baseline-{run}.json')
    r = json.loads(p.read_text())
    assert r['sha256'] == '54a8750dd56d3aab85ae5da90642b4c59a24fe37987b5b11fd2f932b1006e7de'
    assert len(r['fleets']) == 7 and r['differences'] == [dict(action=0,current_bits=0,current_target=-1,fleet=31,original_bits=2,original_target=0)]
    assert digest(r['source'])['sha256'] == r['sha256']
    eligibility.append(r)
a,b = (dict(r) for r in eligibility)
a.pop('source');b.pop('source');assert a == b

if args.freeze:
    assert not target.exists()
    paths = set(typed['Sources'])
    for directory in ('internal','cmd'):
        paths.update(str(p) for p in Path(directory).rglob('*.go'))
    for revision in ('r1','r2','r3'):
        paths.update(str(p) for p in Path('workplace/landing48-r110-probe-'+revision).iterdir())
        paths.add('workplace/landing48-r110-typed-'+revision+'.json')
    paths.update([str(parentpath),str(typedpath),str(candidatepath),str(exportpath),str(exportpath.with_suffix('.com')),
        *historical.values(), 'docs/goal/2026-10-06-landing-caller-round-110.md',
        'docs/re/438-landing-caller.md','docs/playtest/098-landing-caller.md','docs/spec/118-landing-scene.md',
        'docs/internals/data/engagement-panel.json','docs/internals/data/fleet-presentation-spec.json',
        'tools/scripts/landing_caller_prepare.py','tools/scripts/landing_caller_probe.go',
        'tools/scripts/landing_caller_ida.py','tools/scripts/landing_caller_formats.py',
        'tools/scripts/landing_caller_candidate.go','tools/scripts/landing_caller_receipt.py',
        'tools/scripts/planet_eligibility_probe.go',
        'workplace/planet-eligibility-r110-baseline-a.json','workplace/planet-eligibility-r110-baseline-b.json',
        *[r['source'] for r in eligibility],
        'workplace/round110-standard-tests.log','go.mod','go.sum','tools/gui.sh','tools/ida.sh'])
    result = dict(schema='logh3.landing-caller.v1', date='2026-10-06',issue=48,parent_commit='246da42',
        parent_sources_verified=len(parent['sources']),historical_sources=historical,
        classification='same-state complete candidate renderer diagnostic;118 READY;not Game player path',
        typed=typed,comparisons=comparisons,candidate_panel=candidate,ida=export,
        eligibility_diagnostic=eligibility,
        gui_image=parent['gui_image'],ida_image=parent['ida_image'],go_image=parent['go_image'],
        sources={name:digest(name) for name in sorted(paths)},originals=parent['originals'],
        original_bundle=parent['original_bundle'],
        limits=['正式資料生成被自動審核拒絕，使用者確認待答；候選未接Game',
                '瞬時規則快照、正式分段、正常GUI與公開存讀仍缺',
                '不修Targets eligibility；不外推要塞、完整原版正常回合或音訊／人耳'])
    with target.open('x') as f:
        json.dump(result,f,ensure_ascii=False,indent=2)
        f.write('\n')
else:
    result = json.loads(target.read_text())
    assert result['typed'] == typed and result['comparisons'] == comparisons
    assert result['candidate_panel'] == candidate and result['ida'] == export
    assert result['eligibility_diagnostic'] == eligibility
    for name,identity in result['sources'].items():
        assert digest(name) == identity,name
print(json.dumps(dict(sources=len(result['sources']),originals=len(result['originals']),frames=297,
                      pixels=91238400,index=digest(target)),ensure_ascii=False))
