"""規格 076：重生 Windows OPL Unicode 輸入／輸出的完整波形探針。

既有 music_cross.sh 執行 --build，music_wine.sh 執行 --wave。
--inputs 指向三首由固定 dosgolem 取得的 VGM，必須符合 FM 技術收據。
預設輸出 opl-unicode-probe；重跑可指定另一個 /out 下的 --output-dir，拒絕覆蓋。
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shlex
import shutil
import subprocess
import wave


def sha(p):
    assert p.is_file() and not p.is_symlink() and (p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid()), p
    with p.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def build(root, inputs):
    assert root.is_relative_to('/out') and not root.exists() and not root.is_symlink()
    fm_path = Path('/metadata/windows-fm-pcm.json')
    assert sha(fm_path) == '4fc93a59c4a9b1f65b9e1a0114bc919126471b0d13fcef9102fdc29df93bac5a'
    fm = json.loads(fm_path.read_text())
    baseline = {r['track']: r for r in fm['fm']['tracks']}
    originals = {n: inputs/(n+'.vgm') for n in ['BGM_0F', 'END_2F', 'OPEN_F']}
    for name, p in originals.items():
        assert sha(p) == baseline[name]['vgm_sha256'], name
    old_tools = Path('/out/pcm-gate/封裝 測試/tools')
    for name, row in fm['package_files'].items():
        if name != 'music-import.exe':
            assert sha(old_tools/name) == row['sha256'], name
    root.mkdir()
    (root/'inputs').mkdir()
    for name, p in originals.items():
        shutil.copyfile(p, root/'inputs'/p.name)
    shutil.copytree(old_tools, root/'tools')
    source = Path('/tools/opl2wav.cpp')
    shutil.copyfile(source, root/'opl2wav.cpp')
    env = dict(os.environ, PKG_CONFIG_LIBDIR='/out/prefix/lib/pkgconfig', PKG_CONFIG_PATH='')
    flags = shlex.split(subprocess.check_output(['pkg-config', '--cflags', '--libs', 'adplug'], env=env, text=True))
    args = ['x86_64-w64-mingw32-g++', '-O2', '-std=c++11', str(root/'opl2wav.cpp')]+flags+['-lshell32', '-o', str(root/'tools/opl2wav.exe')]
    subprocess.run(args, env=env, check=True)
    record = {'scope': '相同固定 VGM／OPL DSP，Windows Unicode 檔案探針', 'command': args,
              'cpp_sha256': sha(source), 'program_sha256': sha(root/'tools/opl2wav.exe'),
              'inputs': {n: sha(p) for n, p in originals.items()}, 'builder_sha256': sha(Path(__file__))}
    (root/'BUILD.json').write_text(json.dumps(record, ensure_ascii=False, indent=2)+'\n')
    print('Unicode OPL 探針已建置', record['program_sha256'], flush=True)


def render(root):
    def win(p):
        return 'Z:'+str(p).replace('/', '\\')
    out = root/'wave'
    assert not out.exists() and not out.is_symlink()
    out.mkdir()
    build = json.loads((root/'BUILD.json').read_text())
    new, old = root/'tools/opl2wav.exe', Path('/out/pcm-gate/封裝 測試/tools/opl2wav.exe')
    assert sha(new) == build['program_sha256']
    records, commands = {}, []
    for name in ['BGM_0F', 'END_2F', 'OPEN_F']:
        source = root/'inputs'/(name+'.vgm')
        assert sha(source) == build['inputs'][name]
        chinese = root/'輸入 曲目'/source.name
        chinese.parent.mkdir(exist_ok=True)
        shutil.copyfile(source, chinese)
        output = root/'輸出 波形'
        output.mkdir(exist_ok=True)
        variants = [('legacy', old, source, out/(name+'-legacy.wav')),
                    ('utf8', new, chinese, output/(name+'-utf8.wav'))]
        if name == 'BGM_0F':
            variants.append(('utf8-repeat', new, chinese, output/(name+'-utf8-repeat.wav')))
        rows, samples = {}, {}
        for side, program, vgm, dest in variants:
            args = ['wine', win(program), win(vgm), win(dest), '44100']
            commands.append(args)
            p = subprocess.run(args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=900)
            (out/(name+'-'+side+'.log')).write_bytes(p.stdout)
            assert p.returncode == 0, p.stdout.decode(errors='replace')
            with wave.open(str(dest)) as w:
                assert (w.getnchannels(), w.getsampwidth(), w.getframerate()) == (2, 2, 44100)
                pcm = w.readframes(w.getnframes())
                samples[side] = pcm
                rows[side] = {'wav_sha256': sha(dest), 'pcm_sha256': hashlib.sha256(pcm).hexdigest(),
                              'frames': w.getnframes(), 'program_sha256': sha(program), 'path': str(dest.relative_to(root))}
            print(name, side, rows[side]['wav_sha256'], flush=True)
        assert rows['legacy']['wav_sha256'] == rows['utf8']['wav_sha256'] and samples['legacy'] == samples['utf8']
        if name == 'BGM_0F':
            assert rows['utf8']['wav_sha256'] == rows['utf8-repeat']['wav_sha256']
        records[name] = {'vgm_sha256': sha(source), 'renders': rows, 'wav_pcm_equal': True}
    result = {'scope': '三路固定 VGM 相同 DSP，Unicode 檔名／寬字元輸入輸出探針',
              'tracks': records, 'commands': commands, 'builder_sha256': sha(Path(__file__))}
    (root/'WAVE.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print('Unicode 三路完整 WAV／PCM 與首路重跑相同。', flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument('--build', action='store_true')
    mode.add_argument('--wave', action='store_true')
    parser.add_argument('--inputs', type=Path)
    parser.add_argument('--output-dir', type=Path, default=Path('/out/opl-unicode-probe'))
    args = parser.parse_args()
    assert args.output_dir.is_absolute() and args.output_dir.is_relative_to('/out') and '..' not in args.output_dir.parts
    if args.build:
        assert args.inputs is not None and args.inputs.is_absolute()
        build(args.output_dir, args.inputs)
    else:
        render(args.output_dir)
