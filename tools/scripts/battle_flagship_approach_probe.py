"""第113輪：公開行進與射擊都以旗艦為目標，不改產品或出貨局面。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('--revision', required=True)
p.add_argument('--scenario', choices=['S1','S2','S3'], default='S3')
p.add_argument('--weapon', choices=['fighter','beam','missile'], default='beam')
p.add_argument('--scan-hit', action='store_true')
p.add_argument('--detach-blockers', action='store_true')
p.add_argument('--solo', action='store_true')
p.add_argument('--standoff', type=int, choices=[64,96], default=0)
p.add_argument('--versus-approach', action='store_true')
p.add_argument('--flank', action='store_true')
p.add_argument('--faction', type=int, choices=[0,1], default=1)
p.add_argument('--observe-winner', action='store_true')
p.add_argument('--composition-only', action='store_true')
a = p.parse_args()
assert a.revision.isascii() and a.revision.isalnum()
assert not a.solo or a.detach_blockers
assert not a.detach_blockers or a.scenario == 'S3'
assert not a.detach_blockers or a.faction == 1
assert not a.flank or a.standoff
assert not a.composition_only or (a.detach_blockers and not a.versus_approach)
source = Path('tools/scripts/battle_winner_player_probe.py')
parent = json.loads(Path('docs/internals/data/battle-winner-player-search.json').read_text())
assert hashlib.sha256(source.read_bytes()).hexdigest() == parent['sources'][str(source)]['sha256']
generator = Path('workplace/winner79-r113-wrapper-'+a.revision+'.py')
with generator.open('x') as f:
    f.write(source.read_text().replace('winner79-r112','winner79-r113'))
subprocess.run(['python3',str(generator),'--revision',a.revision,'--scenario',a.scenario,
                '--weapon',a.weapon,'--flagships'],check=True)
go = Path('workplace/winner79-r113-'+a.revision+'-probe_test.go')
s = go.read_text()
s = s.replace('for _, faction := range []int{1} {', 'for _, faction := range []int{'+str(a.faction)+'} {')
anchor = 'if v.Faction != faction && v.Ships > 0 && v.Squadron >= 0 {'
assert s.count(anchor) == 1
s = s.replace(anchor, 'if v.Faction != faction && v.Ships > 0 && v.Squadron == 0 {')
anchor = '"approach": "projected"'
assert s.count(anchor) == 1
s = s.replace(anchor, '"approach": "projected toward flagships"')
if a.detach_blockers:
    s = s.replace('var entries []map[string]any', 'var entries []map[string]any\nvar compositionSteps []map[string]any')
    s = s.replace('"battle_entries": entries,', '"battle_entries": entries, "composition_steps": compositionSteps,')
    anchor = '''if g.state != StateCommandPhase {
\t\t\t\tt.Fatal("正常S3載入失敗", g.state)
\t\t\t}'''
    assert s.count(anchor) == 1
    slots = '19,21,28,29,30,38' if a.solo else '28,29,38'
    s = s.replace(anchor, anchor + '''
            // 唯讀定位公開地圖入口；編組僅以正式點擊修改。
            mapPX,mapPY:=-1,-1
            for py:=mapY;py<mapY+mapH&&mapPX<0;py++ {for px:=mapX;px<mapX+mapW;px++ {
                if g.mapRegionAt(px,py)==27 {mapPX,mapPY=px,py;break}
            }}
            if mapPX<0 {t.Fatal("地圖沒有星域27入口")}
            click(mapPX,mapPY)
            click(350,290)
            click(617,388)
            click(450,388)
            fi:=g.FleetCommandFleet()
            if g.state!=StateFleetWindow||fi<0||g.fleets[fi].Index!=31 {t.Fatal("公開清單未選到艦隊31",g.state,fi)}
            snapshotComposition:=func(stage string) {
                compositionSteps=append(compositionSteps,map[string]any{"stage":stage,"at_ns":elapsed,
                    "fleet":g.fleets[fi],"region_stock":append([]int(nil),rules.RegionShips(g.regionRecord(27),faction)...),
                    "fleet_table":fmt.Sprintf("%x",g.slot.Fleet),"stellar_table":fmt.Sprintf("%x",g.res.Stellar)})
            }
            snapshotComposition("before public edit")
            click(490,365)
            click(497,392)
            if g.state!=StateFleetChangeOrg||g.editor==nil {t.Fatal("公開編組未開",g.state)}
            spec:=g.res.Screens["screen-fleet-composition"]
            if spec==nil||spec.Windows["overlay"]==nil||g.res.CompositionPanel==nil {t.Fatal("真實編組資源缺項")}
            grid:=spec.Windows["overlay"].Grid
            for _,slot:=range []int{SLOTS} {
                x,y:=-1,-1
                for _,cell:=range g.res.CompositionPanel.GridSlots.Cells {
                    if cell.Index==slot {x,y=grid.X+cell.Col*grid.Tile+8,grid.Y+cell.Row*grid.Tile+8;break}
                }
                if x<0||g.editorGridSlotAt(spec,x,y)!=slot {t.Fatal("公開格盤槽位缺項",slot)}
                before:=len(g.editor.Groups())
                right(x,y)
                if g.msgOn||len(g.editor.Groups())!=before-1 {t.Fatal("公開編組未拆回分隊",slot,g.msgOn)}
                compositionSteps=append(compositionSteps,map[string]any{"stage":"public right remove","slot":slot,"x":x,"y":y,
                    "at_ns":elapsed,"groups":g.editor.Groups(),"region_stock":append([]int(nil),rules.RegionShips(g.regionRecord(27),faction)...)})
            }
            click(510,350)
            if g.state!=StateFleetWindow||g.fleets[fi].Groups!=7-len([]int{SLOTS}) {t.Fatal("公開編組決定未生效",g.state,g.fleets[fi].Groups)}
            snapshotComposition("after public decide")
            click(490,390)
            if g.state!=StateCommandPhase {t.Fatal("公開艦隊視窗未回地圖",g.state)}
            '''.replace('SLOTS',slots))
if a.scan_hit:
    anchor = '''if g.battleUnitAt(x+8, y+8) != i {
						continue
					}'''
    assert s.count(anchor) == 1
    s = s.replace(anchor, '''if g.battleUnitAt(x+8,y+8)!=i {
                        found:=false
                        hitIDs:=map[int]bool{}
                        for py:=y;py<y+16&&!found;py++ {for px:=x;px<x+16;px++ {
                            hit:=g.battleUnitAt(px,py);hitIDs[hit]=true
                            if hit==i {x,y=px-8,py-8;found=true;break}
                        }}
                        if !found {t.Logf("旗艦公開點選完全被遮 fleet%d unit%d xy%d,%d hit%v",u.Fleet,i,u.X,u.Y,hitIDs);continue}
                    }''')
if a.standoff:
    anchor = 'cost := fleetChainRoute(g, g.battleUnits[foe].X, g.battleUnits[foe].Y)'
    assert s.count(anchor) == 1
    destination = '''goalX,goalY:=g.battleUnits[foe].X,g.battleUnits[foe].Y
                    // 玩家目的地離敵旗艦64像素；數值是測試輸入，不改正式傷害。
                    deltaX,deltaY:=u.X-goalX,u.Y-goalY
                    if abs(deltaY)>=abs(deltaX) {if deltaY>=0 {goalY+=64} else {goalY-=64}} else {if deltaX>=0 {goalX+=64} else {goalX-=64}}
                    cost := fleetChainRoute(g,goalX,goalY)'''
    s = s.replace(anchor, destination.replace('64',str(a.standoff)))
    anchor = 'd := rules.Distance16(wx, wy, v.X, v.Y)'
    assert s.count(anchor) == 1
    s = s.replace('v := g.battleUnits[foe]\n\t\t\t\t\t\t\t\td := rules.Distance16(wx, wy, v.X, v.Y)',
                  'd := rules.Distance16(wx, wy, goalX, goalY)')
    s = s.replace('"approach": "projected toward flagships",', '"approach": "projected toward flagships", "standoff_pixels":'+str(a.standoff)+',')
if a.versus_approach:
    anchor = 'click(318, 208+20*faction)'
    assert s.count(anchor) == 1
    s = s.replace(anchor, '''click(318,248)
            if g.state!=StateVersusInput||!g.versus||g.faction!=0 {t.Fatal("正常雙人載入未到帝國交接",g.state)}
            click(432,270)
            click(544,125)
            click(560,125)
            if g.state!=StateVersusInput||g.faction!=1 {t.Fatal("正常雙人未交接同盟",g.state)}
            click(432,270)''')
    # 正式戰場只允許活動陣營開艦隊命令；另一方沿既有對戰移動。
    # 交戰命令階段則兩席均走原探針的公開目標／武器選單。
    s = s.replace('"approach": "projected toward flagships",', '"approach": "projected toward flagships", "versus":true,')
if a.flank:
    anchor = 'deltaX,deltaY:=u.X-goalX,u.Y-goalY\n                    if abs(deltaY)>=abs(deltaX) {if deltaY>=0 {goalY+=64} else {goalY-=64}} else {if deltaX>=0 {goalX+=64} else {goalX-=64}}'
    anchor = anchor.replace('64',str(a.standoff))
    assert s.count(anchor) == 1
    s = s.replace(anchor, 'goalX+='+str(a.standoff)+' // 正常玩家目的地改在敵旗艦東側。')
    s = s.replace('"standoff_pixels":'+str(a.standoff)+',', '"standoff_pixels":'+str(a.standoff)+', "flank":"east",')
if a.observe_winner:
    s = s.replace('func TestBattleFinishNormalProbe(t *testing.T) {',
                  'var winner113Observation func(*Game,*rules.WinnerCapture)\nfunc TestBattleFinishNormalProbe(t *testing.T) {')
    anchor = 'var entries []map[string]any'
    assert s.count(anchor) == 1
    s = s.replace(anchor, anchor+'''
            var occupationResults []json.RawMessage
            winner113Observation=func(observed *Game,event *rules.WinnerCapture) {
                tables:=map[string]string{}
                for k,v:=range map[string][]byte{"fleet":observed.slot.Fleet,"ship":observed.slot.Ship,"admiral":observed.slot.Admiral,
                    "work":observed.workRecord(),"stellar":observed.res.Stellar,"planet":observed.res.Planets.Data} {tables[k]=fmt.Sprintf("%x",v)}
                b,e:=json.Marshal(map[string]any{"region":observed.battleRegion,"battle_turn":observed.battleTurn,
                    "active_faction":observed.faction,"at_ns":elapsed,"seed":observed.battleRand.Seed(),"event":event,
                    "result":observed.battleOccupation.Result(),"done":observed.battleOccupation.Done(),"tables":tables})
                if e!=nil {t.Fatal(e)}
                occupationResults=append(occupationResults,json.RawMessage(b))
            }
            t.Cleanup(func(){winner113Observation=nil})''')
    s = s.replace('"battle_entries": entries,', '"battle_entries": entries, "occupation_results":occupationResults, "active_faction":g.faction,')
    core = Path('internal/game/battlefinish.go')
    frozen = json.loads(Path('docs/internals/data/planet-eligibility.json').read_text())
    assert hashlib.sha256(core.read_bytes()).hexdigest() == frozen['implementation_sources'][str(core)]['sha256']
    anchor = '\t\tg.syncWinnerOccupation()'
    assert core.read_text().count(anchor) == 1
    observed_core = Path('workplace/winner79-r113-'+a.revision+'-observed-battlefinish.go')
    with observed_core.open('x') as f:
        f.write(core.read_text().replace(anchor,anchor+'\nif winner113Observation!=nil {winner113Observation(g,event)}'))
    subprocess.run(['gofmt','-w',str(observed_core)],check=True)
    overlay = Path('workplace/winner79-r113-'+a.revision+'-overlay.json')
    mapping = json.loads(overlay.read_text())
    mapping['Replace'][str(core.resolve())] = str(observed_core.resolve())
    overlay.write_text(json.dumps(mapping,indent=2)+'\n')
if a.composition_only:
    anchor = 'if g.state!=StateCommandPhase {t.Fatal("公開艦隊視窗未回地圖",g.state)}'
    assert s.count(anchor) == 1
    s = s.replace(anchor,anchor+'''
            folder:=filepath.Join(os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT"),fmt.Sprintf("normal-%d",faction))
            if os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT")=="" {t.Fatal("編組診斷必須指定獨立輸出")}
            if e:=os.MkdirAll(folder,0755);e!=nil {t.Fatal(e)}
            b,e:=json.MarshalIndent(map[string]any{"faction":faction,"inputs":inputs,"classification":"public composition only; no battle or RNG"},"","  ")
            if e!=nil {t.Fatal(e)}
            if e=os.WriteFile(filepath.Join(folder,"inputs.json"),append(b,'\\n'),0644);e!=nil {t.Fatal(e)}
            if e=g.SaveSlot(folder,0);e!=nil {t.Fatal(e)}
            if e=g.Frame().WritePNG(filepath.Join(folder,"screen.png"),g.Palette());e!=nil {t.Fatal(e)}
            if e=g.Frame().WriteIndex(filepath.Join(folder,"screen.bin"));e!=nil {t.Fatal(e)}
            t.Log("公開編組完成；未進戰鬥，不是勝後占領")
            return
            ''')
go.write_text(s)
subprocess.run(['gofmt','-w',str(go)],check=True)
print('normal public flagship approach prepared',a.scenario,a.revision,a.weapon)
