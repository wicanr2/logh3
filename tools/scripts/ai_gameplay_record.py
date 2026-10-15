"""第126輪：正常X11鍵鼠、正式遊戲音訊及完整視窗錄影，供AI／S1影片。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import threading
import time

from PIL import Image, ImageChops

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--package', required=True)
p.add_argument('--binary', help='可選待驗二進位；仍走正常主選單入口')
p.add_argument('--art', help='可選待驗素材目錄')
p.add_argument('--graphics', choices=['original', 'hd'], default='hd')
p.add_argument('--out', required=True)
p.add_argument('--plan', required=True)
p.add_argument('--references', help='可選正常模型PNG目錄；存在的同名圖做完整核對')
p.add_argument('--ai', type=int, choices=[-1, 0, 1, 2, 3], default=-1)
p.add_argument('--scale', type=float, default=2)
p.add_argument('--limit', type=int, default=1000)
p.add_argument('--fps', type=int, choices=[30,60], default=30)
p.add_argument('--frame-trace', action='store_true', help='保存正常顯示節奏的唯讀收據')
a = p.parse_args()
package, out = Path(a.package), Path(a.out)
assert os.getuid() == 1000 and package.is_dir() and not out.exists()
assert 1 <= a.scale <= 3 and 60 <= a.limit <= 1200
assert out.parent.stat().st_uid == os.getuid()
out.mkdir()
runtime = Path('/tmp/logh3-promo-pulse')
runtime.mkdir(mode=0o700)
env = dict(os.environ, XDG_RUNTIME_DIR=str(runtime), PULSE_SERVER='unix:'+str(runtime/'native'))
env.update(GOMEMLIMIT='1200MiB', GOGC='50', GOMAXPROCS='2', LP_NUM_THREADS='2')
alsa = runtime/'alsa.conf'
alsa.write_text('</usr/share/alsa/alsa.conf>\npcm.!default { type pulse server "'+env['PULSE_SERVER']+'" }\n')
env['ALSA_CONFIG_PATH'] = str(alsa)
processes, handles, trace = [], [], []
monitor_stop=threading.Event()


def monitor():
    with (out/'RESOURCES.jsonl').open('x') as log:
        while not monitor_stop.wait(2):
            row={'utc':time.time(),'processes':[]}
            for key in ('memory.current','memory.peak','memory.events'):
                path=Path('/sys/fs/cgroup')/key
                if path.exists():row[key]=path.read_text()
            for child in list(processes):
                status=Path('/proc')/str(child.pid)/'status'
                lines=status.read_text().splitlines() if status.exists() else []
                row['processes'].append({'pid':child.pid,'returncode':child.poll(),
                    'memory':[s for s in lines if s.startswith(('Name:','VmRSS:','VmSize:'))]})
            log.write(json.dumps(row)+'\n');log.flush()


def terminate_requested(signum, frame):
    raise RuntimeError('錄製收到SIGTERM，保存失敗收據並關閉子程序')


signal.signal(signal.SIGTERM,terminate_requested)
monitor_thread=threading.Thread(target=monitor,daemon=True)
monitor_thread.start()


def logfile(name):
    f = (out/name).open('x')
    handles.append(f)
    return f


def run(*args, **kwargs):
    return subprocess.run([str(v) for v in args], check=True, env=env, **kwargs)


def x(*args, timeout=15):
    return subprocess.check_output(['xdotool', *map(str, args)], env=env, text=True, timeout=timeout).strip()


def write(name, data):
    (out/name).write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')

def trace_value(row, key):
    for part in key.split('.'):
        row = row.get(part) if isinstance(row, dict) else None
    return row


def latest_frame(path):
    # 目前觀測含分隊清單，一行可能大於舊版8KB；往前讀到完整JSON。
    with path.open('rb') as frame_log:
        end=frame_log.seek(0,2)
        window=min(end,65536)
        while window:
            frame_log.seek(end-window)
            for line in reversed(frame_log.read(window).splitlines()):
                try:return json.loads(line)
                except json.JSONDecodeError:continue
            if window==end:break
            window=min(end,window*2)
    return {}


def screenshot(path):
    run('import', '-window', window, path, timeout=15)
    with Image.open(path) as im:
        return im.convert('RGB')


def expected(path, size):
    with Image.open(path) as im:
        assert im.size == (1920, 1440)
        return im.convert('RGB').resize(size, Image.Resampling.NEAREST)


try:
    pulse = subprocess.Popen(['pulseaudio', '--daemonize=no', '--use-pid-file=no', '--exit-idle-time=-1',
                              '--disable-shm=yes', '--high-priority=no', '--realtime=no', '--log-level=warning', '-n',
                              '--load=module-native-protocol-unix socket='+str(runtime/'native')+' auth-anonymous=1',
                              '--load=module-null-sink sink_name=logh3 rate=44100 channels=2 format=float32le norewinds=yes'],
                             env=env, stdout=logfile('pulse.log'), stderr=subprocess.STDOUT)
    processes.append(pulse)
    deadline = time.monotonic()+15
    while True:
        assert pulse.poll() is None, 'PulseAudio提前結束'
        ready = subprocess.run(['pactl', 'info'], env=env, capture_output=True, text=True, timeout=3)
        if ready.returncode == 0:
            (out/'pulse-info.txt').write_text(ready.stdout)
            break
        assert time.monotonic() < deadline
        time.sleep(.1)
    command = [a.binary or str(package/'gin3'), '-orig', str(package/'original'), '-data', str(package/'data'),
               '-art', a.art or str(package/'art'), '-graphics', a.graphics, '-music', str(package/'music'),
               '-ai', str(a.ai), '-scale', str(a.scale), '-save', str(out/'save')]
    if a.frame_trace:
        command += ['-frame-trace', str(out/'FRAMES.jsonl')]
    game = subprocess.Popen(command, env=env, stdout=logfile('game.log'), stderr=subprocess.STDOUT)
    processes.append(game)
    window = x('search', '--sync', '--onlyvisible', '--name', '^銀河英雄傳說III SP$', timeout=60).splitlines()[-1]
    x('windowfocus', window)
    x('windowmove', window, 0, 0)
    geometry = dict(line.split('=',1) for line in x('getwindowgeometry','--shell',window).splitlines() if '=' in line)
    assert (int(geometry['X']),int(geometry['Y']),int(geometry['WIDTH']),int(geometry['HEIGHT'])) == (0,0,int(640*a.scale),int(480*a.scale)), geometry
    time.sleep(1)
    record_command = ['ffmpeg', '-nostdin', '-y', '-loglevel', 'warning', '-thread_queue_size', '8',
                      '-f', 'x11grab', '-draw_mouse', '0', '-framerate', str(a.fps),
                      '-video_size', f'{int(640*a.scale)}x{int(480*a.scale)}', '-i', os.environ['DISPLAY']+'.0+0,0',
                      '-thread_queue_size', '64', '-f', 'pulse', '-i', 'logh3.monitor',
                      '-t', str(a.limit), '-af', 'aresample=async=1:first_pts=0',
                      '-c:v', 'libx264', '-preset', 'ultrafast', '-crf', '18', '-threads', '2',
                      '-pix_fmt', 'yuv420p', '-c:a', 'aac', '-b:a', '192k',
                      '-movflags', '+frag_keyframe+empty_moov+default_base_moof', str(out/'RAW.mp4')]
    record_started_utc_ns = time.time_ns()
    recorder = subprocess.Popen(record_command, env=env, stdout=logfile('record.log'), stderr=subprocess.STDOUT)
    processes.append(recorder)
    time.sleep(.5)
    input_started_utc_ns = time.time_ns()
    started = time.monotonic()
    # 正常片頭跳過，沒有-script、direct-entry或局面／seed注入。
    events = [{'kind':'click','x':600,'y':440}, {'kind':'click','x':600,'y':440}]
    plan = json.loads(Path(a.plan).read_text())
    if plan[:2] == events:
        events = []
    events += plan
    for event in events:
        assert time.monotonic()-started < a.limit-5
        assert game.poll() is None and recorder.poll() is None and pulse.poll() is None, ('程序提前結束',game.poll(),recorder.poll(),pulse.poll())
        begin = time.monotonic()-started
        wall_begin_ns = time.time_ns()
        kind = event['kind']
        if kind == 'stop':
            break
        if kind in ('move', 'click'):
            x('mousemove', '--window', window, int((event['x']+.5)*a.scale), int((event['y']+.5)*a.scale))
            time.sleep(.2)
            if kind == 'click':
                # 120ms低於137ms連發閘門；錄製期間避免多個HD工作負載爭用CPU。
                x('mousedown', event.get('button', 1), 'sleep', '.12', 'mouseup', event.get('button', 1))
        elif kind == 'key':
            x('keydown', event['key'], 'sleep', '.07', 'keyup', event['key'])
        elif kind == 'wait_image':
            reference, box = Path(event['reference']), event['box']
            assert 0 <= box[0] < box[2] <= 1920 and 0 <= box[1] < box[3] <= 1440
            deadline = min(started+a.limit-5, time.monotonic()+event.get('timeout', 240))
            while True:
                actual = screenshot(out/'WAIT.png')
                expected_image = expected(reference, actual.size)
                scaled_box = tuple(round(v*actual.width/1920) for v in box)
                if ImageChops.difference(expected_image.crop(scaled_box), actual.crop(scaled_box)).getbbox() is None:
                    break
                assert time.monotonic() < deadline, ('未到正常畫面', event)
                time.sleep(.7)
        elif kind == 'wait_state':
            assert a.frame_trace, 'wait_state需要唯讀顯示收據'
            deadline = min(started+a.limit-5, time.monotonic()+event.get('timeout', 180))
            while True:
                row = latest_frame(out/'FRAMES.jsonl')
                if row.get('state') == event['state'] and (not event.get('animation') or row.get('animations')) and all(trace_value(row,k) == v for k, v in event.get('expect', {}).items()):
                    break
                combat = row.get('tactical') or {}
                if event.get('auto_execute') and row.get('battle_ready') and combat.get('turn',0) < event.get('expect',{}).get('tactical.turn',0):
                    command_begin=time.time_ns()
                    x('mousemove','--window',window,int(600.5*a.scale),int(242.5*a.scale))
                    x('mousedown',1,'sleep','.07','mouseup',1)
                    trace.append({'kind':'auto_execute','point':[600,242],
                                  'wall_begin_ns':command_begin,'wall_end_ns':time.time_ns()})
                    time.sleep(.15)
                if event.get('auto_combat') and row.get('state') == 'StateBattle' and combat.get('combat_input'):
                    # 回授式真實滑鼠輸入，沿畫面「自動」開火；不寫遊戲狀態。
                    point = [240,470] if combat['combat_input']=='target' else combat['combat_point']
                    command_begin = time.time_ns()
                    x('mousemove','--window',window,int((point[0]+.5)*a.scale),int((point[1]+.5)*a.scale))
                    x('mousedown',1,'sleep','.07','mouseup',1)
                    trace.append({'kind':'auto_combat','input':combat['combat_input'],'point':point,
                                  'wall_begin_ns':command_begin,'wall_end_ns':time.time_ns()})
                assert game.poll() is None and time.monotonic() < deadline, ('未到正常狀態', event, row)
                time.sleep(.05)
        elif kind == 'capture':
            name = event['name']
            model = Path(a.references)/(name+'.png') if a.references else None
            for attempt in range(3):
                actual = screenshot(out/(name+'.png'))
                if model is None or not model.exists():
                    break
                diff = ImageChops.difference(expected(model, actual.size), actual)
                channels = diff.split()
                changed = ImageChops.lighter(ImageChops.lighter(channels[0],channels[1]),channels[2])
                result = {'pixels':actual.width*actual.height, 'different':sum(changed.point(lambda n:255 if n else 0).histogram()[1:]),
                          'bbox':changed.getbbox(), 'attempt':attempt+1, 'model_sha256':hashlib.sha256(model.read_bytes()).hexdigest(),
                          'scope':'完整正常X11視窗與模型；不是DOS原版parity'}
                write(name+'-comparison.json', result)
                if result['different'] == 0:
                    break
                diff.save(out/(name+'-difference.png'))
                for suffix in ('.png','-comparison.json','-difference.png'):
                    (out/(name+suffix)).rename(out/(name+f'-attempt-{attempt+1}'+suffix))
                assert attempt < 2, (name, result)
                time.sleep(.7)
        elif kind != 'wait':
            raise ValueError(event)
        time.sleep(event.get('wait', .7))
        trace.append(dict(event, start_seconds=begin, end_seconds=time.monotonic()-started,
                          wall_begin_ns=wall_begin_ns, wall_end_ns=time.time_ns()))
        write('EVENTS.json', trace)
        print(kind, event.get('name', event.get('key','')), round(time.monotonic()-started,2), flush=True)
    recorder.send_signal(signal.SIGINT)
    recorder.wait(timeout=30)
    assert recorder.returncode in (0,255)
    probe=json.loads(subprocess.check_output(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(out/'RAW.mp4')]))
    assert {s['codec_type'] for s in probe['streams']} >= {'video','audio'}
    write('FFPROBE.json',probe)
    write('TRACE.json', {'command':command,'record_command':record_command,'normal_x11_input':True,
                        'script_flag':False,'state_injection':False,'seed_injection':False,'ai':a.ai,
                        'audio':'正式ALSA後端→PulseAudio有時鐘虛擬輸出→同次FFmpeg錄製遊戲配樂及音效；未代稱實體音效卡或人耳',
                        'runtime_environment':{k:env[k] for k in ('GOMEMLIMIT','GOGC','GOMAXPROCS','LP_NUM_THREADS')},
                        'input_started_utc_ns':input_started_utc_ns,
                        'record_started_utc_ns':record_started_utc_ns,
                        'frame_trace_sha256':hashlib.sha256((out/'FRAMES.jsonl').read_bytes()).hexdigest() if a.frame_trace else None,
                        'binary_sha256':hashlib.sha256(Path(a.binary or package/'gin3').read_bytes()).hexdigest(),
                        'raw_sha256':hashlib.sha256((out/'RAW.mp4').read_bytes()).hexdigest(),'events':trace})
finally:
    monitor_stop.set()
    monitor_thread.join(timeout=3)
    if not (out/'TRACE.json').exists():
        memory=Path('/sys/fs/cgroup/memory.events')
        write('FAILURE-TRACE.json',{'events':trace,'normal_x11_input':True,
              'process_returncodes':[p.poll() for p in processes],
              'memory_events':memory.read_text() if memory.exists() else None})
    for child in reversed(processes):
        if child.poll() is None:
            child.terminate()
            try:child.wait(timeout=5)
            except subprocess.TimeoutExpired:child.kill();child.wait(timeout=5)
    for handle in handles:
        handle.close()
    if a.frame_trace and (out/'TRACE.json').exists():
        write('FRAME-TRACE-CLOSE.json', {
            'frame_trace_sha256':hashlib.sha256((out/'FRAMES.jsonl').read_bytes()).hexdigest(),
            'finalized_after_child_termination':True,
            'note':'TRACE.json在關閉程序前保存，FRAMES.jsonl的最終雜湊以本檔為準。'})
