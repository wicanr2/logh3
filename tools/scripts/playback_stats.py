"""核PulseAudio monitor全部float32樣本；不把技術檢測當真人聆聽。"""
import array
import hashlib
import json
import math
from pathlib import Path
import sys

root=Path(sys.argv[1])
path=root/'monitor.f32le'
assert path.stat().st_size%8==0 and path.stat().st_size>0
target=root/'monitor-stats.json'
assert not target.exists()
count=nonzero=clipped=0
peak=squares=0.0
first=last=None
digest=hashlib.sha256()
with path.open('rb') as f:
    while b:=f.read(65536):
        digest.update(b)
        values=array.array('f');values.frombytes(b)
        if sys.byteorder!='little':values.byteswap()
        for v in values:
            assert math.isfinite(v),'輸出含NaN／Inf'
            a=abs(v)
            if a>0:
                if first is None:first=count
                last=count
                nonzero+=1
            if a>=1:clipped+=1
            peak=max(peak,a);squares+=v*v;count+=1
result={'classification':'正式ALSA經PulseAudio虛擬輸出的技術樣本，非真人聆聽',
        'bytes':path.stat().st_size,'sha256':digest.hexdigest(),'samples':count,'sample_rate':44100,'channels':2,
        'duration_seconds':count/2/44100,'nonzero_samples':nonzero,'peak':peak,'rms':math.sqrt(squares/count),
        'samples_at_or_above_one':clipped,'first_nonzero_seconds':None if first is None else first/2/44100,
        'last_nonzero_seconds':None if last is None else last/2/44100}
target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(result,ensure_ascii=False))
