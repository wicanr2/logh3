"""第139輪：全新遊戲實錄剪輯，整片單首連續配樂及獨立遊戲音效。Docker內執行。"""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import re
import shutil
import subprocess
import wave

import numpy as np
from PIL import Image, ImageChops, ImageStat

ROOT=Path('/src')
VERSION='v.1.4.0-20261015'
BATCH=ROOT/'workplace/promo139'
RATE=44100
MOVIE='LOGH3-HERO-PROMO.mp4'


def sha(path):
    h=hashlib.sha256()
    with Path(path).open('rb') as f:
        for chunk in iter(lambda:f.read(1024*1024),b''):h.update(chunk)
    return h.hexdigest()


def read(path):return json.loads(Path(path).read_text())
def write(path,data):Path(path).write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
def run(*args,**kw):return subprocess.run(list(map(str,args)),check=True,timeout=600,**kw)
def capture(*args):return subprocess.check_output(list(map(str,args)),timeout=120)
def decode(path):
    return np.frombuffer(capture('ffmpeg','-nostdin','-v','error','-threads','2','-i',path,
        '-vn','-ar',RATE,'-ac','2','-f','f32le','-'),dtype='<f4').reshape(-1,2).copy()


def record(folder):
    trace=read(folder/'TRACE.json')
    assert trace['normal_x11_input'] and not any(trace[k] for k in ['script_flag','state_injection','seed_injection'])
    assert sha(folder/'RAW.mp4')==trace['raw_sha256']
    assert sha(folder/'FRAMES.jsonl')==read(folder/'FRAME-TRACE-CLOSE.json')['frame_trace_sha256']
    rows=[json.loads(l) for l in (folder/'FRAMES.jsonl').open()]
    return trace,rows


def plan():
    scenes=[]
    def scene(kind,duration,title,caption,**kw):
        scenes.append(dict(kind=kind,duration=math.ceil(duration*60)/60,title=title,caption=caption,**kw))
    def game(folder,start,duration,title,caption,**kw):
        scene('game',duration,title,caption,file=str(folder/'RAW.mp4'),start=start,**kw)
    old=ROOT/'workplace/hero130'
    scene('card',6,'銀河英雄傳說 III SP｜v.1.4.0','帝國與同盟，下一道命令由你下達',poster=str(old/'anime-sources/remaster-poster.png'))
    for start,title in [(238,'萊因哈特｜向銀河的頂點進軍'),(778,'楊威利｜魔術師的下一步')]:
        scene('anime',3,title,'1988動畫來源，原生480p；本機推廣片',file=str(old/'ANIME-SOURCE-CLEAN.mp4'),start=start)
    inputs={};groups=[];unselected_anomalies=[]
    settings=BATCH/'compact-delivery-f0'
    ui_trace,_=record(settings)
    ui_captures={e['name']:(e['wall_begin_ns']-ui_trace['record_started_utc_ns'])/1e9 for e in ui_trace['events'] if e.get('kind')=='capture'}
    for name,duration,title,caption in [
        ('settings-on',3,'系統設定收進框內｜顯示敵軍可開關','敵軍預設關閉；此處手動開啟，轉移旗艦與名將移動加成預設開啟'),
        ('settings-off',2,'可選規則，由你決定','轉移旗艦與名將移動加成可分別關閉，退出重開仍保存'),
        ('settings-original',2,'按F7，即時切換原版／HD','縮小文字與同一對話框，原版／HD都能完整顯示'),
        ('ai-menu',3,'AI強化，在系統設定選擇','原版與強化0至3｜本次執行有效，數字越大越強')]:
        game(settings,ui_captures[name],duration,title,caption)
    for name in ['TRACE.json','FRAMES.jsonl','FRAME-TRACE-CLOSE.json','RESULT.json']:
        inputs[str(settings/name)]=sha(settings/name)
    for faction in [0,1]:
        folder=BATCH/('record-f0-final' if faction==0 else 'record-f1')
        trace,rows=record(folder)
        origin=trace['record_started_utc_ns']
        captures={e['name']:(e['wall_begin_ns']-origin)/1e9 for e in trace['events'] if e['kind']=='capture'}
        for name,duration,title,caption in (
            ('entry',4,'帝國的雄心' if faction==0 else '同盟的信念','本輪重新錄製S1正常操作，強化AI；沒有局面或seed注入'),
            ('turn-02',4,'戰線推進，名將協同','楊威利／萊因哈特帶隊參戰時，全陣營新移動額度1.5倍；僅強化模式')):
            game(folder,max(0,captures[name]-.1),duration,title,caption)
        fleet={}
        for row in rows:
            pb=row.get('playback') or {}
            if pb.get('kind')=='fleet':fleet.setdefault(pb['id'],[]).append(row)
        for identity,frames in fleet.items():
            if not frames[-1]['playback']['result']:
                unselected_anomalies.append(dict(source=str(folder),scene=identity,reason='來源未播完整，不選用'))
                continue
            applied=0
            skipped=[]
            for frame in frames:
                pb=frame['playback']
                assert pb['applied']==sum(p['tick']<=pb['tick'] for p in pb['pictures'])
                if pb['applied']>applied+1 and pb['interval']>0:skipped.append([applied,pb['applied']])
                applied=pb['applied']
            if skipped:unselected_anomalies.append(dict(source=str(folder),scene=identity,reason='來源漏格，未列為通過',gaps=skipped))
            start=math.floor((frames[0]['wall_ns']-origin)/1e9*60)/60
            end=math.ceil((frames[-1]['wall_ns']-origin)/1e9*60+2)/60
            groups.append((folder,start,end-start,identity,frames[-1]['playback']['weapon'],skipped))
        for name in ['TRACE.json','FRAMES.jsonl','FRAME-TRACE-CLOSE.json']:
            inputs[str(folder/name)]=sha(folder/name)
    fortress=BATCH/'record-fortress'
    trace,rows=record(fortress)
    origin=trace['record_started_utc_ns']
    gun=[r for r in rows if r.get('fortress')]
    assert gun and any(r['fortress']['Unit']>=0 and r['fortress']['Result'] for r in gun)
    first=math.floor((gun[0]['wall_ns']-origin)/1e9*60)/60
    last=math.ceil((gun[-1]['wall_ns']-origin)/1e9*60+2)/60
    game(fortress,first,last-first,'伊謝爾倫要塞砲｜發射至戰損','本輪正常操作試射；含友軍命中，保留完整演出與音效',fortress_animation=True)
    for name in ['TRACE.json','FRAMES.jsonl','FRAME-TRACE-CLOSE.json','RESULT.json']:
        inputs[str(fortress/name)]=sha(fortress/name)
    total=0
    for folder,start,duration,identity,weapon,skipped in groups:
        if total>=50 and {s['weapon'] for s in scenes if s.get('battle_animation')}=={0,1,2}:break
        assert not skipped,('成片選用的完整來源場次漏格，須重拍',identity,skipped)
        game(folder,start,duration,'光束交錯，飛彈破陣','完整交火原速實錄；背景配樂跨片段連續播放',battle_animation=True,playback_id=identity,weapon=weapon)
        total+=duration
    assert total>=50,'全新錄影未包含足夠完整交火'
    assert {s['weapon'] for s in scenes if s.get('battle_animation')}=={0,1,2},'須包含三種武器完整演出'
    scene('card',6,'你的戰術，決定銀河的明天','v.1.4.0｜Linux・Windows・macOS｜系統設定可選AI與規則')
    result=dict(version=VERSION,background=str(ROOT/'workplace/hd127/promo-final/BACKGROUND.png'),
        music=str(ROOT/'dist-all'/VERSION/'full'/('LOGH3-'+VERSION+'-linux-amd64')/'music/BGM_0F.ogg'),
        scenes=scenes,fresh_gameplay_inputs=inputs,rights='含原版及衍生音畫、指定動畫，只留本機，不公開上傳',
        unselected_source_anomalies=unselected_anomalies,
        playback_audit_scope='成片實際選用的完整艦隊與要塞場次；不宣稱整份來源錄影或完整遊戲無漏格',
        anime_source='https://www.bilibili.com/video/BV1G87szDEh7/',
        recording_scope='全部遊戲鏡頭為本輪正常入口新錄；設定為147後正式ZIP，交火為147排版修正前同規則候選；動畫／海報沿已核來源')
    assert sum(s['duration'] for s in scenes)<125,'不截斷結尾、不默默循環配樂'
    write(BATCH/'PROMO-PLAN-COMPACT.json',result)
    print('READY',len(scenes),'段',sum(s['duration'] for s in scenes),'秒；新交火',total,'秒',flush=True)


def render(out):
    data=read(out/'PLAN.json')
    for path,digest in {**data['inputs'],**data['fresh_gameplay_inputs']}.items():assert sha(path)==digest,path
    reuse=Path(args.reuse_dir) if args.reuse_dir else None
    old=read(reuse/'PLAN.json') if reuse else None
    reused=[]
    for scene in data['scenes']:
        target=out/f'scene-{scene["id"]:02}.mp4'
        assert not target.exists()
        overlay=out/f'overlay-{scene["id"]:02}.png'
        if old:
            signature={k:v for k,v in scene.items() if k not in ['id','start_at']}
            match=next((s for s in old['scenes'] if {k:v for k,v in s.items() if k not in ['id','start_at']}==signature),None)
            if match is not None:
                previous=reuse/f'scene-{match["id"]:02}.mp4'
                assert sha(reuse/f'overlay-{match["id"]:02}.png')==sha(overlay)
                if scene['kind']!='card':assert old['inputs'][scene['file']]==data['inputs'][scene['file']]
                streams=read_probe(previous)['streams']
                assert len(streams)==1 and streams[0]['codec_type']=='video'
                shutil.copyfile(previous,target)
                reused.append(dict(scene=scene['id'],source=str(previous),sha256=sha(previous),same_source_time_and_overlay=True))
                print('REUSED VIDEO',scene['id'],flush=True)
                continue
        if scene['kind']=='card':
            inputs=['-loop','1','-framerate','60','-i',out/f'card-{scene["id"]:02}.png','-loop','1','-i',overlay]
            filters='[0:v][1:v]overlay=0:0:shortest=1[out]'
        else:
            anime=scene['kind']=='anime'
            size,x,y=('1440:810',240,105) if anime else ('1280:960',320,60)
            inputs=['-loop','1','-framerate','60','-i',out/'BACKGROUND.png','-ss',scene['start'],'-i',scene['file'],'-loop','1','-i',overlay]
            filters=f'[1:v]scale={size}:flags=lanczos,setsar=1,fps=60[v];[0:v][v]overlay={x}:{y}:shortest=1[c];[c][2:v]overlay=0:0:shortest=1[out]'
        with (out/f'encode-{scene["id"]:02}.log').open('x') as log:
            run('ffmpeg','-nostdin','-v','warning',*inputs,'-filter_complex',filters,'-map','[out]','-an',
                '-frames:v',round(scene['duration']*60),'-c:v','libx264','-preset','veryfast','-crf','19','-threads','2',
                '-filter_complex_threads','1','-pix_fmt','yuv420p',target,stdout=log,stderr=subprocess.STDOUT)
        print('ENCODED',scene['id'],scene['kind'],flush=True)
    write(out/'REUSE.json',dict(video_only=True,segments=reused,original_clip_audio_mixed=False))
    listing=out/'CONCAT.txt'
    listing.write_text(''.join(f"file 'scene-{s['id']:02}.mp4'\nduration {s['duration']:.9f}\n" for s in data['scenes']))
    run('ffmpeg','-nostdin','-v','warning','-f','concat','-safe','0','-i',listing,'-c:v','copy','-an',out/'SILENT.mp4')
    count=round(data['seconds']*RATE)
    music=decode(data['music'])
    assert len(music)>=count,'單曲不足以覆蓋全片'
    music=music[:count]
    gain=10**(-20/20)/float(np.sqrt(np.mean(music.astype(np.float64)**2)))
    music*=gain
    music[:RATE]*=np.linspace(0,1,RATE,dtype=np.float32)[:,None]
    tail=round(2.5*RATE);music[-tail:]*=np.linspace(1,0,tail,dtype=np.float32)[:,None]
    sfx=np.zeros_like(music);caches={};records={};sounds=[]
    for scene in data['scenes']:
        if scene['kind']!='game':continue
        folder=Path(scene['file']).parent
        if folder not in records:
            trace,rows=record(folder);events=[];seen=set()
            for row in rows:
                pb=row.get('playback') or {}
                for cue in pb.get('sounds',[]):
                    key=(pb['id'],cue['id'],cue['tick'],cue.get('defence'))
                    if not cue['played'] or key in seen:continue
                    seen.add(key);events.append(dict(id=cue['id'],playback=pb['id'],start=(row['wall_ns']-trace['record_started_utc_ns'])/1e9))
            records[folder]=events
        for event in records[folder]:
            number=event['id'];path=Path(data['music']).parent/f'SFX_{number:02}.ogg'
            if number not in caches:caches[number]=decode(path)
            sample=caches[number]
            begin=max(scene['start'],event['start']);end=min(scene['start']+scene['duration'],event['start']+len(sample)/RATE)
            if end<=begin:continue
            source=round((begin-event['start'])*RATE);dest=round((scene['start_at']+begin-scene['start'])*RATE)
            length=min(round((end-begin)*RATE),len(sample)-source,count-dest)
            assert dest>=0 and length>0
            sfx[dest:dest+length]+=sample[source:source+length]*.7
            sounds.append(dict(scene=scene['id'],id=number,source_sample=source,output_sample=dest,samples=length,playback=event['playback'],source_sha256=sha(path)))
    assert sounds
    headroom=min(1,10**(-1.5/20)/float(np.max(np.abs(music+sfx))))
    music*=headroom;sfx*=headroom
    for name,array in [('MUSIC.wav',music),('SFX.wav',sfx),('MIX.wav',music+sfx)]:
        with wave.open(str(out/name),'wb') as f:
            f.setnchannels(2);f.setsampwidth(2);f.setframerate(RATE)
            f.writeframes(np.round(np.clip(array,-1,1)*32767).astype('<i2').tobytes())
    write(out/'SOUNDTRACK.json',dict(music_track_count=1,music_loops=0,music_start_sample=0,music_samples=count,
        gain=gain*headroom,fades_seconds=[1,2.5],original_clip_audio_mixed=False,sfx_events=sounds,
        method='全片單一連續音樂底軌；僅整片頭尾淡入淡出。遊戲音效依本輪played觀測重混。'))
    run('ffmpeg','-nostdin','-v','warning','-i',out/'SILENT.mp4','-i',out/'MIX.wav','-map','0:v:0','-map','1:a:0',
        '-c:v','copy','-c:a','aac','-b:a','256k','-ar',RATE,'-ac','2','-threads','2',
        '-movie_timescale',math.lcm(RATE,15360),'-movflags','+faststart',out/MOVIE)
    print('RENDERED 全新畫面、單曲連續配樂',flush=True)


def frame(path,seconds,filters):
    raw=capture('ffmpeg','-nostdin','-v','error','-threads','2','-ss',seconds,'-i',path,
        '-frames:v','1','-vf',filters,'-pix_fmt','rgb24','-f','rawvideo','-')
    return Image.frombytes('RGB',(320,240),raw)


def read_probe(path):
    return json.loads(capture('ffprobe','-v','error','-show_streams','-show_format','-of','json',path))


def source_audit(out):
    data=read(out/'PLAN.json');groups={}
    for scene in data['scenes']:
        if scene.get('battle_animation'):
            groups.setdefault(Path(scene['file']).parent,[]).append(scene['playback_id'])
    audits=[]
    for folder,ids in groups.items():
        writable=Path('/work')/folder.relative_to(ROOT/'workplace')
        with (out/(folder.name+'-source-audit.log')).open('x') as log:
            run('python3',ROOT/'tools/scripts/battle_playback_audit.py',writable,
                '--kinds','fleet','--scene-ids',*ids,stdout=log,stderr=subprocess.STDOUT)
        value=read(writable/'PLAYBACK-AUDIT.json')
        assert value['passed'] and not value['skipped_frames']
        audits.append(dict(source=str(folder),sha256=sha(writable/'PLAYBACK-AUDIT.json'),audit=value))
    fortress=BATCH/'record-fortress/PLAYBACK-AUDIT.json'
    value=read(fortress)
    assert value['passed'] and not value['skipped_frames']
    audits.append(dict(source=str(fortress.parent),sha256=sha(fortress),audit=value))
    write(out/'SOURCE-PLAYBACK.json',dict(passed=True,scope=data['playback_audit_scope'],
        recordings=audits,unselected_source_anomalies=data['unselected_source_anomalies']))
    print('PASS 成片選用完整演出逐格、結果、來源媒體核對；未選用場次另列',flush=True)


def audit(out):
    data=read(out/'PLAN.json');sound=read(out/'SOUNDTRACK.json')
    source=decode(data['music'])[:sound['music_samples']]*sound['gain']
    source[:RATE]*=np.linspace(0,1,RATE,dtype=np.float32)[:,None]
    tail=round(2.5*RATE);source[-tail:]*=np.linspace(1,0,tail,dtype=np.float32)[:,None]
    def pcm(name):
        with wave.open(str(out/name),'rb') as f:return np.frombuffer(f.readframes(f.getnframes()),dtype='<i2').reshape(-1,2)
    expected=np.round(np.clip(source,-1,1)*32767).astype('<i2');bed=pcm('MUSIC.wav')
    assert np.max(np.abs(expected.astype(np.int32)-bed))<=1
    mixed=pcm('MIX.wav').astype(np.float32)/32767;decoded=decode(out/MOVIE)[:len(mixed)]
    assert len(decoded)==len(mixed) and np.max(np.abs(decoded))<1
    error=decoded-mixed
    def snr(left,diff):return float(10*np.log10(max(float(np.mean(left.astype(np.float64)**2)),1e-16)/max(float(np.mean(diff.astype(np.float64)**2)),1e-16)))
    whole=snr(mixed,error);assert whole>25,whole
    cuts=[]
    for scene in data['scenes'][1:]:
        center=round(scene['start_at']*RATE);left=center-round(.2*RATE);right=center+round(.2*RATE)
        score=snr(mixed[left:right],error[left:right]);assert score>20,(scene['id'],score)
        cuts.append(dict(scene=scene['id'],sample=center,decoded_snr_db=score,continuous_source_samples=True))
    freezes=re.findall(r'freeze_start: ([0-9.]+).*?freeze_duration: ([0-9.]+).*?freeze_end: ([0-9.]+)',(out/'VIDEO-AUDIT.log').read_text(),re.S)
    reviews=[]
    for start,duration,end in freezes:
        lo,hi=float(start),float(end);checks=[]
        for s in data['scenes']:
            left=max(lo,s['start_at']);right=min(hi,s['start_at']+s['duration'])
            if right-left<.2:continue
            if s['kind']=='card':checks.append(dict(scene=s['id'],classification='刻意靜態標題卡'));continue
            for fraction in [.1,.5,.9]:
                at=left+(right-left)*fraction
                crop='crop=1440:810:240:105' if s['kind']=='anime' else 'crop=1280:960:320:60'
                actual=frame(out/MOVIE,at,crop+',scale=320:240:flags=area')
                original=frame(s['file'],s['start']+at-s['start_at'],'scale=320:240:flags=area')
                difference=sum(ImageStat.Stat(ImageChops.difference(actual,original)).mean)/3
                assert difference<12,('凍結區間與來源不同',s['id'],at,difference)
                checks.append(dict(scene=s['id'],seconds=at,source_difference=difference,classification='原速對應來源等待／低變動'))
        assert checks
        reviews.append(dict(start=lo,end=hi,checks=checks))
    write(out/'FREEZE-REVIEW.json',dict(passed=True,events=reviews))
    (out/'RIGHTS.txt').write_text(data['rights']+'\n全部遊戲畫面為本輪重新拍攝；1988動畫及海報沿已核來源。\n')
    settings=next(s for s in data['scenes'] if s['title'].startswith('系統設定收進框內'))
    fortress=next(s for s in data['scenes'] if s.get('fortress_animation'))
    for name,scene in [('settings',settings),('fortress',fortress),('ending',data['scenes'][-1])]:
        run('ffmpeg','-nostdin','-v','error','-threads','2','-ss',scene['start_at']+scene['duration']/2,
            '-i',out/MOVIE,'-frames:v','1',out/f'REVIEW-{name}.png')
    write(out/'CONTINUITY-AUDIT.json',dict(passed=True,version=VERSION,music_sha256=sha(data['music']),
        music_track_count=1,music_loops=0,original_clip_audio_mixed=False,decoded_mix_snr_db=whole,
        cuts=cuts,peak=float(np.max(np.abs(decoded))),sfx_events=len(sound['sfx_events']),
        all_gameplay_new=True,freeze_review=reviews,visual_review='待目視CONTACT',
        limits=['macOS與Windows真機、人耳驗收未新增','音效時刻依frame trace首次played取樣，不稱逐取樣同步或DOS整場同態']))
    print('PASS 單曲全片連續、',len(cuts),'個剪接點、AAC SNR',whole,'dB',flush=True)


def publish(out):
    data=read(out/'CONTINUITY-AUDIT.json');assert data['passed'] and args.visual_reviewed
    destination=ROOT/'dist-all'/VERSION/'promo'
    assert not list(destination.iterdir()),'拒絕覆寫推廣片'
    data['visual_review']='本輪CONTACT及完整原尺寸抽樣已目視，字幕、來源與片尾完整'
    write(out/'CONTINUITY-AUDIT.json',data)
    names=[MOVIE,'PLAN.json','AUDIT.json','TEXT-BOUNDS.json','SOURCE-FRAMES.json','SOURCE-PLAYBACK.json','CONTINUITY-AUDIT.json',
        'SOUNDTRACK.json','REUSE.json','FREEZE-REVIEW.json','AUDIO-AUDIT.log','VIDEO-AUDIT.log','CONTACT.jpg','RIGHTS.txt']
    names+=['REVIEW-settings.png','REVIEW-fortress.png','REVIEW-ending.png']
    files=[]
    for name in names:
        target=destination/(f'LOGH3-{VERSION}-promo.mp4' if name==MOVIE else name)
        shutil.copyfile(out/name,target)
        files.append(dict(path=str(target.relative_to(ROOT)),bytes=target.stat().st_size,sha256=sha(target)))
    write(destination/'SHA256SUMS.json',dict(version=VERSION,files=files))
    index=ROOT/'docs/internals/data/v140-release-promo.json'
    assert not index.exists()
    write(index,dict(version=VERSION,source_commit=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
        working_tree=subprocess.check_output(['git','status','--porcelain'],text=True),
        frozen_product_inputs_sha256=sha(ROOT/'workplace'/('package92-'+VERSION+'-release139')/'INPUTS.json'),
        generator_sha256={name:sha(ROOT/'tools/scripts'/name) for name in ['promo_v140_prepare.py','promo_v140_movie.py','v140_platform_delivery.py','ai_gameplay_record.py','fortress_presentation_gui.py','hero_story_promo.py']},
        tests=dict(command='tools/gui.sh test ./internal/... ./cmd/...',packages=19,passed=15,no_test_files=4),
        platform_receipts={str(p.relative_to(ROOT)):sha(p) for p in [
            ROOT/'dist-all'/VERSION/'smoke/full-linux/SMOKE.json',
            ROOT/'dist-all'/VERSION/'smoke/release-linux/SMOKE.json',
            ROOT/'dist-all'/VERSION/'smoke/macos/MACOS.json',
            ROOT/'workplace/music69-windows/windows139-compact/RESULT.json',
            ROOT/'workplace/music69-windows/compact139-gui-f1/RESULT.json',
            ROOT/'workplace/promo139/compact-delivery-f0/RESULT.json']},
        packages=read(ROOT/'dist-all'/VERSION/'SHA256SUMS-platforms.json'),promo_files=files,
        recording_sources=read(out/'PLAN.json')['fresh_gameplay_inputs'],audit=data,
        source_playback=read(out/'SOURCE-PLAYBACK.json'),
        platforms=dict(linux='修正版實際ZIP、15條入口與正常帝國GUI',windows='修正版實際ZIP、8條Wine入口與正常同盟GUI',macos='雙架構／依賴／ARM簽章與ZIP，未真機執行'),
        rights='完整版與推廣片僅本機，Release只有三平台引擎、來源與校驗檔'))
    print('DELIVERED',destination,flush=True)


if __name__=='__main__':
    assert os.getuid()==1000
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('mode',choices=['plan','render','source_audit','audit','publish'])
    p.add_argument('--out',default='/work/promo139/movie-compact')
    p.add_argument('--reuse-dir',help='只重用已核對同來源時碼及相同字幕的無聲畫面片段；音軌仍整片重建')
    p.add_argument('--visual-reviewed',action='store_true')
    args=p.parse_args()
    if args.mode=='plan':plan()
    else:globals()[args.mode](Path(args.out))
