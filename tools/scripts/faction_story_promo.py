"""第126輪：HD雙旗故事片頭、正式封包實錄與同局面AI比較。"""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import random
import subprocess

from PIL import Image, ImageDraw, ImageFont, ImageChops, ImageStat

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('action',choices=['prepare','assemble','audit'])
p.add_argument('--root',default='/work/hd126/promo-r1')
p.add_argument('--version',default='v.1.1.2-20261011')
p.add_argument('--plan')
p.add_argument('--flag-background',help='可選已檢視的HD雙旗原PNG，保留來源不覆寫')
p.add_argument('--resume',action='store_true',help='比對前一PLAN與輸出雜湊，只重編已修訂場景')
p.add_argument('--fps',type=int,choices=[30,60],default=30)
p.add_argument('--label',help='預覽字幕；未定版時避免把中間影片稱為正式發行')
p.add_argument('--movie-name',help='可選成片檔名，不含目錄')
a=p.parse_args()
root=Path(a.root)
package=Path('/src/dist-all')/a.version/'full'/('LOGH3-'+a.version+'-linux-amd64')
music=package/'music/BGM_0F.ogg'
font=Path('/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc')
bold=Path('/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc')
assert os.getuid()==1000 and root.parent.stat().st_uid==1000
assert not a.movie_name or Path(a.movie_name).name==a.movie_name
movie=root/(a.movie_name or ('LOGH3-'+a.version+'-promo.mp4'))
bounds=[]

def write(path,obj):path.write_text(json.dumps(obj,ensure_ascii=False,indent=2)+'\n')
def sha(path):return hashlib.sha256(path.read_bytes()).hexdigest()
def run(*args,**kw):return subprocess.run([str(v) for v in args],check=True,**kw)
def ff(*args,**kw):run('ffmpeg','-nostdin','-y','-v','warning',*args,**kw)
def codec():return ['-c:v','libx264','-preset','veryfast','-crf','18','-threads','2','-pix_fmt','yuv420p','-r',str(a.fps),'-c:a','aac','-b:a','192k','-ar','44100','-ac','2','-movflags','+faststart']

def text(im,xy,content,size=34,colour='#fff0d1',weight=False,width=None):
 d=ImageDraw.Draw(im);f=ImageFont.truetype(str(bold if weight else font),size)
 box=d.textbbox(xy,content,font=f);assert box[0]>=0 and box[1]>=0 and box[2]<=1920 and box[3]<=1080,(content,box)
 if width is not None:assert box[2]-box[0]<=width,(content,box,width)
 d.text(xy,content,font=f,fill=colour,stroke_width=0)
 bounds.append(dict(text=content,bounds=list(box)))

def background():return Image.open(root/'BACKGROUND.png').convert('RGB')

def animate(path,ending=False):
 seconds=6 if ending else 12
 command=['ffmpeg','-nostdin','-y','-v','warning','-f','rawvideo','-pixel_format','rgb24','-video_size','1920x1080','-framerate',str(a.fps),'-i','pipe:0',
          '-ss','12' if ending else '0','-i',str(music),'-map','0:v','-map','1:a','-t',str(seconds),
          '-af',f'volume=0.8,afade=t=in:st=0:d=0.5,afade=t=out:st={seconds-.6}:d=0.6',*codec(),str(path)]
 rng=random.Random(126);stars=[(rng.randrange(1920),rng.randrange(1080),rng.randrange(1,4),rng.randrange(25,90)) for _ in range(100)]
 cards=[]
 captions=[('百餘年的對峙，戰火仍未熄滅。','一個銀河，兩種未來。'),
           ('遠征軍深入帝國領土。','補給線遭截斷，亞姆力札的命運即將揭曉。'),
           ('如果這一次，換你下令？','改變歷史，扭轉局勢。')]
 if ending:captions=[('下一頁歷史，由你寫下。','銀河英雄傳說 III SP')]
 for headline,subtitle in captions:
  im=Image.new('RGBA',(1920,1080));text(im,(90,75),'銀河帝國',52,'#ead6a2',True,width=780)
  text(im,(1130,75),'自由行星同盟',52,'#edb15d',True,width=700)
  text(im,(120,650),headline,62,'#fff5dc',True,width=1690)
  text(im,(125,760),subtitle,40,'#e9d7b6',width=1680)
  text(im,(125,965),(a.label or a.version+'　Windows / Linux / macOS'),28,'#e5d1a6',width=1660)
  cards.append(im)
 with (root/(path.stem+'-encode.log')).open('x') as log:
  proc=subprocess.Popen(command,stdin=subprocess.PIPE,stdout=log,stderr=subprocess.STDOUT)
  try:
   base=background()
   for n in range(seconds*a.fps):
    t=n/a.fps;im=base.copy();d=ImageDraw.Draw(im)
    for x,y,r,s in stars:
     yy=int((y+t*s)%1080);xx=int((x+t*(s/8))%1920)
     light=int(145+70*math.sin(t*1.4+x));d.ellipse((xx,yy,xx+r,yy+r),fill=(light,light,min(255,light+22)))
    d.line((960,0,960,1080),fill='#cba35d',width=3)
    index=0 if ending else min(2,n//(4*a.fps))
    overlay=cards[index];alpha=min(1,(n%(4*a.fps)+1)/(a.fps/2)) if not ending else min(1,(n+1)/(a.fps/2))
    im.paste(overlay,(0,0),overlay.getchannel('A').point(lambda v:round(v*alpha)))
    proc.stdin.write(im.tobytes())
   proc.stdin.close();assert proc.wait(timeout=60)==0
  finally:
   if proc.poll() is None:proc.terminate();proc.wait(timeout=10)

def prepare():
 assert not root.exists();root.mkdir()
 flags=Path('/work/hd123-s1-20261011');sources=json.loads((flags/'FLAG-SOURCES.json').read_text())
 out=Image.new('RGB',(1920,1080))
 for faction,x,box in [('empire',0,(20,90,480,404)),('alliance',960,(40,88,454,406))]:
  path=flags/(faction+'-flag.jpg');entry=next(s for s in sources if s['file']==path.name);assert sha(path)==entry['sha256']
  im=Image.open(path).convert('RGB').crop(box);im.save(root/(faction+'-flag.png'))
  wall=im.resize((960,1080),Image.Resampling.LANCZOS);out.paste(wall,(x,0))
 if a.flag_background:
  source=Path(a.flag_background)
  out=Image.open(source).convert('RGB').resize((1920,1080),Image.Resampling.LANCZOS)
  (root/'FLAGS-HD-ORIGINAL.png').write_bytes(source.read_bytes())
 out=Image.blend(out,Image.new('RGB',out.size,'#030711'),.50 if a.flag_background else .64);out.save(root/'BACKGROUND.png')
 animate(root/'INTRO.mp4');animate(root/'ENDING.mp4',True)
 write(root/'PREPARE.json',dict(version=a.version,flags=sources,music=dict(path=str(music),sha256=sha(music)),
       fonts=[dict(path=str(f),sha256=sha(f)) for f in (font,bold)],text_bounds=bounds,
       animated_intro=f'原生1920×1080文字、雙旗與逐幀星點；片頭{12*a.fps}幀，戰鬥另取正常實錄',
       flag_background=None if not a.flag_background else dict(path=a.flag_background,sha256=sha(Path(a.flag_background)),method='以原旗幟照片作身份參考的imagegen高清重建，非官方原圖逐像素放大'),
       story_sources=['https://www.production-ig.co.jp/works/ginei/story.html','https://ginei.jp/story1.php'],
       rights='雙旗、遊戲美術及配樂為原版衍生素材，只留本機及授權私有範圍'))

def assemble():
 scenes=json.loads(Path(a.plan).read_text());timeline=0;entries=[];parts=[root/'INTRO.mp4'];timeline=12
 previous=json.loads((root/'PLAN.json').read_text())['storyboard'] if a.resume else []
 for i,s in enumerate(scenes):
  assert s['kind'] in ('game','compare','still') and s['seconds']>=4
  card=background();d=ImageDraw.Draw(card)
  if s['kind']=='compare':
   text(card,(60,45),s['title'],48,weight=True,width=1800)
   text(card,(20,165),'原版規則模式',32,width=920);text(card,(980,165),'強化 AI',32,'#f0b26b',width=920)
   text(card,(40,950),s['subtitle'],30,width=1840)
   text(card,(40,1010),'兩側皆為 remake 實錄，從相同初始六表開始。',26,'#dbc59d',width=1820)
  else:
   text(card,(330,8),s['title'],32,weight=True,width=1270)
   text(card,(330,1019),s['subtitle'],27,width=1260)
   for faction,x,label in [('empire',16,'銀河帝國'),('alliance',1616,'自由行星同盟')]:
    flag=Image.open(root/(faction+'-flag.png')).convert('RGB');flag.thumbnail((288,220),Image.Resampling.LANCZOS);card.paste(flag,(x,410))
    text(card,(x+10,350),label,28,'#ead5a0',width=285)
  card_path=root/f'CARD-{i:02}.png';card.save(card_path)
  output=root/f'SCENE-{i:02}.mp4'
  if i<len(previous):
   old=previous[i];old_input={k:v for k,v in old.items() if k not in ('timeline_start','output','output_sha256')}
   if s==old_input:
    assert sha(output)==old['output_sha256'],output
    entries.append(dict(s,timeline_start=timeline,output=str(output),output_sha256=sha(output)))
    timeline+=s['seconds'];parts.append(output);continue
  command=['-loop','1','-framerate',str(a.fps),'-i',str(card_path)]
  if s['kind']=='game':
   raw=Path(s['video']);assert raw.is_file();command+=['-ss',str(s['start']),'-i',str(raw)]
   filters='[1:v]scale=1280:960:flags=lanczos,setpts=PTS-STARTPTS[g];[0:v][g]overlay=320:60:shortest=1[v]'
   audio='1:a'
  elif s['kind']=='compare':
   for side in ('left','right'):command+=['-ss',str(s[side]['start']),'-i',s[side]['video']]
   filters='[1:v]scale=920:690:flags=lanczos,setpts=PTS-STARTPTS[l];[2:v]scale=920:690:flags=lanczos,setpts=PTS-STARTPTS[r];[0:v][l]overlay=20:230[t];[t][r]overlay=980:230:shortest=1[v]'
   audio='2:a'
  else:
   im=Image.open(s['image']).convert('RGB').resize((1280,960),Image.Resampling.LANCZOS);card.paste(im,(320,60));card.save(card_path)
   command+=['-ss',str(timeline%20),'-i',str(music)];filters='[0:v]null[v]';audio='1:a'
  command+=['-filter_complex_threads','1','-filter_complex',filters,'-map','[v]','-map',audio,'-t',str(s['seconds']),
            '-af',f'volume=0.8,aresample=async=1:first_pts=0,afade=t=in:st=0:d=0.06,afade=t=out:st={s["seconds"]-.1}:d=0.1',*codec(),str(output)]
  with (root/f'SCENE-{i:02}-encode.log').open('w' if a.resume else 'x') as log:ff(*command,stdout=log,stderr=subprocess.STDOUT)
  entries.append(dict(s,timeline_start=timeline,output=str(output),output_sha256=sha(output)));timeline+=s['seconds'];parts.append(output)
  print('完成場景',i,s['title'],flush=True)
 parts.append(root/'ENDING.mp4');timeline+=6
 concat=root/'CONCAT.txt';concat.write_text(''.join("file '"+str(f)+"'\n" for f in parts))
 ff('-f','concat','-safe','0','-i',concat,'-c:v','copy','-c:a','aac','-b:a','192k',
    '-af','aresample=async=1:first_pts=0','-movflags','+faststart',movie)
 write(root/'PLAN.json',dict(version=a.version,label=a.label,fps=a.fps,duration=timeline,intro_seconds=12,ending_seconds=6,storyboard=entries,text_bounds=bounds,
  audio='遊戲段及AI比較保留同次正式程式錄音；比較音軌取右側強化模式。首尾及介面展示沿正式遊戲OGG。',
  inputs=[dict(path=str(f),sha256=sha(f)) for f in set([Path(s['video']) for s in scenes if s['kind']=='game']+[Path(s[k]['video']) for s in scenes if s['kind']=='compare' for k in ('left','right')])]))

def audit():
 plan=json.loads((root/'PLAN.json').read_text());probe=json.loads(subprocess.check_output(['ffprobe','-v','error','-show_format','-show_streams','-of','json',str(movie)]))
 streams={s['codec_type']:s for s in probe['streams']};v=streams['video'];assert (v['width'],v['height'],v['r_frame_rate'])==(1920,1080,f'{a.fps}/1')
 assert streams['audio']['channels']==2 and abs(float(probe['format']['duration'])-plan['duration'])<.3
 with (root/'VIDEO-AUDIT.txt').open('x') as log:
  run('ffmpeg','-nostdin','-v','info','-i',movie,'-vf','blackdetect=d=0.2:pix_th=0.05,freezedetect=n=0.001:d=2','-an','-f','null','-',stdout=log,stderr=subprocess.STDOUT)
 # volumedetect需要info層級，另存音訊量測。
 with (root/'AUDIO-AUDIT.txt').open('x') as log:
  run('ffmpeg','-nostdin','-v','info','-i',movie,'-vn','-af','volumedetect','-f','null','-',stdout=log,stderr=subprocess.STDOUT)
 import re
 volume=(root/'AUDIO-AUDIT.txt').read_text();mean=float(re.search(r'mean_volume: ([-.0-9]+)',volume).group(1));peak=float(re.search(r'max_volume: ([-.0-9]+)',volume).group(1))
 assert -55<mean<-5 and peak<0,(mean,peak)
 video_log=(root/'VIDEO-AUDIT.txt').read_text()
 black=[dict(start=float(x),end=float(y),seconds=float(z)) for x,y,z in re.findall(r'black_start:([0-9.]+) black_end:([0-9.]+) black_duration:([0-9.]+)',video_log)]
 assert not black,black
 starts=[float(x) for x in re.findall(r'freeze_start: ([0-9.]+)',video_log)]
 ends=[float(x) for x in re.findall(r'freeze_end: ([0-9.]+)',video_log)]
 if len(ends)<len(starts):ends.append(plan['duration'])
 assert len(ends)==len(starts)
 freezes=[];extra_offsets={}
 for start,end in zip(starts,ends):
  scenes=[(i,s) for i,s in enumerate(plan['storyboard']) if s['timeline_start']<end and s['timeline_start']+s['seconds']>start]
  if not scenes:
   assert end<=12.05 or start>=plan['duration']-6,(start,end)
   frames=[]
   for n,t in enumerate((start+.25,min(end-.25,plan['duration']-.1))):
    path=root/f'ANIMATION-CHECK-{len(freezes)}-{n}.png'
    ff('-ss',str(t),'-i',movie,'-frames:v','1','-threads','2',path)
    frames.append(Image.open(path).convert('RGB'))
   assert ImageChops.difference(*frames).getbbox() is not None
   freezes.append(dict(start=start,end=end,scenes=[],classification='片頭／片尾字幕停留，星點動畫兩幀確有變化'))
   continue
  for i,s in scenes:
   if s['kind']!='still':
    # 同一實錄若也停在輸入畫面，即為玩家停留；增加該凍結中點的來源核對。
    mid=(max(start,s['timeline_start'])+min(end,s['timeline_start']+s['seconds']))/2-s['timeline_start']
    extra_offsets.setdefault(i,[]).append(round(mid,3))
  freezes.append(dict(start=start,end=end,scenes=[i for i,s in scenes],classification='介面靜圖，或與來源實錄相同的輸入停留；逐項抽核來源中點'))
 samples=[]
 for i,s in enumerate(plan['storyboard']):
  if s['kind']=='still':continue
  for offset in sorted(set([1,s['seconds']-1]+extra_offsets.get(i,[]))):
   for side,box,size in ([('',(320,60,1280,960),(1280,960))] if s['kind']=='game' else [('left',(20,230,920,690),(920,690)),('right',(980,230,920,690),(920,690))]):
    source=s if not side else s[side];raw=root/f'CHECK-{i}-{offset}-{side}-raw.png';actual=root/f'CHECK-{i}-{offset}-{side}-movie.png'
    ff('-ss',str(source['start']+offset),'-i',source['video'],'-vf',f'scale={size[0]}:{size[1]}:flags=lanczos','-frames:v','1','-threads','2',raw)
    ff('-ss',str(s['timeline_start']+offset),'-i',movie,'-vf',f'crop={box[2]}:{box[3]}:{box[0]}:{box[1]}','-frames:v','1','-threads','2',actual)
    d=ImageChops.difference(Image.open(raw).convert('RGB'),Image.open(actual).convert('RGB'));error=sum(ImageStat.Stat(d).mean)/3
    assert error<4,(i,offset,side,error);d.save(root/f'CHECK-{i}-{offset}-{side}-diff.png')
    samples.append(dict(scene=i,offset=offset,side=side,mean_channel_error=error))
 write(root/'AUDIT.json',dict(passed=True,version=a.version,label=a.label,duration=float(probe['format']['duration']),video=[1920,1080,f'{a.fps}fps'],
   audio=dict(mean_db=mean,peak_db=peak,channels=2,gameplay_source='同次PulseAudio遊戲配樂及音效'),source_frame_samples=samples,
   black_segments=black,freeze_segments=freezes,movie_sha256=sha(movie),rights='本機私有；未公開發布',not_claimed=['DOS全戰場同態','AI必勝','Windows／macOS真機','人耳驗收']))
 write(root/'FFPROBE.json',probe);print('PASS 成片影音及實錄來源抽樣',len(samples),mean,peak,flush=True)

{'prepare':prepare,'assemble':assemble,'audit':audit}[a.action]()
