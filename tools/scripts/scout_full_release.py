"""171：P索敵、整隊行軍及按需HD，三平台同源本機完整版。"""
import argparse
import datetime
import os
from pathlib import Path
import shutil
import subprocess
import zipfile

import poster_frame_release as old
import prepare_hd_release

app=old.app;package=app.package
ROOT=Path('/src');VERSION='v.1.6.4-20261031';OUT=ROOT/'dist-all'/VERSION
STAGE=ROOT/'workplace'/('package92-'+VERSION+'-release171')
BATCH=ROOT/'workplace/release171';MAC=ROOT/'workplace/macos-release171'
BASE=ROOT/'workplace/poster166/extracted/squashfs-root/usr/bin'
PARENT=ROOT/'dist-all/v.1.6.3-20261026/full/LOGH3-v.1.6.3-20261026-linux-x86_64.AppImage'
sha,write,read=app.sha,app.write,app.read


def notes(platform,full=True):
    text=old.notes(platform,full).decode('utf-8-sig').replace('\r\n','\n')
    text=text.replace('v.1.6.3-20261026',VERSION)
    text+='\n本版新增：P索敵可見圖示與視野圈，指定位置或取消後回原視窗。\n'
    text+='敵我AI改為整隊下令，保留編隊與原有避障。\n'
    text+='改善啟動速度與記憶體用量，圖片需要時載入。發行人物圖預縮，顯示像素保持。\n'
    if platform=='windows':
        text+='先前Windows交火閃退尚未取得對應日誌，本版不宣稱已修復該未知問題。\n'
        text+='若仍退出，可用capture-battle-original.cmd或capture-battle-enhanced.cmd啟動，日誌在%TEMP%\\logh3-battle-*.log。\n'
    return b'\xef\xbb\xbf'+(text.replace('\n','\r\n')if platform=='windows'else text).encode()


def configure():
    app.VERSION,app.OUT,app.STAGE,app.BATCH,app.MAC,app.BASE=VERSION,OUT,STAGE,BATCH,MAC,BASE
    app.configure();app.notes=notes;app.previous.notes=notes;package.readme=notes


def prepare():
    assert os.getuid()==os.getgid()==1000
    assert not OUT.exists()and not STAGE.exists()and not BATCH.exists()
    for p in [ROOT/'workplace',ROOT/'dist-all',BASE]:assert p.is_dir()and p.stat().st_uid==1000
    assert shutil.disk_usage(ROOT).free>8*1024**3
    package.verify(BASE)
    assert sha(PARENT)=='96f519353fe1e94c405fd56dfd0a91a0da48f481e2abd5b59924bf917b5f1e1d'
    BATCH.mkdir();STAGE.mkdir();source=STAGE/'source';source.mkdir()
    for n in ['cmd','internal']:shutil.copytree(ROOT/n,source/n)
    for n in ['go.mod','go.sum','LICENSE']:shutil.copyfile(ROOT/n,source/n)
    lock=source/'docs/internals/data/fm-import-lock.json';lock.parent.mkdir(parents=True)
    shutil.copyfile(ROOT/'docs/internals/data/fm-import-lock.json',lock)
    shutil.copytree(ROOT/'assets/hd/poster/approved',source/'assets/hd/poster/approved')
    subprocess.run(['go','mod','vendor'],cwd=source,env=dict(os.environ,GOWORK='off',GOPROXY='off',GOSUMDB='off'),check=True,timeout=180)
    common=STAGE/'common';common.mkdir()
    for n in ['original','data','music','soundfont']:shutil.copytree(BASE/n,common/n)
    # Linux專用runtime授權在AppDir依實際相依重建，不放進三平台共用輸入。
    shutil.copytree(BASE/'licenses',common/'licenses',ignore=shutil.ignore_patterns('linux-appimage'))
    shutil.copyfile(ROOT/'LICENSE',common/'LICENSE')
    prepare_hd_release.prepare(BASE/'art',BASE/'original',common/'art')
    shutil.copyfile(ROOT/'internal/localize/ui.tsv',common/'data/ui-locales.tsv')
    runtime=BATCH/'runtime-x86_64'
    with PARENT.open('rb')as f:runtime.write_bytes(f.read(944632))
    assert sha(runtime)=='1cc49bcf1e2ccd593c379adb17c9f85a36d619088296504de95b1d06215aebbf';runtime.chmod(0o755)
    shutil.copytree(BASE/'licenses/linux-appimage/runtime',BATCH/'runtime-licenses')
    (BATCH/'SOURCE_DATE_EPOCH').write_text(str(int(datetime.datetime(2026,10,31,tzinfo=datetime.timezone.utc).timestamp()))+'\n')
    inputs=dict(version=VERSION,source_files=package.inventory(source),common_files=package.inventory(common),
        base=str(BASE.relative_to(ROOT)),base_manifest_sha256=sha(BASE/'MANIFEST.json'),parent_appimage_sha256=sha(PARENT),
        runtime_sha256=sha(runtime),head=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
        working_tree=subprocess.check_output(['git','status','--porcelain'],text=True),
        generators={n:sha(ROOT/n)for n in ['tools/scripts/scout_full_release.py','tools/scripts/prepare_hd_release.py',
            'tools/scripts/full_appimage_release.py','tools/scripts/poster_release.py','tools/scripts/v140_platform_delivery.py',
            'tools/scripts/local_test_macos_build.sh','tools/macos.sh','tools/appimage.sh','tools/scripts/appimage_build.sh','tools/Dockerfile.release']})
    write(STAGE/'INPUTS.json',inputs);(OUT/'full').mkdir(parents=True)
    print('READY',len(inputs['source_files']),'frozen sources',len(inputs['common_files']),'shared files',flush=True)


def windows():
    app.previous.source_check();target=STAGE/'gin3.exe';assert not target.exists()
    command=['go','build','-mod=vendor','-trimpath','-buildvcs=false','-ldflags=-X main.buildVersion='+VERSION,'-o',str(target),'./cmd/gin3']
    subprocess.run(command,cwd=STAGE/'source',env=dict(os.environ,GOOS='windows',GOARCH='amd64',CGO_ENABLED='0',GOWORK='off',GOPROXY='off',GOSUMDB='off'),check=True,timeout=600)
    folder=OUT/'full'/('LOGH3-'+VERSION+'-windows-amd64');assert not folder.exists()
    shutil.copytree(STAGE/'common',folder);shutil.copyfile(target,folder/'gin3.exe')
    (folder/'README.txt').write_bytes(notes('windows'));write(folder/'VERSION.json',package.metadata('windows-amd64'))
    for mode,level in [('original',-1),('enhanced',1)]:
        quality=' -music-quality hd'if mode=='enhanced'else''
        (folder/('start-'+mode+'.cmd')).write_bytes(('@echo off\r\ncd /d "%~dp0"\r\ngin3.exe -ai '+str(level)+' -scale 1.5'+quality+' %*\r\nif errorlevel 1 pause\r\n').encode('ascii'))
        shutil.copyfile(ROOT/'tools'/('capture-battle-'+mode+'.cmd'),folder/('capture-battle-'+mode+'.cmd'))
    item=package.archive(folder,'含原版及衍生素材，只限本機完整版')
    write(STAGE/'WINDOWS.json',dict(version=VERSION,package=item,binary_sha256=sha(target),command=command,go_version=subprocess.check_output(['go','version'],text=True).strip()))


def extract_windows():
    parent=ROOT/'workplace/music69-windows/full171';assert not parent.exists();parent.mkdir()
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
    elif a.mode=='windows':windows()
    elif a.mode=='macos':app.previous.macos()
    elif a.mode=='extract-windows':extract_windows()
    elif a.mode=='extract-appimage':app.extract_appimage()
    else:getattr(app,a.mode)()
