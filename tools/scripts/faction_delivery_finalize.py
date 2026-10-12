"""第126輪：核對最新赤焰戰線、實錄影片及三平台本機交付。僅在Docker執行。"""
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
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1024*1024), b''):
            h.update(block)
    return h.hexdigest()


def read(path):
    return json.loads(path.read_text())


def write(path, value):
    assert path.parent.stat().st_uid == os.getuid()
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')


def record(path):
    assert path.stat().st_uid == os.getuid() and path.is_file()
    return dict(path=str(path.relative_to(ROOT)), bytes=path.stat().st_size, sha256=sha(path))


def copy(source, target):
    assert source.is_file() and not source.is_symlink()
    target.parent.mkdir(parents=True, exist_ok=True)
    assert target.parent.stat().st_uid == os.getuid()
    if target.exists():
        assert sha(source) == sha(target), target
    else:
        shutil.copyfile(source, target)


def receipts(source, target, recursive=False):
    paths = source.rglob('*') if recursive else source.glob('*')
    for p in sorted(paths):
        if p.is_file() and p.suffix in ('.json', '.jsonl', '.png', '.txt', '.log', '.bin', '.md5'):
            # 只複製完整驗收項，不把失敗重試及錄製中的暫存圖當成驗收。
            if '-attempt-' not in p.name and p.name != 'WAIT.png':
                copy(p, target/p.relative_to(source))


def main(version):
    assert os.getuid() == 1000
    out = ROOT/'dist-all'/version
    batch = ROOT/'workplace/hd126'
    build = read(out/'BUILD.json')
    assert build['version'] == version and build['hd_complete_presentation']
    frozen = ROOT/f'workplace/package92-{version}-hd/source'
    for name, expected in build['source_files'].items():
        assert sha(frozen/name) == expected['sha256'], name
        if not name.startswith('vendor/'):
            assert sha(ROOT/name) == expected['sha256'], name
    assert len(build['packages']) == 8
    manifests = []
    for item in build['packages']:
        path = out/item['path']
        assert path.stat().st_size == item['bytes'] and sha(path) == item['sha256'], path
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
    full_art = out/'full'/f'LOGH3-{version}-linux-amd64/art'
    for p in (ROOT/'assets/hd').rglob('*'):
        if p.is_file() and p.name != 'README.md':
            assert sha(p) == sha(full_art/p.relative_to(ROOT/'assets/hd')), p
    for platform, binary in [('linux-amd64','gin3'), ('windows-amd64','gin3.exe'), ('macos-universal','銀河英雄傳說III.app/Contents/MacOS/gin3')]:
        assert sha(out/'full'/f'LOGH3-{version}-{platform}'/binary) == sha(out/'release'/f'LOGH3-{version}-{platform}'/binary)
    macos = read(out/'smoke/macos/MACOS.json')
    assert {s['architecture'] for s in macos['slices']} == {'amd64','arm64'}
    assert macos['zip_integrity'] and macos['executable_permissions_preserved']
    assert all(c['all_page_hashes_match'] for s in macos['slices'] for c in s['code_directory'])
    print('PASS 1523凍結來源、八封包／六ZIP、三平台相同引擎與最新美術', flush=True)

    log = batch/'faction-standard-tests-r7.log'
    lines = log.read_text().splitlines()
    assert not any(l.startswith(('FAIL','--- FAIL')) for l in lines)
    tests = dict(passed=sum(l.startswith('ok ') for l in lines), no_tests=sum(l.startswith('? ') for l in lines))
    assert tests == dict(passed=15, no_tests=4)
    copy(log, out/'smoke/standard-tests.log')
    gui = []
    for faction in ('empire','alliance'):
        folder = batch/f'theme-x11-{faction}-r7'
        results = [read(p) for p in folder.glob('*-x11-comparison.json') if '-attempt-' not in p.name]
        assert len(results) == 34 and all(r['different'] == 0 for r in results)
        gui.append(dict(faction=faction, screens=34, pixels=sum(r['pixels'] for r in results), different=0))
        receipts(folder, out/'smoke'/f'gui-{faction}')
    approved = read(batch/'APPROVED-UI-COMPARISON-r7.json')
    assert len(approved['comparisons']) == 8 and all(r['different'] == 0 for r in approved['comparisons'])
    for name in ('APPROVED-UI-COMPARISON-r7.json','FACTION-GAME-BROWSER-RESULTS-r7.json'):
        copy(batch/name, out/'smoke'/name)
    copy(batch/'FACTION-HD-GAME.html', out/'promo/FACTION-HD-GAME.html')
    smoke_counts = {}
    for folder, name, count in [
        (ROOT/f'workplace/package92-smoke-{version}','linux-full',11),
        (ROOT/f'workplace/music69-windows/package92-smoke-{version}','windows-full',6),
        (ROOT/f'workplace/music69-windows/release-smoke-{version}','windows-release',4),
        (out/'smoke/release-linux','release-linux',4),
    ]:
        result = read(folder/'SMOKE.json')
        assert result['version'] == version and len(result['runs']) == count
        smoke_counts[name] = count
        if folder != out/'smoke'/name:
            receipts(folder, out/'smoke'/name)
    for folder, name in [(batch/'scenarios-linux-v112','scenarios-linux'), (ROOT/f'workplace/music69-windows/scenarios-{version}','scenarios-windows')]:
        s = read(folder/'SCENARIOS.json')
        assert [r['loaded_turn'] for r in s['runs']] == [346,93,346,380]
        assert all(r['reloaded_s3_turn'] == 380 for r in s['runs'])
        receipts(folder, out/'smoke'/name, True)

    tactical = read(batch/'ai-model-r2/AI.json')
    strategic = read(batch/'strategy-ai-model-r4/STRATEGY.json')
    assert tactical['same_initial_tables'] and strategic['same_initial_tables']
    assert not strategic['state_injection'] and not strategic['seed_injection']
    assert strategic['results']['original']['commands'] != strategic['results']['enhanced']['commands']
    for stage in ('strategy','entry-off'):
        for stem in ('ADMIRAL','FLEET','SHIP','WORK','STELLAR','PLANET'):
            assert (batch/'ai-model-r2/original'/stage/(stem+'.DT0')).read_bytes() == (batch/'ai-model-r2/enhanced'/stage/(stem+'.DT0')).read_bytes()
    normal_recordings = []
    names = [
        ('s1-record-r4','S1-HD',0),
        ('ai-record-original-r3','TACTICAL-ORIGINAL',10),
        ('ai-record-enhanced-r2','TACTICAL-ENHANCED',10),
        ('strategy-record-original-r2','STRATEGY-ORIGINAL',6),
        ('strategy-record-enhanced-r2','STRATEGY-ENHANCED',6),
    ]
    for name, delivery_name, count in names:
        folder = batch/name
        trace = read(folder/'TRACE.json')
        assert trace['normal_x11_input'] and not any(trace[k] for k in ('script_flag','state_injection','seed_injection'))
        assert trace['binary_sha256'] == sha(out/'full'/f'LOGH3-{version}-linux-amd64/gin3')
        assert trace['raw_sha256'] == sha(folder/'RAW.mp4')
        comparisons = [read(p) for p in folder.glob('*-comparison.json') if '-attempt-' not in p.name]
        assert len(comparisons) == count and all(c['different'] == 0 for c in comparisons)
        if 'ai-record-' in name:
            assert sha(folder/'entry-off.png') == sha(folder/'entry-off-restored.png')
            assert sha(folder/'entry-off.png') != sha(folder/'entry-on.png')
        copy(folder/'RAW.mp4', out/'promo'/('RAW-'+delivery_name+'.mp4'))
        receipts(folder, out/'smoke'/('recording-'+delivery_name))
        normal_recordings.append(dict(name=name, raw_sha256=trace['raw_sha256'], comparisons=count, full_window_different=0))
    volumes = read(batch/'RECORDINGS-AUDIO.json')
    assert len(volumes) == 5 and all(r['passed'] and -55<r['mean_db']<-5 for r in volumes)
    copy(batch/'RECORDINGS-AUDIO.json', out/'promo/RECORDINGS-AUDIO.json')
    for name in ('s1-three-turn-plan.json','ai-original-four-turn-plan.json','ai-enhanced-four-turn-plan.json','FLAG-BACKGROUND-HD.json'):
        copy(batch/name, out/'promo'/name)
    for name in ('AI.json',):
        copy(batch/'ai-model-r2'/name, out/'smoke'/name)
    copy(batch/'strategy-ai-model-r4/STRATEGY.json', out/'smoke/STRATEGY.json')
    # 第125輪地圖／動畫證據只沿用既有有限範圍，不宣稱本輪重新驗原版。
    for folder, name in [(ROOT/'workplace/hd125/maps-r3','maps'), (ROOT/'workplace/hd125/animations-r11','animations')]:
        receipts(folder, out/'smoke'/name)

    promo = batch/'promo-r2'
    audit = read(promo/'AUDIT.json')
    movie = promo/f'LOGH3-{version}-promo.mp4'
    assert audit['passed'] and audit['movie_sha256'] == sha(movie)
    assert audit['video'] == [1920,1080,'30fps'] and abs(audit['duration']-118)<.3
    assert not audit['black_segments'] and audit['audio']['peak_db'] < 0
    assert len(audit['source_frame_samples']) >= 20
    for p in sorted(promo.iterdir()):
        if p.is_file() and '-first-cut' not in p.name and (p.suffix in ('.json','.txt','.png') or p.name == movie.name):
            copy(p, out/'promo'/p.name)
    visual = read(promo/'VISUAL-REVIEW.json')
    assert visual['passed'] and len(visual['samples']) >= 12
    copy(promo/'VISUAL-REVIEW.json', out/'promo/VISUAL-REVIEW.json')
    copy(batch/'flag-background-hd.png', out/'promo/flag-background-hd.png')
    check = dict(version=version, passed=True, source_files=len(build['source_files']), packages=build['packages'],
                 manifests=manifests, tests=tests, gui=gui, approved_ui=approved,
                 platform_smokes=smoke_counts, macos=macos, recordings=normal_recordings,
                 tactical_ai=tactical['same_initial_tables'], strategic_ai=strategic, promo=audit, visual_review=visual)
    write(out/'smoke/PACKAGE-CHECK.json', check)
    source_paths = {ROOT/name for name in build['source_files'] if not name.startswith('vendor/')}
    source_paths.update(p for p in (ROOT/'assets/hd').rglob('*') if p.is_file())
    source_paths.update(ROOT/'tools/scripts'/name for name in (
        'faction_delivery_finalize.py','faction_story_promo.py','ai_gameplay_record.py','ai_promo_model.go',
        'strategy_ai_promo_model.go','faction_hd_check.py','faction_hd_game_browser_check.py',
        'v1_release.py','v1_release_smoke.py','local_test_package.py','local_test_macos_build.sh',
        'local_test_macos_package.py','local_test_smoke.py','local_test_windows_smoke.py','bundled_scenarios_smoke.py'))
    source_paths.update(ROOT/name for name in ('docs/playtest/115-faction-hd-ai-promo.md',
        'docs/spec/134-enemy-visibility-display.md','docs/spec/135-original-style-hd-ui.md',
        'docs/internals/data/alliance-fireline-theme.json','docs/internals/data/faction-themes.json'))
    source_paths.add(batch/'FACTION-HD-GAME.html')
    sources = [record(p) for p in sorted(source_paths)]
    snapshot = batch/'delivery115-source'
    for p in sorted(source_paths):
        copy(p, snapshot/p.relative_to(ROOT))
    evidence = [record(p) for folder in (out/'smoke',out/'promo') for p in sorted(folder.rglob('*')) if p.is_file()]
    index = dict(version=version, status='PASS 赤焰戰線與雙陣營HD／實錄AI影片本機交付', source_head=build['head'],
                 dirty_source_snapshot=True, frozen_build_source=str(frozen.relative_to(ROOT)),
                 frozen_delivery_source=str(snapshot.relative_to(ROOT)), delivery=str(out.relative_to(ROOT)),
                 sources=sources, evidence=evidence, packages=build['packages'], verification=check,
                 limits=['DOS全戰場same-state','AI必勝','完整登陸分段流程#48','Windows／macOS真機','人耳驗收','逐幀AI新繪'],
                 rights='完整版／HD／影片本機限定；引擎包不含原版與HD衍生素材，RRSAL-1.0', remote_release=False)
    write(ROOT/'docs/internals/data/faction-hd-ai-delivery.json', index)
    write(out/'DELIVERY.json', index)
    checksums = build['packages']+[dict(path=str(p.relative_to(out)), bytes=p.stat().st_size, sha256=sha(p))
        for p in sorted((out/'promo').glob('*.mp4'))]
    write(out/'SHA256SUMS.json', dict(version=version, files=checksums))
    print('PASS 68正常視窗、32完整實錄視窗、五段非靜音遊玩、118秒HD故事影片與八封包', flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--version', default='v.1.1.2-20261011')
    main(parser.parse_args().version)
