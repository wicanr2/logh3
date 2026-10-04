"""第80輪：沿第79輪真實鍵鼠，只加正式播放器的只讀觀測。"""
from pathlib import Path
import json
import os
import subprocess
import sys

os.chdir('/src')
revision=sys.argv[1]
assert revision.isalnum()
def create(path,text):
    p=Path(path)
    assert not p.exists() and p.parent.stat().st_uid==os.getuid(),path
    p.write_text(text)

subprocess.run(['python3','tools/scripts/fleet_player_gui_prepare.py',revision],check=True)
overlay=Path(f'workplace/round79-fleet-gui-overlay-{revision}.json')
mapping=json.loads(overlay.read_text())
mainout=Path(f'workplace/round79-fleet-gui-main-{revision}.go')
s=mainout.read_text().replace('st["edges"] = fleetGUIEdges', '''
if p,ok:=a.music.(*musicPlayer);ok {st["music_track"]=p.Playing();st["music_playing"]=p.cur!=nil&&p.cur.IsPlaying()}
if p,ok:=a.effects.(*effectPlayer);ok {st["effect_playing"]=p.decoder.cur!=nil&&p.decoder.cur.IsPlaying()}
st["edges"] = fleetGUIEdges''')
assert 'st["effect_playing"]' in s
mainout.write_text(s)

music='cmd/gin3/music.go'
s=Path(music).read_text().replace('import (','import (\n"encoding/json"\n"path/filepath"\n"time"',1)
s=s.replace('p.cur, p.name = pl, track','p.cur, p.name = pl, track\n fleetPlaybackEvent("music_start",track,-1,pl,st.Length())',1)
s=s.replace('_ = p.cur.Close()','if p.name!="" {fleetPlaybackEvent("music_stop",p.name,-1,p.cur,0)}\n _ = p.cur.Close()',1)
s+='''
// 不替換解碼器、播放器或PCM；只記正式呼叫後的播放器狀態。
func fleetPlaybackEvent(kind,track string,id int,p *eaudio.Player,length int64) {
 dir:=os.Getenv("LOGH3_FLEET_GUI_OUT");if dir=="" {return}
 b,err:=json.Marshal(map[string]any{"kind":kind,"track":track,"id":id,"unix_ns":time.Now().UnixNano(),
 "playing":p.IsPlaying(),"position_ns":p.Position().Nanoseconds(),"stream_bytes":length})
 if err!=nil {panic(err)}
 f,err:=os.OpenFile(filepath.Join(dir,"playback-events.jsonl"),os.O_WRONLY|os.O_CREATE|os.O_APPEND,0644)
 if err!=nil {panic(err)}
 if _,err=f.Write(append(b,'\\n'));err!=nil {f.Close();panic(err)}
 if err=f.Close();err!=nil {panic(err)}
}
'''
musicout=f'workplace/round80-music-observe-{revision}.go'
create(musicout,s)
effects='cmd/gin3/effects.go'
s=Path(effects).read_text().replace('p.decoder.cur = pl','p.decoder.cur = pl\n fleetPlaybackEvent("effect_start","",id,pl,st.Length())',1)
s=s.replace('func (p *effectPlayer) Stop() { p.decoder.Stop() }', '''func (p *effectPlayer) Stop() {
 if p.decoder.cur!=nil {fleetPlaybackEvent("effect_stop","",-1,p.decoder.cur,0)}
 p.decoder.Stop()
}''')
effectout=f'workplace/round80-effects-observe-{revision}.go'
create(effectout,s)
for src,dst in [(music,musicout),(effects,effectout)]:mapping['Replace'][str(Path(src).resolve())]=str(Path(dst).resolve())
overlay.write_text(json.dumps(mapping,indent=2)+'\n')

plan=Path('tools/scripts/fleet_player_gui_plan.go').read_text()
plan=plan.replace('var actions []map[string]any','if err=g.EnableSoundEffects(true);err!=nil {panic(err)}\n var sounds []int\n var actions []map[string]any',1)
plan=plan.replace('g.Frame()\n','g.Frame()\n sounds=append(sounds,g.DrainSoundEffects()...)\n')
plan=plan.replace('"faction": parent.Faction, "actions": actions','"sound_ids":sounds,"sound_enabled":true,"faction": parent.Faction, "actions": actions',1)
planout=f'workplace/round80-plan-source-{revision}.go'
create(planout,plan)
subprocess.run(['gofmt','-w',str(mainout),musicout,effectout,planout],check=True)

driver=Path('tools/scripts/fleet_player_gui_driver.py').read_text()
driver=driver.replace("    raise RuntimeError('真實右鍵鏡頭未收斂 '+str(want)+' '+str(status()))", "    final=status()\n    if final['view']==want:return final\n    raise RuntimeError('真實右鍵鏡頭未收斂 '+str(want)+' '+str(final))",1)
driver=driver.replace("default='/src/workplace/round79-gin3-v3'",f"default='/src/workplace/round80-gin3-{revision}'",1)
driver=driver.replace("'-save',str((out/'player-save').resolve())","'-music','/src/workplace/round80-library','-save',str((out/'player-save').resolve())",1)
driver=driver.replace("request({'Save':True})",'''request({'Save':True})
    wait(lambda s:s is not None and s.get('effect_playing') is False,'等正式音效完成',20)
    time.sleep(.2)
    events=[json.loads(line) for line in (out/'playback-events.jsonl').read_text().splitlines()]
    starts=[e for e in events if e['kind']=='effect_start']
    assert [e['id'] for e in starts]==plan['sound_ids'],'正式播放器音效序列不同'
    assert all(e['playing'] and e['stream_bytes']>0 for e in starts)
    stopped=[e for e in events if e['kind']=='effect_stop']
    inputs=subprocess.run(['pactl','-f','json','list','sink-inputs'],capture_output=True,text=True,check=True).stdout
    (out/'actual-sink-inputs.json').write_text(inputs)
    assert json.loads(inputs),'正式後端未連至實際虛擬輸出'
    (out/'playback-summary.json').write_text(json.dumps({'effect_ids':[e['id'] for e in starts],
        'effect_stops':len(stopped),'effect_stops_while_playing':sum(e['playing'] for e in stopped),
        'music_starts':[e['track'] for e in events if e['kind']=='music_start'],
        'physical_device_tested':False,'human_listening':False},ensure_ascii=False,indent=2)+'\\n')''',1)
driver=driver.replace("'sound_device_tested':False","'sound_device_tested':'ALSA default / PulseAudio clocked virtual sink'",1)
driverout=f'workplace/round80-driver-{revision}.py'
create(driverout,driver)
print('正式播放器只讀觀測，正常X11輸入；第79輪來源未修改')
