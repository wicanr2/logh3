"""研究 363：固定來源 Windows GM 試建，未審 READY 前不更換正式工具。

入口 bash tools/music_cross.sh python3 /tools/scripts/music_windows_gm_probe.py。
拒絕覆蓋既有 gm-build；同版本來源先驗 SHA-256，網路關閉，不改 SDK 或 FM prefix。
"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tarfile

assert sys.argv[1:] in [[], ['--utf8']], '未知探針選項'
utf8 = sys.argv[1:] == ['--utf8']
root = Path('/out/gm-build-utf8' if utf8 else '/out/gm-build')
assert not root.exists(), '來源試建必須從乾淨目錄開始'
root.mkdir()
prefix, src = root/'prefix', root/'src'
prefix.mkdir()
src.mkdir()
lock = json.loads(Path('/metadata/windows-music-source-lock.json').read_text())
cache = Path('/source-cache')
env = dict(os.environ, PKG_CONFIG_LIBDIR=str(prefix/'lib/pkgconfig'), PKG_CONFIG_PATH='',
           CFLAGS='-O2', CXXFLAGS='-O2', CPPFLAGS=f'-I{prefix}/include', LDFLAGS=f'-L{prefix}/lib')
commands = []


def sha(path):
    with path.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def run(args, cwd):
    commands.append({'args': args, 'cwd': str(cwd)})
    name = f'{len(commands):02d}-{Path(args[0]).name}.log'
    with (root/name).open('w') as log:
        result = subprocess.run(args, cwd=cwd, env=env, stdout=log, stderr=subprocess.STDOUT)
    print(name, result.returncode, flush=True)
    if result.returncode:
        print((root/name).read_text()[-7000:], flush=True)
        raise RuntimeError(name)


selected = ['libffi', 'pcre2', 'zlib', 'glib2.0', 'libsndfile', 'fluidsynth']
sources = []
for name in selected:
    if name == 'zlib':
        continue  # Debian DFSG 移除 win32；下面使用完整同版本上游來源。
    row = next(s for s in lock['sources'] if s['source'] == name)
    for n, record in {row['description']['filename']: row['description'], **row['files']}.items():
        p = cache/n
        assert sha(p) == record['sha256'] and p.stat().st_size == record['bytes'], n
    run(['dpkg-source', '--no-check', '-x', str(cache/row['description']['filename']), str(src/name)], root)
    sources.append(row)

def extract_extra(record, destination, top, limit):
    archive = Path('/out')/record['filename']
    assert not archive.is_symlink() and sha(archive) == record['sha256'] and archive.stat().st_size == record['bytes']
    destination.mkdir()
    with tarfile.open(archive) as tf:
        total = 0
        for member in tf.getmembers():
            parts = Path(member.name).parts
            assert parts and parts[0] == top
            assert '..' not in parts and not member.name.startswith('/') and (member.isdir() or member.isfile())
            if len(parts) == 1:
                continue
            p = destination.joinpath(*parts[1:])
            if member.isdir():
                p.mkdir(parents=True, exist_ok=True)
            else:
                total += member.size
                assert total < limit and not p.exists()
                p.parent.mkdir(parents=True, exist_ok=True)
                p.write_bytes(tf.extractfile(member).read())


proxy = next(s for s in lock['extra_sources'] if s['source'] == 'proxy-libintl')
extract_extra(proxy, src/'glib2.0/subprojects/proxy-libintl', 'proxy-libintl-'+proxy['commit'], 1_000_000)
zlib = next(s for s in lock['extra_sources'] if s['source'] == 'zlib-upstream')
extract_extra(zlib, src/'zlib', 'zlib-'+zlib['version'], 20_000_000)
patch_record = None
if utf8:
    original = src/'fluidsynth/src/fluidsynth.c'
    renderer = src/'fluidsynth/src/bindings/fluid_filerenderer.c'
    assert sha(original) == '13a26a2c7bd62a4f7d041c38471ad2dc40bd7dc7f48c918b64d9a1413e2ceb04'
    assert sha(renderer) == '3e1166dc2098c64f1c8485640ec5f29173d0c80abb14e12540cfd8409489f753'
    patch_file = Path('/tools/scripts/music_windows_fluidsynth_utf8.patch')
    run(['patch', '--dry-run', '-p1', '-i', str(patch_file)], src/'fluidsynth')
    run(['patch', '-p1', '-i', str(patch_file)], src/'fluidsynth')
    patch_record = {'purpose': 'Windows CLI Unicode 參數探針；不改 DSP',
                    'patch_sha256': sha(patch_file), 'patched_source_sha256': sha(original),
                    'patched_renderer_sha256': sha(renderer)}

cross = root/'windows.ini'
cross.write_text("""[binaries]
c = 'x86_64-w64-mingw32-gcc'
cpp = 'x86_64-w64-mingw32-g++'
ar = 'x86_64-w64-mingw32-ar'
strip = 'x86_64-w64-mingw32-strip'
windres = 'x86_64-w64-mingw32-windres'
pkgconfig = 'pkg-config'
[host_machine]
system = 'windows'
cpu_family = 'x86_64'
cpu = 'x86_64'
endian = 'little'
[properties]
needs_exe_wrapper = true
""")
cmake_base = ['-DCMAKE_SYSTEM_NAME=Windows', '-DCMAKE_SYSTEM_VERSION=10',
    '-DCMAKE_C_COMPILER=x86_64-w64-mingw32-gcc', '-DCMAKE_CXX_COMPILER=x86_64-w64-mingw32-g++',
    '-DCMAKE_RC_COMPILER=x86_64-w64-mingw32-windres', '-DCMAKE_BUILD_TYPE=Release',
    f'-DCMAKE_INSTALL_PREFIX={prefix}', '-DCMAKE_INSTALL_LIBDIR=lib', '-DBUILD_SHARED_LIBS=ON']

for name in selected:
    source = src/name
    if name in ['libffi', 'pcre2']:
        # libffi release 自帶 configure；README 只要求 Git 來源執行 autogen。
        # 不用主機 libtool 重生它的私有宏，保持 release 已產生的配方。
        extra = ['--disable-docs'] if name == 'libffi' else ['--disable-pcre2grep', '--disable-pcre2test', '--disable-pcre2-16', '--disable-pcre2-32']
        run(['./configure', '--host=x86_64-w64-mingw32', f'--prefix={prefix}', '--enable-shared', '--disable-static']+extra, source)
        run(['make', '-j2'], source)
        run(['make', 'install'], source)
    elif name == 'zlib':
        # 官方 MinGW Makefile 的 libz.dll.a 與 zlib.pc 名稱一致；
        # 1.2.13 CMake 在 Windows 裝 libzlib.dll.a 卻仍宣告 -lz。
        settings = ['-fwin32/Makefile.gcc', 'PREFIX=x86_64-w64-mingw32-', 'SHARED_MODE=1',
                    f'prefix={prefix}', f'BINARY_PATH={prefix}/bin',
                    f'INCLUDE_PATH={prefix}/include', f'LIBRARY_PATH={prefix}/lib']
        run(['make', '-j2']+settings, source)
        run(['make', 'install']+settings, source)
    elif name == 'glib2.0':
        build = root/'glib-build'
        run(['meson', 'setup', str(build), str(source), '--cross-file', str(cross), '--wrap-mode=nodownload',
             f'--prefix={prefix}', '--libdir=lib', '--buildtype=release', '-Ddefault_library=shared',
             '-Dtests=false', '-Dinstalled_tests=false', '-Dnls=disabled', '-Dlibmount=disabled',
             '-Dxattr=false', '-Dselinux=disabled', '-Dman=false', '-Dgtk_doc=false', '-Dlibelf=disabled', '-Dsysprof=disabled'], root)
        run(['ninja', '-C', str(build), '-j2'], root)
        run(['ninja', '-C', str(build), 'install'], root)
    else:
        build = root/(name+'-build')
        args = ['cmake', '-S', str(source), '-B', str(build)]+cmake_base
        if name == 'libsndfile':
            args += ['-DENABLE_EXTERNAL_LIBS=OFF', '-DENABLE_MPEG=OFF', '-DBUILD_TESTING=OFF',
                     '-DBUILD_PROGRAMS=OFF', '-DBUILD_EXAMPLES=OFF', '-DENABLE_COMPATIBLE_LIBSNDFILE_NAME=ON']
        if name == 'fluidsynth':
            args += ['-Denable-floats=ON', '-Denable-libinstpatch=OFF', '-Denable-network=OFF',
                     '-Denable-dbus=OFF', '-Denable-readline=OFF', '-Denable-ladspa=OFF', '-DBUILD_TESTING=OFF']
            args += ['-Denable-'+k+'=OFF' for k in ['dsound', 'wasapi', 'waveout', 'winmidi', 'jack', 'portaudio',
                    'oss', 'pulseaudio', 'pipewire', 'sdl2', 'midishare', 'opensles', 'oboe']]
        run(args, root)
        run(['cmake', '--build', str(build), '-j2'], root)
        run(['cmake', '--install', str(build)], root)

for dll in ['libgcc_s_seh-1.dll', 'libstdc++-6.dll', 'libwinpthread-1.dll', 'libgomp-1.dll']:
    p = Path(subprocess.check_output(['x86_64-w64-mingw32-g++', '-print-file-name='+dll], text=True).strip())
    assert p.is_file(), dll
    shutil.copyfile(p, prefix/'bin'/dll)
record = {'status': 'compiled-probe-only', 'commands': commands, 'sources': sources, 'extra_sources': [proxy, zlib],
          'cli_patch': patch_record, 'builder_sha256': sha(Path(__file__)),
          'cross_file': cross.read_text(), 'files': {p.name: {'bytes': p.stat().st_size, 'sha256': sha(p)}
            for p in sorted((prefix/'bin').iterdir()) if p.is_file()}}
(root/'BUILD.json').write_text(json.dumps(record, ensure_ascii=False, indent=2)+'\n')
print('Windows GM 來源試建完成；仍需實跑與來源／授權審查。')
