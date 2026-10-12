"""規格139：核對正常視窗實際顯示的COMANI段、逐格順序及結果時點。"""
import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
import statistics
import re
import subprocess

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('recording', type=Path)
a = p.parse_args()
root = a.recording
trace = json.loads((root/'TRACE.json').read_text())
assert trace['normal_x11_input']
closed = json.loads((root/'FRAME-TRACE-CLOSE.json').read_text())
path = root/'FRAMES.jsonl'
assert hashlib.sha256(path.read_bytes()).hexdigest() == closed['frame_trace_sha256']
scenes, modes, gaps = {}, Counter(), []
samples = []
for line in path.open():
    row = json.loads(line)
    v = row.get('playback')
    if not v:
        continue
    modes[row['graphics']] += 1
    s = scenes.setdefault(v['id'], {'kind':v['kind'], 'first_tick':v['tick'],
        'result_seen':False, 'result_first_tick':None,'applied':0, 'frames':[], 'cadence_ms':[], 'modes':set(), 'last':None})
    s['modes'].add(row['graphics'])
    last = s['last']
    if last:
        assert v['tick'] >= last['tick'], ('時間倒退', v['id'])
        assert v['applied'] >= last['applied'], ('畫格倒退', v['id'])
    assert v['applied'] == sum(q['tick'] <= v['tick'] for q in v['pictures'])
    pictures = v['pictures']
    for i, q in enumerate(pictures):
        seg = bytes.fromhex(q['segment_hex'])
        assert len(seg) == 4 and q['side'] in (0,1)
        if i and q['segment_hex'] == pictures[i-1]['segment_hex'] and q['side'] == pictures[i-1]['side']:
            # The first attack/damage frame may repeat a previous D0 base segment.
            assert q['frame'] == 0 or q['frame'] == pictures[i-1]['frame']+1
        else:
            assert q['frame'] == 0
        if i:
            assert q['tick']-pictures[i-1]['tick'] >= v['interval']
    if v['applied'] != s['applied']:
        if v['interval'] > 0 and v['applied']-s['applied'] != 1:
            gaps.append({'id':v['id'],'from':s['applied'],'to':v['applied'],'tick':v['tick']})
        if s['frames']:
            s['cadence_ms'].append((row['wall_ns']-s['frames'][-1]['wall_ns'])/1e6)
        s['frames'].append({'applied':v['applied'],'tick':v['tick'],'wall_ns':row['wall_ns']})
        samples.append({'id':v['id'],'weapon':v['weapon'],'applied':v['applied'],
                        'wall_ns':row['wall_ns'],'graphics':row['graphics']})
        s['applied'] = v['applied']
    if v['kind'] in ('fleet','planet'):
        assert len(pictures) >= 3 and pictures[0]['frame'] == pictures[1]['frame'] == 0
        assert bytes.fromhex(pictures[0]['segment_hex'])[2:] == b'D0'
        assert bytes.fromhex(pictures[1]['segment_hex'])[2:] == b'D0'
        if v['kind']=='fleet':assert pictures[0]['side'] != pictures[1]['side']
        attack = [q for q in pictures[2:] if bytes.fromhex(q['segment_hex'])[2] == ord('A')]
        damage = [q for q in pictures[2:] if bytes.fromhex(q['segment_hex'])[2] == ord('D')]
        assert attack and [q['frame'] for q in attack] == list(range(len(attack)))
        assert all(q['side'] == pictures[0]['side'] for q in attack)
        assert all(q['side'] == pictures[1]['side'] for q in damage)
        assert not damage or max(q['tick'] for q in attack)<min(q['tick'] for q in damage)
        assert not v['result'] or v['result_tick']==pictures[-1]['tick']
    if v['kind'] == 'fleet':
        assert all(bytes.fromhex(q['segment_hex'])[3] == b'BMF'[v['weapon']] for q in attack)
        assert all(q['tick'] <= v['attack_end'] for q in attack)
        assert all(q['tick'] > v['attack_end'] for q in damage)
        if v['damage_applied']:
            assert v['tick'] >= v['attack_end']
            assert bool(damage) == (v['lost'] > 0)
            assert v['result_tick'] == pictures[-1]['tick']
        if v['result']:
            assert v['damage_applied'] and v['tick'] >= v['result_tick']
        for cue in v.get('sounds', []):
            following = damage if cue['defence'] else attack
            if any(q['tick'] <= v['tick'] for q in following):
                assert cue['played'], ('首格早於音效閘門', v['id'], cue)
    if v['result']:
        assert v['applied'] == len(pictures), ('結果早於最後一格', v['id'])
        s['result_seen'] = True
        if s['result_first_tick'] is None:s['result_first_tick']=v['tick']
    s['last'] = v
assert scenes, '沒有正常交戰'
assert all(s['result_seen'] for s in scenes.values()), '存在未播完場次'
assert not gaps, ('正常速度有未顯示格', gaps)
summary = []
for identity, s in scenes.items():
    v = s['last']
    summary.append({'id':identity,'kind':s['kind'],'graphics':sorted(s['modes']),
        'weapon':v['weapon'], 'lost':v['lost'], 'pictures':len(v['pictures']),
        'observed_transitions':len(s['frames']), 'result_seen':s['result_seen'],
        'observed_result_pause_ticks':v['tick']-s['result_first_tick'],
        'median_frame_ms':statistics.median(s['cadence_ms']) if s['cadence_ms'] else None,
        'segments':list(dict.fromkeys(q['segment_hex'] for q in v['pictures']))})
    summary[-1]['same_side_planet'] = v['kind']=='planet' and v['pictures'][0]['side']==v['pictures'][1]['side']
result = {'passed':True,'normal_entry':True,'state_injection':False,
          'ai':trace['command'][trace['command'].index('-ai')+1] if 'command' in trace else None,
          'frame_trace_sha256':closed['frame_trace_sha256'],
          'graphics_observations':dict(modes),'skipped_frames':gaps,'scenes':summary}
result['scope'] = '艦隊／行星既有演出段序、逐格及結果時點；完整登陸及占領呈現仍沿原有未完成項，不宣稱由本輪閉合'
raw = root/'RAW.mp4'
probe = json.loads(subprocess.check_output(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(raw)]))
video = next(q for q in probe['streams'] if q['codec_type']=='video')
audio = next(q for q in probe['streams'] if q['codec_type']=='audio')
assert video['width']*3==video['height']*4 and float(probe['format']['duration'])>10
sound = subprocess.run(['ffmpeg','-nostdin','-v','info','-i',str(raw),'-vn','-af','volumedetect','-f','null','-'],capture_output=True,text=True,timeout=90,check=True).stderr
(root/'AUDIO-CHECK.log').write_text(sound)
mean = float(re.search(r'mean_volume: (-?[0-9.]+) dB', sound)[1])
assert mean>-55, ('實錄近乎靜音', mean)
result['media'] = {'raw_sha256':hashlib.sha256(raw.read_bytes()).hexdigest(),
                   'duration':float(probe['format']['duration']), 'width':video['width'],
                   'height':video['height'],'fps':video['r_frame_rate'],
                   'audio_channels':audio['channels'],'mean_volume_db':mean}
visual = subprocess.run(['ffmpeg','-nostdin','-v','info','-i',str(raw),'-an',
    '-vf','scale=320:240,blackdetect=d=0.25:pic_th=0.98:pix_th=0.1,freezedetect=n=-50dB:d=1',
    '-f','null','-'],capture_output=True,text=True,timeout=120,check=True).stderr
(root/'VIDEO-CHECK.log').write_text(visual)
result['media']['black_intervals'] = re.findall(r'black_start:([0-9.]+) black_end:([0-9.]+) black_duration:([0-9.]+)', visual)
result['media']['freeze_events'] = re.findall(r'lavfi.freezedetect.([a-z_]+): ([0-9.]+)', visual)
result['media']['freeze_scope'] = '完整GUI包含刻意停留的選單、訊息及音效等待；逐格閘門另由實際Draw與來源時間線核對'
# Extract real recorded frames at observed source changes for visual review.
start = trace.get('record_started_utc_ns', trace.get('input_started_utc_ns',0)-500_000_000)
result['sample_timing'] = '錄影程序開始時點' if 'record_started_utc_ns' in trace else '舊收據按輸入前500ms估計，只供目視導覽'
chosen = []
for weapon in sorted({q['weapon'] for q in samples}):
    group = [q for q in samples if q['weapon']==weapon]
    for q in [group[len(group)//4],group[len(group)//2],group[-2]]:
        at = max(0,(q['wall_ns']-start)/1e9)
        destination = root/f'review-{len(chosen):02d}.png'
        subprocess.run(['ffmpeg','-nostdin','-v','error','-y','-ss',str(at),'-i',str(raw),'-frames:v','1',str(destination)],check=True,timeout=30)
        chosen.append(dict(q,video_seconds=at,file=destination.name))
result['visual_samples'] = chosen
(root/'PLAYBACK-AUDIT.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
print(json.dumps(result, ensure_ascii=False, indent=2))
