"""第107輪：出貨劇本的正常玩家占領探索，禁止局面注入與重擲種子。"""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src')
p = argparse.ArgumentParser()
p.add_argument('--scenario', choices=['S1', 'S2', 'S3'], required=True)
p.add_argument('--revision', required=True)
p.add_argument('--faction', type=int, choices=[0, 1], default=1)
p.add_argument('--weapon', choices=['fighter', 'beam', 'missile'], default='fighter')
p.add_argument('--approach', choices=['existing', 'projected'], default='existing')
p.add_argument('--objective', choices=['fleet', 'planet'], default='fleet')
a = p.parse_args()
assert a.revision.isascii() and a.revision.isalnum()
source = Path('tools/scripts/battle_finish_victory_probe.py')
parent = json.loads(Path('docs/internals/data/battle-finish-natural.json').read_text())
assert hashlib.sha256(source.read_bytes()).hexdigest() == parent['sources'][str(source)]['sha256']
text = source.read_text().replace('finish79-r94', 'occupy79-r107')
generator = Path('workplace') / ('occupy79-r107-prepare-' + a.revision + '.py')
assert not generator.exists()
generator.write_text(text)
subprocess.run(['python3', str(generator), a.revision], check=True)
go = Path('workplace') / ('occupy79-r107-' + a.revision + '-probe_test.go')
s = go.read_text()
assert s.count('for _, faction := range []int{1}') == 1
s = s.replace('for _, faction := range []int{1}', 'for _, faction := range []int{%d}' % a.faction)
slot_y = {'S1': 282, 'S2': 302, 'S3': 322}[a.scenario]
assert s.count('click(300, 322)') == 1
s = s.replace('click(300, 322)', 'click(300, %d)' % slot_y)
s = s.replace('"scenario":"S3"', '"scenario":"' + a.scenario + '"')
s = s.replace('rules.MaxBattleTurns*4 && victoryReports == 0',
              'rules.MaxBattleTurns*4 && captureReports == 0')
s = s.replace('if victoryReports == 0 || scenes == 0 || orders == 0',
              'if captureReports == 0 || scenes == 0 || orders == 0')
s = s.replace('正常UI未到自然勝負', '正常UI未到占領報告')
anchor = 'seed := func() {'
assert s.count(anchor) == 1
s = s.replace(anchor, 'seeded := false\n            seed := func() {', 1)
anchor = 'inputs = append(inputs, map[string]any{"kind": "seed"'
assert s.count(anchor) == 1
s = s.replace(anchor, 'if seeded { return }; seeded = true\n                ' + anchor, 1)
anchor = 'row=2;for row>0&&candidate.Mask&(1<<row)==0 {row--}'
assert s.count(anchor) == 1
if a.weapon == 'beam':
    s = s.replace(anchor, 'row=0;for row<2&&candidate.Mask&(1<<row)==0 {row++}')
elif a.weapon == 'missile':
    s = s.replace(anchor, 'row=1;if candidate.Mask&2==0 {row=2;if candidate.Mask&4==0 {row=0}}')
if a.approach == 'projected':
    # 只規劃玩家輸入：依正式戰鬥行程預估下一次移動後可進原有300距離。
    anchor = 'if near <= int64(300*rules.FixOne) || g.battleTurn == rules.MaxBattleTurns {'
    assert s.count(anchor) == 1
    s = s.replace(anchor, '_, travel := g.fleetSpeeds(u.Fleet, true)\n                    if near <= int64(travel+300*rules.FixOne) || g.battleTurn == rules.MaxBattleTurns {')
# 前一戰之後仍可能停在全圖星域報告。只以公開點擊看完，再送出下一回合。
anchor = 'if g.state != StateBattle {\n\t\t\t\t\tif g.state != StateCommandPhase {'
assert s.count(anchor) == 1
s = s.replace(anchor, '''if g.state != StateBattle {
                    for n:=0;n<15000 && g.state!=StateCommandPhase && g.state!=StateBattle;n++ {
                        switch g.state {
                        case StateTurnThinking,StateTurnWindow,StateTurnReport,StateEngagement:
                            tick()
                        case StateOccupationReport,StateRebellionReport,StateRegionFlipReport,StateAdmiralFateReport,StateShipDelivery,StatePhezzanReport:
                            click(240,240)
                        default: t.Fatal("戰後公開報告停在",g.state)
                        }
                    }
                    if g.state==StateBattle {seed()} else {
\t\t\t\t\tif g.state != StateCommandPhase {''')
anchor = '\t\t\t\t\t// 下個自然戰場仍沿既有公開入口。\n\t\t\t\t}'
assert s.count(anchor) == 1
s = s.replace(anchor, anchor.replace('\n\t\t\t\t}', '\n                    }\n\t\t\t\t}'))
s = s.replace('"scenario":"' + a.scenario + '"',
              '"scenario":"' + a.scenario + '","weapon_preference":"' + a.weapon + '","approach":"' + a.approach + '"')
if a.objective == 'planet':
    # 沿019的公開戰鬥選點。讀取行星位置只用於測試玩家規劃。
    anchor = 'for _, v := range g.battleUnits {\n\t\t\t\t\t\tif v.Faction == faction || v.Ships == 0 || v.Squadron != 0 {'
    assert s.count(anchor) == 1
    s = s.replace(anchor, '''var viewTargets []rules.BattleUnit
                    for _,p:=range g.battlePlanets {
                        if p.Owner!=faction {viewTargets=append(viewTargets,rules.BattleUnit{Faction:1-faction,Ships:1,Squadron:0,X:p.X,Y:p.Y})}
                    }
                    for _,v:=range viewTargets {
                        if v.Faction == faction || v.Ships == 0 || v.Squadron != 0 {''')
    anchor = '\t\t\t\t\tnear := best'
    assert s.count(anchor) == 1
    s = s.replace(anchor, '''goalX,goalY := g.battleUnits[foe].X,g.battleUnits[foe].Y
                    best = int64(1<<62)
                    for _,p := range g.battlePlanets {
                        if p.Owner==faction {continue}
                        d:=rules.Distance16(u.X,u.Y,p.X,p.Y)
                        if d<best {best,goalX,goalY=d,p.X,p.Y}
                    }
                    if best==int64(1<<62) {continue}
                    near := best''')
    s = s.replace('fleetChainRoute(g, g.battleUnits[foe].X, g.battleUnits[foe].Y)',
                  'fleetChainRoute(g, goalX,goalY)')
    s = s.replace('v := g.battleUnits[foe]\n\t\t\t\t\t\t\t\td := rules.Distance16(wx, wy, v.X, v.Y)',
                  'd := rules.Distance16(wx,wy,goalX,goalY)')
    # 定位成功後留在公開戰場輸入，再以保存鏈驗證。勝方占領另計，不混用。
    s = s.replace('victoryReports, captureReports := 0, 0', 'victoryReports, captureReports := 0, 0\n var planetCaptures []map[string]any\n var ownershipReports []map[string]any')
    anchor = 'observe();observeCommands()'
    assert s.count(anchor) == 1
    s = s.replace(anchor, '''observe();observeCommands()
                if g.state==StateEngagement && len(g.engQueue)>0 && g.engQueue[0].label==screen.EngagementCaptured {
                    e:=g.engQueue[0]
                    if len(planetCaptures)==0 {
                        planetCaptures=append(planetCaptures,map[string]any{"at_ns":elapsed,"region":g.battleRegion,"battle_turn":g.battleTurn,"engagement":map[string]int{"planet":e.planet,"fleet":e.fleet,"label":int(e.label)},"planets":append([]rules.BattlePlanet(nil),g.battlePlanets...),"assaults":append([]rules.PlanetAssaultHit(nil),g.battleAssaults...)})
                    }
                }
                if len(planetCaptures)>0 && g.state==StateTurnReport && g.turnRep.battle && !g.battleFinishing && len(g.turnRep.pending)>0 && len(ownershipReports)==0 {
                    r:=g.turnRep.pending[0]
                    ownershipReports=append(ownershipReports,map[string]any{"at_ns":elapsed,"region":g.battleRegion,"battle_turn":g.battleTurn,"face":r.face,"text_hex":fmt.Sprintf("%x",r.text),"stellar_hex":fmt.Sprintf("%x",g.regionRecord(g.battleRegion)),"planets":append([]rules.BattlePlanet(nil),g.battlePlanets...)})
                }''')
    s = s.replace('turn < rules.MaxBattleTurns*4 && captureReports == 0',
                  'turn < rules.MaxBattleTurns*4 && len(ownershipReports)==0')
    s = s.replace('if captureReports == 0 || scenes == 0 || orders == 0',
                  'if len(planetCaptures)==0 || len(ownershipReports)==0 || orders == 0')
    s = s.replace('"capture_reports":captureReports,"observations"',
                  '"capture_reports":captureReports,"planet_captures":planetCaptures,"ownership_reports":ownershipReports,"observations"')
    s = s.replace('"capture_reports": captureReports, "no_state_injection"',
                  '"capture_reports": captureReports,"planet_captures":planetCaptures,"ownership_reports":ownershipReports, "no_state_injection"')
    s = s.replace('"scenario":"' + a.scenario + '"',
                  '"scenario":"' + a.scenario + '","objective":"planet"')
go.write_text(s)
print('prepared', a.scenario, 'normal public input probe', a.revision)
