"""容器內保存image_gen艦型原圖；/src及/input唯讀，/output是assets/hd。"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import sys
from PIL import Image

plan=json.loads(Path(sys.argv[1]).read_text())
root=Path('/output');assert root.stat().st_uid==os.getuid()!=0
out=root/'animation';out.mkdir(exist_ok=True);assert out.stat().st_uid==os.getuid()
manifest=out/'MANIFEST.json'
m=json.loads(manifest.read_text()) if manifest.exists() else dict(schema=1,logical_cell=[256,102],atlases=[],rights='原版艦型衍生重繪，沿本機私有完整包邊界。')
name=f'ships-{plan["index"]:02}.png';source=Path('/input')/Path(plan['generated']).name
target=out/name
assert not target.exists() and source.is_file()
with Image.open(source) as im:
 w,h=im.size
assert w>=1024 and h>=408 and w<=4096 and h<=4096
assert abs(w*204-h*512)*100<=max(w*204,h*512), '圖集比例與原參照差超過1%'
shutil.copyfile(source,target)
prompt=target.with_suffix('.prompt.txt');prompt.write_text(plan['prompt']+'\n')
cells=[]
for i,path in enumerate(plan['sources']):
 p=Path('/src')/path
 with Image.open(p) as im:
  assert im.size==(256,102) and im.mode=='P'
  pixel_sha=hashlib.sha256(bytes(im.getdata())).hexdigest()
 prefix=p.parent.name[:-2]
 cells.append(dict(key=prefix,x=(i%2)*(w//2),y=(i//2)*(h//2),w=w//2,h=h//2,source=path,source_sha256=hashlib.sha256(p.read_bytes()).hexdigest(),pixels_sha256=pixel_sha))
m['atlases'].append(dict(file=name,width=w,height=h,sha256=hashlib.sha256(target.read_bytes()).hexdigest(),prompt=prompt.name,cells=cells))
keys=[c['key'] for a in m['atlases'] for c in a['cells']]
assert len(keys)==len(set(keys))
m['completed_models']=len(keys)
manifest.write_text(json.dumps(m,ensure_ascii=False,indent=2)+'\n')
print(f'保存艦型 {len(keys)}/45，生成PNG不變，{w}×{h}')
