"""第116輪：公開攻擊一般分隊，只讀觀測自然空組清理。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('revision')
p.add_argument('--target', choices=['weakest','middle'], default='weakest')
a = p.parse_args()
assert a.revision.isascii() and a.revision.isalnum()
parent_path = Path('docs/internals/data/battle-natural-promotion.json')
assert hashlib.sha256(parent_path.read_bytes()).hexdigest() == '0790049faea28eb077b82c05ce8cf3a947f65e62d6b9bad564762e972e2e0824'
parent = json.loads(parent_path.read_text())


def checked(name):
    b = Path(name).read_bytes()
    assert len(b) == parent['sources'][name]['bytes']
    assert hashlib.sha256(b).hexdigest() == parent['sources'][name]['sha256'], name
    return b.decode()


s = checked('workplace/promotion79-r115-h-probe_test.go').replace('natural115Observation', 'natural116Observation')
s = s.replace('var natural116Observation func(*Game, *rules.WinnerCapture)',
              'var natural116Observation func(*Game, *rules.WinnerCapture)\nvar natural116CleanupObservation func(*Game,*rules.Slot,int,string,[]int)')
marker = '\t\t\tvar strategicOrders []map[string]any'
assert s.count(marker) == 1
s = s.replace(marker, '''
            var cleanupResults []json.RawMessage
            cleanupEvents:=0
            cleanupSeen:=map[string]bool{}
            natural116CleanupObservation=func(observed *Game,slot *rules.Slot,index int,stage string,mapping []int) {
                key:=fmt.Sprintf("%d:%d:%d:%d",observed.turn,observed.battleRegion,observed.battleTurn,index)
                if stage=="before" {
                    dead:=false;flagship:=false
                    for _,u:=range observed.battleUnits {if u.Fleet==index {if u.Squadron>0&&u.Ships==0 {dead=true};if u.Squadron==0&&u.Ships>0 {flagship=true}}}
                    if !dead||!flagship {return}
                    if cleanupSeen[key] {t.Fatal("同一正常空組事件重複",key)}
                    cleanupSeen[key]=true;cleanupEvents++
                } else if !cleanupSeen[key] {return}
                tables:=map[string]string{}
                for k,v:=range map[string][]byte{"fleet":slot.Fleet,"ship":slot.Ship,"admiral":slot.Admiral,"work":observed.workRecord(),"stellar":observed.res.Stellar,"planet":observed.res.Planets.Data} {tables[k]=fmt.Sprintf("%x",v)}
                b,e:=json.Marshal(map[string]any{"key":key,"stage":stage,"index":index,"turn":observed.turn,"region":observed.battleRegion,"battle_turn":observed.battleTurn,"at_ns":elapsed,"seed":observed.battleRand.Seed(),"mapping":mapping,"tables":tables,"units":observed.battleUnits,"fleets":observed.fleets})
                if e!=nil {t.Fatal(e)}
                cleanupResults=append(cleanupResults,json.RawMessage(b))
            }
            t.Cleanup(func(){natural116CleanupObservation=nil})
''' + marker)
s = s.replace('"promotion_reports": promotionReports,', '"promotion_reports": promotionReports,"cleanup_events":cleanupEvents,"cleanup_results":cleanupResults,')
s = s.replace('(promotionReports == 0 || g.state != StateCommandPhase)', '(cleanupEvents == 0 || g.state != StateCommandPhase)')
s = s.replace('if promotionReports > 0 {', 'if cleanupEvents > 0 {')
s = s.replace('if promotionReports == 0 || g.admiralRec(23)[6] != 5 {', 'if cleanupEvents == 0 {')
s = s.replace('正常UI未到自然副官升級', '正常UI未產生一般空組清理')
marker='\t\t\tobserve := func() {'
assert s.count(marker)==1
s=s.replace(marker, '''
            lastCleanupFrame:=0
''' +marker+'''
                if cleanupEvents>lastCleanupFrame&&g.state==StateBattle&&g.battleInputIdle() {
                    lastCleanupFrame=cleanupEvents
                    if out:=os.Getenv("LOGH3_FLEET_CHAIN_OUTPUT");out!="" {
                        folder:=filepath.Join(out,fmt.Sprintf("normal-%d",faction))
                        stem:=filepath.Join(folder,fmt.Sprintf("cleanup-next-input-%02d",cleanupEvents))
                        if e:=g.Frame().WritePNG(stem+".png",g.Palette());e!=nil {t.Fatal(e)}
                        if e:=g.Frame().WriteIndex(stem+".bin");e!=nil {t.Fatal(e)}
                    }
                }
''')
start = s.index('\t\t\t\t\t\t\tchosen := false')
end = s.index('\n\t\t\t\t\t\t\tif !chosen {', start)
s = s[:start]+'''
                            chosen:=false
                            for pass:=0;pass<2&&!chosen;pass++ {
                                target,px,py,best:=-1,-1,-1,int(1<<30)
                                for _,candidate:=range c.candidates {
                                    u:=g.battleUnits[candidate.Target]
                                    if (u.Squadron>0)!=(pass==0)||candidate.Mask==0||u.Ships==0 {continue}
                                    if u.Ships>=best {continue}
                                    x,y:=unitScreenPos(g,candidate.Target)
                                    for dy:=0;dy<16;dy++ {for dx:=0;dx<16;dx++ {
                                        hits:=g.fleetCommandHit(x+dx,y+dy)
                                        found:=false;for _,hit:=range hits {if hit==candidate.Target {found=true}}
                                        if found {target,px,py,best=candidate.Target,x+dx,y+dy,u.Ships;break}
                                    };if target==candidate.Target {break}}
                                }
                                if target>=0 {
                                    hits:=g.fleetCommandHit(px,py)
                                    for n:=0;n<len(hits)&&g.fleetCombat.target!=target;n++ {right(px,py)}
                                    click(px,py)
                                    if g.fleetCombat.target!=target {t.Fatal("公開一般分隊選敵不同",target,g.fleetCombat.target)}
                                    chosen=true
                                }
                            }
''' + s[end:]
if a.target == 'middle':
    marker='if (u.Squadron>0)!=(pass==0)||candidate.Mask==0||u.Ships==0 {continue}'
    assert s.count(marker)==1
    s=s.replace(marker,marker+'\nif pass==0&&u.Squadron==g.fleetByIndex(u.Fleet).Groups-1 {continue}')
stem = 'workplace/cleanup79-r116-'+a.revision
files = {'-probe_test.go':s,
         '-observed-battlefinish.go':checked('workplace/promotion79-r115-h-observed-battlefinish.go').replace('natural115Observation','natural116Observation')}
cleanup = Path('internal/game/battlecleanup.go').read_text()
marker = '\tmapping, err := work.PruneEmptyBattleGroups(index)'
assert cleanup.count(marker) == 1
cleanup = cleanup.replace(marker, '\tif natural116CleanupObservation!=nil {natural116CleanupObservation(g,&work,index,"before",nil)}\n'+marker)
marker = '\tg.battleUnits = out\n'
assert cleanup.count(marker) == 1
cleanup = cleanup.replace(marker, marker+'\tif natural116CleanupObservation!=nil {natural116CleanupObservation(g,g.slot,index,"after",mapping)}\n')
files['-observed-battlecleanup.go'] = cleanup
files['-overlay.json'] = json.dumps({'Replace': {
    '/src/internal/game/finish_player_probe_test.go':'/src/'+stem+'-probe_test.go',
    '/src/internal/game/battlefinish.go':'/src/'+stem+'-observed-battlefinish.go',
    '/src/internal/game/battlecleanup.go':'/src/'+stem+'-observed-battlecleanup.go'}},indent=2)+'\n'
for suffix,value in files.items():
    path = Path(stem+suffix)
    assert not path.exists() and path.parent.stat().st_uid == path.parent.stat().st_gid == 1000
    with path.open('x') as f:
        f.write(value)
subprocess.run(['gofmt','-w',*[stem+s for s in files if s.endswith('.go')]],check=True)
print('prepared',stem,'normal nonflagship targeting; no death injection')
