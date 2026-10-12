"""第125輪：只在Docker核對並集中原配置HD交付，不發布遠端Release。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import zipfile

ROOT = Path('/src')


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda: f.read(1024*1024), b''):
            h.update(block)
    return h.hexdigest()


def write(path, value):
    assert path.parent.stat().st_uid == os.getuid()
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')


def record(path):
    return dict(path=str(path.relative_to(ROOT)), bytes=path.stat().st_size, sha256=sha(path))


def copy(source, target):
    assert source.is_file() and not source.is_symlink()
    target.parent.mkdir(parents=True, exist_ok=True)
    assert target.parent.stat().st_uid == os.getuid()
    if target.exists():
        assert sha(source) == sha(target), target
    else:
        shutil.copyfile(source, target)


def copy_receipt(source, target, recursive=False):
    files = source.rglob('*') if recursive else source.glob('*')
    for p in sorted(files):
        if p.is_file() and p.suffix in ('.json', '.png', '.txt', '.log', '.bin', '.md5'):
            copy(p, target/p.relative_to(source))


def main(version):
    assert os.getuid() == 1000
    out = ROOT/'dist-all'/version
    batch = ROOT/'workplace/hd125'
    build = json.loads((out/'BUILD.json').read_text())
    assert build['version'] == version and build['hd_complete_presentation']
    frozen = ROOT/f'workplace/package92-{version}-hd/source'
    for name, expected in build['source_files'].items():
        assert sha(frozen/name) == expected['sha256']
        if not name.startswith('vendor/'):
            assert sha(ROOT/name) == expected['sha256'], name
    assert len(build['packages']) == 8
    manifests = []
    for item in build['packages']:
        path = out/item['path']
        assert path.stat().st_size == item['bytes'] and sha(path) == item['sha256']
        if path.suffix != '.zip':
            continue
        with zipfile.ZipFile(path) as z:
            assert z.testzip() is None
            name = path.stem
            m = json.loads(z.read(name+'/MANIFEST.json'))
            assert m['version'] == version
            assert set(z.namelist()) == {name+'/MANIFEST.json', *(name+'/'+f for f in m['files'])}
            for f, expected in m['files'].items():
                data = z.read(name+'/'+f)
                assert len(data) == expected['bytes'] and hashlib.sha256(data).hexdigest() == expected['sha256'], f
            if item['path'].startswith('release/'):
                forbidden = {'.amg', '.bmg', '.img', '.mds', '.dat', '.bit', '.ogg', '.sf2', '.com', '.rgn', '.spr'}
                assert not any('/original/' in f or '/art/' in f or Path(f).suffix.lower() in forbidden for f in z.namelist())
            manifests.append(dict(package=item['path'], files=len(m['files']), valid=True))
    art = ROOT/'assets/hd'
    full_art = out/'full'/f'LOGH3-{version}-linux-amd64/art'
    for p in art.rglob('*'):
        if p.is_file() and p.name != 'README.md':
            assert sha(p) == sha(full_art/p.relative_to(art)), p
    for platform, binary in [('linux-amd64','gin3'), ('windows-amd64','gin3.exe'), ('macos-universal','銀河英雄傳說III.app/Contents/MacOS/gin3')]:
        assert sha(out/'full'/f'LOGH3-{version}-{platform}'/binary) == sha(out/'release'/f'LOGH3-{version}-{platform}'/binary)
    log = batch/'standard-tests-final.log'
    lines = log.read_text().splitlines()
    assert not any(l.startswith(('FAIL', '--- FAIL')) for l in lines)
    tests = dict(passed=sum(l.startswith('ok ') for l in lines), no_tests=sum(l.startswith('? ') for l in lines))
    assert tests == dict(passed=15, no_tests=4), tests
    copy(log, out/'smoke/standard-tests.log')
    gui = []
    for faction in ('empire','alliance'):
        source = batch/f'x11-{faction}-validated'
        results = [json.loads(p.read_text()) for p in source.glob('*-x11-comparison.json') if '-attempt-' not in p.name]
        assert len(results) == 32 and all(r['different'] == 0 for r in results)
        gui.append(dict(faction=faction, screens=32, pixels=sum(r['pixels'] for r in results), different=0))
        copy_receipt(source, out/'smoke'/f'gui-{faction}')
    maps = json.loads((batch/'maps-r3/MAPS.json').read_text())
    assert maps['maps'] == 32 and maps['cameras'] == 64 and all(r['indexed_replay_equal'] for r in maps['runs'])
    animation = json.loads((batch/'animations-r11/ANIMATIONS.json').read_text())
    assert animation['comani_frames'] == 1438 and animation['hd_ship_models'] == 45 and animation['intro_playback_frames'] == 421
    assert len(animation['samples']) == 157 and all(r['indexed_replay_equal'] for r in animation['samples'])
    copy_receipt(batch/'maps-r3', out/'smoke/maps')
    copy_receipt(batch/'animations-r11', out/'smoke/animations')
    receipts = [(ROOT/f'workplace/package92-smoke-{version}', 'linux-full'),
                (ROOT/f'workplace/music69-windows/package92-smoke-{version}', 'windows-full'),
                (ROOT/f'workplace/music69-windows/release-smoke-{version}', 'windows-release')]
    for source, name in receipts:
        assert (source/'SMOKE.json').is_file()
        copy_receipt(source, out/'smoke'/name)
    assert (out/'smoke/release-linux/SMOKE.json').is_file()
    for source, name in [(batch/'scenarios-linux','scenarios-linux'), (ROOT/f'workplace/music69-windows/scenarios-{version}','scenarios-windows')]:
        s = json.loads((source/'SCENARIOS.json').read_text())
        assert [r['loaded_turn'] for r in s['runs']] == [346,93,346,380]
        assert all(r['reloaded_s3_turn'] == 380 for r in s['runs'])
        copy_receipt(source, out/'smoke'/name, True)
    raw = batch/'s1-hd-guarded'
    trace = json.loads((raw/'TRACE.json').read_text())
    assert trace['normal_x11_input'] and not trace['script_flag'] and not trace['state_injection']
    assert sha(raw/'RAW.mp4') == trace['raw_sha256']
    assert sha(batch/'gin3-s1-repro') == trace['binary_sha256'], 'S1候選程式不能由正式來源重生'
    promo = batch/'story/promo'
    audit = json.loads((promo/'AUDIT.json').read_text())
    movie = promo/f'LOGH3-{version}-promo.mp4'
    assert audit['video_sha256'] == sha(movie) and audit['seconds'] == 90 and audit['frames'] == 2700
    assert audit['visual_review']['passed'] and audit['non_silent'] and audit['black_segments'] == 0
    source_check = json.loads((promo/'SOURCE-CHECK.json').read_text())
    assert source_check['passed'] and len(source_check['samples']) == 9
    assert audit['source_frame_check']['passed'] and audit['source_frame_check']['samples'] == 9
    for p in sorted(promo.iterdir()):
        if p.is_file() and (p.suffix in ('.json','.txt','.md5','.png') or p.name == movie.name):
            copy(p, out/'promo'/p.name)
    copy(raw/'RAW.mp4', out/'promo/RAW-S1-HD.mp4')
    for name in ('TRACE.json','EVENTS.json','OPERATOR-EVENTS.json'):
        copy(raw/name, out/'promo'/('S1-'+name))
    copy(batch/'s1-hd-guarded-plan.json', out/'promo/S1-INPUT-PLAN.json')
    copy(batch/'story/SCENES.json', out/'promo/SCENES.json')
    copy(batch/'story/FLAG-SOURCES.json', out/'promo/FLAG-SOURCES.json')
    check = dict(version=version, source_files=len(build['source_files']), packages=build['packages'], manifests=manifests,
                 tests=tests, gui=gui, maps=dict(maps=32,cameras=64), animations=dict(models=45,segments=254,frames=1438,samples=157),
                 s1=record(raw/'TRACE.json'), promo=audit)
    write(out/'smoke/PACKAGE-CHECK.json', check)
    sources = [record(ROOT/name) for name in build['source_files'] if not name.startswith('vendor/')]
    sources += [record(p) for folder in (ROOT/'assets/hd',) for p in sorted(folder.rglob('*')) if p.is_file()]
    sources += [record(ROOT/'tools/scripts'/name) for name in ('hd_finalize.py','hd_promo_source_check.py','hd_complete_check.go','hd_art_gui.py','hd_art_model.go','hd_map_check.go','hd_animation_check.go','hd_ship_import.py','hd_font_atlas.py','s1_gameplay_record.py','s1_story_promo.py','local_test_package.py','local_test_macos_build.sh','local_test_macos_package.py','v1_release.py','v1_release_smoke.py','bundled_scenarios_smoke.py')]
    sources += [record(ROOT/'docs/spec'/f) for f in ('130-hd-drawing-journal.md','131-hd-map-materials.md','132-hd-strategy-and-symbols.md','133-hd-animation-surfaces.md')]
    sources += [record(ROOT/f) for f in ('docs/playtest/111-production-hd.md','docs/goal/2026-10-11-production-hd-round-125.md')]
    evidence = [record(p) for folder in (out/'smoke',out/'promo') for p in sorted(folder.rglob('*')) if p.is_file()]
    index = dict(version=version, status='PASS 原配置HD正式交付', source_head=build['head'], dirty_source_snapshot=True,
                 frozen_source=str(frozen.relative_to(ROOT)), delivery=str(out.relative_to(ROOT)), sources=sources,
                 evidence=evidence, packages=build['packages'], verification=check,
                 limits=['原版全遊戲same-state','完整登陸分段流程#48','Windows／macOS真機','人耳驗收','逐幀AI新繪'],
                 rights='完整版／HD／影片本機限定；引擎包不含原版與HD衍生素材，RRSAL-1.0', remote_release=False)
    write(ROOT/'docs/internals/data/production-hd.json', index)
    write(out/'DELIVERY.json', index)
    write(out/'SHA256SUMS.json', dict(version=version, files=build['packages']+[dict(path=str(movie.relative_to(ROOT/'workplace/hd125/story')), bytes=movie.stat().st_size, sha256=sha(movie))]))
    print('PASS 凍結來源、八封包／六ZIP、64正常視窗、三劇本、S1實錄與90秒HD影片', tests, flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--version', required=True)
    main(parser.parse_args().version)
