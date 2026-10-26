"""157：上下旗艦景深隊列及人物內嵌名言，自包含HTML。"""
import csv
import hashlib
import json
from pathlib import Path
import os
import re
from PIL import Image

import poster_closeup_html as asset

ROOT=Path('/src');OUT=ROOT/'workplace/poster157';sha=asset.sha
assert os.getuid()==os.getgid()==1000 and not OUT.exists()
assert (ROOT/'workplace').is_dir() and (ROOT/'workplace').stat().st_uid==1000
OUT.mkdir()
m=json.loads((ROOT/'assets/hd/animation/MANIFEST.json').read_text())
codes=bytes.fromhex(json.loads((ROOT/'docs/internals/data/engagement-panel.json').read_text())['animation']['flagship_codes'])
names=json.loads((ROOT/'docs/internals/data/flagships.json').read_text())['tables']
wanted=['45B1','45B2','45B3','45B4','45B5','45B6','46BD','46B1','46B3','46B5','46B7','46B8']
ships=[];inputs=[]
for key in wanted:
    atlas,cell=next((a,c) for a in m['atlases'] for c in a['cells'] if c['key']==key)
    file=ROOT/'assets/hd/animation'/atlas['file'];assert sha(file)==atlas['sha256']
    x,y,w,h=[cell[k] for k in ['x','y','w','h']]
    image,bounds=asset.main_hull(Image.open(file).convert('RGBA').crop((x,y,x+w,y+h)))
    image.save(OUT/(key+'.png'));side=0 if key.startswith('45') else 1
    index=next(i for i in range(16) if codes[side*40+i]==int(key[2:],16))
    ships.append(dict(key=key,name=names[side][index]['text'],side=side,width=image.width,height=image.height,src=asset.encoded(image)))
    inputs.append(dict(key=key,atlas=str(file.relative_to(ROOT)),atlas_sha256=sha(file),cell=[x,y,w,h],bounds=list(bounds),png_sha256=sha(OUT/(key+'.png'))))
portraits=[]
for side,name in enumerate(['empire.png','alliance.png']):
    file=ROOT/'assets/hd/poster'/name;im=Image.open(file).convert('RGBA')
    portraits.append(dict(side=side,width=im.width,height=im.height,src=asset.encoded(im)))
    inputs.append(dict(portrait=str(file.relative_to(ROOT)),sha256=sha(file)))
catalog=ROOT/'internal/localize/ui.tsv'
with catalog.open() as f:
    quotes={r['key']:{k:v for k,v in r.items() if k!='key'} for r in csv.DictReader(f,delimiter='\t') if r['key'].startswith(('poster.quote.','poster.name.'))}
base=ROOT/'docs/images/readme-v162-poster.png'
data=dict(background=asset.encoded(Image.open(base).convert('RGB')),ships=ships,portraits=portraits,quotes=quotes)
template=ROOT/'tools/scripts/poster_flagship_rows.html'
shared=re.search(r'<style>([\s\S]*?)</style>',(ROOT/'tools/scripts/poster_closeup.html').read_text()).group(1)
page=template.read_text().replace('/*__SHARED_CSS__*/',shared).replace('/*__ROW_DATA__*/',json.dumps(data,ensure_ascii=False))
(OUT/'proposal.html').write_text(page)
(OUT/'SOURCE.json').write_text(json.dumps(dict(prototype=True,scope='上同盟朝左，下帝國朝右，隊列景深，名言與姓名嵌入人物圖',
    classification='含原版及衍生圖像，本機提案',html_sha256=sha(OUT/'proposal.html'),template_sha256=sha(template),
    generator_sha256=sha(Path(__file__)),shared_template_sha256=sha(ROOT/'tools/scripts/poster_closeup.html'),
    catalog_sha256=sha(catalog),background_sha256=sha(base),inputs=inputs,
    production_sha256=sha(ROOT/'cmd/gin3/poster_decoration.go'),poster_go_sha256=sha(ROOT/'cmd/gin3/poster.go')),
    ensure_ascii=False,indent=2)+'\n')
print('READY 上下旗艦隊列HTML', (OUT/'proposal.html').stat().st_size,flush=True)
