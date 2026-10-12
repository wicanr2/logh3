"""規格139：沿既有公開玩家鏈產生現行AI的艦隊交火計畫；禁止種子注入。"""
import json
import os
from pathlib import Path
import re
import subprocess
import sys

root = Path('/src')
os.chdir(root)
batch = Path('workplace/battle129')
assert batch.stat().st_uid == os.getuid()

def create(path, source):
    p = Path(path)
    assert not p.exists() and p.parent.stat().st_uid == os.getuid(), p
    p.write_text(source)

# Only add observation to the same app. The historical seed setter is disabled.
s = Path('tools/scripts/fleet_player_gui_prepare.py').read_text()
s = s.replace('workplace/round79-', 'workplace/battle129/fleet-')
s = s.replace('g.battleRand = rules.NewRand(seed)', 'panic("本輪禁止注入種子")')
s = s.replace("'import (\\n\"encoding/json\"'", "'import (\\n\"encoding/json\"\\n\"github.com/hajimehoshi/ebiten/v2/inpututil\"'")
revision = sys.argv[1] if len(sys.argv)>1 else 'active'
assert revision in ('active','manual','weapons')
sys.argv = [sys.argv[0], revision]
exec(compile(s, 'fleet-observer-prepare', 'exec'))
main = batch/f'fleet-fleet-gui-main-{revision}.go'
s = main.read_text()
assert s.count('st["cursor"] = []int{x, y}') == 1
main.write_text(s.replace('st["cursor"] = []int{x, y}', 'x,y=a.gamePosition(x,y)\nst["cursor"] = []int{x, y}'))
subprocess.run(['gofmt','-w',str(main)],check=True)

source = Path('internal/game/fleetplayerchain_test.go').read_text()
source = source.replace('g.SetPassiveBattleComputer(true) // 保存102的061歷史玩家收據；正常政策依120另驗。', '// 本輪沿預設活動CPU，不切換被動政策。')
source = re.sub(r'\t\t\tseed := func\(\) \{.*?\n\t\t\t\}', '\t\t\tseed := func() {} // 不改或記錄測試種子', source, count=1, flags=re.S)
source = re.sub(r'\t\t\tfor i, u := range g.battleUnits \{\n\t\t\t\tif before, ok := cpu\[i\].*?\n\t\t\t\}', '\t\t\t_ = cpu // 活動CPU可正常移動', source, count=1, flags=re.S)
source = source.replace('"seed": uint32(0x12345678)', '"seed_injection": false').replace('"cpu_position_unchanged": true', '"active_cpu": true')
if revision in ('manual','weapons'):
    source = source.replace('scenes, orders := 0, 0', 'scenes, orders := 0, 0\nweaponChoices,weaponSeen:=0,0\n_ = weaponSeen')
    if revision == 'weapons':
        source = source.replace('turn < rules.MaxBattleTurns*4 && scenes == 0', 'turn < rules.MaxBattleTurns-1 && weaponSeen != 7')
    source = source.replace('case fleetCommandTarget:\n\t\t\t\t\t\t\tclick(240, 470)', '''case fleetCommandTarget:
                            target:=c.candidates[0].Target
                            x,y:=unitScreenPos(g,target)
                            fld:=g.res.Battle.Field
                            if x>=fld.X && y>=fld.Y && x+16<fld.X+fld.Cols*fld.TileW && y+16<fld.Y+fld.Rows*fld.TileH { click(x+8,y+8) } else {click(240,470)}''')
    source = source.replace('case fleetCommandMenu:\n\t\t\t\t\t\t\tclick(c.menuX+16, c.menuY+20)', '''case fleetCommandMenu:
                            weapon:=0
                            if c.menu==screen.CommandMenuCombatWeapon {
                                mask:=0
                                for _,candidate:=range c.candidates {if candidate.Target==c.target {mask=int(candidate.Mask)}}
                                weapon=weaponChoices%3
                                for preferred:=0;preferred<3;preferred++ {if mask&(1<<preferred)!=0 && weaponSeen&(1<<preferred)==0 {weapon=preferred;break}}
                                if mask&(1<<weapon)==0 {weapon=0;for mask&(1<<weapon)==0 && weapon<2 {weapon++}}
                                weaponChoices++
                                weaponSeen |= 1<<weapon
                                t.Logf("正常手動武器%d mask%d",weapon,mask)
                            }
                            click(c.menuX+16,c.menuY+20+weapon*20)''')
create(batch/f'{revision}-player_test.go', source)
test_overlay = {'Replace':{str(root/'internal/game/fleetplayerchain_test.go'):str(root/batch/f'{revision}-player_test.go')}}
create(batch/f'{revision}-test-overlay.json', json.dumps(test_overlay, indent=2)+'\n')

plan = Path('tools/scripts/fleet_player_gui_plan.go').read_text()
plan = plan.replace('var actions []map[string]any', 'if err=g.EnableSoundEffects(true);err!=nil {panic(err)}\nvar actions []map[string]any')
plan = plan.replace('g.Tick(time.Duration(in.NS))', 'g.Tick(time.Duration(in.NS));g.DrainSoundEffects()')
create(batch/f'{revision}-plan.go', plan)
subprocess.run(['gofmt','-w',str(batch/f'{revision}-player_test.go'),str(batch/f'{revision}-plan.go')],check=True)
input_source = Path('tools/scripts/fleet_player_gui_input.go').read_text()
input_source = input_source.replace('"os"', '"os"\n"time"')
input_source = input_source.replace('Type, Button, Key int', 'Type, Button, Key, TapMs int')
input_source = input_source.replace('xtest.FakeInput(c, byte(p.Type), detail, 0, s.DefaultScreen(c).Root, 0, 0, 0)',
    '''xtest.FakeInput(c, byte(p.Type), detail, 0, s.DefaultScreen(c).Root, 0, 0, 0)
c.Sync()
if p.TapMs>0 {
    if p.Type!=xproto.ButtonPress || p.TapMs>100 {panic("非法短按")}
    time.Sleep(time.Duration(p.TapMs)*time.Millisecond)
    xtest.FakeInput(c,xproto.ButtonRelease,detail,0,s.DefaultScreen(c).Root,0,0,0)
    c.Sync()
}''')
create(batch/f'{revision}-input-sync.go', input_source)
print('READY 活動原版AI、無種子注入；工具只觀測正常app並由X11輸入')
