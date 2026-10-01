"""研究 363：依實際 PE 引用組成隔離 GM 探針，未接正式發行。

bash tools/music_cross.sh python3 /tools/scripts/music_windows_gm_package_probe.py。
"""
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys

ROOT = Path('/out')
assert sys.argv[1:] in [[], ['--utf8']]
utf8 = sys.argv[1:] == ['--utf8']
package = ROOT/('gm-source-utf8-probe/封裝 測試' if utf8 else 'gm-source-probe/封裝 測試')
assert not package.exists(), '拒絕覆蓋試驗包'
build = ROOT/('gm-build-utf8' if utf8 else 'gm-build')
assert json.loads((build/'BUILD.json').read_text())['status'] == 'compiled-probe-only'
assert build.stat().st_uid == os.getuid() and ROOT.stat().st_uid == os.getuid()
tools = package/'tools'
tools.mkdir(parents=True)


def sha(p):
    assert p.is_file() and not p.is_symlink() and p.stat().st_uid == os.getuid(), p
    with p.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


# 只允許已觀測的 Windows 系統 DLL，所有其他引用必須在本次候選工具集合內閉合。
system = {'kernel32.dll', 'msvcrt.dll', 'advapi32.dll', 'ole32.dll', 'shell32.dll',
          'user32.dll', 'ws2_32.dll', 'bcrypt.dll'}
available = {p.name.lower(): p for p in (build/'prefix/bin').iterdir() if p.suffix == '.dll'}
fm = ROOT/'pcm-gate/封裝 測試/tools'
for p in fm.iterdir():
    if p.suffix == '.dll' and p.name.lower() not in available:
        available[p.name.lower()] = p
pending = [build/'prefix/bin/fluidsynth.exe', fm/'ffmpeg.exe']
metadata = {}
while pending:
    p = pending.pop()
    if p.name.lower() in metadata:
        continue
    imports = re.findall(r'DLL Name: (\S+)', subprocess.check_output(
        ['x86_64-w64-mingw32-objdump', '-p', str(p)], text=True))
    shutil.copyfile(p, tools/p.name)
    assert sha(tools/p.name) == sha(p)
    metadata[p.name.lower()] = {'filename': p.name, 'source': str(p.relative_to(ROOT)),
        'sha256': sha(p), 'bytes': p.stat().st_size, 'imports': imports}
    for name in imports:
        name = name.lower()
        if name not in system:
            assert name in available, '缺少非系統 DLL：'+name
            pending.append(available[name])
importer = ROOT/'music-import-utf8-gate.exe' if utf8 else ROOT/'pcm-gate/封裝 測試/music-import.exe'
shutil.copyfile(importer, package/'music-import.exe')
font = ROOT/'gm-probe/soundfont/FluidR3_GM.sf2'
assert sha(font) == '74594e8f4250680adf590507a306655a299935343583256f3b722c48a1bc1cb0'
(package/'soundfont').mkdir()
shutil.copyfile(font, package/'soundfont/FluidR3_GM.sf2')

licenses = package/'licenses'
licenses.mkdir()
license_files = {}
source_dirs = list((build/'src').iterdir()) + [ROOT/'src'/n for n in ['ffmpeg', 'libogg', 'libvorbis']]
for source in source_dirs:
    if not source.is_dir():
        continue
    candidates = [source/'debian/copyright']
    candidates += [p for p in source.iterdir() if p.is_file() and
                   p.name.upper().startswith(('COPYING', 'LICENSE', 'COPYRIGHT', 'AUTHORS'))]
    if source.name == 'glib2.0':
        candidates.append(source/'subprojects/proxy-libintl/COPYING')
    for p in candidates:
        if p.is_file():
            target = licenses/source.name/p.relative_to(source)
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(p, target)
            license_files[str(target.relative_to(licenses))] = sha(target)
for name, expected in {'COPYING': '8ef830b65c97a976b86e34bb5fde08d99dfb1db13c4149b5b20bc837ac6c4568',
                       'README': '2f92d462684629f207e43815115b79d7f1ce130d30ca8792b7d82b398a8306f9'}.items():
    p = ROOT/'soundfont-licenses'/name
    assert sha(p) == expected
    target = licenses/'soundfont'/name
    target.parent.mkdir(exist_ok=True)
    shutil.copyfile(p, target)
    license_files[str(target.relative_to(licenses))] = sha(target)
for name in ['gcc-mingw-w64-x86-64-win32-runtime', 'mingw-w64-x86-64-dev']:
    p = Path('/usr/share/doc')/name/'copyright'
    target = licenses/(name+'-copyright')
    shutil.copyfile(p, target)
    license_files[target.name] = sha(target)
for p in Path('/usr/share/common-licenses').iterdir():
    if p.is_file():
        target = licenses/'common'/p.name
        target.parent.mkdir(exist_ok=True)
        shutil.copyfile(p, target)
        license_files[str(target.relative_to(licenses))] = sha(target)
record = {'status': 'isolated-gm-source-probe-only',
          'gm_build_sha256': sha(build/'BUILD.json'), 'files': metadata,
          'program_sha256': sha(package/'music-import.exe'), 'soundfont_sha256': sha(font),
          'licenses_sha256': license_files, 'formal_delivery_complete': False}
(package/'PROBE.json').write_text(json.dumps(record, ensure_ascii=False, indent=2)+'\n')
print('隔離 GM 探針引用閉合：', len(metadata), '個工具／DLL；尚未接正式發行。')
