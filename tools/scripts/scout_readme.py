"""171：README 的正式 AppImage 索敵與行軍畫面，保留完整原視窗。"""
import argparse
import json
from pathlib import Path
import shutil
import subprocess

import scout_full_release as r

def main(check):
    folder=r.BATCH/'battle-linux';receipt=r.read(folder/'RESULT.json');index=r.ROOT/'docs/images/readme-v164.json'
    if check:
        data=r.read(index)
        for name,h in data['input_sha256'].items():assert r.sha(r.ROOT/name)==h,name
        for item in data['images']:assert r.sha(r.ROOT/item['path'])==item['sha256']
        print('PASS 2 full-canvas README images');return
    assert not index.exists()and index.parent.stat().st_uid==1000
    assert receipt['passed']and receipt['appimage']and receipt['automatic_resources']
    movie=r.read(r.BATCH/'movie/PLAN.json');march=movie['march_movement'][-1]['seconds']+.15
    images=[]
    for name,source,at in [('scout',folder/'scout-range.png',None),('march',folder/'RAW.mp4',march)]:
        target=index.parent/('readme-v164-'+name+'.png');assert not target.exists()
        if at is None:shutil.copyfile(source,target)
        else:subprocess.run(['ffmpeg','-nostdin','-v','error','-threads','2','-ss',str(at),'-i',str(source),'-frames:v','1','-threads','1',str(target)],check=True,timeout=30)
        from PIL import Image
        assert Image.open(target).size==(1280,720)
        images.append(dict(path=str(target.relative_to(r.ROOT)),sha256=r.sha(target),source=str(source.relative_to(r.ROOT)),seconds=at,size=[1280,720]))
    paths=[folder/'RESULT.json',folder/'RAW.mp4',folder/'scout-range.png',folder/'FRAMES.jsonl',Path(__file__).resolve()]
    r.write(index,dict(version=r.VERSION,normal_entry=True,state_injection=False,full_canvas=True,
        input_sha256={str(p.relative_to(r.ROOT)):r.sha(p)for p in paths},march_movement=movie['march_movement'],images=images))
    print('READY 2 full-canvas README images')

if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('--check',action='store_true');a=p.parse_args();main(a.check)
