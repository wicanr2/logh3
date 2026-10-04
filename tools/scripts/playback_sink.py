"""固定PulseAudio有時鐘虛擬輸出，錄正式ALSA後端；有界清理所有子程序。"""
import json
import os
from pathlib import Path
import subprocess
import sys
import time

os.chdir('/src')
out=Path('workplace')/sys.argv[1]
assert out.parent.stat().st_uid==os.getuid() and not out.exists()
out.mkdir()
runtime=Path('/tmp/logh3-playback-pulse')
assert not runtime.exists()
runtime.mkdir(mode=0o700)
env=dict(os.environ,HOME='/tmp',XDG_RUNTIME_DIR=str(runtime),PULSE_SERVER='unix:'+str(runtime/'native'),LOGH3_AUDIO_SINK_OUT=str(out.resolve()))
config=runtime/'alsa.conf'
config.write_text('</usr/share/alsa/alsa.conf>\npcm.!default { type pulse server "'+env['PULSE_SERVER']+'" }\n')
env['ALSA_CONFIG_PATH']=str(config)
pulse_log=(out/'pulse.log').open('w')
capture_log=(out/'capture.log').open('w')
raw=(out/'monitor.f32le').open('wb')
pulse=None
capture=None
child=None
started=time.monotonic()
def control(*args):
    return subprocess.run(['pactl',*args],env=env,capture_output=True,text=True,timeout=10,check=True).stdout
def terminate(proc):
    if proc is not None and proc.poll() is None:
        proc.terminate()
        try:proc.wait(timeout=5)
        except subprocess.TimeoutExpired:proc.kill();proc.wait(timeout=5)
try:
    pulse=subprocess.Popen(['pulseaudio','--daemonize=no','--use-pid-file=no','--exit-idle-time=-1','--disable-shm=yes',
        '--high-priority=no','--realtime=no','--log-level=warning','-n',
        '--load=module-native-protocol-unix socket='+str(runtime/'native')+' auth-anonymous=1',
        '--load=module-null-sink sink_name=logh3 rate=44100 channels=2 format=float32le norewinds=yes'],
        env=env,stdout=pulse_log,stderr=subprocess.STDOUT)
    deadline=time.monotonic()+15
    while time.monotonic()<deadline:
        assert pulse.poll() is None,'PulseAudio提早結束'
        try:
            info=control('info')
            break
        except subprocess.CalledProcessError:time.sleep(.05)
    else:raise RuntimeError('虛擬音訊服務未就緒')
    (out/'pulse-info.txt').write_text(info)
    (out/'sinks.json').write_text(control('-f','json','list','sinks'))
    capture=subprocess.Popen(['parec','--device=logh3.monitor','--format=float32le','--rate=44100','--channels=2','--latency-msec=10'],
        env=env,stdout=raw,stderr=capture_log)
    time.sleep(.2)
    assert capture.poll() is None,'monitor擷取未啟動'
    cmd=sys.argv[2:]
    assert cmd
    child=subprocess.Popen(cmd,env=env)
    code=child.wait(timeout=600)
    (out/'sink-inputs-end.json').write_text(control('-f','json','list','sink-inputs'))
    time.sleep(.15)
    assert pulse.poll() is None and capture.poll() is None
    result={'classification':'PulseAudio clocked virtual sink，正式ALSA輸出；非實體音效卡／人耳',
            'command':cmd,'exit_code':code,'server':env['PULSE_SERVER'],'alsa_config':config.read_text(),
            'sample_rate':44100,'channels':2,'format':'float32le','physical_dev_snd_present':Path('/dev/snd').exists(),
            'elapsed_seconds':time.monotonic()-started,'pulse_version':subprocess.run(['pulseaudio','--version'],capture_output=True,text=True,check=True).stdout.strip()}
    (out/'sink-receipt.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    if code:raise RuntimeError('正式程式結束碼 '+str(code))
finally:
    terminate(child)
    terminate(capture)
    terminate(pulse)
    raw.close();capture_log.close();pulse_log.close()
