"""第114輪：正常移動與占領命令增援，不改正式規則或出貨局面。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('--revision', required=True)
p.add_argument('--direct', action='store_true')
p.add_argument('--hold-main', action='store_true')
p.add_argument('--versus', action='store_true')
p.add_argument('--finish-month', action='store_true')
p.add_argument('--player-updates', action='store_true')
p.add_argument('--trace-input', action='store_true')
p.add_argument('--formal-assets', action='store_true')
p.add_argument('--battle-limit', type=int, choices=[36,108], default=36)
a = p.parse_args()
assert a.revision.isascii() and a.revision.isalnum()
parent = json.loads(Path('docs/internals/data/battle-flagship-approach.json').read_text())
source = Path('tools/scripts/battle_flagship_approach_probe.py')
assert hashlib.sha256(source.read_bytes()).hexdigest() == parent['sources'][str(source)]['sha256']
generator = Path('workplace/reinforce79-r114-outer-'+a.revision+'.py')
with generator.open('x') as f:
    f.write(source.read_text().replace('winner79-r113', 'reinforce79-r114'))
subprocess.run(['python3', str(generator), '--revision', a.revision, '--scenario', 'S1',
                '--faction', '0', '--weapon', 'missile', '--standoff', '96',
                '--observe-winner'], check=True)
go = Path('workplace/reinforce79-r114-'+a.revision+'-probe_test.go')
s = go.read_text()
if a.formal_assets:
    s = s.replace('import (', 'import (\n"encoding/hex"', 1)
    anchor = '\t\t\tfleetChainAssets(t, g)'
    assert s.count(anchor) == 1
    s = s.replace(anchor, anchor+'''
            // 正式bootstrap有這些資源；舊畫面助手沒有，會跳過補給與修護的可見等待。
            var publicNames struct {Factions []struct {Bytes string `json:"bytes"`} `json:"factions"`}
            names,e:=os.ReadFile(filepath.Join(root,"docs/internals/data/faction-and-kind-names.json"))
            if e!=nil {t.Fatal(e)}
            if e=json.Unmarshal(names,&publicNames);e!=nil {t.Fatal(e)}
            for _,name:=range publicNames.Factions {
                b,e:=hex.DecodeString(name.Bytes);if e!=nil {t.Fatal(e)}
                g.res.FactionNames=append(g.res.FactionNames,b)
            }
            g.res.SupplyPanel,e=screen.LoadSupplyPanelSpec(filepath.Join(root,"docs/internals/data/supply-panel.json"))
            if e!=nil {t.Fatal(e)}
            fleetSoundAssets(t,g)
''')
anchor = 'var entries []map[string]any'
assert s.count(anchor) == 1
s = s.replace(anchor, anchor+'\nvar strategicOrders []map[string]any\nreinforcementPhase:=0')
anchor = '"battle_entries": entries,'
assert s.count(anchor) == 1
s = s.replace(anchor, anchor+' "strategic_orders":strategicOrders,')
anchor = '\t\t\tclick(544, 125)\n\t\t\tclick(560, 145)'
assert s.count(anchor) == 1
helper = r'''
            // 所有狀態讀取僅定位正式入口，命令只由 click 寫入。
            regionPoint:=func(region int)(int,int) {
                for py:=mapY;py<mapY+mapH;py++ {for px:=mapX;px<mapX+mapW;px++ {
                    if g.mapRegionAt(px,py)==region {return px,py}
                }}
                t.Fatal("地圖沒有目標星域",region);return -1,-1
            }
            publicOrder:=func(index,target,action int) {
                if g.state!=StateCommandPhase {t.Fatal("增援命令不在正常地圖",g.state)}
                fi:=-1
                for i,f:=range g.fleets {if f.Index==index&&f.Faction==faction {fi=i;break}}
                if fi<0 {t.Fatal("增援艦隊不在正式名冊",index)}
                before:=g.fleets[fi]
                if g.listRegion()!=before.Region {x,y:=regionPoint(before.Region);click(x,y)}
                click(350,290)
                row:=g.fleetListRowOf(fi)
                if row<0||row>=5 {t.Fatal("增援艦隊不在正常清單首頁",index,row)}
                click(450,308+20*row)
                if g.state!=StateFleetWindow||g.FleetCommandFleet()!=fi {t.Fatal("公開清單選錯艦隊",index,g.state)}
                click(490,345)
                if g.state!=StateFleetActionMenu||!g.fleetMenuEnabled()[action] {t.Fatal("公開作戰命令停用",index,action,g.state)}
                click(497,312+20*action)
                if g.state!=StateFleetTargetPick||!g.FleetTargetsAllowed()[target] {t.Fatal("公開目的地不允許",index,target,g.state)}
                x,y:=regionPoint(target);click(x,y)
                if g.state!=StateFleetConfirm {t.Fatal("公開目的地未到確認",index,g.state)}
                click(432,263)
                after:=g.fleets[fi]
                if g.state!=StateCommandPhase||after.Target!=target||after.Move!=action {t.Fatal("公開增援命令未寫入",index,g.state,after)}
                strategicOrders=append(strategicOrders,map[string]any{"phase":reinforcementPhase,"turn":g.turn,"at_ns":elapsed,
                    "before":before,"after":after,"fleet_table":fmt.Sprintf("%x",g.slot.Fleet),"stellar_table":fmt.Sprintf("%x",g.res.Stellar)})
                t.Logf("公開增援phase%d fleet%d region%d target%d action%d",reinforcementPhase,index,after.Region,target,action)
            }
            reinforcement:=func() {
                if reinforcementPhase>=2 {return}
                for _,index:=range []int{4,5,10,12} {
                    target,action:=11,fleetActionMove
                    if reinforcementPhase==1 {target,action=13,fleetActionOccupy}
                    publicOrder(index,target,action)
                }
                reinforcementPhase++
            }
'''
s = s.replace(anchor, helper+anchor)
if a.versus:
    assert a.hold_main
    anchor = 'click(318, 208+20*faction)'
    assert s.count(anchor) == 1
    s = s.replace(anchor, '''click(318,248)
                if g.state!=StateVersusInput||g.faction!=0||!g.versus {t.Fatal("正常雙人載入未到帝國交接",g.state,g.faction)}
                click(432,270)''')
    anchor = 'click(544, 125)\n\t\t\tclick(560, 125)'
    assert s.count(anchor) == 1
    s = s.replace(anchor, anchor+'''
                if g.state!=StateVersusInput||g.faction!=1 {t.Fatal("正常雙人實行未到同盟交接",g.state,g.faction)}
                click(432,270)
                click(544,125);click(560,125)''')
    anchor = 'click(544, 125)\n\t\t\t\t\t\tclick(560, 125)'
    assert s.count(anchor) == 1
    s = s.replace(anchor, anchor+'''
                        if g.state!=StateVersusInput||g.faction!=1 {t.Fatal("後續雙人未到同盟交接",g.state,g.faction)}
                        click(432,270)
                        click(544,125);click(560,125)''')
    anchor = 'case StateTurnThinking, StateTurnWindow, StateTurnReport, StateEngagement:'
    assert s.count(anchor) == 3
    s = s.replace(anchor, 'case StateVersusAccount, StateVersusShipOrder:\n right(300,240)\ncase StateVersusInput:\n click(432,270)\n'+anchor)
    s = s.replace('click(432,270)\n                        click(544,125);click(560,125)', '''click(432,270)
                        for n:=0;n<3&&(g.state==StateVersusAccount||g.state==StateVersusShipOrder);n++ {right(300,240)}
                        click(544,125);click(560,125)''')
    anchor = 'case fleetCommandTarget:\n'
    assert s.count(anchor) == 1
    s = s.replace(anchor, anchor+'''if g.battleUnits[c.current].Faction!=faction {click(240,470);break}
''')
s = s.replace('rules.MaxBattleTurns*4', str(a.battle_limit))
if a.finish_month:
    s = s.replace('&& captureReports == 0', '&& (captureReports == 0 || g.state != StateCommandPhase)')
anchor = 'case StateTurnThinking, StateTurnWindow, StateTurnReport, StateEngagement:'
assert s.count(anchor) == 3
s = s.replace(anchor, '''case StateOccupyAsk:
                    t.Logf("公開途中占領問話 fleet%d next%d",g.fleets[g.occupyAsk.fleet].Index,g.occupyAsk.next)
                    answered:=false
                    for y:=0;y<480&&!answered;y++ {for x:=0;x<640;x++ {
                        if g.res.OccupyAsk.ButtonAt(x,y)==0 {click(x,y);answered=true;break}
                    }}
                    if !answered {t.Fatal("正式途中占領問話沒有占領按鈕")}
                '''+anchor)
if a.direct:
    s = s.replace('reinforcementPhase>=2', 'reinforcementPhase>=1')
    s = s.replace('target,action:=11,fleetActionMove', 'target,action:=13,fleetActionOccupy')
if a.hold_main:
    assert not a.direct
    anchor = 'after:=g.fleets[fi]\n                if g.state!=StateCommandPhase||after.Target!=target||after.Move!=action'
    assert s.count(anchor) == 1
    s = s.replace(anchor, '''after:=g.fleets[fi]
                expectedTarget,expectedAction:=target,action
                if target==before.Region {
                    expectedTarget,expectedAction=rules.NoTarget,rules.MoveParked
                    if g.state!=StateFleetWindow {t.Fatal("正常待命取消未保留艦隊視窗",g.state)}
                    click(490,390)
                }
                if g.state!=StateCommandPhase||after.Target!=expectedTarget||after.Move!=expectedAction''')
    anchor = 'for _,index:=range []int{4,5,10,12} {'
    assert s.count(anchor) == 1
    s = s.replace(anchor, '''if reinforcementPhase==0 {publicOrder(1,11,fleetActionMove)}
                ids:=[]int{4,5,10,12}
                if reinforcementPhase==1 {ids=append(ids,1)}
                for _,index:=range ids {''')
anchor = '\t\t\t\tt.Fatal("正常S3載入失敗", g.state)\n\t\t\t}'
assert s.count(anchor) == 1
s = s.replace(anchor, anchor+'\nreinforcement()')
anchor = '\t\t\t\t\t\tclick(544, 125)\n\t\t\t\t\t\tclick(560, 125)'
assert s.count(anchor) == 1
s = s.replace(anchor, '\t\t\t\t\t\treinforcement()\n'+anchor)
if a.finish_month:
    anchor = '\t\t\t\t\t\treinforcement()'
    assert s.count(anchor) == 1
    s = s.replace(anchor, 'if captureReports>0 {break}\n'+anchor)
if a.player_updates:
    anchor = '\t\t\t\tg.Tick(d)'
    assert s.count(anchor) == 1
    s = s.replace(anchor, anchor+'\n g.MoveMouse(g.mouseX,g.mouseY)\n g.Frame()')
if a.trace_input:
    anchor = '\t\t\tclick := func(x, y int) {'
    assert s.count(anchor) == 1
    stamp = '''stamp:=func()map[string]any {
        stage,current,at:=-1,-1,-1
        if g.fleetCombat.active {stage,current,at=int(g.fleetCombat.stage),g.fleetCombat.current,g.fleetCombat.commandAt}
        return map[string]any{"state":g.state.String(),"turn":g.turn,"battle_turn":g.battleTurn,"region":g.battleRegion,
            "view":[]int{g.battleView.Col,g.battleView.Row},"idle":g.battleInputIdle(),"phase":int(g.battlePhase),
            "battle_menu":g.battleMenu,"battle_cmd":g.battleCmd,"order":int(g.fleetOrder),"msg":g.msgOn,
            "stage":stage,"current":current,"command_at":at,"typed":string(g.typed),"report_layer":g.reportLayer}
    }
'''
    s = s.replace(anchor, stamp+anchor+'\n beforeInput:=stamp()\n defer func(){inputs[len(inputs)-1]["before"]=beforeInput;inputs[len(inputs)-1]["after"]=stamp()}()')
    anchor = '\t\t\tright := func(x, y int) {'
    assert s.count(anchor) == 1
    s = s.replace(anchor, anchor+'\n beforeInput:=stamp()\n defer func(){inputs[len(inputs)-1]["before"]=beforeInput;inputs[len(inputs)-1]["after"]=stamp()}()')
go.write_text(s)
subprocess.run(['gofmt','-w',str(go)],check=True)
print('normal public reinforcement prepared',a.revision)
