"""166：新版銀河海報及名言，同源三平台本機完整版。Docker內執行。"""
import argparse
import datetime
import os
from pathlib import Path
import shutil
import subprocess
import zipfile

import full_appimage_release as app

ROOT=app.ROOT;VERSION='v.1.6.3-20261026';OUT=ROOT/'dist-all'/VERSION
STAGE=ROOT/'workplace'/('package92-'+VERSION+'-release166')
BATCH=ROOT/'workplace/poster166';MAC=ROOT/'workplace/macos-release166'
BASE=ROOT/'workplace/package92-v.1.6.2-20261025-release154/common'
sha,write,read=app.sha,app.write,app.read
base_notes=app.notes

def notes(platform,full=True):
    value=base_notes(platform,full).decode('utf-8-sig').replace('\r\n','\n')
    value=value.replace('底部斜體名言','人物圖內斜體名言與姓名')
    value=value.replace('譯文唯一來源為data/ui-locales.tsv，可修改後重啟。','主要介面譯表為data/ui-locales.tsv，可修改後重啟；新版海報名言為已核對PNG圖層。')
    value+='\n核准新版海報：三組人物與上下銀河銜接，上方休伯利安朝左，下方伯倫希爾朝右。\n'
    value+='船體固定，只循環短光束；關閉動畫時光束停止。中央畫面與滑鼠命中同步縮放。\n'
    value+='v1五組雙人名言每60秒同步輪替，繁／簡／日／英／韓共用組別；重啟從首組開始。\n'
    value+='名言時間獨立於遊戲回合與動畫開關；v2／v3不疊v1名言。\n'
    value+='語錄採已同意網路轉錄與本專案譯文，尚未逐字核對尖端舊版，不稱官方譯文。\n'
    value+='來源見art/poster/approved/QUOTE-SOURCES.json；缺少輪替擴充時保留原首組。\n'
    return b'\xef\xbb\xbf'+(value.replace('\n','\r\n') if platform=='windows' else value).encode()

def configure():
    app.VERSION,app.OUT,app.STAGE,app.BATCH,app.MAC=VERSION,OUT,STAGE,BATCH,MAC
    app.configure();app.notes=notes;app.previous.notes=notes;app.package.readme=notes

def prepare():
    assert os.getuid()==os.getgid()==1000 and not STAGE.exists() and not OUT.exists() and not BATCH.exists()
    for p in [ROOT/'workplace',ROOT/'dist-all',BASE]:assert p.is_dir() and p.stat().st_uid==1000
    assert shutil.disk_usage(ROOT).free>8*1024**3
    old=read(BASE.parent/'INPUTS.json')
    for n,v in old['common_files'].items():assert sha(BASE/n)==v['sha256'],n
    BATCH.mkdir();STAGE.mkdir();source=STAGE/'source';source.mkdir()
    for n in ['cmd','internal']:shutil.copytree(ROOT/n,source/n)
    for n in ['go.mod','go.sum','LICENSE']:shutil.copyfile(ROOT/n,source/n)
    lock=source/'docs/internals/data/fm-import-lock.json';lock.parent.mkdir(parents=True)
    shutil.copyfile(ROOT/'docs/internals/data/fm-import-lock.json',lock)
    shutil.copytree(ROOT/'assets/hd/poster/approved',source/'assets/hd/poster/approved')
    subprocess.run(['go','mod','vendor'],cwd=source,env=dict(os.environ,GOPROXY='off',GOSUMDB='off',GOWORK='off'),check=True,timeout=180)
    common=STAGE/'common';shutil.copytree(BASE,common)
    target=common/'art/poster/approved';assert not target.exists()
    shutil.copytree(ROOT/'assets/hd/poster/approved',target)
    shutil.copyfile(ROOT/'tools/scripts/poster_quote_candidates.json',target/'QUOTE-SOURCES.json')
    shutil.copyfile(ROOT/'internal/localize/ui.tsv',common/'data/ui-locales.tsv')
    shutil.copytree(ROOT/'workplace/appimage153/runtime-licenses',BATCH/'runtime-licenses')
    write(STAGE/'INPUTS.json',dict(version=VERSION,source_files=app.package.inventory(source),common_files=app.package.inventory(common),
        base=str(BASE.relative_to(ROOT)),base_inputs_sha256=sha(BASE.parent/'INPUTS.json'),
        head=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),working_tree=subprocess.check_output(['git','status','--porcelain'],text=True),
        generators={str(p.relative_to(ROOT)):sha(p)for p in[Path(__file__).resolve(),ROOT/'tools/scripts/full_appimage_release.py',ROOT/'tools/scripts/poster_release.py',ROOT/'tools/scripts/v140_platform_delivery.py',ROOT/'tools/scripts/local_test_macos_build.sh',ROOT/'tools/macos.sh',ROOT/'tools/appimage.sh',ROOT/'tools/scripts/appimage_build.sh']}))
    (OUT/'full').mkdir(parents=True)
    (BATCH/'SOURCE_DATE_EPOCH').write_text(str(int(datetime.datetime(2026,10,26,tzinfo=datetime.timezone.utc).timestamp()))+'\n')
    inputs=app.previous.source_check();print('READY',len(inputs['source_files']),'source',len(inputs['common_files']),'common',flush=True)

def extract_windows():
    parent=ROOT/'workplace/music69-windows/full166';assert not parent.exists();parent.mkdir()
    archive=OUT/'full'/('LOGH3-'+VERSION+'-windows-amd64.zip')
    with zipfile.ZipFile(archive)as z:
        assert z.testzip()is None
        for item in z.infolist():
            assert (parent/item.filename).resolve().is_relative_to(parent.resolve())
            z.extract(item,parent);(parent/item.filename).chmod((item.external_attr>>16)&0o777)
    app.package.verify(parent/archive.stem);print('PASS Windows ZIP actually extracted',flush=True)

if __name__=='__main__':
    assert Path.cwd()==ROOT and os.getuid()==os.getgid()==1000
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('mode',choices=['prepare','linux','windows','macos','appimage','extract-appimage','extract-windows']);a=p.parse_args();configure()
    if a.mode=='prepare':prepare()
    elif a.mode=='windows':app.previous.build('windows')
    elif a.mode=='macos':app.previous.macos()
    elif a.mode=='extract-windows':extract_windows()
    elif a.mode=='extract-appimage':app.extract_appimage()
    else:getattr(app,a.mode)()
