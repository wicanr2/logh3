"""正常X11人事選單。單次短點擊或事件回授，完整視窗逐像素核對。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import time

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--binary', required=True)
p.add_argument('--plan', required=True)
p.add_argument('--out', required=True)
p.add_argument('--capture', required=True)
p.add_argument('--windows', action='store_true')
p.add_argument('--package', help='Windows既有完整包的唯讀素材來源')
p.add_argument('--ack', action='store_true', help='慢速虛擬顯示：按下後等到實際取樣再放開，不重送')
a = p.parse_args()
out, plan = Path(a.out), Path(a.plan)
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
out.mkdir()
trace = out/'frames.jsonl'
env = dict(os.environ, GOMAXPROCS='2', LP_NUM_THREADS='1')
def game_path(value):
    return 'Z:'+str(value).replace('/','\\') if a.windows else str(value)
package = Path(a.package) if a.package else None
command = (['wine', game_path(a.binary)] if a.windows else [a.binary])+[
    '-orig', game_path('/orig'),
    '-data', game_path(package/'data' if package else '/src/docs/internals/data'),
    '-art', game_path(package/'art' if package else '/src/assets/hd'),
    '-graphics', 'hd', '-intro=false', '-scale', '1.5', '-music', '',
    '-save', game_path(out/'save'), '-frame-trace', game_path(trace)]
events = []
log = (out/'game.log').open('w')
proc = subprocess.Popen(command, env=env, stdout=log, stderr=subprocess.STDOUT)

def x(*args):
    return subprocess.check_output(['xdotool', *map(str,args)], env=env,
                                   text=True, timeout=20).strip()

def move(px, py):
    x('mousemove','--window', window, int((px+.5)*1.5), int((py+.5)*1.5))
    time.sleep(.12)

def settled():
    # 等已送出的畫面完成呈現；不延長80ms按鍵，也不重送點擊。
    def last():
        rows = trace.read_text().splitlines()
        for row in reversed(rows):
            try: return json.loads(row)
            except json.JSONDecodeError: pass
        return {}
    first = last().get('ticks', -1)
    deadline = time.monotonic()+15
    while time.monotonic() < deadline:
        if last().get('ticks', -1) >= first+3: return
        time.sleep(.05)
    raise AssertionError('畫面未完成呈現')

def button_seen(button, pressed):
    deadline = time.monotonic()+15
    while time.monotonic() < deadline:
        for line in reversed(trace.read_text().splitlines()[-5:]):
            try: row = json.loads(line)
            except json.JSONDecodeError: continue
            if row['wall_ns'] >= button_started and row['mouse_buttons'][0 if button == 1 else 1] == pressed: return
            break
        time.sleep(.02)
    raise AssertionError(('未取樣滑鼠事件',button,pressed))

try:
    deadline = time.monotonic()+60
    while time.monotonic() < deadline:
        assert proc.poll() is None, '啟動失敗'
        if trace.exists() and 'StateMainMenu' in trace.read_text(): break
        time.sleep(.1)
    else: raise AssertionError('主選單未準備好')
    window = x('search','--onlyvisible','--name','^銀河英雄傳說III SP$').splitlines()[-1]
    x('windowfocus',window)
    x('windowmove','--sync',window,0,0)
    for event in json.loads((plan/'PLAN.json').read_text()):
        if event['Kind'] == 'click':
            move(event['X'],event['Y'])
            button_started = time.time_ns()
            x('mousedown',event['Button']); time.sleep(.08)
            if a.ack: button_seen(event['Button'],True)
            x('mouseup',event['Button']); time.sleep(.25)
            if a.ack: button_seen(event['Button'],False)
        elif event['Kind'] == 'key':
            x('keydown',event['Key']); time.sleep(.08)
            x('keyup',event['Key']); time.sleep(.15)
        else:
            move(600,440); settled(); time.sleep(.6)
            base = out/event['Name']
            shutil.copyfile(plan/(event['Name']+'.png'),str(base)+'.png')
            subprocess.run([a.capture,'-window',window,'-base',str(base),'-scale','1.5'],
                           env=env, check=True, timeout=20)
            print('PASS',event['Name'],flush=True)
        events.append(dict(event,wall_ns=time.time_ns()))
    (out/'RESULT.json').write_text(json.dumps(dict(passed=True, normal_entry=True,
        state_injection=False, script_flag=False, button_ms=80, acknowledged=a.ack, command=command,
        binary_sha256=hashlib.sha256(Path(a.binary).read_bytes()).hexdigest(),
        events=events),ensure_ascii=False,indent=2)+'\n')
finally:
    proc.terminate()
    try: proc.wait(timeout=5)
    except subprocess.TimeoutExpired: proc.kill(); proc.wait(timeout=5)
    if a.windows:
        subprocess.run(['wineserver','-k'],env=env,timeout=5,check=False)
    log.close()
