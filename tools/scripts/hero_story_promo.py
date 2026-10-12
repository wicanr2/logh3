"""規格140：雙旗故事、指定1988動畫與有聲實戰；只在Docker本機製片。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

from PIL import Image, ImageChops, ImageDraw, ImageEnhance, ImageFont, ImageStat

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('mode',choices=['prepare','encode','audit'])
p.add_argument('--plan',required=True)
p.add_argument('--out',required=True)
a=p.parse_args()
plan=json.loads(Path(a.plan).read_text())
out=Path(a.out)
assert os.getuid()==1000 and out.parent.stat().st_uid==1000
font='/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc'
bold='/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc'
def sha(path):return hashlib.sha256(Path(path).read_bytes()).hexdigest()
def write(name,data):(out/name).write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
def run(*args,**kwargs):return subprocess.run([str(v) for v in args],check=True,timeout=600,**kwargs)
def text(draw,xy,value,size=30,fill='#f7e7bb',width=1824):
    f=ImageFont.truetype(bold if size>=40 else font,size)
    box=draw.textbbox(xy,value,font=f)
    assert box[0]>=0 and box[2]<=1920 and box[1]>=0 and box[3]<=1080 and box[2]-box[0]<=width,(value,box)
    draw.text(xy,value,font=f,fill=fill,stroke_width=1,stroke_fill='#08101a')
    return {'text':value,'box':box}

if a.mode=='prepare':
    assert not out.exists();out.mkdir()
    background=Image.open(plan['background']).convert('RGB')
    assert background.size==(1920,1080)
    ImageEnhance.Brightness(background).enhance(.7).save(out/'BACKGROUND.png')
    bounds=[];timeline=0;inputs={}
    for n,scene in enumerate(plan['scenes']):
        scene.update(id=n,start_at=timeline);timeline+=scene['duration']
        overlay=Image.new('RGBA',(1920,1080));d=ImageDraw.Draw(overlay)
        d.rectangle((0,0,1919,59),fill=(4,8,16,225));d.rectangle((0,1020,1919,1079),fill=(4,8,16,235))
        bounds.append(text(d,(48,8),scene['title'],30))
        bounds.append(text(d,(48,1027),scene['caption'],28))
        overlay.save(out/f'overlay-{n:02}.png')
        if scene['kind']=='card':
            card=background.copy();card=ImageEnhance.Brightness(card).enhance(.6);c=ImageDraw.Draw(card)
            if scene.get('poster'):
                poster=Image.open(scene['poster']).convert('RGB')
                for box,xy in [((60,45,555,555),(285,170)),((635,45,1130,555),(1115,170))]:
                    portrait=poster.crop(box).resize((520,535),Image.Resampling.LANCZOS)
                    card.paste(portrait,xy)
                bounds.append(text(c,(455,730),'銀河帝國',50,width=500))
                bounds.append(text(c,(1160,730),'自由行星同盟',50,width=600))
                bounds.append(text(c,(570,850),'兩大勢力，戰火橫越星海',54,width=1100))
                inputs[scene['poster']]=sha(scene['poster'])
            else:
                bounds.append(text(c,(360,300),'銀河英雄傳說 III SP',76,width=1400))
                bounds.append(text(c,(455,460),'由你下令，讓歷史轉向',62,width=1300))
                bounds.append(text(c,(525,620),'原版操作 × 陣營 HD × 強化 AI',46,width=1200))
                bounds.append(text(c,(600,750),plan['version'],38,width=1000))
            card.save(out/f'card-{n:02}.png')
        else:
            inputs[scene['file']]=sha(scene['file'])
    inputs[plan['background']]=sha(plan['background']);inputs[plan['music']]=sha(plan['music'])
    write('PLAN.json',dict(plan,seconds=timeline,inputs=inputs))
    write('TEXT-BOUNDS.json',bounds)
    print('READY',timeline,'秒',len(plan['scenes']),'段',flush=True)

elif a.mode=='encode':
    plan=json.loads((out/'PLAN.json').read_text())
    assert all(sha(name)==digest for name,digest in plan['inputs'].items())
    for s in plan['scenes']:
        target=out/f'scene-{s["id"]:02}.mp4'
        if target.exists():continue
        duration=str(s['duration']);overlay=out/f'overlay-{s["id"]:02}.png'
        if s['kind']=='card':
            frames=round(s['duration']*60)
            command=['ffmpeg','-nostdin','-v','warning','-loop','1','-i',out/f'card-{s["id"]:02}.png',
                     '-loop','1','-i',overlay,'-stream_loop','-1','-i',plan['music'],
                     '-filter_complex',f"[0:v]zoompan=z='1+0.00007*on':d={frames}:x='iw/2-iw/zoom/2':y='ih/2-ih/zoom/2':s=1920x1080:fps=60[v];[v][1:v]overlay=0:0:shortest=1[out];[2:a]volume=0.5[a]",
                     '-map','[out]','-map','[a]']
        else:
            size='1440:810' if s['kind']=='anime' else '1280:960'
            x,y=('240','105') if s['kind']=='anime' else ('320','60')
            command=['ffmpeg','-nostdin','-v','warning','-loop','1','-framerate','60','-i',out/'BACKGROUND.png',
                     '-ss',str(s['start']),'-i',s['file'],'-loop','1','-i',overlay,
                     '-stream_loop','-1','-ss',str(s['start_at']%70),'-i',plan['music'],
                     '-filter_complex',f'[1:v]scale={size}:flags=lanczos,setsar=1,fps=60[v];[0:v][v]overlay={x}:{y}:shortest=1[c];[c][2:v]overlay=0:0:shortest=1[out];[1:a]volume=0.95[a1];[3:a]volume=0.1[a2];[a1][a2]amix=inputs=2:duration=first:normalize=0[a]',
                     '-map','[out]','-map','[a]']
        command += ['-t',duration,'-c:v','libx264','-preset','veryfast','-crf','19','-threads','2',
                    '-filter_complex_threads','1','-pix_fmt','yuv420p','-c:a','aac','-b:a','192k','-ar','44100','-ac','2','-movflags','+faststart',target]
        with (out/f'encode-{s["id"]:02}.log').open('w') as log:run(*command,stdout=log,stderr=subprocess.STDOUT)
        print('ENCODED',s['id'],s['kind'],flush=True)
    # Explicit video duration prevents each independently padded AAC packet from
    # shifting subsequent scene cuts. Resample the concatenated audio clock.
    listing=out/'CONCAT.txt';listing.write_text(''.join(f"file 'scene-{s['id']:02}.mp4'\nduration {s['duration']:.6f}\n" for s in plan['scenes']))
    run('ffmpeg','-nostdin','-v','warning','-f','concat','-safe','0','-i',listing,
        '-t',str(plan['seconds']),'-c:v','copy','-af','aresample=async=1:first_pts=0',
        '-c:a','aac','-b:a','192k','-movflags','+faststart',out/'LOGH3-HERO-PROMO.mp4')

else:
    plan=json.loads((out/'PLAN.json').read_text());movie=out/'LOGH3-HERO-PROMO.mp4'
    probe=json.loads(subprocess.check_output(['ffprobe','-v','error','-show_streams','-show_format','-of','json',movie]))
    video=next(s for s in probe['streams'] if s['codec_type']=='video');audio=next(s for s in probe['streams'] if s['codec_type']=='audio')
    assert (video['width'],video['height'],video['r_frame_rate'],audio['channels'])==(1920,1080,'60/1',2)
    assert abs(float(probe['format']['duration'])-plan['seconds'])<.2
    with (out/'AUDIO-AUDIT.log').open('w') as log:run('ffmpeg','-nostdin','-v','info','-i',movie,'-vn','-af','volumedetect,silencedetect=n=-55dB:d=1','-f','null','-',stdout=log,stderr=subprocess.STDOUT)
    with (out/'VIDEO-AUDIT.log').open('w') as log:run('ffmpeg','-nostdin','-v','info','-i',movie,'-vf','blackdetect=d=.1:pix_th=.1,freezedetect=n=-50dB:d=2','-an','-f','null','-',stdout=log,stderr=subprocess.STDOUT)
    assert 'mean_volume: -inf' not in (out/'AUDIO-AUDIT.log').read_text()
    assert 'silence_start:' not in (out/'AUDIO-AUDIT.log').read_text()
    assert 'black_start:' not in (out/'VIDEO-AUDIT.log').read_text()
    video_origin=float(video.get('start_time',0))
    contact=Image.new('RGB',(4*480,((len(plan['scenes'])+3)//4)*270),'#09111d')
    for i,s in enumerate(plan['scenes']):
        raw=subprocess.check_output(['ffmpeg','-v','error','-ss',str(video_origin+s['start_at']+s['duration']/2),'-i',movie,'-frames:v','1','-vf','scale=480:270','-pix_fmt','rgb24','-f','rawvideo','-'])
        contact.paste(Image.frombytes('RGB',(480,270),raw),((i%4)*480,(i//4)*270))
    contact.save(out/'CONTACT.jpg')
    # Compare the content rectangle at the actual planned source time, before
    # overlay text. Compression and one capture-frame timing error are allowed.
    checks=[]
    def frame(path,at,filter):
        data=subprocess.check_output(['ffmpeg','-nostdin','-v','error','-ss',str(at),'-i',str(path),
            '-frames:v','1','-vf',filter,'-pix_fmt','rgb24','-f','rawvideo','-'])
        return Image.frombytes('RGB',(320,240),data)
    for s in plan['scenes']:
        if s['kind']=='card':continue
        times=[s['duration']/2]
        if s.get('battle_animation'):times=[.2,s['duration']/2,s['duration']-.2]
        for relative in times:
            crop='crop=1440:810:240:105' if s['kind']=='anime' else 'crop=1280:960:320:60'
            actual=frame(movie,video_origin+s['start_at']+relative,crop+',scale=320:240:flags=area')
            expected=frame(s['file'],s['start']+relative,'scale=320:240:flags=area')
            error=sum(ImageStat.Stat(ImageChops.difference(actual,expected)).mean)/3
            assert error<12,('來源時碼不符',s['id'],relative,error)
            checks.append(dict(scene=s['id'],relative=relative,mean_absolute_error=error,
                               source_seconds=s['start']+relative))
    write('SOURCE-FRAMES.json',dict(passed=True,checks=checks,video_origin_seconds=video_origin,
        method='實際成片內容矩形對應來源原速時碼；320×240面積取樣，平均通道差小於12'))
    write('AUDIT.json',dict(version=plan['version'],seconds=float(probe['format']['duration']),video_sha256=sha(movie),
        resolution=[1920,1080],fps=60,channels=2,non_silent=True,black_segments=0,
        battle_animation_seconds=sum(s['duration'] for s in plan['scenes'] if s.get('battle_animation')),
        anime_seconds=sum(s['duration'] for s in plan['scenes'] if s['kind']=='anime'),
        visual_review='待目視CONTACT與來源抽樣',rights='本機限定',
        limits=['指定動畫原生852×480，以原比例放大；不稱原生HD','輸出60fps不代表原作補間','未宣稱DOS整場同態或人耳驗收']))
    print('PASS 聲畫格式、長度、非靜音與零黑段；待目視',flush=True)
