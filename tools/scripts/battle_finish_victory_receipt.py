"""第94輪正常勝負的獨立重播與來源收據；只能在專案Docker內執行。"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

assert Path.cwd() == Path('/src') and os.getuid() == 1000

def create(path, text):
    with Path(path).open('x') as stream:
        stream.write(text)

if sys.argv[1] == '--prepare-replay':
    parent = Path('tools/scripts/fleet_player_chain_replay.go')
    frozen = json.loads(Path('docs/internals/data/fleet-order-action.json').read_text())
    assert hashlib.sha256(parent.read_bytes()).hexdigest() == frozen['sources'][str(parent)]['sha256']
    game = Path('internal/game/fleetcombat.go')
    extra = r'''
func (g *Game) FleetChainProbeSeed(seed uint32) { g.battleRand = rules.NewRand(seed) }
func (g *Game) FleetFinishProbeReport() map[string]any {
 if !g.battleFinishing || g.state != StateTurnReport || len(g.turnRep.pending)==0 { return nil }
 r:=g.turnRep.pending[0]
 kind:="victory";if g.battleOccupation!=nil {kind="capture"}
 tables:=map[string]string{}
 for k,v:=range map[string][]byte{"fleet":g.slot.Fleet,"ship":g.slot.Ship,"admiral":g.slot.Admiral,"work":g.workRecord(),"stellar":g.res.Stellar,"planet":g.res.Planets.Data} {tables[k]=fmt.Sprintf("%x",v)}
 return map[string]any{"kind":kind,"text_hex":fmt.Sprintf("%x",r.text),"face":r.face,"battle_turn":g.battleTurn,"region":g.battleRegion,"seed":g.battleRand.Seed(),"tables":tables}
}
'''
    create('workplace/finish79-r94-replay-game.go', game.read_text()+extra)
    s = parent.read_text()
    s = s.replace('scenes := 0', 'scenes := 0\n var reports []map[string]any\n lastReport := ""')
    s = s.replace('before := g.State()', '''before := g.State()
        if r:=g.FleetFinishProbeReport();r!=nil {
            b,_:=json.Marshal(r);identity:=string(b)
            if identity!=lastReport {
                lastReport=identity;reports=append(reports,r)
                stem:=filepath.Join(*out,fmt.Sprintf("finish-report-%02d",len(reports)))
                if err=g.Frame().WriteIndex(stem+".bin");err!=nil {panic(err)}
                if err=g.Frame().WritePNG(stem+".png",g.Palette());err!=nil {panic(err)}
            }
        } else {lastReport=""}''')
    s = s.replace('g.State() != game.StateBattle || scenes == 0', 'g.State() != game.StateCommandPhase || scenes == 0 || len(reports)==0')
    s = s.replace('// 直接核對六份正式保存', '''var expected struct { Observations []map[string]any }
    b,err=os.ReadFile(filepath.Join(filepath.Dir(*input),"observations.json"));if err!=nil {panic(err)}
    if err=json.Unmarshal(b,&expected);err!=nil {panic(err)}
    if len(expected.Observations)!=len(reports) {panic("自然報告數不同")}
    for i,r:=range reports {
        delete(expected.Observations[i],"at_ns")
        a,_:=json.Marshal(expected.Observations[i]);b,_:=json.Marshal(r)
        if string(a)!=string(b) {panic("自然報告、六表或seed不同")}
        name:=fmt.Sprintf("finish-report-%02d.bin",i+1)
        a,err=os.ReadFile(filepath.Join(filepath.Dir(*input),name));if err!=nil {panic(err)}
        b,err=os.ReadFile(filepath.Join(*out,name));if err!=nil {panic(err)}
        if string(a)!=string(b) {panic("自然報告畫面不同")}
    }
    // 直接核對六份正式保存''')
    s = s.replace('"six_save_files_equal": true', '"six_save_files_equal": true, "finish_reports":len(reports), "report_tables_seed_pixels_equal":true')
    s = s.replace('正常bootstrap、交戰、續回戰場及六檔保存一致', '正常bootstrap、自然敗退報告、六表與seed、回命令及六檔保存一致')
    create('workplace/finish79-r94-replay.go', s)
    create('workplace/finish79-r94-replay-overlay.json', json.dumps({'Replace':{str(game.resolve()):'/src/workplace/finish79-r94-replay-game.go'}},indent=2)+'\n')
    subprocess.run(['gofmt','-w','workplace/finish79-r94-replay-game.go','workplace/finish79-r94-replay.go'],check=True)
elif sys.argv[1] == '--prepare-gui':
    subprocess.run(['python3','tools/scripts/fleet_player_gui_prepare.py','finish94'],check=True)
    game = Path('workplace/round79-fleet-gui-game-finish94.go')
    report = Path('workplace/finish79-r94-replay-game.go').read_text()
    report = report[report.index('func (g *Game) FleetFinishProbeReport()'):]
    game.write_text(game.read_text()+report)
    main = Path('workplace/round79-fleet-gui-main-finish94.go')
    s=main.read_text()
    assert s.count('err = ebiten.RunGame(a)')==1
    s=s.replace('err = ebiten.RunGame(a)', 'ebiten.SetTPS(10) // 僅測試排程；規則、輸入及renderer保持正式路徑。\nerr = ebiten.RunGame(a)')
    s=s.replace('st := a.g.FleetGUIStamp()', 'st := a.g.FleetGUIStamp()\nst["finish_report"]=a.g.FleetFinishProbeReport()')
    main.write_text(s)
    s=Path('tools/scripts/fleet_player_gui_plan.go').read_text()
    s=s.replace('var actions []map[string]any', 'var actions []map[string]any\nreportSeen:=false')
    s=s.replace('g.Tick(time.Duration(in.NS))', 'g.Tick(time.Duration(in.NS))')
    anchor='g.Frame()\n\t\t\tcontinue'
    assert s.count(anchor)==1
    s=s.replace(anchor, '''g.Frame()
            if r:=g.FleetFinishProbeReport();r!=nil&&!reportSeen {
                reportSeen=true
                name:="finish-report-01"
                if err=g.Frame().WriteIndex(filepath.Join(*out,name+".bin"));err!=nil {panic(err)}
                actions=append(actions,map[string]any{"kind":"wait_report","report":r,"snapshot":name})
            }
            continue''')
    s=s.replace('g.State() != game.StateBattle || !g.FleetGUIStamp()["idle"].(bool)', 'g.State() != game.StateCommandPhase || !reportSeen')
    create('workplace/finish79-r94-gui-plan.go',s)
    s=Path('tools/scripts/fleet_player_gui_driver.py').read_text()
    anchor="        before=wait(lambda s:match(s,action['before']),f'輸入{i}尚未就緒')"
    assert s.count(anchor)==1
    s=s.replace(anchor, '''        if action['kind']=='wait_report':
            st=wait(lambda s:s is not None and s.get('finish_report')==action['report'],'自然敗退報告')
            snapshot(action['snapshot'])
            trace.append({'index':i,'kind':'wait_report','state':st})
            continue
'''+anchor)
    s=s.replace("'sound_device_tested':False", "'sound_device_tested':False,'test_tps':10,'natural_finish':True")
    create('workplace/finish79-r94-gui-driver.py',s)
    subprocess.run(['gofmt','-w',str(game),str(main),'workplace/finish79-r94-gui-plan.go'],check=True)
elif sys.argv[1] in ('--freeze','--check'):
    target=Path('docs/internals/data/battle-finish-natural.json')
    if sys.argv[1]=='--freeze':
        assert not target.exists()
        normal=json.loads(Path('workplace/finish79-r94-normal-c/normal-1/observations.json').read_text())
        gui=json.loads(Path('workplace/finish79-r94-gui-a/receipt.json').read_text())
        assert normal['victory_reports']==1 and normal['capture_reports']==0 and normal['state']=='StateCommandPhase'
        assert gui['natural_finish'] and gui['test_tps']==10 and gui['six_diagnostic_saves_equal']
        assert all(c['x11_different']==0 and c['game_index_equal'] for c in gui['comparisons'])
        for revision in ('a','b'):
            r=json.loads(Path(f'workplace/finish79-r94-bootstrap-{revision}/replay.json').read_text())
            assert r['six_save_files_equal'] and r['report_tables_seed_pixels_equal'] and r['finish_reports']==1
        sources=json.loads(Path('docs/internals/data/fleet-order-action.json').read_text())['sources']
        for name,expected in sources.items():
            assert hashlib.sha256(Path(name).read_bytes()).hexdigest()==expected['sha256'],name
        paths=set(sources)
        paths.add('docs/internals/data/fleet-order-action.json')
        paths.update(('tools/scripts/battle_finish_victory_probe.py','tools/scripts/battle_finish_victory_receipt.py',
            'docs/goal/2026-10-05-battle-finish-natural-round-94.md','docs/re/428-battle-finish-natural.md',
            'docs/playtest/083-battle-finish-natural.md'))
        for root in Path('workplace').glob('finish79-r94*'):
            if root.is_file(): paths.add(str(root))
            elif root.is_dir(): paths.update(str(p) for p in root.rglob('*') if p.is_file())
        paths.update(str(p) for p in Path('workplace').glob('round79-fleet-gui-*-finish94.*') if p.is_file())
        paths.update(('workplace/order79-r93-input','workplace/round79-capture'))
        paths.update(('tools/scripts/sound_tail_input.go','tools/scripts/fleet_player_gui_capture.go'))
        sources={}
        for name in sorted(paths):
            data=Path(name).read_bytes()
            sources[name]={'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest()}
        result={'schema':'logh3.battle-finish-natural.v1','date':'2026-10-05','issue':79,
            'parent_commit':'601476e810a8bd6bd36b5a237f0435714f3036f9',
            'classification':'remake正常玩家自然勝負；原版正常整回合同態未驗',
            'scenario':'S3','faction':1,'initial_seed':0x12345678,'seed_method':'戰場開始後、首次實行前測試overlay設定一次',
            'state_injection':False,'native_new_receipts':False,
            'normal':{'inputs':len(normal['inputs']),'victory_reports':1,'capture_reports':0,
                'battle_turn':9,'orders':50,'scenes':19,'terminal_state':'StateCommandPhase','bootstrap_repeats':2,
                'report_six_tables_bytes':sum(len(bytes.fromhex(x)) for x in normal['observations'][0]['tables'].values()),
                'report_seed':normal['observations'][0]['seed']},
            'gui':{'test_tps':10,'actions':len(gui['actions']),'frames':len(gui['comparisons']),
                'pixels':sum(c['pixels'] for c in gui['comparisons']),'x11_different':0,
                'extra_scroll_gestures':len(gui['camera_gestures']),'six_diagnostic_saves_equal':True,
                'limitations':['正式60Hz長流程未新增驗收','單人CPU沿061停留','未新增占領、自然死亡、到期或升級GUI','不是原版same-state']},
            'sources':sources}
        create(target,json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    v=json.loads(target.read_text())
    for name,expected in v['sources'].items():
        b=Path(name).read_bytes()
        assert len(b)==expected['bytes'] and hashlib.sha256(b).hexdigest()==expected['sha256'],name
    print(len(v['sources']),'來源身份符合，索引SHA-256',hashlib.sha256(target.read_bytes()).hexdigest())
else:
    raise SystemExit('未知模式')
