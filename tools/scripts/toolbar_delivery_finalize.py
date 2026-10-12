"""規格138：驗收正式封包、兩個正常啟動器與來源身份，通過才輸出八包索引。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import tarfile
import zipfile

ROOT = Path('/src')


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for chunk in iter(lambda: stream.read(1024*1024), b''):
            h.update(chunk)
    return h.hexdigest()


def read(path):
    return json.loads(path.read_text())


def write(path, value):
    assert path.parent.stat().st_uid == os.getuid()
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')


def record(path):
    assert path.is_file() and path.stat().st_uid == os.getuid(), path
    return {'path': str(path.relative_to(ROOT)), 'bytes': path.stat().st_size, 'sha256': sha(path)}


def main(version):
    assert Path.cwd() == ROOT and os.getuid() == 1000
    assert re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}', version)
    out = ROOT/'dist-all'/version
    batch = ROOT/'workplace/launch128'
    build = read(out/'BUILD.json')
    assert build['version'] == version and len(build['packages']) == 8
    frozen = ROOT/'workplace'/('package92-'+version+'-ui')/'source'
    for name, item in build['source_files'].items():
        path = frozen/name
        assert path.stat().st_size == item['bytes'] and sha(path) == item['sha256'], path
        if not name.startswith('vendor/'):
            assert sha(ROOT/name) == item['sha256'], ('來源已改動', name)
    test_log = batch/'final-standard-tests.log'
    lines = test_log.read_text().splitlines()
    assert not any(line.startswith(('FAIL', '--- FAIL')) for line in lines)
    tests = {'passed': sum(line.startswith('ok ') for line in lines),
             'no_tests': sum(line.startswith('? ') for line in lines)}
    assert tests == {'passed': 15, 'no_tests': 4}, tests
    manifests = []
    for item in build['packages']:
        path = out/item['path']
        assert path.stat().st_size == item['bytes'] and sha(path) == item['sha256'], path
        if path.suffix == '.zip':
            with zipfile.ZipFile(path) as archive:
                assert archive.testzip() is None
                name = path.stem
                names = set(archive.namelist())
                assert all(not Path(n).is_absolute() and '..' not in Path(n).parts for n in names)
                manifest = read(path.parent/name/'MANIFEST.json')
                assert manifest['version'] == version
                assert names == {name+'/'+n for n in manifest['files']} | {name+'/MANIFEST.json'}
                for n, expected in manifest['files'].items():
                    data = archive.read(name+'/'+n)
                    assert len(data) == expected['bytes'] and hashlib.sha256(data).hexdigest() == expected['sha256'], n
                    unpacked = path.parent/name/n
                    assert sha(unpacked) == expected['sha256'], unpacked
                    if path.parent.name == 'release':
                        assert not {'original', 'art', 'music', 'soundfont'}.intersection(Path(n).parts), n
                    if path.parent.name == 'full' and n == 'README.txt':
                        assert all(key in data.decode('utf-8-sig') for key in ['F1', 'F7', 'F8']), n
                    if 'windows' not in name and (n.endswith(('.sh', '.command')) or n.endswith(('/MacOS/gin3', '/MacOS/launch')) or n == 'gin3'):
                        assert archive.getinfo(name+'/'+n).external_attr >> 16 & 0o111, n
                manifests.append({'path': item['path'], 'files': len(manifest['files']), 'passed': True})
        else:
            assert path.name.endswith('-engine-source.tar.gz')
            with tarfile.open(path, 'r:gz') as archive:
                files = [m for m in archive.getmembers() if m.isfile()]
                assert len(files) == len(build['source_files'])
                for member in files:
                    n = str(Path(member.name).relative_to('logh3-source'))
                    expected = build['source_files'][n]
                    assert member.size == expected['bytes']
                    assert hashlib.sha256(archive.extractfile(member).read()).hexdigest() == expected['sha256'], n
    linux = out/'full'/('LOGH3-'+version+'-linux-amd64')
    for platform in ['linux-amd64', 'windows-amd64', 'macos-universal']:
        binary = 'gin3.exe' if platform.startswith('windows') else '銀河英雄傳說III.app/Contents/MacOS/gin3' if platform.startswith('macos') else 'gin3'
        name = 'LOGH3-'+version+'-'+platform
        assert sha(out/'full'/name/binary) == sha(out/'release'/name/binary)
    for source in (ROOT/'assets/hd').rglob('*'):
        if source.is_file() and source.name != 'README.md':
            assert sha(source) == sha(linux/'art'/source.relative_to(ROOT/'assets/hd')), source
    normal = []
    for mode, faction in [('original', 0), ('enhanced', 1)]:
        folder = batch/('full-'+mode)
        result = read(folder/'RESULT.json')
        assert result['passed'] and result['normal_entry'] and not result['script_flag'] and not result['state_injection']
        assert result['faction'] == faction and result['command'][0] == str(linux/('start-'+mode+'.sh'))
        assert len([e for e in result['events'] if 'check' in e]) >= 25
        assert result['account_hover']['changed_pixels'] >= 60 and result['account_hover']['brightness_increase'] >= 6
        dest = out/'smoke'/('normal-'+mode)
        assert not dest.exists()
        shutil.copytree(folder, dest, ignore=shutil.ignore_patterns('pulse', 'save'))
        normal.append({'mode': mode, 'faction': faction, 'checks': sum('check' in e for e in result['events']),
                       'screenshots': sum('screenshot' in e for e in result['events']),
                       'hover': result['account_hover'], 'result_sha256': sha(folder/'RESULT.json')})
    wine = ROOT/'workplace/music69-windows'/('package92-smoke-'+version)
    wine_result = read(wine/'SMOKE.json')
    assert wine_result['version'] == version and len(wine_result['runs']) == 6
    shutil.copytree(wine, out/'smoke/full-windows-wine', ignore=shutil.ignore_patterns('LOGH3-*', 'save-*'))
    engine_wine = ROOT/'workplace/music69-windows'/('release-smoke-'+version)
    assert len(read(engine_wine/'SMOKE.json')['runs']) == 4
    shutil.copytree(engine_wine, out/'smoke/release-windows-wine', ignore=shutil.ignore_patterns('LOGH3-*', 'save-*'))
    assert len(read(out/'smoke/release-linux/SMOKE.json')['runs']) == 4
    mac = read(out/'smoke/macos/MACOS.json')
    assert {s['architecture'] for s in mac['slices']} == {'amd64', 'arm64'}
    assert all(c['all_page_hashes_match'] for s in mac['slices'] for c in s['code_directory'])
    shutil.copyfile(test_log, out/'smoke/standard-tests.log')
    previous = ROOT/'docs/internals/data/live-graphics-review.json'
    prior = read(previous)
    movie = ROOT/'dist-all'/prior['version']/'promo'/('LOGH3-'+prior['version']+'-promo.mp4')
    # 前版影片只驗身份並連回前版收據，不改名、不冒稱本輪新錄製。
    assert sha(movie) == '52c7353548a1175153a5d79f312f9145a386eafaa7ef5f61f0702b9127be37cb'
    (out/'promo/README.md').write_text('本版只更新介面。現行推廣片沿v.1.1.3-20261011的116秒實錄，沒有重標版號。\n\n'
        '影片位於../../v.1.1.3-20261011/promo/LOGH3-v.1.1.3-20261011-promo.mp4。\n'
        '該片使用舊快捷鍵F6切換顯示、F7顯示敵軍；新版改為F7／F8，F1說明。\n'
        '音畫及權利驗收見docs/playtest/116-live-graphics-review.md。\n')
    check = {'version': version, 'passed': True, 'source_files': len(build['source_files']),
             'tests': tests, 'normal_gui': normal, 'manifests': manifests,
             'windows_wine_runs': 6, 'linux_engine_runs': 4, 'windows_engine_wine_runs': 4,
             'macos': mac, 'previous_promo': record(movie),
             'not_claimed': ['Windows／macOS真機', '人耳驗收', 'DOS全遊戲parity', '完整登陸分段', '新版影片錄製']}
    write(out/'smoke/PACKAGE-CHECK.json', check)
    sources = [record(ROOT/n) for n in build['source_files'] if not n.startswith('vendor/')]
    sources += [record(p) for p in (ROOT/'assets/hd').rglob('*') if p.is_file()]
    sources += [record(ROOT/n) for n in [
        'tools/scripts/toolbar_gui_check.py', 'tools/scripts/toolbar_delivery_finalize.py',
        'tools/scripts/local_test_package.py', 'tools/scripts/local_test_macos_package.py',
        'tools/scripts/local_test_macos_build.sh', 'tools/scripts/v1_release.py',
        'docs/spec/138-top-toolbar-and-hover.md']]
    evidence = [record(p) for folder in ['full-original', 'full-enhanced'] for p in (batch/folder).rglob('*')
                if p.is_file() and not {'pulse', 'save'}.intersection(p.relative_to(batch).parts)]
    evidence += [record(batch/n) for n in ['final-standard-tests.log', 'package-build.log',
        'macos-package.log', 'engine-package.log', 'concept/PREVIEW.html', 'concept/VARIANTS.json',
        'gui-empire/RESULT.json', 'gui-empire-r2/RESULT.json', 'gui-empire-r3/RESULT.json',
        'gui-alliance-r3/RESULT.json', 'hover-probe.txt']]
    index = {'version': version, 'status': 'PASS 已核准A調暗、頂端功能列、反光及三平台本機交付',
             'source_head': build['head'], 'dirty_source_snapshot': True,
             'frozen_build_source': str(frozen.relative_to(ROOT)), 'packages': build['packages'],
             'checks': check, 'sources': sources, 'evidence': evidence,
             'previous_promo_index': record(previous)}
    write(ROOT/'docs/internals/data/top-toolbar-hover.json', index)
    print('PASS 八封包、兩陣營正式啟動器正常視窗、來源與標準測試', flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--version', required=True)
    main(parser.parse_args().version)
