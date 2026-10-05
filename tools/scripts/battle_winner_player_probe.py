"""#79正常勝後占領輸入探針；沿用出貨劇本，不改局面或產品。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('--revision', required=True)
p.add_argument('--reserve31', action='store_true')
p.add_argument('--scenario', choices=['S1','S2','S3'], default='S3')
p.add_argument('--weapon', choices=['fighter','beam','missile'], default='fighter')
p.add_argument('--flagships', action='store_true')
p.add_argument('--first-battle-cruise', action='store_true')
p.add_argument('--pursue28', action='store_true')
p.add_argument('--expect-late-victory', action='store_true')
a = p.parse_args()
assert a.revision.isascii() and a.revision.isalnum()
source = Path('tools/scripts/battle_occupation_player_probe.py')
parent = json.loads(Path('docs/internals/data/battle-occupation-player.json').read_text())
assert hashlib.sha256(source.read_bytes()).hexdigest() == parent['sources'][str(source)]['sha256']
text = source.read_text().replace('occupy79-r107','winner79-r112')
generator = Path('workplace/winner79-r112-base-'+a.revision+'.py')
with generator.open('x') as f:
    f.write(text)
subprocess.run(['python3',str(generator),'--revision',a.revision,'--scenario',a.scenario,
                '--weapon',a.weapon,'--approach','projected','--objective','fleet'],check=True)
go = Path('workplace/winner79-r112-'+a.revision+'-probe_test.go')
subprocess.run(['gofmt','-w',str(go)],check=True)
text = go.read_text()
if a.expect_late_victory:
    assert a.scenario == 'S3' and not a.first_battle_cruise and not a.pursue28
    anchor = 'rules.MaxBattleTurns*4 && captureReports == 0'
    assert text.count(anchor) == 1
    text = text.replace(anchor, 'rules.MaxBattleTurns*4 && victoryReports == 0')
    anchor = 'if captureReports == 0 || scenes == 0 || orders == 0 {'
    assert text.count(anchor) == 1
    text = text.replace(anchor, 'if victoryReports != 1 || captureReports != 0 || g.battleTurn != 10 || scenes == 0 || orders == 0 {')
    text = text.replace('正常UI未到占領報告', '正常第九回合勝利或零占領不符')
metadata = {'reserve_fleet31':a.reserve31,'flagships_first':a.flagships,
            'first_battle_cruise':a.first_battle_cruise,'pursue28':a.pursue28,
            'expect_late_victory':a.expect_late_victory}
anchor = '"scope": "normal Game helper exploration; independent bootstrap replay pending"'
assert text.count(anchor) == 1
text = text.replace(anchor, anchor+','+','.join('"'+k+'":'+str(v).lower() for k,v in metadata.items()))
if a.pursue28:
    assert a.scenario == 'S3' and a.first_battle_cruise
    anchor = '\n\t\t\t\t\t\tclick(544, 125)'
    assert text.count(anchor) == 1
    text = text.replace(anchor, '''
                        point:=func(region int)(int,int){for y:=mapY;y<mapY+mapH;y+=4 {for x:=mapX;x<mapX+mapW;x+=4 {if g.mapRegionAt(x,y)==region{return x,y}}};t.Fatal("星圖找不到星域",region);return 0,0}
                        sx,sy:=point(26);dx,dy:=point(28)
                        if g.selectedRegion!=26 {click(sx,sy)}
                        click(350,290)
                        for row:=0;row<5;row++ {
                            click(450,308+20*row)
                            if g.state!=StateFleetWindow {t.Fatal("正常追擊沒有艦隊視窗",row,g.state)}
                            click(490,345);click(497,332)
                            if g.state!=StateFleetTargetPick {t.Fatal("正常追擊沒有占領選點",g.state)}
                            click(dx,dy)
                            if g.state!=StateFleetConfirm {t.Fatal("正常追擊星域不可選",g.state)}
                            click(432,263)
                            if g.state!=StateCommandPhase {t.Fatal("正常追擊命令未提交",g.state)}
                        }
''' + anchor)
if a.first_battle_cruise:
    anchor = 'seed := func() {'
    assert text.count(anchor) == 1
    text = text.replace(anchor, 'entryCalls:=0\n            seed := func() {entryCalls++')
    anchor = '\n\t\t\t\t\tclick(g.battleCmdX+l.HighlightDX+8'
    assert text.count(anchor) == 1
    text = text.replace(anchor, '\n                    if entryCalls==1 {row=0}' + anchor)
if a.flagships:
    anchor = 'chosen := false\n\t\t\t\t\t\t\tfor _, candidate := range c.candidates {'
    assert text.count(anchor) == 1
    text = text.replace(anchor, 'chosen := false\n                            for pass:=0;pass<2&&!chosen;pass++ {\n                            for _, candidate := range c.candidates {\n                                if (g.battleUnits[candidate.Target].Squadron==0)!=(pass==0) {continue}')
    anchor = '\n\t\t\t\t\t\t\tif !chosen {'
    assert text.count(anchor) == 1
    text = text.replace(anchor, '\n                            }' + anchor)
if a.reserve31:
    # 測試玩家不向31下艦隊命令，讓其他艦隊自行接敵；仍由公開UI下令。
    anchor = 'if u.Faction != faction || u.Squadron != 0 || u.Out || g.battleFleet(u.Fleet).Flag0 {'
    assert text.count(anchor) == 1
    text = text.replace(anchor,anchor.replace('g.battleFleet(u.Fleet).Flag0','g.battleFleet(u.Fleet).Flag0 || u.Fleet==31'))
text = text.replace('"no_state_injection": true', '"no_state_injection": true,"reserve_fleet31":'+str(a.reserve31).lower())
go.write_text(text)
subprocess.run(['gofmt','-w',str(go)],check=True)
print('winner probe prepared',a.scenario,a.revision,'reserve31',a.reserve31)
