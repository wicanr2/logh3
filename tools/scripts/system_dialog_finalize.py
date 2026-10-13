"""第131輪：驗證正式封包、正常玩家收據與原曲HD來源，保存唯一交付索引。"""
import argparse
import hashlib
import io
import json
import os
from pathlib import Path
import shutil
import subprocess
import tarfile

ROOT = Path('/src')
VERSION = 'v.1.3.0-20261013'
BATCH = ROOT / 'workplace/modal131'
OUT = ROOT / 'dist-all' / VERSION


def read(path):
    return json.loads(path.read_text())


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b''):
            h.update(chunk)
    return h.hexdigest()


def record(path):
    assert path.is_file() and path.stat().st_uid == os.getuid(), path
    return dict(path=str(path.relative_to(ROOT)), bytes=path.stat().st_size, sha256=sha(path))


def write(path, value):
    assert path.parent.is_dir() and path.parent.stat().st_uid == os.getuid()
    if path.exists():
        assert path.stat().st_uid == os.getuid()
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def main(commit=None):
    build = read(OUT / 'BUILD.json')
    assert build['version'] == VERSION and len(build['packages']) == 8
    packages = []
    for item in build['packages']:
        p = OUT / item['path']
        assert p.stat().st_size == item['bytes'] and sha(p) == item['sha256'], p
        packages.append(dict(item, classification='僅本機完整版' if item['path'].startswith('full/') else '無原版素材的引擎／來源包'))
    frozen = ROOT / ('workplace/package92-' + VERSION + '-final/source')
    sources = []
    for name, expected in build['source_files'].items():
        assert sha(frozen / name) == expected['sha256'], name
        if not name.startswith('vendor/'):
            assert sha(ROOT / name) == expected['sha256'], name
            sources.append(record(ROOT / name))
    assert len(sources) == 575
    evidence = []
    normal = []
    for faction in (0, 1):
        folder = BATCH / ('full-f' + str(faction))
        r = read(folder / 'RESULT.json')
        assert r['passed'] and r['normal_entry'] and not r['script_flag'] and not r['state_injection']
        assert r['quit_exit_code'] == 0 and len(r['save_sha256']) == 6
        assert len(r['cursor_checks']) == 6
        normal.append(dict(faction=faction, passed=True, cursor_checks=r['cursor_checks'],
                           elapsed_seconds=r['elapsed_seconds'], receipt=str((folder / 'RESULT.json').relative_to(ROOT))))
        for p in sorted(folder.glob('*.png')):
            evidence.append(record(p))
        for name in ['RESULT.json', 'FRAMES.jsonl', 'AUDIO.wav']:
            evidence.append(record(folder / name))
    windows = read(ROOT / 'workplace/music69-windows/modal131-windows/RESULT.json')
    assert windows['passed'] and len(windows['runs']) == 4
    engine = []
    for p in [OUT / 'smoke/release-linux/SMOKE.json', ROOT / ('workplace/music69-windows/release-smoke-' + VERSION + '/SMOKE.json')]:
        r = read(p)
        assert r['external_original_only'] and len(r['runs']) == 4
        engine.append(dict(platform=r['platform'], runs=4, diagnostic=True, receipt=record(p)))
        evidence.append(record(p))
    error = read(BATCH / 'startup-error/RESULT.json')
    assert error['passed'] and error['escape_exit'] == 1
    tests = (BATCH / 'TESTS-FINAL.log').read_text().splitlines()
    assert sum(line.startswith('ok ') for line in tests) == 15
    assert sum('[no test files]' in line for line in tests) == 4
    music = read(BATCH / 'music-hd/MANIFEST.json')
    midi = read(BATCH / 'MUSIC-DURATION-CHECK.json')
    assert len(music['files']) == len(midi) == 11 and all(i['midi_same'] for i in midi)
    for item in music['files']:
        assert sha(BATCH / 'music-hd' / item['file']) == item['sha256']
        assert item['peak_db'] < -0.1 and item['loop_boundary_jump'] < 0.01 and item['duration'] > 0
    for name in ['TESTS-FINAL.log', 'PACKAGE.log', 'MACOS-BUILD.log', 'MACOS-PACKAGE.log',
                 'RELEASE-PACKAGE.log', 'WINDOWS-SMOKE.log', 'ENGINE-LINUX.log', 'ENGINE-WINDOWS.log',
                 'MUSIC-DURATION-CHECK.json', 'music-hd/MANIFEST.json', 'music-source/SOURCE.json',
                 'startup-error/RESULT.json', 'startup-error/error.png']:
        evidence.append(record(BATCH / name))
    evidence.append(record(ROOT / 'workplace/music69-windows/modal131-windows/RESULT.json'))
    evidence.append(record(OUT / 'BUILD.json'))
    evidence.append(record(OUT / 'smoke/macos/MACOS.json'))
    images = ROOT / 'docs/images'
    assert images.stat().st_uid == os.getuid()
    shots = []
    for src, name in [('full-f0/00-loading.png', 'loading'), ('full-f0/settings-serif.png', 'empire-font'),
                      ('full-f1/settings-sans.png', 'alliance-font')]:
        p = images / ('readme-v130-' + name + '.png')
        shutil.copyfile(BATCH / src, p)
        shots.append(dict(record(p), source=record(BATCH / src), normal_entry=True, cropped=False, resized=False))
    write(images / 'readme-v130.json', dict(version=VERSION, shots=shots,
        binary=record(OUT / ('full/LOGH3-' + VERSION + '-linux-amd64/gin3')),
        normal_receipts=[n['receipt'] for n in normal]))
    approved = []
    for item in read(BATCH / 'font-preview/MANIFEST.json'):
        source = BATCH / 'font-preview' / item['file']
        assert sha(source) == item['sha256']
        target = images / ('font-v130-' + item['file'])
        shutil.copyfile(source, target)
        approved.append(dict(item, saved=record(target)))
    html = (BATCH / 'font-preview/index.html').read_text()
    for name in ['classic', 'serif', 'sans']:
        html = html.replace('src="' + name + '.png"', 'src="font-v130-' + name + '.png"')
    (images / 'font-v130-preview.html').write_text(html)
    write(images / 'font-v130-preview.json', dict(classification='已核准的可丟棄試樣，保留原提案文字，不代表正式遊戲截圖',
        decision='使用者確認三款加入、原版強化預設，追加保留原版點陣，共四款。', variants=approved))
    extra = ['tools/audio.sh', 'tools/scripts/hd_font_atlas.py', 'tools/scripts/startup_font.py',
             'tools/scripts/hd_music_export.py', 'tools/scripts/hd_music_render.sh', 'tools/scripts/hd_music_check.py',
             'tools/scripts/system_dialog_gui_check.py', 'tools/scripts/system_dialog_finalize.py',
             'tools/scripts/local_test_package.py', 'tools/scripts/local_test_macos_package.py',
             'tools/scripts/v1_release.py', 'tools/scripts/v1_release_smoke.py', 'tools/scripts/release_publish_check.py']
    sources += [record(ROOT / name) for name in extra]
    previous = ROOT / 'docs/internals/data/system-dialog-polish.json'
    old = read(previous) if previous.exists() else {}
    data = dict(schema=1, version=VERSION, status='PASS 本機交付與機械驗證；HD聽感及真機保留#69',
        build_head=build['head'], build_working_tree='當時含未提交修改；1553項凍結檔符合BUILD，575項非vendor與工作樹相同；另以commit收據綁定正式來源',
        frozen_source_files=1553, non_vendor_source_files=575, sources=sources, packages=packages,
        normal_gui=normal, windows=dict(runs=windows['runs'], passed=True, audio=False, real_hardware=False),
        engine_smoke=engine, tests=dict(command='tools/gui.sh test ./internal/... ./cmd/...', passed=15, no_tests=4),
        verified_features=['讀檔／保存／退出即時單游標與原圖／HD', '主選單讀檔右鍵返回主選單',
            '正常兩陣營三劇本及六檔保存重讀', '四字型切換與F7返回保留選擇', 'HD／原音色與靜音正常切換',
            '真實七階段載入視窗與錯誤退出'],
        music=dict(tracks=11, original_midi_bytes_unchanged=True, source_manifest=record(BATCH / 'music-hd/MANIFEST.json'),
            maximum_sample_peak_db=max(i['peak_db'] for i in music['files']),
            maximum_loop_edge_jump=max(i['loop_boundary_jump'] for i in music['files']), human_listening=False),
        evidence=evidence, rights='完整版、原字庫、HD素材與原曲衍生OGG只留本機；GitHub只有引擎與來源；內嵌啟動字集為授權Noto衍生',
        limits=['不是DOS同狀態對拍', '戰術讀檔取消只有回歸測試，未補正常戰術GUI',
                'Windows Wine抽樣音訊停用；Windows／macOS真機未驗', '人耳、完整曲目播放及音樂性循環待#69',
                '本輪沒有重製推廣片，v.1.2.0影片保持原身份'])
    for key in ['published_assets', 'published_checksum', 'remote_release', 'source_commit_check']:
        if key in old:
            data[key] = old[key]
    if commit:
        assert len(commit) == 40 and all(c in '0123456789abcdef' for c in commit)
        blob = subprocess.check_output(['git', '-C', str(ROOT), 'archive', commit, '--',
            'LICENSE', 'go.mod', 'go.sum', 'cmd', 'internal', 'docs/internals/data/fm-import-lock.json'])
        wanted = {name: item for name, item in build['source_files'].items() if not name.startswith('vendor/')}
        with tarfile.open(fileobj=io.BytesIO(blob)) as archive:
            members = [m for m in archive.getmembers() if m.isfile()]
            assert {m.name for m in members} == set(wanted)
            for member in members:
                assert hashlib.sha256(archive.extractfile(member).read()).hexdigest() == wanted[member.name]['sha256'], member.name
        proof = dict(passed=True, commit=commit, files_checked=len(wanted),
                     version=VERSION, method='git archive commit的575項非vendor與BUILD凍結雜湊相同')
        write(BATCH / 'COMMIT-SOURCE-CHECK.json', proof)
        data['source_commit_check'] = dict(proof, receipt=record(BATCH / 'COMMIT-SOURCE-CHECK.json'))
    write(previous, data)
    lines = [i['sha256'] + '  ' + Path(i['path']).name for i in build['packages'] if i['path'].startswith('release/')]
    (OUT / 'release/SHA256SUMS.txt').write_text('\n'.join(lines) + '\n')
    print('PASS 1553凍結來源、八封包、兩正常GUI、Windows四入口、兩平台引擎、四字型與11HD曲目')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--commit', help='核對待發布的正式Git來源commit')
    main(parser.parse_args().commit)
