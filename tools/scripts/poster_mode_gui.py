"""151正常視窗：海報開關、兩側防點穿、鍵盤、F7、縮放、存讀及重啟。Docker內執行。"""
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
    Image = ImageChops = None  # Wine影像不額外安裝套件，最後由既有影像容器核對PNG。

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--binary', required=True)
p.add_argument('--out', required=True)
p.add_argument('--windows', action='store_true')
p.add_argument('--original-only', action='store_true', help='不載HD renderer，驗原生640×480中央')
p.add_argument('--center-only', action='store_true', help='中央像素與側欄檢查後結束')
a = p.parse_args()
out = Path(a.out); assert not out.exists() and out.parent.stat().st_uid == os.getuid(); out.mkdir()
env = dict(os.environ, GOMAXPROCS='2', LP_NUM_THREADS='1', GOMEMLIMIT='1600MiB', GOGC='60')
events = []; proc = None; logs = []; trace = out/'FRAMES.jsonl'
sha = lambda f: hashlib.sha256(Path(f).read_bytes()).hexdigest()
def path(s): return 'Z:'+str(s).replace('/', '\\') if a.windows else str(s)
def x(*args): return subprocess.check_output(['xdotool', *map(str, args)], text=True, env=env, timeout=20).strip()


def observe():
    assert proc is not None and proc.poll() is None, '程式提早結束'
    if not trace.exists(): return {}
    with trace.open('rb') as f:
        end = f.seek(0, 2); f.seek(max(0, end-131072))
        for line in reversed(f.read().splitlines()):
            try: return json.loads(line)
            except json.JSONDecodeError: pass
    return {}


def wait(test, label, seconds=180):
    deadline = time.monotonic()+seconds
    while time.monotonic() < deadline:
        r = observe()
        if r and test(r): return r
        time.sleep(.04)
    raise AssertionError((label, observe()))


def ack(button=None, pressed=None):
    tick = observe().get('ticks', -1)
    return wait(lambda r: r['ticks'] >= tick+2 and (button is None or r['mouse_buttons'][0 if button == 1 else 1] == pressed), '輸入回授', 30)


def geometry():
    values = dict(line.split('=', 1) for line in x('getwindowgeometry', '--shell', window).splitlines() if '=' in line)
    return int(values['WIDTH']), int(values['HEIGHT'])


def move(gx, gy):
    w, h = geometry(); scale = h/480
    offset = (w-640*scale)/2 if observe()['poster_mode'] else 0
    x('mousemove', '--window', window, int(offset+(gx+.5)*scale), int((gy+.5)*scale)); ack()


def raw_click(px, py, button=1):
    x('mousemove', '--window', window, px, py); time.sleep(.07)
    x('mousedown', button); ack(button, True); x('mouseup', button); ack(button, False)


def click(gx, gy, button=1):
    move(gx, gy)
    x('mousedown', button); ack(button, True); x('mouseup', button); ack(button, False)


def key(name):
    x('keydown', name); ack(); x('keyup', name); ack()


def snap(name):
    move(600, 440); time.sleep(.4)
    file = out/(name+'.png'); subprocess.run(['import', '-window', window, str(file)], check=True, env=env, timeout=20)
    row = observe(); events.append(dict(name=name, screenshot=file.name, sha256=sha(file), observed=row, window=geometry()))
    print('PASS', name, geometry(), flush=True)


def state(name): return wait(lambda r: r['state'] == name, name)
def settings(): click(544, 125); click(560, 185); state('StateSettings')
def poster(enabled):
    rect = observe()['settings']['poster']; click((rect[0]+rect[2])//2, (rect[1]+rect[3])//2)
    wait(lambda r: r['poster_mode'] == enabled, '海報開關')
    time.sleep(.25)
    w, h = geometry(); assert abs(w/h-(16/9 if enabled else 4/3)) < .005


def start(stem, extra=None):
    global proc, trace, window
    trace = out/(stem+'-frames.jsonl')
    base = (['wine', path(a.binary)] if a.windows else [a.binary])+['-intro=false', '-graphics', 'original' if a.original_only else 'hd', '-scale', '1.5', '-music', '', '-save', path(out/'save'), '-frame-trace', path(trace)]
    if not a.windows: base += ['-orig', '/orig', '-data', '/src/docs/internals/data', '-art', '/src/assets/hd']
    base += extra or []
    log = (out/(stem+'.log')).open('w'); logs.append(log)
    proc = subprocess.Popen(base, env=env, stdout=log, stderr=subprocess.STDOUT)
    state('StateMainMenu')
    window = x('search', '--onlyvisible', '--name', '^銀河英雄傳說III SP$').splitlines()[-1]
    x('windowfocus', window); x('windowmove', '--sync', window, 0, 0)
    return base


def stop():
    global proc
    if proc is not None and proc.poll() is None:
        proc.terminate()
        try: proc.wait(timeout=5)
        except subprocess.TimeoutExpired: proc.kill(); proc.wait(timeout=5)
    proc = None
    if a.windows: subprocess.run(['wineserver', '-k'], env=env, check=False, timeout=10)


result = dict(passed=False, normal_entry=True, script_flag=False, state_injection=False, seed_injection=False, events=events)
try:
    command = start('first'); assert observe()['poster_available'] and not observe()['poster_mode']
    click(300, 230); state('StateCopyProtection')
    for k in ['1', '3', '8', 'Return']: key(k)
    click(300, 240, 3); click(300, 240, 3); state('StateCommandPhase'); snap('map-off')
    unchanged = observe()['strategic']
    settings(); snap('settings-off'); poster(True); snap('settings-on')
    before = observe()['strategic']
    for side in ['left', 'right']:
        for button in [1, 3]:
            w, h = geometry(); raw_click(20 if side == 'left' else w-20, h//2, button)
            assert observe()['state'] == 'StateSettings' and observe()['poster_mode'] and observe()['strategic'] == before
    snap('side-clicks-ignored')
    click(300, 240, 3); state('StateCommandPhase'); snap('map-on')
    assert observe()['strategic'] == unchanged
    if Image is not None:
        original = Image.open(out/'map-off.png').convert('RGB'); wide = Image.open(out/'map-on.png').convert('RGB')
        box = ((wide.width-original.width)//2, 0, (wide.width+original.width)//2, wide.height)
        diff = ImageChops.difference(original, wide.crop(box)); different = sum(max(p) > 0 for p in diff.getdata())
        result['center_comparison'] = dict(pixels=original.width*original.height, different=different, box=box, difference_box=diff.getbbox())
        diff.save(out/'CENTER-DIFF.png'); assert different == 0, result['center_comparison']
    else:
        result['center_comparison'] = dict(pending_external=True)
    if a.center_only:
        result.update(passed=True, command=command, binary_sha256=sha(a.binary), original_only=a.original_only, driver_sha256=sha(__file__))
        (out/'RESULT.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
        raise SystemExit(0)
    key('F7'); wait(lambda r: r['graphics'] == 'original' and r['poster_mode'], 'F7保留海報'); snap('original-on')
    key('F1'); wait(lambda r: r['help_visible'], 'F1說明'); snap('help-on'); key('F1')
    x('windowsize', '--sync', window, 1152, 648); ack(); settings(); assert geometry() == (1152, 648); snap('resized-settings')
    poster(False); snap('resized-off'); poster(True); click(300, 240, 3); state('StateCommandPhase')
    x('windowsize', '--sync', window, 1280, 720); ack(); key('F7'); wait(lambda r: r['graphics'] == 'hd', '回HD')
    # 公開UI存讀後，原存檔不包含海報欄位。
    click(544, 125); click(560, 165); state('StateSaveGame'); click(450, 152)
    wait(lambda r: (out/'save/SHIP.DT0').is_file(), '六表保存')
    click(300, 240, 3); click(544, 125); click(560, 145); state('StateLoadGame'); click(300, 132); state('StateLoadFaction'); click(318, 228)
    state('StateCommandPhase'); assert observe()['poster_mode']; snap('loaded-on')
    pref = json.loads((out/'save/preferences.json').read_text()); assert pref['poster_mode']
    result.update(command=command, binary_sha256=sha(a.binary), saved={p.name: sha(p) for p in sorted((out/'save').glob('*.DT0'))})
    assert len(result['saved']) == 6
    stop(); start('restart'); assert observe()['poster_mode']; snap('restart-on')
    # 游標停海報，防拷題仍可正常鍵盤輸入。
    click(300, 230); state('StateCopyProtection'); raw_click(20, 360, 1)
    for k in ['1', '3', '8', 'Return']: key(k)
    wait(lambda r: r['state'] != 'StateCopyProtection', '海報側鍵盤輸入'); snap('typing-from-side')
    stop(); empty = out/'empty-art'; empty.mkdir()
    start('missing', ['-art', path(empty), '-graphics', 'original', '-poster=true'])
    assert not observe()['poster_mode'] and not observe()['poster_available'] and geometry() == (960, 720)
    snap('missing-fallback')
    result.update(passed=True, driver_sha256=sha(__file__))
    (out/'RESULT.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
except Exception as error:
    result['failure'] = repr(error); (out/'FAILED.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n'); raise
finally:
    stop()
    for log in logs: log.close()
