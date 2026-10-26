"""154：凍結海報交戰裝飾來源，沿153工具鏈打包三平台本機完整版。"""
import argparse
import datetime
import os
from pathlib import Path
import shutil
import subprocess
import zipfile

import full_appimage_release as app

ROOT=app.ROOT; VERSION='v.1.6.2-20261025'; OUT=ROOT/'dist-all'/VERSION
STAGE=ROOT/'workplace'/('package92-'+VERSION+'-release154')
BATCH=ROOT/'workplace/poster154'; MAC=ROOT/'workplace/macos-release154'
BASE=ROOT/'workplace/package92-v.1.6.1-20261025-release153/common'
sha,write,read=app.sha,app.write,app.read
base_notes=app.notes

def notes(platform,full=True):
    value=base_notes(platform,full).decode('utf-8-sig').replace('\r\n','\n')
    value+='\n海報模式上下留白加入輕量交戰循環：上方雙方戰機，下方16種帝國與12種同盟旗艦。\n'
    value+='每12秒換一批旗艦；關閉系統設定的動畫時顯示靜態裝飾。海報關閉時不顯示。\n'
    value+='使用既有旗艦圖與新戰機小圖；來源與提示見art/poster/fighters.SOURCE.json。\n'
    return b'\xef\xbb\xbf'+(value.replace('\n','\r\n') if platform=='windows' else value).encode()

def configure():
    app.VERSION,app.OUT,app.STAGE,app.BATCH,app.MAC=VERSION,OUT,STAGE,BATCH,MAC
    app.configure()
    app.notes=notes; app.previous.notes=notes; app.package.readme=notes

def prepare():
    assert os.getuid()==os.getgid()==1000 and not STAGE.exists() and not OUT.exists()
    for f in [ROOT/'workplace',ROOT/'dist-all',BATCH,BASE]:
        assert f.is_dir() and (f.stat().st_uid,f.stat().st_gid)==(1000,1000)
    assert shutil.disk_usage(ROOT).free>8*1024**3
    old=read(BASE.parent/'INPUTS.json')
    for n,v in old['common_files'].items(): assert sha(BASE/n)==v['sha256'],n
    STAGE.mkdir(); source=STAGE/'source';source.mkdir()
    for n in ['cmd','internal']: shutil.copytree(ROOT/n,source/n)
    for n in ['go.mod','go.sum','LICENSE']:shutil.copyfile(ROOT/n,source/n)
    lock=source/'docs/internals/data/fm-import-lock.json';lock.parent.mkdir(parents=True)
    shutil.copyfile(ROOT/'docs/internals/data/fm-import-lock.json',lock)
    subprocess.run(['go','mod','vendor'],cwd=source,env=dict(os.environ,GOPROXY='off',GOSUMDB='off',GOWORK='off'),check=True,timeout=180)
    common=STAGE/'common';shutil.copytree(BASE,common)
    for n in ['fighters.png','fighters.prompt.txt','fighters.SOURCE.json']:
        assert not (common/'art/poster'/n).exists()
        shutil.copyfile(ROOT/'assets/hd/poster'/n,common/'art/poster'/n)
    shutil.copytree(ROOT/'workplace/appimage153/runtime-licenses',BATCH/'runtime-licenses')
    write(STAGE/'INPUTS.json',dict(version=VERSION,source_files=app.package.inventory(source),common_files=app.package.inventory(common),
        base=str(BASE.relative_to(ROOT)),base_inputs_sha256=sha(BASE.parent/'INPUTS.json'),
        head=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
        working_tree=subprocess.check_output(['git','status','--porcelain'],text=True),
        generators={str(f.relative_to(ROOT)):sha(f) for f in [Path(__file__).resolve(),ROOT/'tools/scripts/full_appimage_release.py',
            ROOT/'tools/scripts/poster_release.py',ROOT/'tools/scripts/v140_platform_delivery.py',ROOT/'tools/scripts/local_test_macos_build.sh',
            ROOT/'tools/macos.sh',ROOT/'tools/appimage.sh',ROOT/'tools/scripts/appimage_build.sh']}))
    (OUT/'full').mkdir(parents=True)
    (BATCH/'SOURCE_DATE_EPOCH').write_text(str(int(datetime.datetime(2026,10,25,tzinfo=datetime.timezone.utc).timestamp()))+'\n')
    data=app.previous.source_check()
    print('READY',len(data['source_files']),'source',len(data['common_files']),'common',flush=True)

def extract_windows():
    parent=ROOT/'workplace/music69-windows/full154'; assert not parent.exists();parent.mkdir()
    archive=OUT/'full'/('LOGH3-'+VERSION+'-windows-amd64.zip')
    with zipfile.ZipFile(archive) as z:
        assert z.testzip() is None
        for item in z.infolist():
            assert (parent/item.filename).resolve().is_relative_to(parent.resolve())
            z.extract(item,parent);(parent/item.filename).chmod((item.external_attr>>16)&0o777)
    app.package.verify(parent/archive.stem)
    print('PASS Windows ZIP實際解出',flush=True)

if __name__=='__main__':
    assert Path.cwd()==ROOT and os.getuid()==os.getgid()==1000
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('mode',choices=['prepare','linux','windows','macos','appimage','extract-appimage','extract-windows'])
    a=p.parse_args();configure()
    if a.mode=='prepare':prepare()
    elif a.mode=='windows':app.previous.build('windows')
    elif a.mode=='macos':app.previous.macos()
    elif a.mode=='extract-windows':extract_windows()
    elif a.mode=='extract-appimage':app.extract_appimage()
    else:getattr(app,a.mode)()
