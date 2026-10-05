"""第115輪：正常存檔接續自然升級，正式bootstrap及60 Hz GUI重播。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('revision')
p.add_argument('--input', default='workplace/promotion79-r115-h/normal-0/inputs.json')
p.add_argument('--gui', action='store_true')
a = p.parse_args()
assert a.revision.isascii() and a.revision.isalnum()
index = json.loads(Path('docs/internals/data/battle-reinforcement.json').read_text())
source = Path('workplace/reinforce79-r114-replay-c-outer.py')
assert hashlib.sha256(source.read_bytes()).hexdigest() == index['sources'][str(source)]['sha256']
inputs = json.loads(Path(a.input).read_text())
assert inputs['promotion_reports'] == 1 and inputs['capture_reports'] == inputs['victory_reports'] == 0
assert sum(v['kind'] == 'seed' for v in inputs['inputs']) == 1
s = source.read_text().replace('reinforce79-r114-replay-', 'promotion79-r115-replay-')
marker = "    create(stem + '-game.go', game)"
assert s.count(marker) == 1
s = s.replace(marker, '''
    game=game.replace('kind := "battle-report"', 'kind := "battle-report"\\nif r.customWait&&r.face==23 {kind="promotion"}')
    game=game.replace('"kind": kind, "text_hex":', '"kind": kind, "wait":r.wait, "text_hex":')
''' + marker)
marker = "    create(stem + '-plan.go', plan)"
assert s.count(marker) == 1
s = s.replace(marker, r'''
    marker='if err = bootstrap.UseSaveDir(r, filepath.Join(*out, "player-save"));'
    assert plan.count(marker)==1
    plan=plan.replace(marker,r"""
    fixture:="workplace/reinforce79-r114-replay-d-plan/player-save"
    if err=os.Mkdir(filepath.Join(*out,"player-save"),0755);err!=nil {panic(err)}
    for _,name:=range append(rules.SlotFileNames(0),"SAVEDATA.IDX") {
        data,e:=os.ReadFile(filepath.Join(fixture,name));if e!=nil {panic(e)}
        if e=os.WriteFile(filepath.Join(*out,"player-save",name),data,0644);e!=nil {panic(e)}
    }
    """+marker)
    plan=plan.replace('if captures!=1||victories!=2 {panic("實際勝後占領／敗退報告數不同")}',r"""
    promotions:=0
    for _,r:=range lifeReports {if r["kind"]=="promotion"&&r["face"].(int)==23 {promotions++}}
    if captures!=0||victories!=0||promotions!=1 {panic("自然升級／戰後報告數不同")}
    """)
''' + marker)
marker = "    create(stem + '-driver.py', driver)"
assert s.count(marker) == 1
s = s.replace(marker, r'''
    driver=driver.replace('import argparse', 'import argparse\nimport shutil',1)
    marker='proc=subprocess.Popen([a.binary'
    assert driver.count(marker)==1
    driver=driver.replace(marker,"shutil.copytree('/src/workplace/reinforce79-r114-replay-d-plan/player-save',out/'player-save')\n"+marker)
    driver=driver.replace('正常主選單、S3、真實X11鍵鼠、正常雙人載入、增援與第4回合勝後占領；RND test-only，公開存讀',
                          '正常主選單、S1存檔接續、真實X11鍵鼠、自然功績升級；RND test-only，公開存讀')
    driver=driver.replace("'battle_planet_capture':True", "'battle_planet_capture':False,'natural_promotion':True")
    driver=driver.replace("'natural_finish':True,'natural_annihilation':True", "'natural_finish':False,'natural_annihilation':False")
''' + marker)
generator = Path('workplace/promotion79-r115-replay-'+a.revision+('-gui-outer.py' if a.gui else '-outer.py'))
assert not generator.exists()
assert generator.parent.stat().st_uid == generator.parent.stat().st_gid == 1000
with generator.open('x') as f:
    f.write(s)
args = ['python3', str(generator), a.revision, '--input', a.input]
if a.gui:
    args.append('--gui')
subprocess.run(args, check=True)
