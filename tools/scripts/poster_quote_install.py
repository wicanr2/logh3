"""165：核對並安裝已核准的名言透明層，原藝術層保持。僅Docker內執行。"""
import hashlib
import json
import os
from pathlib import Path
import shutil

from PIL import Image

root=Path('/src')
work=root/'workplace/poster165'
dest=root/'assets/hd/poster/approved'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
assert dest.is_dir() and dest.stat().st_uid==os.getuid()
old=json.loads((dest/'manifest.json').read_text())
before={name:sha(dest/name) for name in old['files']}
assert all(before[name]==item['sha256'] for name,item in old['files'].items())
data=root/'tools/scripts/poster_quote_candidates.json'
source=json.loads(data.read_text())
m=json.loads((work/'art/quotes.json').read_text())
assert source['status']=='user-approved-2026-10-26'
assert m['format']=='logh3-poster-quotes-v1' and m['interval_seconds']==60
assert m['source_sha256']==sha(data) and m['proposal_sha256']==old['proposal_sha256']
assert m['rect']==[0,880,2560,1280] and len(m['sets'])==4
files={}
for i,s in enumerate(m['sets']):
    assert s['id']==source['pairs'][i+1]['id']
    assert set(s['files'])=={'zh-TW','zh-CN','ja','en','ko'}
    for locale,item in s['files'].items():
        name=f'quote-{i+2}-{locale}.png'
        assert item['file']==name and name not in before
        path=work/'art'/name
        assert sha(path)==item['sha256']
        with Image.open(path) as image:
            assert image.size==(2560,400) and image.mode=='RGBA'
        files[name]=item['sha256']
for name in [*files,'quotes.json']:
    target=dest/name
    assert not target.exists() or target.stat().st_uid==os.getuid()
    shutil.copyfile(work/'art'/name,target)
    assert target.stat().st_uid==os.getuid()
assert before=={name:sha(dest/name) for name in before}
(work/'INSTALL.json').write_text(json.dumps(dict(passed=True,files=files,source_sha256=sha(data),
    original_unchanged=before,manifest_sha256=sha(dest/'quotes.json')),ensure_ascii=False,indent=2)+'\n')
print('PASS installed20; original9 unchanged')
