"""155：檢查草稿的留白邊界、固定本體及實錄，不作正式版驗收。"""
import hashlib
import json
import math
from pathlib import Path
import subprocess

import numpy as np
from PIL import Image

ROOT=Path('/src');BATCH=ROOT/'workplace/poster155';OUT=BATCH/'preview-r2'
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
info=read(OUT/'RESULT.json');source=read(BATCH/'SOURCE.json')
assert info['passed'] and info['prototype'] and info['normal_entry']
assert sha(ROOT/'cmd/gin3/poster_decoration.go')==source['production_sha256']
assert sha(ROOT/source['draw_source'])==source['draw_sha256']
events={x['name']:x for x in info['events']}
for item in events.values():assert sha(OUT/item['screenshot'])==item['sha256']
before=np.asarray(Image.open(OUT/'elegant.png').convert('RGB'))
after=np.asarray(Image.open(OUT/'elegant-after.png').convert('RGB'))
plain=np.asarray(Image.open(ROOT/'workplace/poster154/player-linux/map-off.png').convert('RGB'))
center=before[:,160:1120];delta=np.any(center!=plain,axis=2);allowed=np.zeros(delta.shape,dtype=bool)
for x,y,x2,y2 in events['elegant']['observed']['poster_decoration']['bands']:
    allowed[math.floor(y*1.5):math.ceil(y2*1.5),math.floor(x*1.5):math.ceil(x2*1.5)]=True
assert not np.any(delta&~allowed)
poster=np.asarray(Image.open(ROOT/'docs/images/readme-v162-poster.png').convert('RGB'))
for left,right in [(0,160),(1120,1280)]:assert np.array_equal(before[:,left:right],poster[:,left:right])
weapons=np.zeros(before.shape[:2],dtype=bool)
weapons[24:42,480:800]=True;weapons[662:685,610:670]=True
assert not np.any(np.any(before!=after,axis=2)&~weapons)
file=OUT/'elegant.mp4';assert sha(file)==info['video_sha256']
ff=json.loads(subprocess.check_output(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(file)],text=True))
assert len(ff['streams'])==1 and float(ff['format']['duration'])==18
p=subprocess.Popen(['ffmpeg','-v','error','-threads','1','-i',str(file),'-f','rawvideo','-pix_fmt','rgb24','-'],stdout=subprocess.PIPE)
reference=None;frames=0;motion=0;motion_max=0
try:
    while True:
        data=p.stdout.read(1280*720*3)
        if not data:break
        assert len(data)==1280*720*3
        frame=np.frombuffer(data,dtype=np.uint8).reshape(720,1280,3)
        if reference is None:reference=frame.copy()
        change=np.abs(frame[weapons].astype(np.int16)-reference[weapons].astype(np.int16))
        motion_max=max(motion_max,int(change.max()))
        if np.any(change>20):motion+=1
        frames+=1
    assert p.wait(timeout=20)==0
finally:
    if p.poll() is None:p.kill();p.wait()
assert frames==270 and motion>0
result=dict(prototype=True,passed=True,production_unchanged=True,nonblank_pixels=int(np.count_nonzero(~allowed)),
    outside_different=0,poster_pixels=230400,poster_different=0,fixed_composition_endpoint_different=0,
    seconds=18,frames=frames,weapon_motion_frames=motion,weapon_max_change=motion_max,
    source_sha256=sha(BATCH/'SOURCE.json'),player_sha256=sha(OUT/'RESULT.json'),
    video_sha256=sha(file),image_sha256=sha(OUT/'elegant.png'))
(BATCH/'CHECK.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(result,ensure_ascii=False))
