"""156：使用原始HD格建立自包含HTML提案，素材只留本機。"""
import base64
import argparse
import hashlib
import io
import json
import os
from pathlib import Path

import numpy as np
from PIL import Image,ImageFilter

ROOT=Path('/src');OUT=ROOT/'workplace/poster156'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def encoded(im):
    b=io.BytesIO();im.save(b,format='PNG');return 'data:image/png;base64,'+base64.b64encode(b.getvalue()).decode()
def main_hull(im):
    pixels=np.array(im);mask=np.max(pixels[:,:,:3],axis=2)>16
    height,width=mask.shape;flat=mask.ravel();seen=np.zeros(flat.size,dtype=bool);largest=[]
    for seed in np.flatnonzero(flat):
        if seen[seed]:continue
        stack=[int(seed)];seen[seed]=True;component=[]
        while stack:
            i=stack.pop();component.append(i);x=i%width
            for j in ([i-1] if x else [])+([i+1] if x+1<width else [])+([i-width] if i>=width else [])+([i+width] if i+width<flat.size else []):
                if flat[j] and not seen[j]:seen[j]=True;stack.append(j)
        if len(component)>len(largest):largest=component
    keep=np.zeros(flat.size,dtype=np.uint8);keep[largest]=255
    matte=Image.fromarray(keep.reshape(height,width)).filter(ImageFilter.MaxFilter(7))
    pixels[:,:,3]=np.where((np.array(matte)>0)&(np.max(pixels[:,:,:3],axis=2)>6),255,0)
    result=Image.fromarray(pixels);bounds=result.getbbox();assert bounds is not None
    return result.crop(bounds),bounds
def main():
    assert os.getuid()==os.getgid()==1000 and not OUT.exists()
    assert (ROOT/'workplace').is_dir() and (ROOT/'workplace').stat().st_uid==1000
    OUT.mkdir();manifest=ROOT/'assets/hd/animation/MANIFEST.json';m=json.loads(manifest.read_text())
    codes=bytes.fromhex(json.loads((ROOT/'docs/internals/data/engagement-panel.json').read_text())['animation']['flagship_codes'])
    names=json.loads((ROOT/'docs/internals/data/flagships.json').read_text())['tables']
    wanted=['45B1','45B2','45B3','46BD','46B1','46B3'];ships=[];records=[]
    for key in wanted:
        atlas,cell=next((a,c) for a in m['atlases'] for c in a['cells'] if c['key']==key)
        file=ROOT/'assets/hd/animation'/atlas['file'];assert sha(file)==atlas['sha256']
        x,y,w,h=(cell[k] for k in ['x','y','w','h']);im=Image.open(file).convert('RGBA').crop((x,y,x+w,y+h))
        # 原格另有小型陪襯艦，以最大連通主體取出主艦；不經192×76中間圖。
        im,bound=main_hull(im);im.save(OUT/(key+'.png'))
        side=0 if key.startswith('45') else 1;code=int(key[2:],16)
        index=next((i for i in range(16) if codes[side*40+i]==code),None)
        label=names[side][index]['text'] if index is not None else key
        ships.append(dict(key=key,name=label,side=side,width=im.width,height=im.height,src=encoded(im)))
        records.append(dict(key=key,atlas=str(file.relative_to(ROOT)),atlas_sha256=sha(file),cell=[x,y,w,h],crop=list(bound),
            extraction='最大連通主艦，邊緣擴3像素保留輪廓；只供構圖原型',png_sha256=sha(OUT/(key+'.png'))))
    file=ROOT/'assets/hd/poster/fighters.png';im=Image.open(file).convert('RGBA');fighters=[]
    for side in range(2):
        cell=im.crop((side*im.width//2,0,(side+1)*im.width//2,im.height));alpha=np.array(cell)[:,:,3]
        yy,xx=np.where(alpha>=192);bounds=(max(0,int(xx.min())-8),max(0,int(yy.min())-8),min(cell.width,int(xx.max())+9),min(cell.height,int(yy.max())+9))
        cell=cell.crop(bounds);cell.save(OUT/('fighter-'+str(side)+'.png'))
        fighters.append(dict(side=side,width=cell.width,height=cell.height,src=encoded(cell)))
    base=ROOT/'docs/images/readme-v162-poster.png';background=Image.open(base).convert('RGB')
    assert background.size==(1280,720)
    data=dict(background=encoded(background),ships=ships,fighters=fighters)
    template=ROOT/'tools/scripts/poster_closeup.html'
    page=template.read_text().replace('/*__PROPOSAL_DATA__*/',json.dumps(data,ensure_ascii=False))
    assert '/*__PROPOSAL_DATA__*/' not in page
    (OUT/'proposal.html').write_text(page)
    (OUT/'SOURCE.json').write_text(json.dumps(dict(prototype=True,classification='含衍生圖像，只限本機提案檢視',
        background=str(base.relative_to(ROOT)),background_sha256=sha(base),ship_assets=records,
        fighters_source=str(file.relative_to(ROOT)),fighters_sha256=sha(file),template_sha256=sha(template),
        html_sha256=sha(OUT/'proposal.html'),production_sha256=sha(ROOT/'cmd/gin3/poster_decoration.go')),
        ensure_ascii=False,indent=2)+'\n')
    print('READY 自包含HTML', (OUT/'proposal.html').stat().st_size,'bytes',flush=True)
def refresh():
    assert os.getuid()==os.getgid()==1000 and OUT.is_dir()
    source=json.loads((OUT/'SOURCE.json').read_text());assert sha(OUT/'proposal.html')==source['html_sha256']
    old=(OUT/'proposal.html').read_text()
    data=json.loads(old.split('const DATA=',1)[1].split(';\nconst $=',1)[0])
    template=ROOT/'tools/scripts/poster_closeup.html'
    (OUT/'proposal.html').write_text(template.read_text().replace('/*__PROPOSAL_DATA__*/',json.dumps(data,ensure_ascii=False)))
    source.update(template_sha256=sha(template),html_sha256=sha(OUT/'proposal.html'))
    (OUT/'SOURCE.json').write_text(json.dumps(source,ensure_ascii=False,indent=2)+'\n')
    print('READY 更新HTML模板，原圖雜湊保持',flush=True)
if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--refresh',action='store_true');a=p.parse_args()
    refresh() if a.refresh else main()
