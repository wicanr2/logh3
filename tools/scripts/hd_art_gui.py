"""規格122：HD模式由片頭开始，X11真實輸入與全視窗核對。"""
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
p.add_argument('--art', required=True)
p.add_argument('--graphics', choices=['original', 'hd'], default='hd')
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
       '-graphics',a.graphics,'-art',a.art, '-save', str(out/'save'), '-scale', str(a.scale)]
proc = subprocess.Popen(cmd, stdout=log, stderr=subprocess.STDOUT)

def x(*args, timeout=15):
    return subprocess.check_output(['xdotool', *map(str,args)], text=True, timeout=timeout).strip()

def move(px, py):
    # 非整數縮放取邏輯像素中心，避免0.5四捨五入落到鄰格。
    x('mousemove', '--window', window, int((px+.5)*a.scale), int((py+.5)*a.scale))
    time.sleep(.18)

def click(px, py, button=1):
    move(px,py)
    # 同一個X11用戶端發送短按，避免兩次啟動xdotool的延遲跨過137ms連發閘門。
    press_start = time.monotonic()
    x('mousedown',button,'sleep','0.05','mouseup',button)
    events.append({'kind':'press-timing', 'x':px, 'y':py, 'button':button,
                   'requested_hold_seconds':.05, 'command_seconds':time.monotonic()-press_start})
    time.sleep(.22)

try:
    # 83張生成PNG的首次解碼已量到十餘秒，啟動與一般輸入採不同上限。
    window = x('search','--sync','--onlyvisible','--name','^銀河英雄傳說III SP$',timeout=60).splitlines()[-1]
    events.append({'kind':'window-ready','seconds':time.monotonic()-started})
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
            # 真實視窗更新與X11取圖非同步。只等待完整預期畫面，不遮罩游標或忽略差異。
            # 每次失敗的原圖、差圖與數字保留，最多三次；不重送玩家輸入。
            for attempt in range(3):
                result = subprocess.run([a.capture,'-window',window,'-base',str(base),'-scale',str(a.scale)], timeout=15)
                events.append({'kind':'capture-attempt','name':event['name'],'attempt':attempt+1,
                               'returncode':result.returncode,'seconds':time.monotonic()-started})
                if result.returncode == 0:
                    break
                if attempt == 2:
                    result.check_returncode()
                for suffix in ('-x11.png','-x11-diff.png','-x11-comparison.json'):
                    source = Path(str(base)+suffix)
                    if source.exists():
                        source.rename(str(base)+f'-attempt-{attempt+1}'+suffix)
                time.sleep(.5)
        else:
            raise AssertionError(event)
        events.append(dict(event, seconds=time.monotonic()-started))
    (out/'TRACE.json').write_text(json.dumps({'command':cmd,'normal_input':True,'script_flag':False,
         'scale':a.scale,'events':events}, ensure_ascii=False, indent=2)+'\n')
    print('正常X11逐項全圖通過', out, flush=True)
finally:
    # 失敗也保存按壓實際耗時，供辨別連發、更新追趕或捕捉時機。
    if not (out/'TRACE.json').exists():
        (out/'FAILURE-TRACE.json').write_text(json.dumps({'command':cmd,'normal_input':True,
             'scale':a.scale,'events':events}, ensure_ascii=False, indent=2)+'\n')
    if proc.poll() is None:
        proc.terminate()
        try:
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            proc.kill(); proc.wait(timeout=5)
    log.close()
