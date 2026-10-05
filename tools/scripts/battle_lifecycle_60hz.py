"""第102輪：目前正式程式的正常60 Hz戰鬥長流程。只在Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('revision')
p.add_argument('--gui', action='store_true')
a = p.parse_args()
assert a.revision.isascii() and a.revision.isalnum()
stem = f'workplace/lifecycle79-r102-{a.revision}'
parent = json.loads(Path('docs/internals/data/battle-lifecycle-player.json').read_text())


def historical(name):
    data = Path(name).read_bytes()
    assert hashlib.sha256(data).hexdigest() == parent['sources'][name]['sha256'], name
    return data.decode()


def create(name, value):
    with Path(name).open('x') as stream:
        stream.write(value)


if not a.gui:
    historical('workplace/lifecycle79-r95-normal-b/normal-1/inputs.json')
    # 只沿用已凍結的觀測函式。正式程式正文從目前工作樹讀取。
    old_game = historical('workplace/lifecycle79-r95-gui-game.go')
    marker = '// 僅供觀測：不寫局面、命令、鏡頭或輸入。種子另行明示固定。'
    assert old_game.count(marker) == 1
    game = Path('internal/game/fleetcombat.go').read_text() + old_game[old_game.index(marker):]
    create(stem+'-game.go', game)
    old_main = historical('workplace/lifecycle79-r95-gui-main.go')
    marker = 'var fleetGUIEdges [3]int'
    assert old_main.count(marker) == 1
    helper = old_main[old_main.index(marker):]
    anchor = 'st := a.g.FleetGUIStamp()'
    assert helper.count(anchor) == 1
    helper = helper.replace(anchor, anchor+'\n st["tps"]=ebiten.TPS()')
    main = Path('cmd/gin3/main.go').read_text()
    assert 'ebiten.SetTPS' not in main
    anchor = 'func (a *app) Update() error {'
    assert main.count(anchor) == 1
    main = main.replace('import (', 'import (\n"encoding/json"', 1)
    main = main.replace(anchor, anchor+'\n defer a.fleetGUIObserve()') + '\n' + helper
    create(stem+'-main.go', main)
    create(stem+'-plan.go', historical('workplace/lifecycle79-r95-gui-plan.go'))
    create(stem+'-overlay.json', json.dumps({'Replace': {
        '/src/internal/game/fleetcombat.go': '/src/'+stem+'-game.go',
        '/src/cmd/gin3/main.go': '/src/'+stem+'-main.go',
    }}, indent=2)+'\n')
    subprocess.run(['gofmt', '-w', stem+'-game.go', stem+'-main.go', stem+'-plan.go'], check=True)
    subprocess.run(['go', 'run', '-overlay', stem+'-overlay.json', stem+'-plan.go',
                    '-input', 'workplace/lifecycle79-r95-normal-b/normal-1/inputs.json',
                    '-out', stem+'-plan'], check=True)
    subprocess.run(['go', 'build', '-overlay', stem+'-overlay.json', '-o', stem+'-gin3', './cmd/gin3'], check=True)
else:
    # 建置及計畫是相依步驟，必須等前一命令成功完成。
    assert Path(stem+'-plan/plan.json').is_file(), '先完成計畫及建置，再啟動GUI'
    assert Path(stem+'-gin3').is_file(), '正式GUI建置尚未完成'
    driver = historical('workplace/lifecycle79-r95-gui-driver.py')
    assert driver.count("'test_tps':10") == 1
    driver = driver.replace("'test_tps':10", "'test_tps':60")
    anchor = "    wait(lambda s:match(s,plan['final']),'正常交戰返回Input')"
    assert driver.count(anchor) == 1
    driver = driver.replace(anchor, anchor+"\n    assert status()['tps']==60\n    assert all(t['before']['tps']==60 and t['after']['tps']==60 for t in trace if 'before' in t)")
    create(stem+'-driver.py', driver)
    subprocess.run(['python3', stem+'-driver.py', '--plan', stem+'-plan/plan.json',
                    '--out', stem+'-gui', '--scale', '1', '--binary', '/src/'+stem+'-gin3',
                    '--input-binary', '/src/workplace/order79-r93-input'], check=True)
