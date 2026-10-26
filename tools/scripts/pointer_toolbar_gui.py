"""152：正常視窗游標、hover、五語工具列及六表存讀。Docker內執行。"""
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


# 152：下方正常玩家路徑；X11共用函式由既有poster_sets_gui.py快照置於前段。
import statistics


def shot(name, gx=600, gy=440):
    move(gx-1, gy-1); time.sleep(.1); move(gx, gy); time.sleep(.4)
    f=out/(name+'.png')
    subprocess.run(['import','-window',window,str(f)],env=env,check=True,timeout=20)
    row=observe(); events.append(dict(name=name,screenshot=f.name,sha256=sha(f),observed=row,window=geometry(),wall_ns=time.time_ns()))
    print('PASS',name,flush=True)


def bar(): move(320,2); wait(lambda r:r['toolbar_visible'],'toolbar reveal')
def tool(name):
    b=observe()['toolbar_controls'][name]; click((b[0]+b[2])//2,(b[1]+b[3])//2)


result=dict(passed=False,normal_entry=True,state_injection=False,seed_injection=False,script_flag=False,events=events,
            driver_sha256=sha(__file__),baseline=a.baseline)
try:
    command=start('first'); shot('main')
    click(300,entry_y); state('StateCopyProtection')
    for k in ['1','3','8','Return']: key(k)
    click(300,240,3); click(300,240,3); state('StateCommandPhase')
    initial=observe()['strategic']; shot('map-off'); shot('button-hover',548,125)
    click(544,125); state('StateSysMenu'); shot('menu-off'); shot('menu-hover',548,165)
    click(300,240,3); state('StateCommandPhase'); move(600,440); time.sleep(.4)
    before=observe(); movements=[]
    for i in range(24):
        gx=500+(i%6)*18; gy=335+(i//6)*20
        w,h=geometry(); scale=h/480; offset=(w-640*scale)/2 if observe()['poster_mode'] else 0
        began=time.time_ns()
        x('mousemove','--window',window,int(offset+(gx+.5)*scale),int((gy+.5)*scale))
        row=wait(lambda r:r['cursor']==[gx,gy] and r['wall_ns']>=began,'pointer draw',30)
        movements.append(dict(target=[gx,gy],issued_ns=began,draw_start_ns=row['wall_ns'],latency_ms=(row['wall_ns']-began)/1e6,frame_ms=row['frame_ms'],draw_ms=row['draw_ms'],uploads=row.get('background_uploads')))
    after=observe()
    assert after['strategic']==initial
    if not a.baseline:
        assert before['background_uploads']==after['background_uploads'],(before['background_uploads'],after['background_uploads'])
        bar(); tool('font'); wait(lambda r:r['toolbar_menu']=='font','font menu'); shot('toolbar-font',350,80)
        tool('font-sans'); wait(lambda r:r['font_style']=='sans' and r['toolbar_menu']=='','Noto Sans')
        move(600,440); shot('map-sans')
        settings(); shot('settings-sans'); assert observe()['font_style']=='sans'
        bar(); tool('ai'); wait(lambda r:r['toolbar_menu']=='ai','AI menu'); shot('toolbar-ai',500,70)
        tool('ai-3'); wait(lambda r:r['ai_level']==3 and r['toolbar_menu']=='','AI strong3'); move(600,440); shot('settings-ai3')
        bar(); tool('font'); wait(lambda r:r['toolbar_menu']=='font','font cancel'); key('Escape'); wait(lambda r:r['toolbar_menu']=='','Esc cancels'); assert observe()['state']=='StateSettings'
        bar(); tool('ai'); wait(lambda r:r['toolbar_menu']=='ai','AI right cancel'); click(500,180,3); wait(lambda r:r['toolbar_menu']=='','right cancels'); assert observe()['state']=='StateSettings'
        for code in ['zh-TW','zh-CN','ja','en','ko']:
            locale(code)
            bar(); tool('font'); wait(lambda r:r['toolbar_menu']=='font','localized font'); shot('font-'+code,350,80); key('Escape')
            bar(); tool('ai'); wait(lambda r:r['toolbar_menu']=='ai','localized AI'); shot('ai-'+code,500,70); key('Escape')
        locale('zh-TW')
        move(600,440); close(); assert observe()['strategic']==initial
        key('F1'); wait(lambda r:r['help_visible'],'F1'); shot('help'); key('Escape'); wait(lambda r:not r['help_visible'],'help close')
        key('F7'); wait(lambda r:r['graphics']=='original','F7 original'); shot('original-pointer',580,420)
        shot('original-off'); shot('original-button-hover',548,125)
        click(544,125); state('StateSysMenu'); shot('original-menu-off'); shot('original-menu-hover',548,165)
        click(300,240,3); state('StateCommandPhase')
        key('F7'); wait(lambda r:r['graphics']=='hd','F7 HD'); shot('hd-pointer',580,420)
        settings(); poster('v1'); close(); shot('poster-map'); shot('poster-hover',548,125)
        bar(); tool('font'); wait(lambda r:r['toolbar_menu']=='font','poster font'); shot('poster-toolbar',350,80); key('Escape')
        # 側欄按鍵不能點穿，工具列取消後仍可操作正常系統指令。
        w,h=geometry(); raw_click(20,h//2); assert observe()['strategic']==initial
        move(600,440)
    click(544,125); click(560,165); state('StateSaveGame'); click(450,152)
    wait(lambda r:(save_dir/'SHIP.DT0').is_file(),'six tables')
    click(300,240,3); click(544,125); click(560,145); state('StateLoadGame'); click(300,132)
    state('StateLoadFaction')
    menu=json.loads((menu_data/'battle-command-menus.json').read_text()); layout=menu['layout']
    # loadfaction.go的正式原點(289,181)，列內位置沿共用CommandMenuSpec。
    click(289+layout['highlight_dx']+menu['menus']['load-faction']['width']*layout['cell_w']//2,
          181+layout['highlight_dy']+a.faction*layout['row_pitch']+layout['row_pitch']//2)
    state('StateCommandPhase'); shot('loaded')
    if not a.baseline: assert observe()['font_style']=='sans' and observe()['ai_level']==3
    result.update(passed=True,command=command,binary_sha256=sha(a.binary),movements=movements,
        performance=dict(samples=len(movements),latency_median_ms=statistics.median(v['latency_ms'] for v in movements),draw_median_ms=statistics.median(v['draw_ms'] for v in movements),frame_median_ms=statistics.median(v['frame_ms'] for v in movements),uploads_before=before.get('background_uploads'),uploads_after=after.get('background_uploads'),scope='Python啟動xdotool前至唯讀draw_start觀測，包含程序啟動與排程，非硬體端到端延遲'),
        saved={f.name:sha(f) for f in sorted(save_dir.glob('*.DT0'))},save_dir=str(save_dir),appimage=a.appimage,automatic_resources=a.appimage)
    assert len(result['saved'])==6
    (out/'RESULT.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(result['performance'],ensure_ascii=False),flush=True)
except Exception as error:
    result['failure']=repr(error); (out/'FAILED.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n'); raise
finally:
    stop()
    for log in logs: log.close()
