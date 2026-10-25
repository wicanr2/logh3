"""153：同源三平台完整版，Linux為完整AppImage。Docker內執行。"""
import argparse
import datetime
import json
import os
from pathlib import Path
import re
import shutil
import subprocess

import poster_release as previous
import v140_platform_delivery as package

ROOT=Path('/src'); VERSION='v.1.6.1-20261025'; OUT=ROOT/'dist-all'/VERSION
STAGE=ROOT/'workplace'/('package92-'+VERSION+'-release153')
MAC=ROOT/'workplace/macos-release153'; BATCH=ROOT/'workplace/appimage153'
BASE=ROOT/'dist-all/v.1.6.0-20261024/full/LOGH3-v.1.6.0-20261024-windows-amd64'
sha,write=package.sha,package.write
read=lambda p:json.loads(p.read_text())
old_notes=previous.notes

def notes(platform,full=True):
    text=old_notes(platform,full).decode('utf-8-sig').replace('\r\n','\n')
    text=text.replace('原版AI／強化AI啟動器均保留，強化預設難度1。',
        '游標獨立繪製，減少HD重繪延遲；按鈕與下拉列有滑鼠停留反光。\n'
        '滑鼠移至頂端，工具列可直接選思源黑體等既有字型及AI原版／強化0至3。\n'
        '字型與AI與系統設定共用當前值，本次執行有效；原版顯示固定原字形。\n'
        '原版AI／強化AI啟動器均保留，強化預設難度1。')
    if platform=='linux':
        text+='\nLinux主要交付為x86_64 AppImage。將.AppImage標為可執行後啟動，已包含完整素材。\n'
        text+='缺FUSE時可執行 ./LOGH3-'+VERSION+'-linux-x86_64.AppImage --appimage-extract-and-run\n'
        text+='可傳-ai -1或-ai 1選初始模式；遊戲內工具列仍可直接改AI。\n'
        text+='存檔仍在既有XDG設定目錄，不寫唯讀AppImage。附帶ALSA與X11函式庫，OpenGL與音效服務沿主機。\n'
    if platform=='windows': text=text.replace('\n','\r\n')
    return b'\xef\xbb\xbf'+text.encode('utf-8')

def configure():
    previous.VERSION,previous.OUT,previous.STAGE,previous.MAC,previous.BASE=VERSION,OUT,STAGE,MAC,BASE
    previous.notes=notes
    package.VERSION,package.OUT,package.STAGE,package.MACOS,package.readme=VERSION,OUT,STAGE,MAC,notes

def prepare():
    assert os.getuid()==os.getgid()==1000
    assert not STAGE.exists() and not (OUT/'full').exists() and not BATCH.exists()
    for p in [ROOT/'workplace',OUT,ROOT/'workplace/music69-windows']:
        assert p.is_dir() and (p.stat().st_uid,p.stat().st_gid)==(1000,1000)
    package.verify(BASE)
    old=read(OUT/'BUILD.json'); frozen=ROOT/old['source']
    assert old['version']==VERSION
    for n,h in old['source_files'].items():
        assert sha(frozen/n)==h,n
        if not n.startswith('vendor/'): assert sha(ROOT/n)==h,n
    STAGE.mkdir(); shutil.copytree(frozen,STAGE/'source')
    common=STAGE/'common'; common.mkdir()
    for n in ['original','data','art','music','soundfont','licenses']: shutil.copytree(BASE/n,common/n)
    shutil.copyfile(BASE/'LICENSE',common/'LICENSE')
    inputs=dict(version=VERSION,source_files=package.inventory(STAGE/'source'),common_files=package.inventory(common),
        base=str(BASE.relative_to(ROOT)),base_manifest_sha256=sha(BASE/'MANIFEST.json'),
        pointer_build_sha256=sha(OUT/'BUILD.json'),head=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
        working_tree=subprocess.check_output(['git','status','--porcelain'],text=True),
        generators={str(p.relative_to(ROOT)):sha(p) for p in [Path(__file__).resolve(),ROOT/'tools/scripts/poster_release.py',
            ROOT/'tools/scripts/v140_platform_delivery.py',ROOT/'tools/scripts/local_test_macos_build.sh',ROOT/'tools/appimage.sh']})
    assert set(inputs['source_files'])==set(old['source_files'])
    assert sha(common/'data/ui-locales.tsv')==sha(ROOT/'internal/localize/ui.tsv')
    for n,row in inputs['common_files'].items(): assert row['sha256']==read(BASE/'MANIFEST.json')['files'][n]['sha256']
    write(STAGE/'INPUTS.json',inputs)
    (OUT/'full').mkdir(); BATCH.mkdir()
    (BATCH/'SOURCE_DATE_EPOCH').write_text(str(int(datetime.datetime(2026,10,25,tzinfo=datetime.timezone.utc).timestamp()))+'\n')
    print('READY',len(inputs['source_files']),'sources',len(inputs['common_files']),'common files',flush=True)

def linux():
    previous.source_check()
    target=STAGE/'gin3'; assert not target.exists()
    command=['go','build','-mod=vendor','-trimpath','-buildvcs=false','-ldflags=-X main.buildVersion='+VERSION,'-o',str(target),'./cmd/gin3']
    subprocess.run(command,cwd=STAGE/'source',env=dict(os.environ,GOPROXY='off',GOSUMDB='off',GOWORK='off',GOOS='linux',GOARCH='amd64',CGO_ENABLED='1'),check=True,timeout=600)
    folder=package.unix_folder('linux-amd64',True)
    write(folder/'MANIFEST.json',dict(version=VERSION,files=package.inventory(folder)))
    appdir=BATCH/'AppDir'; (appdir/'usr').mkdir(parents=True)
    shutil.move(str(folder),appdir/'usr/bin')
    lib=appdir/'usr/lib'; lib.mkdir()
    seeds=[target]+[Path('/lib/x86_64-linux-gnu')/n for n in ['libXcursor.so.1','libXrandr.so.2','libXinerama.so.1','libXi.so.6','libXxf86vm.so.1','libXrender.so.1','libXext.so.6']]
    exclude={'libc.so.6','libm.so.6','libdl.so.2','libpthread.so.0','librt.so.1','ld-linux-x86-64.so.2'}
    deps={};queue=list(seeds)
    while queue:
        f=queue.pop(0); assert f.exists(),str(f)
        if f!=target and f.name not in deps: deps[f.name]=f
        output=subprocess.check_output(['ldd',str(f)],text=True)
        assert 'not found' not in output,output
        for name,path in re.findall(r'\s*(\S+) => (/\S+)',output):
            if name not in exclude and name not in deps:
                deps[name]=Path(path); queue.append(Path(path))
    licenses=appdir/'usr/bin/licenses/linux-appimage'; licenses.mkdir()
    packages={}; requirements={}
    for n,f in sorted(deps.items()):
        shutil.copyfile(f,lib/n)
        ownership=None
        for candidate in dict.fromkeys([str(f),str(f.resolve()),str(f.resolve()).replace('/usr/lib/','/lib/')]):
            result=subprocess.run(['dpkg-query','-S',candidate],text=True,stdout=subprocess.PIPE,stderr=subprocess.DEVNULL)
            if result.returncode==0:
                ownership=result.stdout.splitlines()[0].split(': ',1)[0];break
        assert ownership is not None,str(f)
        package_name=ownership.split(':')[0]
        packages[package_name]=subprocess.check_output(['dpkg-query','-W','-f=${Version}',package_name],text=True)
        copyright=Path('/usr/share/doc')/package_name/'copyright'
        assert copyright.is_file(),str(copyright)
        shutil.copyfile(copyright,licenses/(package_name+'.copyright'))
        versions=re.findall(r'Name: GLIBC_([0-9.]+)',subprocess.check_output(['readelf','--version-info',str(f)],text=True))
        requirements[n]=max(versions,key=lambda v:tuple(map(int,v.split('.')))) if versions else None
    requirements['gin3']=max(re.findall(r'Name: GLIBC_([0-9.]+)',subprocess.check_output(['readelf','--version-info',str(target)],text=True)),key=lambda v:tuple(map(int,v.split('.'))))
    shutil.copytree('/usr/share/alsa',appdir/'usr/share/alsa')
    shutil.copyfile('/usr/share/doc/libasound2-data/copyright',licenses/'libasound2-data.copyright')
    shutil.copytree(BATCH/'runtime-licenses',licenses/'runtime')
    (appdir/'AppRun').write_text('#!/bin/sh\nset -eu\n'
        'LOGH3_APPDIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)\n'
        'export LD_LIBRARY_PATH="$LOGH3_APPDIR/usr/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"\n'
        'export ALSA_CONFIG_PATH="$LOGH3_APPDIR/usr/share/alsa/alsa.conf"\n'
        'export ALSA_CONFIG_DIR="$LOGH3_APPDIR/usr/share/alsa"\n'
        'cd "$LOGH3_APPDIR/usr/bin"\nexec ./gin3 -scale 1.5 -ai -1 "$@"\n')
    (appdir/'AppRun').chmod(0o755)
    (appdir/'logh3.desktop').write_text('[Desktop Entry]\nType=Application\nName=LOGH3 Remake\nName[zh_TW]=銀河英雄傳說III SP\nExec=AppRun\nIcon=logh3\nTerminal=false\nCategories=Game;StrategyGame;\n')
    (appdir/'logh3.svg').write_text('<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256"><rect width="256" height="256" rx="32" fill="#142a46"/><circle cx="128" cy="128" r="92" fill="none" stroke="#b99450" stroke-width="8"/><path d="M128 38L148 108L218 128L148 148L128 218L108 148L38 128L108 108Z" fill="#ffecc7"/><circle cx="128" cy="128" r="23" fill="#b99450"/></svg>\n')
    (appdir/'.DirIcon').symlink_to('logh3.svg')
    write(BATCH/'DEPENDENCIES.json',dict(libraries=package.inventory(lib),packages=packages,glibc_requirements=requirements,
        host_requirements='x86_64、主機glibc、X11或XWayland、OpenGL/GPU驅動、音效服務；不附主機驅動'))
    write(STAGE/'LINUX.json',dict(version=VERSION,binary_sha256=sha(target),command=command,
        appdir=str(appdir.relative_to(ROOT)),go_version=subprocess.check_output(['go','version'],text=True).strip()))
    print('READY AppDir',len(deps),'runtime libraries',requirements,flush=True)

def appimage():
    file=BATCH/('LOGH3-'+VERSION+'-linux-x86_64.AppImage')
    assert file.is_file() and os.access(file,os.X_OK)
    assert subprocess.check_output([str(file),'--appimage-offset'],text=True).strip()=='944632'
    assert file.read_bytes()[8:11]==b'AI\x02'
    import hashlib
    with file.open('rb') as f: assert hashlib.sha256(f.read(944632)).hexdigest()=='1cc49bcf1e2ccd593c379adb17c9f85a36d619088296504de95b1d06215aebbf'
    target=OUT/'full'/file.name; assert not target.exists(); shutil.copyfile(file,target); target.chmod(0o755)
    write(BATCH/'APPDIR-MANIFEST.json',dict(version=VERSION,files=package.inventory(BATCH/'AppDir')))
    write(OUT/'full/APPIMAGE.json',dict(version=VERSION,package=dict(path=str(target.relative_to(OUT)),bytes=target.stat().st_size,sha256=sha(target)),
        runtime_sha256='1cc49bcf1e2ccd593c379adb17c9f85a36d619088296504de95b1d06215aebbf',offset=944632,
        appdir_manifest_sha256=sha(BATCH/'APPDIR-MANIFEST.json'),dependencies=read(BATCH/'DEPENDENCIES.json')))
    (OUT/'full/README-Linux.txt').write_bytes(notes('linux'))
    print('PASS AppImage runtime／版本輸入／SHA-256',sha(target),flush=True)

def extract_windows():
    import zipfile
    parent=ROOT/'workplace/music69-windows/full153'; assert not parent.exists(); parent.mkdir()
    archive=OUT/'full'/('LOGH3-'+VERSION+'-windows-amd64.zip')
    shutil.copyfile(archive,parent/archive.name)
    with zipfile.ZipFile(archive) as z:
        assert z.testzip() is None
        for item in z.infolist():
            assert (parent/item.filename).resolve().is_relative_to(parent.resolve())
            z.extract(item,parent)
    package.verify(parent/archive.stem)
    print('PASS 實際Windows ZIP解出',flush=True)

def extract_appimage():
    file=OUT/'full'/('LOGH3-'+VERSION+'-linux-x86_64.AppImage')
    target=BATCH/'extracted'; assert not target.exists(); target.mkdir()
    with (BATCH/'EXTRACT.log').open('x') as log:
        subprocess.run([str(file),'--appimage-extract'],cwd=target,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=180)
    folder=target/'squashfs-root'
    expected=read(BATCH/'APPDIR-MANIFEST.json')['files']
    for n,row in expected.items():
        f=folder/n;assert f.is_file() and sha(f)==row['sha256'],n
    assert package.inventory(folder)==expected
    print('PASS 實際AppImage全部檔案解出與權限',len(expected),flush=True)

if __name__=='__main__':
    assert Path.cwd()==ROOT and os.getuid()==os.getgid()==1000
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('mode',choices=['prepare','linux','windows','macos','appimage','extract-windows','extract-appimage']);a=p.parse_args();configure()
    if a.mode=='windows': previous.build('windows')
    elif a.mode=='macos': previous.macos()
    elif a.mode=='extract-windows': extract_windows()
    elif a.mode=='extract-appimage': extract_appimage()
    else: globals()[a.mode]()
