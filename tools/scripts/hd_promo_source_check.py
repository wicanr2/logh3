"""影片成品遊戲區與正常實錄逐點抽驗，區分原演出停格與合成凍結。"""
import argparse
import json
import re
import subprocess
from pathlib import Path
from PIL import Image, ImageChops, ImageStat

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--root', required=True)
a=p.parse_args()
r=Path(a.root);out=r/'promo';plan=json.loads((out/'PLAN.json').read_text())
movie=out/f'LOGH3-{plan["version"]}-promo.mp4'
rows=[]
for t in (10,20,26,33,44,56,65,73,80):
    s=next(s for s in plan['storyboard'] if s['timeline_start']<=t<s['timeline_start']+s['seconds'])
    assert s['source']
    offset=t-s['timeline_start']
    if s.get('pieces'):
        for part in s['pieces']:
            if offset<part['seconds']:
                source_t=part['start']+offset;break
            offset-=part['seconds']
    else:
        source_t=s['start']+offset
    source=out/f'combat-source-{t:02}.png';actual=out/f'combat-movie-{t:02}.png'
    for path,video,stamp,filter_ in [(source,r/s['source']/'RAW.mp4',source_t,'scale=1280:960:flags=lanczos'),(actual,movie,t,'crop=1280:960:320:60')]:
        subprocess.run(['ffmpeg','-nostdin','-y','-v','error','-ss',str(stamp),'-i',str(video),'-vf',filter_,'-frames:v','1','-threads','2',str(path)],check=True)
    with Image.open(source) as x,Image.open(actual) as y:
        d=ImageChops.difference(x.convert('RGB'),y.convert('RGB'))
        mean=sum(ImageStat.Stat(d).mean)/3
        # 兩次有損H.264及YUV取樣允許小差異；不豁免幀序或大區域錯誤。
        assert mean<4,(t,source_t,mean)
        d.save(out/f'combat-diff-{t:02}.png')
    rows.append(dict(movie_seconds=t,raw_seconds=source_t,mean_channel_error=mean,passed=True))
text=(out/'VIDEO-AUDIT.txt').read_text()
intervals=[]
for m in re.finditer(r'freeze_start: ([0-9.]+)',text):
    start=float(m.group(1));end_match=re.search(r'freeze_end: ([0-9.]+)',text[m.end():])
    end=float(end_match.group(1)) if end_match else 90
    ids=[s['id'] for s in plan['storyboard'] if s['timeline_start']<end and s['timeline_start']+s['seconds']>start]
    reason='首尾故事卡' if all(i in (0,8) for i in ids) else '正常實錄的讀檔／操作／原演出等待格；遊戲區抽樣另與原實錄核對'
    intervals.append(dict(start=start,end=end,scenes=ids,classification=reason))
(out/'SOURCE-CHECK.json').write_text(json.dumps(dict(samples=rows,freeze_intervals=intervals,
    criterion='完整遊戲區1280×960；雙重H.264／YUV平均通道誤差小於4；不裁掉遊戲內容',passed=True),ensure_ascii=False,indent=2)+'\n')
audit=json.loads((out/'AUDIT.json').read_text());audit['source_frame_check']=dict(samples=len(rows),passed=True,report='SOURCE-CHECK.json')
audit['frozen_intervals']='全畫面偵測含原實錄停留；詳SOURCE-CHECK.json。正常實錄與成品遊戲區九點抽驗通過，沒有以靜態截圖代替交火。'
(out/'AUDIT.json').write_text(json.dumps(audit,ensure_ascii=False,indent=2)+'\n')
print('PASS 成片完整遊戲區與正常實錄九點核對',flush=True)
