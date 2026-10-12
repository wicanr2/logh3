"""取得石黑版官方角色圖供本機影片試剪；不授予公開散布權。"""
import hashlib
import json
import os
from pathlib import Path
import urllib.request

from PIL import Image

out=Path('/work/hero130')
assert os.getuid()==1000 and out.is_dir() and out.stat().st_uid==1000
destination=out/'anime-sources'
destination.mkdir(exist_ok=True)
assert destination.stat().st_uid==1000
inputs=[
    ('reinhard.jpg','https://www.ginei.jp/img/char/g01.jpg','石黑版官方角色圖：萊因哈特'),
    ('yang.jpg','https://www.ginei.jp/img/char/d01.jpg','石黑版官方角色圖：楊威利'),
    ('empire.html','https://www.ginei.jp/character_g.php','官方人物來源與權利聲明'),
    ('alliance.html','https://www.ginei.jp/character_d.php','官方人物來源與權利聲明'),
    ('4k-remaster.html','https://www.ginei.jp/4kremaster.php','1988與1993兩劇場版官方年代對照'),
    ('remaster-poster.png','https://www.ginei.jp/img/remaster/remasterPoster.png','1988及1993兩作品官方海報，須分別標明年代'),
]
records=[]
for name,url,description in inputs:
    path=destination/name
    if not path.exists():
        request=urllib.request.Request(url,headers={'User-Agent':'LOGH3-local-source-reference/1.0'})
        with urllib.request.urlopen(request,timeout=25) as response:
            assert response.status==200
            data=response.read(8*1024*1024+1)
            assert 0<len(data)<=8*1024*1024
        path.write_bytes(data)
    data=path.read_bytes()
    record=dict(file=name,url=url,description=description,bytes=len(data),
                sha256=hashlib.sha256(data).hexdigest(),classification='本機試剪，公開散布權未知')
    if path.suffix in ('.jpg','.png'):
        with Image.open(path) as im:
            im.load();assert im.format in ('JPEG','PNG')
            record['size']=list(im.size)
    records.append(record)
manifest=dict(source_family='1988年起石黑版，非2018新版',
    provenance='官方網站角色圖；不是本輪擷取的1988正片影格，個別角色圖出現集數未知',
    use='只供本機預覽；不加入Git、公開Release或可散布引擎包',records=records)
(destination/'SOURCES.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(manifest,ensure_ascii=False,indent=2))
