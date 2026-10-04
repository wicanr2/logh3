"""延長既有正常Game玩家，唯讀觀測自然勝負；測試overlay不改產品。"""
import json
import sys
from pathlib import Path

source = Path("internal/game/fleetplayerchain_test.go").read_text()
test = source[source.index("func TestNormalFleetPlayerChain(t *testing.T) {"):]
assert test.count("func TestNormalFleetPlayerChain") == 1


def replace(old, new):
    global test
    assert test.count(old) == 1, old
    test = test.replace(old, new, 1)


replace("func TestNormalFleetPlayerChain", "func TestBattleFinishNormalProbe")
replace("for _, faction := range []int{0, 1}", "for _, faction := range []int{1}")
replace("click(300, 322)", "click(300, 282) // 公開S1列：第8列，122+8*20。")
replace("scenes, orders := 0, 0", "scenes, orders := 0, 0")
replace("rules.MaxBattleTurns*4 && scenes == 0", "rules.MaxBattleTurns*4 && victoryReports == 0")
replace("u.CruiseSpeed > g.battleUnits[lead].CruiseSpeed", "u.Ships > g.battleUnits[lead].Ships")
replace("\t\t\t\t\tfor _, v := range g.battleUnits {", r'''nearest, nearestDist := -1, int64(1<<62)
                    for j,v:=range g.battleUnits {
                        if v.Faction!=faction&&v.Ships>0&&v.Squadron>=0 {
                            d:=rules.Distance16(u.X,u.Y,v.X,v.Y)
                            if d<nearestDist {nearest,nearestDist=j,d}
                        }
                    }
                    for j, v := range g.battleUnits {
                        if j!=nearest {continue}''')
replace("v.Squadron != 0", "v.Squadron < 0")
replace("\t\t\t\t\t\tcase fleetCommandTarget:\n\t\t\t\t\t\t\tclick(240, 470)", r'''case fleetCommandTarget:
                            chosen:=false
                            for _,candidate:=range c.candidates {
                                if candidate.Mask==0 {continue}
                                x,y:=unitScreenPos(g,candidate.Target)
                                if len(g.fleetCommandHit(x+8,y+8))==0 {continue}
                                click(x+8,y+8);chosen=true;break
                            }
                            if !chosen {click(240,470)}''')
replace("\t\t\t\t\t\tcase fleetCommandMenu:\n\t\t\t\t\t\t\tclick(c.menuX+16, c.menuY+20)", r'''case fleetCommandMenu:
                            row:=0
                            if c.menu==screen.CommandMenuCombatWeapon {
                                for _,candidate:=range c.candidates {
                                    if candidate.Target!=c.target {continue}
                                    for row<2&&candidate.Mask&(1<<row)==0 {row++}
                                    break
                                }
                            }
                            click(c.menuX+16,c.menuY+20+row*g.res.CommandMenus.Layout.RowPitch)''')
replace("\t\t\t\t\t\t\t\tif !g.res.Terrain.LineOpen(g.battleCells(g.battleRegion), u.X, u.Y, wx, wy) {\n\t\t\t\t\t\t\t\t\tcontinue\n\t\t\t\t\t\t\t\t}", "// 目的地仍由battleWorldAt驗通行；正式移動自行繞障。")
replace("cpu := g.snapshotSide(1 - faction)", "// 唯讀觀測；占領定位允許移動勝方，不沿用原被動CPU終點斷言。")
replace("cpu = g.snapshotSide(1 - faction)", "// 下個自然戰場仍沿既有公開入口。")
replace("\t\t\tfor i, u := range g.battleUnits {\n\t\t\t\tif before, ok := cpu[i]; ok && (u.X != before.X || u.Y != before.Y) {\n\t\t\t\t\tt.Fatal(\"被動CPU被移動\")\n\t\t\t\t}\n\t\t\t}", "")
replace("\"cpu_position_unchanged\": true", "\"victory_reports\": victoryReports, \"capture_reports\": captureReports, \"no_state_injection\": true")
replace("elapsed := int64(0)", r'''elapsed := int64(0)
            victoryReports, captureReports := 0, 0
            var observations []map[string]any
            var entries []map[string]any
            var commands []map[string]any
            commandSeen:=map[string]bool{}
            observeCommands:=func(){
                c:=g.fleetCombat
                if !c.active||c.current<0||c.current>=len(g.battleUnits) {return}
                key:=fmt.Sprintf("%d:%d:%d:%d:%d",g.turn,g.battleRegion,g.battleTurn,c.current,c.stage)
                if commandSeen[key] {return};commandSeen[key]=true
                a:=g.battleUnits[c.current]
                targets:=[]map[string]any{}
                for j,b:=range g.battleUnits {
                    if b.Faction==a.Faction||b.Squadron<0 {continue}
                    dist:=rules.FixedDist(a.X,a.Y,b.X,b.Y)
                    if dist>rules.FixedFromInt(400) {continue}
                    targets=append(targets,map[string]any{"unit":j,"fleet":b.Fleet,"ships":b.Ships,"xy":[]int{b.X,b.Y},"detail":b.Detail,"out":b.Out,"action":b.Action,"distance":int(dist>>16),"angle":rules.Atan256(b.X-a.X,b.Y-a.Y)})
                }
                commands=append(commands,map[string]any{"turn":g.turn,"battle_turn":g.battleTurn,"region":g.battleRegion,"current":c.current,"stage":int(c.stage),"human":g.humanSeat(a.Faction),"automatic":c.automatic,"actor":a,"facing":g.battleFleet(a.Fleet).Facing,"candidates":c.candidates,"targets":targets})
            }
            lastReport := ""
            observe := func() {
                if !g.battleFinishing || g.state != StateTurnReport || len(g.turnRep.pending) == 0 {
                    lastReport = ""
                    return
                }
                r := g.turnRep.pending[0]
                identity := fmt.Sprintf("%d:%d:%x",g.battleRegion,r.face,r.text)
                if identity == lastReport { return }
                lastReport = identity
                kind := "victory"
                if g.battleOccupation == nil { victoryReports++ } else {kind="capture";captureReports++}
                tables := map[string]string{}
                for k,v := range map[string][]byte{"fleet":g.slot.Fleet,"ship":g.slot.Ship,"admiral":g.slot.Admiral,"work":g.workRecord(),"stellar":g.res.Stellar,"planet":g.res.Planets.Data} {tables[k]=fmt.Sprintf("%x",v)}
                observations=append(observations,map[string]any{"kind":kind,"text_hex":fmt.Sprintf("%x",r.text),"face":r.face,"at_ns":elapsed,"battle_turn":g.battleTurn,"region":g.battleRegion,"seed":g.battleRand.Seed(),"tables":tables})
                if out:=os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT");out!="" {
                    folder:=filepath.Join(out,fmt.Sprintf("normal-%d",faction)); if e:=os.MkdirAll(folder,0755);e!=nil {t.Fatal(e)}
                    stem:=filepath.Join(folder,fmt.Sprintf("finish-report-%02d",len(observations)))
                    if e:=g.Frame().WritePNG(stem+".png",g.Palette());e!=nil {t.Fatal(e)}
                    if e:=g.Frame().WriteIndex(stem+".bin");e!=nil {t.Fatal(e)}
                }
            }
            defer func(){
                if out:=os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT");out!="" {
                    folder:=filepath.Join(out,fmt.Sprintf("normal-%d",faction));if e:=os.MkdirAll(folder,0755);e!=nil {t.Fatal(e)}
                    b,e:=json.MarshalIndent(map[string]any{"scope":"normal Game helper exploration; independent bootstrap replay pending","scenario":"S1","faction":faction,"victory_reports":victoryReports,"capture_reports":captureReports,"observations":observations,"battle_entries":entries,"commands":commands,"inputs":inputs,"state":g.state.String(),"save_error":fmt.Sprint(g.saveErr)},"","  ");if e!=nil {t.Fatal(e)}
                    if e=os.WriteFile(filepath.Join(folder,"observations.json"),append(b,'\n'),0644);e!=nil {t.Fatal(e)}
                }
            }()''')
replace("\t\t\ttick := func() {\n\t\t\t\td := g.TimerTick()", "\t\t\ttick := func() {\n                observe();observeCommands()\n\t\t\t\td := g.TimerTick()")
replace("\t\t\t\t\tif g.fleetCombat.active {", "\t\t\t\t\tif g.fleetCombat.active {\n observeCommands()")
replace("\t\t\t\tfor _, f := range g.battleFleets {\n\t\t\t\t\tif f.Faction == faction {", r'''eligible,activeShips:=0,0
                for _,u:=range g.battleUnits {if u.Faction==faction {activeShips++;if u.Action==127&&!u.Out {eligible++}}}
                for i,a:=range g.battleUnits {
                    if a.Faction!=faction||a.Squadron<0 {continue}
                    targets:=[]map[string]any{}
                    for j,b:=range g.battleUnits {
                        if b.Faction==a.Faction||b.Squadron<0 {continue}
                        dist:=rules.FixedDist(a.X,a.Y,b.X,b.Y)
                        if dist>rules.FixedFromInt(400) {continue}
                        targets=append(targets,map[string]any{"unit":j,"fleet":b.Fleet,"ships":b.Ships,"xy":[]int{b.X,b.Y},"detail":b.Detail,"out":b.Out,"action":b.Action,"distance":int(dist>>16),"angle":rules.Atan256(b.X-a.X,b.Y-a.Y)})
                    }
                    if len(targets)>0 {
                        candidates,err:=g.fleetCombatCandidates(i)
                        sr:=g.slot.Ship[g.battleUnitPool(a)*rules.ShipRecord:]
                        for _,target:=range targets {
                            b:=g.battleUnits[target["unit"].(int)]
                            raw:=g.slot.Ship[g.battleUnitPool(b)*rules.ShipRecord:]
                            target["raw_detail"],target["raw_action"]=int(raw[0x17]),int(raw[0x1e])
                        }
                        commands=append(commands,map[string]any{"scope":"post-turn pure query; not original command receipt","turn":g.turn,"battle_turn":g.battleTurn,"region":g.battleRegion,"current":i,"stage":-1,"human":g.humanSeat(a.Faction),"actor":a,"raw_action":int(sr[0x1e]),"raw_detail":int(sr[0x17]),"facing":g.battleFleet(a.Fleet).Facing,"candidates":candidates,"query_error":fmt.Sprint(err),"targets":targets})
                    }
                }
                t.Logf("eligibility turn%d ships%d action7F%d roster%v commandErr%v",turn,activeShips,eligible,g.fleetCombat.fleets,g.fleetCombat.err)
                for _, f := range g.battleFleets {
                    if f.Faction == faction {''')
replace("\t\t\tseed := func() {", r'''seed := func() {
                profiles:=[]map[string]any{}
                for _,f:=range g.battleFleets {
                    ships:=0;for _,u:=range g.battleUnits {if u.Fleet==f.Index&&u.Squadron>=0 {ships+=u.Ships}}
                    profiles=append(profiles,map[string]any{"fleet":f.Index,"faction":f.Faction,"unable":f.Flag0,"ships":ships,"xy":[]int{f.X,f.Y}})
                }
                tables:=map[string]string{}
                for k,v:=range map[string][]byte{"fleet":g.slot.Fleet,"ship":g.slot.Ship,"admiral":g.slot.Admiral,"work":g.workRecord(),"stellar":g.res.Stellar,"planet":g.res.Planets.Data} {tables[k]=fmt.Sprintf("%x",v)}
                var beforeSeed any
                if g.battleRand!=nil {beforeSeed=g.battleRand.Seed()}
                entries=append(entries,map[string]any{"at_ns":elapsed,"region":g.battleRegion,"battle_turn":g.battleTurn,"before_seed":beforeSeed,"profiles":profiles,"tables":tables})''')
replace("if scenes == 0 || orders == 0", "if victoryReports == 0 || scenes == 0 || orders == 0")
replace("正常UI未進入交戰", "正常UI未到自然勝負")

prefix = '''package game
import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "testing"
    "github.com/wicanr2/logh3/internal/rules"
    "github.com/wicanr2/logh3/internal/screen"
)
'''
revision = sys.argv[1] if len(sys.argv) > 1 else "first"
assert revision.isalnum()
stem = "finish79-r93" if revision == "first" else "finish79-r93-" + revision
dest = Path("workplace/" + stem + "-probe_test.go")
overlay = Path("workplace/" + stem + "-overlay.json")
assert not dest.exists() and not overlay.exists() and dest.parent.stat().st_uid == 1000
dest.write_text(prefix + test)
overlay.write_text(json.dumps({"Replace": {"/src/internal/game/finish_player_probe_test.go": str(dest.resolve())}}, indent=2)+"\n")
print("正常玩家探針已生成；無正式程式修改")
