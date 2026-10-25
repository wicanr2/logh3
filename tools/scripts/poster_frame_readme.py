"""166：從實際新版AppImage正常玩家收據複製README完整視窗，不重畫。"""
import argparse
import os
from pathlib import Path
import shutil

from PIL import Image
import poster_frame_release as build

ROOT,BATCH,VERSION=build.ROOT,build.BATCH,build.VERSION
sha,read,write=build.sha,build.read,build.write

def main(check=False):
    destination=ROOT/'docs/images'
    assert destination.is_dir() and destination.stat().st_uid==os.getuid()==1000
    files=[];receipts=[]
    for folder,names in [(BATCH/'player-linux',[
        ('zh-TW-map-v1-first','poster-v1'),('zh-TW-map-v1','quotes'),
        ('zh-TW-map-v2','poster-v2'),('zh-TW-map-v3','poster-v3'),
        ('zh-CN-map-v1','locale-zh-CN'),('ja-map-v1','locale-ja'),
        ('en-map-v1','locale-en'),('ko-map-v1','locale-ko'),
        ('zh-TW-editor-basic','editor-basic'),('zh-TW-editor-special','editor-special')]),
        (BATCH/'garrison-linux-r3',[('option-default-off','settings-alliance')])]:
        result=read(folder/'RESULT.json')
        assert result['passed'] and result['normal_entry'] and result['appimage'] and result['automatic_resources']
        assert not any(result[k]for k in ['state_injection','seed_injection','script_flag'])
        receipts.append(dict(path=str((folder/'RESULT.json').relative_to(ROOT)),sha256=sha(folder/'RESULT.json')))
        events={e['name']:e for e in result['events']if e.get('screenshot')or e.get('kind')=='capture'}
        for name,suffix in names:
            event=events[name];source=folder/event.get('screenshot',event.get('file'));target=destination/('readme-v163-'+suffix+'.png')
            assert sha(source)==event['sha256']
            with Image.open(source)as image:assert image.size==(1280,720)
            if check or target.exists():assert sha(target)==sha(source)
            else:
                assert not target.exists();shutil.copyfile(source,target)
            files.append(dict(file=str(target.relative_to(ROOT)),source=str(source.relative_to(ROOT)),
                sha256=sha(target),bytes=target.stat().st_size,event=name))
    value=dict(version=VERSION,source='實際正式AppImage正常主選單操作，完整1280×720原視窗，不重畫、不縮放',
        canvas=[1280,720],normal_entry=True,state_injection=False,seed_injection=False,
        appimage_sha256=sha(build.OUT/'full'/('LOGH3-'+VERSION+'-linux-x86_64.AppImage')),
        receipts=receipts,files=files,generator_sha256=sha(Path(__file__).resolve()))
    target=destination/'readme-v163.json'
    if check:assert read(target)==value
    else:assert not target.exists();write(target,value)
    print('PASS README',len(files),'actual package screenshots',flush=True)

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--check',action='store_true');a=p.parse_args();main(a.check)
