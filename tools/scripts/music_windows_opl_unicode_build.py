"""規格 076：以既有固定 AdPlug／MinGW 工具鏈建置 Windows Unicode 檔案 adapter。

bash tools/music_cross.sh python3 /tools/scripts/music_windows_opl_unicode_build.py。
只重建 opl2wav；DSP 與共享 library 沿用已驗版本，不覆寫舊 prefix 或失敗收據。
"""
import hashlib
import json
import os
from pathlib import Path
import shlex
import subprocess

root = Path('/out/opl-unicode-production')
assert not root.exists() and not root.is_symlink(), '拒絕覆蓋正式路徑建置'
source = Path('/tools/opl2wav.cpp')
probe = Path('/out/opl-unicode-probe')


def sha(p):
    assert p.is_file() and not p.is_symlink() and (p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid()), p
    with p.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


wave = json.loads((probe/'WAVE.json').read_text())
assert sha(source) == sha(probe/'opl2wav.cpp'), '正式來源必須與已驗探針相同'
for name, row in wave['tracks'].items():
    assert row['wav_pcm_equal']
    for side, record in row['renders'].items():
        assert sha(probe/record['path']) == record['wav_sha256']
root.mkdir()
env = dict(os.environ, PKG_CONFIG_LIBDIR='/out/prefix/lib/pkgconfig', PKG_CONFIG_PATH='')
flags = shlex.split(subprocess.check_output(['pkg-config', '--cflags', '--libs', 'adplug'], env=env, text=True))
args = ['x86_64-w64-mingw32-g++', '-O2', '-std=c++11', str(source)]+flags+['-lshell32', '-o', str(root/'opl2wav.exe')]
subprocess.run(args, env=env, check=True)
record = {'scope': 'Windows Unicode 參數／檔案入口；原 DSP 不改', 'command': args,
          'source_sha256': sha(source), 'program_sha256': sha(root/'opl2wav.exe'),
          'probe_wave_sha256': sha(probe/'WAVE.json'), 'builder_sha256': sha(Path(__file__))}
(root/'BUILD.json').write_text(json.dumps(record, ensure_ascii=False, indent=2)+'\n')
print('Windows Unicode OPL 正式來源建置', record['program_sha256'], flush=True)
