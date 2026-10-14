"""第132輪：核對原版雙跑、297幀與兩陣營音色實錄，保存可追溯索引。"""
import argparse
import hashlib
import json
import os
from pathlib import Path

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--out', default='docs/internals/data/landing-soundfont.json')
a = p.parse_args()
assert Path.cwd() == Path('/src') and os.getuid() == 1000
target = Path(a.out)
assert not target.exists() and target.parent.stat().st_uid == os.getuid()

def identity(path):
    path = Path(path)
    with path.open('rb') as f:
        sha = hashlib.file_digest(f, 'sha256').hexdigest()
    return {'bytes': path.stat().st_size, 'sha256': sha}

def read(path):
    return json.loads(Path(path).read_text())

native = []
for kind in ('baseline', 'failure', 'capture', 'counterfirst', 'empire', 'multiple'):
    prefix = ('landing48-r108-r5-' if kind in ('baseline', 'failure', 'capture')
              else 'landing48-r109-branch-' + ('r2-' if kind == 'multiple' else 'r1-'))
    dirs = [Path('workplace/oracle')/(prefix+kind+'-'+side) for side in ('a', 'b')]
    reports = [read(d/'report.json') for d in dirs]
    files = sorted(p.name for p in dirs[0].iterdir() if p.suffix == '.bin' or p.name.endswith('-palette.json'))
    assert files and files == sorted(p.name for p in dirs[1].iterdir() if p.suffix == '.bin' or p.name.endswith('-palette.json'))
    identities = {}
    for name in files:
        x, y = identity(dirs[0]/name), identity(dirs[1]/name)
        assert x == y, (kind, name)
        identities[name] = x
    for r in reports:
        assert r['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
        assert not r['dosgolem_dirty'] and r['seed_once'] and r['irq0_every'] == 20000
    native.append({'case': kind, 'equal_files': len(files), 'files': identities,
                   'runs': [{'path': str(d), 'report_identity': identity(d/'report.json'),
                             'metadata': {k: v for k, v in r.items() if k not in ('events', 'frames')}}
                            for d, r in zip(dirs, reports)]})

comparison = Path('workplace/round132/landing-comparison/comparison.json')
stats = read(comparison)
assert len(stats) == 297 and sum(s['pixels'] for s in stats) == 91238400
assert all(s['different'] == s['max_channel_difference'] == 0 and s['difference_bbox'] is None
           and len(s['regions']) == 8 and all(r['different'] == 0 for r in s['regions']) for s in stats)

gui = []
for faction, directory in enumerate(('gui-f0-final', 'gui-f1-r3')):
    d = Path('workplace/round132')/directory
    r = read(d/'RESULT.json')
    assert r['passed'] and r['faction'] == faction and r['normal_entry']
    assert not any(r[k] for k in ('script_flag', 'state_injection', 'seed_injection'))
    audio = [{k: e[k] for k in ('audio', 'peak', 'rms', 'pcm_sha256')}
             | {'library': e['observed']['music_library']} for e in r['events'] if 'audio' in e]
    assert len(audio) == 2 and {x['audio'] for x in audio} == {'gs', 'mt32'}
    assert all(x['peak'] > 32 and x['rms'] > 4 for x in audio)
    gui.append({'faction': faction, 'path': str(d), 'result': identity(d/'RESULT.json'),
                'elapsed_seconds': r['elapsed_seconds'], 'audio': audio,
                'checks': [e['check'] for e in r['events'] if 'check' in e],
                'screenshots': [e for e in r['events'] if 'screenshot' in e]})

mt32 = read('workplace/round132/mt32/SOURCES.json')
assert len(mt32['tracks']) == 11
assert all(-90 < t['rms_db'] and t['peak_db'] < -.2 for t in mt32['tracks'])
for t in mt32['tracks']:
    assert identity('workplace/round132/music/mt32/'+t['name']+'.ogg')['sha256'] == t['ogg_sha256']

sources = [
    *[str(p) for directory in ('cmd/gin3', 'internal/game', 'internal/rules', 'internal/screen')
      for p in Path(directory).glob('*.go')],
    'docs/internals/data/engagement-panel.json', 'docs/internals/data/fleet-presentation-spec.json',
    'tools/scripts/landing_soundfont_gui.py', 'tools/scripts/landing_soundfont_receipt.py',
    'tools/scripts/mt32_prepare.py', 'tools/scripts/landing_caller_formats.py',
    'tools/scripts/landing_caller_candidate.go', 'tools/Dockerfile.audio', 'tools/audio.sh',
    'tools/gui.sh', 'tools/with-xvfb.sh', 'go.mod', 'go.sum',
]
result = {'schema': 'logh3.landing-soundfont.v1', 'date': '2026-10-14',
          'soundfont_gui': gui, 'mt32': mt32, 'native_pairs': native,
          'renderer': {'classification': 'same-state complete renderer diagnostic; not Game player path',
                       'frames': 297, 'pixels': 91238400, 'different': 0,
                       'comparison': {'path': str(comparison), **identity(comparison)}},
          'sources': {s: identity(s) for s in sorted(set(sources))},
          'binary': identity('workplace/round132/gin3'),
          'original_bundle': identity('org_game/银河英雄传说3SP.zip'),
          'originals': {n: identity('/orig/'+n) for n in
                        ('GIN3PS.EXE', 'ANI_PNL1.A98', 'COMANI.BMG', '3DBLOCK.SPR', 'COM_FLAG.SPR', 'GRAPH.IMG')},
          'standard_tests': {'command': 'tools/gui.sh test ./cmd/... ./internal/...',
                             'passed_packages': 15, 'no_test_packages': 4},
          'limits': ['#48 remains open: fortress-gun scene, current-policy normal GUI/save and full DOS same-state integration missing',
                     'legacy round107 helper explicitly uses passive CPU; its capture probe is diagnostic and cannot prove current default player path',
                     'MT-32 synthesis/channel routing is an approximation; no real hardware waveform or new human listening claim',
                     'this is a development build; published v.1.3.0 artifacts unchanged']}
target.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
print(json.dumps({'index': str(target), **identity(target), 'native_pairs': len(native),
                  'native_equal_files': sum(x['equal_files'] for x in native), 'renderer_frames': 297,
                  'gui_factions': len(gui)}, ensure_ascii=False))
