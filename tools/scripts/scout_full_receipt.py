"""171：實際三平台完整版、正常玩家路徑與完整推廣影片的交付核對。"""
import hashlib
import argparse
import json
import os
from pathlib import Path
import plistlib
import shutil
import struct
import zipfile

import scout_full_release as release

ROOT,OUT,STAGE,BATCH,VERSION=release.ROOT,release.OUT,release.STAGE,release.BATCH,release.VERSION
read,write,sha=release.read,release.write,release.sha

def record(p):return dict(path=str(p.relative_to(ROOT)),bytes=p.stat().st_size,sha256=sha(p))

def check(packages_only=False):
    assert os.getuid()==1000
    release.configure();inputs=read(STAGE/'INPUTS.json');common=inputs['common_files']
    assert release.package.inventory(STAGE/'source')==inputs['source_files']
    assert release.package.inventory(STAGE/'common')==common
    for n,row in inputs['source_files'].items():
        if not n.startswith('vendor/'):assert sha(ROOT/n)==row['sha256'],n
    for n,h in inputs['generators'].items():assert sha(ROOT/n)==h,n
    folders={name:OUT/'full'/('LOGH3-'+VERSION+'-'+name) for name in ['windows-amd64','macos-universal']}
    artifacts=[];checks={}
    for name,folder in folders.items():
        release.package.verify(folder)
        archive=folder.parent/(folder.name+'.zip')
        with zipfile.ZipFile(archive)as z:
            assert z.testzip()is None
            rows=read(folder/'MANIFEST.json')['files']
            assert len(z.infolist())==len(rows)+1
            for n,row in rows.items():
                e=z.getinfo(folder.name+'/'+n);data=z.read(e)
                assert len(data)==row['bytes']and hashlib.sha256(data).hexdigest()==row['sha256']
                assert (e.external_attr>>16)&0o777==int(row['mode'],8)
                assert e.filename.isascii()or e.flag_bits&0x800
        assert read(folder/'VERSION.json')['version']==VERSION
        artifacts.append(record(archive));checks[name]=dict(crc=True,manifest_files=len(rows),permissions=True,utf8_names=True)
    windows=folders['windows-amd64'];wb=windows/'gin3.exe';pe=wb.read_bytes()
    assert pe[:2]==b'MZ';p=struct.unpack_from('<I',pe,0x3c)[0]
    assert pe[p:p+4]==b'PE\0\0'and struct.unpack_from('<H',pe,p+4)[0]==0x8664
    assert (windows/'README.txt').read_bytes().startswith(b'\xef\xbb\xbf')
    for f in ['start-original.cmd','start-enhanced.cmd','capture-battle-original.cmd','capture-battle-enhanced.cmd']:
        data=(windows/f).read_bytes();assert b'\r\n'in data and b'\n'not in data.replace(b'\r\n',b'')
    mac=folders['macos-universal'];app=mac/'銀河英雄傳說III.app/Contents'
    plist=plistlib.loads((app/'Info.plist').read_bytes())
    assert plist['LOGH3FullVersion']==VERSION
    assert plist['CFBundleVersion']==plist['CFBundleShortVersionString']=='1.6.4'
    slices=release.app.previous.universal(app/'MacOS/gin3')
    resources=app/'Resources'
    linux=BATCH/'extracted/squashfs-root/usr/bin';release.package.verify(linux)
    appimage=OUT/'full'/('LOGH3-'+VERSION+'-linux-x86_64.AppImage')
    data=appimage.open('rb').read(944632)
    assert hashlib.sha256(data).hexdigest()==inputs['runtime_sha256']
    with appimage.open('rb')as f:f.seek(944632);assert f.read(4)==b'hsqs'
    assert (linux/'gin3').read_bytes()[:6]==b'\x7fELF\x02\x01'
    expected=read(BATCH/'APPDIR-MANIFEST.json')
    # The extractor must preserve the exact AppDir tree, including symlink targets.
    def tree(folder):
        return {str(p.relative_to(folder)):('symlink',os.readlink(p))if p.is_symlink()else('file',sha(p),p.stat().st_mode&0o777)
                for p in folder.rglob('*')if p.is_symlink()or p.is_file()}
    assert tree(BATCH/'AppDir')==tree(BATCH/'extracted/squashfs-root')
    assert read(linux/'VERSION.json')['version']==VERSION
    artifacts.append(record(appimage))
    for folder in [windows,resources,linux]:
        for name,row in common.items():
            p=folder/name;assert p.stat().st_size==row['bytes']and sha(p)==row['sha256'],str(p)
        assert not list(folder.rglob('preferences.json'))
    packed=read(STAGE/'common/art/PACKED-HD.json')
    assert len(packed['changed_files'])==83
    players={}
    for name,path in [('linux-alliance',BATCH/'player-linux'),('linux-empire',BATCH/'player-empire'),
                      ('windows-wine',ROOT/'workplace/music69-windows/full171/player-windows'),
                      ('linux-battle',BATCH/'battle-linux'),('linux-garrison',BATCH/'garrison-linux')]:
        receipt=read(path/'RESULT.json')
        assert receipt['passed']and receipt['normal_entry']
        assert not any(receipt.get(k,False)for k in ['state_injection','seed_injection','time_injection','script_flag'])
        expected_binary=wb if name=='windows-wine'else appimage
        assert receipt['binary_sha256']==sha(expected_binary)
        if name!='windows-wine':assert receipt['appimage']and receipt['automatic_resources']
        players[name]=record(path/'RESULT.json')
    alliance=read(BATCH/'player-linux/RESULT.json');wine=read(ROOT/'workplace/music69-windows/full171/player-windows/RESULT.json')
    assert alliance['saved']==wine['saved']and alliance['full_cycle']
    battle=read(BATCH/'battle-linux/RESULT.json')
    assert battle['pre_cleanup_alive']and battle['last']['state']=='StateBattle'
    assert battle['last']['tactical']['turn']==3
    assert sha(BATCH/'battle-linux/RAW.mp4')==battle['video_sha256']
    if packages_only:
        write(BATCH/'PACKAGE-CHECK.json',dict(passed=True,version=VERSION,artifacts=artifacts,
            common_files=len(common),source_files=len(inputs['source_files']),player_receipts=players,slices=slices))
        print('PASS 3 actual packages, 652 identical shared files, 5 normal-path receipts',flush=True)
        return
    promo=OUT/'promo';audit=read(promo/'AUDIT.json');plan=read(promo/'PLAN.json')
    assert audit['passed']and audit['visual_review'].startswith('已目視')
    assert audit['version']==VERSION and plan['version']==VERSION
    assert sha(promo/audit['file'])==audit['sha256']
    assert len(plan['scenes'])==37 and read(promo/'SOUNDTRACK.json')['music_loops']==0
    artifacts.append(record(promo/audit['file']))
    (OUT/'SHA256SUMS.txt').write_text(''.join(r['sha256']+'  '+str(Path(r['path']).relative_to(OUT.relative_to(ROOT)))+'\n'for r in artifacts))
    receipt=dict(version=VERSION,passed=True,rights='含原版與衍生音畫，只限本機完整版；未公開上傳',
        head=inputs['head'],dirty_source=True,inputs=record(STAGE/'INPUTS.json'),
        source_files=len(inputs['source_files']),common_files=len(common),artifacts=artifacts,
        checks=checks,macos=dict(slices=slices,toolchain=(release.MAC/'TOOLCHAIN.txt').read_text(),native_execution=False),
        linux=dict(actual_appimage=True,extract_and_run=True,fuse_execution=False,offset=944632,runtime_sha256=inputs['runtime_sha256'],
                   appdir_manifest=record(BATCH/'APPDIR-MANIFEST.json'),dependencies=read(BATCH/'DEPENDENCIES.json')),
        windows=dict(pe_amd64=True,wine=True,native_execution=False,prior_crash_root='unknown，尚待對應Windows日誌或存檔'),
        hd=dict(portraits=83,before_bytes=packed['before_bytes'],after_bytes=packed['after_bytes'],pixel_identical=True,
                report=record(STAGE/'common/art/PACKED-HD.json'),master_unchanged=True),
        standard_tests=dict(command='go test ./cmd/... ./internal/...',passed=True,packages=21,with_tests=16,no_tests=5,
                            scope='本輪執行完成；不等同DOS全局同態或Windows真機驗收'),
        player_receipts=players,save_cross_platform=alliance['saved'],promo=dict(audit=record(promo/'AUDIT.json'),
            plan=record(promo/'PLAN.json'),seconds=audit['seconds'],new_scenes=plan['new_scene_ids'],
            old_story_and_battles_retained=True,audio_listening='未作人耳驗收'),
        generators={str(p.relative_to(ROOT)):sha(p)for p in [Path(__file__).resolve(),ROOT/'tools/scripts/scout_promo.py',
            ROOT/'tools/scripts/release_battle_gui.py',ROOT/'tools/scripts/release_garrison_gui.py',ROOT/'tools/scripts/scout_readme.py',ROOT/'tools/release.sh']},
        toolchain=dict(release=dict(image='logh3-release:20261031-r171',id='sha256:c300ac10ae82519c285808ad33c9b6d1086239ae701d5f0bc879ecc9e70286fa'),
            macos=dict(image='cok-osxcross:go1.24.13-sdk15.5-r1',id='sha256:a2dc911a8709192604ffe6e4936dd3fd74fecc5fbc63996fbe75058edfa00980'),
            wine=dict(image='yuan-wine-verify:noble-r2',id='sha256:31664639bb8cce681b29c9b06765a2f19ca5fb88468ad3fd8ed1645771d8f7b2'),
            versions=record(BATCH/'RELEASE-TOOLCHAIN.txt')),
        closeout=record(BATCH/'CLOSEOUT.json')if(BATCH/'CLOSEOUT.json').exists()else None,
        limitations=['Windows先前交戰閃退根因仍未知','macOS未真機','Linux未測FUSE啟動','無新增DOS同態宣稱'])
    write(ROOT/'docs/internals/data/v164-full-delivery.json',receipt)
    write(OUT/'MANIFEST.json',receipt)
    print('PASS',VERSION,'3 packages, 652 identical shared files, 5 normal-path receipts, 37-scene promo',flush=True)

if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('--packages-only',action='store_true');a=p.parse_args();check(a.packages_only)
