"""第81輪：真正正式播放器PCM探針，28段完整與舊替換方式；不產測試音源。"""
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

source='cmd/gin3/effects_test.go'
s=Path(source).read_text().replace('import (','import (\n eaudio "github.com/hajimehoshi/ebiten/v2/audio"',1)
s+='''
// 診斷入口直接用正式effectPlayer及真正Context，沒有Game局面或人工PCM。
func TestSoundTailDeviceProbe(t *testing.T) {
 out:=os.Getenv("LOGH3_AUDIO_SINK_OUT")
 if out=="" {t.Skip("需tools/playback.sh實際裝置")}
 mode:=os.Getenv("LOGH3_SOUND_TAIL_MODE")
 if mode!="complete"&&mode!="replace" {t.Fatal("需complete或replace")}
 lib:=audio.NewSoundLibrary("/src/workplace/round80-library")
 if lib.Have()!=28 {t.Skip("缺完整真實28段音效")}
 p:=newEffectPlayer(eaudio.NewContext(audio.SampleRate),lib)
 defer p.Stop()
 var ids []int
 if mode=="complete" {for i:=0;i<28;i++ {ids=append(ids,i)}} else {ids=[]int{19,4}}
 var events []map[string]any
 for index,id:=range ids {
  st,ok,err:=p.decoder.decode(audio.SoundTrack(id));if err!=nil||!ok {t.Fatal(id,err)}
  pcm,err:=io.ReadAll(st);if err!=nil||int64(len(pcm))!=st.Length() {t.Fatal("完整解碼",id,err)}
  file:=filepath.Join(out,fmt.Sprintf("source-%02d.s16le",id))
  if _,err=os.Stat(file);!os.IsNotExist(err) {t.Fatal("拒絕覆寫",file)}
  if err=os.WriteFile(file,pcm,0644);err!=nil {t.Fatal(err)}
  if p.decoder.cur!=nil {
   events=append(events,map[string]any{"kind":"before_replace","id":ids[index-1],"playing":p.decoder.cur.IsPlaying(),"buffer":p.decoder.cur.SoundTailProbe(),"unix_ns":time.Now().UnixNano()})
  }
  if err=p.PlayEffect(id);err!=nil {t.Fatal(err)}
  length:=time.Duration(int64(len(pcm))*int64(time.Second)/(4*44100))
  events=append(events,map[string]any{"kind":"start","id":id,"bytes":len(pcm),"duration_ns":length.Nanoseconds(),"unix_ns":time.Now().UnixNano(),"buffer":p.decoder.cur.SoundTailProbe()})
  if mode=="replace"&&index==0 {
   // 第78輪原始Game兩個請求間隔；固定一次，不為了截短挑時點。
   time.Sleep(961184000*time.Nanosecond)
  } else {
   deadline:=time.Now().Add(length+3*time.Second)
   for p.decoder.cur.IsPlaying()&&time.Now().Before(deadline) {time.Sleep(time.Millisecond)}
   if p.decoder.cur.IsPlaying() {t.Fatal("正式播放未完成",id)}
   events=append(events,map[string]any{"kind":"complete","id":id,"buffer":p.decoder.cur.SoundTailProbe(),"unix_ns":time.Now().UnixNano()})
   p.releaseFinished()
   time.Sleep(150*time.Millisecond)
  }
 }
 b,err:=json.MarshalIndent(map[string]any{"classification":"正式effectPlayer局部診斷／ALSA PCM，非正常Game入口","mode":mode,"events":events},"","  ")
 if err!=nil {t.Fatal(err)}
 if err=os.WriteFile(filepath.Join(out,"probe.json"),append(b,'\\n'),0644);err!=nil {t.Fatal(err)}
}
'''
out=f'workplace/round81-tail-test-{revision}.go'
create(out,s)

cache=Path(os.environ['GOMODCACHE'])/'github.com/hajimehoshi/ebiten/v2@v2.8.8/audio/audio.go'
assert cache.is_file()
observer=cache.read_text()+'''
// Test-only：只讀正式Oto緩衝量及來源位置，不碰buffer或音量。
func (p *Player) SoundTailProbe() map[string]any {
 p.p.m.Lock();defer p.p.m.Unlock()
 if p.p.player==nil {return map[string]any{"available":false}}
 return map[string]any{"available":true,"playing":p.p.player.IsPlaying(),
 "buffer_bytes":p.p.player.BufferedSize(),"stream_bytes":p.p.stream.position(),"sample_bytes":p.p.bytesPerSample}
}
'''
observerout=f'workplace/round81-audio-observe-{revision}.go'
create(observerout,observer)
subprocess.run(['gofmt','-w',out,observerout],check=True)
create(f'workplace/round81-tail-overlay-{revision}.json',json.dumps({'Replace':{
    str(Path(source).resolve()):str(Path(out).resolve()),str(cache):str(Path(observerout).resolve())}},indent=2)+'\n')
print('28段真實素材、正式一次性播放器、只讀Oto／來源觀測；拒絕覆寫')
