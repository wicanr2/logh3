"""規格138：正式視窗的頂端功能列、停留與快捷鍵正常輸入驗收。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time
from PIL import Image, ImageChops, ImageStat

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--binary')
p.add_argument('--launcher')
p.add_argument('--package', required=True)
p.add_argument('--art', required=True)
p.add_argument('--out', required=True)
p.add_argument('--faction', type=int, choices=[0, 1], required=True)
a = p.parse_args()
assert bool(a.binary) != bool(a.launcher)
out, package = Path(a.out), Path(a.package)
assert package.is_dir() and not out.exists()
assert out.parent.stat().st_uid == os.getuid()
out.mkdir()
assert out.stat().st_uid == os.getuid()
env = dict(os.environ, GOMAXPROCS='2', LP_NUM_THREADS='2', GOMEMLIMIT='1200MiB', GOGC='50')
runtime = out/'pulse'
runtime.mkdir(mode=0o700)
env.update(XDG_RUNTIME_DIR=str(runtime), PULSE_SERVER='unix:'+str(runtime/'native'))
alsa = runtime/'alsa.conf'
alsa.write_text('</usr/share/alsa/alsa.conf>\npcm.!default { type pulse server "'+env['PULSE_SERVER']+'" }\n')
env['ALSA_CONFIG_PATH'] = str(alsa)
frames = out/'FRAMES.jsonl'
arguments = ['-save', str(out/'save'), '-frame-trace', str(frames)]
if a.launcher:
    command = [a.launcher]+arguments
else:
    command = [a.binary, '-orig', str(package/'original'), '-data', str(package/'data'),
               '-art', a.art, '-graphics', 'hd', '-scale', '1.5']+arguments
events, processes, handles = [], [], []
started = time.monotonic()
result = {'normal_entry': True, 'script_flag': False, 'state_injection': False,
          'command': command, 'faction': a.faction, 'events': events, 'passed': False}


def log(name):
    handle = (out/name).open('x')
    handles.append(handle)
    return handle


def x(*args, timeout=15):
    return subprocess.check_output(['xdotool', *map(str, args)], env=env, text=True, timeout=timeout).strip()


def observe():
    assert game.poll() is None, '遊戲提前退出'
    if not frames.exists():
        return {}
    with frames.open('rb') as stream:
        size = stream.seek(0, 2)
        stream.seek(max(0, size-16384))
        lines = stream.read().splitlines()
    for line in reversed(lines):
        try:
            return json.loads(line)
        except json.JSONDecodeError:
            pass
    return {}


def check(label, **expected):
    deadline = time.monotonic()+12
    while True:
        row = observe()
        if row and row['wall_ns'] > time.time_ns()-2_000_000_000 and all(row.get(k) == v for k, v in expected.items()):
            events.append({'check': label, 'expected': expected, 'observed': row,
                           'seconds': time.monotonic()-started})
            print(label, expected, flush=True)
            return row
        assert time.monotonic() < deadline, (label, expected, row)
        time.sleep(.15)


def move(px, py):
    x('mousemove', '--window', window, int((px+.5)*1.5), int((py+.5)*1.5))
    events.append({'move': [px, py]})
    time.sleep(.3)


def click(px, py, button=1):
    move(px, py)
    x('mousedown', button, 'sleep', '.05', 'mouseup', button)
    events.append({'click': [px, py, button]})
    time.sleep(.3)


def key(name):
    x('keydown', name, 'sleep', '.07', 'keyup', name)
    events.append({'key': name})
    time.sleep(.25)


def snap(name):
    path = out/(name+'.png')
    subprocess.run(['import', '-window', window, str(path)], env=env, check=True, timeout=15)
    events.append({'screenshot': path.name, 'sha256': hashlib.sha256(path.read_bytes()).hexdigest(),
                   'observed': observe()})


try:
    pulse = subprocess.Popen(['pulseaudio', '--daemonize=no', '--use-pid-file=no', '--exit-idle-time=-1',
        '--disable-shm=yes', '--high-priority=no', '--realtime=no', '-n',
        '--load=module-native-protocol-unix socket='+str(runtime/'native')+' auth-anonymous=1',
        '--load=module-null-sink sink_name=logh3 rate=44100 channels=2'], env=env,
        stdout=log('pulse.log'), stderr=subprocess.STDOUT)
    processes.append(pulse)
    deadline = time.monotonic()+10
    while subprocess.run(['pactl', 'info'], env=env, capture_output=True, timeout=3).returncode:
        assert pulse.poll() is None and time.monotonic() < deadline
        time.sleep(.1)
    game = subprocess.Popen(command, cwd='/tmp', env=env, stdout=log('game.log'), stderr=subprocess.STDOUT)
    processes.append(game)
    window = x('search', '--sync', '--onlyvisible', '--name', '^銀河英雄傳說III SP$', timeout=60).splitlines()[-1]
    events.append({'window_ready_seconds': time.monotonic()-started})
    x('windowfocus', window)
    time.sleep(1)
    click(600, 440)
    click(600, 440)
    check('正常主選單', state='StateMainMenu', graphics='hd')
    move(320, 2)
    check('頂緣展開', toolbar_visible=True)
    move(320, 15)
    check('列內保持', toolbar_visible=True)
    snap('01-toolbar')
    click(80, 15)
    check('點擊F1', help_visible=True, state='StateMainMenu')
    snap('02-help')
    click(300, 210)
    key('1')
    key('Return')
    check('說明攔截遊戲輸入', help_visible=True, state='StateMainMenu')
    key('Escape')
    move(600, 420)
    check('Escape關閉並離開收起', help_visible=False, toolbar_visible=False, state='StateMainMenu')
    for button in (1, 3):
        move(320, 2)
        x('mousedown', button, 'sleep', '.12')
        check('列內按住'+str(button), toolbar_visible=True, state='StateMainMenu')
        move(300, 210)
        x('mouseup', button)
        time.sleep(.3)
        check('列內拖出釋放不點穿'+str(button), state='StateMainMenu', help_visible=False)
    click(300, 210+20*a.faction)
    check('選擇陣營', state='StateCopyProtection')
    for name in '138':
        key(name)
    click(432, 270)
    check('第二畫面確認', state='StateMainGame')
    move(600, 420)
    snap('03-account-base')
    move(448, 350)
    snap('04-account-hover')
    # 左半部遠離游標，驗真正的按鈕反光，不能只因指標移動就通過。
    region = tuple(round(v*1.5) for v in (432, 342, 446, 356))
    with Image.open(out/'03-account-base.png') as base, Image.open(out/'04-account-hover.png') as hover:
        before, after = base.convert('RGB').crop(region), hover.convert('RGB').crop(region)
        diff = ImageChops.difference(before, after)
        changed = sum(max(pixel) >= 10 for pixel in diff.getdata())
        brightness = sum(ImageStat.Stat(after).mean)/3-sum(ImageStat.Stat(before).mean)/3
    result['account_hover'] = {'region': region, 'changed_pixels': changed, 'brightness_increase': brightness}
    assert changed >= 60 and brightness >= 6, result['account_hover']
    click(448, 350)
    check('財政變更', state='StateBudgetPlan')
    move(487, 329)
    snap('05-budget-hover')
    move(600, 420)
    snap('06-budget-leave')
    click(463, 370)
    check('財政結束', state='StateMainGame')
    key('F7')
    check('F7切原版', graphics='original', state='StateMainGame')
    snap('07-original')
    key('F7')
    check('F7切HD', graphics='hd', state='StateMainGame')
    key('F8')
    check('F8不切敵軍', show_enemies=False, state='StateMainGame')
    move(320, 2)
    click(490, 15)
    check('舊敵軍位置無動作', show_enemies=False, state='StateMainGame')
    click(250, 15)
    check('功能列切原版', graphics='original', state='StateMainGame')
    click(250, 15)
    check('功能列切HD', graphics='hd', state='StateMainGame')
    key('F1')
    check('F1說明開啟', help_visible=True, state='StateMainGame')
    click(320, 348)
    check('關閉說明不點穿', help_visible=False, state='StateMainGame')
    move(600, 420)
    click(448, 370)
    check('訂造確認', state='StateShipOrder')
    move(448, 350)
    snap('08-order-hover')
    click(448, 350)
    check('建造變更', state='StateBuildPlan')
    move(232, 210)
    snap('09-build-hover')
    move(600, 420)
    snap('10-build-leave')
    click(512, 370)
    check('建造結束', state='StateShipOrder')
    click(448, 370)
    check('進入戰略指令', state='StateCommandPhase')
    move(600, 420)
    snap('11-strategy')
    click(544, 125)
    click(560, 185)
    check('戰略設定入口', state='StateSettings', show_enemies=False)
    click(232, 370)
    check('設定開啟敵軍', state='StateSettings', show_enemies=True)
    snap('12-settings-on')
    key('F8')
    check('設定內F8無動作', state='StateSettings', show_enemies=True)
    click(416, 370)
    check('設定關閉保持偏好', state='StateCommandPhase', show_enemies=True)
    click(544, 125)
    click(560, 185)
    click(232, 370)
    check('設定關閉敵軍', state='StateSettings', show_enemies=False)
    snap('13-settings-off')
    click(416, 370)
    check('返回戰略', state='StateCommandPhase', show_enemies=False)
    result['passed'] = True
finally:
    result['elapsed_seconds'] = time.monotonic()-started
    for child in reversed(processes):
        if child.poll() is None:
            child.terminate()
            try:
                child.wait(timeout=5)
            except subprocess.TimeoutExpired:
                child.kill()
                child.wait(timeout=5)
    for handle in handles:
        handle.close()
    (out/'RESULT.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
