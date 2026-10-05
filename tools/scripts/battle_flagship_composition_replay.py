"""第113輪：公開編組的獨立 bootstrap、60 Hz 視窗及存讀。"""
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
parent = json.loads(Path('docs/internals/data/battle-winner-player-search.json').read_text())
source = Path('workplace/winner79-r112-replay-a-generator.py')
assert hashlib.sha256(source.read_bytes()).hexdigest() == parent['sources'][str(source)]['sha256']
input_path = Path('workplace/winner79-r113-p/normal-1/inputs.json')
inputs = json.loads(input_path.read_text())['inputs']
assert len(inputs) == 15 and all(x['at_ns'] == 0 for x in inputs)
s = source.read_text().replace('winner79-r112-replay-', 'flagship79-r113-replay-')
s = s.replace("default='workplace/winner79-r112-latea/normal-1/inputs.json'", "default='"+str(input_path)+"'")
anchor = "    create(stem + '-plan.go', plan)"
assert s.count(anchor) == 1
s = s.replace(anchor, """    plan = plan.replace('g.State() != game.StateCommandPhase || len(lifeReports) != 2',
                        'g.State() != game.StateCommandPhase || len(lifeReports) != 0')
    marker = 'actions = append(actions, a)'
    assert plan.count(marker) == 1
    plan = plan.replace(marker, r'''
        if g.State()==game.StateFleetChangeOrg || g.State()==game.StateFleetWindow {
            name:=fmt.Sprintf("composition-%05d",len(actions))
            if err=g.Frame().WritePNG(filepath.Join(*out,name+".png"),g.Palette());err!=nil {panic(err)}
            if err=g.Frame().WriteIndex(filepath.Join(*out,name+".bin"));err!=nil {panic(err)}
            a["snapshot"]=name
        }
        actions = append(actions, a)
    ''')
    plan = plan.replace('snapshot("post-victory")','snapshot("composition-decided")')
""" + anchor)
anchor = '''    driver = driver.replace("'natural_finish':True,'natural_annihilation':True", "'natural_finish':True,'natural_annihilation':True,'battle_planet_capture':False,'ownership_report':False")'''
assert s.count(anchor) == 1
s = s.replace(anchor, '''    driver = driver.replace("'natural_finish':True,'natural_annihilation':True", "'natural_finish':False,'natural_annihilation':False,'battle_planet_capture':False,'ownership_report':False")''')
s = s.replace('第九回合勝利與零占領', '公開編組拆隊')
generator = Path('workplace/flagship79-r113-replay-'+a.revision+('-gui-wrapper.py' if a.gui else '-wrapper.py'))
with generator.open('x') as f:
    f.write(s)
args = ['python3',str(generator),a.revision,'--input',str(input_path)]
if a.gui:
    args.append('--gui')
subprocess.run(args,check=True)
