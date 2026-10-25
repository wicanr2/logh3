"""165：正常GUI的25組名言／語系與已核准透明層比對。Docker內執行。"""
import hashlib
import json
from pathlib import Path
import re

import numpy as np
from PIL import Image

root=Path('/src');work=root/'workplace/poster165';art=root/'assets/hd/poster/approved'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
result=json.loads((work/'alliance/RESULT.json').read_text())
assert result['passed'] and result['full_cycle'] and not result['time_injection']
base=Image.open(art/'base-v1.png').convert('RGBA')
mask=np.zeros((720,1280),bool)
mask[440:640,:217]=True;mask[440:640,1064:]=True
rows=[]
for event in result['events']:
    match=re.fullmatch(r'group-([1-5])-(zh-TW|zh-CN|ja|en|ko)',event['name'])
    if not match:continue
    group,locale=int(match[1]),match[2]
    observed=event['observed']
    assert observed['locale']==locale and observed['poster_decoration']['quote_group']==group
    image=base.copy()
    if group==1:
        image.alpha_composite(Image.open(art/f'labels-{locale}.png').convert('RGBA'))
    else:
        image.alpha_composite(Image.open(art/f'quote-{group}-{locale}.png').convert('RGBA'),(0,880))
    expected=np.asarray(image.convert('RGB').resize((1280,720),Image.Resampling.NEAREST),dtype=np.int16)
    path=work/'alliance'/event['screenshot']
    actual=np.asarray(Image.open(path).convert('RGB'),dtype=np.int16)
    delta=np.abs(actual-expected);difference=delta.max(2)
    assert int(difference[mask].max())<=3,(group,locale,int(difference[mask].max()))
    diff=work/f'diff-quote-{group}-{locale}.png'
    Image.fromarray(np.where(mask[:,:,None],np.minimum(delta*10,255),0).astype('uint8')).save(diff)
    rows.append(dict(group=group,locale=locale,pixels=int(mask.sum()),different=int(((difference>0)&mask).sum()),
        max_channel=int(difference[mask].max()),actual_sha256=sha(path),diff_sha256=sha(diff)))
assert len(rows)==25 and len({(r['group'],r['locale'])for r in rows})==25
# Group posters never receive the v1 text layer.
for event in result['events']:
    if event['name'] not in ['v2-no-quotes','v3-no-quotes']:continue
    name=event['name'][:2]
    expected=np.asarray(Image.open(art/f'base-{name}.png').convert('RGB').resize((1280,720),Image.Resampling.NEAREST),dtype=np.int16)
    actual=np.asarray(Image.open(work/'alliance'/event['screenshot']).convert('RGB'),dtype=np.int16)
    assert not np.abs(actual-expected)[mask].any(),'群像混入名言'
report=dict(passed=True,comparison='approved-remake-text, not official translation or DOS parity',rows=rows,
    sampler='Ebitengine 2:1 nearest odd pixels; alpha rounding <=3 from approved-frame check',
    source_sha256=sha(root/'tools/scripts/poster_quote_candidates.json'),manifest_sha256=sha(art/'quotes.json'))
(work/'CHECK.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print('PASS 25 quote/locale images; no v1 labels on v2/v3; max',max(r['max_channel']for r in rows))
