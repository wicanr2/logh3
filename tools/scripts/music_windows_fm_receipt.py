"""規格 074：只收實際完成的 Windows 十一首／三路重跑及 Linux 抽驗。

LOGH3_EXTRA_RO 指向本專案，tools/py.sh tools/scripts/music_windows_fm_receipt.py。
不替換來源、不複製舊曲庫為原版收據；大型音檔只留 workplace。
"""
import hashlib
import json
import os
from pathlib import Path
import struct
import sys

root = Path('/ref')
work = Path('/work/workplace/music69-windows')
docs = Path('/work/docs')


def sha(path):
    assert path.is_file() and not path.is_symlink(), path
    assert (path.stat().st_uid, path.stat().st_gid) == (os.getuid(), os.getgid()), path
    with path.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def load(path):
    return json.loads(path.read_text())


lock_path = docs/'internals/data/fm-import-lock.json'
lock = load(lock_path)
fm = load(work/'fm-full/FM-SOURCES.json')
repeat = load(work/'fm-repeat/FM-SOURCES.json')
linux = load(work/'linux-pcm-sampling/FM-SOURCES.json')
sample = {'BGM_0F', 'END_2F', 'OPEN_F'}
assert fm['status'] == repeat['status'] == linux['status'] == 'complete-fm-stage'
assert fm['platform'] == repeat['platform'] == 'windows/amd64'
assert linux['platform'] == 'linux/amd64'
assert {t['track'] for t in fm['tracks']} == set(lock['tracks']) and len(fm['tracks']) == 11
assert {t['track'] for t in repeat['tracks']} == {t['track'] for t in linux['tracks']} == sample
assert len(repeat['tracks']) == len(linux['tracks']) == 3
bundle = root/'org_game/银河英雄传说3SP.zip'
assert sha(bundle) == 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673'
for receipt in [fm, repeat, linux]:
    assert receipt['probe_head'] == lock['probe_head']
    assert receipt['source_executable_sha256'] == lock['executables']
    for name, expected in receipt['source_executable_sha256'].items():
        assert sha(Path('/orig')/name) == expected

files = load(work/'fm-pcm-imports.json')
package = work/'pcm-gate/封裝 測試'
for name, metadata in files.items():
    p = package/name if name == 'music-import.exe' else package/'tools'/name
    assert sha(p) == metadata['sha256']
assert sha(package/'music-import.exe') == fm['program_sha256'] == repeat['program_sha256']
assert fm['tools_sha256'] == repeat['tools_sha256']
for path, expected in fm['tools_sha256'].items():
    assert sha(package/'tools'/path.rsplit('\\', 1)[-1]) == expected

all_tracks = {t['track']: t for t in fm['tracks']}
platform_differences = {}
for dirname, receipt in [('fm-full', fm), ('fm-repeat', repeat), ('linux-pcm-sampling', linux)]:
    for track in receipt['tracks']:
        name = track['track']
        reference = lock['tracks'][name]
        for key in ['source_sha256', 'vgm_sha256', 'wav_sha256', 'pcm_sha256']:
            assert track[key] == reference[key], (dirname, name, key)
        assert track['gain_db'] == '-1.00' and track['sample_rate'] == 44100
        assert sha(Path('/orig')/(name+'.MDS')) == reference['source_sha256']
        assert sha(work/dirname/(name+'.vgm')) == reference['vgm_sha256']
        ogg = work/dirname/(name+'.ogg')
        assert sha(ogg) == track['ogg_sha256']
        assert struct.unpack_from('<I', ogg.read_bytes(), 14)[0] == reference['serial']
        assert track['seconds'] > 0 and 0 < track['peak'] <= 1
        if dirname == 'fm-repeat':
            for key in ['vgm_sha256', 'wav_sha256', 'pcm_sha256', 'ogg_sha256']:
                assert track[key] == all_tracks[name][key]
            for suffix in ['.vgm', '.ogg']:
                assert (work/dirname/(name+suffix)).read_bytes() == (work/'fm-full'/(name+suffix)).read_bytes()
        if dirname == 'linux-pcm-sampling':
            assert track['ogg_sha256'] == reference['ogg_sha256']
        if dirname == 'fm-full':
            platform_differences[name] = {
                'windows_ogg_sha256': track['ogg_sha256'],
                'linux_historical_ogg_sha256': reference['ogg_sha256'],
                'ogg_equal': track['ogg_sha256'] == reference['ogg_sha256'],
                'lossless_stages_equal': True,
            }
    assert not list((work/dirname).glob('.fm-import-*'))

test_log = (work/'player-pcm-test.log').read_text()
for test in ['TestWindowsFMImportPlayerDecode', 'TestWindowsGMImportPlayerDecode', 'TestPlayerPathMusicSwitches']:
    assert '--- PASS: '+test in test_log
products = (work/'product-tests.log').read_text()
assert 'FAIL' not in products and 'github.com/wicanr2/logh3/cmd/gin3' in products
assert sha(work/'music-import-pcm-linux') == linux['program_sha256']
build = load(work/'fm-pcm-build.json')
meta = load(work/'pcm-runmeta.json')
assert meta['probe_head'] == lock['probe_head'] and meta['probe_clean']
assert meta['program_sha256'] == fm['program_sha256']
scripts = ['music_windows_fm_probe.py', 'music_windows_fm_receipt.py',
           'music_fm_lock.py', 'music_fm_pcm_baseline.sh']
result = {
    'schema': 'logh3.music-tools.windows.fm-pcm.v1', 'date': '2026-10-01',
    'scope': 'Windows 原生十一首 FM 無損閘門與三路重跑；Linux 三路回歸；正式播放器解碼',
    'user_decision': 'VGM／WAV／增益後 PCM 相同，允許各平台 OGG 差異',
    'original_bundle_sha256': sha(bundle), 'fm_lock_sha256': sha(lock_path),
    'run': meta, 'fm_build': build, 'package_files': files,
    'receipt_runtime': {'python': sys.version, 'image': 'sha256:229a2c5bfa27522db7815ea81f9bed70af17ccb9de9fc7ad142b1877b5830d36'},
    'fm': fm, 'repeat': repeat, 'linux_sampling': linux,
    'platform_differences': platform_differences,
    'scripts_sha256': {n: sha(Path('/work/tools/scripts')/n) for n in scripts},
    'source_sha256': {n: sha(root/n) for n in ['cmd/music-import/fm.go', 'cmd/music-import/fmdata.go', 'cmd/music-import/fm_test.go', 'cmd/gin3/music_integration_test.go']},
    'receipts_sha256': {n: sha(work/n) for n in ['fm-full/FM-SOURCES.json', 'fm-repeat/FM-SOURCES.json', 'linux-pcm-sampling/FM-SOURCES.json', 'player-pcm-test.log', 'product-tests.log']},
    'buildinfo': (work/'importer-pcm-buildinfo.txt').read_text(),
    'limits': '原版 OPL 串已驗；合成仍為 hardware-spec approximation，OPEN_F 為 substituted。只驗 Wine 原生 CLI 與 Linux 正式播放層，不宣稱實機音效、人耳、循環、macOS 或正式來源／授權交付已完成。',
}
target = docs/'internals/data/windows-fm-pcm.json'
assert target.parent.stat().st_uid == os.getuid()
if target.exists():
    sha(target)
target.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
print('Windows FM 無損閘門收據 SHA-256', sha(target))
