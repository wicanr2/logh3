"""第132輪：正式片頭、兩音色選擇／取消／偏好重開及實際音訊。"""
import argparse
import array
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time
import wave

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--binary',required=True)
p.add_argument('--music',required=True)
p.add_argument('--out',required=True)
p.add_argument('--faction',type=int,choices=[0,1],required=True)
a=p.parse_args()
out=Path(a.out)
assert not out.exists() and out.parent.stat().st_uid==os.getuid()
out.mkdir()
env=dict(os.environ,GOMAXPROCS='2',LP_NUM_THREADS='2',GOMEMLIMIT='1200MiB',GOGC='50')
runtime=out/'pulse';runtime.mkdir(mode=0o700)
env.update(XDG_RUNTIME_DIR=str(runtime),PULSE_SERVER='unix:'+str(runtime/'native'))
alsa=runtime/'alsa.conf';alsa.write_text('</usr/share/alsa/alsa.conf>\npcm.!default { type pulse server "'+env['PULSE_SERVER']+'" }\n')
env['ALSA_CONFIG_PATH']=str(alsa)
events=[];children=[];handles=[];started=time.monotonic()
result={'normal_entry':True,'script_flag':False,'state_injection':False,'seed_injection':False,'faction':a.faction,'events':events,'passed':False,
        'binary_sha256':hashlib.sha256(Path(a.binary).read_bytes()).hexdigest()}
game=None;frames=None;window=None

def log(name):
    f=(out/name).open('x');handles.append(f);return f
def x(*args,timeout=20):
    return subprocess.check_output(['xdotool',*map(str,args)],env=env,text=True,timeout=timeout).strip()
def observe():
    assert game.poll() is None,'game exited early'
    if not frames.exists():return {}
    with frames.open('rb') as f:
        size=f.seek(0,2);f.seek(max(0,size-32768));lines=f.read().splitlines()
    for line in reversed(lines):
        try:return json.loads(line)
        except json.JSONDecodeError:pass
    return {}
def check(label,timeout=20,**want):
    until=time.monotonic()+timeout
    while time.monotonic()<until:
        r=observe()
        if r and r['wall_ns']>time.time_ns()-2_000_000_000 and all(r.get(k)==v for k,v in want.items()):
            events.append({'check':label,'expected':want,'observed':r,'seconds':time.monotonic()-started})
            print(label,want,flush=True);return r
        time.sleep(.15)
    raise AssertionError((label,want,observe()))
def move(px,py):
    x('mousemove','--window',window,int((px+.5)*1.5),int((py+.5)*1.5));time.sleep(.15)
def click(px,py,button=1):
    move(px,py);x('mousedown',button,'sleep','.05','mouseup',button)
    events.append({'click':[px,py,button],'seconds':time.monotonic()-started});time.sleep(.22)
def key(name):
    x('keydown',name,'sleep','.07','keyup',name);events.append({'key':name});time.sleep(.2)
def snap(name):
    move(600,420)
    path=out/(name+'.png')
    subprocess.run(['import','-window',window,str(path)],env=env,check=True,timeout=15)
    events.append({'screenshot':path.name,'sha256':hashlib.sha256(path.read_bytes()).hexdigest(),'observed':observe()})
def launch(index):
    global game,frames,window
    frames=out/f'FRAMES-{index}.jsonl'
    command=[a.binary,'-orig','/orig','-data','/src/docs/internals/data','-art','/src/assets/hd','-graphics','hd','-scale','1.5',
             '-music',a.music,'-music-quality','hd','-save',str(out/'save'),'-frame-trace',str(frames)]
    events.append({'launch':command})
    game=subprocess.Popen(command,cwd='/tmp',env=env,stdout=log(f'game-{index}.log'),stderr=subprocess.STDOUT);children.append(game)
    window=x('search','--sync','--onlyvisible','--name','^銀河英雄傳說III SP$',timeout=60).splitlines()[-1]
    x('windowfocus',window);x('windowmove',window,0,0);time.sleep(.6)
    click(600,440);click(600,440);check('main menu',state='StateMainMenu')
def enter():
    click(300,210+20*a.faction);check('copy protection',state='StateCopyProtection')
    for c in '138':key(c)
    click(432,270);check('accounting',state='StateMainGame')
    click(448,370);check('order',state='StateShipOrder')
    click(448,370);check('strategy',state='StateCommandPhase')
def settings():
    click(544,125);click(560,185);check('settings',state='StateSettings')
def quit_game():
    click(544,125);click(560,205);check('quit confirm',state='StateQuitConfirm')
    click(432,270);game.wait(timeout=10);assert game.returncode==0
def audio_sample(name):
    path=out/(name+'.wav')
    subprocess.run(['ffmpeg','-nostdin','-v','error','-f','pulse','-i','logh3.monitor','-t','2','-ar','44100','-ac','2','-c:a','pcm_s16le',str(path)],env=env,check=True,timeout=15)
    with wave.open(str(path),'rb') as f:
        assert f.getframerate()==44100 and f.getnchannels()==2 and f.getsampwidth()==2
        pcm=f.readframes(f.getnframes())
    samples=array.array('h',pcm);peak=max(map(abs,samples));rms=(sum(v*v for v in samples)/len(samples))**.5
    assert peak>32 and rms>4,(name,peak,rms)
    events.append({'audio':name,'peak':peak,'rms':rms,'pcm_sha256':hashlib.sha256(pcm).hexdigest(),'observed':observe()})

try:
    pulse=subprocess.Popen(['pulseaudio','--daemonize=no','--use-pid-file=no','--exit-idle-time=-1','--disable-shm=yes','--high-priority=no','--realtime=no','-n',
        '--load=module-native-protocol-unix socket='+str(runtime/'native')+' auth-anonymous=1','--load=module-null-sink sink_name=logh3 rate=44100 channels=2'],env=env,stdout=log('pulse.log'),stderr=subprocess.STDOUT);children.append(pulse)
    until=time.monotonic()+10
    while subprocess.run(['pactl','info'],env=env,capture_output=True,timeout=3).returncode:
        assert pulse.poll() is None and time.monotonic()<until;time.sleep(.1)
    launch(0);enter();settings()
    check('GS initial',soundfont='gs',mt32_available=True,music_quality='hd')
    snap('01-gs');audio_sample('gs')
    click(350,370);check('open choices',soundfont_menu=True);snap('02-choice')
    click(350,349);check('MT32 selected',soundfont='mt32',soundfont_menu=False,music_quality='hd')
    assert observe()['music_library'].endswith('/mt32')
    snap('03-mt32');audio_sample('mt32')
    click(350,370);click(300,240,3);check('right cancel',soundfont='mt32',soundfont_menu=False,state='StateSettings')
    click(350,370);key('Escape');check('escape cancel',soundfont='mt32',soundfont_menu=False,state='StateSettings')
    click(416,110);check('mute',music_track='')
    click(350,370);click(350,327);check('GS stays muted',soundfont='gs',music_track='')
    click(350,370);click(350,349);check('MT32 stays muted',soundfont='mt32',music_track='')
    click(320,110);r=check('HD unmute',soundfont='mt32',music_quality='hd');assert r['music_track']
    key('F7');check('original display keeps font choice',graphics='original',soundfont='mt32');snap('04-original')
    key('F7');check('HD returns',graphics='hd',soundfont='mt32')
    click(416,370);check('settings exit works',state='StateCommandPhase')
    pref=json.loads((out/'save/preferences.json').read_text());assert pref['soundfont']=='mt32'
    quit_game()
    launch(1);check('preference restored',soundfont='mt32',music_quality='hd');enter();settings();snap('05-restored')
    click(416,370);check('second settings exit',state='StateCommandPhase')
    result['soundfont_passed']=True
    quit_game()
    result['passed']=True
finally:
    result['elapsed_seconds']=time.monotonic()-started
    for child in reversed(children):
        if child.poll() is None:
            child.terminate()
            try:child.wait(timeout=5)
            except subprocess.TimeoutExpired:child.kill();child.wait(timeout=5)
    for f in handles:f.close()
    (out/'RESULT.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
