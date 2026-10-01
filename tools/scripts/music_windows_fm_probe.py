"""Windows FM 固定來源建置；實跑閘門依規格 074，單純編譯不算通過。

由 bash tools/music_cross.sh python3 /tools/scripts/music_windows_fm_probe.py 執行。
證據及限制：docs/re/362-windows-music-tools.md。拒絕覆蓋既有 src／prefix。
"""
import os, json, hashlib, subprocess, shutil, shlex
from pathlib import Path

root = Path('/out')
cache = Path('/source-cache')
lock = json.loads(Path('/metadata/linux-music-source-lock.json').read_text())
prefix = root/'prefix'
src = root/'src'
src.mkdir(exist_ok=True)
prefix.mkdir(exist_ok=True)
env = dict(os.environ, PKG_CONFIG_LIBDIR=str(prefix/'lib/pkgconfig'), PKG_CONFIG_PATH='',
           CFLAGS='-O2', CXXFLAGS='-O2', CPPFLAGS=f'-I{prefix}/include', LDFLAGS=f'-L{prefix}/lib')
commands = []
def run(args, cwd):
    commands.append({'args':args, 'cwd':str(cwd)})
    name = f'{len(commands):02d}-{Path(args[0]).name}.log'
    with (root/name).open('w') as out:
        p = subprocess.run(args, cwd=cwd, env=env, stdout=out, stderr=subprocess.STDOUT)
    print(name, p.returncode, flush=True)
    if p.returncode:
        print((root/name).read_text()[-7000:], flush=True)
        raise RuntimeError(name)

for package in ['libogg','libvorbis','libbinio','adplug','ffmpeg']:
    row = next(r for r in lock['sources'] if r['source']==package)
    for n,r in row['files'].items():
        p=cache/n
        assert p.stat().st_size==r['bytes'] and hashlib.file_digest(p.open('rb'),'sha256').hexdigest()==r['sha256']
    dsc=cache/row['description']['filename']
    assert hashlib.sha256(dsc.read_bytes()).hexdigest()==row['description']['sha256']
    dest=src/package
    assert not dest.exists(), '探針不覆蓋已存在來源'
    run(['dpkg-source','--no-check','-x',str(dsc),str(dest)], root)
    if package in ['libbinio','adplug']:
        run(['autoreconf','-fi'],dest)
    if package!='ffmpeg':
        args=['./configure','--host=x86_64-w64-mingw32',f'--prefix={prefix}','--enable-shared','--disable-static','--disable-dependency-tracking']
        run(args,dest)
        overrides=[]
        if package=='libbinio': overrides=['libbinio_la_LDFLAGS=-no-undefined -version-info 1:0:0']
        if package=='adplug': overrides=['src_libadplug_la_LDFLAGS=-no-undefined -release 2.3.3 -version-info 0']
        run(['make','-j2']+overrides,dest)
        run(['make','install']+overrides,dest)
    else:
        args=['./configure','--enable-cross-compile','--target-os=mingw32','--arch=x86_64','--cross-prefix=x86_64-w64-mingw32-',
              '--pkg-config=pkg-config',f'--prefix={prefix}','--disable-autodetect','--disable-everything','--disable-network',
              '--disable-doc','--disable-debug','--disable-x86asm','--disable-ffplay','--disable-ffprobe','--enable-ffmpeg',
              '--enable-protocol=file','--enable-demuxer=wav,ogg','--enable-decoder=pcm_s16le,pcm_f32le,vorbis',
              '--enable-encoder=libvorbis,pcm_f32le','--enable-muxer=ogg,pcm_f32le','--enable-filter=volume,aresample,anull','--enable-libvorbis',
              '--enable-static','--disable-shared']
        run(args,dest)
        run(['make','-j2'],dest)
        run(['make','install'],dest)
flags=shlex.split(subprocess.check_output(['pkg-config','--cflags','--libs','adplug'], env=env, text=True))
run(['x86_64-w64-mingw32-g++','-O2','-std=c++11','/tools/opl2wav.cpp']+flags+['-o',str(prefix/'bin/opl2wav.exe')],root)
for dll in ['libstdc++-6.dll','libgcc_s_seh-1.dll','libwinpthread-1.dll']:
    p=Path(subprocess.check_output(['x86_64-w64-mingw32-g++',f'-print-file-name={dll}'],text=True).strip())
    if p.is_file(): shutil.copyfile(p,prefix/'bin'/dll)
(root/'fm-build.json').write_text(json.dumps({'status':'compiled-probe-only','commands':commands,
    'sources':[{k:r[k] for k in ('source','version','description','files')} for r in lock['sources'] if r['source'] in ['libogg','libvorbis','libbinio','adplug','ffmpeg']],
    'files':{p.name:{'bytes':p.stat().st_size,'sha256':hashlib.file_digest(p.open('rb'),'sha256').hexdigest()} for p in (prefix/'bin').iterdir() if p.is_file()}},ensure_ascii=False,indent=2)+'\n')
