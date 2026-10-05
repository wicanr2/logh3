"""第106輪：正常公開存讀、退出修正、完整檔案與歷史來源核對。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import struct
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
parser = argparse.ArgumentParser()
parser.add_argument('--freeze', action='store_true')
parser.add_argument('--check', action='store_true')
args = parser.parse_args()
assert args.freeze != args.check
target = Path('docs/internals/data/windows-save-load.json')
root = Path('workplace/music69-windows')
stems = ['STELLAR', 'PLANET', 'ADMIRAL', 'FLEET', 'SHIP', 'WORK']
parent = '91d85785b937145f060066cf9bb29edc643a7c05'


def sha(path):
    with Path(path).open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def identity(path):
    return {'bytes': Path(path).stat().st_size, 'sha256': sha(path)}


def read_json(path):
    return json.loads(Path(path).read_text())


def owned(path):
    assert Path(path).stat().st_uid == os.getuid() and Path(path).stat().st_gid == os.getgid(), path


def inspect():
    build = root / 'saveload69-r106-build'
    builds = read_json(build / 'BUILD.json')
    for row in builds['binaries']:
        assert sha(build / row['file']) == row['sha256']
    assert builds['binaries'][0]['sha256'] == builds['binaries'][1]['sha256'] == 'fa2995765c7ee2ed5d458e46f7b3536a93a1f82b408130320d6f5b2c9cab29fb'
    assert builds['binaries'][2]['sha256'] == builds['binaries'][3]['sha256'] == 'abab7e58f88b0f37dc6a7e69bf326426ec95c94aec4fe8f899a2781621292343'
    for folder in ['cmd', 'internal']:
        for p in (build / 'source' / folder).rglob('*'):
            if p.is_file(): assert p.read_bytes() == Path(p.relative_to(build / 'source')).read_bytes(), p
    for n in ['go.mod', 'go.sum']:
        assert (build / 'source' / n).read_bytes() == Path(n).read_bytes()
    assert '正式 app 未結束 <nil>' in Path('workplace/round106-quit-before.log').read_text()
    assert 'FAIL' not in Path('workplace/round106-quit-after.log').read_text()
    lines = Path('workplace/round106-standard-product.log').read_text().splitlines()
    assert sum(l.split()[:1] == ['ok'] for l in lines) == 13
    assert sum('[no test files]' in l for l in lines) == 5 and not any(l.startswith('FAIL') for l in lines)
    references = {}
    for rev in ['refc', 'refd']:
        folder = root / ('saveload69-r106-' + rev)
        ref = read_json(folder / 'REFERENCES.json')
        assert len(ref['runs']) == 15 and ref['script_diagnostic']
        assert ref['binary_sha256'] == builds['binaries'][0]['sha256']
        for row in ref['runs']:
            assert sha(folder / (row['name'] + '.png')) == row['png_sha256']
            assert sha(folder / (row['name'] + '.bin')) == row['index_sha256']
        references[rev] = ref
    a, b = root / 'saveload69-r106-refc', root / 'saveload69-r106-refd'
    for p in a.rglob('*'):
        if p.is_file() and (p.suffix in ['.bin', '.png'] or '.DT' in p.name or p.name == 'SAVEDATA.IDX'):
            assert p.read_bytes() == (b / p.relative_to(a)).read_bytes(), p
    runtime, frame_count, action_count = {}, 0, 0
    for rev in ['b', 'c', 'linuxb']:
        folder = root / ('saveload69-r106-' + rev)
        check = read_json(folder / 'CHECK.json')
        assert not check['script_used'] and not check['state_injection'] and check['fresh_process_load']
        assert [p['phase'] for p in check['phases']] == ['new', 'load']
        for phase in check['phases']:
            assert phase['exit'] == 0 and phase['public_quit']
            assert phase['actions'] == (59 if phase['phase'] == 'new' else 41)
            assert len(phase['frames']) == (9 if phase['phase'] == 'new' else 7)
            trace = read_json(folder / phase['phase'] / 'TRACE.json')
            assert len(trace['actions']) == phase['actions'] and trace['frames'] == phase['frames']
            start = read_json(folder / phase['phase'] / 'START.json')
            command = start['command']
            assert start['default_intro'] and not start['script_used'] and not start['state_injection']
            assert '-script' not in command and '-shot' not in command and '-shot-bin' not in command
            assert start['platform'] == ('linux/amd64' if rev == 'linuxb' else 'windows/amd64')
            if rev != 'linuxb':
                assert start['working_directory'] == '/player'
                assert start['readonly_inputs'] == ['/player', '/orig', '/data']
            for frame in phase['frames']:
                assert frame['exit'] == 0 and frame['different_pixels'] == '0'
                png = Path(frame['path'].replace('/out/', '/src/workplace/music69-windows/'))
                assert png.read_bytes()[:8] == b'\x89PNG\r\n\x1a\n'
                assert struct.unpack('>II', png.read_bytes()[16:24]) == (640, 480)
            for f in phase['files']:
                assert f['different'] == 0
                if f['file'] == 'SAVEDATA.IDX' and phase['phase'] == 'new':
                    # 第二次公開保存會更新同一IDX；CHECK保留首次實際讀取的雜湊。
                    initial_index = a / '保存 空間-saved' / 'SAVEDATA.IDX'
                    assert f['sha256'] == sha(initial_index) and f['bytes'] == initial_index.stat().st_size
                else:
                    assert sha(folder / '玩家 保存 空間' / f['file']) == f['sha256']
            frame_count += len(phase['frames']); action_count += phase['actions']
        save = folder / '玩家 保存 空間'
        differences = {}
        for stem in stems:
            x, y = (save / (stem + '.DT0')).read_bytes(), (save / (stem + '.DT1')).read_bytes()
            assert len(x) == len(y)
            differences[stem] = [[i, m, n] for i, (m, n) in enumerate(zip(x, y)) if m != n]
            assert x == (a / '保存 空間-saved' / (stem + '.DT0')).read_bytes()
            assert y == (a / '保存 空間-resave' / (stem + '.DT1')).read_bytes()
        assert all(not differences[s] for s in stems if s != 'WORK')
        assert differences['WORK'] == [[14, 0, 1], [16, 0, 1], [18, 0, 1], [20, 0, 31]]
        fleet = (save / 'FLEET.DT0').read_bytes()
        assert len(fleet) == 3904
        assert [i for i in range(32) if fleet[i * 122 + 4] != 255] == [0]
        assert fleet[1] == 0 and fleet[4] == 0
        runtime[rev] = {'gui': check, 'roundtrip_differences': differences,
                        'six_file_bytes_per_save': 44560, 'active_fleet_pool_ids': [0]}
    assert frame_count == 48 and action_count == 300
    assert not (root / 'saveload69-r106-a/CHECK.json').exists()
    failed = read_json(root / 'saveload69-r106-a/new/TRACE.json')
    assert len(failed['frames']) == 9 and len(failed['actions']) == 59
    history = {}
    for name in ['current-source', 'windows-player', 'windows-music-playback']:
        file = Path('docs/internals/data') / (name + '.json')
        old = read_json(file)
        changed = []
        for path, expected in old['sources'].items():
            if identity(path) != expected:
                assert path == 'cmd/gin3/main.go', path
                original = subprocess.check_output(['git', 'show', parent + ':' + path])
                assert len(original) == expected['bytes'] and hashlib.sha256(original).hexdigest() == expected['sha256']
                changed.append({'path': path, 'historical_commit': parent, 'historical': expected, 'current': identity(path)})
        history[name] = {'index_sha256': sha(file), 'sources': len(old['sources']), 'historical_changes': changed}
    parent_index = read_json('docs/internals/data/windows-player.json')
    for path, expected in parent_index['original_files'].items():
        assert identity(Path('org_game/extracted') / path) == expected
    assert sha(parent_index['original_bundle']['path']) == parent_index['original_bundle']['sha256']
    images = Path('workplace/round106-images.txt').read_text().splitlines()
    assert images == ['sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
                      'sha256:7d3c7ff554db39300e3e7a92a7b79c2d5fcd647289596d22fa10899baac61201']
    return {'schema': 'logh3.windows-save-load.v1', 'date': '2026-10-05', 'issues': [69, 90],
            'parent_commit': parent, 'build': builds, 'images': images, 'references': references,
            'runtime': runtime, 'pixels': {'frames': frame_count, 'pixels': frame_count * 307200,
            'different': 0, 'max_channel_difference': 0, 'bbox': None, 'mask': None},
            'actions': action_count, 'original_bundle': parent_index['original_bundle'],
            'original_files': parent_index['original_files'], 'history': history,
            'before_fix': {'revision': 'a', 'public_save_matches_linux': True, 'public_quit': 'FAIL timeout 10 seconds'},
            'additional_linuxa': '同一GUI流程通過；CHECK.scope誤寫Windows，START.platform為Linux；保留原批次，修正文字後正式重跑linuxb',
            'standard': {'packages': 18, 'passed': 13, 'no_test': 5},
            'limitations': {'physical_windows': False, 'new_original_same_state': False,
                            'whole_music_playback': False, 'human_listening': False,
                            'macos': False, 'formal_release': False,
                            'prior_source_kit_is_current_player': False}}


if args.freeze:
    assert not target.exists()
    value = inspect()
    files = set()
    for folder in ['cmd', 'internal']:
        files.update(p for p in Path(folder).rglob('*') if p.is_file())
    for n in ['go.mod', 'go.sum', 'tools/windows_save_load.sh', 'tools/scripts/windows_save_load.py',
              'tools/scripts/windows_save_load_receipt.py', 'tools/scripts/fleet_player_gui_capture.go',
              'docs/goal/2026-10-05-windows-save-load-round-106.md', 'docs/spec/117-player-public-quit.md',
              'docs/playtest/094-windows-save-load.md', 'docs/re/341-fleet-number-list-and-quit.md',
              'docs/internals/data/current-source.json', 'docs/internals/data/windows-player.json',
              'docs/internals/data/windows-music-playback.json', 'workplace/round106-build.log',
              'workplace/round106-images.txt', 'workplace/round106-quit-before.log',
              'workplace/round106-quit-after.log', 'workplace/round106-standard-gui.log',
              'workplace/round106-standard-product.log']:
        files.add(Path(n))
    files.update(p for p in Path('docs/internals/data').glob('*.json') if p != target)
    for rev in ['a', 'b', 'c', 'linuxa', 'linuxb', 'refa', 'refb', 'refc', 'refd']:
        files.update(p for p in (root / ('saveload69-r106-' + rev)).rglob('*') if p.is_file())
    build = root / 'saveload69-r106-build'
    files.update(p for p in build.iterdir() if p.is_file())
    for folder in ['cmd', 'internal']:
        files.update(p for p in (build / 'source' / folder).rglob('*') if p.is_file())
    for n in ['go.mod', 'go.sum']:
        files.add(build / 'source' / n)
    value['sources'] = {str(p): identity(p) for p in sorted(files)}
    for p in files: owned(p)
    target.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')
    print('FROZEN', len(files), 'sources', sha(target))
else:
    value = read_json(target)
    current = inspect()
    assert {k: v for k, v in value.items() if k != 'sources'} == current
    for path, expected in value['sources'].items():
        assert identity(path) == expected, path
        owned(path)
    print('PASS', len(value['sources']), 'sources / 164 originals / 48 full frames / 300 X11 commands')
