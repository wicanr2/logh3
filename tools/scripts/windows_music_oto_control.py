"""第105輪：Windows鎖版Oto對照，隔離GUI及OGG解碼負載。"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

assert Path.cwd() == Path('/out') and os.getuid() == 1000
batch = sys.argv[1]
assert Path(batch).name == batch
out = Path(batch)
assert (out / 'READY.json').is_file()
binary = Path('windows-music-oto-probe-r105.exe')
source = Path('playeraudio69-r105-decode-windows/OPEN_M.f32le')
command = ['wine', str(binary.resolve()),
           'Z:\\out\\' + str(source).replace('/', '\\'),
           'Z:\\out\\' + batch + '\\OTO.json']
before = {name: Path('/sys/fs/cgroup/' + name).read_text()
          for name in ['cpu.max', 'cpu.stat']}
with (out / 'oto.log').open('x') as log:
    result = subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, timeout=60)
after = {name: Path('/sys/fs/cgroup/' + name).read_text()
         for name in ['cpu.max', 'cpu.stat']}
with (out / 'CONTROL.json').open('x') as stream:
    json.dump({'command': command, 'exit': result.returncode,
               'binary_sha256': hashlib.sha256(binary.read_bytes()).hexdigest(),
               'source_sha256': hashlib.sha256(source.read_bytes()).hexdigest(),
               'cpu_before': before, 'cpu_after': after,
               'gui_used': False, 'production_player_verified': False}, stream, indent=2)
with (out / 'CLIENT-DONE.json').open('x') as stream:
    json.dump({'exit': result.returncode}, stream)
assert result.returncode == 0, (out / 'oto.log').read_text()
print((out / 'oto.log').read_text())
