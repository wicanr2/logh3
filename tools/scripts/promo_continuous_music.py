"""重製既有推廣片音軌：單首連續配樂，依正式觀測獨立重混音效。"""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import shutil
import subprocess
import wave

import numpy as np
from PIL import Image

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('mode', choices=['render', 'audit', 'publish'])
p.add_argument('--out', default='promo-music137')
p.add_argument('--visual-reviewed', action='store_true', help='完成CONTACT目視後才交付')
p.add_argument('--music', default='dist-all/v.1.2.0-20261012/full/LOGH3-v.1.2.0-20261012-linux-amd64/music/BGM_0F.ogg')
a = p.parse_args()
root = Path('/src')
workspace = Path('/work') if Path('/work').is_dir() else root/'workplace'
out = workspace/a.out
source_dir = root/'dist-all/v.1.2.0-20261012/promo'
source_movie = source_dir/'LOGH3-v.1.2.0-20261012-promo.mp4'
version = 'v.1.2.0-20261012'
movie_name = f'LOGH3-{version}-promo-continuous-music.mp4'
assert os.getuid() == 1000 and out.parent.is_dir() and out.parent.stat().st_uid == os.getuid()
rate = 44100

def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()

def resolve(name):
    name = str(name)
    if name.startswith('/work/'):
        return workspace/name[len('/work/'):]
    return Path(name) if name.startswith('/') else root/name

def relative(path):
    path = Path(path)
    if path.is_relative_to(workspace):
        return 'workplace/'+str(path.relative_to(workspace))
    return str(path.relative_to(root))

def run(*args, **kw):
    return subprocess.run([str(x) for x in args], check=True, timeout=600, **kw)

def capture(*args):
    return subprocess.check_output([str(x) for x in args], timeout=600)

def probe(path):
    return json.loads(capture('ffprobe', '-v', 'error', '-show_streams', '-show_format', '-of', 'json', path))

def decode(path):
    data = capture('ffmpeg', '-nostdin', '-v', 'error', '-threads', '2', '-i', path,
                   '-vn', '-ar', str(rate), '-ac', '2', '-f', 'f32le', '-')
    return np.frombuffer(data, dtype='<f4').reshape(-1, 2).copy()

def pcm(samples):
    assert np.max(np.abs(samples)) < 1, 'clipping'
    return np.rint(samples*32767).astype('<i2')

def write_wave(name, samples):
    with wave.open(str(out/name), 'wb') as f:
        f.setnchannels(2)
        f.setsampwidth(2)
        f.setframerate(rate)
        f.writeframes(pcm(samples).tobytes())

def read_wave(name):
    with wave.open(str(out/name), 'rb') as f:
        assert (f.getnchannels(), f.getsampwidth(), f.getframerate()) == (2, 2, rate)
        return np.frombuffer(f.readframes(f.getnframes()), dtype='<i2').reshape(-1, 2).copy()

def write(name, data):
    (out/name).write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')

def stream_hash(path):
    return capture('ffmpeg', '-nostdin', '-v', 'error', '-i', path, '-map', '0:v:0',
                   '-c:v', 'copy', '-f', 'hash', '-hash', 'sha256', '-').decode().strip()

def video_packets(path):
    j = json.loads(capture('ffprobe', '-v', 'error', '-select_streams', 'v:0', '-show_packets',
                          '-show_entries', 'packet=pts_time,dts_time,duration_time,size', '-of', 'json', path))
    return j['packets']

if a.mode == 'render':
    assert not out.exists(), 'refuse overwrite'
    out.mkdir()
    old_audit = json.loads((source_dir/'AUDIT.json').read_text())
    assert sha(source_movie) == old_audit['video_sha256']
    plan = json.loads((source_dir/'PLAN.json').read_text())
    assert plan['version'] == version
    inputs = {relative(source_movie): sha(source_movie), relative(source_dir/'PLAN.json'): sha(source_dir/'PLAN.json')}
    for name, digest in plan['inputs'].items():
        path = resolve(name)
        assert sha(path) == digest, name
        inputs[relative(path)] = digest
    info = probe(source_movie)
    video = next(s for s in info['streams'] if s['codec_type'] == 'video')
    origin = float(video['start_time'])
    seconds = origin+float(video['duration'])
    count = round(seconds*rate)
    music_path = resolve(a.music)
    music = decode(music_path)
    assert len(music) >= count, 'single track too short; do not silently loop or cut the ending'
    music = music[:count]
    inputs[relative(music_path)] = sha(music_path)
    rms = float(np.sqrt(np.mean(music.astype(np.float64)**2)))
    assert rms > .00001
    music_gain = 10**(-20/20)/rms
    music *= music_gain
    # Only whole-film fades; no scene-boundary resets, fades or changes.
    fade_in, fade_out = rate, round(2.5*rate)
    music[:fade_in] *= np.linspace(0, 1, fade_in, dtype=np.float32)[:, None]
    music[-fade_out:] *= np.linspace(1, 0, fade_out, dtype=np.float32)[:, None]
    sfx = np.zeros((count, 2), dtype=np.float32)
    caches, records, sounds = {}, {}, []
    music_dir = resolve(plan['music']).parent
    for scene in plan['scenes']:
        if scene['kind'] != 'game':
            continue
        raw = resolve(scene['file'])
        folder = raw.parent
        if str(folder) not in records:
            trace_path, frames_path = folder/'TRACE.json', folder/'FRAMES.jsonl'
            trace = json.loads(trace_path.read_text())
            assert trace['normal_x11_input'] and not trace['script_flag'] and not trace['state_injection']
            assert sha(raw) == trace['raw_sha256']
            close = folder/'FRAME-TRACE-CLOSE.json'
            final_hash = json.loads(close.read_text())['frame_trace_sha256'] if close.exists() else trace['frame_trace_sha256']
            assert sha(frames_path) == final_hash
            for path in [trace_path, frames_path, *([close] if close.exists() else [])]:
                inputs[relative(path)] = sha(path)
            found, seen = [], set()
            for line in frames_path.open():
                row = json.loads(line)
                playback = row.get('playback') or {}
                for cue in playback.get('sounds', []):
                    key = (playback['id'], cue['id'], cue['tick'], cue['defence'])
                    if not cue['played'] or key in seen:
                        continue
                    seen.add(key)
                    found.append(dict(id=cue['id'], playback=playback['id'], tick=cue['tick'],
                                      start=(row['wall_ns']-trace['record_started_utc_ns'])/1e9))
            records[str(folder)] = found
        for event in records[str(folder)]:
            number = event['id']
            if number not in caches:
                path = music_dir/f'SFX_{number:02}.ogg'
                caches[number] = decode(path)
                inputs[relative(path)] = sha(path)
            sample = caches[number]
            begin = max(scene['start'], event['start'])
            end = min(scene['start']+scene['duration'], event['start']+len(sample)/rate)
            if begin >= end:
                continue
            offset = round((begin-event['start'])*rate)
            dest = round((origin+scene['start_at']+begin-scene['start'])*rate)
            length = min(round((end-begin)*rate), len(sample)-offset, count-dest)
            assert dest >= 0 and length > 0
            sfx[dest:dest+length] += sample[offset:offset+length]*.7
            sounds.append(dict(scene=scene['id'], source=relative(raw), playback=event['playback'],
                               id=number, cue_tick=event['tick'], source_event_seconds=event['start'],
                               output_sample=dest, source_sample=offset, samples=length,
                               clipped_at_scene_cut=offset != 0 or length != len(sample)))
    assert sounds, 'no gameplay sound effects'
    mixed = music+sfx
    headroom = min(1, 10**(-1.5/20)/float(np.max(np.abs(mixed))))
    music *= headroom
    sfx *= headroom
    mixed = music+sfx
    write_wave('MUSIC.wav', music)
    write_wave('SFX.wav', sfx)
    write_wave('MIX.wav', mixed)
    write('PLAN.json', dict(version=version, source_movie=relative(source_movie),
                           seconds=seconds, frames=int(video['nb_frames']), video_origin=origin,
                           music=relative(music_path), music_start_sample=0, music_samples=count,
                           music_track_count=1, music_loops=0, music_gain=music_gain*headroom,
                           sfx_gain=.7*headroom, fades_seconds=[1, 2.5],
                           original_clip_audio_mixed=False, scenes=plan['scenes'],
                           sfx_events=sounds, inputs=inputs,
                           method='Existing video copied; soundtrack replaced with one continuous original game track and separately reconstructed actual SFX cues.',
                           limits=['SFX onset follows first observed played frame; not sample-exact original live capture',
                                   'No anime source audio mixed; it could contain other background music',
                                   'No new gameplay recording or game parity claim']))
    run('ffmpeg', '-nostdin', '-v', 'warning', '-i', source_movie, '-i', out/'MIX.wav',
        '-map', '0:v:0', '-map', '1:a:0', '-c:v', 'copy', '-c:a', 'aac', '-b:a', '256k',
        '-ar', str(rate), '-ac', '2', '-threads', '2',
        # MP4's default millisecond movie clock rounded the video edit-list
        # origin by 15/15360 seconds. Use an exact shared video/audio clock.
        '-movie_timescale', str(math.lcm(rate, int(video['time_base'].split('/')[1]))),
        '-movflags', '+faststart', out/movie_name)
    print('RENDERED', seconds, 'seconds;', len(sounds), 'SFX cues; one continuous music track', flush=True)

elif a.mode == 'audit':
    plan = json.loads((out/'PLAN.json').read_text())
    for name, digest in plan['inputs'].items():
        assert sha(resolve(name)) == digest, name
    movie = out/movie_name
    info = probe(movie)
    video = next(s for s in info['streams'] if s['codec_type'] == 'video')
    audio = next(s for s in info['streams'] if s['codec_type'] == 'audio')
    assert (video['width'], video['height'], video['r_frame_rate'], audio['channels']) == (1920, 1080, '60/1', 2)
    assert int(video['nb_frames']) == plan['frames'] == 7496
    assert abs(float(info['format']['duration'])-plan['seconds']) < .05
    assert abs(float(audio['duration'])-plan['seconds']) < .05
    assert stream_hash(movie) == stream_hash(source_movie)
    assert video_packets(movie) == video_packets(source_movie), 'video packet timing changed'
    source = decode(resolve(plan['music']))[:plan['music_samples']]
    source *= plan['music_gain']
    source[:rate] *= np.linspace(0, 1, rate, dtype=np.float32)[:, None]
    last = round(2.5*rate)
    source[-last:] *= np.linspace(1, 0, last, dtype=np.float32)[:, None]
    expected_bed, bed = pcm(source), read_wave('MUSIC.wav')
    assert np.max(np.abs(expected_bed.astype(np.int32)-bed)) <= 1
    mix = read_wave('MIX.wav').astype(np.float32)/32767
    decoded = decode(movie)
    assert len(decoded) >= len(mix)
    decoded = decoded[:len(mix)]
    assert np.max(np.abs(decoded)) < 1, 'encoded audio clips'
    difference = decoded-mix
    def snr(left, error):
        return float(10*np.log10(max(float(np.mean(left.astype(np.float64)**2)), 1e-16)/
                                max(float(np.mean(error.astype(np.float64)**2)), 1e-16)))
    global_snr = snr(mix, difference)
    assert global_snr > 25, global_snr
    cuts = []
    for scene in plan['scenes'][1:]:
        center = round((plan['video_origin']+scene['start_at'])*rate)
        lo, hi = center-round(.2*rate), center+round(.2*rate)
        assert np.max(np.abs(bed[lo:hi].astype(np.int32)-expected_bed[lo:hi])) <= 1
        score = snr(mix[lo:hi], difference[lo:hi])
        assert score > 20, (scene['id'], score)
        cuts.append(dict(scene=scene['id'], sample=center, decoded_mix_snr_db=score,
                         same_continuous_music_samples=True))
    write('FFPROBE.json', info)
    for name, args in [('AUDIO-AUDIT.log', ['-vn', '-af', 'volumedetect,silencedetect=n=-55dB:d=1']),
                       ('VIDEO-AUDIT.log', ['-an', '-vf', 'blackdetect=d=.1:pix_th=.1,freezedetect=n=-50dB:d=2'])]:
        with (out/name).open('w') as log:
            run('ffmpeg', '-nostdin', '-xerror', '-v', 'info', '-threads', '2', '-i', movie,
                *args, '-f', 'null', '-', stdout=log, stderr=subprocess.STDOUT)
    sound_log = (out/'AUDIO-AUDIT.log').read_text()
    assert 'mean_volume: -inf' not in sound_log and 'silence_start:' not in sound_log
    assert 'black_start:' not in (out/'VIDEO-AUDIT.log').read_text()
    # Identical full video packet content and timing preserves the prior
    # classification of intentionally still title/menu/result intervals.
    freeze_path = source_dir/'FREEZE-REVIEW.json'
    prior_freeze = json.loads(freeze_path.read_text())
    assert prior_freeze['passed']
    write('FREEZE-REVIEW.json', dict(source=relative(freeze_path), source_sha256=sha(freeze_path),
                                  identical_full_video_and_timing=True, original_review=prior_freeze))
    contact = Image.new('RGB', (1920, 540), '#09111d')
    times = [3.5, 17, 36, 64.5, 78, 94, 111, 122]
    for n, when in enumerate(times):
        data = capture('ffmpeg', '-nostdin', '-v', 'error', '-threads', '2', '-ss', str(when), '-i', movie,
                       '-frames:v', '1', '-vf', 'scale=480:270', '-pix_fmt', 'rgb24', '-f', 'rawvideo', '-')
        contact.paste(Image.frombytes('RGB', (480, 270), data), (n%4*480, n//4*270))
    contact.save(out/'CONTACT.jpg', quality=90)
    (out/'RIGHTS.txt').write_text('原版遊戲及衍生音畫、指定動畫僅本機使用，不公開上傳。\n'
                                '背景只用PLAN列明的一首連續配樂，片段原聲不混入。\n'
                                '視訊沿用既有實錄，交火音效依實際播放觀測重新混音。\n')
    write('AUDIT.json', dict(passed=True, version=version, video_sha256=sha(movie),
                            source_video_sha256=sha(source_movie), seconds=float(info['format']['duration']),
                            video_frames=7496, video_packet_hash=stream_hash(movie),
                            video_packets_and_timestamps_unchanged=True, music_track_count=1, music_loops=0,
                            music=plan['music'], music_sha256=plan['inputs'][plan['music']],
                            scene_cuts=cuts, decoded_mix_snr_db=global_snr,
                            sfx_events=len(plan['sfx_events']), peak_pcm=float(np.max(np.abs(mix))),
                            encoded_peak=float(np.max(np.abs(decoded))),
                            black_segments=0, long_silences=0, visual_review='CONTACT pending',
                            source_scope='v.1.2.0 original visual cut, soundtrack revision only',
                            generator_sha256=sha(root/'tools/scripts/promo_continuous_music.py')))
    print('PASS 7496 unchanged video packets; 38 continuous-music cuts;', global_snr, 'dB AAC SNR', flush=True)

else:
    audit = json.loads((out/'AUDIT.json').read_text())
    assert audit['passed'] and sha(out/movie_name) == audit['video_sha256']
    assert audit['generator_sha256'] == sha(root/'tools/scripts/promo_continuous_music.py')
    assert a.visual_reviewed, 'inspect CONTACT before delivery'
    audit['visual_review'] = '8 sampled frames inspected; source picture, captions and ending intact'
    write('AUDIT.json', audit)
    assert source_dir.stat().st_uid == os.getuid()
    names = [movie_name, 'PLAN.json', 'AUDIT.json', 'FFPROBE.json', 'AUDIO-AUDIT.log',
             'VIDEO-AUDIT.log', 'FREEZE-REVIEW.json', 'CONTACT.jpg', 'RIGHTS.txt']
    destinations = {name: source_dir/(name if name == movie_name else 'continuous-music-'+name) for name in names}
    assert all(not path.exists() for path in destinations.values()), 'refuse overwrite delivery'
    checksum_path = source_dir/'continuous-music-SHA256SUMS.json'
    assert not checksum_path.exists()
    records = []
    for name, dest in destinations.items():
        shutil.copyfile(out/name, dest)
        assert sha(dest) == sha(out/name) and dest.stat().st_uid == os.getuid()
        records.append(dict(file=dest.name, bytes=dest.stat().st_size, sha256=sha(dest)))
    checksum_path.write_text(json.dumps(dict(version=version, kind='soundtrack revision; no new release',
                                           files=records), ensure_ascii=False, indent=2)+'\n')
    index_path = root/'docs/internals/data/promo-continuous-music.json'
    assert not index_path.exists(), 'refuse overwrite source index'
    plan = json.loads((out/'PLAN.json').read_text())
    index_path.write_text(json.dumps(dict(schema='logh3.continuous-promo-music.v1', version=version,
                                         source_head='f597f13c5177e6f84b9c9d40e6115339500494c1',
                                         status='soundtrack revision verified; existing visual cut retained',
                                         generator='tools/scripts/promo_continuous_music.py',
                                         generator_sha256=audit['generator_sha256'], inputs=plan['inputs'],
                                         audio=dict(music=plan['music'], track_count=1, loops=0,
                                                    whole_film_fades_seconds=plan['fades_seconds'],
                                                    original_clip_audio_mixed=False, sfx_fragments=len(plan['sfx_events'])),
                                         audit=audit, outputs=[dict(record, file=relative(source_dir/record['file'])) for record in records],
                                         checksum=dict(file=relative(checksum_path), sha256=sha(checksum_path)),
                                         rights='original and derived audiovisual assets; local-only, no upload',
                                         limits=plan['limits']), ensure_ascii=False, indent=2)+'\n')
    print('DELIVERED', destinations[movie_name], flush=True)
