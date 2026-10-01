"""研究 362：合併 Windows 音樂探針與反例；不產生正式發行聲明。"""
import argparse
import array
import hashlib
import json
import math
import os
from pathlib import Path

root = Path('/work/workplace/music69-windows')


def sha(p):
    assert p.is_file() and not p.is_symlink()
    assert (p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid())
    with p.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def read(name):
    return json.loads((root/name).read_text())


def packets(data):
    result, buf, off = [], b'', 0
    while off < len(data):
        assert data[off:off+4] == b'OggS'
        count = data[off+26]
        sizes = data[off+27:off+27+count]
        pos = off+27+count
        for size in sizes:
            buf += data[pos:pos+size]
            pos += size
            if size < 255:
                result.append(buf)
                buf = b''
        off = pos
    assert not buf and off == len(data)
    return result


args = argparse.ArgumentParser()
args.add_argument('--current-cross-image', required=True)
options = args.parse_args()
fm = read('pre-pcm-gate/fm-build.json')
failure = read('imported/FM-SOURCES.json')
assert failure['status'] == 'failed' and 'OGG SHA-256' in failure['error']
assert failure['probe_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
for name, record in fm['files'].items():
    p = root/'封裝 測試/tools'/name
    if p.exists():
        assert sha(p) == record['sha256'], name
for windows_path, expected in failure['tools_sha256'].items():
    assert sha(root/'封裝 測試/tools'/windows_path.rsplit('\\', 1)[-1]) == expected

diag = root/'diagnostic'
wav = read('diagnostic/pcm-comparison.json')
volume = read('diagnostic/volume-comparison.json')
assert volume['equal'] and wav['win'] == wav['linux']
for side in ['win', 'linux']:
    assert sha(diag/f'BGM_0F-{side}.wav') == wav[side]['wav_sha256']
    assert sha(diag/f'BGM_0F-{side}-volume.f32') == volume[side]['sha256']
encoded = {}
streams = {}
for side in ['win', 'linux', 'win-math']:
    p = diag/f'BGM_0F-{side}-fixed.ogg'
    streams[side] = packets(p.read_bytes())
    encoded[side] = {'sha256': sha(p), 'bytes': p.stat().st_size, 'packets': len(streams[side])}
    encoded[side]['headers_equal'] = streams[side][:3] == streams['win'][:3]
for side in ['win', 'win-math']:
    assert len(streams[side]) == len(streams['linux'])
    encoded[side]['different_packets'] = sum(a != b for a, b in zip(streams[side], streams['linux']))
assert encoded['win']['different_packets'] == 260
assert encoded['win-math']['different_packets'] == 232

a, b = array.array('f'), array.array('f')
a.frombytes((diag/'BGM_0F-win-decoded.f32').read_bytes())
b.frombytes((diag/'BGM_0F-linux-decoded.f32').read_bytes())
assert len(a) == len(b)
count, squared, largest = 0, 0., 0.
for x, y in zip(a, b):
    e = abs(x-y)
    count += e != 0
    squared += e*e
    largest = max(largest, e)
decoded = {'samples': len(a), 'different_samples': count, 'max_abs_difference': largest,
           'rmse': math.sqrt(squared/len(a)),
           'files_sha256': {s: sha(diag/f'BGM_0F-{s}-decoded.f32') for s in ['win', 'linux']}}
previous = read('diagnostic/decoded-difference.json')
assert decoded['samples'] == previous['samples'] and decoded['different_samples'] == previous['different_samples']
for key in ['max_abs_difference', 'rmse']:
    assert math.isclose(decoded[key], previous[key], rel_tol=1e-10, abs_tol=1e-12), (key, decoded[key], previous[key])

gm = read('gm-full/GM-SOURCES.json')
assert gm['status'] == 'complete-gm-stage' and gm['platform'] == 'windows/amd64' and len(gm['tracks']) == 11
baseline = json.loads(Path('/work/workplace/music69-fm/full/GM-SOURCES.json').read_text())
old = {r['track']: r for r in baseline['tracks']}
for track in gm['tracks']:
    assert sha(root/'gm-full'/f'{track["track"]}.ogg') == track['ogg_sha256']
    assert track['midi_sha256'] == old[track['track']]['midi_sha256']
    track['duration_difference_from_linux'] = track['seconds']-old[track['track']]['seconds']
assert '--- PASS: TestWindowsGMImportPlayerDecode' in (root/'player-test.log').read_text()
sdk = root/'fluidsynth-2.3.1-win10-x64.zip'
assert sha(sdk) == '7bf4ae8ae1856376266280bf680d465a96d3f66e2b875cdd7f3539b549c4d702'
assets = read('fluidsynth-release.json')['assets']
url = next(a['browser_download_url'] for a in assets if a['name'] == sdk.name)
sdk_files = read('gm-sdk-files.json')
for name, record in sdk_files.items():
    assert sha(root/'gm-probe/tools'/name) == record['sha256']

script_names = ['music_windows_fm_probe.py', 'music_windows_vorbis_probe.py',
                'music_windows_f32_probe.py', 'music_windows_receipt.py']
result = {
    'schema': 'logh3.music-tools.windows.probes.v1', 'date': '2026-10-01',
    'status': 'Windows FM 反例與 GM 十一首技術探針完成；正式封裝未完成',
    'source_parent_head': '456098ca4b48da01b03f9729ec787bc42c1a0df0', 'source_dirty': True,
    'images': {'build_cross': 'sha256:829642b4f69c06c48deeeb90b558b90b08e29c6cfeaa3162f120e63f4b51eadb',
               'current_pinned_cross': options.current_cross_image,
               'wine': 'sha256:7d3c7ff554db39300e3e7a92a7b79c2d5fcd647289596d22fa10899baac61201',
               'linux_audio': 'sha256:da8c479c3466feed5cd363ed439b064303d68c9789abc02a37fb09560285b126',
               'go': 'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac',
               'player_test_gui': 'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8'},
    'fm_build': fm, 'fm_failure': failure, 'waveform': wav, 'post_volume_pcm': volume,
    'encoded': encoded, 'decoded_difference': decoded, 'math_probe': read('vorbis-math-build.json'),
    'gm': gm, 'gm_sdk': {'url': url, 'sha256': sha(sdk), 'bytes': sdk.stat().st_size,
                        'files': sdk_files, 'complete_licenses_and_sources_verified': False},
    'buildinfo': {n: (root/f'{n}-buildinfo.txt').read_text() for n in ['importer', 'probe']},
    'test_log_sha256': sha(root/'player-test.log'),
    'scripts_sha256': {n: sha(Path('/work/tools/scripts')/n) for n in script_names},
    'diagnostic_files': {p.name: {'sha256': sha(p), 'bytes': p.stat().st_size} for p in diag.iterdir() if p.is_file()},
    'limits': '此探針使用舊 FM 完整 OGG 雜湊門檻；數學路徑試驗仍不符。2026-10-01 使用者另裁決改驗 VGM／WAV／增益後 PCM 相同，允許平台 OGG 差異；正式接線由規格 074 驗收。SDK 與 C++ runtime 隔離，未作正式來源／授權交付。只驗 Wine 原生 CLI 及 Linux 正式播放層解碼，不宣稱 Windows 實機／音效裝置／人耳／循環接點。'}
target = Path('/work/docs/internals/data/windows-music-probes.json')
assert target.parent.stat().st_uid == os.getuid()
if target.exists():
    sha(target)
target.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
print('Windows 探針收據 SHA-256', sha(target))
