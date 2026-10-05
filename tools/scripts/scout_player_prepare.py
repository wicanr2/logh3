"""第101輪只讀觀測；正式程式不加測試入口。Docker內執行。"""
import json
import os
from pathlib import Path
import subprocess
import sys

assert Path.cwd() == Path('/src') and os.getuid() == 1000
revision = sys.argv[1]
assert revision.isascii() and revision.isalnum()
source = Path('internal/game/scoutfuel.go')
dest = Path(f'workplace/scout79-r101-{revision}-game.go')
extra = r'''
type ScoutPlayerState struct {
 State string
 Turn, Region, Faction int
 View [2]int
 Idle, Move, Menu, Command bool
 MenuPoint, CommandPoint [2]int
 Selected int
 Enabled []bool
 Units []rules.BattleUnit
 Fleets []rules.BattleFleet
 ReportText string
 ReportFace int
 Tables map[string]string
 Seed uint32
}
// 除執行前明示固定seed外，只讀取正式Game狀態。
func (g *Game) ScoutPlayerSeed(seed uint32) {g.battleRand=rules.NewRand(seed)}
func (g *Game) ScoutPlayerCells() []byte {return append([]byte(nil),g.battleCells(g.battleRegion)...)}
func (g *Game) ScoutPlayerEmptyAt(x,y int) bool {return g.battleUnitAt(x,y)<0&&g.battlePlanetAt(x,y)<0}
func (g *Game) ScoutPlayerStamp() map[string]any {
 s:=g.ScoutPlayerState();m:=map[string]any{"state":s.State,"turn":s.Turn,"view":s.View,"idle":s.Idle,"move":s.Move,"menu":s.Menu,"command":s.Command,"selected":s.Selected,"report_text":s.ReportText,"report_face":s.ReportFace,"unit_count":len(s.Units)}
 for _,f:=range s.Fleets {if f.Index==12 {m["mode"]=f.ScoutMode;m["fuel"]=f.ScoutState;m["scout_out"]=f.ScoutOut}}
 for _,u:=range s.Units {if u.Fleet==12&&u.Squadron<0 {m["unit"]=u}}
 return m
}
func (g *Game) ScoutPlayerState() ScoutPlayerState {
 s:=ScoutPlayerState{State:g.state.String(),Turn:g.battleTurn,Region:g.battleRegion,Faction:g.faction,
 View:[2]int{g.battleView.Col,g.battleView.Row},Idle:g.battleInputIdle(),Move:g.battleMove,
 Menu:g.battleMenu,Command:g.battleCmd,Selected:g.battleCmdUnit,
 Units:append([]rules.BattleUnit(nil),g.battleUnits...),Fleets:append([]rules.BattleFleet(nil),g.battleFleets...)}
 if g.battleCmd {l:=g.res.CommandMenus.Layout;s.CommandPoint=[2]int{g.battleCmdX+l.HighlightDX+8,g.battleCmdY+l.HighlightDY+8};s.Enabled=g.commandMenuEnabled(g.battleCmdKind,g.battleUnits[g.battleCmdUnit])}
 if g.battleMenu {l:=g.res.BattleMenu.Layout;s.MenuPoint=[2]int{g.battleMenuX+l.Highlight.X-l.Window.X+8,g.battleMenuY+l.Highlight.Y-l.Window.Y+8}}
 if g.state==StateTurnReport && g.turnRep.battle && len(g.turnRep.pending)>0 {s.ReportText=fmt.Sprintf("%x",g.turnRep.pending[0].text);s.ReportFace=g.turnRep.pending[0].face}
 if g.slot!=nil {s.Tables=map[string]string{};for k,v:=range map[string][]byte{"fleet":g.slot.Fleet,"ship":g.slot.Ship,"admiral":g.slot.Admiral,"work":g.workRecord(),"stellar":g.res.Stellar,"planet":g.res.Planets.Data} {s.Tables[k]=fmt.Sprintf("%x",v)}}
 if g.battleRand!=nil {s.Seed=g.battleRand.Seed()}
 return s
}
'''
text = source.read_text().replace('"encoding/binary"', '"encoding/binary"\n"fmt"', 1) + extra
with dest.open('x') as stream:
    stream.write(text)
overlay = Path(f'workplace/scout79-r101-{revision}-overlay.json')
with overlay.open('x') as stream:
    json.dump({'Replace': {str(source.resolve()): str(dest.resolve())}}, stream, indent=2)
    stream.write('\n')
subprocess.run(['gofmt', '-w', str(dest)], check=True)
subprocess.run(['python3', 'tools/scripts/fleet_player_gui_prepare.py', 'scout101'+revision], check=True)
parent_overlay = json.loads(Path(f'workplace/round79-fleet-gui-overlay-scout101{revision}.json').read_text())
parent_overlay['Replace'].update(json.loads(overlay.read_text())['Replace'])
main = Path(f'workplace/round79-fleet-gui-main-scout101{revision}.go')
text = main.read_text()
anchor = 'st := a.g.FleetGUIStamp()'
assert text.count(anchor) == 1
text = text.replace(anchor, anchor+'\nst["scout"]=a.g.ScoutPlayerStamp()')
main.write_text(text)
overlay.write_text(json.dumps(parent_overlay, indent=2)+'\n')
subprocess.run(['gofmt', '-w', str(main)], check=True)
