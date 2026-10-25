"""166：保留雙國徽故事及完整實戰，替換所有舊海報操作段。Docker內執行。"""
import argparse
import copy
import math
import os
from pathlib import Path
import random
import shutil
import subprocess

from PIL import Image, ImageDraw

import poster_story_promo as movie
import poster_frame_release as build

ROOT=build.ROOT;VERSION=build.VERSION;BATCH=build.BATCH;WORK=BATCH/'movie'
OUT=build.OUT/'promo';NAME='LOGH3-'+VERSION+'-promo.mp4'
PARENT=ROOT/'dist-all/v.1.6.0-20261024/promo'
OLDWORK=ROOT/'workplace/poster151/movie'
PLAYER=BATCH/'player-linux';GARRISON=BATCH/'garrison-linux-r3'
sha,read,write,run,probe=movie.sha,movie.read,movie.write,movie.run,movie.probe

def configure():
    movie.VERSION,movie.BATCH,movie.WORK,movie.OUT,movie.NAME=VERSION,BATCH,WORK,OUT,NAME
    movie.SFX=build.STAGE/'common/music'

def prepare():
    assert os.getuid()==1000 and BATCH.stat().st_uid==1000 and not WORK.exists() and not OUT.exists()
    parent=read(PARENT/'PLAN.json');audit=read(PARENT/'AUDIT.json')
    assert audit['passed'] and audit['visual_review'].startswith('已目視')
    assert sha(PARENT/audit['file'])==audit['sha256']
    inputs={}
    def bind(path):inputs[str(path.relative_to(ROOT))]=sha(path)
    for name,h in parent['input_sha256'].items():assert sha(ROOT/name)==h;inputs[name]=h
    for path in [PARENT/'PLAN.json',PARENT/'AUDIT.json',PARENT/audit['file'],Path(__file__).resolve(),
        build.OUT/'full'/('LOGH3-'+VERSION+'-linux-x86_64.AppImage')]:bind(path)
    def marks(folder):
        r=read(folder/'RESULT.json')
        assert r['passed'] and r['normal_entry'] and r['appimage'] and r['automatic_resources']
        assert not any(r[k]for k in['state_injection','seed_injection','script_flag'])
        assert sha(folder/'RAW.mp4')==r['video_sha256']
        assert r['binary_sha256']==sha(build.OUT/'full'/('LOGH3-'+VERSION+'-linux-x86_64.AppImage'))
        bind(folder/'RESULT.json');bind(folder/'RAW.mp4')
        return r,{e['name']:(e.get('wall_ns',e.get('wall_begin_ns',0))-r['record_started_utc_ns'])/1e9 for e in r['events']if e.get('name')}
    player,pm=marks(PLAYER);garrison,gm=marks(GARRISON)
    scenes=copy.deepcopy(parent['scenes'])
    def fresh(index,folder,start,duration=None):
        s=scenes[index];s['source']=str(folder/'RAW.mp4');s['start']=max(0,start)
        if duration is not None:s['duration']=math.ceil(duration*60-1e-6)/60
        s['layout']='wide';s['provenance']=VERSION+'正式AppImage正常入口新錄'
        s.pop('clip_sha256',None)
    transition=(player['quote_transition']['wall_ns']-player['record_started_utc_ns'])/1e9
    fresh(4,PLAYER,transition-1.5)
    scenes[4]['caption']='v1雙主角｜五組名言每60秒輪替；此處為實際切換時刻。'
    for index,name in [(5,'zh-TW-map-v2'),(6,'zh-TW-map-v3'),(7,'zh-CN-map-v1'),(8,'ja-map-v1'),(9,'en-map-v1'),(10,'ko-map-v1'),(11,'zh-TW-editor-special')]:
        fresh(index,PLAYER,pm[name]-.15)
    fresh(12,GARRISON,gm['garrison-orders']-2)
    fresh(13,GARRISON,gm['option-default-off']+1)
    begin,stock,ready,end=gm['first-commander']-1.6,gm['local-ship-stock']-1,gm['fleet-ready']-1,gm['departure-order']+4
    for index,start,stop in [(14,begin,stock),(15,stock,ready),(16,ready,end)]:fresh(index,GARRISON,start,stop-start)
    fresh(17,GARRISON,gm['departed']-.2)
    scenes[-1]['caption']='銀河英雄傳說 III SP｜v1.6.3'
    scenes[-1]['subtitle']='三組海報・五語名言輪替・前線整軍｜Windows / Linux / macOS'
    retained=[]
    for s in scenes:
        if s['kind']!='story' and s['layout']!='wide':
            path=OLDWORK/f'scene-{s["id"]:02}.mp4'
            assert sha(path)==parent['scenes'][s['id']]['clip_sha256'];bind(path)
            retained.append(dict(scene=s['id'],path=str(path.relative_to(ROOT)),sha256=sha(path)))
    # Keep continuous normal-speed reorganisation. Only shorten optional holds
    # if the new real-input duration would exceed the existing single BGM.
    limit=float(probe(movie.MUSIC)['format']['duration'])-.25
    if sum(s['duration']for s in scenes)>limit:
        for index,minimum in [(12,3),(13,3),(0,3),(3,3),(34,3),(5,2),(6,2),(7,1.5),(8,1.5),(9,1.5),(10,1.5)]:
            excess=sum(s['duration']for s in scenes)-limit
            if excess<=0:break
            scenes[index]['duration']=max(minimum,scenes[index]['duration']-math.ceil(excess*60)/60)
    elapsed=0
    for s in scenes:s['start_at']=elapsed;elapsed+=s['duration']
    assert elapsed<=limit,('正常整軍過長，不加速或循環配樂',elapsed,limit)
    assert not any(s['start']<0 for s in scenes if s['kind']!='story')
    WORK.mkdir();OUT.mkdir()
    for path in [movie.FLAGS,movie.MUSIC,movie.SFX/'SFX_00.ogg']:
        if path.exists():bind(path)
    plan=dict(parent,version=VERSION,seconds=elapsed,scenes=scenes,input_sha256=inputs,text_bounds=[],
        inherited_text_bounds=parent['text_bounds'],retained_clips=retained,
        parent_plan=dict(path=str((PARENT/'PLAN.json').relative_to(ROOT)),sha256=sha(PARENT/'PLAN.json')),
        quote_transition=player['quote_transition'],new_scene_ids=list(range(4,18)),
        full_canvas_layout=dict(source=[1280,720],scaled=[1664,936],position=[128,72],subtitle_bands=[[0,0,1920,60],[0,1023,1920,1080]]),
        scope='海報／五語／提督及前線整軍取新版正式AppImage新錄；動畫與完整實戰沿151已驗來源。實際60秒換名言，連續整軍原速。')
    write(WORK/'PLAN.json',plan)
    write(WORK/'PARENT.json',dict(movie=str((PARENT/audit['file']).relative_to(ROOT)),sha256=audit['sha256'],retained=retained))
    print('READY',len(scenes),'scenes',elapsed,'seconds; continuous garrison',end-begin,'seconds',flush=True)

def render():
    plan=read(WORK/'PLAN.json');movie.validate(plan)
    flags=Image.open(movie.FLAGS).convert('RGB');flags.save(WORK/'BACKGROUND.png')
    retained={r['scene']:r for r in plan['retained_clips']}
    for s in plan['scenes']:
        target=WORK/f'scene-{s["id"]:02}.mp4';assert not target.exists()
        if s['id']in retained:
            r=retained[s['id']];shutil.copyfile(ROOT/r['path'],target);assert sha(target)==r['sha256']
        elif s['kind']=='story':
            overlay=Image.new('RGBA',flags.size);draw=ImageDraw.Draw(overlay)
            for y in range(785,1080):draw.line((0,y,1919,y),fill=(2,6,14,round((y-785)/295*195)))
            for xy,value,size,bounds,bold in [
                ((960,70),s['title'],46,(60,20,1860,115),True),((480,160),'銀河帝國',48,(60,115,900,210),True),
                ((1440,160),'自由行星同盟',48,(1020,115,1860,210),True),((960,875),s['caption'],58,(60,810,1860,930),True),
                ((960,970),s['subtitle'],34,(60,935,1860,1010),False),((960,1040),VERSION,26,(60,1015,1860,1075),False)]:
                movie.text(overlay,xy,value,size,bounds,plan,bold)
            overlay.save(WORK/f'overlay-{s["id"]:02}.png')
            command=['ffmpeg','-nostdin','-v','error','-f','rawvideo','-pixel_format','rgb24','-video_size','1920x1080','-framerate','60','-i','pipe:0',
                '-an','-frames:v',str(round(s['duration']*60)),'-c:v','libx264','-preset','veryfast','-crf','19','-threads','2','-pix_fmt','yuv420p','-video_track_timescale','60000',str(target)]
            rng=random.Random(126);stars=[(rng.randrange(1920),rng.randrange(1080),rng.randrange(1,3),rng.randrange(25,70))for _ in range(100)]
            process=subprocess.Popen(command,stdin=subprocess.PIPE)
            try:
                for n in range(round(s['duration']*60)):
                    image=flags.copy();draw=ImageDraw.Draw(image);t=n/60
                    for x,y,r,speed in stars:
                        xx=int((x+t*speed/8)%1920);yy=int((y+t*speed)%1080);draw.ellipse((xx,yy,xx+r,yy+r),fill=(155,165,185))
                    image.paste(overlay,(0,0),overlay);process.stdin.write(image.tobytes())
                process.stdin.close();assert process.wait(timeout=60)==0
            finally:
                if process.poll()is None:process.terminate();process.wait(timeout=10)
        else:
            overlay=Image.new('RGBA',flags.size);draw=ImageDraw.Draw(overlay)
            draw.rectangle((0,0,1919,59),fill=(4,8,16,255));draw.rectangle((0,1023,1919,1079),fill=(4,8,16,255))
            movie.text(overlay,(960,29),s['title'],32,(20,0,1900,60),plan,True)
            movie.text(overlay,(960,1050),s['caption'],26,(20,1025,1900,1079),plan)
            path=WORK/f'overlay-{s["id"]:02}.png';overlay.save(path)
            run('ffmpeg','-nostdin','-v','error','-loop','1','-framerate','60','-i',WORK/'BACKGROUND.png','-ss',s['start'],'-i',s['source'],'-loop','1','-i',path,
                '-filter_complex_threads','1','-filter_complex','[1:v]scale=1664:936:flags=neighbor,setsar=1,fps=60,setpts=PTS-STARTPTS[g];[0:v][g]overlay=128:72:shortest=1[b];[b][2:v]overlay=0:0:shortest=1[out]',
                '-map','[out]','-an','-frames:v',round(s['duration']*60),'-r','60','-c:v','libx264','-preset','veryfast','-crf','19','-threads','2','-pix_fmt','yuv420p','-video_track_timescale','60000',target)
        s['clip_sha256']=sha(target);print('RENDER',s['id'],s['title'],flush=True)
    write(WORK/'PLAN.json',plan)
    (WORK/'concat.txt').write_text(''.join("file '"+str(WORK/f'scene-{s["id"]:02}.mp4')+"'\n"for s in plan['scenes']))
    movie.soundtrack(plan)
    run('ffmpeg','-nostdin','-v','error','-f','concat','-safe','0','-i',WORK/'concat.txt','-i',WORK/'MIX.wav','-map','0:v:0','-map','1:a:0',
        '-t',plan['seconds'],'-c:v','copy','-video_track_timescale','60000','-c:a','aac','-b:a','256k','-ar',movie.RATE,'-ac','2','-movflags','+faststart',WORK/NAME)

def audit():
    movie.audit()
    value=read(WORK/'AUDIT.json');plan=read(WORK/'PLAN.json')
    value['generator_wrapper']=dict(path=str(Path(__file__).resolve().relative_to(ROOT)),sha256=sha(Path(__file__).resolve()))
    value['retained_scene_count']=len(plan['retained_clips'])
    value['new_poster_scene_count']=len(plan['new_scene_ids'])
    value['full_source_canvas_preserved']=True
    value['parent_movie_sha256']=read(WORK/'PARENT.json')['sha256']
    write(WORK/'AUDIT.json',value)

def publish():
    plan=read(WORK/'PLAN.json');movie.validate(plan);audit=read(WORK/'AUDIT.json')
    assert audit['passed'] and audit['visual_review'].startswith('已目視') and sha(WORK/NAME)==audit['sha256']
    names=[NAME,'PLAN.json','AUDIT.json','CONTACT.png','SOUNDTRACK.json','FFPROBE.json','SOURCE-PLAYBACK.json','SAMPLES.json','PARENT.json']
    for name in names:assert not (OUT/name).exists();shutil.copyfile(WORK/name,OUT/name)
    (OUT/'SHA256SUMS.txt').write_text(''.join(sha(OUT/name)+'  '+name+'\n'for name in names))
    print('INSTALLED',OUT,flush=True)

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('mode',choices=['prepare','render','audit','publish']);a=p.parse_args();configure();globals()[a.mode]()
