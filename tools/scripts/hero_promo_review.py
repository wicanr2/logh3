"""規格140：回查全片低變動區間，區分剪輯凍結與來源自身停留。"""
import json
from pathlib import Path
import re
import subprocess

from PIL import Image, ImageChops, ImageStat

out=Path('/work/hero130/promo')
plan=json.loads((out/'PLAN.json').read_text())
origin=json.loads((out/'SOURCE-FRAMES.json').read_text())['video_origin_seconds']
events=re.findall(r'freeze_start: ([0-9.]+).*?freeze_duration: ([0-9.]+).*?freeze_end: ([0-9.]+)',
                  (out/'VIDEO-AUDIT.log').read_text(),re.S)


def frame(path,t,filter):
    raw=subprocess.check_output(['ffmpeg','-nostdin','-v','error','-ss',str(t),'-i',str(path),
        '-frames:v','1','-vf',filter,'-pix_fmt','rgb24','-f','rawvideo','-'])
    return Image.frombytes('RGB',(320,240),raw)


def difference(a,b):
    return sum(ImageStat.Stat(ImageChops.difference(a,b)).mean)/3


reviews=[]
for start,duration,end in events:
    begin=float(start)-origin;finish=float(end)-origin
    mid=(begin+finish)/2
    scene=next(s for s in plan['scenes'] if s['start_at']<=mid<s['start_at']+s['duration'])
    assert scene['kind']=='game'
    first=max(begin,scene['start_at'])+.15
    last=min(finish,scene['start_at']+scene['duration'])-.15
    points=[first+(last-first)*i/4 for i in range(5)]
    checks=[];sources=[]
    for point in points:
        source_time=scene['start']+point-scene['start_at']
        source=frame(scene['file'],source_time,'scale=320:240:flags=area')
        actual=frame(out/'LOGH3-HERO-PROMO.mp4',point+origin,
                     'crop=1280:960:320:60,scale=320:240:flags=area')
        error=difference(actual,source)
        assert error<12,('成片與來源不符',scene['id'],point,error)
        sources.append(source)
        checks.append(dict(movie_seconds=point+origin,source_seconds=source_time,mean_difference=error))
    reviews.append(dict(start=float(start),end=float(end),scene=scene['id'],checks=checks,
        source_change=max(difference(sources[0],s) for s in sources[1:]),
        classification='原實錄低變動區間，包含選單、戰術等待或音效閘門；成片按原速對應，沒有新增凍結',
        playback_id=scene.get('playback_id')))
(out/'FREEZE-REVIEW.json').write_text(json.dumps(dict(passed=True,events=reviews,
    method='固定背景可能觸發低變動偵測；逐區間五點核對來源時碼與實際成片，未將低變動直接當播放失敗'),
    ensure_ascii=False,indent=2)+'\n')
print('PASS',len(reviews),'個低變動區間，逐項回查來源原速畫面')
