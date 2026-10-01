"""研究 363：保留 WAV，比較新舊 Windows 編碼器與格式轉換路徑。"""
from pathlib import Path
import subprocess
import json

root = Path('/out')
out = root/'gm-encoding-diagnostic'
assert not out.exists()
out.mkdir()


def win(p):
    return 'Z:'+str(p).replace('/', '\\')


wav = root/'gm-wave-comparison/BGM_0M-source.wav'
new = root/'gm-source-probe/封裝 測試/tools/ffmpeg.exe'
old = root/'gm-probe/tools/ffmpeg.exe'
variants = {'normal': [], 'no-cpu': ['-cpuflags', '0'], 'anull': ['-af', 'anull'],
            'volume-zero': ['-af', 'volume=0dB'], 'old-ffmpeg': []}
commands = []
for name, options in variants.items():
    program = old if name == 'old-ffmpeg' else new
    args = ['wine', win(program), '-v', 'error', '-y']
    if name == 'no-cpu':
        args += options
    args += ['-i', win(wav)]
    if name != 'no-cpu':
        args += options
    args += ['-ar', '44100', '-c:a', 'libvorbis', '-q:a', '5', win(out/(name+'.ogg'))]
    commands.append(args)
    result = subprocess.run(args, capture_output=True, timeout=900)
    (out/(name+'.log')).write_bytes(result.stdout+result.stderr)
    assert result.returncode == 0, result.stderr.decode(errors='replace')[-2000:]
    print(name, '完成', flush=True)
(out/'COMMANDS.json').write_text(json.dumps(commands, ensure_ascii=False, indent=2)+'\n')
