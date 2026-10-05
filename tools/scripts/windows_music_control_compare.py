"""核固定8秒來源的環境對照；不把局部後端診斷升格為玩家驗收。"""
import array
import hashlib
import json
import math
from pathlib import Path
import sys

assert Path.cwd() == Path('/src')
batch = sys.argv[1]
assert Path(batch).name == batch
root = Path('workplace/music69-windows')
out = root / batch
assert json.loads((out / 'CONTROL.json').read_text())['exit'] == 0
source = (root / 'playeraudio69-r105-decode/OPEN_M.f32le').read_bytes()[:8*44100*8]
raw = (out / 'monitor.f32le').read_bytes()
assert len(source) == 8*44100*8 and len(raw) % 8 == 0
found = raw.find(source)
offset = 2*44100*8
needle = source[offset:offset+16384]
feature = array.array('f'); feature.frombytes(needle)
assert math.sqrt(sum(v*v for v in feature)/len(feature)) > .001
feature_found = raw.find(needle)
sb = se = mb = me = 0
if feature_found >= 0:
    assert feature_found % 8 == 0
    sb, mb = offset, feature_found
    se, me = offset+len(needle), feature_found+len(needle)
    while sb >= 8 and mb >= 8 and source[sb-8:sb] == raw[mb-8:mb]:
        sb -= 8; mb -= 8
    while se+8 <= len(source) and me+8 <= len(raw) and source[se:se+8] == raw[me:me+8]:
        se += 8; me += 8
result = {'source_sha256': hashlib.sha256(source).hexdigest(),
          'monitor_sha256': hashlib.sha256(raw).hexdigest(),
          'full_source_found': found >= 0, 'full_source_monitor_frame': found//8 if found >= 0 else None,
          'feature_found': feature_found >= 0, 'continuous_exact_seconds': (me-mb)/352800,
          'source_frames': [sb//8, se//8], 'monitor_frames': [mb//8, me//8],
          'gui_used': False, 'production_player_verified': False}
with (out / 'CONTROL-CHECK.json').open('x') as stream:
    json.dump(result, stream, indent=2)
print(json.dumps(result, indent=2))
