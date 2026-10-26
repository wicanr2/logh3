"""153：實際AppImage的開場、原版／強化參數、自動素材及XDG路徑診斷。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess

p=argparse.ArgumentParser(description=__doc__);p.add_argument('--appimage',required=True);p.add_argument('--out',required=True);a=p.parse_args()
out=Path(a.out);image=Path(a.appimage);assert not out.exists() and out.parent.stat().st_uid==os.getuid()==1000;out.mkdir()
sha=lambda f:hashlib.sha256(f.read_bytes()).hexdigest()
version='v.1.6.1-20261025'
env=dict(os.environ,APPIMAGE_EXTRACT_AND_RUN='1',GOMAXPROCS='2',LP_NUM_THREADS='1',GOMEMLIMIT='1800MiB',GOGC='70')
assert subprocess.check_output([str(image),'-version'],env=env,text=True,timeout=60).strip()==version
rows=[]
for name,level in [('direct-intro',None),('original',-1),('enhanced',1)]:
    runenv=dict(env,XDG_CONFIG_HOME=str(out/('config-'+name)))
    command=[str(image),'-shot',str(out/(name+'.png')),'-state-out',str(out/(name+'.txt')),
        '-frame-trace',str(out/(name+'.jsonl')),'-shot-after','4']
    if level is not None:
        command+=['-ai',str(level),'-intro=false','-script','click 300 230;type 138;enter;rclick 300 240;rclick 300 240']
        if level==1: command+=['-music-quality','hd']
    with (out/(name+'.log')).open('x') as log:
        proc=subprocess.Popen(command,cwd='/tmp',env=runenv,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
        try: code=proc.wait(timeout=180)
        except subprocess.TimeoutExpired:
            os.killpg(proc.pid,signal.SIGKILL);proc.wait();raise
    assert code==0,(out/(name+'.log')).read_text()
    frame=json.loads((out/(name+'.jsonl')).read_text().splitlines()[-1])
    expected='StateIntro' if level is None else 'StateCommandPhase'
    assert frame['state']==expected and frame['graphics']=='hd'
    assert frame['hd_music_available'] and frame['mt32_available'] and frame['poster_available']
    assert not frame['poster_mode'] and not frame['rules']['non_battleship_flagship'] and frame['locale']=='zh-TW'
    if level is not None: assert frame['ai_level']==level and frame['music_quality']==('hd' if level==1 else 'original')
    rows.append(dict(name=name,command=command,cwd='/tmp',exit=code,frame=frame,
        screenshot_sha256=sha(out/(name+'.png')),trace_sha256=sha(out/(name+'.jsonl'))))
    print('PASS AppImage',name,'自動素材／HD／音色／預設關閉',flush=True)
(out/'RESULT.json').write_text(json.dumps(dict(passed=True,version=version,appimage_sha256=sha(image),
    scope='實際AppImage的診斷啟動，extract-and-run，不取代正常GUI或人耳驗收',automatic_resources=True,
    xdg_config_isolated=True,runs=rows),ensure_ascii=False,indent=2)+'\n')
