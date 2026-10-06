"""核對第一批 HD 交付、實際封包與正常 GUI，保存同一份來源索引。容器內執行。"""
import ast
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tarfile
import zipfile

ROOT = Path('/src')
VERSION = 'v.0.1.4-20261006'


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda: f.read(1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()


def record(path, data):
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')


def entry(path):
    assert path.is_file() and path.stat().st_uid == os.getuid(), path
    return dict(path=str(path.relative_to(ROOT)), bytes=path.stat().st_size, sha256=sha(path))


def main():
    assert Path.cwd() == ROOT and os.getuid() == 1000
    root = ROOT/'dist-all'/VERSION
    build = json.loads((root/'BUILD.json').read_text())
    assert build['version'] == VERSION and len(build['packages']) == 4
    assert build['graphics'] == 'hd' and build['hd_portraits'] == 5
    art = json.loads((ROOT/'assets/hd/MANIFEST.json').read_text())
    assert len(art['files']) == 5
    for item in art['files']:
        path = ROOT/'assets/hd'/item['file']
        assert sha(path) == item['sha256'] and path.stat().st_size == item['bytes']
        original = ROOT/'assets/original/face'/path.name
        assert sha(original) == item['original_sha256']
    assert sha(ROOT/'assets/hd/face/face-00.png') == '411468e4d80cacc10edd2c82d5b4c0a35955a0580842cd8187835d37d87fe285'
    for package in build['packages']:
        path = root/package['path']
        assert path.stat().st_size == package['bytes'] and sha(path) == package['sha256'], path
        if path.suffix == '.zip':
            with zipfile.ZipFile(path) as z:
                assert z.testzip() is None
                top = path.stem
                mf = json.loads(z.read(top+'/MANIFEST.json'))
                assert mf['version'] == VERSION
                for name, expected in mf['files'].items():
                    data = z.read(top+'/'+name)
                    assert len(data) == expected['bytes'] and hashlib.sha256(data).hexdigest() == expected['sha256']
                assert set(z.namelist()) == {top+'/MANIFEST.json', *(top+'/'+name for name in mf['files'])}
                art_prefix = '銀河英雄傳說III.app/Contents/Resources/art/' if 'macos' in top else 'art/'
                for item in art['files']:
                    data = z.read(top+'/'+art_prefix+item['file'])
                    assert hashlib.sha256(data).hexdigest() == item['sha256']
                launch = '銀河英雄傳說III.app/Contents/MacOS/launch' if 'macos' in top else 'start-original.cmd' if 'windows' in top else 'start-original.sh'
                assert '-graphics hd' in z.read(top+'/'+launch).decode()
        else:
            with tarfile.open(path) as t:
                files = [m for m in t if m.isfile()]
                assert len(files) == len(build['source_files'])
                for member in files:
                    data = t.extractfile(member).read()
                    expected = build['source_files'][member.name.removeprefix('logh3-source/')]
                    assert len(data) == expected['bytes'] and hashlib.sha256(data).hexdigest() == expected['sha256']
    record(root/'SHA256SUMS.json', {'version': VERSION, 'files': build['packages']})
    gui = []
    for name in ['empire', 'alliance']:
        out = root/'smoke'/('gui-'+name)
        trace = json.loads((out/'TRACE.json').read_text())
        assert trace['normal_input'] and not trace['script_flag']
        assert str(root/'full-local'/('LOGH3-'+VERSION+'-linux-amd64/gin3')) == trace['command'][0]
        comparisons = sorted(out.glob('*-x11-comparison.json'))
        assert len(comparisons) == 28
        values = [json.loads(p.read_text()) for p in comparisons]
        assert all(v['different'] == 0 and v['max_channel_difference'] == 0 for v in values)
        gui.append(dict(faction=name, windows=len(values), pixels=sum(v['pixels'] for v in values), different=0, scale=trace['scale']))
    win = ROOT/'workplace/music69-windows'/('package92-smoke-'+VERSION)
    smoke = json.loads((win/'SMOKE.json').read_text())
    assert smoke['version'] == VERSION and len(smoke['runs']) == 6
    destination = root/'smoke/windows'
    destination.mkdir(exist_ok=True)
    for p in win.iterdir():
        if p.is_file() and p.suffix in ['.json', '.png', '.txt', '.log']:
            target = destination/p.name
            if target.exists():
                assert sha(target) == sha(p), '拒絕覆寫不同的Windows收據'
            else:
                shutil.copyfile(p, target)
    tests = ROOT/'workplace/hd-art-preview-20261006/standard-tests-v2.log'
    text = tests.read_text()
    assert 'FAIL' not in text and sum(line.startswith('ok') for line in text.splitlines()) == 14
    assert sum('[no test files]' in line for line in text.splitlines()) == 5
    shutil.copyfile(tests, root/'smoke/standard-tests.log')
    assert (ROOT/'workplace/hd-art-preview-20261006/portrait-in-game.bin').read_bytes() == (ROOT/'workplace/hd-art-preview-20261006/portrait-original-in-game.bin').read_bytes()
    old = json.loads((ROOT/'docs/internals/data/player-input-hotspots.json').read_text())
    previous = ROOT/'dist-all'/old['version']
    for item in old['delivery_files']:
        p = previous/item['path']
        assert p.stat().st_size == item['bytes'] and sha(p) == item['sha256']
    sources = []
    for rel in ['cmd/gin3/main.go', 'cmd/gin3/artwork_test.go', 'internal/asset/amg.go', 'internal/asset/draw.go',
                'internal/gfx/framebuffer.go', 'internal/gfx/artwork.go', 'internal/presentation/artwork.go',
                'internal/presentation/artwork_test.go', 'docs/spec/122-hd-artwork.md', 'docs/playtest/108-hd-artwork.md',
                'tools/scripts/hd_art_model.go', 'tools/scripts/hd_art_gui.py', 'tools/scripts/hd_art_capture.go',
                'tools/scripts/hd_art_finalize.py', 'tools/scripts/local_test_package.py',
                'tools/scripts/local_test_macos_package.py', 'tools/scripts/local_test_windows_smoke.py']:
        p = ROOT/rel
        if p.suffix == '.py': ast.parse(p.read_text())
        sources.append(entry(p))
    sources.extend(entry(ROOT/rel) for rel in ['README.md', 'CONTEXT.md', 'docs/internals/00-README.md',
        'docs/goal/2026-10-06-hd-artwork-round-121.md'])
    sources.extend(entry(p) for p in sorted((ROOT/'assets/hd').rglob('*')) if p.is_file())
    # 引擎封包的來源與目前正式Go檔逐項核對；vendor另由來源tar核對。
    for rel, expected in build['source_files'].items():
        if rel.startswith(('cmd/', 'internal/')) or rel in ['go.mod', 'go.sum', 'LICENSE']:
            p = ROOT/rel
            assert p.stat().st_size == expected['bytes'] and sha(p) == expected['sha256'], rel
    delivery = [entry(p) for p in sorted(root.rglob('*')) if p.is_file() and 'full-local' not in p.parts]
    delivery.extend(entry(root/p['path']) for p in build['packages'])
    record(ROOT/'docs/internals/data/hd-artwork.json', dict(version=VERSION, status='first-batch-conformed; full-HD-incomplete',
        generated_by='tools/scripts/hd_art_finalize.py', hd_portraits=5, total_original_portraits=83,
        source_head=build['head'], working_tree=build['working_tree'], source_files=sources,
        delivery_files=delivery, gui=gui, windows_runs=6, tests=dict(packages=19, passed=14, no_tests=5),
        macos='universal and ARM signature checked; no hardware runtime', previous_delivery_unchanged=True,
        historical_source_rule='previous frozen sources refer to e216da2 or earlier recorded revisions',
        not_claimed=['全套HD美術', '原版像素parity', 'Windows與macOS真機', '人耳']))
    print('PASS HD first batch: ZIP manifests, 5 portraits, 56 GUI frames, Windows6, macOS, previous delivery, tests')
    print(json.dumps(gui, ensure_ascii=False))


if __name__ == '__main__':
    main()
