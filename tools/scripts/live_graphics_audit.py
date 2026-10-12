"""核對正常F6實錄、原生動畫節奏、同次音訊及靜態畫面復原。Docker audio入口執行。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import statistics
import subprocess

from PIL import Image, ImageChops

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--run', required=True)
p.add_argument('--out', required=True)
a = p.parse_args()
root, output = Path(a.run), Path(a.out)
assert root.is_dir() and not output.exists()


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def summary(values):
    if not values:
        return None
    values = sorted(values)
    return dict(count=len(values), mean=statistics.mean(values),
                median=statistics.median(values), p95=values[int((len(values)-1)*.95)],
                maximum=max(values))


def cadence(rows, mode):
    selected = [r for r in rows if r['state'] == 'StateEngagement'
                and r['graphics'] == mode and r.get('animations')]
    groups, changes, ticks = [], [], []
    previous = None
    for row in rows:
        valid = row['state'] == 'StateEngagement' and row['graphics'] == mode and row.get('animations')
        if not valid:
            groups = []
            previous = None
            continue
        if previous and row['ticks'] > previous['ticks']:
            ticks.append((row['wall_ns']-previous['wall_ns'])/1e6/(row['ticks']-previous['ticks']))
        key = tuple(row['animations'])
        if not groups or groups[-1]['key'] != key:
            current = dict(key=key, wall=row['wall_ns'], tick=row['ticks'])
            if groups and len(groups[-1]['key']) == len(key):
                last = groups[-1]
                changes.append(dict(wall_ms=(current['wall']-last['wall'])/1e6,
                                    update_steps=current['tick']-last['tick']))
            groups.append(current)
        previous = row
    # Native time is independent of Ebitengine's update count after a drawing
    # stall. Include every continuous source change, including short segments
    # at an F6 boundary; never filter observations by update_steps.
    native = changes
    return dict(draw_ms=summary([r['draw_ms'] for r in selected]),
                game_frame_ms=summary([r['frame_ms'] for r in selected]),
                wall_ms_per_update=summary(ticks),
                source_change_ms=summary([c['wall_ms'] for c in native]),
                source_changes=native,
                limits='以交戰狀態、可見來源數及顯示模式分段，所有連續來源變化皆納入；F6邊界可能為部分間隔。update_steps是app更新數，不是DOS硬體tick。原版速度0每格8×17.164ms。')


trace = json.loads((root/'TRACE.json').read_text())
assert trace['normal_x11_input'] and not trace['script_flag'] and not trace['state_injection'] and not trace['seed_injection']
assert digest(root/'RAW.mp4') == trace['raw_sha256']
rows = [json.loads(line) for line in (root/'FRAMES.jsonl').read_text().splitlines()]
if trace.get('frame_trace_sha256'):
    closure = json.loads((root/'FRAME-TRACE-CLOSE.json').read_text())
    assert closure['finalized_after_child_termination']
    assert digest(root/'FRAMES.jsonl') == closure['frame_trace_sha256']
events = [e for e in trace['events'] if e.get('key') == 'F6']
assert len(events) >= 6
switches = []
for i, event in enumerate(events):
    begin = event['wall_begin_ns']
    before = [r for r in rows if r['wall_ns'] < begin][-1]
    after = next(r for r in rows if r['wall_ns'] >= begin and r['graphics'] != before['graphics'])
    delay = (after['wall_ns']-begin)/1e6
    assert 0 <= delay < 1500, (i, delay)
    start = max(0, event['start_seconds']-.1)
    audio = subprocess.run(['ffmpeg', '-nostdin', '-v', 'info', '-ss', str(start), '-t', '1.5',
                            '-i', str(root/'RAW.mp4'), '-vn', '-af', 'volumedetect', '-f', 'null', '-'],
                           check=True, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT).stdout
    mean = float(re.search(r'mean_volume: ([-.0-9]+)', audio).group(1))
    assert mean > -55, (i, mean)
    switches.append(dict(event=i, start_seconds=event['start_seconds'],
                         before=before['graphics'], after=after['graphics'],
                         state=after['state'], delay_ms=delay, audio_mean_db=mean))
assert {s['state'] for s in switches} >= {'StateCommandPhase', 'StateBattle', 'StateEngagement'}

restored = []
for before, after in [('loaded', 'map-hd'), ('battle', 'battle-hd')]:
    left = Image.open(root/(before+'.png')).convert('RGB')
    right = Image.open(root/(after+'.png')).convert('RGB')
    assert left.size == right.size == (1280, 960)
    box = ImageChops.difference(left, right).getbbox()
    assert box is None, (before, after, box)
    restored.append(dict(before=before, after=after, different_pixels=0))

probe = json.loads((root/'FFPROBE.json').read_text())
streams = {s['codec_type']:s for s in probe['streams']}
assert (streams['video']['width'], streams['video']['height']) == (1280, 960)
assert streams['video']['r_frame_rate'] in ('30/1', '60/1')
assert streams['audio']['channels'] == 2
data = dict(normal_input=True, display_switch_passed=True, restored=restored, switches=switches,
            cadence={mode:cadence(rows, mode) for mode in ['hd', 'original']},
            raw_sha256=digest(root/'RAW.mp4'), trace_sha256=digest(root/'TRACE.json'),
            frame_trace_sha256=digest(root/'FRAMES.jsonl'), binary_sha256=trace['binary_sha256'],
            capture_frame_rate=streams['video']['r_frame_rate'],
            fps_note='幀率是擷取格式；順滑聲明依實際來源格間隔及更新耗時，不由容器幀率推論。',
            not_claimed=['DOS全場同態', '補間動畫', '實體音效卡／人耳驗收'])
output.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print(json.dumps({k:data[k] for k in ['switches', 'restored', 'cadence']}, ensure_ascii=False, indent=2))
