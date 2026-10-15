"""147：正常S1設定、敵軍實際可見、AI五值、四字型及F7。只在GUI／Wine Docker內執行。"""
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
p.add_argument('--package',required=True)
p.add_argument('--out',required=True)
p.add_argument('--faction',type=int,choices=[0,1],default=0)
p.add_argument('--windows',action='store_true')
p.add_argument('--video',action='store_true')
args=p.parse_args()
out,package=Path(args.out),Path(args.package)
assert os.getuid()==1000 and package.is_dir() and not out.exists()
assert out.parent.stat().st_uid==os.getuid()
out.mkdir()
trace=out/'FRAMES.jsonl'
env=dict(os.environ,GOMAXPROCS='2',LP_NUM_THREADS='1',GOMEMLIMIT='1200MiB',GOGC='50')
handles=[];children=[];events=[];game=None;recorder=None
result=dict(passed=False,normal_entry=True,script_flag=False,state_injection=False,seed_injection=False,
            platform='Windows Wine' if args.windows else 'Linux X11',faction=args.faction,
            binary_sha256=hashlib.sha256(Path(args.binary).read_bytes()).hexdigest())


def write(name,value):(out/name).write_text(json.dumps(value,ensure_ascii=False,indent=2)+'\n')
def path(value):return 'Z:'+str(value).replace('/','\\') if args.windows else str(value)
def log(name):
    f=(out/name).open('x');handles.append(f);return f
def x(*values,timeout=15):return subprocess.check_output(['xdotool',*map(str,values)],env=env,text=True,timeout=timeout).strip()
def observe():
    assert game.poll() is None,'遊戲提前結束'
    if not trace.exists():return {}
    with trace.open('rb') as f:
        end=f.seek(0,2);size=min(end,65536)
        while size:
            f.seek(end-size)
            for line in reversed(f.read(size).splitlines()):
                try:return json.loads(line)
                except json.JSONDecodeError:continue
            if size==end:break
            size=min(end,size*2)
    return {}
def check(label,**want):
    deadline=time.monotonic()+45
    while time.monotonic()<deadline:
        row=observe()
        if row and row['wall_ns']>time.time_ns()-2_000_000_000 and all(row.get(k)==v for k,v in want.items()):
            events.append(dict(check=label,observed=row));print('PASS',label,flush=True);return row
        time.sleep(.1)
    raise AssertionError((label,want,observe()))
def ack(button=None,pressed=None):
    first=observe().get('ticks',-1);deadline=time.monotonic()+15
    while time.monotonic()<deadline:
        row=observe()
        if row.get('ticks',-1)>=first+2 and (button is None or row.get('mouse_buttons',[False,False])[0 if button==1 else 1]==pressed):return
        time.sleep(.04)
    raise AssertionError(('未確認輸入',button,pressed,observe()))
def click(px,py,button=1):
    x('mousemove','--window',window,round((px+.5)*1.5),round((py+.5)*1.5));time.sleep(.08)
    x('mousedown',button);ack(button,True);x('mouseup',button);ack(button,False);time.sleep(.12)
    events.append(dict(kind='click',point=[px,py,button],wall_ns=time.time_ns()))
def key(name):
    x('keydown',name);ack();x('keyup',name);ack();time.sleep(.12)
    events.append(dict(kind='key',key=name,wall_ns=time.time_ns()))
def control(name):
    r=observe()['settings'][name]
    click((r[0]+r[2])//2,(r[1]+r[3])//2)
def snap(name,hold=.2):
    x('mousemove','--window',window,900,675);time.sleep(.25)
    begin=time.time_ns()
    file=out/(name+'.png')
    subprocess.run(['import','-window',window,str(file)],env=env,check=True,timeout=20)
    events.append(dict(kind='capture',name=name,wall_begin_ns=begin,wall_end_ns=time.time_ns(),
                       screenshot=file.name,sha256=hashlib.sha256(file.read_bytes()).hexdigest(),observed=observe()))
    time.sleep(hold)
def battle_settings():
    click(600,240);click(600,342);check('戰場設定',state='StateSettings')
def close_settings():
    control('close');check('返回戰場',state='StateBattle',battle_ready=True)


try:
    if not args.windows:
        runtime=out/'pulse';runtime.mkdir(mode=0o700)
        env.update(XDG_RUNTIME_DIR=str(runtime),PULSE_SERVER='unix:'+str(runtime/'native'))
        alsa=runtime/'alsa.conf';alsa.write_text('</usr/share/alsa/alsa.conf>\npcm.!default { type pulse server "'+env['PULSE_SERVER']+'" }\n');env['ALSA_CONFIG_PATH']=str(alsa)
        pulse=subprocess.Popen(['pulseaudio','--daemonize=no','--use-pid-file=no','--exit-idle-time=-1','--disable-shm=yes','--high-priority=no','--realtime=no','-n',
            '--load=module-native-protocol-unix socket='+str(runtime/'native')+' auth-anonymous=1','--load=module-null-sink sink_name=logh3 rate=44100 channels=2'],env=env,stdout=log('pulse.log'),stderr=subprocess.STDOUT)
        children.append(pulse);deadline=time.monotonic()+10
        while subprocess.run(['pactl','info'],env=env,capture_output=True,timeout=3).returncode:
            assert pulse.poll() is None and time.monotonic()<deadline;time.sleep(.1)
    command=(['wine',path(args.binary)] if args.windows else [args.binary])+[
        '-orig',path(package/'original'),'-data',path(package/'data'),'-art',path(package/'art'),'-graphics','hd',
        '-scale','1.5','-ai','1','-save',path(out/'save'),'-frame-trace',path(trace),'-music',path(package/'music') if not args.windows else '']
    game=subprocess.Popen(command,cwd='/tmp',env=env,stdout=log('game.log'),stderr=subprocess.STDOUT);children.append(game)
    window=x('search','--sync','--onlyvisible','--name','^銀河英雄傳說III SP$',timeout=75).splitlines()[-1]
    deadline=time.monotonic()+45
    while True:
        g=dict(line.split('=',1) for line in x('getwindowgeometry','--shell',window).splitlines())
        if (int(g['WIDTH']),int(g['HEIGHT']))==(960,720):break
        assert time.monotonic()<deadline;time.sleep(.1)
    x('windowfocus',window);x('windowmove','--sync',window,0,0)
    if args.video:
        assert not args.windows
        record_command=['ffmpeg','-nostdin','-v','warning','-thread_queue_size','8','-f','x11grab','-draw_mouse','0',
            '-framerate','60','-video_size','960x720','-i',os.environ['DISPLAY']+'.0+0,0','-thread_queue_size','64',
            '-f','pulse','-i','logh3.monitor','-t','300','-af','aresample=async=1:first_pts=0','-c:v','libx264',
            '-preset','ultrafast','-crf','18','-threads','2','-pix_fmt','yuv420p','-c:a','aac','-b:a','192k',
            '-movflags','+frag_keyframe+empty_moov+default_base_moof',str(out/'RAW.mp4')]
        record_start=time.time_ns();recorder=subprocess.Popen(record_command,env=env,stdout=log('record.log'),stderr=subprocess.STDOUT);children.append(recorder);time.sleep(.5)
    click(600,440);click(600,440);check('片頭至主選單',state='StateMainMenu')
    click(300,270);check('防拷',state='StateCopyProtection')
    for c in '138':key(c)
    click(432,270);check('取進度',state='StateLoadGame')
    click(300,282);check('S1陣營',state='StateLoadFaction')
    click(318,208+args.faction*20);check('S1戰略',state='StateCommandPhase')
    click(544,125);click(560,185)
    row=check('緊湊設定與敵軍預設關閉',state='StateSettings',show_enemies=False,ai_level=1)
    assert row['rules']==dict(flagship_transfer=True,hero_movement=True)
    snap('default-off')
    control('enemies');check('顯示敵軍開啟',show_enemies=True);snap('settings-on',3 if args.video else .2)
    control('flagship');control('movement');check('兩項規則關閉',rules=dict(flagship_transfer=False,hero_movement=False));snap('settings-off',2 if args.video else .2)
    control('flagship');control('movement')
    key('F7');check('原版緊湊設定',graphics='original');snap('settings-original',2 if args.video else .2)
    key('F7');check('HD設定',graphics='hd')
    control('ai');check('AI五項選單',ai_menu=True);snap('ai-menu',3 if args.video else .2)
    key('Escape');check('AI取消保持',ai_level=1,ai_menu=False)
    for level in [-1,0,1,2,3,1]:
        control('ai');control('ai-choice-'+str(level));check('AI切換'+str(level),ai_level=level,ai_menu=False)
    for style in ['serif','sans','original','classic']:
        control('font');check('字型'+style,font_style=style);snap('font-'+style)
    control('soundfont');check('音色選單',soundfont_menu=True);snap('soundfont-menu');key('Escape');check('音色取消',soundfont_menu=False)
    snap('settings-final-on')
    control('close');check('返回戰略',state='StateCommandPhase')
    click(544,125);click(560,125);click(240,240)
    row=check('正常戰場',state='StateBattle',battle_ready=True)
    field=json.loads((package/'data/screen-battle.json').read_text())['field']
    tactical=row['tactical'];targets=[u for u in tactical['units'] if u['Faction']!=args.faction and not u['Out']]
    assert targets
    target=min(targets,key=lambda u:u['X']**2+u['Y']**2)
    want=[max(0,min(field['map_cols']-field['cols'],(target['X']-field['cols']*field['tile_w']//2)//field['tile_w'])),
          max(0,min(field['map_rows']-field['rows'],(target['Y']-field['rows']*field['tile_h']//2)//field['tile_h']))]
    for _ in range(200):
        view=observe()['tactical']['view']
        if abs(view[0]-want[0])<=2 and abs(view[1]-want[1])<=2:break
        if view[0]!=want[0]:click(400 if view[0]<want[0] else 64,220,3)
        else:click(224,400 if view[1]<want[1] else 80,3)
    else:raise AssertionError(('敵軍鏡頭未到',want,observe()['tactical']['view']))
    before=observe()['tactical'];snap('battle-enemies-on')
    battle_settings()
    for level in [3,-1,1]:
        control('ai');control('ai-choice-'+str(level));check('戰場內AI切換'+str(level),ai_level=level,ai_menu=False)
    control('enemies');check('敵軍關閉',show_enemies=False);close_settings();snap('battle-enemies-off')
    battle_settings();control('enemies');check('敵軍重開',show_enemies=True);close_settings();snap('battle-enemies-on-again')
    after=observe()['tactical'];assert before['turn']==after['turn'] and before['units']==after['units'] and before['view']==after['view']
    px=field['x']+target['X']-before['view'][0]*field['tile_w'];py=field['y']+target['Y']-before['view'][1]*field['tile_h']
    box=tuple(round(v*1.5) for v in [px-8,py-8,px+24,py+24])
    geometry=f'{box[2]-box[0]}x{box[3]-box[1]}+{box[0]}+{box[1]}'
    def pixels(name):
        return subprocess.check_output(['convert',str(out/name),'-crop',geometry,'+repage','-depth','8','rgb:-'],timeout=15)
    on,off=pixels('battle-enemies-on-again.png'),pixels('battle-enemies-off.png')
    assert on!=off,'敵軍開關沒有改變目標艦隊圖示'
    result.update(passed=True,target=target,target_screen_box=box,turn_unchanged=True,units_unchanged=True,
                  show_enemies_default=False,show_enemies_functional=True,target_changed_channels=sum(a!=b for a,b in zip(on,off)),
                  ai_levels=[-1,0,1,2,3],font_styles=['serif','sans','original','classic'],command=command)
finally:
    if recorder is not None:
        if recorder.poll() is None:recorder.send_signal(signal.SIGINT)
        recorder.wait(timeout=30);assert recorder.returncode in (0,255)
    for child in reversed(children):
        if child.poll() is None:
            child.terminate()
            try:child.wait(timeout=5)
            except subprocess.TimeoutExpired:child.kill();child.wait(timeout=5)
    if args.video and recorder is not None and result['passed']:
        write('TRACE.json',dict(command=command,record_command=record_command,normal_x11_input=True,script_flag=False,state_injection=False,
            seed_injection=False,ai=1,record_started_utc_ns=record_start,binary_sha256=result['binary_sha256'],events=events,
            raw_sha256=hashlib.sha256((out/'RAW.mp4').read_bytes()).hexdigest(),frame_trace_sha256=hashlib.sha256(trace.read_bytes()).hexdigest()))
        write('FRAME-TRACE-CLOSE.json',dict(frame_trace_sha256=hashlib.sha256(trace.read_bytes()).hexdigest(),finalized_after_child_termination=True))
    result['events']=events;write('RESULT.json',result)
    for f in handles:f.close()
