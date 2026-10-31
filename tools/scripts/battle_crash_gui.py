"""169：正式封包正常入口交火重播；支援海報實際viewport、音效及Windows Wine。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import time

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--binary',required=True);p.add_argument('--package',required=True)
p.add_argument('--plan',required=True);p.add_argument('--out',required=True)
p.add_argument('--poster',choices=['off','v1','v2','v3'],default='v1')
p.add_argument('--graphics',choices=['hd','original'],default='hd')
p.add_argument('--locale',default='zh-TW');p.add_argument('--ai',type=int,default=-1)
p.add_argument('--audio',choices=['on','off'],default='on');p.add_argument('--windows',action='store_true')
a=p.parse_args();out=Path(a.out);package=Path(a.package)
assert os.getuid()==1000 and out.parent.stat().st_uid==1000 and not out.exists()
assert package.is_dir();out.mkdir();trace=out/'FRAMES.jsonl';events=[];children=[];handles=[]
env=dict(os.environ,HOME='/tmp',GOMAXPROCS='2',LP_NUM_THREADS='2',GOMEMLIMIT='1800MiB',GOGC='70',APPIMAGE_EXTRACT_AND_RUN='1')
result=dict(passed=False,normal_entry=True,state_injection=False,seed_injection=False,time_injection=False,
    script_flag=False,poster=a.poster,graphics=a.graphics,locale=a.locale,audio=a.audio,events=events)
game=None;window=None
def sha(path):
    with Path(path).open('rb')as f:return hashlib.file_digest(f,'sha256').hexdigest()
def write(name,value):(out/name).write_text(json.dumps(value,ensure_ascii=False,indent=2)+'\n')
def log(name):
    f=(out/name).open('x');handles.append(f);return f
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
    w,h=geometry();row=observe();decor=row.get('poster_decoration')or{}
    if row['poster_mode']and decor.get('approved'):
        vx,vy,vw,vh=decor['viewport'];px=(vx+(gx+.5)*vw/640)*w/2560;py=(vy+(gy+.5)*vh/480)*h/1440
    else:
        scale=h/480;offset=(w-640*scale)/2 if row['poster_mode']else 0
        px=offset+(gx+.5)*scale;py=(gy+.5)*scale
    x('mousemove','--window',window,round(px),round(py));ack()
def click(gx,gy,button=1):
    move(gx,gy);x('mousedown',button);ack(button,True);x('mouseup',button);ack(button,False)
def capture(name):
    file=out/(name+'.png');subprocess.run(['import','-window',window,str(file)],env=env,check=True,timeout=20)
    events.append(dict(kind='capture',name=name,file=file.name,sha256=sha(file),observed=observe(),wall_ns=time.time_ns(),window=geometry()))
    print('CAPTURE',name,observe()['state'],flush=True)

try:
    if a.audio=='on':
        runtime=Path('/tmp/logh3-battle169-audio');runtime.mkdir(mode=0o700)
        env.update(XDG_RUNTIME_DIR=str(runtime),PULSE_SERVER='unix:'+str(runtime/'native'))
        config=runtime/'alsa.conf';config.write_text('</usr/share/alsa/alsa.conf>\npcm.!default { type pulse server "'+env['PULSE_SERVER']+'" }\n')
        env['ALSA_CONFIG_PATH']=str(config)
        pulse=subprocess.Popen(['pulseaudio','--daemonize=no','--use-pid-file=no','--exit-idle-time=-1','--disable-shm=yes',
            '--high-priority=no','--realtime=no','--log-level=warning','-n',
            '--load=module-native-protocol-unix socket='+str(runtime/'native')+' auth-anonymous=1',
            '--load=module-null-sink sink_name=logh3 rate=44100 channels=2 format=float32le norewinds=yes'],
            env=env,stdout=log('pulse.log'),stderr=subprocess.STDOUT);children.append(pulse)
        until=time.monotonic()+20
        while subprocess.run(['pactl','info'],env=env,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL).returncode:
            assert pulse.poll()is None and time.monotonic()<until;time.sleep(.1)
    command=(['wine',path(a.binary)]if a.windows else[a.binary])+[
        '-intro=false','-orig',path(package/'original'),'-data',path(package/'data'),'-art',path(package/'art'),
        '-save',path(out/'save'),'-frame-trace',path(trace),'-graphics',a.graphics,'-scale','1.5',
        '-poster='+str(a.poster!='off').lower(),'-poster-set',a.poster if a.poster!='off'else'v1',
        '-locale',a.locale,'-ai',str(a.ai),'-music',path(package/'music')if a.audio=='on'else'']
    result.update(command=command,binary_sha256=sha(a.binary),plan_sha256=sha(a.plan),driver_sha256=sha(__file__))
    game=subprocess.Popen(command,env=env,stdout=log('game.log'),stderr=subprocess.STDOUT,start_new_session=not a.windows)
    children.append(game);wait(lambda r:r['state']=='StateMainMenu')
    window=x('search','--onlyvisible','--name','^銀河英雄傳說III SP$').splitlines()[-1]
    x('windowfocus',window);x('windowmove','--sync',window,0,0);capture('main')
    until=time.monotonic()+900
    for event in json.loads(Path(a.plan).read_text()):
        assert time.monotonic()<until
        begin=time.time_ns();kind=event['kind']
        if kind=='stop':break
        if kind in['click','move']:
            if kind=='click':click(event['x'],event['y'],event.get('button',1))
            else:move(event['x'],event['y'])
        elif kind=='key':x('keydown',event['key']);ack();x('keyup',event['key']);ack()
        elif kind=='capture':capture(event['name'])
        elif kind=='wait_state':
            deadline=time.monotonic()+event.get('timeout',240)
            def value(row,key):
                for part in key.split('.'):row=row.get(part)if isinstance(row,dict)else None
                return row
            while True:
                row=observe()
                if row['state']==event['state']and(not event.get('animation')or row.get('animations'))and all(value(row,k)==v for k,v in event.get('expect',{}).items()):break
                combat=row.get('tactical')or{}
                if event.get('auto_combat')and row['state']=='StateBattle'and combat.get('combat_input'):
                    point=[240,470]if combat['combat_input']=='target'else combat['combat_point'];click(*point)
                    events.append(dict(kind='auto_combat',point=point,wall_ns=time.time_ns()))
                if event.get('auto_execute')and row.get('battle_ready')and combat.get('turn',0)<event.get('expect',{}).get('tactical.turn',0):click(600,242)
                assert time.monotonic()<deadline,('未到正常戰鬥狀態',event,row['state']);time.sleep(.04)
        elif kind!='wait':raise ValueError(event)
        time.sleep(event.get('wait',.35));events.append(dict(kind='input',source=event,wall_begin_ns=begin,wall_end_ns=time.time_ns()))
        write('EVENTS.json',events)
    time.sleep(3);result.update(passed=True,last=observe(),pre_cleanup_alive=game.poll()is None)
except BaseException as error:
    result.update(error=repr(error),pre_cleanup_returncode=game.poll()if game else None)
    print('FAIL',repr(error),flush=True)
    raise
finally:
    result['observed_until_utc_ns']=time.time_ns()
    for child in reversed(children):
        if child.poll()is None:
            child.terminate()
            try:child.wait(timeout=10)
            except subprocess.TimeoutExpired:child.kill();child.wait(timeout=5)
    for handle in handles:handle.close()
    if trace.exists():result['trace_sha256']=sha(trace)
    if(out/'game.log').exists():result['log_sha256']=sha(out/'game.log')
    write('RESULT.json',result)
    if result['passed']:print('PASS complete normal battle',flush=True)
