"""從實際封包的公開讀檔／存檔入口驗三劇本及個人索引回退。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import struct
import subprocess

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--package',required=True)
p.add_argument('--out',required=True)
p.add_argument('--windows',action='store_true')
a=p.parse_args()
package=Path(a.package).resolve()
out=Path(a.out).resolve()
assert os.getuid()==1000 and package.is_dir() and not out.exists()
out.mkdir()
original=package/'original'
index=(original/'SAVEDATA.IDX').read_bytes()
binary=package/('gin3.exe' if a.windows else 'gin3')
def path(p):
    return 'Z:'+str(p).replace('/','\\') if a.windows else str(p)
results=[]
for row,turn in [(0,346),(8,93),(9,346),(10,380)]:
    folder=out/f'slot-{row}'
    folder.mkdir()
    save=folder/'save'
    save.mkdir()
    personal=bytearray(index)
    personal[8*48:]=bytes(len(personal)-8*48)
    personal[:48]=index[9*48:10*48]
    (save/'SAVEDATA.IDX').write_bytes(personal)
    for stem in ['WORK','STELLAR','PLANET','ADMIRAL','SHIP','FLEET']:
        shutil.copyfile(original/(stem+'.DT9'),save/(stem+'.DT0'))
    script=f'click 300 270;type 138;enter;click 300 {122+row*20};click 318 208;click 544 125;click 560 165;click 300 142'
    # 同一遊戲保存後再選S3，保存至另一格，不能靠重新啟動補回清單。
    script+=';rclick 300 240;click 544 125;click 560 145;click 300 322;click 318 208;click 544 125;click 560 165;click 300 162'
    cmd=(['wine'] if a.windows else [])+[str(binary),'-orig',path(original),'-data',path(package/'data'),
        '-save',path(save),'-script',script,'-shot',path(folder/'screen.png'),'-state-out',path(folder/'state.txt')]
    with (folder/'run.log').open('x') as log:
        subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=75)
    got=struct.unpack_from('<H',(save/'WORK.DT0').read_bytes(),4)[0]
    after=struct.unpack_from('<H',(save/'WORK.DT1').read_bytes(),4)[0]
    assert got==turn and after==380,(row,got,after)
    assert 'state=StateSaveGame' in (folder/'state.txt').read_text()
    results.append(dict(row=row,loaded_turn=got,reloaded_s3_turn=after,command=cmd,
                        screenshot_sha256=hashlib.sha256((folder/'screen.png').read_bytes()).hexdigest()))
    print('PASS 個人索引／額外劇本／保存後再載S3',row,got,flush=True)
(out/'SCENARIOS.json').write_text(json.dumps(dict(platform='Windows Wine' if a.windows else 'Linux',
    binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),source_index_sha256=hashlib.sha256(index).hexdigest(),
    classification='實際封包公開CLI存讀抽驗；正常X11 S1路徑與實錄另存',runs=results),ensure_ascii=False,indent=2)+'\n')
