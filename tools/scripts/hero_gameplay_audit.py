"""規格140：核對正常S1實錄、公開輸入模型及兩種AI的同初始狀態。"""
import bisect
import hashlib
import json
from pathlib import Path

ROOT = Path('/src')
BATCH = ROOT/'workplace/hero130'
VERSION = 'v.1.2.0-20261012'


def read(path):
    return json.loads(path.read_text())


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1024*1024), b''):
            h.update(block)
    return h.hexdigest()


results = []
states = {}
for name, model, faction, level in [
    ('full-record-f0', 'full-s1-f0-ai1', 0, 1),
    ('full-record-f1-r2', 'full-s1-f1-ai1', 1, 1),
    ('record-f0-ai-1', 'normal-s1-f0-ai-1', 0, -1),
    ('record-f1-ai-1', 'normal-s1-f1-ai-1', 1, -1),
]:
    run = BATCH/name
    trace = read(run/'TRACE.json')
    assert trace['normal_x11_input'] and not any(trace[k] for k in ('state_injection', 'seed_injection', 'script_flag'))
    assert trace['ai'] == level
    assert trace['binary_sha256'] == sha(ROOT/'dist-all'/VERSION/'full'/('LOGH3-'+VERSION+'-linux-amd64')/'gin3')
    assert trace['raw_sha256'] == sha(run/'RAW.mp4')
    assert read(run/'FRAME-TRACE-CLOSE.json')['frame_trace_sha256'] == sha(run/'FRAMES.jsonl')
    rows = [json.loads(line) for line in (run/'FRAMES.jsonl').open()]
    times = [r['wall_ns'] for r in rows]
    expected = read(BATCH/model/'STATES.json')
    captures = [e for e in trace['events'] if e['kind'] == 'capture']
    assert len(expected) == len(captures)
    matched = []
    for event, want in zip(captures, expected):
        if want is None:
            continue
        i = bisect.bisect_left(times, event['wall_begin_ns'])
        row = min(rows[max(0,i-1):i+2], key=lambda r:abs(r['wall_ns']-event['wall_begin_ns']))
        got = row.get('tactical')
        assert got == want, (name,event['name'])
        assert row['show_enemies']
        assert got['commanders'] == (['萊因哈特','楊威利'] if level >= 0 else ['',''])
        assert all(f['command_boost'] == (level >= 0) for f in got['fleets'])
        matched.append({'capture':event['name'],'turn':got['turn'],'fleets':len(got['fleets']),
                        'video_seconds':(event['wall_begin_ns']-trace['record_started_utc_ns'])/1e9})
        states[name,event['name']] = got
    assert [m['turn'] for m in matched] == list(range(1, (6 if level >= 0 else 4)))
    initial = states[name,'entry']
    final = states[name,matched[-1]['capture']]
    moving = [f['fleet'] for f in final['fleets'] if f['faction'] != faction and
              any(o['fleet']==f['fleet'] and (o['x'],o['y'])!=(f['x'],f['y']) for o in initial['fleets'])]
    assert moving, name
    results.append({'recording':name,'model':model,'faction':faction,'ai':level,
                    'matched_captures':matched,'enemy_moved':moving,'passed':True,
                    'frame_trace_sha256':sha(run/'FRAMES.jsonl'),'raw_sha256':sha(run/'RAW.mp4')})

comparisons = []
for faction, enhanced in [(0,'full-record-f0'),(1,'full-record-f1-r2')]:
    original = f'record-f{faction}-ai-1'
    def physical(s):
        return [{k:v for k,v in f.items() if k!='command_boost'} for f in s['fleets']]
    assert physical(states[enhanced,'entry']) == physical(states[original,'entry'])
    assert physical(states[enhanced,'turn-02']) != physical(states[original,'turn-02'])
    comparisons.append({'player_faction':faction,'enemy_hero':['楊威利','萊因哈特'][faction],
                        'same_initial_fleets':True,'different_turn_2':True,
                        'level':'remake同初始資料及等價輸入；不稱DOS原版same-state'})

switch = BATCH/'record-f0-ai1'
trace = read(switch/'TRACE.json')
rows = [json.loads(line) for line in (switch/'FRAMES.jsonl').open()]
keys = [e for e in trace['events'] if e.get('key')=='F7']
assert len(keys)==2
sequence = []
for row in rows:
    if not sequence or sequence[-1]!=row['graphics']:
        sequence.append(row['graphics'])
assert sequence == ['hd','original','hd'], sequence
result = {'version':VERSION,'passed':True,'normal_entry':True,'state_injection':False,
          'recordings':results,'comparisons':comparisons,'f7_sequence':sequence,
          'f7_recording':str(switch.relative_to(ROOT)),
          'limits':['正常錄影與公開輸入模型逐回合相同；模型不取代實錄',
                    '觸發取消、1.5倍距離與保存邊界另由標準測試驗證',
                    '原版規則模式指remake既有模式，不宣稱DOS全局同態']}
(Path('/work/hero130')/'GAMEPLAY-AUDIT.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('PASS 四條正常玩家實錄、16個戰術狀態、兩側同初始比較及F7實際雙向切換')
