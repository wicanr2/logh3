"""#69 完整原生匯入收據；依 docs/playtest/031-native-fm-import.md 在 Docker 執行。

LOGH3_EXTRA_RO 須指向本專案。只讀實際完成的完整批次與三條觸發路徑重跑，
不將進行中批次、舊 OGG 或交叉建置當成完整平台驗收。
"""
import hashlib
import json
import os
from pathlib import Path
import struct

root = Path('/ref')
work = Path('/work/workplace/music69-fm')
docs = Path('/work/docs')


def sha(path):
    with path.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def load(path):
    return json.loads(path.read_text())


meta = load(work / 'full-runmeta.json')
bundle_hash = sha(root / 'org_game/银河英雄传说3SP.zip')
assert bundle_hash == 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673'
fm = load(work / 'full/FM-SOURCES.json')
gm = load(work / 'full/GM-SOURCES.json')
repeat = load(work / 'repeat/FM-SOURCES.json')
lock_path = docs / 'internals/data/fm-import-lock.json'
lock = load(lock_path)
assert fm['status'] == repeat['status'] == 'complete-fm-stage'
assert gm['status'] == 'complete-gm-stage'
assert len(fm['tracks']) == len(gm['tracks']) == 11
assert len(repeat['tracks']) == 3
assert fm['probe_head'] == repeat['probe_head'] == meta['probe_head']
assert fm['probe_head'] == lock['probe_head']
assert fm['source_executable_sha256'] == lock['executables']
assert fm['program_sha256'] == gm['program_sha256'] == repeat['program_sha256'] == meta['program_sha256']
assert meta['program_sha256'] == sha(work / 'music-import')
for name, expected in meta['source_sha256'].items():
    assert sha(root / name) == expected, name
for name, expected in fm['source_executable_sha256'].items():
    assert sha(Path('/orig') / name) == expected, name
for track in fm['tracks'] + gm['tracks']:
    n = track['track']
    assert sha(Path('/orig') / (n+'.MDS')) == track['source_sha256'], n
    assert sha(work / 'full' / (n+'.ogg')) == track['ogg_sha256'], n
    assert track['seconds'] > 0 and track['peak'] > 0
for track in fm['tracks']:
    n = track['track']
    for field in ('source_sha256', 'vgm_sha256', 'ogg_sha256'):
        assert track[field] == lock['tracks'][n][field], (n, field)
    assert track['ogg_serial'] == lock['tracks'][n]['serial']
    assert sha(work / 'full' / (n+'.vgm')) == track['vgm_sha256'], n
    assert sha(root / 'assets/original/music/vgm' / (n+'.vgm')) == track['vgm_sha256'], n
    assert sha(root / 'assets/original/music/ogg' / (n+'.ogg')) == track['ogg_sha256'], n
for track in gm['tracks']:
    assert sha(root / 'assets/original/music/mid' / (track['track']+'.mid')) == track['midi_sha256']
reproduction = {}
full_fm = {t['track']: t for t in fm['tracks']}
assert {t['track'] for t in repeat['tracks']} == {'BGM_0F', 'END_2F', 'OPEN_F'}
for track in repeat['tracks']:
    n = track['track']
    assert track['vgm_sha256'] == full_fm[n]['vgm_sha256']
    assert track['ogg_sha256'] == full_fm[n]['ogg_sha256']
    for suffix in ('.vgm', '.ogg'):
        assert (work / 'full' / (n+suffix)).read_bytes() == (work / 'repeat' / (n+suffix)).read_bytes()
    reproduction[n] = {'route': track['route'], 'vgm_sha256': track['vgm_sha256'], 'ogg_sha256': track['ogg_sha256'], 'result': '獨立原版固定序列重跑，完整 VGM／OGG 逐位元組相同'}
assert fm['tools_sha256'] == repeat['tools_sha256']
tool_hashes = {line.split()[1]: line.split()[0] for line in (work / 'tool-hashes.txt').read_text().splitlines()}
assert all(tool_hashes[n] == h for n, h in fm['tools_sha256'].items())
assert all(tool_hashes[n] == h for n, h in gm['tools_sha256'].items())
assert tool_hashes['/usr/share/sounds/sf2/FluidR3_GM.sf2'] == gm['soundfont_sha256']
for name, expected in fm['tools_sha256'].items():
    actual = work / 'tools' / Path(name).name
    # 研究批次工具有指向 image 的連結；由音訊 image 內另行核對。
    if not actual.is_symlink():
        assert sha(actual) == expected
assert not list((work / 'full').glob('.fm-import-*'))
assert not list((work / 'repeat').glob('.fm-import-*'))
programs = {}
for name, platform in [('music-import', 'linux/amd64'), ('music-import-windows-amd64.exe', 'windows/amd64'), ('music-import-darwin-amd64', 'darwin/amd64'), ('music-import-darwin-arm64', 'darwin/arm64')]:
    p = work / name
    raw = p.read_bytes()
    if platform.startswith('linux'):
        assert raw[:4] == b'\x7fELF' and raw[4] == 2 and struct.unpack_from('<H', raw, 18)[0] == 62
    elif platform.startswith('windows'):
        pe = struct.unpack_from('<I', raw, 0x3c)[0]
        assert raw[:2] == b'MZ' and raw[pe:pe+4] == b'PE\0\0' and struct.unpack_from('<H', raw, pe+4)[0] == 0x8664
    else:
        assert raw[:4] == b'\xcf\xfa\xed\xfe'
        assert struct.unpack_from('<I', raw, 4)[0] == (0x1000007 if platform.endswith('amd64') else 0x100000c)
    assert (p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid())
    programs[platform] = {'sha256': sha(p), 'bytes': len(raw), 'scope': 'Linux 真實執行；其他平台僅交叉建置與檔頭驗證'}
test_log = work / 'player-test.log'
assert '--- PASS: TestNativeFullMusicImportPlayerDecode' in test_log.read_text()
assert '--- PASS: TestPlayerPathMusicSwitches' in test_log.read_text()
rejections = {}
for name in ('exclusive', 'track', 'tools', 'original-output', 'probe', 'source'):
    p = work / 'check' / (name+'.log')
    assert '音樂匯入失敗' in p.read_text()
    assert not (work / ('rejected-'+name)).exists()
    rejections[name] = {'log_sha256': sha(p), 'output_created': False}
receipt = {
    'date': '2026-10-01', 'scope': 'Linux 原生完整曲庫匯入／正式解碼；平台依賴封裝與聆聽未完成',
    'original_bundle_sha256': bundle_hash,
    'run': meta, 'fm': fm, 'gm': gm, 'repeat': repeat, 'preflight_rejections': rejections,
    'audio_environment': {'tool_sha256': tool_hashes, 'package_versions': (work / 'tool-versions.txt').read_text(), 'hashes_log_sha256': sha(work / 'tool-hashes.txt'), 'versions_log_sha256': sha(work / 'tool-versions.txt')},
    'validation': {'fm_lock_sha256': sha(lock_path), 'full_fm_receipt_sha256': sha(work / 'full/FM-SOURCES.json'), 'full_gm_receipt_sha256': sha(work / 'full/GM-SOURCES.json'), 'repeat_receipt_sha256': sha(work / 'repeat/FM-SOURCES.json'), 'reproduction': reproduction, 'programs': programs, 'player_test_log_sha256': sha(test_log), 'source_inputs_unchanged': True, 'temporary_game_directories_removed': True, 'limits': '原版 OPL 暫存器串已重現；時間與合成為 hardware-spec approximation。GM 沿用 069 近似，部分峰值 1.0，未宣稱無 clipping；Windows／macOS 依賴封裝、實機與人耳／loop 驗收仍待完成。'},
}
target = docs / 'internals/data/native-fm-import.json'
for p in (target.parent, *([target] if target.exists() else [])):
    assert (p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid())
target.write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
print('完整原生曲庫收據 SHA-256：', sha(target))
