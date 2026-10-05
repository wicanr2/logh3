"""第105輪：直接送入固定PCM，驗Pulse monitor本身，不驗Windows播放器。"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

assert Path.cwd() == Path('/src') and os.getuid() == 1000
batch = sys.argv[1]
assert Path(batch).name == batch
root = Path('workplace/music69-windows')
out = root / batch
ready = json.loads((out / 'READY.json').read_text())
source = root / 'playeraudio69-r105-decode/OPEN_M.f32le'
raw = source.read_bytes()[:8*44100*8]
target = out / 'control-source.f32le'
with target.open('xb') as stream:
    stream.write(raw)
command = ['paplay', '--raw', '--format=float32le', '--rate=44100',
           '--channels=2', str(target)]
env = dict(os.environ, PULSE_SERVER='unix:' + str((out / 'pulse/native').resolve()))
result = subprocess.run(command, env=env, timeout=30)
with (out / 'CONTROL.json').open('x') as stream:
    json.dump({'command': command, 'exit': result.returncode,
               'source_sha256': hashlib.sha256(raw).hexdigest(),
               'gui_used': False, 'production_player_verified': False}, stream, indent=2)
with (out / 'CLIENT-DONE.json').open('x') as stream:
    json.dump({'exit': result.returncode}, stream)
assert result.returncode == 0
