"""172：v.1.6.5三平台完整版的實際封包與正常玩家路徑核對。推廣片沿用v.1.6.4，不在本版範圍。"""
import hashlib
import json
import os
from pathlib import Path
import plistlib
import struct
import subprocess
import zipfile

import new_game_full_release as release

ROOT,OUT,STAGE,BATCH,VERSION=release.ROOT,release.OUT,release.STAGE,release.BATCH,release.VERSION
read,write,sha=release.read,release.write,release.sha
WINE=ROOT/'workplace/music69-windows/full172'
GO_TESTS=ROOT/'workplace/release172/standard-tests.txt'


def record(p):return dict(path=str(p.relative_to(ROOT)),bytes=p.stat().st_size,sha256=sha(p))


def main():
    assert os.getuid()==1000
    release.configure();inputs=read(STAGE/'INPUTS.json');common=inputs['common_files']
    assert release.package.inventory(STAGE/'source')==inputs['source_files']
    assert release.package.inventory(STAGE/'common')==common
    assert common==read(release.PARENT_INPUTS)['common_files'],'共用素材與v1.6.4不同'
    for n,row in inputs['source_files'].items():
        if not n.startswith('vendor/'):assert sha(ROOT/n)==row['sha256'],n
    for n,h in inputs['generators'].items():assert sha(ROOT/n)==h,n
    folders={name:OUT/'full'/('LOGH3-'+VERSION+'-'+name)for name in['windows-amd64','macos-universal']}
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
                assert(e.external_attr>>16)&0o777==int(row['mode'],8)
                assert e.filename.isascii()or e.flag_bits&0x800
        assert read(folder/'VERSION.json')['version']==VERSION
        artifacts.append(record(archive));checks[name]=dict(crc=True,manifest_files=len(rows),permissions=True,utf8_names=True)
    windows=folders['windows-amd64'];wb=windows/'gin3.exe';pe=wb.read_bytes()
    assert pe[:2]==b'MZ';p=struct.unpack_from('<I',pe,0x3c)[0]
    assert pe[p:p+4]==b'PE\0\0'and struct.unpack_from('<H',pe,p+4)[0]==0x8664
    assert VERSION.encode()in pe
    assert(windows/'README.txt').read_bytes().startswith(b'\xef\xbb\xbf')
    for f in['start-original.cmd','start-enhanced.cmd','capture-battle-original.cmd','capture-battle-enhanced.cmd']:
        data=(windows/f).read_bytes();assert b'\r\n'in data and b'\n'not in data.replace(b'\r\n',b'')
    mac=folders['macos-universal'];app=mac/'銀河英雄傳說III.app/Contents'
    plist=plistlib.loads((app/'Info.plist').read_bytes())
    assert plist['LOGH3FullVersion']==VERSION
    assert plist['CFBundleVersion']==plist['CFBundleShortVersionString']=='1.6.5'
    slices=release.app.previous.universal(app/'MacOS/gin3')
    resources=app/'Resources'
    linux=BATCH/'extracted/squashfs-root/usr/bin';release.package.verify(linux)
    appimage=OUT/'full'/('LOGH3-'+VERSION+'-linux-x86_64.AppImage')
    data=appimage.open('rb').read(944632)
    assert hashlib.sha256(data).hexdigest()==inputs['runtime_sha256']
    with appimage.open('rb')as f:f.seek(944632);assert f.read(4)==b'hsqs'
    assert(linux/'gin3').read_bytes()[:6]==b'\x7fELF\x02\x01'
    assert VERSION.encode()in(linux/'gin3').read_bytes()
    def tree(folder):
        return{str(p.relative_to(folder)):('symlink',os.readlink(p))if p.is_symlink()else('file',sha(p),p.stat().st_mode&0o777)
               for p in folder.rglob('*')if p.is_symlink()or p.is_file()}
    assert tree(BATCH/'AppDir')==tree(BATCH/'extracted/squashfs-root')
    assert read(linux/'VERSION.json')['version']==VERSION
    artifacts.append(record(appimage))
    for folder in[windows,resources,linux]:
        for name,row in common.items():
            f=folder/name;assert f.stat().st_size==row['bytes']and sha(f)==row['sha256'],str(f)
        assert not list(folder.rglob('preferences.json'))
    players={}
    for name,path,binary in[('linux-alliance',BATCH/'player-linux',appimage),('linux-empire',BATCH/'player-empire',appimage),
                            ('windows-wine',WINE/'player-windows',wb),('linux-battle',BATCH/'battle-linux',appimage),
                            ('linux-garrison',BATCH/'garrison-linux',appimage),
                            ('linux-newgame-empire',BATCH/'newgame-empire',appimage),('linux-newgame-alliance',BATCH/'newgame-alliance',appimage),
                            ('windows-newgame-empire',WINE/'newgame-empire',wb),('windows-newgame-alliance',WINE/'newgame-alliance',wb)]:
        receipt=read(path/'RESULT.json')
        assert receipt['passed']and receipt['normal_entry'],name
        assert not any(receipt.get(k,False)for k in['state_injection','seed_injection','time_injection','script_flag']),name
        assert receipt['binary_sha256']==sha(binary),name
        if binary==appimage:assert receipt['appimage']and receipt['automatic_resources'],name
        players[name]=record(path/'RESULT.json')
    battle=read(BATCH/'battle-linux/RESULT.json')
    assert battle['pre_cleanup_alive']and battle['last']['state']=='StateBattle'and battle['last']['tactical']['turn']==3
    new={}
    for f in['empire','alliance']:
        a,b=read(BATCH/('newgame-'+f)/'RESULT.json'),read(WINE/('newgame-'+f)/'RESULT.json')
        assert not a['panic']and not b['panic']and a['battle_after_load']and b['battle_after_load']
        assert a['saved_sha256']==b['saved_sha256'],f
        new[f]=dict(battle_region=a['battle_region'],saves_identical_across_platforms=True,saved_sha256=a['saved_sha256'])
    # 回歸路徑是新遊戲直接存檔：除了WORK+0C..14（規格159刻意寫入），六表與v1.6.4逐位元組相同。
    regression={}
    for n,p,q,b in[('linux-alliance',BATCH/'player-linux',ROOT/'workplace/release171/player-linux',BATCH/'baseline-alliance.json'),
                   ('linux-empire',BATCH/'player-empire',ROOT/'workplace/release171/player-empire',BATCH/'baseline-empire.json'),
                   ('windows-wine',WINE/'player-windows',ROOT/'workplace/music69-windows/full171/player-windows',WINE/'baseline-alliance.json')]:
        now,old=read(p/'RESULT.json')['saved'],read(q/'RESULT.json')['saved']
        assert now==read(b)['saved'],n
        assert{k for k in now if now[k]!=old[k]}=={'WORK.DT0'},n
        regression[n]=dict(other_tables_identical=True,work_differs_only_in_seat_control=True,baseline=record(b))
    garrison_now,garrison_old=read(BATCH/'garrison-linux/RESULT.json')['saved'],read(ROOT/'workplace/release171/garrison-linux/RESULT.json')['saved']
    regression['linux-garrison']=dict(identical=garrison_now==garrison_old,changed=sorted(k for k in garrison_now if garrison_now[k]!=garrison_old[k]))
    (OUT/'SHA256SUMS.txt').write_text(''.join(r['sha256']+'  '+str(Path(r['path']).relative_to(OUT.relative_to(ROOT)))+'\n'for r in artifacts))
    tests=GO_TESTS.read_text().splitlines()
    assert sum(t.startswith('ok')for t in tests)==16 and sum('no test files'in t for t in tests)==5 and not any('FAIL'in t for t in tests)
    receipt=dict(version=VERSION,passed=True,rights='含原版與衍生音畫，只限本機完整版；未公開上傳',
        head=inputs['head'],branch=subprocess.check_output(['git','rev-parse','--abbrev-ref','HEAD'],text=True).strip(),product_sources_clean=True,
        working_tree_at_prepare=inputs['working_tree'],inputs=record(STAGE/'INPUTS.json'),
        source_files=len(inputs['source_files']),common_files=len(common),common_identical_to='v.1.6.4-20261031',artifacts=artifacts,
        checks=checks,macos=dict(slices=slices,toolchain=(release.MAC/'TOOLCHAIN.txt').read_text(),native_execution=False),
        linux=dict(actual_appimage=True,extract_and_run=True,fuse_execution=False,offset=944632,runtime_sha256=inputs['runtime_sha256'],
                   appdir_manifest=record(BATCH/'APPDIR-MANIFEST.json'),dependencies=read(BATCH/'DEPENDENCIES.json')),
        windows=dict(pe_amd64=True,wine=True,native_execution=False,
                     prior_crash_root='第169輪閃退已定位：新遊戲未存檔進第一場戰鬥，見docs/playtest/143'),
        standard_tests=dict(command='go test ./cmd/... ./internal/...',passed=True,packages=21,with_tests=16,no_tests=5,
                            log=record(GO_TESTS),scope='不等同DOS全局同態或Windows真機驗收'),
        new_game=new,regression_against_v164=regression,player_receipts=players,
        promo='沿用v.1.6.4推廣片，本版未重錄',
        generators={str(p.relative_to(ROOT)):sha(p)for p in[Path(__file__).resolve(),ROOT/'tools/scripts/new_game_battle_gui.py',
            ROOT/'tools/scripts/release_battle_gui.py',ROOT/'tools/scripts/release_garrison_gui.py',ROOT/'tools/scripts/poster_frame_gui.py',ROOT/'tools/scripts/new_game_player_baseline.py',ROOT/'tools/release.sh']},
        toolchain=dict(release=dict(image='logh3-release:20261031-r171',id='sha256:c300ac10ae82519c285808ad33c9b6d1086239ae701d5f0bc879ecc9e70286fa'),
            macos=dict(image='cok-osxcross:go1.24.13-sdk15.5-r1',id='sha256:a2dc911a8709192604ffe6e4936dd3fd74fecc5fbc63996fbe75058edfa00980'),
            wine=dict(image='yuan-wine-verify:noble-r2',id='sha256:31664639bb8cce681b29c9b06765a2f19ca5fb88468ad3fd8ed1645771d8f7b2')),
        limitations=['Windows只在Wine驗證，未上實機','macOS未真機','Linux未測FUSE啟動','無新增DOS同態宣稱'])
    write(ROOT/'docs/internals/data/v165-full-delivery.json',receipt)
    write(OUT/'MANIFEST.json',receipt)
    print('PASS',VERSION,'3 packages,',len(common),'identical shared files,',len(players),'normal-path receipts',flush=True)


if __name__=='__main__':
    main()
