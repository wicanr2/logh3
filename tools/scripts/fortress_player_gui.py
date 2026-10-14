"""S1正常載入、撤退先前戰場、伊謝爾倫行星選單瞄準及回合開火。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--binary',required=True)
p.add_argument('--out',required=True)
a=p.parse_args()
out=Path(a.out)
assert not out.exists() and out.parent.stat().st_uid==os.getuid()
out.mkdir()
trace=out/'FRAMES.jsonl'
env=dict(os.environ,GOMAXPROCS='2',LP_NUM_THREADS='2',GOMEMLIMIT='1200MiB',GOGC='50')
cmd=[a.binary,'-orig','/orig','-data','/src/docs/internals/data','-art','/src/assets/hd',
     '-graphics','hd','-scale','1.5','-ai','-1','-save',str(out/'save'),'-frame-trace',str(trace)]
log=(out/'game.log').open('x')
game=subprocess.Popen(cmd,cwd='/tmp',env=env,stdout=log,stderr=subprocess.STDOUT)
events=[];started=time.monotonic();window=None
result={'normal_entry':True,'state_injection':False,'seed_injection':False,'script_flag':False,
        'ai':-1,'passive_override':False,'command':cmd,'binary_sha256':hashlib.sha256(Path(a.binary).read_bytes()).hexdigest(),
        'events':events,'passed':False}
menu=json.loads(Path('/src/docs/internals/data/battle-menu.json').read_text())
cm=json.loads(Path('/src/docs/internals/data/battle-command-menus.json').read_text())
field=json.loads(Path('/src/docs/internals/data/screen-battle.json').read_text())['field']

def x(*args,timeout=15):
    return subprocess.check_output(['xdotool',*map(str,args)],env=env,text=True,timeout=timeout).strip()
def observe():
    assert game.poll() is None,'game exited early'
    if not trace.exists():return {}
    with trace.open('rb') as f:
        size=f.seek(0,2);f.seek(max(0,size-65536));lines=f.read().splitlines()
    for line in reversed(lines):
        try:return json.loads(line)
        except json.JSONDecodeError:pass
    return {}
def move(px,py):
    x('mousemove','--window',window,int((px+.5)*1.5),int((py+.5)*1.5));time.sleep(.12)
def click(px,py,button=1):
    move(px,py);x('mousedown',button,'sleep','.04','mouseup',button);time.sleep(.2)
    events.append({'click':[px,py,button],'seconds':time.monotonic()-started,'observed':observe()})
def key(name):
    x('keydown',name,'sleep','.07','keyup',name);time.sleep(.2);events.append({'key':name})
def snap(name):
    path=out/(name+'.png');subprocess.run(['import','-window',window,str(path)],env=env,check=True,timeout=15)
    events.append({'screenshot':path.name,'sha256':hashlib.sha256(path.read_bytes()).hexdigest(),'observed':observe()})
def wait(label,predicate,timeout=40,handle=False):
    until=time.monotonic()+timeout
    while time.monotonic()<until:
        row=observe()
        if row and row['wall_ns']>time.time_ns()-2_000_000_000 and predicate(row):
            events.append({'check':label,'observed':row});print(label,flush=True);return row
        if handle and row:handle_wait(row)
        else:time.sleep(.1)
    raise AssertionError((label,observe()))
def handle_wait(r):
    tactical=r.get('tactical') or {}
    if r['state'] in ('StateTurnReport','StateOccupationReport','StateRebellionReport','StateRegionFlipReport','StateShipDelivery','StateAdmiralFateReport','StatePhezzanReport'):
        click(240,240)
    elif tactical.get('message_open'):
        snap('message-'+str(len(events)));raise AssertionError('unexpected modal; inspect actual button before continuing')
    elif tactical.get('combat_input'):
        if tactical['combat_input']=='target':click(240,470)
        else:click(*tactical['combat_point'])
    elif tactical.get('planet_menu') or tactical.get('menu_open'):
        # A report-dismiss click can reach the next screen before the latest
        # frame is observed. Close the visible menu before declaring idle.
        click(240,240,3)
    else:time.sleep(.08)
def system_execute():
    click(544,125);click(560,125)
def battle_menu(row):
    wait('idle before tactical menu',lambda r:(r.get('tactical') or {}).get('idle'))
    px,py=240,470
    click(px,py);wait('tactical menu opened',lambda r:(r.get('tactical') or {}).get('menu_open'),timeout=4)
    l=menu['layout'];o=l['origin'];w=l['window'];h=l['highlight']
    cx=max(0,px//o['cell_w']+o['cursor_cell_bias']);cy=max(0,py//o['cell_h']+o['cursor_cell_bias'])
    cy=min(cy,o['clamp_rows']-len(menu['items']))
    left,top=cx*o['cell_w']+o['origin_bias'],cy*o['cell_h']+o['origin_bias']
    click(left+h['x']-w['x']+8,top+h['y']-w['y']+row*l['text']['row_pitch']+8)
def find_planet(r,index):
    return next((p for p in (r.get('tactical') or {}).get('planets',[]) if p['index']==index),None)
def pan_to(p):
    want=[max(0,min(field['map_cols']-field['cols'],(p['x']-field['cols']*field['tile_w']//2)//field['tile_w'])),
          max(0,min(field['map_rows']-field['rows'],(p['y']-field['rows']*field['tile_h']//2)//field['tile_h']))]
    for _ in range(220):
        r=observe();v=r['tactical']['view']
        if abs(v[0]-want[0])<=2 and abs(v[1]-want[1])<=2:return
        if v[0]!=want[0]:click(400 if v[0]<want[0] else 64,220,3)
        else:click(224,400 if v[1]<want[1] else 80,3)
    raise AssertionError(('camera did not reach fortress',want,observe()))
def arm(p):
    pan_to(p);r=observe();v=r['tactical']['view']
    px=field['x']+p['x']-v[0]*field['tile_w'];py=field['y']+p['y']-v[1]*field['tile_h']
    chosen=None
    for dx,dy in ((0,0),(-6,-8),(6,-8),(6,6)):
        click(px+dx,py+dy)
        if observe()['tactical']['planet_menu']:chosen=(px+dx,py+dy);break
        click(240,240,3)
    assert chosen,'fortress cell remains covered by a unit'
    snap('fortress-menu-'+str(len(events)))
    l=cm['layout'];cx=max(0,chosen[0]//l['cell_w']+l['cursor_cell_bias']);cy=max(0,chosen[1]//l['cell_h']+l['cursor_cell_bias'])
    cy=min(cy,l['clamp_rows']-len(cm['menus']['fortress']['items']))
    click(cx*l['cell_w']+l['origin_bias']+l['highlight_dx']+8,
          cy*l['cell_h']+l['origin_bias']+l['highlight_dy']+3*l['row_pitch']+8)
    wait('normal fortress aiming',lambda r:r['tactical']['fortress_target'],timeout=4)
    # Aim toward the nearest fleet, including friendly fleets: the original
    # query deliberately includes both factions. This is an explicit test shot.
    fleets=observe()['tactical']['fleets']
    target=min(fleets,key=lambda f:(f['x']-p['x'])**2+(f['y']-p['y'])**2)
    dx,dy=target['x']-p['x']-7,target['y']-p['y']-7
    scale=90/max(1,abs(dx),abs(dy))
    tx,ty=int(px+7+dx*scale),int(py+7+dy*scale)
    move(tx,ty);snap('aim-ellipse-'+str(len(events)));click(tx,ty)
    r=wait('aim committed as state3',lambda r:find_planet(r,p['index']) and find_planet(r,p['index'])['gun_state']==3,timeout=4)
    events.append({'test_shot_target_fleet':target,'planet':p,'observed':r})

try:
    window=x('search','--sync','--onlyvisible','--name','^銀河英雄傳說III SP$',timeout=60).splitlines()[-1]
    x('windowfocus',window);x('windowmove',window,0,0);time.sleep(.6)
    click(600,440);click(600,440);wait('main menu',lambda r:r['state']=='StateMainMenu')
    click(300,270);wait('copy protection',lambda r:r['state']=='StateCopyProtection')
    for c in '138':key(c)
    click(432,270);wait('load list',lambda r:r['state']=='StateLoadGame')
    click(300,282);wait('S1 faction',lambda r:r['state']=='StateLoadFaction')
    click(318,228);wait('S1 alliance',lambda r:r['state']=='StateCommandPhase')
    click(544,125);click(560,185);wait('settings',lambda r:r['state']=='StateSettings')
    for _ in range(9):click(424,310)
    click(320,350);click(368,330);click(416,370)
    system_execute()
    fortress=None
    for attempt in range(6):
        r=wait('normal battle',lambda r:r['state']=='StateBattle' and (r.get('tactical') or {}).get('idle'),timeout=180,handle=True)
        print('battle',r['tactical']['region'],r['tactical']['turn'],r['tactical']['planets'],flush=True)
        fortress=next((p for p in r['tactical']['planets'] if p['fortress'] and p['owner']==1),None)
        if fortress:break
        previous_region=r['tactical']['region']
        battle_menu(2)
        r=wait('withdraw continued',lambda r:r['state']=='StateCommandPhase' or r['state']=='StateBattle' and r['tactical']['region']!=previous_region,timeout=180,handle=True)
        if r['state']=='StateCommandPhase':system_execute()
    assert fortress,'no owned real fortress battle reached'
    snap('fortress-entry')
    for turn in range(8):
        r=wait('fortress input',lambda r:r['state']=='StateBattle' and r['tactical']['idle'],timeout=180,handle=True)
        p=find_planet(r,fortress['index']);assert p and p['owner']==1
        previous=r['tactical']['turn']
        if p['gun_state'] in (0,3):arm(p)
        battle_menu(0)
        r=wait('fortress turn completed',lambda r:r['state']=='StateBattle' and r['tactical']['idle'] and r['tactical']['turn']>previous,timeout=180,handle=True)
        snap('after-turn-'+str(previous))
        hits=r['tactical'].get('gun_hits',[])
        if hits:
            assert sum(h['lost'] for h in hits)>0
            result['hits']=hits;result['planet']=find_planet(r,fortress['index']);result['region']=r['tactical']['region']
            result['passed']=True;break
    assert result['passed'],'normal fortress shot never hit a squadron'
finally:
    result['elapsed_seconds']=time.monotonic()-started
    if window and game.poll() is None:
        try:snap('final')
        except Exception:pass
    if game.poll() is None:
        game.terminate()
        try:game.wait(timeout=5)
        except subprocess.TimeoutExpired:game.kill();game.wait(timeout=5)
    log.close();(out/'RESULT.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
