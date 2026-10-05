"""第105輪：未修改Windows玩家程式，真實輸入驗片頭／FM換曲／關閉。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import time

assert Path.cwd() == Path('/out') and os.getuid() == os.getgid() == 1000
parser = argparse.ArgumentParser()
parser.add_argument('batch')
args = parser.parse_args()
assert re.fullmatch(r'[A-Za-z0-9][A-Za-z0-9_-]*', args.batch)
out = Path('/out') / args.batch
assert (out / 'READY.json').is_file() and (out / 'pulse/native').is_socket()
client = out / 'client'
client.mkdir()


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def write(name, value):
    with (client / name).open('x') as stream:
        json.dump(value, stream, ensure_ascii=False, indent=2)
        stream.write('\n')


def win(path):
    return 'Z:' + str(path).replace('/', '\\')


binary = Path('/player/gin3.exe')
assert sha(binary) == '6af940c72f340a5ecc49b2c6dc092e0e6dca077c1136abada3cad09b49e8ffa9'
metadata = Path('/data/windows-combined-music-tools.json')
parent = json.loads(Path('/data/windows-player.json').read_text())
assert sha(metadata) == parent['sources']['docs/internals/data/windows-combined-music-tools.json']['sha256']
catalog = json.loads(metadata.read_text())
music = Path('/music/生成 曲庫')
inputs = {}
for group in ['FM', 'GM']:
    for row in catalog['full'][group]['tracks']:
        path = music / (row['track'] + '.ogg')
        assert not path.is_symlink() and sha(path) == row['ogg_sha256'], row['track']
        inputs[row['track']] = row['ogg_sha256']
assert len(inputs) == 22 and {p.stem for p in music.glob('*.ogg')} == set(inputs)
reference = Path('/out/player69-r104-reference2')
reference_records = json.loads((reference / 'REFERENCES.json').read_text())
for row in reference_records['runs']:
    assert sha(reference / (row['name'] + '.png')) == row['png_sha256']
with (client / 'driver.py').open('xb') as stream:
    stream.write(Path(__file__).read_bytes())
command = ['wine', str(binary), '-orig', win('/orig'), '-data', win('/data'),
           '-save', win(client / 'save'), '-music', win(music),
           '-shot', win(client / 'final.png'), '-shot-bin', win(client / 'final.bin'),
           '-state-out', win(client / 'final.txt'), '-shot-after', '3600']
write('START.json', {'command': command, 'binary_sha256': sha(binary),
                    'music_sha256': inputs, 'metadata_sha256': sha(metadata),
                    'pulse_server': os.environ['PULSE_SERVER'], 'default_intro': True,
                    'mesa_threads': os.environ.get('LP_NUM_THREADS'),
                    'script_used': False, 'state_injection': False, 'port_mapping_changed': False})
write('CPU-START.json', {name: Path('/sys/fs/cgroup/' + name).read_text()
                        for name in ['cpu.max', 'cpu.stat']})
actions, frames, markers = [], [], []
started = time.monotonic()
rc = 99
process = None


def mark(name):
    markers.append({'name': name, 'monotonic': time.monotonic(), 'elapsed': time.monotonic()-started})


try:
    with (client / 'player.log').open('x') as log:
        process = subprocess.Popen(command, stdout=log, stderr=subprocess.STDOUT)
        write('PROCESS.json', {'pid': process.pid, 'command': command})
        deadline, window = time.monotonic() + 30, None
        while window is None:
            assert process.poll() is None, (client / 'player.log').read_text()
            assert time.monotonic() < deadline, 'Windows視窗30秒期限'
            result = subprocess.run(['xdotool', 'search', '--onlyvisible', '--name',
                                     '^銀河英雄傳說III SP$'], capture_output=True, text=True)
            if result.returncode == 0:
                window = result.stdout.strip().splitlines()[-1]
            else:
                time.sleep(.1)

        def event(args):
            command = ['xdotool'] + args
            subprocess.run(command, check=True)
            actions.append({'monotonic': time.monotonic(), 'command': command})

        def click(button, x, y):
            event(['mousemove', '--window', window, str(x), str(y)]); time.sleep(.08)
            event(['mousedown', str(button)]); time.sleep(.08)
            event(['mouseup', str(button)]); time.sleep(.15)

        def capture(name):
            path = client / (name + '-window.png')
            subprocess.run(['import', '-window', window, str(path)], check=True)
            return path

        def frame(name):
            deadline, attempt = time.monotonic() + 3, 0
            while True:
                stem = name if attempt == 0 else name + '-wait-%02d' % attempt
                image = capture(stem)
                result = subprocess.run(['compare', '-metric', 'AE', str(reference / (name + '.png')),
                                         str(image), str(client / (stem + '-diff.png'))],
                                        capture_output=True, text=True)
                frames.append({'name': name, 'path': stem, 'different': result.stderr.strip(),
                               'exit': result.returncode, 'monotonic': time.monotonic()})
                if result.returncode == 0: break
                assert result.returncode == 1 and time.monotonic() < deadline, (name, result.stderr)
                time.sleep(.1); attempt += 1

        event(['windowfocus', window]); time.sleep(.5)
        mark('intro'); capture('intro'); time.sleep(8)
        click(1, 300, 240); frame('main'); mark('main'); time.sleep(5)
        click(1, 300, 210); frame('copy')
        for key in ['1', '3', '8', 'Return']:
            event(['keydown', '--window', window, key]); time.sleep(.08)
            event(['keyup', '--window', window, key]); time.sleep(.15)
        frame('month'); mark('fm0')
        click(3, 300, 240); frame('ship')
        click(3, 300, 240); frame('input'); time.sleep(6)
        click(1, 544, 125); frame('system')
        click(1, 560, 185); capture('settings')
        settings = json.loads(Path('/data/screen-settings.json').read_text())

        def choice(action):
            hot = next(h for h in settings['hotspots'] if h['action'] == action)
            assert not hot.get('disabled', False)
            click(1, hot['x']+hot['w']//2, hot['y']+hot['h']//2)

        choice('bgm-1'); capture('fm1'); mark('fm1'); time.sleep(6)
        choice('device-2'); capture('off'); mark('off'); time.sleep(5)
        rc = process.wait(timeout=110)
    assert rc == 0, (client / 'player.log').read_text()
    assert 'state=StateSettings\n' in (client / 'final.txt').read_text()
    assert not any(word in (client / 'player.log').read_text() for word in ['panic:', '配樂：', 'oto:'])
    write('CHECK.json', {'exit': rc, 'state': (client / 'final.txt').read_text(),
                         'actions': len(actions), 'frames': frames, 'markers': markers,
                         'expected_tracks': ['OPEN_M', '', 'BGM_0F', 'BGM_1F', ''],
                         'physical_device': False, 'human_listening': False})
    print('PASS Windows normal music player UI', flush=True)
finally:
    write('CPU-END.json', {name: Path('/sys/fs/cgroup/' + name).read_text()
                          for name in ['cpu.max', 'cpu.stat']})
    write('TRACE.json', {'actions': actions, 'frames': frames, 'markers': markers})
    if process is not None and process.poll() is None:
        process.kill(); process.wait(timeout=5)
    with (out / 'CLIENT-DONE.json').open('x') as stream:
        json.dump({'exit': rc}, stream)
