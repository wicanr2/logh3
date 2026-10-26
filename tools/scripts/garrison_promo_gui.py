"""150正常視窗：駐留、改選提督、當地編成與出發。只送真實鍵鼠，不注入局面或seed。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import time

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--binary',required=True)
p.add_argument('--plan',required=True)
p.add_argument('--out',required=True)
p.add_argument('--windows',action='store_true')
p.add_argument('--package')
p.add_argument('--video',action='store_true')
p.add_argument('--poster',action='store_true',help='以16:9海報外框錄製；原版座標不變')
p.add_argument('--locale',choices=['zh-TW','zh-CN','ja','en','ko'])
a=p.parse_args()
out=Path(a.out); assert not out.exists() and out.parent.stat().st_uid==os.getuid()
out.mkdir()
plan=json.loads(Path(a.plan).read_text())
env=dict(os.environ,GOMAXPROCS='2',LP_NUM_THREADS='1',GOMEMLIMIT='1600MiB',GOGC='60')
trace=out/'FRAMES.jsonl'
events=[]; proc=None; recorder=None; logs=[]
sha=lambda f:hashlib.sha256(Path(f).read_bytes()).hexdigest()

def path(value):return 'Z:'+str(value).replace('/','\\') if a.windows else str(value)
package=Path(a.package) if a.package else None
base_command=(['wine',path(a.binary)] if a.windows else [a.binary])+[
    '-orig',path(package/'original' if package else '/orig'),'-data',path(package/'data' if package else '/src/docs/internals/data'),
    '-art',path(package/'art' if package else '/src/assets/hd'),'-graphics','hd','-intro=false',
    '-scale','1.5','-save',path(out/'save'),'-music','']
if a.poster:base_command+=['-poster=true','-poster-set','v1']
if a.locale:base_command+=['-locale',a.locale]

def observe():
    assert proc is not None and proc.poll() is None,'遊戲提早退出'
    if not trace.exists():return {}
    with trace.open('rb') as f:
        size=f.seek(0,2);f.seek(max(0,size-262144))
        for line in reversed(f.read().splitlines()):
            try:return json.loads(line)
            except json.JSONDecodeError:pass
    return {}

def wait(predicate,label,seconds=45):
    deadline=time.monotonic()+seconds
    while time.monotonic()<deadline:
        row=observe()
        if row and predicate(row):return row
        time.sleep(.03)
    raise AssertionError((label,observe()))

def x(*values):return subprocess.check_output(['xdotool',*map(str,values)],env=env,text=True,timeout=20).strip()

def ack(button=None,pressed=None):
    first=observe().get('ticks',-1)
    return wait(lambda r:r['ticks']>=first+2 and (button is None or r['mouse_buttons'][0 if button==1 else 1]==pressed),'正常輸入回授')

def click(px,py,button=1):
    values=dict(line.split('=',1) for line in x('getwindowgeometry','--shell',window).splitlines() if '=' in line)
    width,height=int(values['WIDTH']),int(values['HEIGHT']);scale=height/480
    offset=(width-640*scale)/2 if observe().get('poster_mode') else 0
    x('mousemove','--window',window,int(offset+(px+.5)*scale),int((py+.5)*scale));time.sleep(.07)
    began=time.time_ns();x('mousedown',button);ack(button,True)
    x('mouseup',button);ack(button,False)
    events.append(dict(kind='input',x=px,y=py,button=button,wall_begin_ns=began,wall_end_ns=time.time_ns()))

def key(value):
    name='Return' if value=='\r' else value
    x('keydown',name);ack();x('keyup',name);ack()

def matches(row,event):
    if row['state']!=event['state']:return False
    expected=event.get('strategic')
    if not expected:return True
    got=row.get('strategic') or {}
    for field in ['turn','faction','region','detail','message','stock','capital_stock','idle']:
        if field in expected and got.get(field)!=expected[field]:return False
    for field in ['fleets','editor']:
        want=expected.get(field,[]); actual=got.get(field,[])
        if len(want)!=len(actual):return False
        for left,right in zip(want,actual):
            if field=='fleets':
                for k in ['Index','Number','Admiral','Adjutant','Region','Target','Move']:
                    if left[k]!=right[k]:return False
                if len(left['Groups'])!=len(right['Groups']):return False
                pairs=zip(left['Groups'],right['Groups'])
            else:pairs=[(left,right)]
            for l,r in pairs:
                for k in ['kind','ships','slot2','slot3']:
                    if l[k]!=r[k]:return False
    return True

def snap(name):
    x('mousemove','--window',window,900,660);ack();time.sleep(.25)
    file=out/(name+'.png')
    subprocess.run(['import','-window',window,str(file)],env=env,check=True,timeout=20)
    began=time.time_ns();row=observe();time.sleep(1.2 if a.video else .2)
    events.append(dict(kind='capture',name=name,file=file.name,sha256=sha(file),wall_begin_ns=began,wall_end_ns=time.time_ns(),observed=row))
    print('PASS',name,flush=True)

def stop():
    global proc,recorder
    if recorder is not None:
        recorder.send_signal(signal.SIGINT);recorder.wait(timeout=20)
        assert recorder.returncode in (0,255)
        recorder=None
    if proc is not None and proc.poll() is None:
        proc.terminate()
        try:proc.wait(timeout=5)
        except subprocess.TimeoutExpired:proc.kill();proc.wait(timeout=5)
    proc=None
    if a.windows:subprocess.run(['wineserver','-k'],env=env,check=False,timeout=5)

result=dict(passed=False,normal_entry=True,state_injection=False,seed_injection=False,script_flag=False,events=events,driver_sha256=sha(__file__))
try:
    command=base_command+['-frame-trace',path(trace)]
    log=(out/'game.log').open('w');logs.append(log)
    proc=subprocess.Popen(command,env=env,stdout=log,stderr=subprocess.STDOUT)
    wait(lambda r:r['state']=='StateMainMenu','正常主選單',180)
    assert not observe()['rules']['non_battleship_flagship'],'新偏好預設未關閉'
    window=x('search','--onlyvisible','--name','^銀河英雄傳說III SP$').splitlines()[-1]
    x('windowfocus',window);x('windowmove','--sync',window,0,0)
    if a.video:
        log=(out/'record.log').open('w');logs.append(log)
        began=time.time_ns()
        recorder=subprocess.Popen(['ffmpeg','-nostdin','-v','warning','-f','x11grab','-draw_mouse','0','-framerate','60','-video_size','1280x720' if a.poster else '960x720','-i',os.environ['DISPLAY']+'.0+0,0','-t','600','-an','-c:v','libx264','-preset','ultrafast','-crf','18','-threads','2','-pix_fmt','yuv420p',str(out/'RAW.mp4')],env=env,stdout=log,stderr=subprocess.STDOUT)
        result['record_started_utc_ns']=began
        time.sleep(.4)
    for index,event in enumerate(plan['events']):
        kind=event['kind']
        if kind=='click':click(event['x'],event['y'])
        elif kind=='right':click(event['x'],event['y'],3)
        elif kind=='key':key(event['key'])
        elif kind=='rule':
            rect=observe()['settings'][event['id']]
            click((rect[0]+rect[2])//2,(rect[1]+rect[3])//2)
            wait(lambda r:r['rules']['non_battleship_flagship']==event['enabled'],'新開關接線')
        elif kind=='wait_command':
            deadline=time.monotonic()+150
            while time.monotonic()<deadline:
                row=observe()
                if matches(row,event):break
                if row['state'] in ['StateShipDelivery','StateMainGame','StateOccupationReport','StateRebellionReport','StateRegionFlipReport','StateAdmiralFateReport','StatePhezzanReport']:click(300,240,3)
                else:time.sleep(.05)
            else:raise AssertionError(('回合未返回',observe(),event))
        if event['state'] in ['StateTurnThinking','StateTurnWindow','StateTurnReport','StateEngagement']:
            pass # 自動階段由下一個wait_command驗證，不要求輪詢撞上短暫狀態。
        elif kind!='rule':wait(lambda r:matches(r,event),'步驟'+str(index))
        events.append(dict(kind='model-step',index=index,source=event,wall_ns=time.time_ns(),observed=observe()))
        if kind=='capture':snap(event['name'])
    # 正常公開保存及讀回，不使用原始資料注入。
    click(544,125);click(560,165);wait(lambda r:r['state']=='StateSaveGame','存進度')
    data_dir=package/'data' if package else Path('/src/docs/internals/data')
    slots=json.loads((data_dir/'save-slot-list.json').read_text())
    click(450,slots['save']['text_top']+10)
    wait(lambda r:(out/'save/SHIP.DT0').exists(),'六表保存')
    click(300,240,3);click(544,125);click(560,145)
    wait(lambda r:r['state']=='StateLoadGame','取進度')
    click(300,slots['top']+10);wait(lambda r:r['state']=='StateLoadFaction','陣營選擇')
    click(318,228)
    loaded=wait(lambda r:r['state']=='StateCommandPhase' and len((r.get('strategic') or {}).get('fleets',[]))==1,'巡航旗艦讀回')
    fleet=loaded['strategic']['fleets'][0]
    assert fleet['Region']==plan['destination'] and fleet['Number']==3 and fleet['Admiral']==plan['admirals'][1] and fleet['Groups'][0]['kind']==2
    snap('loaded-cruiser')
    pref=json.loads((out/'save/preferences.json').read_text())
    assert pref['non_battleship_flagship'] is True
    result.update(command=command,binary_sha256=sha(a.binary),plan_sha256=sha(a.plan),loaded=loaded,saved={p.name:sha(p) for p in sorted((out/'save').glob('*.DT0'))})
    (out/'CHECKPOINT.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    stop()
    # 同一偏好目錄重啟，確認預設不覆蓋使用者選擇。
    trace=out/'restart-frames.jsonl';log=(out/'restart.log').open('w');logs.append(log)
    proc=subprocess.Popen(base_command+['-frame-trace',path(trace)],env=env,stdout=log,stderr=subprocess.STDOUT)
    restarted=wait(lambda r:r['state']=='StateMainMenu','重啟主選單',180)
    assert restarted['rules']['non_battleship_flagship'] is True
    result.update(passed=True,restarted=restarted)
    if a.video:result['video_sha256']=sha(out/'RAW.mp4')
    (out/'RESULT.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
except Exception as error:
    result['failure']=repr(error)
    (out/'FAILED.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    raise
finally:
    stop()
    for log in logs:log.close()
