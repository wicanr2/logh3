"""第95輪：正常玩家逐擊、戰術報告與公開存讀；Docker內執行。"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

assert Path.cwd() == Path('/src') and os.getuid() == 1000
parent=json.loads(Path('docs/internals/data/battle-finish-natural.json').read_text())

def read_parent(name):
    p=Path(name)
    s=p.read_text()
    assert hashlib.sha256(p.read_bytes()).hexdigest()==parent['sources'][name]['sha256'],name
    return s

def create(name,text):
    with Path(name).open('x') as stream: stream.write(text)

extra=r'''
// 僅測試唯讀資料，不改Game、RND、鏡頭或命令。
func (g *Game) BattleLifecycleProbeTables() map[string]string {
 tables:=map[string]string{}
 for k,v:=range map[string][]byte{"fleet":g.slot.Fleet,"ship":g.slot.Ship,"admiral":g.slot.Admiral,"work":g.workRecord(),"stellar":g.res.Stellar,"planet":g.res.Planets.Data} {tables[k]=fmt.Sprintf("%x",v)}
 return tables
}
func (g *Game) BattleLifecycleProbeReport() map[string]any {
 if g.state!=StateTurnReport || !g.turnRep.battle || len(g.turnRep.pending)==0 {return nil}
 r:=g.turnRep.pending[0]
 kind:="battle-report"
 if g.battleFinishing {kind="victory";if g.battleOccupation!=nil {kind="capture"}}
 for _,bf:=range g.battleFleets {
  f:=g.fleetByIndex(bf.Index);if f==nil {continue}
  if string(r.text)==string(g.battleAnnihilationReport(f).text) {kind="annihilation"}
  if string(r.text)==string(g.scoutLostReport(f).text) {kind="scout-expired"}
 }
 return map[string]any{"kind":kind,"text_hex":fmt.Sprintf("%x",r.text),"face":r.face,"battle_turn":g.battleTurn,"region":g.battleRegion,"seed":g.battleRand.Seed(),"tables":g.BattleLifecycleProbeTables(),"units":append([]rules.BattleUnit(nil),g.battleUnits...)}
}
func (g *Game) BattleLifecycleProbeScene() map[string]any {
 c:=g.fleetCombat.scene
 if !g.fleetCombat.active||c==nil||g.state!=StateEngagement {return nil}
 return map[string]any{"key":fmt.Sprintf("%d:%d:%d",g.battleRegion,g.battleTurn,g.fleetCombat.commandAt),"damaged":c.damaged,
 "own":c.own,"target":c.target,"weapon":c.weapon,"hit":c.hit,"before":c.before,
 "units":append([]rules.BattleUnit(nil),g.battleUnits...),"seed":g.battleRand.Seed(),"tables":g.BattleLifecycleProbeTables()}
}
'''

if sys.argv[1]=='--prepare-probe':
    revision=sys.argv[2]
    assert revision.isascii() and revision.isalnum()
    s=read_parent('workplace/finish79-r94-c-probe_test.go')
    s=s.replace('TestBattleFinishNormalProbe','TestBattleLifecycleNormalProbe')
    anchor='elapsed := int64(0)'
    assert s.count(anchor)==1
    s=s.replace(anchor,anchor+r'''
            var lifecycle []map[string]any
            var lifeReports []map[string]any
            lifeSceneKey,lifeReportKey:="",""
            observeLife:=func(){
                if r:=g.BattleLifecycleProbeReport();r!=nil {
                    key:=fmt.Sprintf("%v:%v:%v",r["battle_turn"],r["face"],r["text_hex"])
                    if key!=lifeReportKey {
                        lifeReportKey=key;lifeReports=append(lifeReports,r)
                        if out:=os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT");out!="" {
                            dir:=filepath.Join(out,fmt.Sprintf("normal-%d",faction));if e:=os.MkdirAll(dir,0755);e!=nil {t.Fatal(e)}
                            stem:=filepath.Join(dir,fmt.Sprintf("lifecycle-report-%02d",len(lifeReports)))
                            if e:=g.Frame().WritePNG(stem+".png",g.Palette());e!=nil {t.Fatal(e)}
                            if e:=g.Frame().WriteIndex(stem+".bin");e!=nil {t.Fatal(e)}
                        }
                    }
                } else {lifeReportKey=""}
                if r:=g.BattleLifecycleProbeScene();r!=nil {
                    key:=fmt.Sprintf("%v:%v",r["key"],r["damaged"])
                    if key!=lifeSceneKey {lifeSceneKey=key;lifecycle=append(lifecycle,r)}
                }
            }
            defer func(){
                if out:=os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT");out!="" {
                    dir:=filepath.Join(out,fmt.Sprintf("normal-%d",faction));if e:=os.MkdirAll(dir,0755);e!=nil {t.Fatal(e)}
                    b,e:=json.MarshalIndent(map[string]any{"classification":"正常玩家Game唯讀逐擊與報告；獨立bootstrap及GUI另驗","reports":lifeReports,"scenes":lifecycle},"","  ");if e!=nil {t.Fatal(e)}
                    if e=os.WriteFile(filepath.Join(dir,"lifecycle.json"),append(b,'\n'),0644);e!=nil {t.Fatal(e)}
                }
            }()
''')
    anchor='observe()\n\t\t\t\tobserveCommands()'
    assert s.count(anchor)==1
    s=s.replace(anchor,'observeLife()\n'+anchor)
    anchor='g.Tick(d)'
    assert s.count(anchor)==1
    s=s.replace(anchor,anchor+'\nobserveLife()')
    anchor='if err := g.Frame().WritePNG(filepath.Join(saveDir, "screen.png"), g.Palette()); err != nil {'
    assert s.count(anchor)==1
    s=s.replace(anchor,r'''
                // 玩家存檔目錄是執行設定，不是局面注入。
                g.res.SaveDir=filepath.Join(saveDir,"player-save")
                if e:=os.MkdirAll(g.res.SaveDir,0755);e!=nil {t.Fatal(e)}
                click(544,125);click(560,165)
                if g.state!=StateSaveGame {t.Fatal("戰後公開存進度未開")}
                click(300,g.res.SlotList.Save.TextTop)
                if g.saveErr!=nil {t.Fatal(g.saveErr)}
                right(300,142)
                if g.state!=StateCommandPhase {t.Fatal("存進度未正常返回")}
                click(544,125);click(560,145);click(300,122)
                if g.state!=StateLoadFaction {t.Fatal("戰後公開取進度未選陣營")}
                click(318,208+20*faction)
                if g.state!=StateCommandPhase {t.Fatal("取進度未回命令畫面")}
'''+anchor)
    s+=extra
    dest=f'workplace/lifecycle79-r95-{revision}-probe_test.go'
    create(dest,s)
    create(f'workplace/lifecycle79-r95-{revision}-probe-overlay.json',json.dumps({'Replace':{'/src/internal/game/lifecycle_player_probe_test.go':str(Path(dest).resolve())}},indent=2)+'\n')
    subprocess.run(['gofmt','-w',dest],check=True)
elif sys.argv[1]=='--prepare-replay':
    revision=sys.argv[2]
    assert revision.isascii() and revision.isalnum()
    stem=f'workplace/lifecycle79-r95-replay-{revision}'
    s=read_parent('workplace/finish79-r94-replay.go')
    s=s.replace('import (','import (\n"bytes"',1)
    game=Path('internal/game/fleetcombat.go')
    base=read_parent('workplace/finish79-r94-replay-game.go')
    create(stem+'-game.go',base+extra)
    anchor='g, err := game.New(r)'
    assert s.count(anchor)==1
    s=s.replace(anchor,'if err=bootstrap.UseSaveDir(r,filepath.Join(*out,"player-save"));err!=nil {panic(err)}\n'+anchor)
    anchor='scenes := 0'
    assert s.count(anchor)==1
    s=s.replace(anchor,anchor+r'''
    var lifeReports,lifeScenes []map[string]any
    lifeReportKey,lifeSceneKey:="",""
    observeLife:=func(){
        if r:=g.BattleLifecycleProbeReport();r!=nil {
            key:=fmt.Sprintf("%v:%v:%v",r["battle_turn"],r["face"],r["text_hex"])
            if key!=lifeReportKey {
                lifeReportKey=key;lifeReports=append(lifeReports,r)
                stem:=filepath.Join(*out,fmt.Sprintf("lifecycle-report-%02d",len(lifeReports)))
                if err=g.Frame().WriteIndex(stem+".bin");err!=nil {panic(err)}
                if err=g.Frame().WritePNG(stem+".png",g.Palette());err!=nil {panic(err)}
            }
        } else {lifeReportKey=""}
        if r:=g.BattleLifecycleProbeScene();r!=nil {
            key:=fmt.Sprintf("%v:%v",r["key"],r["damaged"])
            if key!=lifeSceneKey {lifeSceneKey=key;lifeScenes=append(lifeScenes,r)}
        }
    }
''')
    s=s.replace('before := g.State()','before := g.State()\nobserveLife()',1)
    anchor='if g.State() == game.StateEngagement && before != game.StateEngagement {'
    assert s.count(anchor)==1
    s=s.replace(anchor,'observeLife()\n'+anchor)
    anchor='var expected struct{ Observations []map[string]any }'
    assert s.count(anchor)==1
    s=s.replace(anchor,r'''
    observed,_:=json.MarshalIndent(map[string]any{"reports":lifeReports,"scenes":lifeScenes},"","  ")
    if err=os.WriteFile(filepath.Join(*out,"lifecycle-observed.json"),append(observed,'\n'),0644);err!=nil {panic(err)}
    var lifeExpected struct {Reports,Scenes []map[string]any}
    b,err=os.ReadFile(filepath.Join(filepath.Dir(*input),"lifecycle.json"));if err!=nil {panic(err)}
    decoder:=json.NewDecoder(bytes.NewReader(b));decoder.UseNumber()
    if err=decoder.Decode(&lifeExpected);err!=nil {panic(err)}
    for group,rows:=range map[string][2][]map[string]any{"reports":{lifeExpected.Reports,lifeReports},"scenes":{lifeExpected.Scenes,lifeScenes}} {
        if len(rows[0])!=len(rows[1]) {panic(fmt.Sprintf("%s觀測數不同 %d/%d",group,len(rows[0]),len(rows[1])))}
        for i:=range rows[0] {
            if canonicalLifecycle(rows[0][i])!=canonicalLifecycle(rows[1][i]) {
                var keys []string
                for k,v:=range rows[0][i] {if canonicalLifecycle(v)!=canonicalLifecycle(rows[1][i][k]) {keys=append(keys,k)}}
                panic(fmt.Sprintf("%s第%d筆不同：%v",group,i,keys))
            }
        }
    }
    for i:=range lifeReports {
        name:=fmt.Sprintf("lifecycle-report-%02d.bin",i+1)
        a,e:=os.ReadFile(filepath.Join(filepath.Dir(*input),name));if e!=nil {panic(e)}
        b,e:=os.ReadFile(filepath.Join(*out,name));if e!=nil {panic(e)}
        if string(a)!=string(b) {panic("結算報告畫面不同")}
    }
    for _,name:=range rules.SlotFileNames(0) {
        a,e:=os.ReadFile(filepath.Join(filepath.Dir(*input),"player-save",name));if e!=nil {panic(e)}
        b,e:=os.ReadFile(filepath.Join(*out,"player-save",name));if e!=nil {panic(e)}
        if string(a)!=string(b) {panic("公開保存不同"+name)}
    }
    lifeData,_:=json.MarshalIndent(map[string]any{"reports":lifeReports,"scenes":lifeScenes},"","  ")
    if err=os.WriteFile(filepath.Join(*out,"lifecycle.json"),append(lifeData,'\n'),0644);err!=nil {panic(err)}
    var expected struct{ Observations []map[string]any }''')
    s=s.replace('"six_save_files_equal": true','"six_save_files_equal": true, "lifecycle_reports":len(lifeReports), "scene_observations":len(lifeScenes), "public_save_load_equal":true')
    s+=r'''
// 只正規化JSON物件的欄位順序；數字保留原精度，不刪欄位或容忍數值差。
func canonicalLifecycle(v any) string {
 b,err:=json.Marshal(v);if err!=nil {panic(err)}
 d:=json.NewDecoder(bytes.NewReader(b));d.UseNumber()
 var normalized any;if err=d.Decode(&normalized);err!=nil {panic(err)}
 b,err=json.Marshal(normalized);if err!=nil {panic(err)}
 return string(b)
}
'''
    create(stem+'.go',s)
    create(stem+'-overlay.json',json.dumps({'Replace':{str(game.resolve()):str(Path(stem+'-game.go').resolve())}},indent=2)+'\n')
    subprocess.run(['gofmt','-w',stem+'.go',stem+'-game.go'],check=True)
elif sys.argv[1]=='--prepare-gui':
    game=read_parent('workplace/round79-fleet-gui-game-finish94.go')
    create('workplace/lifecycle79-r95-gui-game.go',game+extra)
    main=read_parent('workplace/round79-fleet-gui-main-finish94.go')
    anchor='st["finish_report"] = a.g.FleetFinishProbeReport()'
    assert main.count(anchor)==1
    main=main.replace(anchor,anchor+'\nst["lifecycle_report"]=a.g.BattleLifecycleProbeReport()\nst["save_error"]=fmt.Sprint(a.g.SaveErr())')
    create('workplace/lifecycle79-r95-gui-main.go',main)
    create('workplace/lifecycle79-r95-gui-overlay.json',json.dumps({'Replace':{
        '/src/internal/game/fleetcombat.go':'/src/workplace/lifecycle79-r95-gui-game.go',
        '/src/cmd/gin3/main.go':'/src/workplace/lifecycle79-r95-gui-main.go'}},indent=2)+'\n')
    s=read_parent('workplace/finish79-r94-gui-plan.go')
    s=s.replace('g, err := game.New(r)','if err=bootstrap.UseSaveDir(r,filepath.Join(*out,"player-save"));err!=nil {panic(err)}\ng, err := game.New(r)',1)
    s=s.replace('reportSeen := false','lifeReportKey:=""\nvar lifeReports []map[string]any',1)
    start=s.index('if r := g.FleetFinishProbeReport();')
    end=s.index('\n\t\t\tcontinue',start)
    s=s[:start]+r'''
            if r:=g.BattleLifecycleProbeReport();r!=nil {
                key:=fmt.Sprintf("%v:%v:%v",r["battle_turn"],r["face"],r["text_hex"])
                if key!=lifeReportKey {
                    lifeReportKey=key;lifeReports=append(lifeReports,r)
                    name:=fmt.Sprintf("lifecycle-report-%02d",len(lifeReports))
                    if err=g.Frame().WriteIndex(filepath.Join(*out,name+".bin"));err!=nil {panic(err)}
                    if err=g.Frame().WritePNG(filepath.Join(*out,name+".png"),g.Palette());err!=nil {panic(err)}
                    actions=append(actions,map[string]any{"kind":"wait_report","report":r,"snapshot":name})
                }
            } else {lifeReportKey=""}
'''+s[end:]
    s=s.replace('!reportSeen','len(lifeReports)!=2',1)
    anchor='points := map[string][2]int{}'
    assert s.count(anchor)==1
    s=s.replace(anchor,r'''
    for _,name:=range rules.SlotFileNames(0) {
        a,e:=os.ReadFile(filepath.Join(*out,"player-save",name));if e!=nil {panic(e)}
        b,e:=os.ReadFile(filepath.Join(filepath.Dir(*input),"player-save",name));if e!=nil {panic(e)}
        if string(a)!=string(b) {panic("正常GUI計畫公開保存不同"+name)}
    }
'''+anchor)
    s=s.replace('"six_saves_equal_parent": true','"six_saves_equal_parent": true, "public_six_saves_equal_parent":true, "lifecycle_reports":lifeReports')
    create('workplace/lifecycle79-r95-gui-plan.go',s)
    s=read_parent('workplace/finish79-r94-gui-driver.py')
    s=s.replace("s.get('finish_report')","s.get('lifecycle_report')")
    s=s.replace('自然敗退報告','戰術結算報告')
    anchor="    assert len(list(out.glob('*.DT0')))==6"
    assert s.count(anchor)==1
    s=s.replace(anchor,anchor+'''
    public=list((planpath.parent/'player-save').glob('*.DT0'))
    assert len(public)==6
    for p in public:
        assert (out/'player-save'/p.name).read_bytes()==p.read_bytes(),'GUI公開保存不同 '+p.name
    assert status()['save_error']=='<nil>'
''')
    s=s.replace("'natural_finish':True","'natural_finish':True,'natural_annihilation':True,'public_save_load_equal':True")
    create('workplace/lifecycle79-r95-gui-driver.py',s)
    subprocess.run(['gofmt','-w','workplace/lifecycle79-r95-gui-game.go','workplace/lifecycle79-r95-gui-main.go','workplace/lifecycle79-r95-gui-plan.go'],check=True)
elif sys.argv[1] in ('--freeze','--check'):
    target=Path('docs/internals/data/battle-lifecycle-player.json')
    if sys.argv[1]=='--freeze':
        assert not target.exists()
        normal=json.loads(Path('workplace/lifecycle79-r95-normal-b/normal-1/lifecycle.json').read_text())
        assert [r['kind'] for r in normal['reports']]==['annihilation','victory']
        assert len(normal['scenes'])==38 and sum(r['damaged'] for r in normal['scenes'])==19
        gui=json.loads(Path('workplace/lifecycle79-r95-gui-a/receipt.json').read_text())
        assert gui['natural_annihilation'] and gui['public_save_load_equal'] and gui['test_tps']==10
        assert all(c['x11_different']==0 and c['game_index_equal'] for c in gui['comparisons'])
        for revision in ('d','e'):
            replay=json.loads(Path(f'workplace/lifecycle79-r95-bootstrap-{revision}/replay.json').read_text())
            assert replay['scene_observations']==38 and replay['lifecycle_reports']==2 and replay['public_save_load_equal']
        for folder in ('workplace/lifecycle79-r95-normal-b/normal-1','workplace/lifecycle79-r95-bootstrap-d',
                       'workplace/lifecycle79-r95-bootstrap-e','workplace/lifecycle79-r95-gui-plan','workplace/lifecycle79-r95-gui-a'):
            p=Path(folder)
            files=list((p/'player-save').glob('*.DT0'));assert len(files)==6,folder
            assert all(f.read_bytes()==(p/f.name).read_bytes() for f in files),folder
        for name,expected in parent['sources'].items():
            assert hashlib.sha256(Path(name).read_bytes()).hexdigest()==expected['sha256'],name
        paths=set(parent['sources'])
        paths.update(('docs/internals/data/battle-finish-natural.json','tools/scripts/battle_lifecycle_player_receipt.py',
            'docs/goal/2026-10-05-battle-lifecycle-player-round-95.md','docs/re/429-battle-lifecycle-player.md',
            'docs/playtest/084-battle-lifecycle-player.md'))
        for p in Path('workplace').glob('lifecycle79-r95*'):
            if p.is_file(): paths.add(str(p))
            elif p.is_dir(): paths.update(str(f) for f in p.rglob('*') if f.is_file())
        sources={}
        for name in sorted(paths):
            data=Path(name).read_bytes();sources[name]={'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest()}
        frames=gui['comparisons']
        result={'schema':'logh3.battle-lifecycle-player.v1','date':'2026-10-05','issue':79,
            'parent_commit':'1bd9dfc1dfd26ee4e2b5d575a67f2f63f0149ec4',
            'classification':'remake正常玩家自然旗艦全滅、敗退及公開存讀；原版正常整回合同態未驗',
            'scenario':'S3','faction':1,'initial_seed':0x12345678,'seed_method':'首次戰場實行前設定一次',
            'normal':{'inputs':6745,'orders':50,'scenes':19,'scene_observations':38,'reports':2,
                'report_kinds':['annihilation','victory'],'table_bytes_compared_per_replay':1784960,
                'bootstrap_repeats':2,'public_save_load_equal':True,'six_storage_bytes':44560,
                'flagship_loss':{'fleet':12,'squadron':0,'ships_before':6,'ships_after':0,'requested':26,'lost':6},
                'admirals_missing':[12,28],'battle_turn':9,'terminal_state':'StateCommandPhase'},
            'gui':{'test_tps':10,'actions':len(gui['actions']),'frames':len(frames),'pixels':sum(f['pixels'] for f in frames),
                'x11_different':0,'extra_scroll_gestures':len(gui['camera_gestures']),'public_save_load_equal':True},
            'limitations':['不是提督死亡判定','未補其他分隊自然死亡或空組搬移GUI','未新增索敵燃料到期、升級或占領GUI',
                '正式60Hz長流程未新增驗收','原版單擊與回收契約沿既有收據；本輪無原版正常玩家收據'],
            'sources':sources}
        create(target,json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    v=json.loads(target.read_text())
    for name,expected in v['sources'].items():
        data=Path(name).read_bytes()
        assert len(data)==expected['bytes'] and hashlib.sha256(data).hexdigest()==expected['sha256'],name
    print(len(v['sources']),'來源身份符合，索引SHA-256',hashlib.sha256(target.read_bytes()).hexdigest())
else:
    raise SystemExit('未知模式')
