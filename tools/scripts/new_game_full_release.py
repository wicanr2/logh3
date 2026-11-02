"""172：修正新遊戲第一場戰鬥崩潰。沿v1.6.4素材，三平台同源本機完整版。"""
import argparse
import datetime
import os
from pathlib import Path
import shutil
import subprocess
import zipfile

import scout_full_release as prior

app=prior.app;package=prior.package
ROOT=prior.ROOT;VERSION='v.1.6.5-20261102';OUT=ROOT/'dist-all'/VERSION
STAGE=ROOT/'workplace'/('package92-'+VERSION+'-release172')
BATCH=ROOT/'workplace/release172';MAC=ROOT/'workplace/macos-release172'
BASE=ROOT/'workplace/release171/extracted/squashfs-root/usr/bin'
PARENT=ROOT/'dist-all/v.1.6.4-20261031/full/LOGH3-v.1.6.4-20261031-linux-x86_64.AppImage'
PARENT_SHA='4b7b3ab7b249effcb22a7c2ce1aacaa97e5ee17be6dc66b1c35aa5f23a5b8884'
PARENT_INPUTS=ROOT/'workplace/package92-v.1.6.4-20261031-release171/INPUTS.json'
RUNTIME_SHA='1cc49bcf1e2ccd593c379adb17c9f85a36d619088296504de95b1d06215aebbf'
sha,write,read=app.sha,app.write,app.read
prior_notes=prior.notes
FIXED=('本版修正：\n'
    '沒存過檔的新遊戲開第一場戰鬥時會閃退，這個問題已修正。\n'
    '新編艦隊進戰場時會配置到入口或待命行星，不再疊在地圖角落。\n'
    '電腦解散艦隊後在同一格重編，不再造成存檔失敗或開戰前卡住。\n\n')


def notes(platform,full=True):
    text=prior_notes(platform,full).decode('utf-8-sig').replace('\r\n','\n')
    marker='本版新增：P索敵';assert text.count(marker)==1
    text=text.replace(marker,FIXED+'v.1.6.4新增：P索敵')
    if platform=='windows':
        old='先前Windows交火閃退尚未取得對應日誌，本版不宣稱已修復該未知問題。\n';assert text.count(old)==1
        text=text.replace(old,'先前回報的Windows進戰場閃退就是上面第一項。此修正在Wine驗證，未上Windows實機。\n')
    return b'\xef\xbb\xbf'+(text.replace('\n','\r\n')if platform=='windows'else text).encode()


def configure():
    prior.VERSION,prior.OUT,prior.STAGE,prior.BATCH,prior.MAC,prior.BASE,prior.PARENT=VERSION,OUT,STAGE,BATCH,MAC,BASE,PARENT
    prior.notes=notes;prior.configure()


def prepare():
    assert os.getuid()==os.getgid()==1000
    assert not OUT.exists()and not STAGE.exists()and not BATCH.exists()
    for p in [ROOT/'workplace',ROOT/'dist-all',BASE]:assert p.is_dir()and p.stat().st_uid==1000
    assert shutil.disk_usage(ROOT).free>8*1024**3
    # 產品來源必須等於已提交的HEAD；工作樹只允許發行工具與文件尚未提交。
    product=subprocess.check_output(['git','status','--porcelain','--','cmd','internal','go.mod','go.sum','LICENSE',
        'docs/internals/data/fm-import-lock.json','assets/hd/poster/approved'],text=True)
    assert product=='',product
    package.verify(BASE);assert sha(PARENT)==PARENT_SHA
    BATCH.mkdir();STAGE.mkdir();source=STAGE/'source';source.mkdir()
    for n in ['cmd','internal']:shutil.copytree(ROOT/n,source/n)
    for n in ['go.mod','go.sum','LICENSE']:shutil.copyfile(ROOT/n,source/n)
    lock=source/'docs/internals/data/fm-import-lock.json';lock.parent.mkdir(parents=True)
    shutil.copyfile(ROOT/'docs/internals/data/fm-import-lock.json',lock)
    shutil.copytree(ROOT/'assets/hd/poster/approved',source/'assets/hd/poster/approved')
    subprocess.run(['go','mod','vendor'],cwd=source,env=dict(os.environ,GOWORK='off',GOPROXY='off',GOSUMDB='off'),check=True,timeout=180)
    common=STAGE/'common';common.mkdir()
    # 素材沒有變動：人物圖已是v1.6.4預縮後的發行版，整包沿用，不再取樣一次。
    for n in ['original','data','music','soundfont','art']:shutil.copytree(BASE/n,common/n)
    shutil.copytree(BASE/'licenses',common/'licenses',ignore=shutil.ignore_patterns('linux-appimage'))
    shutil.copyfile(ROOT/'LICENSE',common/'LICENSE')
    shutil.copyfile(ROOT/'internal/localize/ui.tsv',common/'data/ui-locales.tsv')
    assert package.inventory(common)==read(PARENT_INPUTS)['common_files'],'共用素材與v1.6.4不同'
    runtime=BATCH/'runtime-x86_64'
    with PARENT.open('rb')as f:runtime.write_bytes(f.read(944632))
    assert sha(runtime)==RUNTIME_SHA;runtime.chmod(0o755)
    shutil.copytree(BASE/'licenses/linux-appimage/runtime',BATCH/'runtime-licenses')
    (BATCH/'SOURCE_DATE_EPOCH').write_text(str(int(datetime.datetime(2026,11,2,tzinfo=datetime.timezone.utc).timestamp()))+'\n')
    inputs=dict(version=VERSION,source_files=package.inventory(source),common_files=package.inventory(common),
        base=str(BASE.relative_to(ROOT)),base_manifest_sha256=sha(BASE/'MANIFEST.json'),parent_appimage_sha256=sha(PARENT),
        parent_inputs_sha256=sha(PARENT_INPUTS),runtime_sha256=sha(runtime),
        head=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
        working_tree=subprocess.check_output(['git','status','--porcelain'],text=True),
        generators={n:sha(ROOT/n)for n in ['tools/scripts/new_game_full_release.py','tools/scripts/scout_full_release.py',
            'tools/scripts/poster_frame_release.py','tools/scripts/full_appimage_release.py','tools/scripts/poster_release.py',
            'tools/scripts/v140_platform_delivery.py','tools/scripts/local_test_macos_build.sh','tools/macos.sh','tools/appimage.sh',
            'tools/scripts/appimage_build.sh','tools/Dockerfile.release']})
    write(STAGE/'INPUTS.json',inputs);(OUT/'full').mkdir(parents=True)
    print('READY',len(inputs['source_files']),'frozen sources',len(inputs['common_files']),'shared files, head',inputs['head'],flush=True)


def extract_windows():
    parent=ROOT/'workplace/music69-windows/full172';assert not parent.exists();parent.mkdir()
    archive=OUT/'full'/('LOGH3-'+VERSION+'-windows-amd64.zip')
    with zipfile.ZipFile(archive)as z:
        assert z.testzip()is None
        for item in z.infolist():
            target=parent/item.filename;assert target.resolve().is_relative_to(parent.resolve())
            z.extract(item,parent);target.chmod((item.external_attr>>16)&0o777)
    package.verify(parent/archive.stem);print('PASS actual Windows ZIP',flush=True)


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('mode',choices=['prepare','linux','windows','macos','appimage','extract-appimage','extract-windows']);a=p.parse_args();configure()
    if a.mode=='prepare':prepare()
    elif a.mode=='windows':prior.windows()
    elif a.mode=='macos':app.previous.macos()
    elif a.mode=='extract-windows':extract_windows()
    elif a.mode=='extract-appimage':app.extract_appimage()
    else:getattr(app,a.mode)()
