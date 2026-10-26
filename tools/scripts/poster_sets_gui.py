"""150：實際視窗五語、三組海報、中央像素、輸入與六表存讀。Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import time

try:
    from PIL import Image, ImageChops
except ImportError:
    Image = ImageChops = None

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--binary', required=True); p.add_argument('--out', required=True)
p.add_argument('--package'); p.add_argument('--windows', action='store_true')
p.add_argument('--video', action='store_true'); p.add_argument('--original-only', action='store_true')
p.add_argument('--probe', action='store_true', help='只跑英文外觀與中央比較，不當正式完整驗收')
p.add_argument('--probe-locale', choices=['zh-TW','zh-CN','ja','en','ko'], default='en')
p.add_argument('--faction', type=int, choices=[0,1], default=1)
a = p.parse_args()
out = Path(a.out); assert not out.exists() and out.parent.stat().st_uid == os.getuid(); out.mkdir()
sha = lambda f: hashlib.sha256(Path(f).read_bytes()).hexdigest()
env = dict(os.environ, GOMAXPROCS='2', LP_NUM_THREADS='1', GOMEMLIMIT='1800MiB', GOGC='70')
events = []; proc = recorder = None; logs = []; trace = out/'FRAMES.jsonl'
package = Path(a.package) if a.package else None
menu_data = package/'data' if package else Path('/data' if a.windows else '/src/docs/internals/data')
menu_rows = json.loads((menu_data/'screen-main-menu.json').read_text())['text']['rows_y']
entry_y = menu_rows[a.faction]+8
def path(v): return 'Z:'+str(v).replace('/', '\\') if a.windows else str(v)
def x(*v): return subprocess.check_output(['xdotool', *map(str, v)], env=env, text=True, timeout=20).strip()


def observe():
    assert proc is not None and proc.poll() is None, '遊戲提早退出'
    if not trace.exists(): return {}
    with trace.open('rb') as f:
        end = f.seek(0, 2); f.seek(max(0, end-1024*1024))
        for line in reversed(f.read().splitlines()):
            try: return json.loads(line)
            except json.JSONDecodeError: pass
    return {}


def wait(test, name, seconds=180):
    until = time.monotonic()+seconds
    while time.monotonic() < until:
        r = observe()
        if r and test(r): return r
        time.sleep(.04)
    raise AssertionError((name, observe()))


def state(value): return wait(lambda r: r['state'] == value, value)
def ack(button=None, pressed=None):
    tick = observe().get('ticks', -1)
    return wait(lambda r: r['ticks'] >= tick+2 and (button is None or r['mouse_buttons'][0 if button == 1 else 1] == pressed), '輸入回授', 40)


def geometry():
    v = dict(line.split('=', 1) for line in x('getwindowgeometry', '--shell', window).splitlines() if '=' in line)
    return int(v['WIDTH']), int(v['HEIGHT'])


def move(gx, gy):
    w, h = geometry(); scale = h/480; offset = (w-640*scale)/2 if observe()['poster_mode'] else 0
    x('mousemove', '--window', window, int(offset+(gx+.5)*scale), int((gy+.5)*scale)); ack()


def raw_click(px, py, button=1):
    x('mousemove', '--window', window, px, py); time.sleep(.07)
    x('mousedown', button); ack(button, True); x('mouseup', button); ack(button, False)


def click(gx, gy, button=1):
    move(gx, gy); x('mousedown', button); ack(button, True); x('mouseup', button); ack(button, False)


def key(value):
    x('keydown', value); ack(); x('keyup', value); ack()


def snap(name):
    # Wine視窗改寬後，相同座標不一定產生新的X11移動事件。
    # 先移到鄰點，再移回固定位置，避免把前一個星球的hover留在截圖。
    move(598, 438); time.sleep(.2); move(600, 440); time.sleep(.8)
    f = out/(name+'.png'); subprocess.run(['import', '-window', window, str(f)], env=env, check=True, timeout=20)
    began = time.time_ns(); observed = observe()
    events.append(dict(name=name, screenshot=f.name, sha256=sha(f), observed=observed, window=geometry(), wall_ns=began))
    print('PASS', name, geometry(), flush=True)
    if a.video: time.sleep(3.5)


def control(name):
    r = observe()['settings'][name]; click((r[0]+r[2])//2, (r[1]+r[3])//2)


def settings(): click(544, 125); click(560, 185); state('StateSettings')
def close(): click(300, 240, 3); state('StateCommandPhase')


def poster(mode):
    for _ in range(4):
        r = observe(); current = r['poster_set'] if r['poster_mode'] else 'off'
        if current == mode: return
        control('poster'); ack(); time.sleep(.3)
    raise AssertionError(('poster cycle', mode, observe()))


def locale(code):
    for _ in range(5):
        if observe()['locale'] == code: return
        control('locale'); ack(); time.sleep(.3)
    raise AssertionError(('locale cycle', code, observe()))


def start(stem):
    global proc, trace, window
    trace = out/(stem+'-frames.jsonl')
    command = (['wine', path(a.binary)] if a.windows else [a.binary])+['-intro=false', '-graphics', 'original' if a.original_only else 'hd', '-scale', '1.5', '-music', '', '-save', path(out/'save'), '-frame-trace', path(trace)]
    if package: command += ['-orig', path(package/'original'), '-data', path(package/'data'), '-art', path(package/'art')]
    elif not a.windows: command += ['-orig', '/orig', '-data', '/src/docs/internals/data', '-art', '/src/assets/hd']
    log = (out/(stem+'.log')).open('w'); logs.append(log)
    proc = subprocess.Popen(command, env=env, stdout=log, stderr=subprocess.STDOUT)
    state('StateMainMenu'); window = x('search', '--onlyvisible', '--name', '^銀河英雄傳說III SP$').splitlines()[-1]
    x('windowfocus', window); x('windowmove', '--sync', window, 0, 0)
    return command


def stop():
    global proc, recorder
    if recorder is not None:
        recorder.send_signal(signal.SIGINT); recorder.wait(timeout=20); assert recorder.returncode in (0, 255); recorder = None
    if proc is not None and proc.poll() is None:
        proc.terminate()
        try: proc.wait(timeout=5)
        except subprocess.TimeoutExpired: proc.kill(); proc.wait(timeout=5)
    proc = None
    if a.windows: subprocess.run(['wineserver', '-k'], env=env, check=False, timeout=10)


result = dict(passed=False, normal_entry=True, state_injection=False, seed_injection=False, script_flag=False, events=events, driver_sha256=sha(__file__))
try:
    command = start('first'); assert observe()['locale'] == 'zh-TW' and not observe()['poster_mode']
    if a.video:
        log = (out/'record.log').open('w'); logs.append(log)
        result['record_started_utc_ns'] = time.time_ns()
        recorder = subprocess.Popen(['ffmpeg', '-nostdin', '-v', 'warning', '-f', 'x11grab', '-draw_mouse', '0', '-framerate', '60', '-video_size', '1280x720', '-i', os.environ['DISPLAY']+'.0+0,0', '-t', '900', '-an', '-c:v', 'libx264', '-preset', 'ultrafast', '-crf', '18', '-threads', '2', '-pix_fmt', 'yuv420p', str(out/'RAW.mp4')], env=env, stdout=log, stderr=subprocess.STDOUT)
        time.sleep(.5)
    snap('main-default'); click(300, entry_y); state('StateCopyProtection')
    for k in ['1', '3', '8', 'Return']: key(k)
    click(300, 240, 3); click(300, 240, 3); state('StateCommandPhase')
    initial = observe()['strategic']; comparisons = []
    languages = [a.probe_locale] if a.probe else ['zh-TW', 'zh-CN', 'ja', 'en', 'ko']
    for lang in languages:
        settings(); locale(lang); poster('off'); snap(lang+'-settings-off'); close(); snap(lang+'-map-off')
        assert observe()['strategic'] == initial
        settings(); poster('v1'); snap(lang+'-settings-v1'); close(); snap(lang+'-map-v1')
        assert observe()['strategic'] == initial
        if Image is not None:
            off = Image.open(out/(lang+'-map-off.png')).convert('RGB'); on = Image.open(out/(lang+'-map-v1.png')).convert('RGB')
            diff = ImageChops.difference(off, on.crop((160, 0, 1120, 720))); assert diff.getbbox() is None, lang
            comparisons.append(dict(locale=lang, pixels=off.width*off.height, different=0))
        key('F1'); wait(lambda r: r['help_visible'], 'help'); snap(lang+'-help'); key('F1')
        if lang in ['en', 'ko'] and not a.original_only:
            key('F7'); wait(lambda r: r['graphics'] == 'original', 'F7'); snap(lang+'-original'); key('F7')
        if not a.probe:
            settings(); control('admirals'); wait(lambda r: r.get('admiral_editor') is not None, 'admiral editor'); snap(lang+'-editor-basic')
            click(516, 180); snap(lang+'-editor-special'); click(300, 240, 3)
            wait(lambda r: r.get('admiral_editor') is None, 'close editor'); close()
        settings(); poster('v2'); close(); snap(lang+'-map-v2'); settings(); poster('v3'); close(); snap(lang+'-map-v3')
    if a.probe:
        result.update(passed=True, probe=True, command=command, binary_sha256=sha(a.binary), center_comparisons=comparisons)
        (out/'RESULT.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
        raise SystemExit(0)
    settings(); before = observe()['strategic']
    for side in ['left', 'right']:
        for button in [1, 3]:
            w, h = geometry(); raw_click(20 if side == 'left' else w-20, h//2, button)
            assert observe()['state'] == 'StateSettings' and observe()['strategic'] == before
    snap('side-clicks-ignored'); close()
    # Save/load through the public six-table UI; all five languages used same state.
    click(544, 125); click(560, 165); state('StateSaveGame'); snap('ko-save'); click(450, 152)
    wait(lambda r: (out/'save/SHIP.DT0').is_file(), 'six tables')
    click(300, 240, 3); click(544, 125); click(560, 145); state('StateLoadGame'); snap('ko-load'); click(300, 132)
    state('StateLoadFaction'); click(318, 188 if a.faction == 0 else 228); state('StateCommandPhase'); snap('ko-loaded')
    assert observe()['locale'] == 'ko' and observe()['poster_set'] == 'v3'
    result.update(command=command, binary_sha256=sha(a.binary), center_comparisons=comparisons,
                  saved={f.name: sha(f) for f in sorted((out/'save').glob('*.DT0'))}, preference=json.loads((out/'save/preferences.json').read_text()))
    assert len(result['saved']) == 6
    stop(); start('restart'); assert observe()['locale'] == 'ko' and observe()['poster_mode'] and observe()['poster_set'] == 'v3'; snap('ko-restart-main')
    click(300, entry_y); state('StateCopyProtection'); raw_click(20, 360)
    for k in ['1', '3', '8', 'Return']: key(k)
    wait(lambda r: r['state'] != 'StateCopyProtection', 'side keyboard'); snap('ko-typing-from-side')
    stop()
    if a.video: result['video_sha256'] = sha(out/'RAW.mp4')
    result.update(passed=True)
    (out/'RESULT.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
except Exception as error:
    result['failure'] = repr(error); (out/'FAILED.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n'); raise
finally:
    stop()
    for log in logs: log.close()
