"""規格 072：完成 Linux 音樂工具暫存包；入口見 docs/playtest/032-linux-music-tools.md。

在 tools/py.sh 內執行，LOGH3_EXTRA_RO 指專案。不是正式版號／公開交付器。
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
from urllib.parse import quote


def sha(p):
    with p.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def owned(p):
    st = p.lstat()
    assert not p.is_symlink() and (st.st_uid, st.st_gid) == (os.getuid(), os.getgid()), p


ap = argparse.ArgumentParser()
ap.add_argument('--stage', type=Path, default=Path('/work/workplace/music69-linux/stage'))
ap.add_argument('--audio-image', required=True)
ap.add_argument('--git-head', required=True)
args = ap.parse_args()
root = Path('/ref')
work = Path('/work/workplace')
stage = args.stage
owned(stage)
assert not (stage / 'music-import').exists(), '拒絕覆蓋已完成的包'
tools = stage / 'tools'
records = {}
for paragraph in (stage / 'dpkg-status').read_text().split('\n\n'):
    fields = dict(line.split(': ', 1) for line in paragraph.splitlines() if ': ' in line and not line.startswith(' '))
    if fields.get('Status') != 'install ok installed':
        continue
    package = fields['Package']
    records[package] = fields
    records[package+':'+fields['Architecture']] = fields
packages = {}
for name in (stage / 'package-names.txt').read_text().splitlines():
    r = records[name]
    source = re.fullmatch(r'([\w+.-]+)(?: \(([^)]+)\))?', r.get('Source', r['Package']))
    assert source, r
    source_package, source_version = source[1], source[2] or r['Version']
    copyright = stage / 'licenses/debian' / (name+'.txt')
    assert copyright.is_file()
    packages[name] = {'version': r['Version'], 'source_package': source_package, 'source_version': source_version, 'source_index': f'https://sources.debian.org/src/{quote(source_package)}/{quote(source_version, safe="")}/', 'copyright_file': str(copyright.relative_to(stage)), 'copyright_sha256': sha(copyright)}
owners = {}
for line in (stage / 'owners.tsv').read_text().splitlines():
    src, owner = line.split('\t')
    owners[Path(src).name] = {'input': src, 'package': owner.split(': ')[0]}
libraries = (stage / 'libraries.txt').read_text().splitlines()
assert 0 < len(libraries) <= 512
assert len({Path(x).name for x in libraries}) == len(libraries)
assert {p.name for p in (tools / 'lib').iterdir()} == {Path(x).name for x in libraries}
for p in (tools / 'lib').iterdir():
    assert p.is_file() and not p.is_symlink()
    owned(p)
    owners[p.name]['sha256'] = sha(p)
    owners[p.name]['bytes'] = p.stat().st_size
    assert owners[p.name]['package'] in packages
manifest = {'platform': 'linux/amd64', 'libraries': sorted(owners), 'files': owners}
(tools / 'DEPENDENCIES.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2)+'\n')
launcher = work / 'music69-linux/music-tool'
probe = work / 'music69-fm/tools/dosgolem-probe'
importer = work / 'music69-fm/music-import'
for n in ('ffmpeg', 'fluidsynth', 'opl2wav'):
    assert launcher.is_file() and not launcher.is_symlink()
    shutil.copyfile(launcher, tools/n)
    (tools/n).chmod(0o755)
for src, dst in [(probe, tools/'dosgolem-probe'), (importer, stage/'music-import')]:
    assert src.is_file() and not src.is_symlink()
    owned(src)
    shutil.copyfile(src, dst)
    dst.chmod(0o755)
sf = json.loads((root / 'docs/internals/data/player-music-library.json').read_text())['soundfont']
font_source = work / 'music69-source/prepared'
font = stage / 'soundfont'
font.mkdir()
for member, expected in sf['archive']['member_sha256'].items():
    src = font_source / Path(member).name
    owned(src)
    assert sha(src) == expected
    shutil.copyfile(src, font/src.name)
assert sha(font / sf['file']) == sf['sha256']
for n in ('opl2wav.cpp', 'Dockerfile.audio'):
    shutil.copyfile(root / 'tools' / n, stage / 'licenses' / n)
shutil.copyfile(root / 'docs/internals/data/fluidr3-gm-copyright.txt', font/'debian-copyright.txt')
assert sha(font/'debian-copyright.txt') == sf['copyright_sha256']
receipt = {'schema': 'logh3.music-tools.linux.v1', 'status': '工具已封裝；匯入與對應來源碼交付尚待驗收', 'platform': 'linux/amd64', 'audio_image': args.audio_image, 'git_head': args.git_head, 'working_tree': '本輪 READY 072 啟動器／封裝器未提交', 'packages': packages, 'dependency_manifest_sha256': sha(tools/'DEPENDENCIES.json'), 'soundfont': sf, 'source_sha256': {n: sha(root/n) for n in ['cmd/music-tool/main_linux.go', 'tools/scripts/music_linux_collect.sh', 'tools/scripts/music_linux_finalize.py']}, 'limits': '本機暫存包，不是正式公開封包。Debian 套件來源身份與授權已保存；對應來源碼 archive 須在正式散布前納入交付。其他平台／聆聽未驗收。'}
receipt['files'] = {str(p.relative_to(stage)): {'sha256': sha(p), 'bytes': p.stat().st_size} for p in sorted(stage.rglob('*')) if p.is_file()}
assert not any(Path(n).suffix.upper() in ('.MDS', '.COM', '.OGG', '.MID', '.VGM') for n in receipt['files'])
for p in stage.rglob('*'):
    owned(p)
(stage / 'BUILD.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
print('Linux 暫存工具包完成：', len(packages), '套件、', len(owners), '共享檔案；SF2 已隨包')
