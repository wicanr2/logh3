"""166：實際平台封包的名言、操作及存讀驗證。Docker內執行。"""
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
p.add_argument("--baseline", action="store_true")
p.add_argument('--appimage',action='store_true',help='實際AppImage自動素材與預設XDG存檔，不指定orig/data/art/save')
p.add_argument('--no-video',action='store_true',help='無FFmpeg平台只觀測相位；至少18秒，最長90秒')
p.add_argument("--baseline-receipt")
a = p.parse_args()
out = Path(a.out); assert not out.exists() and out.parent.stat().st_uid == os.getuid(); out.mkdir()
sha = lambda f: hashlib.sha256(Path(f).read_bytes()).hexdigest()
env = dict(os.environ, GOMAXPROCS='2', LP_NUM_THREADS='1', GOMEMLIMIT='1800MiB', GOGC='70')
save_dir=out/'config/logh3/save' if a.appimage else out/'save'
if a.appimage: env.update(APPIMAGE_EXTRACT_AND_RUN='1',XDG_CONFIG_HOME=str(out/'config'))
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
    w, h = geometry(); row=observe(); decor=row.get("poster_decoration") or {}
    if row["poster_mode"] and decor.get("approved"):
        vx,vy,vw,vh=decor["viewport"]
        px=(vx+(gx+.5)*vw/640)*w/2560
        py=(vy+(gy+.5)*vh/480)*h/1440
    else:
        scale=h/480;offset=(w-640*scale)/2 if row["poster_mode"] else 0
        px=offset+(gx+.5)*scale;py=(gy+.5)*scale
    x("mousemove","--window",window,int(px),int(py));ack()


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
    command = (['wine', path(a.binary)] if a.windows else [a.binary])+['-intro=false', '-graphics', 'original' if a.original_only else 'hd', '-scale', '1.5', '-music', '', '-frame-trace', path(trace)]
    if not a.appimage: command += ['-save',path(save_dir)]
    if package and not a.appimage: command += ['-orig', path(package/'original'), '-data', path(package/'data'), '-art', path(package/'art')]
    elif not a.windows and not a.appimage: command += ['-orig', '/orig', '-data', '/src/docs/internals/data', '-art', '/src/assets/hd']
    log = (out/(stem+'.log')).open('w'); logs.append(log)
    proc = subprocess.Popen(command, env=env, stdout=log, stderr=subprocess.STDOUT,start_new_session=a.appimage)
    state('StateMainMenu'); window = x('search', '--onlyvisible', '--name', '^銀河英雄傳說III SP$').splitlines()[-1]
    x('windowfocus', window); x('windowmove', '--sync', window, 0, 0)
    return command


def stop():
    global proc, recorder
    if recorder is not None:
        recorder.send_signal(signal.SIGINT); recorder.wait(timeout=20); assert recorder.returncode in (0, 255); recorder = None
    if proc is not None and proc.poll() is None:
        if a.appimage: os.killpg(proc.pid,signal.SIGTERM)
        else: proc.terminate()
        try: proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            if a.appimage: os.killpg(proc.pid,signal.SIGKILL)
            else: proc.kill()
            proc.wait(timeout=5)
    proc = None
    if a.windows: subprocess.run(['wineserver', '-k'], env=env, check=False, timeout=10)



# 154：正常鍵鼠、裝飾留白、循環與存讀；共用X11函式在凍結前置於本檔。

def capture(name,gx=600,gy=440):
    move(gx-1,gy-1);time.sleep(.1);move(gx,gy);time.sleep(.5)
    f=out/(name+'.png')
    subprocess.run(['import','-window',window,str(f)],env=env,check=True,timeout=20)
    row=observe();events.append(dict(name=name,screenshot=f.name,sha256=sha(f),observed=row,window=geometry()))
    print('PASS',name,flush=True)
    return row

def quote_group(group):
    return wait(lambda r:r['poster_decoration']['quote_group']==group,'名言第'+str(group)+'組',80)

def languages(group):
    settings()
    for code in ['zh-TW','zh-CN','ja','en','ko']:
        locale(code)
        row=capture(f'group-{group}-{code}')
        assert row['poster_decoration']['quote_group']==group, '語系操作跨越本組60秒，需另跑抽樣'
    locale('zh-TW');close()

result=dict(passed=False,normal_entry=True,state_injection=False,seed_injection=False,time_injection=False,
            script_flag=False,events=events,driver_sha256=sha(__file__))
try:
    command=start('first');capture('main')
    click(300,entry_y);state('StateCopyProtection')
    for k in ['1','3','8','Return']:key(k)
    click(300,240,3);click(300,240,3);state('StateCommandPhase')
    initial=observe()['strategic']
    settings();poster('v1');close()
    row=capture('group-1-map')
    assert row['poster_decoration']['quote_count']==5 and row['poster_decoration']['quote_group']==1
    before=observe();samples=[]
    until=time.monotonic()+90
    while observe()['poster_decoration']['quote_group']==1:
        r=observe();assert r['strategic']==initial
        samples.append(dict(ticks=r['ticks'],quote=r['poster_decoration'],background_uploads=r['background_uploads']))
        assert time.monotonic()<until,'實際時間未換名言'
        time.sleep(.2)
    after=quote_group(2);capture('group-2-map')
    assert before['background_uploads']==after['background_uploads'],'換名言上傳了遊戲背景'
    assert 60000<=after['poster_decoration']['quote_elapsed_ms']<65000
    result['idle_transition']=dict(before=before['background_uploads'],after=after['background_uploads'],samples=samples)
    if not a.probe:
        languages(2)
        settings();control('anim-off');close()
        assert not observe()['poster_decoration']['animated']
        quote_group(3);row=capture('group-3-animation-off')
        assert not row['poster_decoration']['animated'] and row['poster_decoration']['phase']==0
        languages(3)
        settings();control('anim-on');close()
        for group in [4,5]:
            quote_group(group);languages(group)
        quote_group(1);row=capture('group-1-wrapped')
        assert row['poster_decoration']['quote_elapsed_ms']>=300000
        languages(1)
    settings()
    for selected in ['v2','v3']:
        poster(selected);row=capture(selected+'-no-quotes')
        assert row['poster_decoration']['quote_group']==0
    poster('v1');close()
    key('F1');wait(lambda r:r['help_visible'],'help');capture('help');key('Escape')
    key('F7');wait(lambda r:r['graphics']=='original','original');capture('original');key('F7');wait(lambda r:r['graphics']=='hd','HD')
    assert observe()['strategic']==initial
    click(544,125);click(560,165);state('StateSaveGame');click(450,152)
    wait(lambda r:(save_dir/'SHIP.DT0').is_file(),'six saved tables')
    click(300,240,3);click(544,125);click(560,145);state('StateLoadGame');click(300,132);state('StateLoadFaction')
    layout=json.loads((menu_data/'battle-command-menus.json').read_text())['layout']
    click(289+layout['highlight_dx']+16,181+layout['highlight_dy']+a.faction*layout['row_pitch']+10)
    state('StateCommandPhase');capture('loaded')
    saved={f.name:sha(f) for f in sorted(save_dir.glob('*.DT0'))};assert len(saved)==6
    baseline=Path(a.baseline_receipt) if a.baseline_receipt else Path('/src/workplace/poster164')/('alliance' if a.faction else 'empire')/'RESULT.json'
    assert saved==json.loads(baseline.read_text())['saved'],'名言或語系改變六表'
    stop();start('restart');row=capture('restart-first-group')
    assert row['poster_decoration']['quote_group']==1 and row['poster_decoration']['quote_elapsed_ms']<15000
    result.update(passed=True,command=command,binary_sha256=sha(a.binary),saved=saved,full_cycle=not a.probe,appimage=a.appimage,automatic_resources=a.appimage)
    (out/'RESULT.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
except Exception as error:
    result['failure']=repr(error);(out/'FAILED.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');raise
finally:
    stop()
    for log in logs:log.close()
