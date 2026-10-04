"""第79輪：建立只讀GUI觀測與固定RND overlay，拒絕覆寫。"""
from pathlib import Path
import json
import os
import subprocess
import sys

os.chdir('/src')
revision=sys.argv[1] if len(sys.argv)>1 else 'first'
assert revision.isalnum()
def create(path, text):
    p=Path(path)
    assert p.parent.stat().st_uid==os.getuid() and not p.exists(),path
    p.write_text(text)

game='internal/game/fleetcombat.go'
gameout=f'workplace/round79-fleet-gui-game-{revision}.go'
create(gameout,Path(game).read_text()+'''
// 僅供觀測：不寫局面、命令、鏡頭或輸入。種子另行明示固定。
func (g *Game) FleetGUISeed(seed uint32) { g.battleRand = rules.NewRand(seed) }
func (g *Game) FleetGUIStamp() map[string]any {
 stage,current,commandAt := -1,-1,-1
 if g.fleetCombat.active {stage=int(g.fleetCombat.stage);current=g.fleetCombat.current;commandAt=g.fleetCombat.commandAt}
 return map[string]any{"state":g.state.String(),"turn":g.turn,"battle_turn":g.battleTurn,"region":g.battleRegion,
 "view":[]int{g.battleView.Col,g.battleView.Row},"idle":g.battleInputIdle(),"phase":int(g.battlePhase),
 "battle_menu":g.battleMenu,"battle_cmd":g.battleCmd,"order":int(g.fleetOrder),"msg":g.msgOn,
 "stage":stage,"current":current,"command_at":commandAt,"typed":string(g.typed),"report_layer":g.reportLayer}
}
''')
main='cmd/gin3/main.go'
mainout=f'workplace/round79-fleet-gui-main-{revision}.go'
s=Path(main).read_text().replace('import (','import (\n"encoding/json"',1)
s=s.replace('func (a *app) Update() error {','func (a *app) Update() error {\n defer a.fleetGUIObserve()',1)
create(mainout,s+'''
var fleetGUIEdges [3]int
var fleetGUIRequestCount int
func (a *app) fleetGUIObserve() {
 dir:=os.Getenv("LOGH3_FLEET_GUI_OUT")
 if dir=="" { return }
 if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {fleetGUIEdges[0]++}
 if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {fleetGUIEdges[1]++}
 for k:=ebiten.Key0;k<=ebiten.Key9;k++ {if inpututil.IsKeyJustPressed(k) {fleetGUIEdges[2]++}}
 if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {fleetGUIEdges[2]++}
 reqPath:=filepath.Join(dir,"request.json")
 if b,err:=os.ReadFile(reqPath);err==nil {
  var req struct {Seed *uint32; Name string; Save bool}
  if err=json.Unmarshal(b,&req);err!=nil {panic(err)}
  if req.Seed!=nil {a.g.FleetGUISeed(*req.Seed)}
  if req.Name!="" {
   if filepath.Base(req.Name)!=req.Name {panic("非法快照名稱")}
   prefix:=filepath.Join(dir,req.Name)
   if _,err=os.Stat(prefix+".bin");!os.IsNotExist(err) {panic("拒絕覆寫快照")}
   fb:=a.g.Frame()
   if err=fb.WriteIndex(prefix+".bin");err!=nil {panic(err)}
   if err=fb.WritePNG(prefix+".png",a.g.Palette());err!=nil {panic(err)}
  }
  if req.Save {
   if err=a.g.SaveSlot(dir,0);err!=nil {panic(err)}
  }
  if err=os.Remove(reqPath);err!=nil {panic(err)}
  fleetGUIRequestCount++
 }
 st:=a.g.FleetGUIStamp()
 st["edges"]=fleetGUIEdges;st["requests"]=fleetGUIRequestCount
 held:=ebiten.IsKeyPressed(ebiten.KeyEnter)
 for k:=ebiten.Key0;k<=ebiten.Key9;k++ {held=held||ebiten.IsKeyPressed(k)}
 st["buttons"]=[3]bool{ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight),held}
 st["cursor"]=[]int{a.g.CursorPosition()}
 st["app_ticks"]=a.ticks
 b,err:=json.Marshal(st);if err!=nil {panic(err)}
 tmp:=filepath.Join(dir,"status.tmp")
 if err=os.WriteFile(tmp,b,0644);err!=nil {panic(err)}
 if err=os.Rename(tmp,filepath.Join(dir,"status.json"));err!=nil {panic(err)}
 if trace,err:=os.OpenFile(filepath.Join(dir,"updates.jsonl"),os.O_WRONLY|os.O_CREATE|os.O_APPEND,0644);err==nil {
  if _,err=trace.Write(append(b,'\\n'));err!=nil {trace.Close();panic(err)}
  if err=trace.Close();err!=nil {panic(err)}
 } else {panic(err)}
}
'''.replace('st["cursor"]=[]int{a.g.CursorPosition()}','x,y:=ebiten.CursorPosition();st["cursor"]=[]int{x,y}'))
subprocess.run(['gofmt','-w',gameout,mainout],check=True)
create(f'workplace/round79-fleet-gui-overlay-{revision}.json',json.dumps({'Replace':{str(Path(game).resolve()):str(Path(gameout).resolve()),str(Path(main).resolve()):str(Path(mainout).resolve())}},indent=2)+'\n')
print('只讀狀態／快照、明示RND；X11輸入沒有替代入口')
