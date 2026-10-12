"""正常遊戲視窗實錄。僅以 X11 鍵鼠操作，不注入局面或跳入戰場。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import time
from PIL import Image, ImageChops

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--package', required=True)
p.add_argument('--out', required=True)
p.add_argument('--plan', required=True)
p.add_argument('--record', action='store_true')
p.add_argument('--follow', action='store_true', help='持續讀取追加的正常輸入，stop事件結束；上限420秒')
p.add_argument('--binary')
p.add_argument('--art', help='本輪正式HD素材；預設沿封包art，不覆寫舊包')
p.add_argument('--scale', type=float, default=2)
a = p.parse_args()
package, out = Path(a.package), Path(a.out)
assert os.getuid() != 0 and package.is_dir() and not out.exists()
out.mkdir()
assert out.stat().st_uid == os.getuid()
binary = Path(a.binary) if a.binary else package / 'gin3'
assert 1 <= a.scale <= 3
art = Path(a.art) if a.art else package/'art'
assert art.is_dir()
command = [str(binary), '-orig', str(package/'original'), '-data', str(package/'data'),
           '-art', str(art), '-graphics', 'hd', '-scale', str(a.scale), '-save', str(out/'save')]
trace = []
game_log = (out/'game.log').open('x')
proc = subprocess.Popen(command, stdout=game_log, stderr=subprocess.STDOUT)
recorder = None
started = time.monotonic()

def x(*args, timeout=15):
    return subprocess.check_output(['xdotool', *map(str, args)], text=True, timeout=timeout).strip()

try:
    window = x('search', '--sync', '--onlyvisible', '--name', '^銀河英雄傳說III SP$',timeout=60).splitlines()[-1]
    x('windowfocus', window)
    time.sleep(1)
    if a.record:
        record_command = ['ffmpeg', '-nostdin', '-y', '-loglevel', 'warning', '-threads', '2',
                          '-f', 'x11grab', '-draw_mouse', '1', '-framerate', '30',
                          '-video_size', f'{int(640*a.scale)}x{int(480*a.scale)}', '-i', os.environ['DISPLAY']+'.0+0,0',
                          '-t', '420', '-an', '-c:v', 'libx264', '-preset', 'veryfast',
                          '-crf', '18', '-threads', '2', '-pix_fmt', 'yuv420p', str(out/'RAW.mp4')]
        recorder = subprocess.Popen(record_command, stdout=(out/'record.log').open('x'), stderr=subprocess.STDOUT)
        time.sleep(.4)
    started = time.monotonic()
    position = 0
    while time.monotonic()-started < 415:
        plan = json.loads(Path(a.plan).read_text())
        if position >= len(plan):
            if a.follow:
                time.sleep(.3)
                continue
            break
        event = plan[position]
        position += 1
        begin = time.monotonic()-started
        kind = event['kind']
        if kind == 'stop':
            break
        if kind in ('click', 'move'):
            x('mousemove', '--window', window, int((event['x']+.5)*a.scale), int((event['y']+.5)*a.scale))
            time.sleep(.2)
            if kind == 'click':
                x('mousedown', event.get('button', 1), 'sleep', '.05', 'mouseup', event.get('button', 1))
        elif kind == 'key':
            x('keydown', event['key'], 'sleep', '.07', 'keyup', event['key'])
        elif kind == 'capture':
            subprocess.run(['import', '-window', window, str(out/(event['name']+'.png'))], check=True, timeout=15)
        elif kind == 'wait_image':
            # 只觀測真實視窗；畫面未到就不送後續命令，不注入遊戲狀態。
            box = tuple(event['box'])
            with Image.open(event['reference']) as source:
                expected = source.convert('RGB').crop(box)
            deadline = time.monotonic()+min(event.get('timeout', 180), 240)
            while True:
                observed = out/'WAIT.png'
                temporary = out/'WAIT-next.png'
                subprocess.run(['import', '-window', window, str(temporary)], check=True, timeout=15)
                temporary.replace(observed)
                with Image.open(observed) as source:
                    actual = source.convert('RGB').crop(box)
                if ImageChops.difference(expected, actual).getbbox() is None:
                    break
                if time.monotonic() >= deadline:
                    raise TimeoutError(f'正常視窗未達指定畫面：{event}')
                time.sleep(1)
        elif kind != 'wait':
            raise ValueError(event)
        time.sleep(event.get('wait', .35))
        trace.append(dict(event, start_seconds=begin, end_seconds=time.monotonic()-started))
        (out/'EVENTS.json').write_text(json.dumps(trace, ensure_ascii=False, indent=2)+'\n')
    if recorder:
        recorder.send_signal(signal.SIGINT)
        recorder.wait(timeout=30)
        # FFmpeg收到正常SIGINT常回255；必須另以ffprobe核實已完整封裝。
        assert recorder.returncode in (0, 255)
        subprocess.run(['ffprobe', '-v', 'error', '-show_entries', 'format=duration', str(out/'RAW.mp4')], check=True)
    result = {'command': command, 'binary_sha256': hashlib.sha256(binary.read_bytes()).hexdigest(),
              'source_index_sha256': hashlib.sha256((package/'original/SAVEDATA.IDX').read_bytes()).hexdigest(),
              'normal_x11_input': True, 'script_flag': False, 'state_injection': False,
              'audio': '畫面實錄；成片配樂另行剪輯', 'events': trace}
    if a.record:
        result['record_command'] = record_command
        result['raw_sha256'] = hashlib.sha256((out/'RAW.mp4').read_bytes()).hexdigest()
    (out/'TRACE.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print('正常 X11 操作完成', out, flush=True)
finally:
    for child in (recorder, proc):
        if child and child.poll() is None:
            child.terminate()
            try:
                child.wait(timeout=10)
            except subprocess.TimeoutExpired:
                child.kill()
                child.wait(timeout=5)
    game_log.close()
