"""反例探針：MinGW 改選 libvorbis Unix 數學路徑，仍未達 OGG 一致。

只供研究 362 的可丟棄副本，不授權將改選路徑放入正式發行工具。
"""
import os,subprocess,shutil,hashlib,json
from pathlib import Path
root=Path('/out'); src=root/'src/libvorbis'; dest=root/'diagnostic/original-dlls'
dest.mkdir(exist_ok=False)
for p in (root/'封裝 測試/tools').glob('libvorbis*.dll'): shutil.copyfile(p,dest/p.name)
p=src/'lib/os.h'; text=p.read_text(); old='''#  include <malloc.h>
#  define rint(x)   (floor((x)+0.5f))
#  define NO_FLOAT_MATH_LIB
#  define FAST_HYPOT(a, b) sqrt((a)*(a) + (b)*(b))'''
new='''#  include <malloc.h>
#  if !defined(__MINGW32__)
#    define rint(x)   (floor((x)+0.5f))
#    define NO_FLOAT_MATH_LIB
#    define FAST_HYPOT(a, b) sqrt((a)*(a) + (b)*(b))
#  endif'''
assert text.count(old)==1
p.write_text(text.replace(old,new))
env=dict(os.environ,PKG_CONFIG_LIBDIR='/out/prefix/lib/pkgconfig',PKG_CONFIG_PATH='',CFLAGS='-O2',CXXFLAGS='-O2',CPPFLAGS='-I/out/prefix/include',LDFLAGS='-L/out/prefix/lib')
commands=[['make','clean'],['make','-j2'],['make','install']]
for i,cmd in enumerate(commands):
 with (root/f'vorbis-math-{i}.log').open('w') as log:
  p=subprocess.run(cmd,cwd=src,env=env,stdout=log,stderr=subprocess.STDOUT)
 print(cmd,p.returncode,flush=True)
 if p.returncode: raise RuntimeError((root/f'vorbis-math-{i}.log').read_text()[-4000:])
for p in (root/'prefix/bin').glob('libvorbis*.dll'): shutil.copyfile(p,root/'封裝 測試/tools'/p.name)
(root/'vorbis-math-build.json').write_text(json.dumps({'scope':'probe-only-original-Unix-math-branch','commands':commands,'source_os_h_sha256':hashlib.sha256((src/'lib/os.h').read_bytes()).hexdigest(),'libraries':{p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in (root/'prefix/bin').glob('libvorbis*.dll')}},indent=2)+'\n')
