"""凍結第109輪完整單幀；舊來源按歷史118核實，不能覆寫凍結索引。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import struct

assert Path.cwd() == Path('/src') and os.getuid() == 1000
parser = argparse.ArgumentParser()
mode = parser.add_mutually_exclusive_group(required=True)
mode.add_argument('--freeze', action='store_true')
mode.add_argument('--check', action='store_true')
args = parser.parse_args()
target = Path('docs/internals/data/landing-frame.json')


def digest(path):
    raw = Path(path).read_bytes()
    return {'bytes': len(raw), 'sha256': hashlib.sha256(raw).hexdigest()}


parentpath = Path('docs/internals/data/landing-scene.json')
assert digest(parentpath)['sha256'] == '55908e574235e713b15c6d36dab79bec58dafe95a87b71423905248956196a0f'
parent = json.loads(parentpath.read_text())
for name, identity in parent['sources'].items():
    path = 'workplace/round109-parent-spec118.md' if name == 'docs/spec/118-landing-scene.md' else name
    assert digest(path) == identity, name
for name, identity in parent['originals'].items():
    assert digest('/orig/'+name.removeprefix('org_game/extracted/')) == identity, name
assert digest(parent['original_bundle']['path'])['sha256'] == parent['original_bundle']['sha256']

typedpath = Path('workplace/landing48-r109-typed-r4.json')
typed = json.loads(typedpath.read_text())
assert typed['Schema'] == 'logh3.landing-frame-probe.v1' and len(typed['Cases']) == 18
for name, identity in typed['Sources'].items():
    assert digest(name) == identity, name
stats = json.loads(Path('workplace/landing48-r109-probe-r4/comparison.json').read_text())
assert len(stats) == 18 and sum(s['pixels'] for s in stats) == 5529600
palette = json.loads(Path('docs/internals/data/fleet-presentation-spec.json').read_text())['palette']
rgb = [[(x << 2) | (x >> 4) for x in color] for color in palette]
for c, s in zip(typed['Cases'], stats, strict=True):
    assert c['Name'] == s['case'] and s['pixels'] == 307200
    assert s['different'] == s['max_channel_difference'] == 0 and s['difference_bbox'] is None
    assert len(s['regions']) == 8 and all(r['different'] == 0 for r in s['regions'])
    assert Path(c['NativePath']).read_bytes() == Path('workplace/landing48-r109-probe-r4',c['Name']+'-remake.bin').read_bytes()
    palpath = Path(c['NativePath'].removesuffix('-screen.bin')+'-palette.json')
    assert json.loads(palpath.read_text())[:16] == rgb

expected = {'counterfirst-r1': (249,59,122,1252923480,3,0,5590,[17,3,22,3]),
            'empire-r1': (249,59,122,1252923509,3,1,5482,[22,3,17,4]),
            'multiple-r1': (249,59,122,1252923522,3,0,5482,[22,3,17,3]),
            'multiple-r2': (377,91,188,1317924331,5,0,5405,[22,3,17,3,22,3])}
branches = {}
paths = set(typed['Sources'])
for key, (files,frames,events,end,waits,owner,defence,sounds) in expected.items():
    kind, revision = key.split('-')
    a = Path(f'workplace/oracle/landing48-r109-branch-{revision}-{kind}-a')
    b = Path(f'workplace/oracle/landing48-r109-branch-{revision}-{kind}-b')
    names = sorted(x.name for x in a.iterdir())
    assert names == sorted(x.name for x in b.iterdir()) and len(names) == files
    for name in names:
        paths.update([str(a/name),str(b/name)])
        if name != 'report.json':
            assert (a/name).read_bytes() == (b/name).read_bytes(), (key,name)
    r, rr = json.loads((a/'report.json').read_text()),json.loads((b/'report.json').read_text())
    suffix = ' r2' if revision == 'r2' else ''
    assert r.pop('command') == f'bash tools/landing_frame_branch_oracle.sh {kind} a{suffix}'
    assert rr.pop('command') == f'bash tools/landing_frame_branch_oracle.sh {kind} b{suffix}'
    assert r == rr
    assert r['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c' and not r['dosgolem_dirty']
    assert r['uid'] == 1000 and r['go_version'] == 'go1.24.13'
    assert len(r['frames']) == frames and len(r['events']) == events and r['end_step'] == end
    assert r['start_step'] == 1138999999 and r['end_original_location'] == '181B:03F0'
    assert r['irq0_every'] == 20000 and r['irq0_pinned'] and not r['cycle_clock']
    assert r['canvas'] == [640,480] and r['seed_once']
    assert len([x for x in r['injections'] if x['address'] == 'lin:5E00C']) == 1
    e = lambda loc: [x for x in r['events'] if x['original_location'] == loc]
    assert len(e('3765:006C')) == waits and all(x['stack_words'][2] == 90 for x in e('3765:006C'))
    assert [x['stack_words'][2] for x in e('3B74:031A')] == sounds
    assert len(e('1DD6:0152')) == len(e('1DD6:026B')) == 1
    endmem = (a/'end-mem.bin').read_bytes()
    assert endmem[0x54571] == owner and struct.unpack_from('<H',endmem,0x54574)[0] == defence
    calls = [x['original_location'] for x in r['events'] if x['original_location'] in ('20B4:0681','20B4:08B1')]
    if kind == 'counterfirst':
        assert calls == ['20B4:08B1','20B4:0681']
    elif revision == 'r2':
        assert calls == ['20B4:0681','20B4:08B1']*2
        assert [x['stack_words'][2] for x in e('20B4:0503')] == [53050,52928]
    else:
        assert calls == ['20B4:0681','20B4:08B1']
    branches[key] = dict(raw_files_per_run=files-1,frame_count=frames,event_count=events,
                         end_step=end,wait_count=waits,owner=owner,defence=defence,sound_ids=sounds,
                         classification='跳過第二攻方的診斷' if key == 'multiple-r1' else '明示注入分支診斷',
                         normalized_report=r)

d = Path('workplace/oracle/landing48-r109-branch-r2-multiple-a')
lower = lambda raw: b''.join(raw[y*640+8:y*640+632] for y in range(208,432))
previous = (d/'frame-3765-006c-03-screen.bin').read_bytes()
nextactor = (d/'frame-20b4-078f-02-screen.bin').read_bytes()
assert lower(previous) == lower(nextactor) and sum(x != y for x,y in zip(previous,nextactor)) == 6881
assert hashlib.sha256(lower(previous)).hexdigest() == '769482473bd9b7266d766febe4d4947d608de6cf83d1fd93102801127d46c0fe'

ram = Path('workplace/oracle/landing48-r108-r5-failure-a/frame-20b4-078f-01-mem.bin').read_bytes()
ida = []
for tag in ('20B4','321F','driver'):
    path = Path(f'workplace/ida/landing48-r109-{tag}.json')
    export = json.loads(path.read_text())
    assert export['kernel_version'] == '9.4' and export['uid'] == 1000
    assert digest(path.with_suffix('.com'))['sha256'] == export['input_sha256']
    for row in export['instructions']:
        a = int(row['runtime_linear_hex'],16)
        raw = bytes.fromhex(row['original_bytes_hex'])
        assert ram[a:a+len(raw)] == raw
    paths.update([str(path),str(path.with_suffix('.com'))])
    ida.append(export)
old = Path('workplace/ida/scene48-r69-3958.json')
export = json.loads(old.read_text())
for row in export['instructions']:
    a = int(row['runtime_linear_hex'],16)
    raw = bytes.fromhex(row['original_bytes_hex'])
    assert ram[a:a+len(raw)] == raw
paths.add(str(old))
assert sum(len(j['instructions']) for j in ida) == 1890
standard = Path('workplace/round109-standard-tests.log').read_text()
assert standard.count('ok  \t') == 13 and standard.count('[no test files]') == 5 and 'FAIL' not in standard

if args.freeze:
    assert not target.exists()
    for directory in ('internal','cmd'):
        paths.update(str(p) for p in Path(directory).rglob('*.go'))
    for directory in ('workplace/landing48-r109-probe-r1','workplace/landing48-r109-probe-r2',
                      'workplace/landing48-r109-probe-r3','workplace/landing48-r109-probe-r4'):
        paths.update(str(p) for p in Path(directory).iterdir())
    paths.update(['workplace/landing48-r109-typed.json','workplace/landing48-r109-typed-r2.json',
                  'workplace/landing48-r109-typed-r3.json',str(typedpath),str(parentpath),
                  'workplace/round109-parent-spec118.md','workplace/round109-standard-tests.log',
                  'workplace/ida/landing48-r109-minimum2.json',
                  'docs/goal/2026-10-06-landing-frame-round-109.md','docs/re/437-landing-frame.md',
                  'docs/playtest/097-landing-frame.md','docs/spec/118-landing-scene.md',
                  'docs/internals/37-ani-pnl-a98.md','go.mod','go.sum','tools/go.sh','tools/gui.sh','tools/py.sh','tools/ida.sh',
                  'tools/landing_frame_branch_oracle.sh','tools/landing_frame_branch_native_run.sh',
                  'tools/scripts/landing_frame_branch_prepare.py','tools/scripts/landing_frame_native.go',
                  'tools/scripts/landing_frame_native_r2.go','tools/scripts/landing_frame_ida.py',
                  'tools/scripts/landing_frame_prepare.py','tools/scripts/landing_frame_probe.go',
                  'tools/scripts/landing_frame_receipt.py','docs/internals/data/engagement-panel.json',
                  'docs/internals/data/fleet-presentation-spec.json'])
    result = dict(schema='logh3.landing-frame.v1',date='2026-10-06',issue=48,
                  parent_commit='e9608737e19ec3fadf5d660307e32f2e9430e06c',
                  parent_sources_verified=len(parent['sources']), parent_spec_verified_by_commit='e960873',
                  classification='same-state complete typed diagnostic;118 DRAFT;not Game player path',
                  go_image=parent['go_image'],ida_image=parent['ida_image'],
                  gui_image=dict(tag='logh3-ebiten:2',id='sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8'),
                  typed=typed,comparisons=stats,branches=branches,ida=ida,
                  sources={p:digest(p) for p in sorted(paths)},originals=parent['originals'],
                  original_bundle=parent['original_bundle'],
                  limits=['完整反擊caller與跨攻方不同上／下方快照待補','118 DRAFT，正式Game未改',
                          '正常GUI／公開保存與原版整回合同態未新增','原版音效選號不代表人耳或硬體時序驗收'])
    with target.open('x') as f:
        json.dump(result,f,ensure_ascii=False,indent=2)
        f.write('\n')
else:
    result = json.loads(target.read_text())
    assert result['typed'] == typed and result['comparisons'] == stats and result['branches'] == branches and result['ida'] == ida
    for name, identity in result['sources'].items():
        assert digest(name) == identity, name
print(json.dumps(dict(sources=len(result['sources']),originals=len(result['originals']),frames=len(stats),
                      pixels=sum(s['pixels'] for s in stats),index=digest(target)),ensure_ascii=False))
