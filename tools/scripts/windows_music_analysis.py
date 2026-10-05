"""保存全部固定曲段及缺口；這份觀測不取代first-feature播放驗收。"""
import argparse
import array
import hashlib
import json
import math
from pathlib import Path
import sys
import wave

assert Path.cwd() == Path('/src') and sys.byteorder == 'little'
parser = argparse.ArgumentParser()
parser.add_argument('batch')
args = parser.parse_args()
assert Path(args.batch).name == args.batch
root = Path('workplace/music69-windows')
out = root / args.batch
target = out / 'ANALYSIS.json'
assert not target.exists()
ready = json.loads((out / 'READY.json').read_text())
check = json.loads((out / 'client/CHECK.json').read_text())
assert check['exit'] == 0 and all(row['exit'] == 0 for row in check['frames'])
assert all(row['different'] == '0' for row in check['frames'])
markers = {m['name']: m['monotonic']-ready['capture_start_monotonic'] for m in check['markers']}
server = json.loads((out / 'SERVER.json').read_text())
assert server['client']['exit'] == 0
active = [c for snapshot in server['samples'] for c in snapshot['sink_inputs']]
pid = str(json.loads((out / 'client/PROCESS.json').read_text())['pid'])
assert active and all(c['properties'].get('application.process.id') == pid for c in active)
assert all(c['sample_specification'] == 'float32le 2ch 44100Hz' and not c['mute'] for c in active)
assert all(all(v['value_percent'] == '100%' for v in c['volume'].values()) for c in active)
raw = (out / 'monitor.f32le').read_bytes()
assert raw and len(raw) % 8 == 0
values = array.array('f'); values.frombytes(raw)
assert all(math.isfinite(v) for v in values)
peak = max(map(abs, values))
assert 0 < peak < 1
stats = {'bytes': len(raw), 'sha256': hashlib.sha256(raw).hexdigest(),
         'seconds': len(raw)/352800, 'peak': peak,
         'rms': math.sqrt(sum(v*v for v in values)/len(values)),
         'nonzero_samples': sum(v != 0 for v in values)}
decode = json.loads((root / 'playeraudio69-r105-decode/DECODE.json').read_text())
records = []
for name, phase in [('OPEN_M', ('intro','main')), ('BGM_0F', ('fm0','fm1')), ('BGM_1F', ('fm1','off'))]:
    source = (root / 'playeraudio69-r105-decode' / (name+'.f32le')).read_bytes()
    row = next(r for r in decode if r['track'] == name)
    assert hashlib.sha256(source).hexdigest() == row['pcm_sha256']
    assert hashlib.sha256(Path(row['ogg']).read_bytes()).hexdigest() == row['ogg_sha256']
    sections = []
    begin, end = markers[phase[0]], markers[phase[1]]
    for second in range(1, 9):
        offset = second*352800
        needle = source[offset:offset+16384]
        feature = array.array('f'); feature.frombytes(needle)
        rms = math.sqrt(sum(v*v for v in feature)/len(feature))
        found = raw.find(needle) if rms > .001 else -1
        section = {'source_seconds':second, 'feature_rms':rms, 'feature_found':found >= 0}
        if found >= 0:
            assert found % 8 == 0
            sb, mb = offset, found
            se, me = offset+len(needle), found+len(needle)
            while sb >= 8 and mb >= 8 and source[sb-8:sb] == raw[mb-8:mb]:
                sb -= 8; mb -= 8
            while se+8 <= len(source) and me+8 <= len(raw) and source[se:se+8] == raw[me:me+8]:
                se += 8; me += 8
            overlap = max(0, min(me/352800, end)-max(mb/352800, begin))
            section.update(source_frames=[sb//8, se//8], monitor_frames=[mb//8, me//8],
                           continuous_exact_seconds=(me-mb)/352800,
                           phase_overlap_seconds=overlap)
        sections.append(section)
    usable = [s for s in sections if s['feature_rms'] > .001]
    assert usable
    first = usable[0]
    phase_samples = [s for s in sections if s.get('phase_overlap_seconds',0) >= 3]
    # 保留原檢查的失敗；後面的符合片段只屬額外觀測。
    first_pass = first.get('continuous_exact_seconds',0) >= 3 and first.get('phase_overlap_seconds',0) >= 3
    spans = sorted({tuple(s['source_frames']+s['monitor_frames']) for s in sections if s['feature_found']})
    gaps = []
    for previous, current in zip(spans, spans[1:]):
        if current[0] < previous[1]:
            continue
        gap = raw[previous[3]*8:current[2]*8]
        gap_values = array.array('f'); gap_values.frombytes(gap)
        gaps.append({'source_frames':[previous[1],current[0]],
                     'monitor_frames':[previous[3],current[2]],
                     'monitor_seconds':len(gap)/352800,
                     'nonzero_samples':sum(v != 0 for v in gap_values),
                     'all_zero': bool(gap_values) and all(v == 0 for v in gap_values)})
    clip_begin = round((begin+1)*44100)*2
    clip_end = min(round((begin+6)*44100)*2, round(end*44100)*2)
    samples = array.array('h', (round(v*32768) for v in values[clip_begin:clip_end]))
    clip = out / (name+'-observed.wav')
    assert not clip.exists()
    with wave.open(str(clip),'wb') as stream:
        stream.setnchannels(2); stream.setsampwidth(2); stream.setframerate(44100)
        stream.writeframes(samples.tobytes())
    records.append({'track':name, 'ogg_sha256':row['ogg_sha256'],
                    'decoded_source_sha256':row['pcm_sha256'], 'sections':sections,
                    'first_feature_gate_passed':first_pass,
                    'additional_three_second_phase_sample':bool(phase_samples),
                    'gaps_between_identified_spans':gaps,
                    'clip':{'path':str(clip),'sha256':hashlib.sha256(clip.read_bytes()).hexdigest(),
                            'seconds':len(samples)/88200, 'contains_observed_output_without_gap_removal':True}})
silent = []
for phase, end_phase in [('main','fm0'), ('off',None)]:
    begin = markers[phase]+1
    end = markers[end_phase]-.5 if end_phase else begin+3
    assert 2 <= end-begin and end < stats['seconds']
    segment = values[round(begin*44100)*2:round(end*44100)*2]
    count = sum(v != 0 for v in segment)
    silent.append({'phase':phase, 'monitor_seconds':[begin,end], 'samples':len(segment), 'nonzero':count})
    assert count == 0, (phase,count)
result = {'schema':'logh3.windows-music-observation.v1','monitor':stats,'tracks':records,
          'silence':silent,'normal_gui_passed':True,'endpoint_observations':len(active),
          'first_feature_gate_passed':all(r['first_feature_gate_passed'] for r in records),
          'additional_phase_samples_verified':all(r['additional_three_second_phase_sample'] for r in records),
          'full_playback_accepted':False,'physical_sound_card':False,'human_listening':False,'loop_verified':False}
with target.open('x') as stream:
    json.dump(result,stream,ensure_ascii=False,indent=2);stream.write('\n')
print('OBSERVED',args.batch,'first-feature gate',result['first_feature_gate_passed'],
      'additional phase samples',result['additional_phase_samples_verified'],'silence',silent)
