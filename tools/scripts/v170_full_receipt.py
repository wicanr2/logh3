"""177：v.1.7.0四平台本機完整版的實際封包、正常玩家路徑與推廣片核對。Docker內執行（tools/release.sh）。"""
import hashlib
import json
import os
from pathlib import Path
import plistlib
import struct
import subprocess
import zipfile

import v170_full_release as release

ROOT, OUT, STAGE, BATCH, VERSION = release.ROOT, release.OUT, release.STAGE, release.BATCH, release.VERSION
read, write, sha = release.read, release.write, release.sha
WINE = ROOT/'workplace/music69-windows/full177'
PREV = ROOT/'workplace/release172'
GO_TESTS = BATCH/'standard-tests.txt'


def record(p): return dict(path=str(p.relative_to(ROOT)), bytes=p.stat().st_size, sha256=sha(p))


def main():
    assert os.getuid() == 1000
    release.configure(); inputs = read(STAGE/'INPUTS.json'); common = inputs['common_files']
    assert release.package.inventory(STAGE/'source') == inputs['source_files']
    assert release.package.inventory(STAGE/'common') == common
    assert common == read(release.PARENT_INPUTS)['common_files'], '共用素材與v1.6.5不同'
    # 凍結之後只改了行動版殼（Android重建APK、整場測試），桌面執行檔不引用它。
    later = sorted(n for n, row in inputs['source_files'].items() if not n.startswith('vendor/') and (not (ROOT/n).exists() or sha(ROOT/n) != row['sha256']))
    assert all(n.startswith('internal/mobileshell/') for n in later), later
    folders = {name: OUT/'full'/('LOGH3-'+VERSION+'-'+name) for name in ['windows-amd64', 'macos-universal']}
    artifacts = []; checks = {}
    for name, folder in folders.items():
        release.package.verify(folder)
        archive = folder.parent/(folder.name+'.zip')
        with zipfile.ZipFile(archive) as z:
            assert z.testzip() is None
            rows = read(folder/'MANIFEST.json')['files']; assert len(z.infolist()) == len(rows)+1
            for n, row in rows.items():
                e = z.getinfo(folder.name+'/'+n); data = z.read(e)
                assert len(data) == row['bytes'] and hashlib.sha256(data).hexdigest() == row['sha256']
                assert (e.external_attr >> 16) & 0o777 == int(row['mode'], 8)
                assert e.filename.isascii() or e.flag_bits & 0x800
        assert read(folder/'VERSION.json')['version'] == VERSION
        artifacts.append(record(archive)); checks[name] = dict(crc=True, manifest_files=len(rows), permissions=True, utf8_names=True)
    windows = folders['windows-amd64']; wb = windows/'gin3.exe'; pe = wb.read_bytes()
    assert pe[:2] == b'MZ'; p = struct.unpack_from('<I', pe, 0x3c)[0]
    assert pe[p:p+4] == b'PE\0\0' and struct.unpack_from('<H', pe, p+4)[0] == 0x8664 and VERSION.encode() in pe
    assert (windows/'README.txt').read_bytes().startswith(b'\xef\xbb\xbf')
    mac = folders['macos-universal']; app = mac/'銀河英雄傳說III.app/Contents'
    plist = plistlib.loads((app/'Info.plist').read_bytes())
    assert plist['LOGH3FullVersion'] == VERSION and plist['CFBundleVersion'] == plist['CFBundleShortVersionString'] == '1.7.0'
    slices = release.app.previous.universal(app/'MacOS/gin3'); resources = app/'Resources'
    linux = BATCH/'extracted/squashfs-root/usr/bin'; release.package.verify(linux)
    appimage = OUT/'full'/('LOGH3-'+VERSION+'-linux-x86_64.AppImage')
    assert hashlib.sha256(appimage.open('rb').read(944632)).hexdigest() == inputs['runtime_sha256']
    with appimage.open('rb') as f: f.seek(944632); assert f.read(4) == b'hsqs'
    assert (linux/'gin3').read_bytes()[:6] == b'\x7fELF\x02\x01' and VERSION.encode() in (linux/'gin3').read_bytes()
    assert read(linux/'VERSION.json')['version'] == VERSION
    artifacts.append(record(appimage))
    for folder in [windows, resources, linux]:
        for name, row in common.items():
            f = folder/name; assert f.stat().st_size == row['bytes'] and sha(f) == row['sha256'], str(f)
        assert not list(folder.rglob('preferences.json'))
    # Android：交付目錄的APK、素材與簽章已由 v170_full_release.py android 核對；這裡複核雜湊。
    apk = OUT/'full'/release.ANDROID_NAME; android = read(OUT/'full/ANDROID.json')
    assert android['package']['sha256'] == sha(apk); artifacts.append(record(apk))
    # 正常玩家路徑。
    players = {}
    for name, path, binary in [('linux-newgame-empire', BATCH/'newgame-empire', appimage), ('linux-newgame-alliance', BATCH/'newgame-alliance', appimage),
                               ('linux-garrison', BATCH/'garrison-linux', appimage),
                               ('windows-newgame-empire', WINE/'newgame-empire', wb), ('windows-newgame-alliance', WINE/'newgame-alliance', wb)]:
        r = read(path/'RESULT.json')
        assert r['passed'] and r['normal_entry'], name
        assert not any(r.get(k, False) for k in ['state_injection', 'seed_injection', 'time_injection', 'script_flag']), name
        assert r['binary_sha256'] == sha(binary), name
        if binary == appimage: assert r['appimage'] and r['automatic_resources'], name
        players[name] = record(path/'RESULT.json')
    new = {}
    for f in ['empire', 'alliance']:
        a, b = read(BATCH/('newgame-'+f)/'RESULT.json'), read(WINE/('newgame-'+f)/'RESULT.json')
        m = read(BATCH/'mobile-e2e'/f'TestMobileNewGameFullBattle-{f}.json')['saved_sha256']
        old = read(PREV/('newgame-'+f)/'RESULT.json')['saved_sha256']
        assert not a['panic'] and not b['panic'] and a['battle_after_load'] and b['battle_after_load'] and a['camera_on_player'] and b['camera_on_player']
        assert a['saved_sha256'] == b['saved_sha256'] == m == old, f
        new[f] = dict(battle_region=a['battle_region'], saves_identical=['AppImage', 'Wine', '行動版殼', 'v1.6.5'], saved_sha256=a['saved_sha256'])
    g_now, g_old = read(BATCH/'garrison-linux/RESULT.json')['saved'], read(PREV/'garrison-linux/RESULT.json')['saved']
    assert g_now == g_old
    staff = {}
    for f in ['empire', 'alliance']:
        folder = BATCH/('staff-linux' if f == 'empire' else 'staff-linux-2')
        shots = sorted(folder.glob(f + '-*.png')); assert len(shots) == 21, (f, len(shots))
        staff[f] = dict(folder=str(folder.relative_to(ROOT)), screenshots={p.name: sha(p) for p in shots},
                        visual='預算、建造、生產三入口更換參謀前後已逐張目視')
    emulator = {}
    for name, folder in [('saveload-alliance', BATCH/'emu-saveload2'), ('promo-recording', ROOT/'workplace/promo177/emu-rec2')]:
        emulator[name] = dict(folder=str(folder.relative_to(ROOT)), timeline=(folder/'timeline.txt').read_text(),
                              shots={p.name: sha(p) for p in sorted((folder/'shots').glob('*.png'))})
    tests = GO_TESTS.read_text().splitlines()
    ok = sum(t.startswith('ok') for t in tests); none = sum('no test files' in t for t in tests)
    assert ok == 17 and none == 6 and not any(t.startswith('FAIL') or t.startswith('--- FAIL') for t in tests), (ok, none)
    promo = read(OUT/'promo/AUDIT.json'); movie = OUT/'promo'/promo['file']
    assert promo['passed'] and promo['visual_review'].startswith('已目視') and sha(movie) == promo['sha256']
    artifacts.append(record(movie))
    (OUT/'SHA256SUMS.txt').write_text(''.join(r['sha256']+'  '+str(Path(r['path']).relative_to(OUT.relative_to(ROOT)))+'\n' for r in artifacts))
    receipt = dict(version=VERSION, passed=True, rights='含原版與衍生音畫，只限本機完整版；未公開上傳',
        head=inputs['head'], product_sources_clean=True, working_tree_at_prepare=inputs['working_tree'], inputs=record(STAGE/'INPUTS.json'),
        source_files=len(inputs['source_files']), common_files=len(common), common_identical_to='v.1.6.5-20261102', artifacts=artifacts, checks=checks,
        macos=dict(slices=slices, toolchain=(release.MAC/'TOOLCHAIN.txt').read_text(), native_execution=False),
        linux=dict(actual_appimage=True, extract_and_run=True, fuse_execution=False, offset=944632, runtime_sha256=inputs['runtime_sha256'],
                   appdir_manifest=record(BATCH/'APPDIR-MANIFEST.json'), dependencies=read(BATCH/'DEPENDENCIES.json')),
        windows=dict(pe_amd64=True, wine=True, native_execution=False),
        android=dict(package=android['package'], abis=android['abis'], min_sdk=android['min_sdk'], signing=android['signing'],
                     apksigner=android['apksigner'], assets=android['assets'], built_from='e51caa7d；桌面封包凍結於'+inputs['head'][:8]+'，其間的產品來源變更只在internal/mobileshell（桌面不引用）',
                     source_changed_after_freeze=later,
                     emulator=emulator, native_device=False),
        standard_tests=dict(command='go test ./cmd/... ./internal/...', passed=True, with_tests=ok, no_tests=none, log=record(GO_TESTS)),
        new_game=new, garrison=dict(saved_identical_to_v165=True, receipt=players['linux-garrison']), staff_pick=staff,
        mobile_shell_full_battle=dict(test='TestMobileNewGameFullBattle（LOGH3_MOBILE_E2E=1）', log=record(BATCH/'mobile-e2e.log'),
                                      results={p.name: sha(p) for p in sorted((BATCH/'mobile-e2e').glob('*.json'))}),
        player_receipts=players, promo=dict(file=record(movie), audit=promo),
        generators={str(p.relative_to(ROOT)): sha(p) for p in [Path(__file__).resolve(), ROOT/'tools/scripts/v170_full_release.py', ROOT/'tools/scripts/v170_promo.py',
            ROOT/'tools/scripts/new_game_battle_gui.py', ROOT/'tools/scripts/release_garrison_gui.py', ROOT/'tools/staff_pick_gui.sh',
            ROOT/'tools/android.sh', ROOT/'tools/android_emulator_check.sh', ROOT/'tools/release.sh']},
        toolchain=dict(release=dict(image='logh3-release:20261031-r171', id='sha256:c300ac10ae82519c285808ad33c9b6d1086239ae701d5f0bc879ecc9e70286fa'),
            macos=dict(image='cok-osxcross:go1.24.13-sdk15.5-r1', id='sha256:a2dc911a8709192604ffe6e4936dd3fd74fecc5fbc63996fbe75058edfa00980'),
            wine=dict(image='yuan-wine-verify:noble-r2', id='sha256:31664639bb8cce681b29c9b06765a2f19ca5fb88468ad3fd8ed1645771d8f7b2'),
            android=dict(image='logh3-android:20261119-r1'), emulator=dict(image='logh3-android-emulator:api35-20261119-r1')),
        limitations=['Windows只在Wine驗證，未上實機', 'macOS未真機', 'Linux未測FUSE啟動', 'Android只在模擬器驗證，未上實機；沒有人耳聽配樂',
                     'Wine同盟戰後截圖仍是戰場畫面（狀態已是戰略畫面，v1.6.5同樣），判為Wine截圖時視窗未重繪', '無新增DOS同態宣稱'])
    write(ROOT/'docs/internals/data/v170-full-delivery.json', receipt)
    write(OUT/'MANIFEST.json', receipt)
    print('PASS', VERSION, len(artifacts), 'artifacts,', len(common), 'identical shared files,', len(players), 'normal-path receipts', flush=True)


if __name__ == '__main__':
    main()
