"""172：正式執行檔正常入口新遊戲。未存檔、編隊、占領、實行到開戰，打完整場，存檔讀回再繼續。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--binary',required=True);p.add_argument('--package',required=True)
p.add_argument('--out',required=True);p.add_argument('--windows',action='store_true')
p.add_argument('--graphics',choices=['hd','original'],default='hd')
# 主選單列、占領目標與讀檔選陣營的落點；預設帝國往星域15，同盟用219／151,216／228往星域13。
p.add_argument('--menu-y',type=int,default=210);p.add_argument('--target',type=int,nargs=2,default=[176,167])
p.add_argument('--load-y',type=int,default=208);p.add_argument('--after-turns',type=int,default=3)
a=p.parse_args();out=Path(a.out);package=Path(a.package)
assert os.getuid()==1000 and out.parent.stat().st_uid==1000 and not out.exists()
assert package.is_dir();out.mkdir();trace=out/'FRAMES.jsonl';events=[];handles=[]
env=dict(os.environ,HOME='/tmp',GOMAXPROCS='2',LP_NUM_THREADS='2',GOMEMLIMIT='1800MiB',GOGC='70')
result=dict(passed=False,normal_entry=True,state_injection=False,seed_injection=False,time_injection=False,
    script_flag=False,graphics=a.graphics,audio='off',events=events)
REPORTS={'StateOccupationReport','StateRebellionReport','StateRegionFlipReport','StateTurnWindow','StateTurnReport',
    'StateShipDelivery','StateAdmiralFateReport'}
game=None;window=None
def sha(path):
    with Path(path).open('rb')as f:return hashlib.file_digest(f,'sha256').hexdigest()
def path(value):return 'Z:'+str(value).replace('/','\\')if a.windows else str(value)
def x(*args):return subprocess.check_output(['xdotool',*map(str,args)],env=env,text=True,timeout=20).strip()
def observe():
    assert game.poll()is None,('遊戲退出',game.poll())
    if not trace.exists():return {}
    with trace.open('rb')as f:
        end=f.seek(0,2);f.seek(max(0,end-1024*1024))
        for line in reversed(f.read().splitlines()):
            try:return json.loads(line)
            except json.JSONDecodeError:pass
    return {}
def wait(test,seconds=240):
    until=time.monotonic()+seconds
    while time.monotonic()<until:
        row=observe()
        if row and test(row):return row
        time.sleep(.04)
    raise AssertionError(('正常狀態逾時',observe().get('state')))
def ack(button=None,pressed=None):
    tick=observe().get('ticks',-1)
    return wait(lambda r:r['ticks']>=tick+2 and(button is None or r['mouse_buttons'][0 if button==1 else 1]==pressed),40)
def geometry():
    fields=dict(s.split('=',1)for s in x('getwindowgeometry','--shell',window).splitlines()if'='in s)
    return int(fields['WIDTH']),int(fields['HEIGHT'])
def move(gx,gy):
    w,h=geometry();scale=h/480
    x('mousemove','--window',window,round((gx+.5)*scale),round((gy+.5)*scale));ack()
def click(gx,gy,button=1):
    move(gx,gy);x('mousedown',button);ack(button,True);x('mouseup',button);ack(button,False)
    events.append(dict(kind='click',x=gx,y=gy,button=button,state=observe().get('state'),wall_ns=time.time_ns()));time.sleep(.2)
def key(name):
    x('keydown',name);ack();x('keyup',name);ack();events.append(dict(kind='key',key=name,wall_ns=time.time_ns()));time.sleep(.1)
def capture(name):
    file=out/(name+'.png');subprocess.run(['import','-window',window,str(file)],env=env,check=True,timeout=20)
    row=observe();events.append(dict(kind='capture',name=name,file=file.name,sha256=sha(file),state=row.get('state'),tactical=row.get('tactical'),wall_ns=time.time_ns()))
    print('CAPTURE',name,row.get('state'),flush=True)

try:
    command=(['wine',path(a.binary)]if a.windows else[a.binary])+[
        '-intro=false','-orig',path(package/'original'),'-data',path(package/'data'),'-art',path(package/'art'),
        '-save',path(out/'save'),'-frame-trace',path(trace),'-graphics',a.graphics,'-scale','1.5',
        '-poster=false','-locale','zh-TW','-ai','-1','-music','']
    result.update(command=command,binary_sha256=sha(a.binary),driver_sha256=sha(__file__))
    log=(out/'game.log').open('x');handles.append(log)
    game=subprocess.Popen(command,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=not a.windows)
    wait(lambda r:r['state']=='StateMainMenu')
    window=x('search','--onlyvisible','--name','^銀河英雄傳說III SP$').splitlines()[-1]
    x('windowfocus',window);x('windowmove','--sync',window,0,0)
    click(300,a.menu_y)
    for name in['1','3','8','Return']:key(name)
    for _ in range(8):
        if observe()['state']=='StateCommandPhase':break
        click(300,240,3);time.sleep(.3)
    assert observe()['state']=='StateCommandPhase',observe()['state']
    capture('strategy')
    for point in[(352,125),(400,125),(400,85),(488,330),(480,350),(510,350),(450,308),(490,345),(497,332),tuple(a.target),(432,263)]:
        click(*point)
    assert observe()['state']=='StateCommandPhase',observe()['state']
    capture('ordered')
    def execute_until(done,seconds,name):
        # 戰略回合：實行、點掉報告。done成立就停；碰到問話或沒見過的畫面照實留痕。
        until=time.monotonic()+seconds
        while True:
            assert time.monotonic()<until,(name,'逾時',observe().get('state'),seen[-12:])
            row=observe();state=row['state']
            if not seen or seen[-1]!=state:seen.append(state)
            if done(row):return row
            assert state!='StateOccupyAsk',('未預期的占領問話',name)
            if state=='StateCommandPhase':click(544,125);click(560,125)
            elif state in('StateMainGame','StateShipOrder'):click(300,240,3)
            elif state in REPORTS:click(560,125)
            else:time.sleep(.1)
    seen=[]
    first=execute_until(lambda r:r['state']=='StateBattle',600,'開戰')
    result['states']=seen
    wait(lambda r:r['state']=='StateBattle'and r.get('battle_ready'),240)
    capture('battle-entry');result['battle_region']=(observe().get('tactical')or{}).get('region')
    # 打完整場：有交火輸入就點，可以下令就按實行，直到離開戰場。
    until=time.monotonic()+1500;shots=set();finished=0
    while True:
        assert time.monotonic()<until,('戰場打不完',observe().get('state'),(observe().get('tactical')or{}).get('turn'))
        row=observe();state=row['state'];combat=row.get('tactical')or{}
        if state=='StateBattle':
            turn=combat.get('turn',0)
            if row.get('battle_ready')and turn>=2 and turn not in shots and turn in(2,3,9):
                shots.add(turn);capture('battle-turn-%02d'%turn)
            if combat.get('combat_input'):
                point=[240,470]if combat['combat_input']=='target'else combat['combat_point'];click(*point)
            elif row.get('battle_ready'):click(600,240);click(600,242)
            else:time.sleep(.1)
        elif state in REPORTS:click(560,125)
        elif state in('StateEngagement','StateTurnThinking'):time.sleep(.1)
        elif state in('StateCommandPhase','StateMainGame','StateShipOrder'):break
        else:
            assert state not in('StateOccupyAsk',),('未預期的畫面',state);time.sleep(.1)
    result['battle_left_state']=observe().get('state')
    execute_until(lambda r:r['state']=='StateCommandPhase',300,'戰後回指令畫面')
    capture('after-battle')
    # 正常存進度，再正常取進度。
    click(544,125);click(560,165);wait(lambda r:r['state']=='StateSaveGame',20);click(300,142);time.sleep(1)
    saved=sorted(p.name for p in(out/'save').iterdir());result['saved_files']=saved
    assert sum(n.split('.')[0]in('STELLAR','PLANET','ADMIRAL','FLEET','SHIP','WORK')for n in saved)>=6,saved
    result['saved_sha256']={n:sha(out/'save'/n)for n in saved}
    click(300,122,3);wait(lambda r:r['state']=='StateCommandPhase',20)
    click(544,125);click(560,145);wait(lambda r:r['state']=='StateLoadGame',20)
    click(300,122);wait(lambda r:r['state']=='StateLoadFaction',20)
    click(318,a.load_y);wait(lambda r:r['state']=='StateCommandPhase',60);capture('loaded')
    for n in range(a.after_turns):
        click(544,125);click(560,125);time.sleep(.3)
        execute_until(lambda r:r['state']in('StateCommandPhase','StateBattle'),300,'讀回後回合')
        if observe()['state']=='StateBattle':result['battle_after_load']=True;wait(lambda r:r.get('battle_ready'),240);break
    capture('after-load');time.sleep(2)
    result.update(passed=True,last_state=observe().get('state'),pre_cleanup_alive=game.poll()is None,all_states=seen)
except BaseException as error:
    result.update(error=repr(error),pre_cleanup_returncode=game.poll()if game else None)
    print('FAIL',repr(error),flush=True)
    raise
finally:
    result['observed_until_utc_ns']=time.time_ns()
    if game and game.poll()is None:
        game.terminate()
        try:game.wait(timeout=10)
        except subprocess.TimeoutExpired:game.kill();game.wait(timeout=5)
    for handle in handles:handle.close()
    if trace.exists():result['trace_sha256']=sha(trace)
    if(out/'game.log').exists():
        text=(out/'game.log').read_text(errors='replace');result['log_sha256']=sha(out/'game.log')
        result['panic']='panic:'in text
        if result['panic']:result['panic_head']=text[text.index('panic:'):][:1600]
    (out/'RESULT.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    if result['passed']:print('PASS new game first battle',flush=True)
