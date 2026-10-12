"""規格139：正常活動AI艦隊交火的X11錄影，沿既有回授式玩家重播器。"""
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import threading
import time

arguments = sys.argv[1:]
hd_playback = '--hd-playback' in arguments
if hd_playback:
    sys.argv.remove('--hd-playback')
    arguments = sys.argv[1:]
out = Path(arguments[arguments.index('--out')+1])
assert out.parent.stat().st_uid == os.getuid() and not out.exists()
runtime = Path('/tmp/battle129-pulse')
runtime.mkdir(mode=0o700)
os.environ.update(XDG_RUNTIME_DIR=str(runtime), PULSE_SERVER='unix:'+str(runtime/'native'),
                  GOMAXPROCS='2',LP_NUM_THREADS='2',GOMEMLIMIT='1200MiB',GOGC='50')
alsa = runtime/'alsa.conf'
alsa.write_text('</usr/share/alsa/alsa.conf>\npcm.!default { type pulse server "'+os.environ['PULSE_SERVER']+'" }\n')
os.environ['ALSA_CONFIG_PATH'] = str(alsa)
pulse = subprocess.Popen(['pulseaudio','--daemonize=no','--use-pid-file=no','--exit-idle-time=-1',
    '--disable-shm=yes','--high-priority=no','--realtime=no','-n',
    '--load=module-native-protocol-unix socket='+str(runtime/'native')+' auth-anonymous=1',
    '--load=module-null-sink sink_name=logh3 rate=44100 channels=2'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
recorder = None
record_log = None
switches = []
stop_switches = threading.Event()
switch_thread = None
record_started_ns = 0
command = []

def playback_switch():
    deadline = time.monotonic()+700
    try:
        while not stop_switches.wait(.03):
            path = out/'FRAMES.jsonl'
            if not path.exists():
                continue
            with path.open('rb') as stream:
                stream.seek(max(0,path.stat().st_size-65536))
                rows = stream.read().splitlines()
            row = json.loads(rows[-1])
            v = row.get('playback', {})
            if v.get('kind') == 'fleet' and v['tick'] >= 2*v['interval']:
                for pause in [1.8, 0]:
                    subprocess.run(['xdotool','keydown','F7','sleep','.07','keyup','F7'],check=True,timeout=10)
                    switches.append({'key':'F7','wall_ns':time.time_ns(),'scene':v['id']})
                    if pause:
                        time.sleep(pause)
                return
            assert time.monotonic() < deadline, '沒有到達正常艦隊交火'
    except BaseException as error:
        switches.append({'error':str(error)})

def start_recording(window, scale):
    global recorder, record_log, switch_thread, record_started_ns
    # HD assets are loaded at startup; the baseline comparisons use original RGB.
    if not hd_playback:
        subprocess.run(['xdotool','keydown','F7','sleep','.07','keyup','F7'],check=True,timeout=10)
        time.sleep(.3)
    record_log = (out/'record.log').open('x')
    record_started_ns = time.time_ns()
    recorder = subprocess.Popen(['ffmpeg','-nostdin','-y','-loglevel','warning',
        '-thread_queue_size','8','-f','x11grab','-draw_mouse','0','-framerate','60',
        '-video_size',f'{640*scale}x{480*scale}','-i',os.environ['DISPLAY']+'.0+0,0',
        '-thread_queue_size','64','-f','pulse','-i','logh3.monitor','-t','800',
        '-af','aresample=async=1:first_pts=0','-c:v','libx264','-preset','ultrafast',
        '-crf','18','-threads','2','-pix_fmt','yuv420p','-c:a','aac','-b:a','192k',
        '-movflags','+frag_keyframe+empty_moov+default_base_moof',str(out/'RAW.mp4')],
        stdout=record_log,stderr=subprocess.STDOUT)
    if not hd_playback:
        switch_thread = threading.Thread(target=playback_switch,daemon=True)
        switch_thread.start()

try:
    deadline = time.monotonic()+15
    while subprocess.run(['pactl','info'],capture_output=True,timeout=3).returncode:
        assert pulse.poll() is None and time.monotonic()<deadline
        time.sleep(.1)
    source = Path('/src/tools/scripts/fleet_player_gui_driver.py').read_text()
    source = source.replace("'/src/workplace/round79-capture'", "'/src/workplace/battle129/capture'")
    old = "'-save',str((out/'player-save').resolve()),'-scale',str(a.scale)]"
    new = "'-music','/src/dist-all/v.1.1.6-20261012/full/LOGH3-v.1.1.6-20261012-linux-amd64/music','-art','/src/assets/hd','-graphics','hd','-ai','-1','-frame-trace',str(out/'FRAMES.jsonl'),'-save',str((out/'player-save').resolve()),'-scale',str(a.scale)]"
    assert source.count(old)==1
    source = source.replace(old,new)
    old = "x('windowfocus',window);x('windowmove',window,0,0)"
    assert source.count(old)==1
    source = source.replace(old,old+'\n    start_recording(window,a.scale)')
    source = source.replace("after=request({'Seed':action['seed']})", "raise AssertionError('本輪禁止種子注入')")
    old = "send_input({'Type':4,'Button':3})\n        try:\n            down=wait(lambda s:s is not None and s['edges'][1]>=edge,'鏡頭按下',5)\n        finally:send_input({'Type':5,'Button':3})"
    assert source.count(old)==1
    source = source.replace(old, "send_input({'Type':4,'Button':3,'TapMs':20})\n        wait(lambda s:s is not None and s['edges'][1]>=edge,'鏡頭短按',5)")
    source = source.replace('escape=visits%2==1','escape=False')
    source = source.replace("    raise RuntimeError('真實右鍵鏡頭未收斂 '", "    if status()['view']==want:return status()\n    raise RuntimeError('真實右鍵鏡頭未收斂 '")
    if hd_playback:
        source = source.replace('def snapshot(name):', 'def original_snapshot(name):')
        point = '\ndef align_view(want):'
        assert source.count(point)==1
        source = source.replace(point, '''
def snapshot(name):
    assert status()['state']!='StateEngagement', '不得為診斷快照中斷交火'
    x('keydown','F7','sleep','.07','keyup','F7')
    original_snapshot(name)
    x('keydown','F7','sleep','.07','keyup','F7')
    switches.append({'key':'F7-pair','snapshot':name,'wall_ns':time.time_ns()})
\ndef align_view(want):''')
    source = source.replace('正常主選單、S3、真實X11鍵鼠、命令與交戰；RND test-only，保存由診斷API',
        '正常主選單、S3、活動原版AI、真實X11鍵鼠；無種子注入，保存由診斷API')
    namespace = {'__name__':'__main__','start_recording':start_recording,'switches':switches}
    exec(compile(source,'normal-fleet-gui-driver','exec'),namespace)
    command = namespace['proc'].args
finally:
    stop_switches.set()
    if switch_thread:
        switch_thread.join(timeout=5)
    if recorder and recorder.poll() is None:
        recorder.send_signal(signal.SIGINT)
        try:
            recorder.wait(timeout=30)
        except subprocess.TimeoutExpired:
            recorder.kill();recorder.wait(timeout=5)
    if record_log:
        record_log.close()
    pulse.terminate()
    try:
        pulse.wait(timeout=5)
    except subprocess.TimeoutExpired:
        pulse.kill();pulse.wait(timeout=5)
    if out.exists():
        (out/'SWITCHES.json').write_text(json.dumps(switches,indent=2)+'\n')

assert switches and all('error' not in s for s in switches), switches
assert hd_playback or len(switches)==2
assert recorder.returncode in (0,255)
probe = json.loads(subprocess.check_output(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(out/'RAW.mp4')]))
assert {s['codec_type'] for s in probe['streams']} >= {'video','audio'}
(out/'FFPROBE.json').write_text(json.dumps(probe,indent=2)+'\n')
digest = hashlib.sha256((out/'FRAMES.jsonl').read_bytes()).hexdigest()
(out/'TRACE.json').write_text(json.dumps({'normal_x11_input':True,'script_flag':False,
    'state_injection':False,'seed_injection':False,'ai':-1,'command':command,
    'readonly_observation_overlay':True,'binary_sha256':hashlib.sha256(Path(command[0]).read_bytes()).hexdigest(),
    'record_started_utc_ns':record_started_ns,'frame_trace_sha256':digest,
    'raw_sha256':hashlib.sha256((out/'RAW.mp4').read_bytes()).hexdigest()},indent=2)+'\n')
(out/'FRAME-TRACE-CLOSE.json').write_text(json.dumps({'frame_trace_sha256':digest,
    'finalized_after_child_termination':True},indent=2)+'\n')
print('PASS 正常艦隊交火、實際X11；' + ('交火全程HD，診斷快照只在交火外切換' if hd_playback else '交火中兩次F7顯示切換'))
