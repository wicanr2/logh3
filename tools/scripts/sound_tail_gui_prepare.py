"""第81輪：正常公開設定關配樂，沿既有真實X11，記正式Oto未送出buffer。"""
from pathlib import Path
import json
import os
import subprocess
import sys

os.chdir('/src')
revision=sys.argv[1]
mode=sys.argv[2]
assert revision.isalnum() and mode in ['off','mixed']
subprocess.run(['python3','tools/scripts/fleet_playback_prepare.py',revision],check=True)

gameout=Path(f'workplace/round79-fleet-gui-game-{revision}.go')
gameout.write_text(gameout.read_text()+'''
// Test-only：從正式commandBar／settings熱區讀中心，不寫設定或局面。
func (g *Game) SoundTailPoint(action string) (int,int) {
 if action=="system"&&g.state==StateCommandPhase {
  for _,b:=range commandBar {if b.Menu==StateSysMenu {return b.X+b.W/2,barRow+rowHeight/2}}
 }
 if action=="settings"&&g.state==StateSysMenu {
  b:=commandBar[menuOf(g.state)]
  for n,state:=range b.Items {if state==StateSettings {return b.X+b.ItemW/2,barRow+n*rowHeight+rowHeight/2}}
 }
 if spec:=g.commandSpec(g.state);spec!=nil {
  for _,h:=range spec.Hotspots {if h.Action==action&&!h.Disabled {return h.X+h.W/2,h.Y+h.H/2}}
 }
 panic("正式熱區不存在："+g.state.String()+" "+action)
}
''')
planout=Path(f'workplace/round80-plan-source-{revision}.go')
s=planout.read_text()
if mode=='off':
    s=s.replace('for _, in := range parent.Inputs {', '''configured:=false
 for _, in := range parent.Inputs {
  if !configured&&g.State()==game.StateCommandPhase&&g.MusicSettings().BGM==3 {
   for _,act:=range []string{"system","settings","device-2","close"} {
    px,py:=g.SoundTailPoint(act);step("click",px,py,0,0)
   }
   if g.State()!=game.StateCommandPhase||g.MusicSettings().Device!=game.MusicOff {panic("正常關配樂失敗")}
   configured=true
  }''',1)
    s=s.replace('if g.State() != game.StateBattle ||', 'if !configured {panic("沒有正常S3設定操作")}\n if g.State() != game.StateBattle ||',1)
    s=s.replace('if err != nil || string(a) != string(b) {', '''if n=="WORK.DT0" {b=append([]byte(nil),b...);b[0x1F]=byte(game.MusicOff)}
 if err != nil || string(a) != string(b) {''',1)
    s=s.replace('"sound_enabled": true,','"music_off_by_public_settings":true,"sound_enabled": true,',1)
else:s=s.replace('"sound_enabled": true,','"music_off_by_public_settings":false,"sound_enabled": true,',1)
planout.write_text(s)

musicout=Path(f'workplace/round80-music-observe-{revision}.go')
s=musicout.read_text().replace('"stream_bytes": length','"buffer":p.SoundTailProbe(),"stream_bytes": length',1)
assert 'SoundTailProbe()' in s
musicout.write_text(s)
overlay=Path(f'workplace/round79-fleet-gui-overlay-{revision}.json')
d=json.loads(overlay.read_text())
cache=Path(os.environ['GOMODCACHE'])/'github.com/hajimehoshi/ebiten/v2@v2.8.8/audio/audio.go'
assert cache.is_file()
d['Replace'][str(cache)]=str(Path('workplace/round81-audio-observe-baseline.go').resolve())
overlay.write_text(json.dumps(d,indent=2)+'\n')
driverout=Path(f'workplace/round80-driver-{revision}.py')
s=driverout.read_text().replace("default='/src/workplace/round79-input-v2'", "default='/src/workplace/round81-input'")
s=s.replace("send_input({'Type':4,'Button':3})", "send_input({'Type':4,'Button':3,'Status':str(statuspath),'Edge':edge})")
# 正常右鍵同樣由常駐工具確認正式edge後放開。左鍵／鍵盤仍原送法。
s=s.replace("send_input(down)\n            try:", "send_input(dict(down,Status=str(statuspath),Edge=edge) if slot==1 else down)\n            try:")
s=s.replace("    for i,action in enumerate(plan['actions']):", """    # 模型的一次RightClick只移一格；實體按住可跨數個Update。
    # 合併沒有命令／快照的連續捲動，只核下一個玩家操作前的等價鏡頭。
    actions=[]
    j=0
    def scroll_action(a):
        b=a['before']
        return a['kind']=='right' and b['state']=='StateBattle' and b['idle'] and b['stage']==-1 and b['order']==0
    while j<len(plan['actions']):
        first=plan['actions'][j];end=j+1
        if scroll_action(first):
            while end<len(plan['actions']) and scroll_action(plan['actions'][end]):end+=1
            assert all('snapshot' not in a for a in plan['actions'][j:end])
            unchanged=lambda b:{k:v for k,v in b.items() if k!='view'}
            assert unchanged(first['before'])==unchanged(plan['actions'][end-1]['after'])
        action=dict(plan['actions'][end-1]);action['before']=first['before'];action['model_indices']=list(range(j,end))
        actions.append(action);j=end
    for i,action in enumerate(actions):""",1)
s=s.replace("'at_seconds':time.monotonic()-started,'before':before,'after':after})", "'model_indices':action['model_indices'],'at_seconds':time.monotonic()-started,'before':before,'after':after})",1)
s=s.replace("'input_mapping':'右鍵保持以實際鏡頭回授定位；額外手勢完整記錄，非相同實體時序'", "'input_mapping':'連續模型捲動合併為公開手勢，命令前等價鏡頭；原模型索引及額外手勢完整記錄，非相同實體時序'")
s=s.replace("escape=visits%2==1\n        if escape:delta=[-v for v in delta]", """# 舊控制器把任何重訪都反向，單軸逼近13時反而退到0。
        # 重訪且兩軸皆未到位時，改走誤差較大的一軸以解除對角循環。
        # 單軸始終朝目標，仍保留40次上限與最後精確比較。
        escape=visits>0 and all(delta)
        if escape:
            if abs(want[0]-st['view'][0])>=abs(want[1]-st['view'][1]):delta[1]=0
            else:delta[0]=0""",1)
driverout.write_text(s)
assert all(part in s for part in ['model_indices','delta[1]=0','/src/workplace/round81-input','\'Status\':str(statuspath)'])
subprocess.run(['gofmt','-w',str(gameout),str(planout),str(musicout)],check=True)
print('正常公開設定',mode,'正式buffer只讀，來源皆新檔')
