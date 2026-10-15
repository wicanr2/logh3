"""正常片頭至系統設定、兩項規則開關、F7與偏好重開；Linux或Wine。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--binary',required=True)
p.add_argument('--out',required=True)
p.add_argument('--windows',action='store_true')
p.add_argument('--faction',type=int,choices=[0,1],default=0)
a=p.parse_args()
out=Path(a.out)
assert not out.exists() and out.parent.is_dir() and out.parent.stat().st_uid==os.getuid()
out.mkdir()
env=dict(os.environ,GOMAXPROCS='2',LP_NUM_THREADS='1',GOMEMLIMIT='1200MiB',GOGC='50')
events=[];handles=[];process=None;window=None;trace=None
result=dict(passed=False,normal_entry=True,script_flag=False,state_injection=False,seed_injection=False,
            platform='Windows Wine' if a.windows else 'Linux X11',faction=a.faction,
            binary_sha256=hashlib.sha256(Path(a.binary).read_bytes()).hexdigest(),events=events)


def path(value):
    return 'Z:'+str(value).replace('/','\\') if a.windows else str(value)


def x(*args):
    return subprocess.check_output(['xdotool',*map(str,args)],env=env,text=True,timeout=15).strip()


def observe():
    if process.poll() is not None:
        raise AssertionError('遊戲提前結束')
    if not trace.exists():return {}
    with trace.open('rb') as f:
        f.seek(max(0,f.seek(0,2)-65536));lines=f.read().splitlines()
    for line in reversed(lines):
        try:return json.loads(line)
        except json.JSONDecodeError:pass
    return {}


def check(label,**want):
    deadline=time.monotonic()+30
    while time.monotonic()<deadline:
        row=observe()
        if row and row['wall_ns']>time.time_ns()-2_000_000_000 and all(row.get(k)==v for k,v in want.items()):
            events.append(dict(check=label,expected=want,observed=row))
            print('PASS',label,flush=True)
            return row
        time.sleep(.12)
    raise AssertionError((label,want,observe()))


def click(xp,yp,button=1):
    x('mousemove','--window',window,int((xp+.5)*1.5),int((yp+.5)*1.5))
    time.sleep(.1)
    if a.windows:
        x('mousedown',button)
        acknowledge(button=button,pressed=True)
        x('mouseup',button)
        acknowledge(button=button,pressed=False)
    else:
        x('mousedown',button,'sleep','.07','mouseup',button)
    time.sleep(.2)
    events.append(dict(click=[xp,yp,button]))


def key(k):
    if a.windows:
        x('keydown',k);acknowledge()
        x('keyup',k);acknowledge()
    else:
        x('keydown',k,'sleep','.08','keyup',k)
    time.sleep(.2)
    events.append(dict(key=k))


def acknowledge(button=None,pressed=None):
    # Wine software rendering can update slower than a short synthetic press.
    # Observe actual updates instead of guessing a longer wall-clock delay.
    first=observe().get('ticks',-1)
    deadline=time.monotonic()+8
    while time.monotonic()<deadline:
        row=observe()
        if row.get('ticks',-1)>=first+2 and (button is None or row.get('mouse_buttons',[False,False])[0 if button==1 else 1]==pressed):
            events.append(dict(input_ack_ticks=row['ticks'],button=button,pressed=pressed))
            return
        time.sleep(.04)
    raise AssertionError(('輸入未由遊戲幀確認',button,pressed,observe()))


def snap(name):
    x('mousemove','--window',window,900,675)
    time.sleep(.15)
    f=out/(name+'.png')
    subprocess.run(['import','-window',window,str(f)],env=env,check=True,timeout=15)
    events.append(dict(screenshot=f.name,sha256=hashlib.sha256(f.read_bytes()).hexdigest(),observed=observe()))


def launch(index):
    global process,trace,window
    trace=out/f'FRAMES-{index}.jsonl'
    command=(['wine',path(a.binary)] if a.windows else [a.binary])+[
        '-scale','1.5','-ai','1','-save',path(out/'save'),'-frame-trace',path(trace)]
    if not a.windows:
        command+=['-orig','/orig','-data','/src/docs/internals/data','-graphics','hd','-art','/src/assets/hd']
    log=(out/f'game-{index}.log').open('x');handles.append(log)
    events.append(dict(command=command,cwd='/tmp'))
    process=subprocess.Popen(command,cwd='/tmp',env=env,stdout=log,stderr=subprocess.STDOUT)
    # Main window is created after the startup progress window completes.
    window=subprocess.check_output(['xdotool','search','--sync','--onlyvisible','--name','^銀河英雄傳說III SP$'],env=env,text=True,timeout=75).strip().splitlines()[-1]
    x('windowfocus',window);x('windowmove',window,0,0);time.sleep(.6)
    click(600,440);click(600,440);check('片頭至主選單',state='StateMainMenu')


def enter():
    click(300,210+20*a.faction);check('防拷',state='StateCopyProtection')
    for k in '138':key(k)
    click(432,270);check('月報',state='StateMainGame')
    click(448,370);check('訂造',state='StateShipOrder')
    click(448,370);check('戰略',state='StateCommandPhase')
    click(544,125);click(560,185);check('系統設定',state='StateSettings')


def quit_game():
    click(416,370);check('設定返回',state='StateCommandPhase')
    click(544,125);click(560,205);check('退出確認',state='StateQuitConfirm')
    click(432,270);process.wait(timeout=15)
    assert process.returncode==0


try:
    launch(0);enter()
    check('兩項預設開啟',rules=dict(flagship_transfer=True,hero_movement=True));snap('01-default-on')
    click(370,410);check('轉移關閉',rules=dict(flagship_transfer=False,hero_movement=True))
    click(370,434);check('兩項關閉',rules=dict(flagship_transfer=False,hero_movement=False));snap('02-off-hd')
    key('F7');check('原圖設定',graphics='original',rules=dict(flagship_transfer=False,hero_movement=False));snap('03-off-original')
    key('F7');check('HD返回',graphics='hd')
    pref=json.loads((out/'save/preferences.json').read_text())
    assert pref['flagship_transfer'] is False and pref['hero_movement'] is False
    quit_game()
    launch(1);enter();check('重開保持關閉',rules=dict(flagship_transfer=False,hero_movement=False));snap('04-restored')
    click(370,410);click(370,434);check('兩項重新開啟',rules=dict(flagship_transfer=True,hero_movement=True));snap('05-on-again')
    quit_game();result['passed']=True
finally:
    if process and process.poll() is None:
        process.terminate()
        try:process.wait(timeout=5)
        except subprocess.TimeoutExpired:process.kill();process.wait(timeout=5)
    for f in handles:f.close()
    (out/'RESULT.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
