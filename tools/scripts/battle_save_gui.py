"""145：正式片頭、S3與公開戰術保存，僅送X11輸入。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--binary', required=True)
p.add_argument('--out', required=True)
a = p.parse_args()
out = Path(a.out)
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
out.mkdir()
trace = out/'FRAMES.jsonl'
env = dict(os.environ, GOMAXPROCS='2', LP_NUM_THREADS='2', GOMEMLIMIT='1200MiB')
# 本輪只驗保存與返回，音訊完成閘門沿用124，不另作音訊驗收聲明。
alsa = out/'alsa.conf'
alsa.write_text('pcm.!default { type null }\n')
env['ALSA_CONFIG_PATH'] = str(alsa)
cmd = [a.binary, '-orig', '/orig', '-data', '/src/docs/internals/data',
       '-art', '/src/assets/hd', '-graphics', 'hd', '-scale', '1.5',
       '-ai', '-1', '-save', str(out/'save'), '-frame-trace', str(trace),
       '-music', '/src/workplace/round132/music']
events = []
result = dict(command=cmd, events=events, normal_entry=True, state_injection=False,
              seed_injection=False, passive_override=False, script_flag=False, passed=False,
              binary_sha256=hashlib.sha256(Path(a.binary).read_bytes()).hexdigest())
log = (out/'game.log').open('x')
game = subprocess.Popen(cmd, cwd='/tmp', env=env, stdout=log, stderr=subprocess.STDOUT)

def x(*args):
    return subprocess.check_output(['xdotool', *map(str, args)], env=env, text=True, timeout=60).strip()

def observe():
    assert game.poll() is None, 'game exited'
    if not trace.exists():
        return {}
    with trace.open('rb') as f:
        size = f.seek(0, 2)
        f.seek(max(0, size-65536))
        for line in reversed(f.read().splitlines()):
            try:
                return json.loads(line)
            except json.JSONDecodeError:
                pass
    return {}

def wait(label, predicate, seconds=90):
    until = time.monotonic()+seconds
    while time.monotonic() < until:
        r = observe()
        if r and r['wall_ns'] > time.time_ns()-2_000_000_000 and predicate(r):
            events.append(dict(check=label, observed=r))
            print(label, r['state'], flush=True)
            return r
        time.sleep(.03)
    raise AssertionError((label, observe()))

def state(name):
    return wait(name, lambda r: r['state'] == name)

def click(px, py):
    x('mousemove', '--window', window, int((px+.5)*1.5), int((py+.5)*1.5))
    wait('cursor', lambda r: r.get('cursor') == [px, py], 10)
    x('mousedown', 1)
    try:
        wait('pressed', lambda r: r.get('mouse_buttons', [False])[0], 10)
    finally:
        x('mouseup', 1)
    wait('released', lambda r: not r.get('mouse_buttons', [True])[0], 10)
    events.append(dict(click=[px, py]))

def snap(name):
    # Frame telemetry is emitted before the window presents that frame.
    # Sample after presentation, keeping the product clock untouched.
    time.sleep(.4)
    path = out/(name+'.png')
    subprocess.run(['import', '-window', window, str(path)], env=env, check=True, timeout=15)
    events.append(dict(screenshot=path.name, sha256=hashlib.sha256(path.read_bytes()).hexdigest(), observed=observe()))

def tactical(turn):
    return wait('battle-turn-'+str(turn), lambda r: r['state'] == 'StateBattle'
                and r.get('tactical', {}).get('idle') and r['tactical']['turn'] == turn)

menu = json.loads(Path('/src/docs/internals/data/battle-menu.json').read_text())
l = menu['layout']
origin, win, highlight = l['origin'], l['window'], l['highlight']
cx = max(0, 240//origin['cell_w']+origin['cursor_cell_bias'])
cy = min(max(0, 470//origin['cell_h']+origin['cursor_cell_bias']), origin['clamp_rows']-len(menu['items']))
menu_x = cx*origin['cell_w']+origin['origin_bias']+highlight['x']-win['x']+8
menu_y = cy*origin['cell_h']+origin['origin_bias']+highlight['y']-win['y']+8

try:
    window = x('search', '--sync', '--onlyvisible', '--name', '^銀河英雄傳說III SP$').splitlines()[-1]
    x('windowfocus', window)
    x('windowmove', window, 0, 0)
    time.sleep(.6)
    click(600, 440)
    click(600, 440)
    state('StateMainMenu')
    click(300, 270)
    state('StateCopyProtection')
    for key in '138':
        x('keydown', key, 'sleep', '.12', 'keyup', key)
        time.sleep(.12)
    click(432, 270)
    state('StateLoadGame')
    click(300, 322)
    state('StateLoadFaction')
    click(318, 208)
    state('StateCommandPhase')
    click(544, 125)
    click(560, 125)
    until = time.monotonic()+90
    while not (observe().get('state') == 'StateBattle' and observe().get('tactical', {}).get('idle')):
        if observe().get('state') in ['StateTurnReport', 'StateOccupationReport', 'StateRegionFlipReport', 'StateRebellionReport']:
            click(240, 240)
        assert time.monotonic() < until, observe()
        time.sleep(.05)
    for slot, mode in enumerate(['hd', 'original']):
        before = tactical(slot+1)['tactical']
        snap(mode+'-before-save')
        click(240, 470)
        click(menu_x, menu_y+4*l['text']['row_pitch'])
        state('StateSaveGame')
        snap(mode+'-save-dialog')
        click(300, 152+20*slot)
        after = tactical(slot+1)['tactical']
        for key in ['region', 'turn', 'view', 'units', 'planets']:
            assert before[key] == after[key], (key, before[key], after[key])
        assert len(list((out/'save').glob('*.DT'+str(slot)))) == 6
        snap(mode+'-returned')
        click(240, 470)
        click(menu_x, menu_y)
        tactical(slot+2)
        snap(mode+'-next-turn')
        if slot == 0:
            x('keydown', 'F7')
            try:
                wait('original graphics', lambda r: r.get('graphics') == 'original', 10)
            finally:
                x('keyup', 'F7')
    result['save_sha256'] = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in (out/'save').glob('*.DT*')}
    result['passed'] = True
finally:
    if game.poll() is None:
        game.terminate()
        try:
            game.wait(timeout=8)
        except subprocess.TimeoutExpired:
            game.kill()
            game.wait(timeout=5)
    log.close()
    (out/'RESULT.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
