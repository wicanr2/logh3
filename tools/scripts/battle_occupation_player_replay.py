"""第107輪：正式bootstrap獨立重播及60 Hz正常GUI，Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('revision')
p.add_argument('--input', default='workplace/occupy79-r107-s3planetb/normal-1/inputs.json')
p.add_argument('--gui', action='store_true')
a = p.parse_args()
assert a.revision.isascii() and a.revision.isalnum()
stem = 'workplace/occupy79-r107-replay-' + a.revision
parent = json.loads(Path('docs/internals/data/battle-lifecycle-player.json').read_text())


def historical(name):
    data = Path(name).read_bytes()
    assert hashlib.sha256(data).hexdigest() == parent['sources'][name]['sha256'], name
    return data.decode()


def create(name, value):
    with Path(name).open('x') as f:
        f.write(value)


if not a.gui:
    game = Path('internal/game/fleetcombat.go').read_text()
    old = historical('workplace/lifecycle79-r95-gui-game.go')
    marker = '// 僅供觀測：不寫局面、命令、鏡頭或輸入。種子另行明示固定。'
    assert old.count(marker) == 1
    game += old[old.index(marker):]
    create(stem + '-game.go', game)
    old = historical('workplace/lifecycle79-r95-gui-main.go')
    helper = old[old.index('var fleetGUIEdges [3]int'):]
    helper = helper.replace('st := a.g.FleetGUIStamp()', 'st := a.g.FleetGUIStamp()\n st["tps"]=ebiten.TPS()')
    main = Path('cmd/gin3/main.go').read_text()
    main = main.replace('import (', 'import (\n"encoding/json"', 1)
    anchor = 'func (a *app) Update() error {'
    assert main.count(anchor) == 1 and 'ebiten.SetTPS' not in main
    main = main.replace(anchor, anchor + '\n defer a.fleetGUIObserve()') + '\n' + helper
    create(stem + '-main.go', main)
    plan = historical('workplace/lifecycle79-r95-gui-plan.go')
    plan = plan.replace('g.State() != game.StateCommandPhase || len(lifeReports) != 2',
                        'g.State() != game.StateBattle || len(lifeReports) != 1')
    # 輸入重播、全六表保存及報告全圖先與獨立Game探針比較。
    anchor = 'for _, name := range rules.SlotFileNames(0) {\n\t\ta, e := os.ReadFile(filepath.Join(*out, "player-save", name))'
    assert plan.count(anchor) == 1
    start = plan.index(anchor)
    end = plan.index('\n\tpoints :=', start)
    plan = plan[:start] + r'''
    snapshot:=func(name string) {
        if err=g.Frame().WritePNG(filepath.Join(*out,name+".png"),g.Palette());err!=nil {panic(err)}
        if err=g.Frame().WriteIndex(filepath.Join(*out,name+".bin"));err!=nil {panic(err)}
        actions=append(actions,map[string]any{"kind":"wait_state","after":g.FleetGUIStamp(),"snapshot":name})
    }
    snapshot("battle-owned")
    step("click",240,470,0,0)
    mx,my:=r.BattleMenu.BattleMenuAt(240,470)
    step("click",mx+20,my+20+4*r.BattleMenu.Layout.Text.RowPitch,0,0)
    if g.State()!=game.StateSaveGame {panic("公開戰場保存未開")}
    step("click",300,r.SlotList.Save.TextTop,0,0)
    snapshot("saved")
    for _,name:=range rules.SlotFileNames(0) {
        x,e:=os.ReadFile(filepath.Join(*out,name));if e!=nil {panic(e)}
        y,e:=os.ReadFile(filepath.Join(*out,"player-save",name));if e!=nil||string(x)!=string(y) {panic("公開保存與完整診斷保存不同 "+name)}
    }
    step("right",300,142,0,0)
    if g.State()!=game.StateCommandPhase {panic("公開保存未回命令")}
    readSaved:=func(name string) []byte {
        b,e:=os.ReadFile(filepath.Join(*out,"player-save",name+".DT0"));if e!=nil {panic(e)}
        return b
    }
    expected,e:=rules.LoadSlot(rules.SlotFiles{Stellar:readSaved("STELLAR"),Planet:readSaved("PLANET"),Admiral:readSaved("ADMIRAL"),Fleet:readSaved("FLEET"),Ship:readSaved("SHIP"),Work:readSaved("WORK"),PlanetHeader:r.Planets.Data[:rules.PlanetHeader]})
    if e!=nil {panic(e)}
    params:=r.UnrestParams
    params.Turn,params.Blend,params.Rate=expected.Turn,expected.Blend,expected.Rate
    for side:=0;side<2;side++ {if id:=rules.StaffAdmiral(expected.Work,rules.PostDomestic,side);id>=0 {params.Ruler[side]=id}}
    expected.RecomputeUnrest(r.Graph,params)
    fleets:=expected.Fleets(r.Weights)
    rules.RecountRegionFleets(expected.Stellar,fleets)
    rules.RecountRegionVisibleFleets(expected.Stellar,fleets)
    step("click",544,125,0,0)
    step("click",560,145,0,0)
    step("click",300,122,0,0)
    if g.State()!=game.StateLoadFaction {panic("公開取進度未到陣營")}
    step("click",318,208+20*parent.Faction,0,0)
    if g.State()!=game.StateCommandPhase {panic("公開取進度未回命令")}
    snapshot("loaded")
    if err=os.Mkdir(filepath.Join(*out,"loaded-save"),0755);err!=nil {panic(err)}
    if err=g.SaveSlot(filepath.Join(*out,"loaded-save"),0);err!=nil {panic(err)}
    for _,name:=range rules.SlotFileNames(0) {
        x,e:=os.ReadFile(filepath.Join(*out,name));if e!=nil {panic(e)}
        if name=="STELLAR.DT0" {x=expected.Stellar}
        y,e:=os.ReadFile(filepath.Join(*out,"loaded-save",name));if e!=nil||string(x)!=string(y) {panic("占領後讀回完整資料不同 "+name)}
    }
''' + plan[end:]
    plan = plan.replace('"public_six_saves_equal_parent": true', '"public_six_saves_equal_diagnostic": true')
    create(stem + '-plan.go', plan)
    create(stem + '-overlay.json', json.dumps({'Replace': {
        '/src/internal/game/fleetcombat.go': '/src/' + stem + '-game.go',
        '/src/cmd/gin3/main.go': '/src/' + stem + '-main.go',
    }}, indent=2) + '\n')
    subprocess.run(['gofmt', '-w', stem + '-game.go', stem + '-main.go', stem + '-plan.go'], check=True)
    subprocess.run(['go', 'run', '-overlay', stem + '-overlay.json', stem + '-plan.go',
                    '-input', a.input, '-out', stem + '-plan'], check=True)
    subprocess.run(['go', 'build', '-overlay', stem + '-overlay.json', '-o', stem + '-gin3', './cmd/gin3'], check=True)
else:
    assert Path(stem + '-plan/plan.json').is_file()
    driver = historical('workplace/lifecycle79-r95-gui-driver.py')
    driver = driver.replace("'test_tps':10", "'test_tps':60")
    anchor = "        if action['kind']=='wait_report':"
    assert driver.count(anchor) == 1
    driver = driver.replace(anchor, """        if action['kind']=='wait_state':
            st=wait(lambda s:match(s,action['after']),'公開存讀快照')
            snapshot(action['snapshot'])
            trace.append({'index':i,'kind':'wait_state','state':st})
            continue
""" + anchor)
    driver = driver.replace("assert (out/p.name).read_bytes()==p.read_bytes(),'GUI診斷保存不同 '+p.name",
                            "assert (out/p.name).read_bytes()==(planpath.parent/'loaded-save'/p.name).read_bytes(),'GUI讀回保存不同 '+p.name")
    driver = driver.replace("'natural_finish':True,'natural_annihilation':True", "'natural_finish':False,'natural_annihilation':False,'battle_planet_capture':True,'ownership_report':True")
    driver = driver.replace('正常主選單、S3、真實X11鍵鼠、命令與交戰；RND test-only，保存由診斷API',
                            '正常主選單、S3、真實X11鍵鼠、行星換手與星域報告；RND test-only，公開存讀')
    anchor = "    snapshot('final')"
    # final現為loaded；plan原final圖是battle，改用已驗loaded快照。
    assert driver.count(anchor) == 1
    driver = driver.replace(anchor, "    assert status()['tps']==60\n    assert all(t['before']['tps']==60 and t['after']['tps']==60 for t in trace if 'before' in t)")
    create(stem + '-driver.py', driver)
    subprocess.run(['python3', stem + '-driver.py', '--plan', stem + '-plan/plan.json',
                    '--out', stem + '-gui', '--scale', '1', '--binary', '/src/' + stem + '-gin3',
                    '--input-binary', '/src/workplace/order79-r93-input'], check=True)
