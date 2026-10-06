"""第117輪正常Game回合的唯讀觀測覆蓋；不改正式產品。"""
import argparse
import json
import os
from pathlib import Path
import subprocess

p=argparse.ArgumentParser()
p.add_argument('revision')
a=p.parse_args()
assert Path.cwd()==Path('/src') and os.getuid()==1000
assert a.revision.isascii() and a.revision.isalnum()
stem='workplace/normal79-r117-'+a.revision
replace={}

def create(suffix,text):
    path=Path(stem+suffix)
    assert not path.exists() and path.parent.stat().st_uid==path.parent.stat().st_gid==1000
    with path.open('x') as f:f.write(text)
    return str(path)

files={
 'battle.go': [('g.beginBattlePreparation(g.advanceBattleTurnAfterPreparation)', 'maintenance'),
               ('func (g *Game) advanceBattleTurnAfterPreparation() {','movement'),
               ('func (g *Game) advanceBattleTurnAfterMovement() {','spotting'),
               ('func (g *Game) advanceBattleTurnAfterSpotting() {','fleet-combat'),
               ('func (g *Game) advanceBattleTurnAfterFleetCombat() {','fortress'),
               ('g.battleAssaults = g.planetAssault()', 'planet'),
               ('func (g *Game) finishBattleTurn() {','final'),
               ('func (g *Game) endBattleTurn() {','turn-return')],
 'battleownership.go': [('func (g *Game) finishBattleOwnership() {','final-2')],
 'scoutfuel.go': [('func (g *Game) finishBattleScoutFuel() {','final-3')],
 'promotion.go': [('func (g *Game) finishBattlePromotions() {','final-4')],
 'battlefinish.go': [('func (g *Game) finishBattleDecision() {','final-5')],
}
for filename,markers in files.items():
    path=Path('internal/game')/filename
    text=path.read_text()
    for marker,label in markers:
        assert text.count(marker)==1,(filename,marker)
        call='g.NormalTurn117Record("'+label+'")\n'
        text=text.replace(marker,marker+'\n'+call if marker.endswith('{') else call+marker)
    if filename=='battle.go':
        text=text.replace('import (','import (\n"fmt"\nejson "encoding/json"',1)
        text+='''
var normalTurn117Rows []ejson.RawMessage
func (g *Game) NormalTurn117Record(name string) {
    tables:=map[string]string{}
    for k,v:=range map[string][]byte{"fleet":g.slot.Fleet,"ship":g.slot.Ship,"admiral":g.slot.Admiral,"work":g.workRecord(),"stellar":g.res.Stellar,"planet":g.res.Planets.Data} {tables[k]=fmt.Sprintf("%x",v)}
    var seed any
    if g.battleRand!=nil {seed=g.battleRand.Seed()}
    b,e:=ejson.Marshal(map[string]any{"name":name,"tables":tables,"seed":seed,"state":g.state.String(),"turn":g.turn,"battle_turn":g.battleTurn,"region":g.battleRegion,"view":g.battleView,"units":g.battleUnits,"fleets":g.battleFleets})
    if e!=nil {panic(e)}
    normalTurn117Rows=append(normalTurn117Rows,b)
}
func (g *Game) NormalTurn117Rows() []ejson.RawMessage {return normalTurn117Rows}
func (g *Game) NormalTurn117Seed(seed uint32) {
    if g.state!=StateBattle||g.battleTurn!=1 {panic("不是正常首戰")}
    if g.battleRand==nil {g.battleRand=rules.NewRand(seed)} else {g.battleRand.SetSeed(seed)}
}
func (g *Game) NormalTurn117Done() bool {return g.state==StateBattle&&g.battleTurn==2&&g.battleInputIdle()}
'''
    dest=create('-'+filename,text)
    replace['/src/'+str(path)]='/src/'+dest
path=Path('internal/game/rebellion.go')
text=path.read_text()+'''
func (g *Game) NormalTurn117StrategySeed(seed uint32) {g.rand94DC().SetSeed(seed)}
'''
replace['/src/'+str(path)]='/src/'+create('-rebellion.go',text)
create('-overlay.json',json.dumps({'Replace':replace},indent=2)+'\n')
subprocess.run(['gofmt','-w',*[v.removeprefix('/src/') for v in replace.values()]],check=True)
print('正常回合唯讀觀測',stem)
