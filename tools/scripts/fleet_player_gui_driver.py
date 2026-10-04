"""第79輪：正式視窗的真實鍵鼠；只讀狀態等待，輸出拒絕覆寫。"""
import argparse
import json
import os
from pathlib import Path
import subprocess
import time

p=argparse.ArgumentParser()
p.add_argument('--plan',required=True)
p.add_argument('--out',required=True)
p.add_argument('--scale',type=int,choices=[1,2],required=True)
p.add_argument('--binary',default='/src/workplace/round79-gin3-v3')
p.add_argument('--input-binary',default='/src/workplace/round79-input-v2')
a=p.parse_args()
planpath=Path(a.plan)
plan=json.loads(planpath.read_text())
out=Path(a.out)
assert not out.exists() and out.parent.stat().st_uid==os.getuid()
out.mkdir()
statuspath=out/'status.json'
started=time.monotonic()
trace=[]
comparisons=[]
camera_gestures=[]
inputproc=None
env=dict(os.environ,LOGH3_FLEET_GUI_OUT=str(out.resolve()))
log=(out/'app.log').open('w')
proc=subprocess.Popen([a.binary,'-orig','/orig','-data','/src/docs/internals/data',
    '-save',str((out/'player-save').resolve()),'-scale',str(a.scale)],env=env,stdout=log,stderr=subprocess.STDOUT)

def x(*args):
    return subprocess.run(['xdotool',*map(str,args)],capture_output=True,text=True,check=True,timeout=30).stdout.strip()
def send_input(data):
    inputproc.stdin.write(json.dumps(data)+'\n');inputproc.stdin.flush()
    reply=inputproc.stdout.readline()
    assert reply and json.loads(reply)['queued'],reply
def input_trace(event,st):
    with (out/'input-trace.jsonl').open('a') as f:
        f.write(json.dumps({'event':event,'at_seconds':time.monotonic()-started,'state':st})+'\n')
def status():
    if proc.poll() is not None:raise RuntimeError('正式app提前退出 '+str(proc.returncode))
    try:return json.loads(statuspath.read_text())
    except FileNotFoundError:return None
def match(st,want):
    return st is not None and all(st.get(k)==v for k,v in want.items())
def wait(test,description,timeout=90):
    deadline=time.monotonic()+timeout
    last=None
    while time.monotonic()<deadline:
        last=status()
        if test(last):return last
        time.sleep(.002)
    raise RuntimeError(description+'；最後狀態 '+json.dumps(last,ensure_ascii=False))
def request(data):
    st=wait(lambda s:s is not None,'等候觀測器')
    q=out/'request.json';assert not q.exists()
    tmp=out/'request.tmp';tmp.write_text(json.dumps(data));tmp.rename(q)
    return wait(lambda s:s is not None and s['requests']==st['requests']+1,'快照／RND請求',10)
def observed_after(want,offset,index):
    # 自動階段可在按下與放開之間完成。每Update的原始紀錄才是完整觀測，
    # 不能要求低頻率的主機輪詢恰好看見短暫階段，也不能猜測它已發生。
    deadline=time.monotonic()+10
    with (out/'updates.jsonl').open() as updates:
        updates.seek(offset)
        while time.monotonic()<deadline:
            st=status()
            if match(st,want):return st
            while True:
                position=updates.tell()
                line=updates.readline()
                if not line or not line.endswith('\n'):
                    updates.seek(position)
                    break
                observed=json.loads(line)
                if match(observed,want):return observed
            time.sleep(.002)
    raise RuntimeError(f'輸入{index}結果未曾觀測到；最後狀態 '+json.dumps(status(),ensure_ascii=False))
def snapshot(name):
    st=request({'Name':name})
    wait(lambda s:s is not None and s['app_ticks']>=st['app_ticks']+3,'等實際Draw',10)
    # Xvfb無視窗管理器，視窗移到0,0；仍核實幾何。
    geometry=x('getwindowgeometry','--shell',window)
    values=dict(line.split('=',1) for line in geometry.splitlines() if '=' in line)
    assert int(values['WIDTH'])==640*a.scale and int(values['HEIGHT'])==480*a.scale,geometry
    subprocess.run(['/src/workplace/round79-capture','-window',window,'-base',str(out/name),'-scale',str(a.scale)],check=True)
    comparison=json.loads((out/(name+'-x11-comparison.json')).read_text())
    different=comparison['different']
    binpath=out/(name+'.bin')
    want=planpath.parent/(name+'.bin')
    assert binpath.read_bytes()==want.read_bytes(),name+'：正式GUI與公開輸入Game不同'
    comparisons.append({'name':name,'pixels':307200,'x11_different':different,'x11_bbox':comparison['bbox'],'game_index_equal':True})
    assert different==0,(name,different,comparison['bbox'])
def align_view(want):
    seen={}
    for attempt in range(40):
        st=status()
        if st['view']==want:return st
        assert st['state']=='StateBattle' and st['idle'] and st['stage']==-1 and st['order']==0,st
        statekey=tuple(st['view'])
        visits=seen.get(statekey,0);seen[statekey]=visits+1
        delta=[(w>v)-(w<v) for w,v in zip(want,st['view'])]
        # 二倍視窗可能一次Draw執行多個Update；近邊界時，往返會被
        # 同一邊界夾回相同位置。重訪位置時先走離目標，解除這個循環。
        escape=visits%2==1
        if escape:delta=[-v for v in delta]
        key=','.join(map(str,delta))
        px,py=plan['scroll_points'][key]
        x('mousemove','--window',window,px*a.scale,py*a.scale)
        wait(lambda s:s is not None and s['cursor']==[px,py],'鏡頭定位游標',10)
        edge=status()['edges'][1]+1
        start=time.monotonic()-started
        send_input({'Type':4,'Button':3})
        try:
            down=wait(lambda s:s is not None and s['edges'][1]>=edge,'鏡頭按下',5)
        finally:send_input({'Type':5,'Button':3})
        after=wait(lambda s:s is not None and not s['buttons'][1],'鏡頭放開',5)
        camera_gestures.append({'before_view':st['view'],'wanted_view':want,'x':px,'y':py,'cycle_escape':escape,
                                'at_seconds':start,'after':after})
    raise RuntimeError('真實右鍵鏡頭未收斂 '+str(want)+' '+str(status()))

try:
    inputproc=subprocess.Popen([a.input_binary],stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=log,text=True)
    # 每個輸入等正式app的狀態，不以固定sleep猜防拷或命令時機。
    window=x('search','--sync','--onlyvisible','--all','--pid',proc.pid,'--name','^銀河英雄傳說III SP$').splitlines()[0]
    x('windowfocus',window);x('windowmove',window,0,0)
    for i,action in enumerate(plan['actions']):
        before=wait(lambda s:match(s,action['before']),f'輸入{i}尚未就緒')
        update_offset=(out/'updates.jsonl').stat().st_size
        kind=action['kind']
        if kind=='seed':
            after=request({'Seed':action['seed']})
        else:
            if kind=='key':
                key='Return' if action['key']==13 else chr(action['key'])
                slot=2;down={'Type':2,'Key':action['key']};up={'Type':3,'Key':action['key']}
            else:
                x('mousemove','--window',window,action['x']*a.scale,action['y']*a.scale)
                wait(lambda s:s is not None and s['cursor']==[action['x'],action['y']],'邏輯游標',10)
                slot=0 if kind=='click' else 1
                button=1 if slot==0 else 3
                down={'Type':4,'Button':button};up={'Type':5,'Button':button}
            edge=before['edges'][slot]+1
            input_trace('before_press',status())
            send_input(down)
            try:
                pressed=wait(lambda s:s is not None and s['edges'][slot]>=edge,'正式按下事件',5)
                input_trace('pressed',pressed)
            finally:send_input(up)
            released=wait(lambda s:s is not None and not s['buttons'][slot],'正式放開事件',5)
            input_trace('released',released)
            if kind=='right' and action['before']['state']=='StateBattle' and action['before']['idle'] and action['before']['stage']==-1:
                remaining={k:v for k,v in action['after'].items() if k!='view'}
                assert match(released,remaining),('捲動改變其他控制狀態',released,remaining)
                align_view(action['after']['view'])
                x('mousemove','--window',window,action['x']*a.scale,action['y']*a.scale)
                wait(lambda s:s is not None and s['cursor']==[action['x'],action['y']],'捲動後游標',10)
            after=observed_after(action['after'],update_offset,i)
        trace.append({'index':i,'kind':kind,'x':action['x'],'y':action['y'],'key':action['key'],'seed':action['seed'],
                      'at_seconds':time.monotonic()-started,'before':before,'after':after})
        if 'snapshot' in action:snapshot(action['snapshot'])
    wait(lambda s:match(s,plan['final']),'正常交戰返回Input')
    snapshot('final')
    request({'Save':True})
    for p in sorted(planpath.parent.glob('*.DT0')):
        assert (out/p.name).read_bytes()==p.read_bytes(),'GUI診斷保存不同 '+p.name
    assert len(list(out.glob('*.DT0')))==6
    result={'classification':'正常主選單、S3、真實X11鍵鼠、命令與交戰；RND test-only，保存由診斷API','faction':plan['faction'],
            'scale':a.scale,'actions':trace,'comparisons':comparisons,'six_diagnostic_saves_equal':True,
            'camera_gestures':camera_gestures,'input_mapping':'右鍵保持以實際鏡頭回授定位；額外手勢完整記錄，非相同實體時序',
            'intro_default':True,'script_flag_used':False,'sound_device_tested':False}
    (out/'receipt.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    print('真實GUI',plan['faction'],a.scale,'輸入',len(trace),'快照',len(comparisons),'六份診斷保存相同')
except BaseException:
    (out/'failed-trace.json').write_text(json.dumps(trace,ensure_ascii=False,indent=2)+'\n')
    raise
finally:
    if inputproc is not None:
        inputproc.terminate()
        try:inputproc.wait(timeout=5)
        except subprocess.TimeoutExpired:inputproc.kill();inputproc.wait(timeout=5)
    proc.terminate()
    try:proc.wait(timeout=10)
    except subprocess.TimeoutExpired:proc.kill();proc.wait(timeout=5)
    log.close()
