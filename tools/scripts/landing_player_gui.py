"""由正式片頭重播目前CPU政策的S3登陸占領，僅送X11公開輸入。"""
import argparse
import array
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time
import wave

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--plan', required=True)
p.add_argument('--binary', required=True)
p.add_argument('--out', required=True)
p.add_argument('--graphics', choices=['original', 'hd'], default='hd')
p.add_argument('--music', required=True)
a = p.parse_args()
plan = json.loads(Path(a.plan).read_text())
assert plan['passed'] and not any(plan[k] for k in ['state_injection', 'seed_injection', 'passive_override'])
out = Path(a.out)
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
out.mkdir()
trace = out/'FRAMES.jsonl'
env = dict(os.environ, GOMAXPROCS='2', LP_NUM_THREADS='2', GOMEMLIMIT='1200MiB', GOGC='50')
runtime = out/'pulse'
runtime.mkdir(mode=0o700)
env.update(XDG_RUNTIME_DIR=str(runtime), PULSE_SERVER='unix:'+str(runtime/'native'))
alsa = runtime/'alsa.conf'
alsa.write_text('</usr/share/alsa/alsa.conf>\npcm.!default { type pulse server "'+env['PULSE_SERVER']+'" }\n')
env['ALSA_CONFIG_PATH'] = str(alsa)
pulse_log = (out/'pulse.log').open('x')
pulse = subprocess.Popen(['pulseaudio', '--daemonize=no', '--use-pid-file=no', '--exit-idle-time=-1', '--disable-shm=yes', '--high-priority=no', '--realtime=no', '-n',
    '--load=module-native-protocol-unix socket='+str(runtime/'native')+' auth-anonymous=1', '--load=module-null-sink sink_name=logh3 rate=44100 channels=2'], env=env, stdout=pulse_log, stderr=subprocess.STDOUT)
until = time.monotonic()+10
while subprocess.run(['pactl', 'info'], env=env, capture_output=True, timeout=3).returncode:
    assert pulse.poll() is None and time.monotonic()<until
    time.sleep(.1)
recorder = subprocess.Popen(['parec', '--device=logh3.monitor', '--file-format=wav', '--format=s16le', '--rate=44100', '--channels=2', str(out/'audio.wav')], env=env)
cmd = [a.binary, '-orig', '/orig', '-data', '/src/docs/internals/data', '-art', '/src/assets/hd', '-graphics', a.graphics,
       '-scale', '1.5', '-ai', '-1', '-save', str(out/'save'), '-frame-trace', str(trace), '-music', a.music]
log = (out/'game.log').open('x')
game = subprocess.Popen(cmd, cwd='/tmp', env=env, stdout=log, stderr=subprocess.STDOUT)
events = []
result = {'normal_entry': True, 'state_injection': False, 'seed_injection': False, 'passive_override': False, 'script_flag': False,
          'command': cmd, 'binary_sha256': hashlib.sha256(Path(a.binary).read_bytes()).hexdigest(),
          'plan_sha256': hashlib.sha256(Path(a.plan).read_bytes()).hexdigest(), 'events': events, 'passed': False}
started = time.monotonic()
window = None
seen = set()
switched = False
scene_index = {}

def x(*args):
    return subprocess.check_output(['xdotool', *map(str, args)], env=env, text=True, timeout=60).strip()

def observe():
    assert game.poll() is None, 'game exited'
    if not trace.exists(): return {}
    with trace.open('rb') as f:
        size=f.seek(0, 2)
        f.seek(max(0, size-131072))
        lines=f.read().splitlines()
    for line in reversed(lines):
        try: return json.loads(line)
        except json.JSONDecodeError: pass
    return {}

def stamp(r):
    s = {'state': r.get('state')}
    t = r.get('tactical')
    if t:
        for k in ['region', 'turn', 'idle', 'view', 'fleet_order', 'command_open', 'menu_open', 'combat_input']:
            s[k] = t[k]
    return s

def match(r, wanted):
    s = stamp(r)
    return all(s.get(k)==v for k, v in wanted.items() if k!='view')

def move(px, py):
    x('mousemove', '--window', window, int((px+.5)*1.5), int((py+.5)*1.5))
    until=time.monotonic()+10
    while observe().get('cursor')!=[px,py]:
        assert time.monotonic()<until,('cursor not observed',px,py,observe().get('cursor'))
        time.sleep(.01)

def click(px, py, button=1):
    move(px, py)
    slot=0 if button==1 else 1
    x('mousedown',button)
    try:
        until=time.monotonic()+10
        while not observe().get('mouse_buttons',[False,False])[slot]:
            assert time.monotonic()<until,'mouse press was not observed'
            time.sleep(.005)
    finally:
        x('mouseup',button)
    until=time.monotonic()+10
    while observe().get('mouse_buttons',[True,True])[slot]:
        assert time.monotonic()<until,'mouse release was not observed'
        time.sleep(.005)

def key(k):
    old=observe().get('graphics')
    x('keydown',k)
    try:
        if k=='F7':
            until=time.monotonic()+10
            while observe().get('graphics')==old:
                assert time.monotonic()<until,'F7 not observed'
                time.sleep(.01)
        else: time.sleep(.12)
    finally: x('keyup',k)
    time.sleep(.1)

def snap(name):
    path = out/(name+'.png')
    assert not path.exists()
    subprocess.run(['import', '-window', window, str(path)], env=env, check=True, timeout=15)
    events.append({'screenshot': path.name, 'sha256': hashlib.sha256(path.read_bytes()).hexdigest(), 'observed': observe()})

def inspect(r):
    global switched
    s, pb = r.get('planet'), r.get('playback')
    if not s or not pb: return
    identity = pb['id']+'/'+str(s['Kind'])
    if identity not in scene_index: scene_index[identity] = len(scene_index)
    name = 'scene-%02d-kind%d-' % (scene_index[identity], s['Kind'])
    boundary = 'result' if pb['result'] else 'start'
    stable = boundary=='start' or pb['tick']>=pb['result_tick']+18
    label = name+boundary
    if stable and label not in seen:
        seen.add(label)
        snap(label)
    if a.graphics=='hd' and not switched and s['Kind']==0 and pb['applied']>=4 and not pb['result']:
        key('F7')
        snap('f7-during-landing')
        key('F7')
        switched = True

def wait(wanted, label, timeout=240):
    until = time.monotonic()+timeout
    while time.monotonic()<until:
        r=observe()
        if r:
            inspect(r)
            if r['wall_ns']>time.time_ns()-2_000_000_000 and match(r, wanted): return r
        time.sleep(.06)
    raise AssertionError((label, wanted, stamp(observe())))

def align_view(want, bounds=None):
    # Right-hold repeats are display paced. Use feedback and shorten the final
    # step instead of changing camera state or trusting a planned click count.
    for attempt in range(220):
        r=observe()
        v=r['tactical']['view']
        valid=bounds is None or all(bounds[n][0]<=v[n]<=bounds[n][1] for n in range(2))
        if valid and max(abs(v[0]-want[0]),abs(v[1]-want[1]))<=1: return
        distance=abs(v[0]-want[0])+abs(v[1]-want[1])
        if v[0]!=want[0]: point=(400 if v[0]<want[0] else 64, 220)
        else: point=(224, 400 if v[1]<want[1] else 80)
        click(*point, button=3)
        events.append({'camera_feedback': True, 'before_view': v, 'wanted_view': want, 'after': stamp(observe())})
    raise AssertionError(('camera alignment failed', want, stamp(observe())))

try:
    window=x('search', '--sync', '--onlyvisible', '--name', '^銀河英雄傳說III SP$').splitlines()[-1]
    x('windowfocus', window)
    x('windowmove', window, 0, 0)
    # bootstrap.Load's non-window plan starts at the main menu. The product
    # app attaches the intro, so dismiss it through the normal window first.
    time.sleep(.6)
    click(600,440)
    click(600,440)
    wait({'state':'StateMainMenu'}, 'normal intro dismissed')
    for i, action in enumerate(plan['actions']):
        if action['kind']=='right' and action['before'].get('idle'):
            # The plan records one Game scroll step per event. X11 repeats while
            # held, so the next public click aligns the camera by feedback and
            # maps the exact same world point through the actual view.
            events.append({'planned_pan_index':i,'wanted_view':action['after']['view']})
            continue
        if action['kind']=='ready':
            wait(action['after'], 'ready '+str(i))
            print('READY', i, action['after']['region'], action['after']['turn'], flush=True)
        elif action['kind']=='snapshot':
            wait(action['after'], 'snapshot '+str(i))
            time.sleep(.3)
            snap(action['name'])
        elif action['kind']=='key':
            key(action['key'])
        else:
            before=action['before']
            if before.get('idle') and 'world' in action:
                wait({k:v for k,v in before.items() if k!='view'}, 'before camera '+str(i))
                field=json.loads(Path('/src/docs/internals/data/screen-battle.json').read_text())['field']
                # Both the flagship and its planned destination must fit. A
                # one-cell camera tolerance alone can clip a destination at
                # the top edge; derive the legal range from all order clicks.
                points=[v['world'] for v in plan['actions'][i:i+4] if 'world' in v]
                sizes=[field['cols']*field['tile_w'],field['rows']*field['tile_h']]
                pitches=[field['tile_w'],field['tile_h']]
                bounds=[[(max(v[n] for v in points)-sizes[n]+pitches[n])//pitches[n],min(v[n] for v in points)//pitches[n]] for n in range(2)]
                align_view(before['view'],bounds)
            r=wait(before, 'before action '+str(i))
            point=action['point']
            if 'world' in action:
                field=json.loads(Path('/src/docs/internals/data/screen-battle.json').read_text())['field']
                v=r['tactical']['view']
                point=[field['x']+action['world'][0]-v[0]*field['tile_w'],field['y']+action['world'][1]-v[1]*field['tile_h']]
                assert field['x']<=point[0]<field['x']+field['cols']*field['tile_w'] and field['y']<=point[1]<field['y']+field['rows']*field['tile_h'],('world point outside view',i,point)
            elif 'command_row' in action:
                point=r['tactical']['command_point'].copy()
                point[1]+=20*action['command_row']
            elif 'weapon_row' in action:
                point=r['tactical']['combat_point'].copy()
                point[1]+=20*action['weapon_row']
            click(*point, button=3 if action['kind']=='right' else 1)
                # An automatic header can leave before the next trace. The next
                # action still waits for its complete public input boundary.
            events.append({'action_index': i, 'action': action, 'observed': observe(), 'seconds': time.monotonic()-started})
            print('ACTION',i,action['kind'],point,stamp(observe()),flush=True)
    r=wait(plan['actions'][-1]['after'], 'post capture')
    assert next(p for p in r['tactical']['planets'] if p['index']==63)['owner']==1
    result['capture_turn']=r['tactical']['turn']
    # Public tactical save, public load, public save to a different slot.
    click(240,470)
    menu=json.loads(Path('/src/docs/internals/data/battle-menu.json').read_text())
    l=menu['layout'];o=l['origin'];w=l['window'];h=l['highlight']
    cx=max(0,240//o['cell_w']+o['cursor_cell_bias']);cy=min(max(0,470//o['cell_h']+o['cursor_cell_bias']),o['clamp_rows']-len(menu['items']))
    click(cx*o['cell_w']+o['origin_bias']+h['x']-w['x']+8,cy*o['cell_h']+o['origin_bias']+h['y']-w['y']+4*l['text']['row_pitch']+8)
    wait({'state':'StateSaveGame'},'tactical save')
    click(300,152)
    until=time.monotonic()+8
    while len(list((out/'save').glob('*.DT0')))!=6:
        assert time.monotonic()<until
        time.sleep(.1)
    first={p.stem:p.read_bytes() for p in (out/'save').glob('*.DT0')}
    assert first['PLANET'][63*82+17]==1
    snap('public-saved')
    wait({'state':'StateBattle'},'save returned')
    click(240,470)
    click(cx*o['cell_w']+o['origin_bias']+h['x']-w['x']+8,cy*o['cell_h']+o['origin_bias']+h['y']-w['y']+3*l['text']['row_pitch']+8)
    wait({'state':'StateLoadGame'},'public load list')
    click(300,132)
    wait({'state':'StateLoadFaction'},'saved faction')
    click(318,228)
    wait({'state':'StateCommandPhase'},'public loaded')
    click(544,125);click(560,165)
    wait({'state':'StateSaveGame'},'public resave')
    click(300,172)
    until=time.monotonic()+8
    while len(list((out/'save').glob('*.DT1')))!=6:
        assert time.monotonic()<until
        time.sleep(.1)
    second={p.stem:p.read_bytes() for p in (out/'save').glob('*.DT1')}
    result['six_tables_equal']={name:first[name]==second[name] for name in first}
    # Public loading recomputes unrest and fleet counts. Compare complete files
    # with the independent public-input plan, rather than masking changed bytes.
    expected=Path(a.plan).parent/'save'
    for name in first:
        assert first[name]==(expected/(name+'.DT0')).read_bytes(),('saved table differs from public Game plan',name)
        assert second[name]==(expected/(name+'.DT1')).read_bytes(),('loaded table differs from public Game plan',name)
    result['six_tables_equal_public_plan']=True
    result['load_changes']={name:[{'offset':i,'before':x,'after':y} for i,(x,y) in enumerate(zip(first[name],second[name])) if x!=y] for name in first if first[name]!=second[name]}
    result['saved_sha256']={name:hashlib.sha256(data).hexdigest() for name,data in first.items()}
    result['loaded_sha256']={name:hashlib.sha256(data).hexdigest() for name,data in second.items()}
    click(400,240,3)
    wait({'state':'StateCommandPhase'},'resave returned')
    snap('public-loaded')
    rows=[json.loads(line) for line in trace.read_text().splitlines()]
    stages=[r for r in rows if r.get('planet')]
    assert {0,1,2}<={r['planet']['Kind'] for r in stages}
    assert any(r['planet']['Kind']==2 and r['planet']['Captured'] for r in stages)
    assert any((r.get('playback') or {}).get('sound_playing') for r in stages)
    assert any(any(c['played'] for c in r['playback'].get('sounds',[])) for r in stages)
    result['scene_count']=len(scene_index)
    result['f7_during_landing']=switched
    if a.graphics=='hd': assert switched
    result['passed']=True
finally:
    result['elapsed_seconds']=time.monotonic()-started
    if window and game.poll() is None:
        try: snap('final')
        except Exception: pass
    if game.poll() is None:
        game.terminate()
        try: game.wait(timeout=5)
        except subprocess.TimeoutExpired: game.kill();game.wait(timeout=5)
    recorder.terminate();recorder.wait(timeout=5)
    with wave.open(str(out/'audio.wav')) as f:
        result['audio_format']=[f.getnchannels(),f.getsampwidth(),f.getframerate()]
        pcm=f.readframes(f.getnframes())
    result['audio_peak']=max(map(abs,array.array('h',pcm)),default=0)
    result['audio_sha256']=hashlib.sha256((out/'audio.wav').read_bytes()).hexdigest()
    pulse.terminate();pulse.wait(timeout=5);pulse_log.close();log.close()
    (out/'RESULT.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    assert result['audio_format']==[2,2,44100] and result['audio_peak']>32
