"""第115輪：從正常公開存檔接續自然功績，不注入升級條件。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('revision')
p.add_argument('--command', choices=['none','move-to-9','occupy13','meet-alliance13','meet-alliance16'], default='meet-alliance16')
a = p.parse_args()
assert a.revision.isascii() and a.revision.isalnum()
parent_path = Path('docs/internals/data/battle-reinforcement.json')
assert hashlib.sha256(parent_path.read_bytes()).hexdigest() == 'ba00445ab34f4f28665122098e690b15fc52db318e36ec19d72b0898611f34cb'
parent = json.loads(parent_path.read_text())

def checked(name):
    path = Path(name)
    data = path.read_bytes()
    expected = parent['sources'][name]
    assert len(data) == expected['bytes'] and hashlib.sha256(data).hexdigest() == expected['sha256'], name
    return data

fixture = 'workplace/reinforce79-r114-replay-d-plan/player-save'
for name in ['ADMIRAL','SHIP','WORK','FLEET','PLANET','STELLAR']:
    checked(fixture+'/'+name+'.DT0')
checked(fixture+'/SAVEDATA.IDX')
s = checked('workplace/reinforce79-r114-m-probe_test.go').decode()
s = s.replace('winner113Observation', 'natural115Observation')
anchor = '\t\t\twithSound := '
assert s.count(anchor) == 1
setup = '''
            fixtureDir:=t.TempDir()
            if out:=os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT");out!="" {fixtureDir=filepath.Join(out,"parent-fixture")}
            if e:=os.MkdirAll(fixtureDir,0755);e!=nil {t.Fatal(e)}
            for _,name:=range append(rules.SlotFileNames(0),"SAVEDATA.IDX") {
                b,e:=os.ReadFile(filepath.Join(root,"FIXTURE",name));if e!=nil {t.Fatal(e)}
                if e=os.WriteFile(filepath.Join(fixtureDir,name),b,0644);e!=nil {t.Fatal(e)}
            }
            // package game的測試不能import bootstrap；沿UseSaveDir同一ParseSaveIndex契約配置fixture。
            indexBytes,e:=os.ReadFile(filepath.Join(fixtureDir,"SAVEDATA.IDX"));if e!=nil {t.Fatal(e)}
            index,e:=rules.ParseSaveIndex(indexBytes);if e!=nil {t.Fatal(e)}
            g.res.SaveDir,g.res.SaveIndex=fixtureDir,index
'''.replace('FIXTURE', fixture)
s = s.replace(anchor, setup+anchor)
s = s.replace('victoryReports, captureReports := 0, 0', 'victoryReports, captureReports := 0, 0\nvar promotionResults []map[string]any\nvar strategicCheckpoints []json.RawMessage\nlastStrategic:=""\npromotionReports:=0\nlastPromotion:=""')
anchor = '\t\t\tobserve := func() {'
assert s.count(anchor) == 1
observe = '''
                if g.state!=StateBattle&&!g.fleetCombat.active {
                    key:=fmt.Sprintf("%d:%s:%v",g.turn,g.state.String(),g.fleets)
                    if key!=lastStrategic {
                        lastStrategic=key
                        b,e:=json.Marshal(map[string]any{"turn":g.turn,"state":g.state.String(),"at_ns":elapsed,"faction":g.faction,"fleets":g.fleets,"military":g.military,"graph":g.res.Graph,"work":fmt.Sprintf("%x",g.workRecord()),"stellar":fmt.Sprintf("%x",g.res.Stellar)})
                        if e!=nil {t.Fatal(e)}
                        strategicCheckpoints=append(strategicCheckpoints,json.RawMessage(b))
                    }
                }
                if g.state==StateTurnReport&&len(g.turnRep.pending)>0 {
                    r:=g.turnRep.pending[0]
                    rec:=g.admiralRec(23)
                    if r.customWait&&r.face==23&&rec[6]>4 {
                        key:=fmt.Sprintf("%d:%d:%d:%x",g.turn,g.battleRegion,g.battleTurn,r.text)
                        if key!=lastPromotion {
                            lastPromotion=key;promotionReports++
                            tables:=map[string]string{}
                            for k,v:=range map[string][]byte{"fleet":g.slot.Fleet,"ship":g.slot.Ship,"admiral":g.slot.Admiral,"work":g.workRecord(),"stellar":g.res.Stellar,"planet":g.res.Planets.Data} {tables[k]=fmt.Sprintf("%x",v)}
                            promotionResults=append(promotionResults,map[string]any{"kind":"promotion","region":g.battleRegion,"battle_turn":g.battleTurn,"turn":g.turn,"at_ns":elapsed,"face":r.face,"text_hex":fmt.Sprintf("%x",r.text),"wait":r.wait,"seed":g.battleRand.Seed(),"tables":tables,"admiral":fmt.Sprintf("%x",rec)})
                            if out:=os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT");out!="" {
                                folder:=filepath.Join(out,fmt.Sprintf("normal-%d",faction));if e:=os.MkdirAll(folder,0755);e!=nil {t.Fatal(e)}
                                stem:=filepath.Join(folder,fmt.Sprintf("promotion-%02d",promotionReports))
                                if e:=g.Frame().WritePNG(stem+".png",g.Palette());e!=nil {t.Fatal(e)}
                                if e:=g.Frame().WriteIndex(stem+".bin");e!=nil {t.Fatal(e)}
                            }
                        }
                    }
                }
'''
s = s.replace(anchor, anchor+observe)
s = s.replace('"strategic_orders": strategicOrders,', '"strategic_orders": strategicOrders, "strategic_checkpoints":strategicCheckpoints, "promotion_reports":promotionReports,"promotions":promotionResults,"parent_fixture":"normal round114 public save",')
start = s.index('\t\t\treinforcement := func() {')
end = s.index('\t\t\tclick(544, 125)', start)
s = s[:start]+'''reinforcement:=func() {
    if reinforcementPhase>=1 {return}
    orders:=[]struct{index,target,action int}{ORDERS}
    for _,o:=range orders {publicOrder(o.index,o.target,o.action)}
    reinforcementPhase++
}
'''.replace('ORDERS',{'none':'','move-to-9':'{7,9,fleetActionMove}','occupy13':'{7,13,fleetActionOccupy}','meet-alliance13':'','meet-alliance16':''}[a.command])+s[end:]
if a.command in ['meet-alliance13','meet-alliance16']:
    start = s.index('reinforcement:=func() {')
    end = s.index('\t\t\tclick(544, 125)',start)
    s = s[:start]+'''
    reinforcement:=func() {
        if reinforcementPhase==0 {publicOrder(7,11,fleetActionMove);reinforcementPhase=1;return}
        if reinforcementPhase==1 {
            f:=g.fleetByIndex(7)
            if f.Region==11 {
                action:=fleetActionOccupy
                if g.res.Planets.RegionOwner(13)==f.Faction {action=fleetActionMove}
                publicOrder(7,13,action);reinforcementPhase=2
            }
        }
    }
    allianceOrdered:=false
    allianceOrder:=func() {
        if !allianceOrdered {publicOrder(16,13,fleetActionMove);allianceOrdered=true}
    }
''' + s[end:]
    if a.command == 'meet-alliance16':
        start=s.index('    reinforcement:=func() {')
        end=s.index('    allianceOrdered:=false',start)
        s=s[:start]+'''
    reinforcement:=func() {
        f:=g.fleetByIndex(7)
        if reinforcementPhase==0 {publicOrder(7,9,fleetActionMove);reinforcementPhase=1;return}
        if reinforcementPhase==1&&f.Region==9 {publicOrder(7,11,fleetActionMove);reinforcementPhase=2;return}
        if reinforcementPhase==2&&f.Region==11 {publicOrder(7,16,fleetActionOccupy);reinforcementPhase=3}
    }
''' +s[end:]
    s = s.replace('f.Index == index && f.Faction == faction', 'f.Index == index && f.Faction == g.faction')
    anchor='click(432, 270)\n\t\t\tclick(544, 125)\n\t\t\tclick(560, 125)'
    assert s.count(anchor)==1
    s=s.replace(anchor,'click(432,270)\nallianceOrder()\n\t\t\tclick(544, 125)\n\t\t\tclick(560, 125)')
    start=s.index('for n := 0; n < 15000 && g.state != StateBattle; n++ {')
    end=s.index('\n\t\t\tseed()',start)
    first=s[start:end]
    anchor='case StateVersusAccount, StateVersusShipOrder:'
    assert first.count(anchor)==1
    first=first.replace(anchor,'''case StateCommandPhase:
        if g.faction!=0 {t.Fatal("下一輪公開命令不是帝國座位",g.faction)}
        reinforcement();click(544,125);click(560,125)
        if g.state!=StateVersusInput||g.faction!=1 {t.Fatal("正常實行未交接同盟",g.state,g.faction)}
        click(432,270)
        for n:=0;n<3&&(g.state==StateVersusAccount||g.state==StateVersusShipOrder);n++ {right(300,240)}
        allianceOrder();click(544,125);click(560,125)
    '''+anchor)
    s=s[:start]+first+s[end:]
s=s.replace('n < 15000 && g.state != StateBattle', 'n < 15000 && g.turn < 107 && g.state != StateBattle')
s=s.replace('g.Click(x, y)\n', 'g.Click(x, y)\nobserve()\n')
s=s.replace('g.RightRelease()\n', 'g.RightRelease()\nobserve()\n')
s=s.replace('公開清單選错艦隊','公開清單選錯艦隊')
assert s.count('click(300, 282)') == 1
s = s.replace('click(300, 282)', 'click(300, 122)')
s = s.replace('正常S3載入失敗', '正常第114輪公開存檔載入失敗')
s = s.replace('turn < 108 && (captureReports == 0 || g.state != StateCommandPhase)', 'turn < 36 && (promotionReports == 0 || g.state != StateCommandPhase)')
s = s.replace('if captureReports > 0 {', 'if promotionReports > 0 {')
s = s.replace('t.Fatal("正常戰場沒有敵方")', 'continue // 敵方只有行星時，不虛構艦隊或戰鬥命令。')
s = s.replace('if captureReports == 0 || scenes == 0 || orders == 0 {', 'if promotionReports == 0 || g.admiralRec(23)[6]!=5 {')
s = s.replace('正常UI未到占領報告', '正常UI未到自然副官升級')
s = s.replace('"capture_reports": captureReports, "no_state_injection":', '"capture_reports": captureReports, "promotion_reports":promotionReports,"no_state_injection":')
marker='\t\t\tsaveDir := t.TempDir()'
assert s.count(marker)==1
s=s.replace(marker, '''
            // 由正常地圖／清單打開已升級艦隊的正式軍階欄位，不注入UI狀態。
            x,y:=regionPoint(g.fleetByIndex(7).Region);click(x,y)
            click(350,290)
            fi:=-1;for i,f:=range g.fleets {if f.Index==7 {fi=i;break}}
            row:=g.fleetListRowOf(fi);if row<0||row>=5 {t.Fatal("升級軍階艦隊不在公開首頁")}
            click(450,308+20*row)
            if g.state!=StateFleetWindow||g.FleetCommandFleet()!=fi {t.Fatal("升級軍階未到正式艦隊視窗")}
            if out:=os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT");out!="" {
                folder:=filepath.Join(out,fmt.Sprintf("normal-%d",faction))
                if e:=g.Frame().WritePNG(filepath.Join(folder,"rank.png"),g.Palette());e!=nil {t.Fatal(e)}
                if e:=g.Frame().WriteIndex(filepath.Join(folder,"rank.bin"));e!=nil {t.Fatal(e)}
            }
            click(490,390)
            if g.state!=StateCommandPhase {t.Fatal("軍階視窗未返回命令")}
''' +marker)
stem = 'workplace/promotion79-r115-'+a.revision
test = Path(stem+'-probe_test.go')
observed = Path(stem+'-observed-battlefinish.go')
overlay = Path(stem+'-overlay.json')
for path in [test,observed,overlay]:
    assert not path.exists()
    assert path.parent.stat().st_uid == 1000 and path.parent.stat().st_gid == 1000
with test.open('x') as f:
    f.write(s)
with observed.open('x') as f:
    f.write(checked('workplace/reinforce79-r114-m-observed-battlefinish.go').decode().replace('winner113Observation','natural115Observation'))
with overlay.open('x') as f:
    json.dump({'Replace':{'/src/internal/game/finish_player_probe_test.go':str(Path('/src')/test),'/src/internal/game/battlefinish.go':str(Path('/src')/observed)}},f,indent=2)
    f.write('\n')
subprocess.run(['gofmt','-w',str(test),str(observed)],check=True)
print('prepared',stem,'normal public fixture; no merit injection')
