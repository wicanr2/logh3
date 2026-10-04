"""組合既有真實22首配樂及28段音效，來源及雜湊逐檔保留。"""
from pathlib import Path
import hashlib
import json
import os
import shutil
import sys

os.chdir('/src')
out=Path('workplace')/sys.argv[1]
assert not out.exists() and out.parent.stat().st_uid==os.getuid()
out.mkdir()
files=[]
for dirname,count in [('workplace/music69-fm/full',22),('workplace/round78-sfx-final-a',28)]:
    paths=sorted(Path(dirname).glob('*.ogg'))
    assert len(paths)==count,dirname
    for p in paths:
        assert p.stat().st_uid==os.getuid() and not (out/p.name).exists()
        shutil.copyfile(p,out/p.name)
        b=p.read_bytes()
        assert (out/p.name).read_bytes()==b
        files.append({'file':p.name,'source':str(p),'bytes':len(b),'sha256':hashlib.sha256(b).hexdigest()})
result={'classification':'玩家自備原版的真實衍生音訊，只留本機，不加入公開發行包','files':files}
(out/'LIBRARY-SOURCES.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('真實配樂22首、音效28段，來源相同')
