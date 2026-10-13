"""沿既有MDS匯出器保存原曲MIDI事件，供HD音色合成；輸出只留本機。"""
import argparse
import hashlib
import json
import os
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--orig', type=Path, required=True)
parser.add_argument('--out', type=Path, required=True)
args = parser.parse_args()
assert args.orig.is_dir() and args.out.parent.stat().st_uid == os.getuid()
args.out.mkdir(exist_ok=True)
source = Path('/src/tools/scripts/mds_export.py').read_text()
scope = {}
exec(compile(source[:source.index('\nos.makedirs(OUT_JSON')], 'mds_export.py', 'exec'), scope)
rows = []
for path in sorted(args.orig.glob('*M.MDS')):
    data = path.read_bytes()
    title, chans, tracks, unknown = scope['parse'](data)
    barlen = scope['bar_length'](tracks)
    bars = [e[1] for t in tracks for e in t['events'] if e[0] == 'bar']
    sane = bool(bars) and sum(v <= 2 for v in bars)/len(bars) > .9
    midi = scope['midi'](title, tracks, barlen, sane)
    name = path.stem
    (args.out/(name+'.mid')).write_bytes(midi)
    rows.append({'track':name,'title':title,'mds_sha256':hashlib.sha256(data).hexdigest(),
                 'midi_sha256':hashlib.sha256(midi).hexdigest(),
                 'notes':sum(e[0]=='note' for t in tracks for e in t['events']),
                 'classification':'原版MIDI樂譜，不更動音符、速度或事件；HD只換音色與混音'})
assert len(rows) == 11, len(rows)
(args.out/'SOURCE.json').write_text(json.dumps(rows,ensure_ascii=False,indent=2)+'\n')
print('保留原曲MIDI事件：',len(rows),'首')
