"""第104輪Windows玩家GUI的來源及範圍核對。只在專案GUI Docker內執行。"""
import argparse
import ast
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
parser = argparse.ArgumentParser()
parser.add_argument('--freeze', action='store_true')
parser.add_argument('--check', action='store_true')
args = parser.parse_args()
assert args.freeze != args.check
target = Path('docs/internals/data/windows-player.json')


def sha(path):
    with Path(path).open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


if args.freeze:
    assert not target.exists()
    root = Path('workplace/music69-windows')
    pixels = json.loads((root / 'player69-r104-PIXELS4.json').read_text())
    assert pixels['different'] == 154407 and len(pixels['reports']) == 44
    assert pixels['total_pixels'] == 16204800
    exact = [r for r in pixels['reports'] if 'sampling' not in r]
    scaled = [r for r in pixels['reports'] if 'sampling' in r]
    assert len(exact) == 37 and len(scaled) == 7
    assert all(r['different'] == r['max_channel_difference'] == 0 and r['bbox'] is None for r in exact)
    assert all(r['max_channel_difference'] == 1 and not r['sampling']['full_window_exact'] for r in scaled)
    for r in scaled:
        s = r['sampling']
        assert s['left_outside_box_rounding_bounds'] == s['right_outside_box_rounding_bounds'] == 0
        assert s['differences_outside_tie_rows_or_columns'] == 0
    for report in pixels['reports']:
        for side in ['left', 'right']:
            assert sha(root / report[side]) == report[side + '_sha256']
    runtime = {}
    for name in ['load1b', 'load15b', 'fleet1']:
        folder = root / ('player69-r104-' + name)
        check = json.loads((folder / 'CHECK.json').read_text())
        start = json.loads((folder / 'START.json').read_text())
        assert check['exit'] == check['different'] == 0 and check['indexed_pixels'] == 307200
        assert not check['script_used'] and not check['state_injection']
        assert start['default_intro'] and not start['script_used']
        assert len(check['frames']) == (6 if name == 'fleet1' else 7)
        assert all(row['exit'] == 0 and row['different_pixels'] == '0' for row in check['frames'])
        assert start['binary_sha256'] == sha(root / 'source103-player/gin3.exe')
        assert (folder / 'final.bin').read_bytes() == (root / 'player69-r104-reference2' / (start['target']+'.bin')).read_bytes()
        runtime[name] = check
    titles = []
    for name in ['startup', 'repeat']:
        folder = root / ('player69-r104-' + name)
        start = json.loads((folder / 'START.json').read_text())
        process = json.loads((folder / 'PROCESS.json').read_text())
        assert process['exit'] == 0 and process['window'] and process['screenshots']
        assert start['default_intro'] and not start['script_used']
        assert start['binary_sha256'] == sha(root / 'source103-player/gin3.exe')
        state = (folder / 'title.txt').read_text()
        assert 'state=StateTitle\n' in state and 'updates=4998\n' in state and 'intro_frame=420\n' in state
        titles.append({'batch': name, 'state': state, 'process': process})
    assert (root / 'player69-r104-startup/title.bin').read_bytes() == (root / 'player69-r104-repeat/title.bin').read_bytes()
    assert (root / 'player69-r104-startup/title.png').read_bytes() == (root / 'player69-r104-repeat/title.png').read_bytes()
    tests = Path('workplace/round104-standard-gui.log').read_text().splitlines()
    assert sum(t.split()[:1] == ['ok'] for t in tests) == 13
    assert sum('[no test files]' in t for t in tests) == 5 and not any(t.startswith('FAIL') for t in tests)
    bundle = Path('org_game/银河英雄传说3SP.zip')
    assert sha(bundle) == 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673'
    paths = {'tools/player_wine.sh', 'tools/music_wine_entry.sh', 'tools/gui.sh',
             'tools/scripts/windows_player_check.py', 'tools/scripts/windows_player_pixels.py',
             'tools/scripts/windows_player_receipt.py', 'workplace/round104-standard-gui.log',
             'docs/goal/2026-10-05-windows-player-round-104.md', 'docs/playtest/092-windows-player.md',
             'docs/spec/116-current-source.md', 'docs/spec/112-title-player.md',
             'docs/internals/data/current-source.json', 'workplace/source69-r103-d/rebuild/REBUILD.json',
             'workplace/music69-windows/source103-player/gin3.exe',
             'workplace/source69-r103-d/rebuild/gin3-linux-source', 'go.mod', 'go.sum'}
    paths.add('tools/scripts/fleet_player_gui_capture.go')
    vendor = 'workplace/source69-r103-d/extracted/logh3/vendor/github.com/hajimehoshi/ebiten/v2/'
    paths.update(vendor+f for f in ['run.go', 'gameforui.go', 'internal/builtinshader/shader.go'])
    paths.update(str(p) for p in Path('internal').rglob('*.go'))
    paths.update(str(p) for p in Path('cmd').rglob('*.go'))
    paths.update(str(p) for p in Path('docs/internals/data').glob('*.json') if p != target)
    for p in root.glob('player69-r104-*'):
        if p.is_file(): paths.add(str(p))
        else: paths.update(str(v) for v in p.rglob('*') if v.is_file())
    sources = {}
    for name in sorted(paths):
        p = Path(name)
        assert p.stat().st_uid == p.stat().st_gid == 1000 and not p.is_symlink(), name
        sources[name] = {'bytes': p.stat().st_size, 'sha256': sha(p)}
    originals = {str(p.relative_to('/orig')): {'bytes': p.stat().st_size, 'sha256': sha(p)}
                 for p in sorted(Path('/orig').rglob('*')) if p.is_file()}
    for name in ['check', 'pixels', 'receipt']:
        ast.parse(Path('tools/scripts/windows_player_' + name + '.py').read_text())
    subprocess.run(['bash', '-n', 'tools/player_wine.sh'], check=True)
    record = {'schema': 'logh3.windows-player-receipts.v1', 'date': '2026-10-05', 'issue': 69,
              'parent_commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
              'original_bundle': {'path': str(bundle), 'sha256': sha(bundle)}, 'original_files': originals,
              'binary_sha256': sha(root / 'source103-player/gin3.exe'),
              'images': {'wine': 'colonization-windows-verification:20261001-r1',
                         'wine_id': 'sha256:7d3c7ff554db39300e3e7a92a7b79c2d5fcd647289596d22fa10899baac61201',
                         'linux_reference': 'logh3-ebiten:2',
                         'linux_id': 'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8'},
              'pixels': pixels, 'titles': titles, 'runtime': runtime,
              'standard': {'packages': 18, 'pass': 13, 'no_tests': 5},
              'limitations': ['Windows Wine GUI，未驗實機或實體音效裝置',
                              '1.5倍兩平台完整視窗154407像素不同，最大通道差1；邏輯畫布一致，未宣稱完整實體像素一致',
                              '只開取進度及艦隊編號頁，未新增Windows公開保存／讀回或戰鬥驗收',
                              '與同來源Linux渲染比對，不新增原版正常整回合同態',
                              '正式版號／dist-all、macOS、人耳及循環仍缺，#69 OPEN'],
              'sources': sources}
    with target.open('x') as stream:
        json.dump(record, stream, ensure_ascii=False, indent=2)
        stream.write('\n')
record = json.loads(target.read_text())
for name, row in record['sources'].items():
    p = Path(name)
    assert p.stat().st_size == row['bytes'] and sha(p) == row['sha256'], name
for name, row in record['original_files'].items():
    p = Path('/orig') / name
    assert p.stat().st_size == row['bytes'] and sha(p) == row['sha256'], name
print(len(record['sources']), '來源與', len(record['original_files']), '原版輸入身份通過；索引SHA-256', sha(target))
