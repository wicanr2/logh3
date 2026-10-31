"""171：保留國徽故事與完整交戰，補錄正式 AppImage 的 P 索敵及行軍。"""
import argparse
import copy
import math
from pathlib import Path

import poster_frame_promo as frame
import poster_frame_mux as mux
import scout_full_release as package

ROOT=frame.ROOT; VERSION=package.VERSION; BATCH=package.BATCH; WORK=BATCH/'movie'
OUT=package.OUT/'promo'; NAME='LOGH3-'+VERSION+'-promo.mp4'
PARENT=ROOT/'dist-all/v.1.6.3-20261026/promo'
read,write,sha=frame.read,frame.write,frame.sha

def configure():
    frame.VERSION,frame.BATCH,frame.WORK,frame.OUT,frame.NAME=VERSION,BATCH,WORK,OUT,NAME
    frame.movie.VERSION,frame.movie.BATCH,frame.movie.WORK=VERSION,BATCH,WORK
    frame.movie.OUT,frame.movie.NAME=OUT,NAME
    frame.movie.SFX=package.STAGE/'common/music'
    # The imported grid and render helpers must use these same release paths.
    frame.configure=lambda:None

def prepare():
    assert BATCH.stat().st_uid==1000 and not WORK.exists()
    assert not OUT.exists() or not list(OUT.iterdir())
    parent=read(PARENT/'PLAN.json'); audit=read(PARENT/'AUDIT.json')
    assert audit['passed'] and audit['visual_review'].startswith('已目視')
    assert sha(PARENT/audit['file'])==audit['sha256']
    inputs={}
    def bind(path):
        path=Path(path); inputs[str(path.relative_to(ROOT))]=sha(path)
    for path,h in parent['input_sha256'].items():
        assert sha(ROOT/path)==h,path
        inputs[path]=h
    binary=package.OUT/'full'/('LOGH3-'+VERSION+'-linux-x86_64.AppImage')
    for p in [PARENT/'PLAN.json',PARENT/'AUDIT.json',PARENT/audit['file'],binary,
              Path(__file__).resolve(),ROOT/'tools/scripts/release_battle_gui.py',
              ROOT/'tools/scripts/release_garrison_gui.py']:
        bind(p)
    def marks(folder):
        r=read(folder/'RESULT.json')
        assert r['passed'] and r['normal_entry'] and r['appimage']
        assert not any(r[k] for k in ['state_injection','seed_injection','script_flag'])
        assert r['binary_sha256']==sha(binary)
        assert sha(folder/'RAW.mp4')==r['video_sha256']
        for p in [folder/'RESULT.json',folder/'RAW.mp4',folder/'FRAMES.jsonl']:bind(p)
        return r,{e['name']:(e.get('wall_ns',e.get('wall_begin_ns',0))-r['record_started_utc_ns'])/1e9
                  for e in r['events'] if e.get('name')}
    battle,bm=marks(BATCH/'battle-linux');garrison,gm=marks(BATCH/'garrison-linux')
    scenes=copy.deepcopy(parent['scenes']);fresh_ids=[]
    def fresh(index,start,stop):
        s=scenes[index];s.update(source=str(BATCH/'garrison-linux/RAW.mp4'),start=max(0,start),
            duration=math.ceil((stop-start)*60-1e-6)/60,layout='wide',
            provenance=VERSION+' 正式 AppImage 正常入口新錄，原速操作')
        s.pop('clip_sha256',None);fresh_ids.append(index)
    begin,stock,ready,end=gm['first-commander']-1.6,gm['local-ship-stock']-1,gm['fleet-ready']-1,gm['departure-order']+4
    for index,start,stop in [(14,begin,stock),(15,stock,ready),(16,ready,end)]:fresh(index,start,stop)
    fresh(17,gm['departed']-.2,gm['departed']+1.2)
    order=next(e for e in battle['events'] if e.get('kind')=='march_order')
    trace=[read_line for read_line in map(__import__('json').loads,(BATCH/'battle-linux/FRAMES.jsonl').read_text().splitlines())]
    movement=[];previous=None
    for r in trace:
        if r['wall_ns']<order['wall_ns'] or not r.get('tactical'):continue
        u=next((u for u in r['tactical']['units'] if u['Fleet']==order['fleet'] and u['Squadron']==0 and not u['Out']),None)
        if not u:continue
        xy=[u['X'],u['Y']]
        if previous is not None and xy!=previous:
            movement.append(dict(seconds=(r['wall_ns']-battle['record_started_utc_ns'])/1e9,xy=xy))
        previous=xy
        if xy==order['to_xy']:break
    assert movement and movement[-1]['xy']==order['to_xy'],movement
    def added(start,stop,title,caption):
        return dict(kind='game',duration=math.ceil((stop-start)*60-1e-6)/60,title=title,caption=caption,
            source=str(BATCH/'battle-linux/RAW.mp4'),start=max(0,start),layout='wide',
            provenance=VERSION+' 正式 AppImage 正常入口新錄，原速操作')
    scout=added(bm['entry']+.8,bm['scout-restored']+.6,'讓指揮 P 探明前路',
        'P 索敵｜顯示範圍，切換視野選定目標，再回到原視窗。')
    march=added(movement[0]['seconds']-.6,movement[-1]['seconds']+.7,'保持隊形，向前推進',
        '整隊行軍｜旗艦帶隊前進，各中隊維持隊形位置。')
    scenes[18:18]=[scout,march]
    scenes[-1]['caption']='銀河英雄傳說 III SP｜v1.6.4'
    scenes[-1]['subtitle']='P 索敵・整隊行軍・前線整軍・三組海報｜Windows / Linux / macOS'
    retained=[]
    for index,s in enumerate(scenes):
        old_id=s.get('id');s['id']=index
        fresh_old=old_id in fresh_ids and s not in (scout,march)
        if s['kind']!='story' and s not in (scout,march) and not fresh_old:
            p=ROOT/'workplace/poster166/movie'/f'scene-{old_id:02}.mp4'
            assert sha(p)==parent['scenes'][old_id]['clip_sha256'];bind(p)
            retained.append(dict(scene=index,old_scene=old_id,path=str(p.relative_to(ROOT)),sha256=sha(p)))
    limit=float(frame.probe(frame.movie.MUSIC)['format']['duration'])-.25
    # Preserve every old battle playback and the entire new commander/stock/sortie path.
    # Shorten only optional card and feature holds to fit the existing one-shot score.
    for index,minimum in [(0,2),(3,2),(36,2),(4,2),(5,1.5),(6,1.5),
                          (7,1),(8,1),(9,1),(10,1),(11,2),(12,2),(13,2),(17,1),(1,2.5),(2,2.5)]:
        excess=sum(s['duration'] for s in scenes)-limit
        if excess<=0:break
        scenes[index]['duration']=max(minimum,scenes[index]['duration']-math.ceil(excess*60)/60)
    # Any shortened retained hold gets a new trim from its verified original clip.
    for r in retained[:]:
        s=scenes[r['scene']]
        if s['duration']!=parent['scenes'][r['old_scene']]['duration']:
            s.update(source=str(ROOT/r['path']),start=0,layout='wide',
                     provenance='v1.6.3 已驗片段裁短停留；未加速，完整畫布與既有字幕保留')
            s['precomposed']=True;retained.remove(r)
    elapsed=0
    for s in scenes:s['start_at']=elapsed;elapsed+=s['duration']
    assert elapsed<=limit,('完整原速操作超出單首配樂',elapsed,limit)
    WORK.mkdir();OUT.mkdir(exist_ok=True)
    plan=dict(parent,version=VERSION,seconds=elapsed,scenes=scenes,input_sha256=inputs,text_bounds=[],
        inherited_text_bounds=parent['text_bounds'],retained_clips=retained,
        parent_plan=dict(path=str((PARENT/'PLAN.json').relative_to(ROOT)),sha256=sha(PARENT/'PLAN.json')),
        new_scene_ids=[14,15,16,17,18,19],march_movement=movement,
        storyboard=parent['storyboard']+'；前線整軍後加入 P 範圍選點／返回視窗與實際隊形行軍，再銜接完整武器演出。',
        scope='國徽、故事、動畫與完整交戰沿既有已驗版本；前線換提督／領船／出擊及 P 索敵、行軍由 v1.6.4 正式 AppImage 正常入口新錄。全部原速，配樂一首不循環。')
    write(WORK/'PLAN.json',plan)
    write(WORK/'PARENT.json',dict(movie=str((PARENT/audit['file']).relative_to(ROOT)),sha256=audit['sha256'],retained=retained))
    print('READY',len(scenes),'scenes',elapsed,'seconds; scout',scout['duration'],'march',march['duration'],flush=True)

def render():
    # Precomposed holds retain their existing full canvas, title, and captions.
    plan=read(WORK/'PLAN.json')
    for s in plan['scenes']:
        if not s.get('precomposed'):continue
        target=WORK/f'pretrim-{s["id"]:02}.mp4'
        frame.run('ffmpeg','-nostdin','-v','error','-i',s['source'],'-an','-frames:v',round(s['duration']*60),
            '-c:v','libx264','-preset','veryfast','-crf','19','-threads','2','-pix_fmt','yuv420p',
            '-video_track_timescale','60000',target)
        plan['retained_clips'].append(dict(scene=s['id'],path=str(target.relative_to(ROOT)),sha256=sha(target)))
        plan['input_sha256'][str(target.relative_to(ROOT))]=sha(target)
    write(WORK/'PLAN.json',plan)
    frame.render()

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('mode',choices=['prepare','render','grid','audit','publish']);a=p.parse_args()
    configure()
    if a.mode=='grid':mux.main()
    elif a.mode in ['audit','publish']:getattr(frame,a.mode)()
    else:globals()[a.mode]()
