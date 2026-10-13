"""規格141／142：系統視窗游標、取消、存讀與啟動的正常玩家驗收。"""
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
p.add_argument('--music')
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
arguments = ['-music', a.music or str(package/'music'), '-save', str(out/'save'), '-frame-trace', str(frames)]
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
    recorder = subprocess.Popen(['ffmpeg','-nostdin','-v','error','-f','pulse','-i','logh3.monitor','-t','180',
        '-c:a','pcm_s16le',str(out/'AUDIO.wav')],env=env,stdout=log('audio.log'),stderr=subprocess.STDOUT)
    processes.append(recorder)
    result['audio_start_seconds'] = time.monotonic()-started
    game = subprocess.Popen(command, cwd='/tmp', env=env, stdout=log('game.log'), stderr=subprocess.STDOUT)
    processes.append(game)
    loading = x('search', '--sync', '--onlyvisible', '--name', '載入中$', timeout=60).splitlines()[-1]
    subprocess.run(['import', '-window', loading, str(out/'00-loading.png')], env=env, check=True, timeout=15)
    events.append({'loading_window': loading, 'seconds': time.monotonic()-started})
    window = x('search' , '--sync', '--onlyvisible', '--name', '^銀河英雄傳說III SP$', timeout=60).splitlines()[-1]
    events.append({'window_ready_seconds': time.monotonic()-started})
    x('windowfocus', window)
    time.sleep(1)
    click(600, 440)
    click(600, 440)
    check('正常主選單', state='StateMainMenu', graphics='hd')
    if a.launcher:
        check('啟動器音色預設',music_quality='hd' if 'enhanced' in Path(a.launcher).name else 'original')

    click(300, 270)
    check('主選單讀檔問答', state='StateCopyProtection')
    for name in '138':
        key(name)
    click(432, 270)
    check('主選單讀檔', state='StateLoadGame')
    move(300, 240)
    snap('01-main-load-a')
    move(400, 240)
    snap('02-main-load-b')
    click(400, 240, button=3)
    check('取消回主選單', state='StateMainMenu')
    snap('03-main-cancel')
    click(300, 210+20*a.faction)
    check('陣營', state='StateCopyProtection')
    for name in '138':
        key(name)
    click(432, 270)
    check('會計', state='StateMainGame')
    click(448, 370)
    check('訂造', state='StateShipOrder')
    click(448, 370)
    check('戰略', state='StateCommandPhase')
    result['cursor_checks'] = []
    for mode in ['hd', 'original']:
        check('畫面模式', graphics=mode)
        for title, row_y, state in [('load',145,'StateLoadGame'), ('save',165,'StateSaveGame'), ('quit',205,'StateQuitConfirm')]:
            click(544,125)
            check('系統選單',state='StateSysMenu')
            click(560,row_y)
            check(title,state=state)
            move(300,240)
            snap(mode+'-'+title+'-a')
            move(400,240)
            snap(mode+'-'+title+'-b')
            with Image.open(out/(mode+'-'+title+'-a.png')) as left, Image.open(out/(mode+'-'+title+'-b.png')) as right:
                bounds = ImageChops.difference(left.convert('RGB'),right.convert('RGB')).getbbox()
            assert bounds and bounds[0]>=450 and bounds[1]>=360 and bounds[2]<=630 and bounds[3]<=390, (title,mode,bounds)
            result['cursor_checks'].append({'dialog':title,'graphics':mode,'difference_bounds':bounds})
            if title == 'quit':
                move(384,270)
                snap(mode+'-quit-button')
                click(384,270)
                check('可見中止返回',state='StateCommandPhase')
            else:
                click(400,240,button=3)
                check('右鍵返回戰略',state='StateCommandPhase')
        key('F7')
    # 正式保存落地與從同一讀檔選單讀回，額外劇本必須持續保留。
    click(544,125)
    click(560,165)
    check('保存',state='StateSaveGame')
    click(300,152)
    assert len(list((out/'save').glob('*.DT0'))) == 6
    click(400,240,button=3)
    check('保存返回',state='StateCommandPhase')
    for slot, row in [('saved',132),('S1',282),('S2',302),('S3',322)]:
        click(544,125)
        click(560,145)
        check('重新讀檔'+slot,state='StateLoadGame')
        snap('slots-'+slot)
        click(300,row)
        check('陣營選單'+slot,state='StateLoadFaction')
        move(318,208+20*a.faction)
        snap('faction-'+slot)
        click(318,208+20*a.faction)
        check('完成讀檔'+slot,state='StateCommandPhase')
        snap('loaded-'+slot)
    click(544,125)
    click(560,185)
    check('呈現設定',state='StateSettings', font_style='classic', hd_music_available=True)
    snap('settings-classic')
    for style in ['serif','sans','original','classic']:
        click(350,90)
        check('字型切換'+style,state='StateSettings',font_style=style)
        move(600,420)
        snap('settings-'+style)
    click(350,90)
    check('宋體',font_style='serif')
    key('F7')
    check('原圖保持原版字模',graphics='original',font_style='serif')
    snap('settings-original-mode')
    key('F7')
    check('返回HD保留宋體',graphics='hd',font_style='serif')
    if observe().get('music_quality') != 'hd':
        click(320,110)
    check('HD音色',music_quality='hd')
    snap('settings-music-hd')
    time.sleep(2)
    click(360,110)
    check('FM音色',music_quality='original')
    time.sleep(2)
    click(416,110)
    check('關閉音樂',music_track='')
    click(320,110)
    row = check('HD解除靜音',music_quality='hd')
    assert row.get('music_track'), row
    click(416,370)
    check('呈現設定返回',state='StateCommandPhase',font_style='serif',music_quality='hd')
    snap('strategy-serif')
    click(544,125)
    click(560,205)
    check('結束確認',state='StateQuitConfirm')
    move(432,270)
    snap('quit-decide')
    click(432,270)
    game.wait(timeout=10)
    assert game.returncode == 0, game.returncode
    result['quit_exit_code'] = game.returncode
    result['save_sha256'] = {p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in (out/'save').glob('*.DT0')}
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
