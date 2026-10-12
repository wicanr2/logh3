"""集中v1封包、正常操作與推廣片證據；沿已驗證產物，拒絕混入後續程式。"""
import hashlib
import json
import os
import re
from pathlib import Path
import shutil
import zipfile
from local_test_macos_package import universal

ROOT = Path('/src')
VERSION = 'v.1.0.0-20261010'
OUT = ROOT / 'dist-all' / VERSION


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as source:
        for block in iter(lambda: source.read(1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def rec(path, relative=OUT):
    return {'path': str(path.relative_to(relative)), 'bytes': path.stat().st_size, 'sha256': sha(path)}


def copy_checked(source, target):
    if target.exists():
        assert sha(source) == sha(target), str(target)
    else:
        shutil.copyfile(source, target)


def main():
    assert os.getuid() == 1000 and OUT.stat().st_uid == 1000
    build = json.loads((OUT / 'BUILD.json').read_text())
    assert build['version'] == VERSION and build['hd_people'] == 82 and build['hd_portraits'] == 83
    stage = ROOT / f'workplace/package92-{VERSION}/source'
    for name, expected in build['source_files'].items():
        path = (stage if name.startswith('vendor/') else ROOT) / name
        assert path.stat().st_size == expected['bytes'] and sha(path) == expected['sha256'], name
    assert len(build['packages']) == 8
    manifests = []
    for package in build['packages']:
        archive = OUT / package['path']
        assert rec(archive) == package
        if archive.suffix != '.zip':
            continue
        name = archive.stem
        with zipfile.ZipFile(archive) as z:
            assert z.testzip() is None
            manifest = json.loads(z.read(name + '/MANIFEST.json'))
            assert manifest['version'] == VERSION
            assert set(z.namelist()) == {name + '/MANIFEST.json', *(name + '/' + p for p in manifest['files'])}
            for file, expected in manifest['files'].items():
                data = z.read(name + '/' + file)
                assert len(data) == expected['bytes'] and hashlib.sha256(data).hexdigest() == expected['sha256'], file
            if package['path'].startswith('release/'):
                assert not any('/original/' in p or '/art/' in p or p.lower().endswith(('.ogg', '.sf2', '.amg', '.bmg', '.img', '.mds', '.dat', '.bit')) for p in z.namelist())
            manifests.append({'package': package['path'], 'files': len(manifest['files']), 'manifest_sha256': sha(archive.parent / name / 'MANIFEST.json')})
    macos = []
    for kind in ('full', 'release'):
        app = OUT / kind / f'LOGH3-{VERSION}-macos-universal/銀河英雄傳說III.app/Contents/MacOS/gin3'
        macos.append({'kind': kind, 'sha256': sha(app), 'slices': universal(app)})
    assert macos[0]['sha256'] == macos[1]['sha256']
    for platform, binary in [('linux-amd64', 'gin3'), ('windows-amd64', 'gin3.exe')]:
        full = OUT / 'full' / f'LOGH3-{VERSION}-{platform}' / binary
        engine = OUT / 'release' / f'LOGH3-{VERSION}-{platform}' / binary
        assert sha(full) == sha(engine)
    gui = []
    for mode in ('0', '1', '2', 'original'):
        folder = OUT / 'smoke' / ('gui-' + mode)
        trace = json.loads((folder / 'TRACE.json').read_text())
        assert trace['normal_input'] is True and trace['script_flag'] is False
        comparisons = [json.loads(p.read_text()) for p in sorted(folder.glob('*-comparison.json'))]
        assert len(comparisons) == (33 if mode == '2' else 32)
        assert all(c['different'] == 0 and c['bbox'] is None and c['max_channel_difference'] == 0 for c in comparisons)
        gui.append({'mode': mode, 'scale': trace['scale'], 'captures': len(comparisons),
                    'pixels': sum(c['pixels'] for c in comparisons), 'different': 0, 'normal_input': True})
    # 收攏已完成的跨平台抽驗，保留原命令與工具等級。
    for source, name in [
        (ROOT / f'workplace/package92-smoke-{VERSION}', 'linux-full'),
        (ROOT / f'workplace/music69-windows/package92-smoke-{VERSION}', 'windows-full'),
        (ROOT / f'workplace/music69-windows/release-smoke-{VERSION}', 'windows-release'),
    ]:
        target = OUT / 'smoke' / name
        target.mkdir(exist_ok=True)
        for p in source.iterdir():
            if p.is_file() and p.suffix in ('.json', '.txt', '.log', '.png', '.bin'):
                copy_checked(p, target / p.name)
    log = ROOT / 'workplace/release100-20261010/standard-tests.log'
    lines = log.read_text().splitlines()
    assert sum(l.startswith('ok ') for l in lines) == 14
    assert sum(l.startswith('? ') for l in lines) == 5
    copy_checked(log, OUT / 'smoke/standard-tests.log')
    for name in ('HD-INVENTORY.json', 'REFINEMENT.json'):
        copy_checked(ROOT / 'workplace/release100-20261010' / name, OUT / 'smoke' / name)
    promo_source = ROOT / ('workplace/promo-' + VERSION)
    audit = json.loads((promo_source / 'AUDIT.json').read_text())
    movie = promo_source / audit['file']
    assert sha(movie) == audit['sha256'] and audit['frames'] == 1980 and audit['seconds'] == 66
    assert audit['no_black_segment'] and audit['audio_present'] and audit['subtitle_clipped'] == 0
    assert len(audit['motion_checks']) == 2 and all(m['different'] for m in audit['motion_checks'])
    promo = OUT / 'promo'
    promo.mkdir(exist_ok=True)
    for p in promo_source.iterdir():
        if p.is_file() and (p.name in ('PLAN.json', 'AUDIT.json', 'FFPROBE.json', 'AUDIO-AUDIT.txt', 'VIDEO-AUDIT.txt', 'TEXT-BOUNDS.json', 'CONTACT.png', movie.name) or p.name.startswith('sample-')):
            copy_checked(p, promo / p.name)
    (promo / 'RIGHTS.txt').write_text('本片音畫來源及hash見PLAN.json。含原版衍生圖像、配樂及HD人物，只留本機。畫面為正式包靜態剪輯；不宣稱連續實況或新增原版same-state證據。\n')
    plan = ROOT / 'docs/goal/2026-10-10-v1-and-hd-plan-round-122.md'
    def rebase_link(match):
        target = match.group(1)
        if '://' in target or target.startswith('#'):
            return match.group(0)
        resolved = (plan.parent / target).resolve()
        assert resolved.is_file(), str(resolved)
        return '](' + os.path.relpath(resolved, OUT) + ')'
    plan_text = re.sub(r'\]\(([^)]+)\)', rebase_link, plan.read_text())
    plan_target = OUT / 'HD-PLAN.md'
    if plan_target.exists():
        assert plan_target.read_text() == plan_text
    else:
        plan_target.write_text(plan_text)
    packages = build['packages'] + [rec(promo / movie.name)]
    evidence = [rec(p, ROOT) for p in sorted((OUT / 'smoke').rglob('*')) if p.is_file()]
    evidence += [rec(p, ROOT) for p in sorted(promo.iterdir()) if p.is_file() and p.suffix != '.mp4']
    source = [rec((stage if p.startswith('vendor/') else ROOT) / p, ROOT) for p in build['source_files']]
    source += [rec(p, ROOT) for p in sorted((ROOT / 'tools/scripts').glob('v1_*.py'))]
    source += [rec(ROOT / 'tools/scripts' / p, ROOT) for p in ('local_test_package.py', 'local_test_macos_package.py', 'local_test_macos_build.sh', 'local_test_smoke.py', 'local_test_windows_smoke.py')]
    source += [rec(ROOT / p, ROOT) for p in ('tools/scripts/hd_art_model.go', 'tools/scripts/hd_art_gui.py', 'assets/hd/MANIFEST.json', 'docs/spec/123-hd-complete-presentation.md', 'docs/spec/128-visible-button-activation.md')]
    summary = {'version': VERSION, 'source_head': build['head'], 'dirty_source_snapshot': True,
        'source': source, 'packages': packages, 'manifests': manifests, 'gui': gui, 'evidence': evidence,
        'tests': {'packages': 19, 'pass': 14, 'without_tests': 5}, 'hd': {'people': 82, 'blank': 1, 'other_art': '三倍邊緣細化，非逐張AI重繪', 'source_art_samples': 33},
        'inventory': json.loads((OUT / 'smoke/HD-INVENTORY.json').read_text()), 'macos': macos,
        'promo': audit, 'hd_plan': rec(plan, ROOT), 'preview': rec(OUT / 'HD-PREVIEW.html'),
        'tools': {'gui': 'logh3-ebiten:2', 'macos': 'psychicwar-osxcross:go1.24.13-15.5-r2', 'windows': 'kol-wine-verify:ubuntu-noble-20261007-r1', 'audio': 'eob-audio-capture:20261009-r3'},
        'rights': {'release': '引擎與重建版面，無原版或HD衍生素材', 'full': '含原版及衍生素材，只留本機', 'promo': '含原版衍生音畫，只留本機', 'hd_preview': '含原版衍生圖像，只留本機'},
        'not_claimed': build['not_claimed'], 'published': False}
    write(OUT / 'FINAL-VERIFICATION.json', summary)
    write(ROOT / 'docs/internals/data/v1-release.json', summary)
    write(OUT / 'SHA256SUMS.json', {'version': VERSION, 'files': packages + [rec(OUT / p) for p in ('BUILD.json', 'FINAL-VERIFICATION.json', 'HD-PLAN.md', 'HD-PREVIEW.html')]})
    print('PASS 8封包及影片身份／6ZIP逐檔／129完整視窗／來源與三平台程式', flush=True)
    print(json.dumps(gui, ensure_ascii=False), flush=True)


if __name__ == '__main__':
    main()
