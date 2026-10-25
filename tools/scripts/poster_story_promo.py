"""151：沿雙國徽、主角動畫及原速實戰手法，重剪v1.6.0完整故事推廣片。Docker內執行。"""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import random
import shutil
import subprocess
import wave

import numpy as np
from PIL import Image, ImageDraw, ImageFont

import garrison_promo_movie as media
import promo_v140_movie as legacy

ROOT = Path('/src'); VERSION = 'v.1.6.0-20261024'
BATCH = ROOT/'workplace/poster151'; WORK = BATCH/'movie'; PREVIOUS = BATCH/'previous'
OUT = ROOT/'dist-all'/VERSION/'promo'; NAME = 'LOGH3-'+VERSION+'-promo.mp4'
SETS = ROOT/'workplace/poster150/player-linux'; GARRISON = ROOT/'workplace/poster150/garrison-linux'
FLAGS = ROOT/'workplace/hd127/promo-final/BACKGROUND.png'
OLD = ROOT/'dist-all/v.1.5.0-20261019/promo'
MUSIC = ROOT/'workplace/round132/music/BGM_0F.ogg'
SFX = ROOT/'workplace/poster150/package-linux'/('LOGH3-'+VERSION+'-linux-amd64')/'music'
FONT = Path('/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc')
BOLD = Path('/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc')
RATE = 44100
sha, read, write, probe, run, capture = media.sha, media.read, media.write, media.probe, media.run, media.capture


def record(path): return dict(path=str(path.relative_to(ROOT)), sha256=sha(path), bytes=path.stat().st_size)


def prepare():
    assert os.getuid() == BATCH.parent.stat().st_uid == 1000 and not BATCH.exists()
    BATCH.mkdir(); PREVIOUS.mkdir(); WORK.mkdir()
    mapping = []
    sources = [*sorted(OUT.iterdir()), ROOT/'docs/internals/data/v160-delivery.json', OUT.parent/'SHA256SUMS.txt',
               ROOT/'tools/scripts/poster_delivery.py', ROOT/'tools/scripts/poster_promo_movie.py', ROOT/'tools/scripts/garrison_promo_movie.py']
    for source in sources:
        assert source.is_file() and source.stat().st_uid == 1000
        target = PREVIOUS/source.relative_to(ROOT); target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(source, target); assert sha(target) == sha(source)
        mapping.append(dict(original=str(source.relative_to(ROOT)), backup=str(target.relative_to(ROOT)), sha256=sha(source)))
    write(PREVIOUS/'MAPPING.json', dict(reason='使用者要求國徽故事與完整手法重剪；保留150初版，不覆寫遠端發行', files=mapping))
    inputs = {}
    def bind(path):
        path = Path(path); inputs[str(path.relative_to(ROOT))] = sha(path)
    for path in [FLAGS, FLAGS.parent/'PREPARE.json', OLD/'PLAN.json', OLD/'SOURCE-PLAYBACK.json', MUSIC, Path(__file__).resolve(), ROOT/'tools/scripts/garrison_promo_movie.py', ROOT/'tools/scripts/promo_v140_movie.py']: bind(path)
    assert sha(MUSIC) == '48c22ed0b5d9fec5d0b9bed70e7a06ae1ab3a07faef99da80ddfd21b4644739c'
    previous = read(OLD/'PLAN.json'); source_audit = read(OLD/'SOURCE-PLAYBACK.json'); assert source_audit['passed']
    def marks(folder):
        r = read(folder/'RESULT.json'); assert r['passed'] and r['normal_entry'] and not any(r[k] for k in ['script_flag','state_injection','seed_injection'])
        assert sha(folder/'RAW.mp4') == r['video_sha256']
        bind(folder/'RESULT.json'); bind(folder/'RAW.mp4')
        return {e['name']:(e.get('wall_ns',e.get('wall_begin_ns',0))-r['record_started_utc_ns'])/1e9 for e in r['events'] if e.get('name')}
    sets = marks(SETS); garrison = marks(GARRISON); scenes = []
    def add(kind, duration, title, caption, **kw):
        scenes.append(dict(kind=kind, duration=math.ceil(duration*60-1e-6)/60, title=title, caption=caption, **kw))
    def game(folder, start, duration, title, caption, **kw):
        add('game', duration, title, caption, source=str(folder/'RAW.mp4'), start=start, layout='wide', provenance='v1.6.0正式ZIP新實錄', **kw)
    add('story',4,'銀河英雄傳說 III SP','帝國與同盟，對峙百餘年。',subtitle='戰火橫越星海，兩個國家仍在追尋各自的未來。')
    for old_id,title,caption in [(1,'萊因哈特，向銀河的頂點進軍','帝國的年輕將領，立志統一銀河。'),(2,'楊威利，戰場上的魔術師','同盟的不敗名將，心中盼望的仍是和平。')]:
        s = previous['scenes'][old_id]; source = Path(s['file']); assert sha(source) == previous['inputs'][str(source)]; bind(source)
        add('anime',3,title,caption,source=str(source),start=s['start'],layout='anime',provenance='沿既有指定1988動畫，原生852×480，保留原字幕與水印，不稱遊戲實錄')
    add('story',4,'亞姆力札，遠征軍陷入危局','深入帝國領土的艦隊，面臨補給困境。',subtitle='如果這一次，前線的下一道命令由你下達？')
    for name,duration,title,caption in [
        ('zh-TW-map-v1',3,'兩位英雄，同一片銀河','v1雙主角｜左右斜體名言，陪你走進這場戰爭。'),
        ('zh-TW-map-v2',2.5,'將領們，集結於各自的旗幟下','v2兩軍將領｜左帝國、右同盟。'),
        ('zh-TW-map-v3',2.5,'英雄身邊，還有值得守護的人','v3親友｜安妮羅潔與齊格飛，菲列特利加、尤里安與夥伴。'),
    ]: game(SETS,sets[name]-.15,duration,title,caption)
    for lang,label in [('zh-CN','简体中文'),('ja','日本語'),('en','English'),('ko','한국어')]:
        game(SETS,sets[lang+'-map-v1']-.1,2,'用熟悉的語言，下達你的命令',label+'｜主要介面與名言同步切換；另支援繁體中文。')
    game(SETS,sets['zh-TW-editor-special']-.1,3,'提督的能力，成為你的戰術選擇','設定中可編輯能力與特殊加成，保存後繼續你的遠征。')
    game(GARRISON,garrison['garrison-orders']-2,4,'兵力已抵前線，遠征仍要繼續','艦隊駐留，把提督與船艦留在當地。')
    game(GARRISON,garrison['option-default-off']+1,4,'讓存活的艦艇，接續旗艦任務','非戰艦旗艦預設關閉；修護艦E與運補艦T始終排除。')
    first=garrison['first-commander']-1.6; stock=garrison['local-ship-stock']-1; ready=garrison['fleet-ready']-1; departed=garrison['departure-order']+4
    for start,end,title,caption in [
        (first,stock,'重新任命，讓提督接掌前線艦隊','從駐地查看提督，再改選派特，組成第3艦隊。'),
        (stock,ready,'就地領船，重整你的戰力','當地庫存用於編成；巡航艦可當旗艦，E／T仍會被拒絕。'),
        (ready,departed,'下一道命令：向目的地出發','確認編成、選定目的地；從改選提督到出發保持連續原速。'),
    ]: game(GARRISON,start,end-start,title,caption,continuous_garrison=True)
    game(GARRISON,garrison['departed']-.2,1.4,'下一回合，艦隊抵達','同一局面接續；此處剪去回合等待。')
    selected=[]
    for s in previous['scenes']:
        if s['id'] in (9,10) or s.get('fortress_animation') or (s.get('battle_animation') and not s.get('reused')):
            source=Path(s['file']); folder=source.parent; bind(source)
            evidence=next(r for r in source_audit['recordings'] if r['source']==str(folder))
            assert evidence['audit']['passed'] and not evidence['audit']['skipped_frames']
            assert sha(folder/'PLAYBACK-AUDIT.json')==evidence['sha256']
            assert sha(source)==evidence['audit']['media']['raw_sha256']
            for name in ['TRACE.json','FRAMES.jsonl','PLAYBACK-AUDIT.json']:
                p=folder/name; bind(p)
            if s.get('playback_id'): assert s['playback_id'] in {row['id'] for row in evidence['audit']['scenes']}
            if s['id'] in (9,10): duration=1.5; title='在星圖部署，於戰場決勝'; caption='兩陣營既有正常實戰，保留原版操作與陣營HD。'
            elif s.get('fortress_animation'): duration=s['duration']; title='伊謝爾倫要塞砲，改變戰場的力量'; caption='既有正式實錄｜從發射到戰損，保留完整原速演出。'
            else:
                duration=s['duration']; title={0:'光線砲，劃破星海',1:'飛彈，突破敵陣',2:'戰鬥機，迎向敵艦'}[s['weapon']]
                caption='既有正式實戰｜完整交火與結果，原速演出及遊戲音效。'
            add('game',duration,title,caption,source=str(source),start=s['start'],layout='legacy',provenance='沿v1.5已驗正常實戰，非v1.6.0新錄',
                battle_animation=s.get('battle_animation',False),fortress_animation=s.get('fortress_animation',False),playback_id=s.get('playback_id'),weapon=s.get('weapon'))
            selected.append(dict(scene=len(scenes)-1,previous_scene=s['id'],source=str(source.relative_to(ROOT)),audit_sha256=evidence['sha256']))
    add('story',4,'下一頁歷史，由你下令','銀河英雄傳說 III SP｜v1.6.0',subtitle='三組海報・五語介面・前線整軍｜Windows / Linux / macOS')
    elapsed=0
    for i,s in enumerate(scenes): s.update(id=i,start_at=elapsed); elapsed+=s['duration']
    assert elapsed<float(probe(MUSIC)['format']['duration']),elapsed
    assert {s['weapon'] for s in scenes if s.get('battle_animation')}=={0,1,2}
    plan=dict(version=VERSION,seconds=elapsed,scenes=scenes,input_sha256=inputs,
        background=record(FLAGS),fonts=[dict(path=str(p),sha256=sha(p)) for p in [FONT,BOLD]],text_bounds=[],selected_previous=selected,
        previous_delivery=record(PREVIOUS/'docs/internals/data/v160-delivery.json'),previous_mapping=record(PREVIOUS/'MAPPING.json'),
        story_sources=read(FLAGS.parent/'PREPARE.json')['story_sources'],
        storyboard='雙國徽與百年對峙→1988動畫雙主角→亞姆力札危局→英雄及親友海報→五語與提督能力→前線整軍及再次出擊→兩軍戰場與要塞／三武器→由玩家下令',
        scope='新版操作為v1.6.0正式ZIP新錄；動畫與戰術實戰沿已驗舊來源。完整交火保留原速，駐地改選至出發連續。',
        rights='含既有指定1988動畫、原版及衍生音畫，只留本機，不公開發布。',music='同首BGM_0F從零連續播放，不循環、不隨剪接重啟；遊戲音效按played觀測獨立重混。')
    write(WORK/'PLAN.json',plan)
    print('READY',len(scenes),'段',elapsed,'秒，雙國徽故事／主角／新版操作／完整原速戰鬥',flush=True)


def text(image,xy,value,size,bounds,plan,bold=False):
    d=ImageDraw.Draw(image); font=ImageFont.truetype(str(BOLD if bold else FONT),size,index=3)
    box=d.textbbox(xy,value,font=font,anchor='mm'); assert box[0]>=bounds[0] and box[1]>=bounds[1] and box[2]<=bounds[2] and box[3]<=bounds[3],(value,box,bounds)
    d.text(xy,value,font=font,anchor='mm',fill='#f2e4c6',stroke_width=1,stroke_fill='#08101a')
    plan['text_bounds'].append(dict(text=value,ink_box=list(box),safe_box=list(bounds)))


def validate(plan):
    assert all(sha(ROOT/n)==digest for n,digest in plan['input_sha256'].items())


def render():
    plan=read(WORK/'PLAN.json'); validate(plan); assert not (WORK/NAME).exists()
    flags=Image.open(FLAGS).convert('RGB'); flags.save(WORK/'BACKGROUND.png')
    for s in plan['scenes']:
        target=WORK/f'scene-{s["id"]:02}.mp4'; assert not target.exists()
        if s['kind']=='story':
            overlay=Image.new('RGBA',flags.size); d=ImageDraw.Draw(overlay)
            for y in range(785,1080): d.line((0,y,1919,y),fill=(2,6,14,round((y-785)/295*195)))
            text(overlay,(960,70),s['title'],46,(60,20,1860,115),plan,True)
            text(overlay,(480,160),'銀河帝國',48,(60,115,900,210),plan,True)
            text(overlay,(1440,160),'自由行星同盟',48,(1020,115,1860,210),plan,True)
            text(overlay,(960,875),s['caption'],58,(60,810,1860,930),plan,True)
            text(overlay,(960,970),s['subtitle'],34,(60,935,1860,1010),plan)
            text(overlay,(960,1040),VERSION,26,(60,1015,1860,1075),plan)
            overlay.save(WORK/f'overlay-{s["id"]:02}.png')
            preview=Image.alpha_composite(flags.convert('RGBA'),overlay).convert('RGB'); preview.save(WORK/f'card-{s["id"]:02}.png')
            command=['ffmpeg','-nostdin','-v','error','-f','rawvideo','-pixel_format','rgb24','-video_size','1920x1080','-framerate','60','-i','pipe:0','-an','-frames:v',str(round(s['duration']*60)),'-c:v','libx264','-preset','veryfast','-crf','19','-threads','2','-pix_fmt','yuv420p','-video_track_timescale','60000',str(target)]
            rng=random.Random(126); stars=[(rng.randrange(1920),rng.randrange(1080),rng.randrange(1,3),rng.randrange(25,70)) for _ in range(100)]
            p=subprocess.Popen(command,stdin=subprocess.PIPE)
            try:
                for n in range(round(s['duration']*60)):
                    im=flags.copy(); d=ImageDraw.Draw(im); t=n/60
                    for x,y,r,speed in stars:
                        xx=int((x+t*speed/8)%1920); yy=int((y+t*speed)%1080); d.ellipse((xx,yy,xx+r,yy+r),fill=(155,165,185))
                    im.paste(overlay,(0,0),overlay); p.stdin.write(im.tobytes())
                p.stdin.close(); assert p.wait(timeout=60)==0
            finally:
                if p.poll() is None: p.terminate(); p.wait(timeout=10)
        else:
            wide=s['layout']=='wide'; overlay=Image.new('RGBA',flags.size); d=ImageDraw.Draw(overlay)
            left,right=(240,1679) if wide else (0,1919)
            d.rectangle((left,0,right,59),fill=(4,8,16,255)); d.rectangle((left,1023,right,1079),fill=(4,8,16,255))
            text(overlay,(960,29),s['title'],32,(left+12,0,right-12,60),plan,True)
            text(overlay,(960,1050),s['caption'],26,(left+12,1025,right-12,1079),plan)
            path=WORK/f'overlay-{s["id"]:02}.png'; overlay.save(path)
            size,position=('1920:1080','0:0') if wide else (('1440:810','240:105') if s['layout']=='anime' else ('1280:960','320:60'))
            run('ffmpeg','-nostdin','-v','error','-loop','1','-framerate','60','-i',WORK/'BACKGROUND.png','-ss',s['start'],'-i',s['source'],'-loop','1','-i',path,
                '-filter_complex_threads','1','-filter_complex',f'[1:v]scale={size}:flags=neighbor,setsar=1,fps=60,setpts=PTS-STARTPTS[g];[0:v][g]overlay={position}:shortest=1[b];[b][2:v]overlay=0:0:shortest=1[out]',
                '-map','[out]','-an','-frames:v',round(s['duration']*60),'-r','60','-c:v','libx264','-preset','veryfast','-crf','19','-threads','2','-pix_fmt','yuv420p','-video_track_timescale','60000',target)
        s['clip_sha256']=sha(target); print('RENDER',s['id'],s['title'],flush=True)
    write(WORK/'PLAN.json',plan)
    (WORK/'concat.txt').write_text(''.join("file '"+str(WORK/f'scene-{s["id"]:02}.mp4')+"'\n" for s in plan['scenes']))
    soundtrack(plan)
    run('ffmpeg','-nostdin','-v','error','-f','concat','-safe','0','-i',WORK/'concat.txt','-i',WORK/'MIX.wav','-map','0:v:0','-map','1:a:0','-t',plan['seconds'],
        '-vf','fps=60','-r','60','-c:v','libx264','-preset','veryfast','-crf','18','-threads','2','-video_track_timescale','60000','-c:a','aac','-b:a','256k','-ar',RATE,'-ac','2','-movflags','+faststart',WORK/NAME)


def decode(path): return np.frombuffer(capture('ffmpeg','-nostdin','-v','error','-threads','2','-i',path,'-vn','-ar',RATE,'-ac','2','-f','f32le','-'),dtype='<f4').reshape(-1,2).copy()


def soundtrack(plan):
    count=round(plan['seconds']*RATE); music=decode(MUSIC); assert len(music)>=count
    music=music[:count]*6; fade=RATE; music[:fade]*=np.linspace(0,1,fade,dtype=np.float32)[:,None]; music[-fade:]*=np.linspace(1,0,fade,dtype=np.float32)[:,None]
    sfx=np.zeros_like(music); records={}; caches={}; sounds=[]
    for s in plan['scenes']:
        if s.get('layout')!='legacy': continue
        folder=Path(s['source']).parent
        if folder not in records:
            trace,rows=legacy.record(folder); events=[]; seen=set()
            for row in rows:
                pb=row.get('playback') or {}
                for cue in pb.get('sounds',[]):
                    key=(pb['id'],cue['id'],cue['tick'],cue.get('defence'))
                    if cue['played'] and key not in seen:
                        seen.add(key); events.append(dict(id=cue['id'],playback=pb['id'],start=(row['wall_ns']-trace['record_started_utc_ns'])/1e9))
            records[folder]=events
        for event in records[folder]:
            number=event['id']; source=SFX/f'SFX_{number:02}.ogg'
            if number not in caches: caches[number]=decode(source)
            sample=caches[number]; begin=max(s['start'],event['start']); end=min(s['start']+s['duration'],event['start']+len(sample)/RATE)
            if end<=begin: continue
            src=round((begin-event['start'])*RATE); dst=round((s['start_at']+begin-s['start'])*RATE); length=min(round((end-begin)*RATE),len(sample)-src,count-dst)
            assert dst>=0 and length>0; sfx[dst:dst+length]+=sample[src:src+length]*.7
            sounds.append(dict(scene=s['id'],id=number,output_sample=dst,source_sample=src,samples=length,playback=event['playback'],source=record(source)))
    assert sounds
    headroom=min(1,10**(-1.5/20)/float(np.max(np.abs(music+sfx)))); music*=headroom; sfx*=headroom
    for name,data in [('MUSIC.wav',music),('SFX.wav',sfx),('MIX.wav',music+sfx)]:
        with wave.open(str(WORK/name),'wb') as f: f.setnchannels(2); f.setsampwidth(2); f.setframerate(RATE); f.writeframes(np.round(np.clip(data,-1,1)*32767).astype('<i2').tobytes())
    write(WORK/'SOUNDTRACK.json',dict(music_track_count=1,music_loops=0,music_start_sample=0,music_samples=count,gain=6*headroom,fades_seconds=[1,1],original_clip_audio_mixed=False,sfx_events=sounds,music_sha256=sha(MUSIC),reference_sha256=sha(WORK/'MIX.wav')))


def audit():
    plan=read(WORK/'PLAN.json'); validate(plan); movie=WORK/NAME; data=probe(movie); v=next(s for s in data['streams'] if s['codec_type']=='video'); a=next(s for s in data['streams'] if s['codec_type']=='audio')
    assert (v['width'],v['height'],v['avg_frame_rate'])==(1920,1080,'60/1') and int(v['nb_frames'])==round(plan['seconds']*60)
    assert a['channels']==2 and abs(float(v['duration'])-plan['seconds'])<.02 and abs(float(a['duration'])-plan['seconds'])<.05
    sound=read(WORK/'SOUNDTRACK.json'); assert sound['music_track_count']==1 and sound['music_loops']==0
    for event in sound['sfx_events']: assert sha(ROOT/event['source']['path'])==event['source']['sha256']
    got=decode(movie); ref=decode(WORK/'MIX.wav'); n=min(len(got),len(ref)); got=got[:n]; ref=ref[:n]
    snr=lambda x,y:float(10*np.log10(np.mean(y*y)/np.mean((x-y)**2)))
    rms=float(np.sqrt(np.mean(got*got))); peak=float(np.max(np.abs(got))); score=snr(got,ref); assert rms>.01 and peak<.999 and score>25,(rms,peak,score)
    cuts=[]
    for s in plan['scenes'][1:]:
        t=round(s['start_at']*RATE); score_at=snr(got[t-4410:t+4410],ref[t-4410:t+4410]); assert score_at>22
        cuts.append(dict(seconds=s['start_at'],snr_db=score_at))
    detected=subprocess.run(['ffmpeg','-nostdin','-hide_banner','-threads','2','-i',str(movie),'-an','-vf','blackdetect=d=0.25:pix_th=0.10:pic_th=0.98,freezedetect=n=0.002:d=2','-f','null','-'],capture_output=True,text=True,check=True,timeout=300)
    (WORK/'DETECT.log').write_text(detected.stderr)
    import re
    black=re.findall(r'black_start:[^\n]+',detected.stderr); assert not black
    sheet=Image.new('RGB',(1920,math.ceil(len(plan['scenes'])/4)*305),'#101924'); draw=ImageDraw.Draw(sheet); samples=[]
    for i,s in enumerate(plan['scenes']):
        at=s['start_at']+s['duration']/2; frame=WORK/f'frame-{i:02}.png'
        run('ffmpeg','-nostdin','-v','error','-threads','2','-ss',at,'-i',movie,'-frames:v','1','-threads','1',frame)
        im=Image.open(frame); im.thumbnail((480,270)); x=(i%4)*480; y=(i//4)*305; sheet.paste(im,(x,y)); draw.text((x+7,y+275),f'{i+1:02} '+s['title'],font=ImageFont.truetype(str(FONT),18,index=3),fill='#e4eaf0')
        if s['kind']!='story':
            samples.append(dict(scene=i,source=s['source'],source_seconds=s['start']+s['duration']/2,movie_seconds=at,frame=record(frame),provenance=s['provenance']))
    sheet.save(WORK/'CONTACT.png'); write(WORK/'FFPROBE.json',data)
    write(WORK/'SOURCE-PLAYBACK.json',dict(passed=True,source=record(OLD/'SOURCE-PLAYBACK.json'),selected=plan['selected_previous'],all_three_weapons=True,complete_original_speed=True,scope='逐段回查既有passed、無漏格及已見結果的完整演出；不聲稱全片新拍或DOS同態'))
    write(WORK/'SAMPLES.json',samples)
    value=dict(passed=True,version=VERSION,file=NAME,sha256=sha(movie),bytes=movie.stat().st_size,seconds=plan['seconds'],
        video=dict(codec=v['codec_name'],width=v['width'],height=v['height'],fps=v['avg_frame_rate'],frames=v['nb_frames']),
        audio=dict(codec=a['codec_name'],sample_rate=a['sample_rate'],channels=2,rms=rms,peak=peak,snr_db=score,cut_checks=cuts,sfx_events=len(sound['sfx_events']),listening='未作人耳驗收'),
        black_segments=black,freeze_detections=re.findall(r'freeze_(?:start|end|duration): [\d.]+',detected.stderr),freeze_classification='國徽底圖與字幕停留含輕微星點，其餘為與來源相同的介面或動畫停留；實戰沒有替換靜圖或延長影格。',
        text_safe_bounds=len(plan['text_bounds']),subtitle_clipping=False,full_canvas=True,story_sources=plan['story_sources'],storyboard=plan['storyboard'],
        source_audio_sha256=sha(MUSIC),plan_sha256=sha(WORK/'PLAN.json'),contact_sha256=sha(WORK/'CONTACT.png'),recording_scope=plan['scope'],visual_review='待目視國徽、故事銜接、主角動畫、操作與實戰完整抽樣',
        generator=record(Path(__file__).resolve()),ffmpeg_version=capture('ffmpeg','-version').decode().splitlines()[0])
    write(WORK/'AUDIT.json',value); print('PASS',plan['seconds'],'秒，SNR',score,'音效',len(sound['sfx_events']),'處',flush=True)


def publish():
    plan=read(WORK/'PLAN.json'); validate(plan); a=read(WORK/'AUDIT.json'); assert a['passed'] and a['visual_review'].startswith('已目視')
    assert sha(WORK/NAME)==a['sha256'] and sha(WORK/'PLAN.json')==a['plan_sha256']
    for row in read(PREVIOUS/'MAPPING.json')['files']:
        if row['original'].startswith(str(OUT.relative_to(ROOT))+'/'): assert sha(ROOT/row['original'])==sha(ROOT/row['backup'])==row['sha256']
    names=[NAME,'PLAN.json','AUDIT.json','CONTACT.png','SOUNDTRACK.json','FFPROBE.json','SOURCE-PLAYBACK.json','SAMPLES.json']
    for name in names:
        temp=OUT/(name+'.tmp'); assert not temp.exists(); shutil.copyfile(WORK/name,temp); os.replace(temp,OUT/name)
    (OUT/'SHA256SUMS.txt').write_text(''.join(sha(OUT/n)+'  '+n+'\n' for n in names))
    print('INSTALLED',OUT,'，初版已保留於previous映射',flush=True)


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__); parser.add_argument('mode',choices=['prepare','render','audit','publish']); args=parser.parse_args()
    globals()[args.mode]()
