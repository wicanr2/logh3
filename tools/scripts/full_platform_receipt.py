"""147：獨立核對三平台完整ZIP、相同來源、素材、啟動器及正常玩家路徑。Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import plistlib
import struct
import zipfile

import full_platform_repack as build
from local_test_macos_package import universal

ROOT, OUT, STAGE, VERSION = build.ROOT, build.OUT, build.STAGE, build.VERSION
read, write, sha = build.read, build.write, build.sha


def record(path):
    return dict(path=str(path.relative_to(ROOT)), bytes=path.stat().st_size, sha256=sha(path))


def main(check):
    inputs = build.source_check()
    rows, architectures = [], {}
    for name, platform in [('WINDOWS', 'windows-amd64'), ('LINUX', 'linux-amd64'), ('MACOS', 'macos-universal')]:
        result = read(STAGE/(name+'.json')); item = result['package']; archive = OUT/item['path']
        assert sha(archive) == item['sha256'] and archive.stat().st_size == item['bytes']
        folder = OUT/'full'/archive.stem
        manifest = read(folder/'MANIFEST.json'); assert manifest['version'] == VERSION
        with zipfile.ZipFile(archive) as z:
            assert z.testzip() is None
            assert set(z.namelist()) == {folder.name+'/'+n for n in manifest['files']} | {folder.name+'/MANIFEST.json'}
            for n, expected in manifest['files'].items():
                full = folder.name+'/'+n; entry = z.getinfo(full)
                assert not Path(full).is_absolute() and '..' not in Path(full).parts
                assert full.isascii() or entry.flag_bits & 0x800
                b = z.read(full)
                assert len(b) == expected['bytes'] and hashlib.sha256(b).hexdigest() == expected['sha256'], n
                assert sha(folder/n) == expected['sha256']
                assert (entry.external_attr >> 16) & 0o777 == int(expected['mode'], 8)
                if n.endswith(('.sh', '.command', '/MacOS/gin3', '/MacOS/launch')) or n == 'gin3':
                    assert (entry.external_attr >> 16) & 0o111, n
        assets = folder/'銀河英雄傳說III.app/Contents/Resources' if platform == 'macos-universal' else folder
        for n, info in inputs['common_files'].items():
            assert (assets/n).stat().st_size == info['bytes'] and sha(assets/n) == info['sha256'], (platform, n)
        assert not any(p.name == 'preferences.json' for p in folder.rglob('*'))
        assert read(folder/'VERSION.json')['version'] == VERSION
        notes = (folder/'README.txt').read_bytes(); assert notes.startswith(b'\xef\xbb\xbf')
        assert '非戰艦旗艦' in notes.decode('utf-8-sig')
        if platform == 'windows-amd64':
            assert b'\n' not in notes.replace(b'\r\n', b'')
            for mode in ['original', 'enhanced']:
                cmd = (folder/('start-'+mode+'.cmd')).read_bytes()
                assert cmd.isascii() and b'\n' not in cmd.replace(b'\r\n', b'')
            binary = folder/'gin3.exe'; data = binary.read_bytes()
            offset = struct.unpack_from('<I', data, 0x3c)[0]
            assert data[:2] == b'MZ' and data[offset:offset+4] == b'PE\0\0' and struct.unpack_from('<H', data, offset+4)[0] == 0x8664
            assert sha(binary) == read(OUT/'BUILD.json')['files']['gin3.exe']['sha256'], '新建EXE須與146已驗來源同一編碼'
            architectures[platform] = 'PE amd64，與146替換包EXE相同'
        elif platform == 'linux-amd64':
            binary = folder/'gin3'; data = binary.read_bytes()
            assert data[:6] == b'\x7fELF\x02\x01' and struct.unpack_from('<H', data, 18)[0] == 62
            architectures[platform] = 'ELF 64-bit amd64'
        else:
            contents = folder/'銀河英雄傳說III.app/Contents'
            info = plistlib.loads((contents/'Info.plist').read_bytes())
            assert info['LOGH3FullVersion'] == VERSION and info['CFBundleShortVersionString'] == '1.5.2' and info['CFBundleVersion'] == '1.5.2'
            binary = contents/'MacOS/gin3'; slices = universal(binary)
            assert slices == result['slices']
            assert all(s['deployment'][0]['sdk'] == '15.5.0' for s in slices)
            architectures[platform] = slices
        assert sha(binary) == result['binary_sha256']
        rows.append(dict(**item, files=len(manifest['files']), binary_sha256=sha(binary), manifest_sha256=sha(folder/'MANIFEST.json')))
        print('PASS', platform, 'ZIP／權限／UTF-8／相同素材／架構', flush=True)

    diagnostics, players = [], []
    linux_folder = STAGE/'unpacked'/('LOGH3-'+VERSION+'-linux-amd64')
    windows_folder = ROOT/'workplace/music69-windows/full147/diagnostic/unpacked/銀河 完整版'/('LOGH3-'+VERSION+'-windows-amd64')
    for base, diagnostic, player in [(linux_folder, STAGE/'diagnostic', STAGE/'player-linux'),
                                      (windows_folder, ROOT/'workplace/music69-windows/full147/diagnostic', ROOT/'workplace/music69-windows/full147/player-windows')]:
        binary = base/('gin3.exe' if base == windows_folder else 'gin3')
        d = read(diagnostic/'RESULT.json')
        assert d['passed'] and d['automatic_resources'] and len(d['runs']) == 3
        assert d['binary_sha256'] == sha(binary)
        for r in d['runs']:
            assert r['exit'] == 0 and r['final_frame']['graphics'] == 'hd'
            assert not r['final_frame']['rules']['non_battleship_flagship']
            assert sha(diagnostic/(r['name']+'.png')) == r['screenshot_sha256']
        value = read(player/'RESULT.json')
        assert value['passed'] and value['normal_entry'] and not any(value[k] for k in ['state_injection', 'seed_injection', 'script_flag'])
        assert value['binary_sha256'] == sha(binary)
        captures = {e['name']: e for e in value['events'] if e['kind'] == 'capture'}
        assert len(captures) == 19
        for e in captures.values(): assert sha(player/e['file']) == e['sha256']
        formed = captures['fleet-ready']['observed']['strategic']; arrived = captures['garrison-arrived']['observed']['strategic']
        assert arrived['capital_stock'] == formed['capital_stock']
        assert formed['region'] == 19 and formed['fleets'][0]['Groups'][0]['kind'] == 2
        assert value['loaded']['strategic']['fleets'][0]['Region'] == 18
        assert value['restarted']['rules']['non_battleship_flagship']
        assert len(value['saved']) == 6
        for n, digest in value['saved'].items(): assert sha(player/'save'/n) == digest
        # GUI沿146鍵鼠器唯讀掛載的原版；該輸入須與實際完整ZIP逐檔一致。
        for n, row in inputs['common_files'].items():
            if n.startswith('original/'):
                assert sha(ROOT/'org_game/extracted'/Path(n).relative_to('original')) == row['sha256'], n
        for n, row in read(base/'MANIFEST.json')['files'].items(): assert sha(base/n) == row['sha256']
        diagnostics.append(record(diagnostic/'RESULT.json'))
        players.append(dict(**record(player/'RESULT.json'), screenshots=19, saved=value['saved'], binary_sha256=value['binary_sha256']))

    frozen = dict(inputs, version=VERSION, packages=rows, architectures=architectures,
                  diagnostics=diagnostics, normal_player=players,
                  build_images=dict(windows_linux='logh3-ebiten:2', macos='wizardry7-osxcross:go1.24.13-python3.13.15-15.5-r1'),
                  limits=['Windows為Wine，未真機', 'macOS為雙架構／依賴／ARM簽章檢查，未真機', '未新增人耳或DOS parity驗收'],
                  tests='產品615項與146最後標準19套件15 PASS／4無測試來源一致；本輪沒有產品程式變更')
    checksum_file = OUT/'SHA256SUMS.txt'; prior_file = STAGE/'previous-SHA256SUMS.txt'
    if not check:
        assert not (OUT/'full/BUILD.json').exists()
        write(OUT/'full/BUILD.json', frozen)
        (OUT/'full/SHA256SUMS.txt').write_text(''.join(r['sha256']+'  '+Path(r['path']).name+'\n' for r in rows)+sha(OUT/'full/BUILD.json')+'  BUILD.json\n')
        checksum_file.write_text(prior_file.read_text()+''.join(sha(p)+'  '+str(p.relative_to(OUT))+'\n' for p in [*[OUT/r['path'] for r in rows], OUT/'full/BUILD.json', OUT/'full/SHA256SUMS.txt']))
    else: assert read(OUT/'full/BUILD.json') == frozen
    for line in checksum_file.read_text().splitlines():
        digest, name = line.split('  ', 1); assert sha(OUT/name) == digest, name
    evidence = dict(passed=True, version=VERSION, packages=rows,
                    build=record(OUT/'full/BUILD.json'), checksums=record(checksum_file),
                    frozen_files=len(inputs['source_files']), product_files=sum(not n.startswith('vendor/') for n in inputs['source_files']),
                    shared_asset_files=len(inputs['common_files']), architectures=architectures,
                    diagnostics=diagnostics, normal_player=players,
                    preserved=inputs['preserved'], previous_checksums=record(prior_file),
                    normal_path='兩平台各96步及19張正常X11；GUI原版掛載與封包original逐檔相同。另6條診斷入口未傳orig／data／art／music，驗自動封包路徑。',
                    test_scope=frozen['tests'], limits=frozen['limits'],
                    generators={str(p.relative_to(ROOT)): sha(p) for p in [Path(__file__).resolve(), ROOT/'tools/scripts/full_platform_repack.py', ROOT/'tools/scripts/full_package_smoke.py', ROOT/'tools/scripts/garrison_promo_gui.py']})
    index = ROOT/'docs/internals/data/v152-full-delivery.json'
    assert index.parent.stat().st_uid == os.getuid()
    if check: assert read(index) == evidence
    else:
        assert not index.exists(); write(index, evidence)
    print(json.dumps(dict(passed=True, packages=rows, frozen=evidence['frozen_files'], common=evidence['shared_asset_files'], screenshots=38), ensure_ascii=False, indent=2))


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__); p.add_argument('--check', action='store_true')
    main(p.parse_args().check)
