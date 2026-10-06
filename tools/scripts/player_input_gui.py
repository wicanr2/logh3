"""規格121：正式遊戲由片頭開始，X11真實輸入與全圖核對；不注入Game。"""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import time

p = argparse.ArgumentParser()
p.add_argument('--binary', required=True)
p.add_argument('--assets', required=True)
p.add_argument('--out', required=True)
p.add_argument('--plan', required=True)
p.add_argument('--capture', required=True)
p.add_argument('--scale', type=float, default=1.5)
a = p.parse_args()
out, assets, plan = Path(a.out), Path(a.assets), Path(a.plan)
assert not out.exists() and assets.is_dir()
out.mkdir()
assert out.stat().st_uid == os.getuid()
events = []
started = time.monotonic()
log = (out/'game.log').open('x')
cmd = [a.binary, '-orig', str(assets/'original'), '-data', str(assets/'data'),
       '-music', str(assets/'music'), '-save', str(out/'save'), '-scale', str(a.scale)]
proc = subprocess.Popen(cmd, stdout=log, stderr=subprocess.STDOUT)

def x(*args):
    return subprocess.check_output(['xdotool', *map(str,args)], text=True, timeout=15).strip()

def move(px, py):
    # 非整數縮放取邏輯像素中心，避免0.5四捨五入落到鄰格。
    x('mousemove', '--window', window, int((px+.5)*a.scale), int((py+.5)*a.scale))
    time.sleep(.18)

def click(px, py, button=1):
    move(px,py)
    x('mousedown',button)
    time.sleep(.08)
    x('mouseup',button)
    time.sleep(.22)

try:
    window = x('search','--sync','--onlyvisible','--name','^銀河英雄傳說III SP$').splitlines()[-1]
    x('windowfocus',window)
    time.sleep(.6)
    click(600,440)
    click(600,440)
    events.append({'kind':'skip-intro','x':600,'y':440,'clicks':2})
    for event in json.loads((plan/'PLAN.json').read_text())['events']:
        kind = event['kind']
        if kind == 'click':
            click(event['x'], event['y'], event['button'])
        elif kind == 'move':
            move(event['x'], event['y'])
        elif kind == 'key':
            x('keydown',event['key']); time.sleep(.07)
            x('keyup',event['key']); time.sleep(.12)
        elif kind == 'capture':
            base = out/event['name']
            shutil.copyfile(plan/(event['name']+'.png'), str(base)+'.png')
            subprocess.run([a.capture,'-window',window,'-base',str(base),'-scale',str(a.scale)], check=True, timeout=15)
        else:
            raise AssertionError(event)
        events.append(dict(event, seconds=time.monotonic()-started))
    (out/'TRACE.json').write_text(json.dumps({'command':cmd,'normal_input':True,'script_flag':False,
         'scale':a.scale,'events':events}, ensure_ascii=False, indent=2)+'\n')
    print('正常X11逐項全圖通過', out, flush=True)
finally:
    if proc.poll() is None:
        proc.terminate()
        try:
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            proc.kill(); proc.wait(timeout=5)
    log.close()
