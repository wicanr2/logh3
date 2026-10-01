"""規格 075：合併已完成的固定來源／Unicode 路徑／正式曲庫技術收據。

LOGH3_EXTRA_RO 指向專案，tools/py.sh /work/tools/scripts/music_windows_gm_source_receipt.py。
未完成批次、不同工具或未通過測試不能產生完成收據。
"""
import hashlib
import json
import os
from pathlib import Path
import struct
import sys
import wave

root = Path('/work/workplace/music69-windows')
repo = Path('/ref')


def sha(p):
    assert p.is_file() and not p.is_symlink(), p
    assert (p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid()), p
    with p.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def load(p):
    return json.loads(p.read_text())


def packets(p):
    data = p.read_bytes()
    result, pending, offset = [], b'', 0
    while offset < len(data):
        assert data[offset:offset+4] == b'OggS'
        count = data[offset+26]
        sizes = data[offset+27:offset+27+count]
        assert len(sizes) == count
        pos = offset+27+count
        for size in sizes:
            assert pos+size <= len(data)
            pending += data[pos:pos+size]
            pos += size
            if size < 255:
                result.append(pending)
                pending = b''
        assert pos > offset
        offset = pos
    assert not pending and offset == len(data)
    assert [r[:7] for r in result[:3]] == [b'\x01vorbis', b'\x03vorbis', b'\x05vorbis']
    return result


lock = repo/'docs/internals/data/windows-music-source-lock.json'
sources = load(root/'source-cache/VERIFIED.json')
assert sources['status'] == 'complete-fixed-windows-probe-source-cache'
assert sources['lock_sha256'] == sha(lock) and sources['debian_sources'] == 14 and sources['extra_sources'] == 2
for name, record in sources['files'].items():
    p = root/'source-cache'/name
    assert sha(p) == record['sha256'] and p.stat().st_size == record['bytes'], name
bundle = load(root/'source-delivery/BUNDLE.json')
archive = repo/bundle['path']
assert sha(archive) == bundle['sha256'] and archive.stat().st_size == bundle['bytes']
assert bundle['lock_sha256'] == sha(lock)

package = root/'gm-source-utf8-probe/封裝 測試'
probe = load(package/'PROBE.json')
build = load(root/'gm-build-utf8/BUILD.json')
assert build['status'] == 'compiled-probe-only' and probe['gm_build_sha256'] == sha(root/'gm-build-utf8/BUILD.json')
patch = repo/'tools/scripts/music_windows_fluidsynth_utf8.patch'
assert build['cli_patch']['patch_sha256'] == sha(patch)
assert build['cli_patch']['patched_source_sha256'] == sha(root/'gm-build-utf8/src/fluidsynth/src/fluidsynth.c')
assert build['cli_patch']['patched_renderer_sha256'] == sha(root/'gm-build-utf8/src/fluidsynth/src/bindings/fluid_filerenderer.c')
assert build['builder_sha256'] == sha(repo/'tools/scripts/music_windows_gm_probe.py')
assert {p.name for p in (package/'tools').iterdir()} == {r['filename'] for r in probe['files'].values()}
for row in probe['files'].values():
    assert sha(package/'tools'/row['filename']) == row['sha256']
for name, expected in probe['licenses_sha256'].items():
    assert sha(package/'licenses'/name) == expected
assert sha(package/'music-import.exe') == probe['program_sha256']
assert sha(package/'soundfont/FluidR3_GM.sf2') == probe['soundfont_sha256'] == '74594e8f4250680adf590507a306655a299935343583256f3b722c48a1bc1cb0'

gm = load(root/'gm-utf8-full/生成 曲庫/GM-SOURCES.json')
sdk = load(root/'gm-full/GM-SOURCES.json')
assert gm['status'] == sdk['status'] == 'complete-gm-stage' and gm['platform'] == 'windows/amd64'
assert gm['program_sha256'] == probe['program_sha256'] and gm['soundfont_sha256'] == probe['soundfont_sha256']
assert len(gm['tracks']) == len(sdk['tracks']) == 11
baseline = {r['track']: r for r in sdk['tracks']}
assert {r['track'] for r in gm['tracks']} == set(baseline)
audio = {}
for row in gm['tracks']:
    name = row['track']
    reference = baseline[name]
    assert row['midi_sha256'] == reference['midi_sha256']
    assert sha(Path('/orig')/(name+'.MDS')) == row['source_sha256'] == reference['source_sha256']
    p = root/'gm-utf8-full/生成 曲庫'/(name+'.ogg')
    old = root/'gm-full'/(name+'.ogg')
    assert sha(p) == row['ogg_sha256'] and sha(old) == reference['ogg_sha256']
    current_packets, old_packets = packets(p), packets(old)
    equal = current_packets[3:] == old_packets[3:]
    assert equal, 'GM Vorbis 音訊封包與 SDK 不同：'+name
    assert row['seconds'] == reference['seconds'] and row['peak'] == reference['peak']
    audio[name] = {'vorbis_audio_packets_equal': equal, 'audio_packets': len(current_packets)-3,
        'audio_packet_sha256': hashlib.sha256(b''.join(struct.pack('<I', len(p))+p for p in current_packets[3:])).hexdigest()}
assert not list((root/'gm-utf8-full/生成 曲庫').glob('.gm-import-*'))

waveform = load(root/'gm-wave-utf8-comparison/RESULT.json')
assert set(waveform['tracks']) == {'BGM_0M', 'END_2M', 'OPEN_M'}
for name, row in waveform['tracks'].items():
    assert row['equal_pcm'] and row['different_samples'] == row['length_difference_samples'] == 0
    assert row['midi_sha256'] == baseline[name]['midi_sha256']
    for side, record in row['renders'].items():
        path = root/record['path']
        assert sha(path) == record['wav_sha256']
        with wave.open(str(path)) as w:
            assert w.getnchannels() == 2 and w.getsampwidth() == 2 and w.getframerate() == 44100
            assert w.getnframes() == record['frames']
            assert hashlib.sha256(w.readframes(w.getnframes())).hexdigest() == record['pcm_sha256']
    assert len({r['wav_sha256'] for r in row['renders'].values()}) == 1
assert waveform['tracks']['BGM_0M']['source_repeat_equal_pcm'] and waveform['tracks']['BGM_0M']['sdk_repeat_equal_pcm']

rejected = load(root/'gm-utf8-rejected/GM-SOURCES.json')
assert rejected['status'] == 'failed' and not rejected['tracks']
assert rejected['program_sha256'] == probe['program_sha256']
assert 'FluidSynth 明示合成或載入失敗' in rejected['error'] and 'not a SoundFont or MIDI file' in rejected['error']
assert not list((root/'gm-utf8-rejected').glob('*.ogg')) and not list((root/'gm-utf8-rejected').glob('.gm-import-*'))
assert '--- PASS: TestWindowsUnicodeGMImportPlayerDecode' in (root/'player-utf8-test.log').read_text()
assert '--- PASS: TestPlayerPathMusicSwitches' in (root/'player-utf8-test.log').read_text()
product_log = (root/'product-utf8-tests.log').read_text()
assert 'FAIL' not in product_log and 'github.com/wicanr2/logh3/cmd/gin3' in product_log
assert 'Ran 5 tests' in (root/'source-boundary-tests.log').read_text() and 'OK' in (root/'source-boundary-tests.log').read_text()
game_bundle = repo/'org_game/银河英雄传说3SP.zip'
assert sha(game_bundle) == 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673'
scripts = ['music_windows_sources.py', 'music_windows_sources_test.py', 'music_windows_gm_probe.py',
    'music_windows_gm_package_probe.py', 'music_windows_gm_wave_probe.py',
    'music_windows_gm_encoding_probe.py', 'music_windows_path_probe.c', 'music_windows_gm_source_receipt.py',
    'music_windows_fluidsynth_utf8.patch']
logs = ['gm-build-utf8-2.log', 'gm-wave-utf8-comparison-2.log', 'gm-utf8-full.log',
        'gm-utf8-rejected.log', 'source-boundary-tests.log', 'source-bundle.log',
        'player-utf8-test.log', 'product-utf8-tests.log', 'gm-font-cjk.log',
        'gm-unicode-path-api-2.log', 'gm-build-probe-3.log', 'gm-build-utf8.log',
        'gm-wave-utf8-comparison.log', 'gm-encoding-probe.log']
first_patch = root/'gm-attempt-utf8-1/CLI-PATCH.diff'
first_build = load(root/'gm-attempt-utf8-1/BUILD.json')
assert sha(first_patch) == first_build['cli_patch']['patch_sha256']
result = {'schema': 'logh3.music-tools.windows.gm-sources.v1', 'date': '2026-10-01',
    'status': 'Windows GM 固定來源、Unicode 路徑與十一首原生技術驗收通過；正式交付未完成',
    'source_parent_head': '851a2f402a2d836b27e4df6e2c83f5294b5bdeb2', 'source_dirty': True,
    'original_bundle_sha256': sha(game_bundle), 'source_lock_sha256': sha(lock),
    'sources': sources, 'source_bundle': bundle, 'gm_build': build, 'package': probe,
    'gm': gm, 'waveform': waveform, 'ogg_sdk_comparison': audio, 'rejected': rejected,
    'buildinfo': (root/'importer-utf8-buildinfo.txt').read_text(),
    'scripts_sha256': {n: sha(repo/'tools/scripts'/n) for n in scripts},
    'product_sha256': {n: sha(repo/n) for n in ['cmd/music-import/main.go', 'cmd/music-import/main_test.go', 'cmd/gin3/music_integration_test.go']},
    'logs_sha256': {n: sha(root/n) for n in logs},
    'first_unicode_attempt': {'build': first_build, 'patch_sha256': sha(first_patch),
        'build_sha256': sha(root/'gm-attempt-utf8-1/BUILD.json'),
        'limit': '命令列轉 UTF-8 後仍有第二次 ANSI 轉碼；波形診斷失敗，未採用此產物'},
    'images': {'cross': 'sha256:0745d96e4029b152f468e2fee63c01e20bd859a369efe80de681e95cf17f0541',
        'wine': 'sha256:7d3c7ff554db39300e3e7a92a7b79c2d5fcd647289596d22fa10899baac61201',
        'python': 'sha256:229a2c5bfa27522db7815ea81f9bed70af17ccb9de9fc7ad142b1877b5830d36',
        'go': 'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac',
        'gui': 'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8'},
    'receipt_python': sys.version, 'formal_delivery_complete': False,
    'limits': '只驗 Windows Wine 原生 CLI 與 Linux 正式播放層；不外推實機、人耳、循環或 macOS。GM 保留既有近似，部分峰值 1.0 未聲稱無 clipping；未逐套重建執行期或驗來源簽章信任鏈。'}
target = Path('/work/docs/internals/data/windows-gm-source-tools.json')
assert target.parent.stat().st_uid == os.getuid()
if target.exists():
    sha(target)
target.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
print('Windows GM 固定來源技術收據 SHA-256', sha(target))
