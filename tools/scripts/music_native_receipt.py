"""#69 原生 GM 驗收收據；由 tools/py.sh 在容器讀本輪實物。

需 LOGH3_EXTRA_RO 指向本專案，與 docs/playtest/030-native-gm-import.md 一起使用。
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import struct

ap = argparse.ArgumentParser()
for name in ('go-version', 'audio-image', 'go-image', 'gui-image', 'git-head', 'working-tree'):
    ap.add_argument('--'+name, required=True)
args = ap.parse_args()
root = Path('/ref')
out = Path('/work/workplace/music69-native')


def sha(path):
    with path.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


source = out / '曲庫 測試/GM-SOURCES.json'
receipt = json.loads(source.read_text())
assert receipt['status'] == 'complete-gm-stage' and len(receipt['tracks']) == 11
for t in receipt['tracks']:
    assert t['source_sha256'] == sha(Path('/orig') / Path(t['source']).name)
    assert t['midi_sha256'] == sha(root / 'assets/original/music/mid' / (t['track']+'.mid'))
    assert t['ogg_sha256'] == sha(out / '曲庫 測試' / (t['track']+'.ogg'))
    assert t['seconds'] > 0 and t['peak'] > 0
checks = {}
for name in ('BGM_0M', 'END_2M', 'OPEN_M'):
    line = (out / 'check' / (name+'-pcm.sha256')).read_text().split()
    assert len(line) == 2 and len(line[0]) == 64
    checks[name] = {'decoded_pcm_sha256': line[0], 'reference_ogg_sha256': sha(out / 'reference' / (name+'.ogg')), 'comparison': 'cmp：舊命令與原生解碼 PCM 完全相同'}
for name in ('no-gm', 'fm', 'original-output', 'missing-tools', 'wrong-font'):
    assert '音樂匯入失敗' in (out / 'check' / (name+'.log')).read_text()
assert not (out / 'rejected').exists()
programs = {}
for name, expected in [('music-import', 'linux/amd64'), ('music-import-windows-amd64.exe', 'windows/amd64'), ('music-import-darwin-amd64', 'darwin/amd64'), ('music-import-darwin-arm64', 'darwin/arm64')]:
    p = out / name
    d = p.read_bytes()
    if expected.startswith('linux'):
        assert d[:4] == b'\x7fELF' and d[4] == 2 and struct.unpack_from('<H', d, 18)[0] == 62
    elif expected.startswith('windows'):
        pe = struct.unpack_from('<I', d, 0x3c)[0]
        assert d[:2] == b'MZ' and d[pe:pe+4] == b'PE\x00\x00' and struct.unpack_from('<H', d, pe+4)[0] == 0x8664
    else:
        assert d[:4] == b'\xcf\xfa\xed\xfe'
        assert struct.unpack_from('<I', d, 4)[0] == (0x1000007 if expected.endswith('amd64') else 0x100000c)
    assert (p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid())
    programs[expected] = {'sha256': sha(p), 'bytes': p.stat().st_size, 'scope': 'Linux 真實執行；其餘只交叉建置與檔頭驗證，未宣稱實機封裝驗收'}
receipt['validation'] = {
    'date': '2026-10-01', 'gm_manifest_sha256': sha(source),
    'go': args.go_version,
    'git_head': args.git_head, 'working_tree': args.working_tree,
    'images': {'logh3-audio:2': args.audio_image, 'golang:1.24-bookworm': args.go_image, 'logh3-ebiten:2': args.gui_image},
    'midi': {p.name: {'source_sha256': sha(p), 'saved_midi_sha256': sha(root / 'assets/original/music/mid' / (p.stem+'.mid'))} for p in sorted(Path('/orig').glob('*.MDS'))},
    'pcm_comparisons': checks,
    'rejected': ['no-gm', 'fm', 'original-output', 'missing-tools', 'wrong-font', '同長度錯誤 SF2：TestMusicImportPreflight'],
    'programs': programs,
    'source_sha256': {name: sha(root / name) for name in ['internal/asset/mds.go', 'cmd/music-import/main.go', 'tools/scripts/music_native_check.sh', 'tools/scripts/mds_export.py']},
    'limits': 'GM 階段；FM、三平台依賴封裝、聆聽／loop 未完成；峰值 1.0 曲目沿用舊增益，不冒稱無 clipping。',
}
target = Path('/work/docs/internals/data/native-gm-import.json')
target.write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
print('原生 GM 收據：', sha(target))
