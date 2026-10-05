"""第105輪：將實際Wine monitor綁回正式Vorbis來源，另核正常停止及關閉。"""
import argparse
import array
import hashlib
import json
import math
from pathlib import Path
import sys
import wave

assert Path.cwd() == Path('/src')
parser = argparse.ArgumentParser()
parser.add_argument('batch')
parser.add_argument('--intro-only', action='store_true')
args = parser.parse_args()
assert Path(args.batch).name == args.batch
root = Path('workplace/music69-windows')
out = root / args.batch
target = out / 'PCM-CHECK.json'
assert not target.exists()
raw = (out / 'monitor.f32le').read_bytes()
assert raw and len(raw) % 8 == 0
values = array.array('f'); values.frombytes(raw)
assert sys.byteorder == 'little'
assert all(math.isfinite(v) for v in values)
stats = {'bytes': len(raw), 'sha256': hashlib.sha256(raw).hexdigest(), 'frames': len(raw)//8,
         'sample_rate': 44100, 'channels': 2, 'format': 'float32le',
         'seconds': len(raw)/352800, 'nonzero_samples': sum(v != 0 for v in values),
         'peak': max(map(abs, values)), 'rms': math.sqrt(sum(v*v for v in values)/len(values)),
         'samples_at_or_above_one': sum(abs(v) >= 1 for v in values)}
assert stats['nonzero_samples'] > 0
server = json.loads((out / 'SERVER.json').read_text())
assert server['client']['exit'] == 0
active = [client for snapshot in server['samples'] for client in snapshot['sink_inputs']]
driver_path = out / 'DRIVER.json'
driver = json.loads(driver_path.read_text())['driver'] if driver_path.is_file() else 'pulse'
if driver == 'alsa':
    pid = str(json.loads((out / 'client/PROCESS.json').read_text())['pid'])
    assert active and all(c['properties'].get('application.process.id') == pid for c in active)
else:
    assert active and all(c['properties'].get('application.name') == 'gin3.exe' for c in active)
assert all(c['sample_specification'] == 'float32le 2ch 44100Hz' and not c['mute'] for c in active)
assert all(all(v['value_percent'] == '100%' for v in c['volume'].values()) for c in active)
decode = json.loads((root / 'playeraudio69-r105-decode/DECODE.json').read_text())
records, silent = [], []
markers = None
if not args.intro_only:
    ready = json.loads((out / 'READY.json').read_text())
    check = json.loads((out / 'client/CHECK.json').read_text())
    markers = {m['name']: m['monotonic'] - ready['capture_start_monotonic'] for m in check['markers']}
tracks = ['OPEN_M'] if args.intro_only else ['OPEN_M', 'BGM_0F', 'BGM_1F']
for name in tracks:
    source = (root / 'playeraudio69-r105-decode' / (name + '.f32le')).read_bytes()
    row = next(r for r in decode if r['track'] == name)
    assert hashlib.sha256(source).hexdigest() == row['pcm_sha256']
    assert hashlib.sha256(Path(row['ogg']).read_bytes()).hexdigest() == row['ogg_sha256']
    # 原本一秒的片頭接近靜音，不能作有效的波形識別。
    # 按固定順序選第一個 RMS 超過 .001 的完整特徵，不按結果挑曲段。
    for second in range(1, 15):
        source_offset = second * 44100 * 8
        needle = source[source_offset:source_offset+16384]
        feature = array.array('f'); feature.frombytes(needle)
        feature_rms = math.sqrt(sum(v*v for v in feature)/len(feature))
        if feature_rms > .001:
            break
    else:
        raise AssertionError(name + '沒有足夠振幅的波形特徵')
    found = raw.find(needle)
    assert found >= 0 and found % 8 == 0, name + '沒有找到正式解碼波形'
    source_begin, monitor_begin = source_offset, found
    while source_begin >= 8 and monitor_begin >= 8 and source[source_begin-8:source_begin] == raw[monitor_begin-8:monitor_begin]:
        source_begin -= 8; monitor_begin -= 8
    source_end, monitor_end = source_offset+len(needle), found+len(needle)
    while source_end+8 <= len(source) and monitor_end+8 <= len(raw) and source[source_end:source_end+8] == raw[monitor_end:monitor_end+8]:
        source_end += 8; monitor_end += 8
    seconds = (monitor_end-monitor_begin)/352800
    assert seconds >= 3, (name, seconds)
    record = {'track': name, 'ogg_sha256': row['ogg_sha256'], 'decoded_source_sha256': row['pcm_sha256'],
              'feature_source_seconds': second, 'feature_rms': feature_rms,
              'source_frames': [source_begin//8, source_end//8],
              'monitor_frames': [monitor_begin//8, monitor_end//8],
              'continuous_exact_bytes': monitor_end-monitor_begin, 'continuous_exact_seconds': seconds,
              'different': 0, 'complete_track_verified': False,
              'scope': '正常Windows實際輸出的連續內段，非整首／loop／實體卡或人耳'}
    if markers is not None:
        phase = {'OPEN_M': ('intro','main'), 'BGM_0F': ('fm0','fm1'), 'BGM_1F': ('fm1','off')}[name]
        begin, end = markers[phase[0]], markers[phase[1]]
        overlap = min(monitor_end/352800, end)-max(monitor_begin/352800, begin)
        assert overlap >= 3, (name, overlap)
        record['overlap_with_requested_state_seconds'] = overlap
    clip_begin = monitor_begin+min(44100*8, (monitor_end-monitor_begin)//4//8*8)
    clip_end = min(monitor_end, clip_begin+5*44100*8)
    floats = array.array('f'); floats.frombytes(raw[clip_begin:clip_end])
    samples = array.array('h', (max(-32768,min(32767,round(v*32768))) for v in floats))
    clip = out / (name+'-monitor.wav')
    assert not clip.exists()
    with wave.open(str(clip),'wb') as stream:
        stream.setnchannels(2); stream.setsampwidth(2); stream.setframerate(44100); stream.writeframes(samples.tobytes())
    record['listening_clip'] = {'path':str(clip), 'bytes':clip.stat().st_size,
                               'sha256':hashlib.sha256(clip.read_bytes()).hexdigest(), 'seconds':len(samples)/88200}
    records.append(record)
if markers is not None:
    for phase, end_phase in [('main','fm0'),('off',None)]:
        begin = markers[phase]+1
        end = min(markers[end_phase]-.5 if end_phase else begin+3, stats['seconds'])
        assert end-begin >= 2, (phase, begin, end)
        segment = values[round(begin*44100)*2:round(end*44100)*2]
        count = sum(v != 0 for v in segment)
        silent.append({'phase':phase,'monitor_seconds':[begin,end],'samples':len(segment),'nonzero':count})
        assert count == 0, (phase,count)
result = {'schema':'logh3.windows-music-monitor.v1','monitor':stats,'matching_sources':records,
          'wine_audio_driver': driver,
          'silent_states':silent,'endpoint_observations':len(active),
          'resample_methods':sorted({c['resample_method'] for c in active}),
          'physical_sound_card':False,'human_listening':False,'loop_verified':False}
with target.open('x') as stream:
    json.dump(result,stream,ensure_ascii=False,indent=2);stream.write('\n')
print('PASS',args.batch,[(r['track'],r['continuous_exact_seconds']) for r in records], 'silence',silent)
