"""第114輪：正常增援與勝後占領的獨立入口、60 Hz視窗及公開存讀。"""
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
p.add_argument('--input', default='workplace/reinforce79-r114-m/normal-0/inputs.json')
a = p.parse_args()
assert a.revision.isascii() and a.revision.isalnum()
parent = json.loads(Path('docs/internals/data/battle-winner-player-search.json').read_text())
source = Path('workplace/winner79-r112-replay-a-generator.py')
assert hashlib.sha256(source.read_bytes()).hexdigest() == parent['sources'][str(source)]['sha256']
input_path = Path(a.input)
inputs = json.loads(input_path.read_text())
assert inputs['capture_reports'] == 1 and inputs['victory_reports'] == 2
assert sum(x['kind'] == 'seed' for x in inputs['inputs']) == 1
s = source.read_text().replace('winner79-r112-replay-', 'reinforce79-r114-replay-')
anchor = "    create(stem + '-plan.go', plan)"
assert s.count(anchor) == 1
s = s.replace(anchor, r'''
    marker = 'Seed uint32'
    assert plan.count(marker)==1
    plan = plan.replace(marker, marker+'\nBefore,After map[string]any')
    marker = 'step(in.Kind, in.X, in.Y, 0, in.Seed)'
    assert plan.count(marker)==1
    plan = plan.replace(marker, marker+r"""
        if in.After!=nil {
            expected,_:=json.Marshal(in.After)
            actual,_:=json.Marshal(g.FleetGUIStamp())
            if string(expected)!=string(actual) {panic(fmt.Sprintf("正常輸入%d首次控制分歧：expected %s actual %s",len(actions),expected,actual))}
        }
    """)
    plan = plan.replace('g.State() != game.StateCommandPhase || len(lifeReports) != 2',
                        'g.State() != game.StateCommandPhase')
    marker = 'if g.State() != game.StateCommandPhase {'
    assert plan.count(marker) == 1
    plan = plan.replace(marker, r"""
    terminal,e:=json.MarshalIndent(map[string]any{"state":g.FleetGUIStamp(),"reports":lifeReports,"actions":actions},"","  ")
    if e!=nil {panic(e)}
    if e=os.WriteFile(filepath.Join(*out,"before-save.json"),terminal,0644);e!=nil {panic(e)}
    """+marker)
    marker = 'if err = g.Frame().WriteIndex(filepath.Join(*out, "final.bin"));'
    assert plan.count(marker)==1
    plan = plan.replace(marker, r"""
    captures,victories:=0,0
    for _,r:=range lifeReports {
        if r["kind"]=="capture" {captures++}
        if r["kind"]=="victory" {victories++}
    }
    if captures!=1||victories!=2 {panic("實際勝後占領／敗退報告數不同")}
    """+marker)
    marker = 'actions = append(actions, a)'
    assert plan.count(marker) == 1
    plan = plan.replace(marker, r"""
        if g.State()==game.StateFleetWindow||g.State()==game.StateFleetConfirm||g.State()==game.StateVersusInput {
            name:=fmt.Sprintf("reinforcement-%05d",len(actions))
            if err=g.Frame().WritePNG(filepath.Join(*out,name+".png"),g.Palette());err!=nil {panic(err)}
            if err=g.Frame().WriteIndex(filepath.Join(*out,name+".bin"));err!=nil {panic(err)}
            a["snapshot"]=name
        }
        actions = append(actions, a)
    """)
    marker = 'step("click",318,208+20*parent.Faction,0,0)'
    assert plan.count(marker) == 1
    plan = plan.replace(marker, r"""
    step("click",318,248,0,0)
    if g.State()!=game.StateVersusInput {panic("讀回未保留雙人模式")}
    step("click",432,270,0,0)
    for n:=0;n<3&&(g.State()==game.StateVersusAccount||g.State()==game.StateVersusShipOrder);n++ {step("right",300,240,0,0)}
    """)
''' + anchor)
anchor = '''    driver = driver.replace("'natural_finish':True,'natural_annihilation':True", "'natural_finish':True,'natural_annihilation':True,'battle_planet_capture':False,'ownership_report':False")'''
assert s.count(anchor) == 1
s = s.replace(anchor, '''    driver = driver.replace("'natural_finish':True,'natural_annihilation':True", "'natural_finish':True,'natural_annihilation':True,'battle_planet_capture':True,'ownership_report':False,'versus_loaded':True")''')
s = s.replace('第九回合勝利與零占領', '正常雙人載入、增援與第4回合勝後占領')
s = s.replace('主選單、S3、真實X11', '主選單、S1、真實X11')
generator = Path('workplace/reinforce79-r114-replay-'+a.revision+('-gui-outer.py' if a.gui else '-outer.py'))
with generator.open('x') as f:
    f.write(s)
args = ['python3',str(generator),a.revision,'--input',str(input_path)]
if a.gui:
    args.append('--gui')
subprocess.run(args,check=True)
