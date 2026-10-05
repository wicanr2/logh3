"""第116輪：正常自然空組清理的正式入口與60 Hz GUI重播。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('revision')
p.add_argument('--input', default='workplace/cleanup79-r116-c/normal-0/inputs.json')
p.add_argument('--gui', action='store_true')
p.add_argument('--snapshot-timing', action='store_true')
a = p.parse_args()
assert a.revision.isascii() and a.revision.isalnum()
parent = json.loads(Path('docs/internals/data/battle-natural-promotion.json').read_text())
source = Path('workplace/promotion79-r115-replay-b-outer.py')
assert hashlib.sha256(source.read_bytes()).hexdigest() == parent['sources'][str(source)]['sha256']
inputs = json.loads(Path(a.input).read_text())
assert inputs['cleanup_events'] > 0
assert sum(v['kind']=='seed' for v in inputs['inputs']) == 1
if a.snapshot_timing:
    assert not a.gui
    stem='workplace/cleanup79-r116-replay-'+a.revision+'-snapshot'
    plan=Path('workplace/cleanup79-r116-replay-b-plan.go').read_text()
    marker='actions = append(actions, a)\n\t\tcollectCleanup()'
    assert plan.count(marker)==1
    plan=plan.replace(marker,marker+r'''
        if kind=="right" && g.State()==game.StateBattle && len(g.BattleCleanupProbeEvents())==2 {
            st:=g.FleetGUIStamp()
            view:=st["view"].([]int)
            if st["turn"].(int)==99 && st["battle_turn"].(int)==4 && len(view)==2 && view[0]==40 && view[1]==47 {
                if err=g.Frame().WritePNG(filepath.Join(*out,"cleanup-first-right.png"),g.Palette());err!=nil {panic(err)}
                if err=g.Frame().WriteIndex(filepath.Join(*out,"cleanup-first-right.bin"));err!=nil {panic(err)}
            }
        }
    ''')
    path=Path(stem+'-plan.go')
    assert not path.exists() and path.parent.stat().st_uid==path.parent.stat().st_gid==1000
    with path.open('x') as f:f.write(plan)
    subprocess.run(['go','run','-overlay','workplace/cleanup79-r116-replay-b-overlay.json',str(path),
                    '-input',a.input,'-out',stem+'-plan'],check=True)
    raise SystemExit
s = source.read_text().replace('promotion79-r115-replay-', 'cleanup79-r116-replay-')
marker = "    create(stem + '-game.go', game)"
assert s.count(marker) == 1
s = s.replace(marker, r'''
    observer=Path('tools/scripts/battle_natural_cleanup_observer.go').read_text()
    observer=observer[observer.index('// natural116CleanupEvents'):]
    game=game.replace('import (','import (\nejson "encoding/json"',1)+'\n'+observer
''' + marker)
marker = "    create(stem + '-main.go', main)"
assert s.count(marker) == 1
s = s.replace(marker, r'''
    marker='if req.Save {'
    assert main.count(marker)==1
    main=main.replace(marker,marker+r"""
    b,e:=json.MarshalIndent(a.g.BattleCleanupProbeEvents(),"","  ");if e!=nil {panic(e)}
    if e=os.WriteFile(filepath.Join(dir,"cleanup-events.json"),append(b,'\n'),0644);e!=nil {panic(e)}
    """)
    main=main.replace('st["lifecycle_report"] = a.g.BattleLifecycleProbeReport()',
                      'st["lifecycle_report"] = a.g.BattleLifecycleProbeReport()\nst["cleanup_event_records"]=len(a.g.BattleCleanupProbeEvents())')
''' + marker)
marker = "    create(stem + '-plan.go', plan)"
assert s.count(marker) == 1
s = s.replace(marker, r'''
    plan=plan.replace('Faction int','Faction int\nExpectedCleanups int `json:"cleanup_events"`',1)
    marker='step := func(kind string'
    assert plan.count(marker)==1
    plan=plan.replace(marker,r"""
    observedCleanup:=0
    collectCleanup:=func() {
        events:=g.BattleCleanupProbeEvents()
        st:=g.FleetGUIStamp()
        if len(events)<=observedCleanup||g.State()!=game.StateBattle||!st["idle"].(bool) {return}
        observedCleanup=len(events)
        name:=fmt.Sprintf("cleanup-next-input-%02d",len(events)/2)
        if err=g.Frame().WritePNG(filepath.Join(*out,name+".png"),g.Palette());err!=nil {panic(err)}
        if err=g.Frame().WriteIndex(filepath.Join(*out,name+".bin"));err!=nil {panic(err)}
        actions=append(actions,map[string]any{"kind":"wait_state","after":g.FleetGUIStamp(),"snapshot":name})
    }
    """+marker)
    marker='actions = append(actions, a)'
    assert plan.count(marker)==1
    plan=plan.replace(marker,marker+'\ncollectCleanup()')
    marker='g.Tick(time.Duration(in.NS))'
    assert plan.count(marker)==1
    plan=plan.replace(marker,marker+'\ncollectCleanup()')
    marker='terminal,e:=json.MarshalIndent'
    assert plan.count(marker)==1
    plan=plan.replace(marker,r"""
    events:=g.BattleCleanupProbeEvents()
    if len(events)!=parent.ExpectedCleanups*2||observedCleanup!=len(events) {panic("正常清理事件與下一回合數不同")}
    data,e:=json.MarshalIndent(events,"","  ");if e!=nil {panic(e)}
    if e=os.WriteFile(filepath.Join(*out,"cleanup-events.json"),append(data,'\n'),0644);e!=nil {panic(e)}
    """+marker)
    cleanup=Path('workplace/cleanup79-r116-c-observed-battlecleanup.go').read_text()
    create(stem+'-cleanup.go',cleanup)
''' + marker)
marker = "'/src/cmd/gin3/main.go': '/src/' + stem + '-main.go',"
assert s.count(marker) == 1
s = s.replace(marker, marker+"\n'/src/internal/game/battlecleanup.go':'/src/'+stem+'-cleanup.go',")
marker = "    create(stem + '-driver.py', driver)"
assert s.count(marker) == 1
s = s.replace(marker, r'''
    driver=driver.replace('正常主選單、S1存檔接續、真實X11鍵鼠、自然功績升級；RND test-only，公開存讀',
                          '正常主選單、S1存檔接續、真實X11鍵鼠、一般分隊自然死亡與空組搬移；RND test-only，公開存讀')
    driver=driver.replace("'natural_promotion':True", "'natural_promotion':True,'natural_nonflagship_cleanup':True")
    # 沿用第81輪的XTEST工具。由送出按下的同一程序讀到正式Update後
    # 立即放開，避開Python往返造成的額外RightHold。
    old="""        send_input({'Type':4,'Button':3})
        try:
            down=wait(lambda s:s is not None and s['edges'][1]>=edge,'鏡頭按下',5)
        finally:send_input({'Type':5,'Button':3})"""
    assert driver.count(old)==1
    driver=driver.replace(old,"        send_input({'Type':4,'Button':3,'Status':str(out/'status.json'),'Edge':edge})")
    subprocess.run(['go','build','-o',stem+'-input','tools/scripts/sound_tail_input.go'],check=True)
''' + marker)
s=s.replace("'/src/workplace/order79-r93-input'", "'/src/'+stem+'-input'")
generator = Path('workplace/cleanup79-r116-replay-'+a.revision+('-gui-outer.py' if a.gui else '-outer.py'))
assert not generator.exists() and generator.parent.stat().st_uid == generator.parent.stat().st_gid == 1000
with generator.open('x') as f:
    f.write(s)
args = ['python3',str(generator),a.revision,'--input',a.input]
if a.gui:
    args.append('--gui')
subprocess.run(args,check=True)
