"""149正式X11玩家路徑。片頭、S1、編輯／取消／保存、重啟與另一陣營讀回。"""
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
p.add_argument('--out',required=True)
p.add_argument('--package',help='正式解壓包；省略時沿開發資料')
p.add_argument('--windows',action='store_true',help='在既有Wine容器內驗正式Windows包')
p.add_argument('--video',action='store_true',help='保存每次正常啟動的完整視窗無聲影片')
args=p.parse_args()
out=Path(args.out)
assert not out.exists() and out.parent.stat().st_uid==os.getuid()==1000
out.mkdir()
game=None
trace=out/'frames-0.jsonl'
events=[]
logs=[]
result=dict(passed=False,normal_entry=True,state_injection=False,script_flag=False,seed_injection=False)
launches=[]
videos=[]
recorder=None
env=dict(os.environ,GOMAXPROCS='2',LP_NUM_THREADS='1',GOMEMLIMIT='1600MiB',GOGC='60')

def x(*values):
    return subprocess.check_output(['xdotool',*map(str,values)],env=env,text=True,timeout=75).strip()

def observe():
    assert game.poll() is None,'遊戲提前結束'
    if not trace.exists():return {}
    with trace.open('rb') as f:
        end=f.seek(0,2)
        size=min(end,262144)
        f.seek(end-size)
        for row in reversed(f.read(size).splitlines()):
            try:return json.loads(row)
            except json.JSONDecodeError:pass
    return {}

def check(label,predicate,seconds=60):
    deadline=time.monotonic()+seconds
    while time.monotonic()<deadline:
        row=observe()
        if row and row['wall_ns']>time.time_ns()-2_000_000_000 and predicate(row):
            events.append(dict(label=label,observed=row))
            print('PASS',label,flush=True)
            return row
        time.sleep(.12)
    raise AssertionError((label,observe()))

def click(px,py,button=1):
    x('windowmove','--sync',window,0,0)
    x('mousemove','--window',window,round((px+.5)*1.5),round((py+.5)*1.5))
    x('mousedown',button)
    ack(button,True)
    x('mouseup',button)
    ack(button,False)

def ack(button=None,pressed=None):
    first=observe().get('ticks',-1);deadline=time.monotonic()+25
    while time.monotonic()<deadline:
        row=observe()
        if row.get('ticks',-1)>=first+2 and (button is None or row.get('mouse_buttons',[False,False])[0 if button==1 else 1]==pressed):return
        time.sleep(.06)
    raise AssertionError(('輸入未被正式視窗觀測',button,pressed,observe()))

def key(name):
    x('keydown',name);ack();x('keyup',name);ack()

def control(name):
    row=observe()
    controls=row.get('admiral_editor',{}).get('controls',row.get('settings',{}))
    r=controls[name];click((r[0]+r[2])//2,(r[1]+r[3])//2)

def snap(name):
    x('mousemove','--window',window,935,705);time.sleep(.25)
    file=out/(name+'.png')
    subprocess.run(['import','-window',window,str(file)],env=env,check=True,timeout=25)
    events.append(dict(capture=name,file=file.name,wall_ns=time.time_ns(),sha256=hashlib.sha256(file.read_bytes()).hexdigest(),observed=observe()))

def edit(field,value):
    control(f'field-{field}-value')
    check('數值輸入視窗',lambda r:r.get('admiral_editor',{}).get('editing')==field)
    for c in str(value):key(c)
    key('Return')
    return check('數值確認'+str(field),lambda r:r.get('admiral_editor',{}).get('editing')==-1)

def select_admiral(index,faction):
    control('faction-'+str(faction))
    check('選取陣營',lambda r:r['admiral_editor']['faction']==faction)
    for _ in range(15):
        if 'admiral-'+str(index) in observe()['admiral_editor']['controls']:break
        control('scroll-down')
    control('admiral-'+str(index))
    check('選取提督'+str(index),lambda r:r['admiral_editor']['selected']==index)

def start(n,faction):
    global game,window,trace,recorder
    trace=out/f'frames-{n}.jsonl'
    log=(out/f'game-{n}.log').open('w');logs.append(log)
    def path(p):return 'Z:'+str(p).replace('/','\\') if args.windows else str(p)
    package=Path(args.package) if args.package else None
    original=package/'original' if package else '/orig'
    data=package/'data' if package else '/src/docs/internals/data'
    art=package/'art' if package else '/src/assets/hd'
    command=(['wine',path(args.binary)] if args.windows else [args.binary])+['-orig',path(original),'-data',path(data),'-art',path(art),'-graphics','hd','-scale','1.5','-ai','1','-save',path(out/'save'),'-frame-trace',path(trace),'-music','']
    game=subprocess.Popen(command,env=env,cwd='/tmp',stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
    executed=Path(args.binary) if args.windows else Path(f'/proc/{game.pid}/exe')
    launches.append(dict(command=command,binary_sha256=hashlib.sha256(executed.read_bytes()).hexdigest(),identity_scope='正式PE檔案及Wine命令' if args.windows else '執行中ELF'))
    window=x('search','--sync','--onlyvisible','--name','^銀河英雄傳說III SP$').splitlines()[-1]
    x('windowfocus',window)
    check('正式視窗啟動',lambda r:r['state'] in ('StateIntro','StateTitle','StateMainMenu'))
    x('windowmove','--sync',window,0,0)
    if args.video:
        file=out/f'RAW-{n}.mp4'
        video_log=(out/f'record-{n}.log').open('w');logs.append(video_log)
        began=time.time_ns()
        recorder=subprocess.Popen(['ffmpeg','-nostdin','-v','warning','-f','x11grab','-draw_mouse','0','-framerate','60','-video_size','960x720','-i',os.environ['DISPLAY']+'.0+0,0','-t','600','-an','-c:v','libx264','-preset','ultrafast','-crf','18','-threads','2','-pix_fmt','yuv420p',str(file)],env=env,stdout=video_log,stderr=subprocess.STDOUT)
        videos.append(dict(launch=n,file=file.name,record_started_utc_ns=began))
        time.sleep(.5)
    for _ in range(3):
        if observe().get('state')=='StateMainMenu':break
        click(600,440)
    check('片頭至主選單',lambda r:r['state']=='StateMainMenu')
    click(300,270)
    check('防拷',lambda r:r['state']=='StateCopyProtection')
    for c in '138':key(c)
    click(432,270)
    check('存檔清單',lambda r:r['state']=='StateLoadGame')
    click(300,282)
    check('S1陣營',lambda r:r['state']=='StateLoadFaction')
    click(318,208+faction*20)
    check('S1正常戰略入口',lambda r:r['state']=='StateCommandPhase')
    click(544,125);click(560,185)
    check('設定頁',lambda r:r['state']=='StateSettings')
    snap(f'{n}-settings')
    control('admirals')
    check('編輯提督入口',lambda r:r.get('admiral_editor') is not None)
    return command

def stop():
    global game,recorder
    if recorder is not None and recorder.poll() is None:
        recorder.send_signal(signal.SIGINT)
        recorder.wait(timeout=20)
        assert recorder.returncode in (0,255)
    recorder=None
    if game is not None and game.poll() is None:
        os.killpg(game.pid,signal.SIGTERM)
        try:game.wait(timeout=10)
        except subprocess.TimeoutExpired:os.killpg(game.pid,signal.SIGKILL);game.wait(timeout=5)
    game=None

try:
    first=start(0,0)
    check('帝國基本頁',lambda r:r['admiral_editor']['selected']==0 and r['admiral_editor']['tab']==0)
    original=observe()['admiral_editor']['game_entry']['Abilities'][0]
    snap('0-basic')
    weapon_before=observe()['admiral_editor']['game_entry']['Abilities'][4]
    edit(4,1);control('apply')
    check('拒絕會造成零分母的武器設定',lambda r:'武器能力至少' in r['admiral_editor']['error'] and r['admiral_editor']['game_entry']['Abilities'][4]==weapon_before)
    control('cancel');control('confirm-yes');control('admirals')
    check('無效設定未保存或污染原局',lambda r:r['admiral_editor']['entry']['Abilities'][4]==weapon_before and not r['admiral_editor']['dirty'])
    edit(0,150)
    check('草稿不穿透原局',lambda r:r['admiral_editor']['entry']['Abilities'][0]==150 and r['admiral_editor']['game_entry']['Abilities'][0]==original)
    control('hidden');snap('0-hidden')
    edit(10,145)
    control('special')
    check('萊因哈特攻擊預設30%',lambda r:r['admiral_editor']['entry']['Special']['attack']==30)
    snap('0-special')
    key('Escape')
    check('取消確認',lambda r:r['admiral_editor']['confirm']=='cancel')
    control('confirm-yes')
    check('取消返回設定',lambda r:r.get('admiral_editor') is None)
    control('admirals')
    check('取消未改能力',lambda r:r['admiral_editor']['entry']['Abilities'][0]==original and not r['admiral_editor']['dirty'])
    edit(0,150);control('hidden');edit(10,145);control('special');edit(13,60);edit(14,55)
    control('apply')
    check('正式套用與偏好保存',lambda r:not r['admiral_editor']['dirty'] and r['admiral_editor']['game_entry']['Abilities'][0]==150 and r['admiral_editor']['game_entry']['Abilities'][10]==145 and r['admiral_editor']['notice']=='已保存全局設定')
    snap('0-applied')
    pref=json.loads((out/'save/preferences.json').read_text())['admiral_editor']
    assert pref['admirals']['0']['abilities']=={'0':150,'10':145}
    assert pref['admirals']['0']['special']==dict(transfer=60,movement=55,attack=30,fighter_attack=0)
    key('F7')
    check('原版顯示編輯頁',lambda r:r['graphics']=='original' and r.get('admiral_editor'))
    snap('0-original')
    key('F7')
    select_admiral(10,0);control('special')
    check('坎普戰鬥機預設30%',lambda r:r['admiral_editor']['entry']['Special']['fighter_attack']==30)
    snap('0-kemp-fighter')
    edit(16,65);control('apply')
    check('坎普戰鬥機保存65%',lambda r:r['admiral_editor']['game_entry']['Special']['fighter_attack']==65 and not r['admiral_editor']['dirty'])
    control('cancel')
    check('編輯頁返回設定',lambda r:r.get('admiral_editor') is None)
    control('close')
    check('設定返回原局',lambda r:r['state']=='StateCommandPhase')
    stop()
    second=start(1,1)
    control('special')
    check('同盟楊威利攻擊預設30%',lambda r:r['admiral_editor']['selected']==52 and r['admiral_editor']['entry']['Special']['attack']==30)
    snap('1-special')
    select_admiral(74,1)
    check('波布蘭戰鬥機預設30%',lambda r:r['admiral_editor']['entry']['Special']['fighter_attack']==30)
    snap('1-poplan-fighter')
    edit(16,75);control('apply')
    check('波布蘭戰鬥機保存75%',lambda r:r['admiral_editor']['game_entry']['Special']['fighter_attack']==75 and not r['admiral_editor']['dirty'])
    control('faction-0');control('basic')
    check('重啟及S1讀檔套用全局',lambda r:r['admiral_editor']['game_entry']['Abilities'][0]==150 and r['admiral_editor']['game_entry']['Abilities'][10]==145 and r['admiral_editor']['game_entry']['Special']==dict(transfer=60,movement=55,attack=30,fighter_attack=0))
    snap('1-global-reload')
    select_admiral(10,0);control('special')
    check('坎普戰鬥機跨重啟與讀檔保存',lambda r:r['admiral_editor']['game_entry']['Special']['fighter_attack']==65)
    final_hash=hashlib.sha256(Path(args.binary).read_bytes()).hexdigest()
    assert all(x['binary_sha256']==final_hash for x in launches),'驗證期間二進位改變'
    result.update(passed=True,commands=[first,second],binary_sha256=final_hash,launches=launches,preferences=pref)
finally:
    stop()
    result['events']=events
    result['launches']=launches
    result['videos']=videos
    result['platform']='Windows Wine' if args.windows else 'Linux X11'
    for video in videos:
        file=out/video['file']
        video['sha256']=hashlib.sha256(file.read_bytes()).hexdigest()
    (out/'RESULT.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    for log in logs:log.close()
