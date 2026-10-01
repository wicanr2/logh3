"""研究 362 的額外 f32 PCM 診斷輸出；不替換原 ffmpeg.exe。"""
import json, os, subprocess, shutil
from pathlib import Path
root=Path('/out')
record=json.loads((root/'fm-build.json').read_text())
entry=next(x for x in record['commands'] if x['args'][0]=='./configure' and x['cwd'].endswith('/ffmpeg'))
env=dict(os.environ,PKG_CONFIG_LIBDIR='/out/prefix/lib/pkgconfig',PKG_CONFIG_PATH='',CFLAGS='-O2',CXXFLAGS='-O2',CPPFLAGS='-I/out/prefix/include',LDFLAGS='-L/out/prefix/lib')
args=entry['args']+['--enable-muxer=pcm_f32le','--enable-encoder=pcm_f32le']
for i,cmd in enumerate([args,['make','-j2']]):
 with (root/f'ffmpeg-f32-{i}.log').open('w') as log:
  p=subprocess.run(cmd,cwd=entry['cwd'],env=env,stdout=log,stderr=subprocess.STDOUT)
 print(cmd[0],p.returncode,flush=True)
 if p.returncode: raise RuntimeError((root/f'ffmpeg-f32-{i}.log').read_text()[-4000:])
shutil.copyfile(root/'src/ffmpeg/ffmpeg.exe',root/'封裝 測試/tools/ffmpeg-f32.exe')
