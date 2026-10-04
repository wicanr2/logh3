"""將正式ALSA monitor與正式解碼signed16全樣本對齊；不縮放、不容差。"""
import argparse
import array
import hashlib
import json
from pathlib import Path
import sys

p=argparse.ArgumentParser()
p.add_argument('root')
p.add_argument('--sources',required=True)
p.add_argument('--gui',action='store_true')
a=p.parse_args()
root=Path(a.root)
target=root/'pcm-comparison.json'
assert not target.exists()
raw=(root/'monitor.f32le').read_bytes()
assert len(raw)%8==0
events=[json.loads(s) for s in (root/'gui/playback-events.jsonl').read_text().splitlines()] if a.gui else json.loads((root/'probe.json').read_text())['events']
starts=[e for e in events if e['kind']==('effect_start' if a.gui else 'start')]
cursor=0
cases=[]
for e in starts:
    id=e['id']
    source=Path(a.sources)/f'source-{id:02d}.s16le'
    pcm=source.read_bytes()
    values=array.array('h');values.frombytes(pcm)
    if sys.byteorder!='little':values.byteswap()
    floats=array.array('f',(v/32768 for v in values))
    if sys.byteorder!='little':floats.byteswap()
    expected=floats.tobytes()
    # 簽章只由來源決定，選前半段辨識度最高的256格；仍比較全段。
    size=min(2048,len(expected)//4//8*8)
    assert size>0
    offsets=range(0,max(8,len(expected)//2-size+1),size)
    offset=max(offsets,key=lambda o:len(set(values[o//4:(o+size)//4])))
    marker=expected[offset:offset+size]
    assert any(marker)
    match=raw.find(marker,cursor+offset)
    if match<0:
        cases.append({'id':id,'source':str(source),'marker_found':False,'complete_exact':False})
        continue
    begin=match-offset
    assert begin>=cursor and begin%8==0
    got=raw[begin:begin+len(expected)]
    equal=got==expected
    first=None if equal else next((i for i,(x,y) in enumerate(zip(got,expected)) if x!=y),len(got))
    prefix=len(expected) if equal else first//8*8
    remaining=expected[prefix:]
    cases.append({'id':id,'source':str(source),'source_s16le_sha256':hashlib.sha256(pcm).hexdigest(),
        'marker_found':True,'source_marker_offset_bytes':offset,'monitor_begin_frame':begin//8,
        'expected_frames':len(expected)//8,'exact_prefix_frames':prefix//8,'complete_exact':equal,
        'first_different_frame':None if first is None else first//8,
        'remaining_source_frames':len(remaining)//8,
        'nonzero_bytes_in_remaining_source':sum(v!=0 for v in remaining),
        'source_duration_seconds':len(expected)/8/44100})
    cursor=begin+prefix
result={'classification':'正式解碼signed16→精確float32／ALSA monitor全段逐位元組；非原版硬體波形或人耳',
    'monitor_sha256':hashlib.sha256(raw).hexdigest(),'sample_rate':44100,'channels':2,
    'calls':len(cases),'complete_exact':sum(c['complete_exact'] for c in cases),'cases':cases}
target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='cases'},ensure_ascii=False))
for c in cases:
    if not c['complete_exact']:print(json.dumps(c,ensure_ascii=False))
