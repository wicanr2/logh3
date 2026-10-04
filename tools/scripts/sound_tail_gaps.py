"""保留全檔比較失敗，另核唯一插入靜音前後的完整來源是否都保留。"""
import argparse
import array
import json
from pathlib import Path
import sys

p=argparse.ArgumentParser()
p.add_argument('root')
a=p.parse_args()
root=Path(a.root)
target=root/'pcm-gap-classification.json'
assert not target.exists()
raw=(root/'monitor.f32le').read_bytes()
comparison=json.loads((root/'pcm-comparison.json').read_text())
cases=[]
for c in comparison['cases']:
    case={'id':c['id'],'complete_exact':c['complete_exact']}
    if c['complete_exact']:
        case['complete_source_retained']=True
    elif c['marker_found']:
        values=array.array('h');values.frombytes(Path(c['source']).read_bytes())
        if sys.byteorder!='little':values.byteswap()
        values=array.array('f',(v/32768 for v in values))
        if sys.byteorder!='little':values.byteswap()
        expected=values.tobytes()
        begin=c['monitor_begin_frame']*8
        prefix=c['exact_prefix_frames']*8
        # 整段餘下PCM須相同；僅找到局部簽章不算尾端保留。
        shifted=raw.find(expected[prefix:],begin+prefix)
        gap=raw[begin+prefix:shifted] if shifted>=0 else b''
        retained=(shifted>=0 and shifted%8==0 and shifted>begin+prefix
                  and not any(gap) and raw[begin:begin+prefix]==expected[:prefix])
        case.update(complete_source_retained=retained,
                    inserted_silent_frames=len(gap)//8 if retained else None,
                    retained_tail_frames=(len(expected)-prefix)//8 if retained else None,
                    source_change_applied=False)
    else:case['complete_source_retained']=False
    cases.append(case)
result={'monitor_sha256':comparison['monitor_sha256'],
        'classification':'全段逐位元組門檻不放寬；插入靜音另列，成因待驗',
        'complete_exact':sum(c['complete_exact'] for c in cases),
        'complete_source_retained':sum(c['complete_source_retained'] for c in cases),
        'calls':len(cases),'cases':cases}
target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(result,ensure_ascii=False))
