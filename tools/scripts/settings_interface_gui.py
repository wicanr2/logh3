"""容器內正常設定入口，真實X11鍵鼠及全圖核對，無script或注入。"""
import argparse
import json
from pathlib import Path
import shutil
import subprocess
import time

p = argparse.ArgumentParser()
p.add_argument('--out', required=True)
p.add_argument('--plan', required=True)
a = p.parse_args()
out, plan = Path(a.out), Path(a.plan)
out.mkdir()
events = []
begin = time.monotonic()


def run(*args):
    return subprocess.run(args, check=True, capture_output=True, text=True, timeout=10).stdout


def note(kind, **extra):
    events.append(dict(kind=kind, seconds=time.monotonic()-begin, **extra))
    print(json.dumps(events[-1], ensure_ascii=False), flush=True)


log = (out/'game.log').open('x')
cmd = ['/src/workplace/settings66-r98-gin3', '-orig', '/orig',
       '-data', '/src/docs/internals/data', '-save', str(out/'save'),
       '-shot', str(out/'final.png'), '-shot-bin', str(out/'final.bin'),
       '-state-out', str(out/'final.txt'), '-shot-after', '900']
proc = subprocess.Popen(cmd, stdout=log, stderr=log)
try:
    window = run('xdotool', 'search', '--sync', '--onlyvisible', '--name',
                 '^銀河英雄傳說III SP$').splitlines()[0]
    run('xdotool', 'windowfocus', window)
    note('normal-start', command=cmd, tps=60, injection=False)

    def click(x, y, button=1):
        run('xdotool', 'mousemove', '--window', window, str(x), str(y))
        time.sleep(.1)
        run('xdotool', 'mousedown', str(button))
        time.sleep(.1)
        run('xdotool', 'mouseup', str(button))
        time.sleep(.2)
        note('click', x=x, y=y, button=button)

    click(300, 210)  # 正常跳過片頭。
    click(300, 210)  # 帝國。
    for key in ['1', '3', '8', 'Return']:
        run('xdotool', 'keydown', key)
        time.sleep(.1)
        run('xdotool', 'keyup', key)
        time.sleep(.1)
        note('key', key=key)
    click(300, 240, 3)
    click(300, 240, 3)
    click(544, 125)
    click(560, 185)
    run('xdotool', 'mousemove', '--window', window, '600', '340')
    time.sleep(.2)

    def capture(name):
        base = out/name
        shutil.copyfile(plan/(name+'.png'), str(base)+'.png')
        run('/src/workplace/settings66-r98-capture', '-window', window, '-base', str(base))
        note('capture', name=name, expected=str(plan/(name+'.png')))

    capture('settings')
    click(410, 65)
    capture('bug')
    proc.wait(timeout=25)
    if proc.returncode:
        raise RuntimeError('正式遊戲退出失敗')
    state = dict(line.split('=', 1) for line in (out/'final.txt').read_text().splitlines())
    assert state['state'] == 'StateSettings' and int(state['updates']) == 900, state
    assert (out/'final.bin').read_bytes() == (plan/'bug.bin').read_bytes()
    note('complete', state=state, full_index_pixels=307200, different=0)
    (out/'receipt.json').write_text(json.dumps(dict(schema='logh3.settings-interface-gui.v1',
                                                  events=events), ensure_ascii=False, indent=2)+'\n')
finally:
    if proc.poll() is None:
        proc.terminate()
        try:
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.wait(timeout=5)
    log.close()
