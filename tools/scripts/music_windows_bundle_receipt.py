"""規格 076：完整預設入口、中文搬移、工具與正式播放層的完成收據。

LOGH3_EXTRA_RO 指向專案，tools/py.sh /work/tools/scripts/music_windows_bundle_receipt.py。
任何階段未完成或身份不符，都不能寫完成收據。
"""
import hashlib
import json
import os
import re
from pathlib import Path
import struct
import wave

ROOT = Path('/work/workplace/music69-windows')
REPO = Path('/ref')


def sha(p):
    assert p.is_file() and not p.is_symlink() and (p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid()), p
    with p.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def load(p):
    return json.loads(p.read_text())


def audio_packets(p):
    raw = p.read_bytes()
    pos, pending, packets = 0, b'', []
    while pos < len(raw):
        assert raw[pos:pos+4] == b'OggS'
        count = raw[pos+26]
        sizes = raw[pos+27:pos+27+count]
        assert len(sizes) == count
        index = pos+27+count
        for size in sizes:
            assert index+size <= len(raw)
            pending += raw[index:index+size]
            index += size
            if size < 255:
                packets.append(pending)
                pending = b''
        assert index > pos
        pos = index
    assert not pending and pos == len(raw)
    assert [p[:7] for p in packets[:3]] == [b'\x01vorbis', b'\x03vorbis', b'\x05vorbis']
    return packets[3:]


package = ROOT/'combined-utf8-stage/封裝 測試'
bundle = load(package/'BUNDLE.json')
assert bundle['schema'] == 'logh3.music-tools.windows.bundle.v1'
assert bundle['builder_sha256'] == sha(REPO/'tools/scripts/music_windows_bundle.py')
assert sha(package/'music-import.exe') == bundle['program_sha256']
for name, row in bundle['files'].items():
    p = package/name
    assert sha(p) == row['sha256'] and p.stat().st_size == row['bytes']
annex = load(package/'ANNEX.json')
assert annex['stage_bundle_sha256'] == sha(package/'BUNDLE.json')
assert annex['builder_sha256'] == sha(REPO/'tools/scripts/music_windows_bundle_archive.py')
for name, row in annex['files'].items():
    assert sha(package/name) == row['sha256']
archive = load(ROOT/'COMBINED-UTF8-ARCHIVE.json')
assert archive['utf8_flag_all'] and archive['readme_bom_crlf']
assert sha(ROOT/archive['archive']) == archive['sha256'] and (ROOT/archive['archive']).stat().st_size == archive['bytes']
relocated = ROOT/archive['relocated_path']
for name, row in archive['files'].items():
    for parent in [package, relocated]:
        assert sha(parent/name) == row['sha256'] and (parent/name).stat().st_size == row['bytes']
assert {str(p.relative_to(package)) for p in package.rglob('*') if p.is_file()} == set(archive['files'])
baseline = load(REPO/'docs/internals/data/fm-import-lock.json')['tracks']
full = ROOT/'combined-utf8-full/生成 曲庫'
previous_gm = ROOT/'gm-utf8-full/生成 曲庫'
previous_fm = ROOT/'fm-full'
records, comparisons = {}, {}
for stage, reference in [('GM', previous_gm), ('FM', previous_fm)]:
    r = load(full/(stage+'-SOURCES.json'))
    old = {row['track']: row for row in load(reference/(stage+'-SOURCES.json'))['tracks']}
    assert r['status'] == 'complete-'+stage.lower()+'-stage' and r['platform'] == 'windows/amd64'
    assert r['program_sha256'] == bundle['program_sha256']
    assert len(r['tracks']) == len(old) == 11 and {row['track'] for row in r['tracks']} == set(old)
    records[stage] = r
    if stage == 'GM':
        assert r['soundfont_sha256'] == bundle['files']['soundfont/FluidR3_GM.sf2']['sha256']
    for row in r['tracks']:
        name = row['track']
        p = full/(name+'.ogg')
        assert sha(p) == row['ogg_sha256']
        assert sha(Path('/orig')/(name+'.MDS')) == row['source_sha256'] == old[name]['source_sha256']
        assert row['seconds'] == old[name]['seconds'] and row['peak'] == old[name]['peak']
        if stage == 'FM':
            for field in ['vgm_sha256', 'wav_sha256', 'pcm_sha256']:
                assert row[field] == baseline[name][field] == old[name][field]
            assert sha(p) == sha(reference/(name+'.ogg')), '同 Windows 編碼與 WAV 不同：'+name
            comparisons[name] = {'vgm_wav_pcm_match': True, 'same_windows_ogg_equal': True}
        else:
            assert row['midi_sha256'] == old[name]['midi_sha256']
            a, b = audio_packets(p), audio_packets(reference/(name+'.ogg'))
            assert a == b, name
            comparisons[name] = {'midi_equal': True, 'vorbis_audio_packets_equal': True,
                                 'audio_packets_sha256': hashlib.sha256(b''.join(struct.pack('<I', len(x))+x for x in a)).hexdigest()}
    assert not list(full.glob('.'+stage.lower()+'-import-*'))
# 所有實際工具必須來自同包；Go 收據只列直接執行的工具，DLL 身份由包內清單另驗。
for r in records.values():
    for path, digest in r['tools_sha256'].items():
        assert path.startswith('Z:\\out\\combined-utf8-stage\\封裝 測試\\tools\\')
        name = path.rsplit('\\', 1)[1]
        assert sha(package/'tools'/name) == digest
assert len(list(full.glob('*.ogg'))) == 22
actual_importer_info = (ROOT/'combined-final-buildinfo.txt').read_text()
actual_probe_info = (ROOT/'combined-probe-buildinfo.txt').read_text()
assert 'vcs.revision=9a2ce97df1e222a4928dd782ddf210033213a96e' in actual_importer_info and 'vcs.modified=false' in actual_importer_info
assert 'vcs.revision=f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c' in actual_probe_info and 'vcs.modified=false' in actual_probe_info
repeat_dir = ROOT/'combined-utf8-relocated-output/生成 曲庫'
for stage, expected in [('GM', {'BGM_0M'}), ('FM', {'BGM_0F', 'END_2F', 'OPEN_F'})]:
    r = load(repeat_dir/(stage+'-SOURCES.json'))
    assert r['status'] == 'complete-'+stage.lower()+'-stage' and r['program_sha256'] == bundle['program_sha256']
    assert {row['track'] for row in r['tracks']} == expected
    for row in r['tracks']:
        name = row['track']; p = repeat_dir/(name+'.ogg')
        assert sha(p) == row['ogg_sha256']
        if stage == 'FM':
            for field in ['vgm_sha256', 'wav_sha256', 'pcm_sha256']:
                assert row[field] == baseline[name][field]
            assert sha(p) == sha(full/(name+'.ogg'))
        else:
            assert audio_packets(p) == audio_packets(full/(name+'.ogg'))
    for path, digest in r['tools_sha256'].items():
        assert path.startswith('Z:\\out\\'+str(relocated.relative_to(ROOT)).replace('/', '\\')+'\\tools\\')
        assert digest == sha(relocated/'tools'/path.rsplit('\\', 1)[1])
assert not list(repeat_dir.glob('.*-import-*'))
waves = load(ROOT/'opl-unicode-probe/WAVE.json')
for name, row in waves['tracks'].items():
    assert row['wav_pcm_equal']
    for side, record in row['renders'].items():
        p = ROOT/'opl-unicode-probe'/record['path']
        assert sha(p) == record['wav_sha256'] == baseline[name]['wav_sha256']
        with wave.open(str(p)) as w:
            pcm = w.readframes(w.getnframes())
            assert hashlib.sha256(pcm).hexdigest() == record['pcm_sha256']
assert load(ROOT/'opl-linux-preprocess/LINUX-CPP.json')['non_windows_preprocessed_equal']
first = load(ROOT/'combined-full/生成 曲庫/FM-SOURCES.json')
assert first['status'] == 'failed' and 'opl2wav.exe' in first['error'] and 'AdPlug' in first['error']
assert load(ROOT/'combined-full/生成 曲庫/GM-SOURCES.json')['status'] == 'complete-gm-stage'
assert sha(ROOT/'combined-bundle-attempt-1.py') == load(ROOT/'combined-stage/封裝 測試/BUNDLE.json')['builder_sha256']
assert sha(ROOT/'combined-archive-attempt-1.py') == load(ROOT/'COMBINED-ARCHIVE.json')['builder_sha256']
logs = ['combined-full.log', 'combined-utf8-full.log', 'combined-utf8-relocated.log',
        'combined-utf8-boundary-tests.log', 'combined-player-test.log', 'combined-product-tests.log',
        'opl-unicode-wave.log', 'opl-unicode-replay-wave.log', 'opl-linux-regression.log', 'opl-linux-branch.log', 'opl-linux-branch-2.log']
assert 'Ran 6 tests' in (ROOT/'combined-utf8-boundary-tests.log').read_text() and 'OK' in (ROOT/'combined-utf8-boundary-tests.log').read_text()
for name in ['combined-player-test.log', 'combined-product-tests.log']:
    data = (ROOT/name).read_text()
    assert 'FAIL' not in data and re.search(r'^ok\s+github.com/wicanr2/logh3/', data, re.M)
assert '--- PASS: TestWindowsCombinedImportPlayerDecode' in (ROOT/'combined-player-test.log').read_text()
replay = load(ROOT/'opl-unicode-replay/WAVE.json')
assert set(replay['tracks']) == set(waves['tracks']) and all(row['wav_pcm_equal'] for row in replay['tracks'].values())
scripts = ['music_windows_bundle.py', 'music_windows_bundle_archive.py', 'music_windows_bundle_test.py',
           'music_windows_opl_unicode_build.py', 'music_windows_opl_unicode_probe.py', 'music_windows_bundle_receipt.py']
assert sha(REPO/'org_game/银河英雄传说3SP.zip') == 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673'
result = {'schema': 'logh3.music-tools.windows.combined.v1', 'date': '2026-10-02',
          'status': 'Windows 合併預設二十二首、Unicode OPL 與中文搬移技術驗收通過；正式發行未完成',
          'source_parent_head': '9a2ce97df1e222a4928dd782ddf210033213a96e', 'validation_tree_dirty': True,
          'original_bundle_sha256': sha(REPO/'org_game/银河英雄传说3SP.zip'),
          'actual_importer_buildinfo': actual_importer_info, 'actual_probe_buildinfo': actual_probe_info,
          'bundle': bundle, 'annex': annex, 'archive': archive, 'full': records, 'comparisons': comparisons,
          'relocated': {stage: load(repeat_dir/(stage+'-SOURCES.json')) for stage in ['GM', 'FM']},
          'opl_wave': waves, 'replay_wave': replay, 'linux_branch': load(ROOT/'opl-linux-preprocess/LINUX-CPP.json'),
          'first_failed_fm': first, 'first_bundle': load(ROOT/'combined-stage/封裝 測試/BUNDLE.json'),
          'first_builder_source': (ROOT/'combined-bundle-attempt-1.py').read_text(),
          'first_archive_builder_source': (ROOT/'combined-archive-attempt-1.py').read_text(),
          'scripts_sha256': {name: sha(REPO/'tools/scripts'/name) for name in scripts},
          'source_files_sha256': {name: sha(REPO/name) for name in ['tools/opl2wav.cpp', 'cmd/gin3/music_integration_test.go']},
          'interface_sources_sha256': {name: sha(ROOT/name) for name in ['src/adplug/src/adplug.h',
              'src/adplug/src/fprovide.h', 'src/adplug/src/fprovide.cpp', 'src/libbinio/src/binfile.h', 'src/libbinio/src/binfile.cpp']},
          'logs_sha256': {name: sha(ROOT/name) for name in logs},
          'images': {'cross': 'sha256:0745d96e4029b152f468e2fee63c01e20bd859a369efe80de681e95cf17f0541',
                     'wine': 'sha256:7d3c7ff554db39300e3e7a92a7b79c2d5fcd647289596d22fa10899baac61201',
                     'gui': 'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8'},
          'full_command': ['LOGH3_TIMEOUT=3h', 'bash', 'tools/music_wine.sh', 'wine',
                           'Z:\\out\\combined-utf8-stage\\封裝 測試\\music-import.exe', '-source', 'Z:\\orig',
                           '-output', 'Z:\\out\\combined-utf8-full\\生成 曲庫'],
          'formal_delivery_complete': False,
          'limits': '限 Wine 原生 CLI／ZIP／中文搬移與 Linux 正式解碼。非 Windows 預處理相同不等於 Linux 新建置實跑。Windows 實機、人耳、循環、macOS、正式版號／dist-all 及本作品授權與原始碼交付仍待完成；GM 峰值 1.0 與原版音源近似保留。'}
target = Path('/work/docs/internals/data/windows-combined-music-tools.json')
assert target.parent.stat().st_uid == os.getuid()
if target.exists():
    sha(target)
target.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
print('Windows 合併曲庫收據 SHA-256', sha(target))
