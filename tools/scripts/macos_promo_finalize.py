"""#94／95：核對三平台與72秒影片，集中交付，另凍結可回查的來源身份。"""
import argparse
from array import array
import hashlib
import json
import math
import os
from pathlib import Path
import re
import shutil
import tarfile
import zipfile


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda: f.read(1048576), b''):
            h.update(block)
    return h.hexdigest()


def entry(path):
    assert path.stat().st_uid == os.getuid() and path.stat().st_gid == os.getgid(), path
    return {'path': str(path), 'bytes': path.stat().st_size, 'sha256': sha(path)}


def write(path, data):
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')


def validate_packages(root):
    build = json.loads((root/'BUILD.json').read_text())
    for name, expected in build['source_files'].items():
        if name.startswith(('cmd/', 'internal/')) or name in ('go.mod', 'go.sum', 'LICENSE'):
            assert sha(Path(name)) == expected['sha256'], name
    for item in build['packages']:
        p = root/item['path']
        assert (p.stat().st_size, sha(p)) == (item['bytes'], item['sha256']), p
        if p.suffix == '.zip':
            with zipfile.ZipFile(p) as z:
                prefix = p.stem+'/'
                manifest = json.loads(z.read(prefix+'MANIFEST.json'))
                assert manifest['version'] == build['version']
                expected_names = {prefix+name for name in manifest['files']} | {prefix+'MANIFEST.json'}
                assert set(z.namelist()) == expected_names and len(z.namelist()) == len(expected_names)
                for name, expected in manifest['files'].items():
                    data = z.read(prefix+name)  # 讀取同時核對ZIP CRC。
                    assert len(data) == expected['bytes'] and hashlib.sha256(data).hexdigest() == expected['sha256']
                assert any(n.endswith('/licenses/Go-LICENSE') for n in expected_names)
                for i in z.infolist():
                    assert not Path(i.filename).is_absolute() and '..' not in Path(i.filename).parts
                    assert i.filename.isascii() or i.flag_bits & 0x800
        else:
            with tarfile.open(p) as t:
                found = {m.name: m for m in t.getmembers() if m.isfile()}
                for name, expected in build['source_files'].items():
                    member = found['logh3-source/'+name]
                    data = t.extractfile(member).read()
                    assert len(data) == expected['bytes'] and hashlib.sha256(data).hexdigest() == expected['sha256']
    return build


def copy_receipts(source, target, skip_package=False):
    assert source.is_dir() and not target.exists()
    target.mkdir()
    for p in source.rglob('*'):
        if not p.is_file() or p.suffix not in ('.json', '.png', '.txt', '.log'):
            continue
        rel = p.relative_to(source)
        if skip_package and rel.parts[0].startswith('LOGH3-'):
            continue
        q = target/rel
        q.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(p, q)
        assert sha(p) == sha(q)


def finalize(root, revision):
    assert not (root/'FINALIZE.json').exists()
    version = root.name
    build = validate_packages(root)
    linux = Path('workplace/package92-smoke-'+version)
    windows = Path('workplace/music69-windows/package92-smoke-'+version)
    for source, count in ((linux, 11), (windows, 4)):
        d = json.loads((source/'SMOKE.json').read_text())
        assert d['version'] == version and len(d['runs']) == count
    assert json.loads((windows/'SMOKE.json').read_text())['archive_sha256'] == build['packages'][0]['sha256']
    mac = json.loads((root/'smoke/macos/MACOS.json').read_text())
    assert mac == build['macos'] and mac['zip_integrity'] and mac['executable_permissions_preserved']
    for source, target, skip in ((linux, 'linux', False), (windows, 'windows', True),
                                 (Path('workplace/package92-r119-smoke'), 'audio', False)):
        copy_receipts(source, root/'smoke'/target, skip)
    standard = Path('workplace/round119-standard.log')
    lines = standard.read_text().splitlines()
    assert sum(line.startswith('ok  ') for line in lines) == 13
    assert sum('[no test files]' in line for line in lines) == 5 and not any(line.startswith('FAIL') for line in lines)
    shutil.copyfile(standard, root/'smoke/standard.log')
    monitor = Path('workplace/package92-r119-smoke/monitor.f32le')
    samples = array('f', monitor.read_bytes())
    assert samples and all(math.isfinite(v) for v in samples)
    peak = max(abs(v) for v in samples)
    rms = math.sqrt(sum(v*v for v in samples)/len(samples))
    assert 0 < peak < 1 and rms > 0
    audio_stats = {'input': entry(monitor), 'seconds': len(samples)/44100/2,
                   'peak': peak, 'rms': rms, 'classification': '正式Linux封包虛擬音效卡，非人耳／macOS真機'}
    write(root/'smoke/audio/AUDIO.json', audio_stats)
    promo = Path('workplace/promo95-'+version)/('render-'+revision)
    probe = json.loads((promo/'FFPROBE.json').read_text())
    video = next(s for s in probe['streams'] if s['codec_type'] == 'video')
    sound = next(s for s in probe['streams'] if s['codec_type'] == 'audio')
    assert (video['width'], video['height'], video['codec_name'], int(video['nb_frames'])) == (1920, 1080, 'h264', 2160)
    assert video['r_frame_rate'] == '30/1' and sound['codec_name'] == 'aac' and sound['channels'] == 2
    assert abs(float(video['duration'])-72) < 0.01 and abs(float(probe['format']['duration'])-72) < 0.1
    audio_log = (promo/'AUDIO.txt').read_text()
    mean = float(re.search(r'mean_volume: ([-0-9.]+) dB', audio_log).group(1))
    maximum = float(re.search(r'max_volume: ([-0-9.]+) dB', audio_log).group(1))
    assert -35 < mean < -10 and maximum < -1 and 'silence_start:' not in audio_log
    video_log = (promo/'VIDEO.txt').read_text()
    assert 'black_start:' not in video_log
    freezes = [float(v) for v in re.findall(r'freeze_start: ([0-9.]+)', video_log)]
    assert len(freezes) == 9 and all(t >= 15 for t in freezes)
    # 九段靜態展示及兩段動態為分鏡既定設計；不把静態畫面當遊戲凍結。
    review = json.loads((promo/'REVIEW.json').read_text())
    assert review['all_11_segments_reviewed'] and review['subtitle_crop'] == 0
    assert review['motion_verified'] and review['previous_current_reviewed']
    assert not any((root/'promo').iterdir())
    keep = {'FFPROBE.json', 'AUDIO.txt', 'VIDEO.txt', 'RIGHTS.txt', 'PLAN.json', 'theme.sh',
            'cards-contact.png', 'samples-contact.png', 'previous-current.png', 'REVIEW.json', 'CACHE.txt', 'MOTION.txt'}
    for p in promo.iterdir():
        if p.name in keep or p.name.startswith('sample-') or p.name == 'LOGH3-'+version+'-promo.mp4':
            shutil.copyfile(p, root/'promo'/p.name)
    write(root/'FINALIZE.json', {'version': version, 'packages': build['packages'],
          'standard_pass': 13, 'standard_no_test_files': 5, 'linux_runs': 11, 'wine_runs': 4,
          'macos': '雙架構、系統依賴、ARM全2994頁簽章、ZIP及資產通過；未真機執行',
          'audio': audio_stats, 'promo': {'seconds': 72, 'frames': 2160, 'layouts': 11,
          'mean_db': mean, 'max_db': maximum, 'black_segments': 0, 'static_segments': 9,
          'dynamic_segments': 2, 'subtitle_crop': 0, 'comparison': 'nearby-state'},
          'rights': '含原版音畫，只留本機；未建立tag或Release',
          'limits': ['macOS與Windows實機、人耳、完整音樂循環仍待驗收', '完整登陸分段演出未接完', '未新增雙玩家或戰場差異驗證']})
    files = []
    for p in sorted(root.rglob('*')):
        if p.is_file() and (p.suffix in ('.zip', '.gz', '.mp4') or p.parent == root or p.parent == root/'promo'):
            if p.name == 'SHA256SUMS.json':
                continue
            e = entry(p)
            e['path'] = str(p.relative_to(root))
            files.append(e)
    write(root/'SHA256SUMS.json', {'version': version, 'files': files})
    print('PASS 三平台ZIP／来源、Linux11、Wine4、Mac架構及簽章、72秒音畫影片；交付身份', len(files))


def receipt(root):
    output = Path('docs/internals/data/macos-retro-promo.json')
    assert not output.exists()
    frozen = json.loads(Path('docs/internals/data/local-test-package.json').read_text())
    for e in frozen['source_files']:
        assert sha(Path(e['path'])) == e['sha256'], e['path']+'第118輪來源被改'
    for base in (root, Path('dist-all/v.0.1.1-20261006')):
        for e in json.loads((base/'SHA256SUMS.json').read_text())['files']:
            assert (base/e['path']).stat().st_size == e['bytes'] and sha(base/e['path']) == e['sha256']
    sources = ['README.md', 'docs/playtest/106-macos-and-retro-promo.md',
               'docs/goal/2026-10-06-macos-retro-promo-round-119.md', 'tools/macos.sh',
               'tools/scripts/local_test_macos_build.sh', 'tools/scripts/local_test_macos_package.py',
               'tools/scripts/retro_promo_prepare.py', 'tools/scripts/retro_promo_theme.sh',
               'tools/scripts/retro_promo_render.sh', 'tools/scripts/macos_promo_finalize.py']
    artifacts = [p for p in (root/'smoke').rglob('*') if p.is_file()]
    artifacts += [p for p in Path('workplace/macos94-r119a').iterdir() if p.suffix == '.txt']
    artifacts += [Path('workplace')/name for name in ('round119-base-build.log', 'round119-macos-build.log',
                   'round119-macos-package-final.log', 'round119-linux-smoke.log', 'round119-windows-smoke-final.log',
                   'round119-promo-cards-final.log', 'round119-promo-encode-final.log', 'round119-promo-audit-final.log',
                   'round119-finalize-final.log', 'round119-standard.log')]
    images = dict(frozen['toolchain_images'])
    images['psychicwar-osxcross:latest'] = 'sha256:0f50c77c732087b104b3aadc3aecb352054aa37dee8e1d293523a66dfa02ba61'
    write(output, {'schema': 'logh3.macos-retro-promo.v1', 'version': root.name, 'date': '2026-10-06',
          'source_revision': json.loads((root/'BUILD.json').read_text())['head'],
          'sources': [entry(Path(p)) for p in sources], 'artifacts': [entry(p) for p in sorted(artifacts)],
          'delivery_files': json.loads((root/'SHA256SUMS.json').read_text())['files'],
          'result': json.loads((root/'FINALIZE.json').read_text()), 'toolchain_images': images,
          'macos': json.loads((root/'smoke/macos/MACOS.json').read_text()),
          'promo': json.loads((root/'promo/PLAN.json').read_text()),
          'frozen_previous_delivery_preserved': True, 'previous_receipt': entry(Path('docs/internals/data/local-test-package.json')),
          'technical_sources': ['https://go.dev/doc/go1.24', 'https://go.dev/wiki/Darwin',
             'https://help.apple.com/pdf/security/en_GB/apple-platform-security-guide-b.pdf',
             'https://support.apple.com/102445'],
          'user_scope': 'macOS本機完整版、依復古重製技巧重做影片及README故事開頭；無新玩法、戰場差異或公開發行'})
    print('PASS 新版來源與交付身份已凍結，前版保持', entry(output))


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--version', required=True)
    p.add_argument('--revision', default='c')
    p.add_argument('mode', choices=('finalize', 'receipt'))
    a = p.parse_args()
    assert Path.cwd() == Path('/src') and os.getuid() == 1000
    assert re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}', a.version)
    assert re.fullmatch('[a-z0-9]+', a.revision)
    root = Path('dist-all')/a.version
    assert root.is_dir() and root.stat().st_uid == os.getuid()
    if a.mode == 'finalize':
        finalize(root, a.revision)
    else:
        receipt(root)
