"""#69：在既有音訊Docker內供Wine連接的有界PulseAudio服務及monitor錄音。"""
import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import time

assert Path.cwd() == Path('/src') and os.getuid() == os.getgid() == 1000
parser = argparse.ArgumentParser()
parser.add_argument('batch')
args = parser.parse_args()
assert re.fullmatch(r'[A-Za-z0-9][A-Za-z0-9_-]*', args.batch)
parent = Path('workplace/music69-windows')
assert parent.stat().st_uid == parent.stat().st_gid == 1000
out = parent / args.batch
assert not out.exists()
out.mkdir()
runtime = out.resolve() / 'pulse'
runtime.mkdir(mode=0o700)
env = dict(os.environ, HOME='/tmp', XDG_RUNTIME_DIR=str(runtime),
           PULSE_SERVER='unix:' + str(runtime / 'native'))


def write(name, value):
    with (out / name).open('x') as stream:
        json.dump(value, stream, ensure_ascii=False, indent=2)
        stream.write('\n')


def control(*argv):
    return subprocess.check_output(['pactl', *argv], env=env, text=True, timeout=5)


def terminate(process):
    if process is not None and process.poll() is None:
        process.terminate()
        try:
            process.wait(timeout=3)
        except subprocess.TimeoutExpired:
            process.kill(); process.wait(timeout=3)


pulse = capture = None
started = time.monotonic()
with (out / 'pulse.log').open('x') as plog, (out / 'capture.log').open('x') as clog, (out / 'monitor.f32le').open('xb') as raw:
    try:
        command = ['pulseaudio', '--daemonize=no', '--use-pid-file=no', '--exit-idle-time=-1',
                   '--disable-shm=yes', '--high-priority=no', '--realtime=no', '--log-level=warning', '-n',
                   '--load=module-native-protocol-unix socket=' + str(runtime / 'native') + ' auth-anonymous=1',
                   '--load=module-null-sink sink_name=logh3_wine rate=44100 channels=2 format=float32le norewinds=yes']
        pulse = subprocess.Popen(command, env=env, stdout=plog, stderr=subprocess.STDOUT)
        deadline = time.monotonic() + 15
        while True:
            assert pulse.poll() is None, 'PulseAudio提早結束'
            try:
                info = control('info')
                break
            except subprocess.CalledProcessError:
                assert time.monotonic() < deadline, 'PulseAudio15秒就緒期限'
                time.sleep(.05)
        write('PULSE-INFO.json', {'text': info, 'sinks': json.loads(control('-f', 'json', 'list', 'sinks'))})
        capture_command = ['parec', '--device=logh3_wine.monitor', '--format=float32le',
                           '--rate=44100', '--channels=2', '--latency-msec=10']
        capture_started = time.monotonic()
        capture = subprocess.Popen(capture_command, env=env, stdout=raw, stderr=clog)
        time.sleep(.2)
        assert capture.poll() is None
        write('READY.json', {'pulse_command': command, 'capture_command': capture_command,
                            'server_in_wine': 'unix:/out/' + args.batch + '/pulse/native',
                            'sample_rate': 44100, 'channels': 2, 'format': 'float32le',
                            'server_start_monotonic': started, 'capture_start_monotonic': capture_started,
                            'version': subprocess.check_output(['pulseaudio', '--version'], text=True).strip(),
                            'physical_device_present': Path('/dev/snd').exists()})
        print('READY', args.batch, flush=True)
        samples = []
        deadline = started + 180
        while not (out / 'CLIENT-DONE.json').exists():
            assert pulse.poll() is None and capture.poll() is None
            assert time.monotonic() < deadline, '虛擬輸出180秒擁有程序期限'
            clients = json.loads(control('-f', 'json', 'list', 'sink-inputs'))
            samples.append({'elapsed': time.monotonic() - started, 'sink_inputs': clients})
            time.sleep(.1)
        time.sleep(.2)
        write('SERVER.json', {'elapsed': time.monotonic() - started, 'samples': samples,
                              'client': json.loads((out / 'CLIENT-DONE.json').read_text()),
                              'scope': 'WASAPI經Wine／PulseAudio虛擬輸出；非實體卡或人耳'})
        print('DONE monitor recording', flush=True)
    finally:
        terminate(capture)
        terminate(pulse)
