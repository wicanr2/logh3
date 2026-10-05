"""第108輪原版登陸逐格收據核對器；在tools/gui.sh容器內執行。"""
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
target = Path('docs/internals/data/landing-scene.json')


def digest(path):
    b = Path(path).read_bytes()
    return {'bytes': len(b), 'sha256': hashlib.sha256(b).hexdigest()}


def folder(kind, side):
    rev = 'r6' if kind == 'click' else 'r5'
    return Path(f'workplace/oracle/landing48-r108-{rev}-{kind}-{side}')


def report(kind, side='a'):
    return json.loads((folder(kind, side) / 'report.json').read_text())


def events(data, location):
    return [x for x in data['events'] if x['original_location'] == location
            and not x.get('callback_resume', False)]


ds, planet = 0x54b30, 0x54560
expected = {
    'baseline': (121, 27, 20, 59, 1187924209, 0, 3300),
    'failure': (249, 59, 39, 122, 1252923503, 0, 5482),
    'capture': (149, 34, 18, 76, 1202723021, 1, 0),
    'off': (101, 22, 2, 73, 1197123834, 0, 5482),
    'click': (249, 59, 39, 126, 1199269708, 0, 5482),
}
cases = {}
for kind, (files, frames, draws, event_count, end, owner, defence) in expected.items():
    a, b = folder(kind, 'a'), folder(kind, 'b')
    names = sorted(x.name for x in a.iterdir())
    assert names == sorted(x.name for x in b.iterdir()) and len(names) == files
    for name in names:
        if name != 'report.json':
            assert (a / name).read_bytes() == (b / name).read_bytes(), (kind, name)
    r, rr = report(kind), report(kind, 'b')
    assert r.pop('command') == f'bash tools/landing_scene_oracle.sh {kind} a'
    assert rr.pop('command') == f'bash tools/landing_scene_oracle.sh {kind} b'
    assert r == rr
    assert r['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
    assert not r['dosgolem_dirty'] and r['uid'] == 1000
    assert r['start_step'] == 1138999999 and r['end_step'] == end
    assert r['end_original_location'] == '181B:03F0'
    assert r['irq0_every'] == 20000 and r['irq0_pinned'] and not r['cycle_clock']
    assert r['canvas'] == [640, 480] and r['seed_once']
    assert len(r['frames']) == frames and len(r['events']) == event_count
    assert len(events(r, '348A:0281')) == draws
    assert sum(x['name'].startswith('draw-frame-return') for x in r['frames']) == draws
    for f in r['frames']:
        raw = (a / (f['name'] + '-screen.bin')).read_bytes()
        png = (a / (f['name'] + '-screen.png')).read_bytes()
        assert len(raw) == 307200 and png[:8] == b'\x89PNG\r\n\x1a\n'
        assert struct.unpack_from('>II', png, 16) == (640, 480)
        assert len((a / (f['name'] + '-mem.bin')).read_bytes()) == 1048576
    seeds = [x for x in r['injections'] if x['address'] == 'lin:5E00C']
    assert len(seeds) == 1 and seeds[0]['bytes_hex'] == '78563412'
    assert seeds[0]['step'] == 1140714675
    mem = (a / 'end-mem.bin').read_bytes()
    assert mem[planet + 17] == owner
    assert struct.unpack_from('<H', mem, planet + 20)[0] == defence
    calls = events(r, '3958:0041')
    sounds = [x['stack_words'][2] for x in events(r, '3B74:031A')]
    waits = events(r, '3765:006C')
    assert all(x['stack_words'][2] == 90 for x in waits)
    if kind in ('failure', 'capture', 'off', 'click'):
        assert len(events(r, '20B4:0711')) == 1
        local = events(r, '20B4:0711')[0]['stack_words']
        assert local[2] == 431 and local[4] == 844
        assert local[1] == (1 if kind == 'capture' else 518)
        assert events(r, '20B4:121F')[0]['stack_words'][2] == (kind == 'capture')
        assert len(events(r, '20B4:08B1')) == (kind != 'capture')
        assert len(waits) == (2 if kind == 'capture' else 3)
        assert calls[0]['comani_record_hex'] == '01080000ff3716'
        assert calls[0]['stack_words'][4:6] == [0, 0]
        assert calls[2]['comani_record_hex'] == '01080000ff3716'
        assert calls[2]['stack_words'][4:6] == [0, 1]
        assert calls[3]['comani_record_hex'] == ('00090102ffff02' if kind == 'capture' else '00090100ffff03')
    if kind == 'off':
        assert not sounds and not events(r, '348A:0228')
    elif kind == 'capture':
        assert sounds == [22, 2]
        assert struct.unpack_from('<II', mem, planet + 60) == (20000, 0)
        assert mem[planet + 32:planet + 60] == bytes(28)
    elif kind in ('failure', 'click'):
        assert sounds == [22, 3, 17, 3]
    cases[kind] = {'raw_files_per_run': files - 1, 'normalized_reports_equal': True,
                   'frame_count': frames, 'drawn_frame_count': draws,
                   'event_count': event_count, 'end_step': end,
                   'owner': owner, 'defence': defence,
                   'final_seed_hex': f'{struct.unpack_from("<I", mem, ds+0x94dc)[0]:08X}',
                   'sound_ids': sounds, 'wait_steps': [x['step'] for x in waits],
                   'comani_calls': calls, 'inputs': r['inputs'], 'injections': r['injections']}

base = folder('failure', 'a')
for kind in ('off', 'click'):
    a = (base / 'end-mem.bin').read_bytes()
    b = (folder(kind, 'a') / 'end-mem.bin').read_bytes()
    ship = struct.unpack_from('<H', a, ds+0xcfb4)[0] + 16 * struct.unpack_from('<H', a, ds+0xcfb6)[0]
    for start, size in [(ds+0xc074, 3904), (ship, 27456), (planet, 246), (ds+0x94dc, 4)]:
        assert a[start:start+size] == b[start:start+size], (kind, hex(start))
for x, y in zip(report('failure')['frames'], report('click')['frames'], strict=True):
    assert x['name'] == y['name']
    for suffix in ('-screen.bin', '-palette.json', '-screen.png'):
        assert (base / (x['name'] + suffix)).read_bytes() == (folder('click', 'a') / (y['name'] + suffix)).read_bytes()

standard = Path('workplace/round108-standard-tests.log').read_text()
assert standard.count('ok  \t') == 13 and standard.count('[no test files]') == 5 and 'FAIL' not in standard
ida = json.loads(Path('workplace/ida/landing48-r108-driver.json').read_text())
assert ida['kernel_version'] == '9.4' and ida['uid'] == 1000
assert len(ida['instructions']) == 924

if args.freeze:
    assert not target.exists()
    paths = {'go.mod', 'go.sum', 'tools/go.sh', 'tools/gui.sh', 'tools/py.sh', 'tools/ida.sh',
             'tools/landing_scene_oracle.sh', 'tools/landing_scene_native_run.sh',
             'tools/scripts/landing_scene_native.go', 'tools/scripts/landing_scene_ida.py',
             'tools/scripts/landing_scene_receipt.py',
             'docs/goal/2026-10-05-landing-scene-round-108.md',
             'docs/re/436-landing-scene.md', 'docs/spec/118-landing-scene.md',
             'docs/playtest/096-landing-scene.md',
             'docs/internals/data/battle-occupation-player.json',
             'docs/internals/data/engagement-gauges.json',
             'workplace/ida/landing48-r108-minimum.json',
             'workplace/ida/landing48-r108-20B4.com',
             'workplace/ida/landing48-r108-driver.com',
             'workplace/ida/landing48-r108-driver.json',
             'workplace/round108-standard-tests.log', 'workplace/round108-previous-source-check.log'}
    for directory in ('internal', 'cmd'):
        paths.update(str(p) for p in Path(directory).rglob('*.go'))
    for kind in expected:
        for side in 'ab':
            paths.update(str(p) for p in folder(kind, side).iterdir())
    previous = json.loads(Path('docs/internals/data/battle-occupation-player.json').read_text())
    result = {'schema': 'logh3.landing-scene.v1', 'date': '2026-10-05', 'issue': 48,
              'parent_commit': '9b034afcf0a25a5e228445913eac421ae2268c8f',
              'go_image': {'tag': 'golang:1.24-bookworm', 'id': 'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac'},
              'ida_image': {'tag': 'ida-pro-9.4-idapython:locked-v1', 'id': 'sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780'},
              'classification': '完整原版第六步保存續行、明示分支注入；未重算eligibility；沒有remake same-state',
              'cases': cases, 'click_equal_frames': 59, 'click_equal_pixels': 59 * 307200,
              'off_and_click_equal_fleet_ship_planet_bytes': 3904 + 27456 + 246,
              'ida': ida,
              'sources': {name: digest(name) for name in sorted(paths)},
              'originals': previous['originals'], 'original_bundle': previous['original_bundle'],
              'limitations': ['118 DRAFT', '未改正式Go或AI', '帝國／先反擊／多人名冊／要塞尚缺',
                              '完整typed單幀、Game正常GUI及公開保存未新增驗收']}
else:
    result = json.loads(target.read_text())
    assert result['cases'] == cases and result['ida'] == ida
    for name, identity in result['sources'].items():
        assert digest(name) == identity, name
for name, identity in result['originals'].items():
    assert name.startswith('org_game/extracted/')
    assert digest(Path('/orig') / name.removeprefix('org_game/extracted/')) == identity, name
assert digest(result['original_bundle']['path'])['sha256'] == result['original_bundle']['sha256']
if args.freeze:
    with target.open('x') as f:
        json.dump(result, f, ensure_ascii=False, indent=2)
        f.write('\n')
print(json.dumps({'sources': len(result['sources']), 'originals': len(result['originals']),
                  'cases': {k: {kk: v[kk] for kk in ('raw_files_per_run', 'frame_count', 'drawn_frame_count', 'end_step')}
                            for k, v in cases.items()},
                  'index': digest(target)}, ensure_ascii=False))
