"""177：v.1.7.0四平台本機完整版。桌面三平台沿v1.6.5素材只換程式；Android由同一份共用素材打包。"""
import argparse
import datetime
import os
from pathlib import Path
import shutil
import subprocess
import zipfile

import new_game_full_release as ng

prior=ng.prior;app=ng.app;package=ng.package
ROOT=prior.ROOT;VERSION='v.1.7.0-20261119';OUT=ROOT/'dist-all'/VERSION
STAGE=ROOT/'workplace'/('package92-'+VERSION+'-release177')
BATCH=ROOT/'workplace/release177';MAC=ROOT/'workplace/macos-release177'
BASE=ROOT/'workplace/release172/extracted/squashfs-root/usr/bin'
PARENT=ROOT/'dist-all/v.1.6.5-20261102/full/LOGH3-v.1.6.5-20261102-linux-x86_64.AppImage'
PARENT_SHA='4220ced089e26fae7f42b1021a8e5029a093c4d08187e71b84f8ba95cd3520b1'
PARENT_INPUTS=ROOT/'workplace/package92-v.1.6.5-20261102-release172/INPUTS.json'
RUNTIME_SHA=ng.RUNTIME_SHA
# Android：主機以tools/android.sh產生，這裡只收進交付目錄並核對。
ANDROID=ROOT/'workplace/android-v170'
ANDROID_NAME='LOGH3-'+VERSION+'-android.apk'
sha,write,read=app.sha,app.write,app.read
NEW=('本版新增：\n'
    'Android版（64位元，Android 5.0以上）。戰略地圖放大到滿高，右側顯示星域摘要，'
    '左側欄放戰略命令、一覽表、系統指令與退回；戰場右側有方向鈕捲動視野。'
    '含原版配樂、音效與HD人物、星圖及介面。\n'
    '啟動時不再一次載入全部圖片與音效，需要時才讀取。\n\n'
    '本版修正：\n'
    '計畫畫面與月初報告的生產計畫、生產額度分配，可以點參謀更換。\n'
    '月初報告的徵稅計畫參謀可以更換，與戰略畫面相同。\n\n')


ANDROID_README=('銀河英雄傳說III SP remake '+VERSION+' Android 本機完整版\n\n'
    '含原版遊戲檔、原版配樂與音效及HD圖，只供合法持有原版者在自己的裝置使用，請勿散布。\n\n'
    '安裝：把APK傳到手機，允許「安裝不明應用程式」後開啟安裝。只支援64位元裝置，Android 5.0以上。\n'
    '第一次開啟會把遊戲資料展開到App私有目錄，需要約十幾秒。之後更新同一把金鑰簽的新版可以直接覆蓋安裝，存檔保留；'
    '解除安裝會連同存檔一起刪除。\n\n'
    '操作：輕觸＝移過去再按左鍵。左側「退回」與系統返回手勢＝右鍵。\n'
    '戰略地圖：左側欄有戰略命令、一覽表、系統指令；點星域選取，再點一次開星系情報。\n'
    '戰場：右側方向鈕捲動視野，一下三格。點空處開戰術選單（實行、存進度等）。\n'
    '防拷問答：右側出現數字鍵盤。\n\n'
    '已知限制：沒有桌面版的海報；觸控板模式、戰場加寬與跳出式視窗在後續版本。'
    '只在Android模擬器驗證，沒有實機收據。\n')


def notes(platform,full=True):
    text=ng.notes(platform,full).decode('utf-8-sig').replace('\r\n','\n')
    marker='本版修正：\n沒存過檔';assert text.count(marker)==1
    text=text.replace(marker,NEW+'v.1.6.5修正：\n沒存過檔')
    return b'\xef\xbb\xbf'+(text.replace('\n','\r\n')if platform=='windows'else text).encode()


def configure():
    prior.VERSION,prior.OUT,prior.STAGE,prior.BATCH,prior.MAC,prior.BASE,prior.PARENT=VERSION,OUT,STAGE,BATCH,MAC,BASE,PARENT
    prior.notes=notes;prior.configure()


def prepare():
    assert os.getuid()==os.getgid()==1000
    assert not OUT.exists()and not STAGE.exists()and not BATCH.exists()
    for p in [ROOT/'workplace',ROOT/'dist-all',BASE]:assert p.is_dir()and p.stat().st_uid==1000
    assert shutil.disk_usage(ROOT).free>8*1024**3
    # 產品來源必須等於已提交的HEAD。
    product=subprocess.check_output(['git','status','--porcelain','--','cmd','internal','mobile','android','go.mod','go.sum','LICENSE',
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
    # 素材沒有變動：發行資料與docs/internals/data逐檔相同（第177輪比對），整包沿用v1.6.5。
    for n in ['original','data','music','soundfont','art']:shutil.copytree(BASE/n,common/n)
    shutil.copytree(BASE/'licenses',common/'licenses',ignore=shutil.ignore_patterns('linux-appimage'))
    shutil.copyfile(ROOT/'LICENSE',common/'LICENSE')
    shutil.copyfile(ROOT/'internal/localize/ui.tsv',common/'data/ui-locales.tsv')
    assert package.inventory(common)==read(PARENT_INPUTS)['common_files'],'共用素材與v1.6.5不同'
    runtime=BATCH/'runtime-x86_64'
    with PARENT.open('rb')as f:runtime.write_bytes(f.read(944632))
    assert sha(runtime)==RUNTIME_SHA;runtime.chmod(0o755)
    shutil.copytree(BASE/'licenses/linux-appimage/runtime',BATCH/'runtime-licenses')
    (BATCH/'SOURCE_DATE_EPOCH').write_text(str(int(datetime.datetime(2026,11,19,tzinfo=datetime.timezone.utc).timestamp()))+'\n')
    inputs=dict(version=VERSION,source_files=package.inventory(source),common_files=package.inventory(common),
        base=str(BASE.relative_to(ROOT)),base_manifest_sha256=sha(BASE/'MANIFEST.json'),parent_appimage_sha256=sha(PARENT),
        parent_inputs_sha256=sha(PARENT_INPUTS),runtime_sha256=sha(runtime),
        head=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
        working_tree=subprocess.check_output(['git','status','--porcelain'],text=True),
        generators={n:sha(ROOT/n)for n in ['tools/scripts/v170_full_release.py','tools/scripts/new_game_full_release.py',
            'tools/scripts/scout_full_release.py','tools/scripts/poster_frame_release.py','tools/scripts/full_appimage_release.py',
            'tools/scripts/poster_release.py','tools/scripts/v140_platform_delivery.py','tools/scripts/local_test_macos_build.sh',
            'tools/macos.sh','tools/appimage.sh','tools/scripts/appimage_build.sh','tools/Dockerfile.release','tools/android.sh']})
    write(STAGE/'INPUTS.json',inputs);(OUT/'full').mkdir(parents=True)
    print('READY',len(inputs['source_files']),'frozen sources',len(inputs['common_files']),'shared files, head',inputs['head'],flush=True)


def extract_windows():
    parent=ROOT/'workplace/music69-windows/full177';assert not parent.exists();parent.mkdir()
    archive=OUT/'full'/('LOGH3-'+VERSION+'-windows-amd64.zip')
    with zipfile.ZipFile(archive)as z:
        assert z.testzip()is None
        for item in z.infolist():
            target=parent/item.filename;assert target.resolve().is_relative_to(parent.resolve())
            z.extract(item,parent);target.chmod((item.external_attr>>16)&0o777)
    package.verify(parent/archive.stem);print('PASS actual Windows ZIP',flush=True)


def android():
    """收APK：素材必須取自本版共用素材（HD圖不含海報、配樂只取根層OGG），簽章為本機發行金鑰。"""
    inputs=read(STAGE/'INPUTS.json');common=STAGE/'common'
    full=ANDROID/'stage/full';apk=ANDROID/'apk/app-build/outputs/apk/release/app-release.apk'
    assert apk.is_file()
    expect={}
    for n,row in inputs['common_files'].items():
        top,_,rest=n.partition('/')
        if top in('data','art')and not n.startswith('art/poster/'):expect[n]=row['sha256']
        elif top=='music'and '/'not in rest and rest.endswith('.ogg'):expect[n]=row['sha256']
    got={str(p.relative_to(full)):sha(p)for p in full.rglob('*')if p.is_file()and not p.relative_to(full).parts[0]=='original'}
    assert got==expect,('Android素材與共用素材不同',sorted(set(got)^set(expect))[:10])
    for p in (full/'original').iterdir():assert sha(p)==sha(ROOT/'org_game/extracted'/p.name),p.name
    signer=(ANDROID/'APKSIGNER.txt').read_text()
    assert 'Verified using v2 scheme (APK Signature Scheme v2): true'in signer and 'CN=LOGH3 remake (local)'in signer,signer
    target=OUT/'full'/ANDROID_NAME;assert not target.exists();shutil.copyfile(apk,target)
    (OUT/'full/README-Android.txt').write_bytes(ANDROID_README.encode())
    write(OUT/'full/ANDROID.json',dict(version=VERSION,package=dict(path=str(target.relative_to(OUT)),bytes=target.stat().st_size,sha256=sha(target)),
        assets=dict(files=len(got)+len(list((full/'original').iterdir())),from_common=len(got)),apksigner=signer,
        abis=['arm64-v8a','x86_64'],min_sdk=21,signing='本機發行金鑰workplace/android-state/release.keystore，不進Git',
        rights='含原版及衍生素材，只限本機完整版'))
    print('PASS Android',sha(target),flush=True)


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('mode',choices=['prepare','linux','windows','macos','appimage','extract-appimage','extract-windows','android'])
    a=p.parse_args();configure()
    if a.mode=='prepare':prepare()
    elif a.mode=='windows':prior.windows()
    elif a.mode=='macos':app.previous.macos()
    elif a.mode=='extract-windows':extract_windows()
    elif a.mode=='extract-appimage':app.extract_appimage()
    elif a.mode=='android':android()
    else:getattr(app,a.mode)()
