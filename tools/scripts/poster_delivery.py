"""150：獨立核對v1.6.0三平台、正式GUI、README實錄與新影片。Docker內執行。"""
import argparse
import json
import os
from pathlib import Path
import plistlib
import shutil
import struct
import zipfile

from PIL import Image, ImageChops

import poster_release as build
from local_test_macos_package import universal

ROOT, OUT, STAGE, VERSION = build.ROOT, build.OUT, build.STAGE, build.VERSION
read, write, sha = build.read, build.write, build.sha
BATCH = ROOT/'workplace/poster150'
LINUX = BATCH/'package-linux'/('LOGH3-'+VERSION+'-linux-amd64')
WINROOT = ROOT/'workplace/music69-windows/poster150'
WINDOWS = WINROOT/'installed'/('LOGH3-'+VERSION+'-windows-amd64')


def record(path): return dict(path=str(path.relative_to(ROOT)), bytes=path.stat().st_size, sha256=sha(path))


def packages(check=False):
    inputs = build.source_check(); rows = []; architectures = {}
    for name, platform in [('WINDOWS', 'windows-amd64'), ('LINUX', 'linux-amd64'), ('MACOS', 'macos-universal')]:
        result = read(STAGE/(name+'.json')); item = result['package']; archive = OUT/item['path']; folder = archive.with_suffix('')
        assert sha(archive) == item['sha256'] and archive.stat().st_size == item['bytes']
        manifest = read(folder/'MANIFEST.json'); assert manifest['version'] == VERSION
        with zipfile.ZipFile(archive) as z:
            assert z.testzip() is None
            assert set(z.namelist()) == {folder.name+'/'+n for n in manifest['files']} | {folder.name+'/MANIFEST.json'}
            for n, expected in manifest['files'].items():
                entry = z.getinfo(folder.name+'/'+n); raw = z.read(entry)
                assert not Path(entry.filename).is_absolute() and '..' not in Path(entry.filename).parts
                assert entry.filename.isascii() or entry.flag_bits & 0x800
                import hashlib
                assert len(raw) == expected['bytes'] and hashlib.sha256(raw).hexdigest() == expected['sha256']
                assert sha(folder/n) == expected['sha256']
                assert (entry.external_attr >> 16)&0o777 == int(expected['mode'], 8)
                if n == 'gin3' or n.endswith(('.sh', '.command', '/MacOS/launch', '/MacOS/gin3')): assert (entry.external_attr >> 16)&0o111
        assets = folder/'銀河英雄傳說III.app/Contents/Resources' if platform == 'macos-universal' else folder
        for n, expected in inputs['common_files'].items(): assert sha(assets/n) == expected['sha256'], (platform, n)
        assert not any(f.name == 'preferences.json' for f in folder.rglob('*'))
        assert read(folder/'VERSION.json')['version'] == VERSION
        notes = (folder/'README.txt').read_bytes(); assert notes.startswith(b'\xef\xbb\xbf') and 'v1' in notes.decode('utf-8-sig') and '한국어' in notes.decode('utf-8-sig')
        poster_notes = (assets/'art/poster/README.txt').read_bytes(); assert poster_notes.startswith(b'\xef\xbb\xbf') and b'\n' not in poster_notes.replace(b'\r\n', b'')
        assert sha(assets/'data/ui-locales.tsv') == sha(ROOT/'internal/localize/ui.tsv')
        for file in ['empire.png','alliance.png','v2-empire-narrow.png','v2-alliance-narrow.png','v3-empire-narrow.png','v3-alliance-narrow.png']:
            assert sha(assets/'art/poster'/file) == sha(ROOT/'assets/hd/poster'/file)
            with Image.open(assets/'art/poster'/file) as im: im.verify()
        if platform == 'windows-amd64':
            assert b'\n' not in notes.replace(b'\r\n', b'')
            for mode in ['original', 'enhanced']:
                cmd = (folder/('start-'+mode+'.cmd')).read_bytes(); assert cmd.isascii() and b'\n' not in cmd.replace(b'\r\n', b'')
            binary = folder/'gin3.exe'; data = binary.read_bytes(); offset = struct.unpack_from('<I', data, 0x3c)[0]
            assert data[:2] == b'MZ' and data[offset:offset+4] == b'PE\0\0' and struct.unpack_from('<H', data, offset+4)[0] == 0x8664
            architectures[platform] = 'PE amd64，Wine執行另驗'
        elif platform == 'linux-amd64':
            binary = folder/'gin3'; data = binary.read_bytes(); assert data[:6] == b'\x7fELF\x02\x01' and struct.unpack_from('<H', data, 18)[0] == 62
            architectures[platform] = 'ELF amd64'
        else:
            contents = folder/'銀河英雄傳說III.app/Contents'; info = plistlib.loads((contents/'Info.plist').read_bytes())
            assert info['LOGH3FullVersion'] == VERSION and info['CFBundleShortVersionString'] == info['CFBundleVersion'] == '1.6.0'
            binary = contents/'MacOS/gin3'; slices = universal(binary); assert slices == result['slices']
            assert all(s['deployment'][0]['sdk'] == '15.5.0' for s in slices)
            architectures[platform] = slices
        assert sha(binary) == result['binary_sha256']
        rows.append(dict(**item, files=len(manifest['files']), binary_sha256=sha(binary), manifest_sha256=sha(folder/'MANIFEST.json')))
        print('PASS', platform, 'ZIP／manifest／UTF-8／權限／615共用檔／三組海報／五語／架構', flush=True)
    result = dict(**inputs, packages=rows, architectures=architectures,
                  build_images={'windows_linux':'logh3-ebiten:2', 'macos':'wizardry7-osxcross:go1.24.13-python3.13.15-15.5-r1'},
                  limits=['Windows採Wine，非真機', 'macOS檢查雙架構、依賴、ARM簽章，未真機', '未新增DOS parity'])
    target = OUT/'full/BUILD.json'
    if check: assert read(target) == result
    else:
        assert not target.exists(); write(target, result)
        (OUT/'full/SHA256SUMS.txt').write_text(''.join(r['sha256']+'  '+Path(r['path']).name+'\n' for r in rows)+sha(target)+'  BUILD.json\n')
    return result


def captures(folder):
    result = read(folder/'RESULT.json'); assert result['passed'] and result['normal_entry']
    assert not any(result[k] for k in ['state_injection', 'seed_injection', 'script_flag'])
    events = {e['name']: e for e in result['events'] if 'name' in e and ('screenshot' in e or 'file' in e)}
    for e in events.values(): assert sha(folder/e.get('screenshot', e.get('file'))) == e['sha256']
    return result, events


def images(check=False):
    player, pe = captures(BATCH/'player-linux'); empire, ee = captures(BATCH/'empire-linux'); garrison, ge = captures(BATCH/'garrison-linux')
    assert player['binary_sha256'] == empire['binary_sha256'] == garrison['binary_sha256'] == sha(LINUX/'gin3')
    choices = [
        ('settings-empire', BATCH/'empire-linux', ee['zh-TW-settings-v1']),
        ('settings-alliance', BATCH/'player-linux', pe['zh-TW-settings-v1']),
        ('map-empire', BATCH/'empire-linux', ee['zh-TW-map-v1']),
        ('map-alliance', BATCH/'player-linux', pe['zh-TW-map-v1']),
        *[('poster-'+v, BATCH/'player-linux', pe['zh-TW-map-'+v]) for v in ['v1','v2','v3']],
        *[('locale-'+v, BATCH/'player-linux', pe[v+'-settings-v1']) for v in ['zh-CN','ja','en','ko']],
        ('editor-basic', BATCH/'player-linux', pe['zh-TW-editor-basic']),
        ('editor-special', BATCH/'player-linux', pe['zh-TW-editor-special']),
        ('editor-en', BATCH/'player-linux', pe['en-editor-special']),
        ('editor-ko', BATCH/'player-linux', pe['ko-editor-basic']),
        ('garrison-commander', BATCH/'garrison-linux', ge['changed-commander']),
        ('garrison-stock', BATCH/'garrison-linux', ge['local-ship-stock']),
        ('garrison-fleet', BATCH/'garrison-linux', ge['fleet-ready']),
    ]
    rows = []
    for name, folder, event in choices:
        source = folder/event.get('screenshot', event.get('file')); target = ROOT/'docs/images'/('readme-v160-'+name+'.png')
        assert Image.open(source).size == (1280, 720)
        if check: assert sha(target) == sha(source)
        else: assert not target.exists(); shutil.copyfile(source, target)
        rows.append(dict(file=str(target.relative_to(ROOT)), source=str(source.relative_to(ROOT)), sha256=sha(target), bytes=target.stat().st_size, observed=event['observed']))
    result = dict(version=VERSION, source='實際v1.6.0 Linux完整ZIP正常主選單操作，完整1280×720視窗原圖，無重畫或縮放',
                  archive_sha256=sha(OUT/'full'/('LOGH3-'+VERSION+'-linux-amd64.zip')), binary_sha256=sha(LINUX/'gin3'),
                  receipts=[record(BATCH/name/'RESULT.json') for name in ['player-linux','empire-linux','garrison-linux']], files=rows)
    target = ROOT/'docs/images/readme-v160.json'
    if check: assert read(target) == result
    else: assert not target.exists(); write(target, result)
    print('PASS README', len(rows), '張正式包完整視窗原圖', flush=True)
    return result


def gui(folder, binary):
    result, events = captures(folder); assert result['binary_sha256'] == sha(binary)
    comparisons = []
    for lang in ['zh-TW','zh-CN','ja','en','ko']:
        off = Image.open(folder/(lang+'-map-off.png')).convert('RGB'); on = Image.open(folder/(lang+'-map-v1.png')).convert('RGB')
        assert off.size == (960,720) and on.size == (1280,720)
        assert ImageChops.difference(off, on.crop((160,0,1120,720))).getbbox() is None, (folder,lang)
        comparisons.append(dict(locale=lang,pixels=691200,different=0))
        for set_name in ['v1','v2','v3']:
            e = events[lang+'-map-'+set_name]; assert e['observed']['poster_mode'] and e['observed']['poster_set'] == set_name and e['observed']['locale'] == lang
        for key in ['settings-v1','editor-basic','editor-special','help']:
            assert events[lang+'-'+key]['observed']['locale'] == lang
    assert result['preference']['locale'] == 'ko' and result['preference']['poster_set'] == 'v3'
    assert events['ko-restart-main']['observed']['locale'] == 'ko'
    assert len(result['saved']) == 6
    for name,digest in result['saved'].items(): assert sha(folder/'save'/name) == digest
    return dict(**record(folder/'RESULT.json'), screenshots=len(events), binary_sha256=sha(binary), comparisons=comparisons, saved=result['saved'])


def garrison(folder, binary):
    result, events = captures(folder); assert result['binary_sha256'] == sha(binary) and len(events) == 19
    get = lambda name: events[name]['observed']['strategic']
    local, ready = get('garrison-arrived'), get('fleet-ready')
    assert local['stock'] == [9000,3000,16000,2000,12000,3000,1000]
    assert ready['stock'] == [6000,3000,12000,2000,9000,3000,1000]
    assert local['capital_stock'] == ready['capital_stock'] == [102000,30000,104000,36000,9000,6000,5000]
    assert ready['fleets'][0]['Groups'][0]['kind'] == 2 and ready['fleets'][0]['Admiral'] == 41
    assert result['loaded']['strategic']['fleets'][0]['Region'] == 18
    assert result['restarted']['rules']['non_battleship_flagship']
    for name in ['support-rejected-5','support-rejected-6']: assert get(name)['message']
    for name,digest in result['saved'].items(): assert sha(folder/'save'/name) == digest
    return dict(**record(folder/'RESULT.json'), screenshots=len(events), binary_sha256=sha(binary), stock_before=local['stock'],stock_after=ready['stock'],saved=result['saved'])


def receipt(check=False, refresh_promo=False):
    previous = None
    if refresh_promo:
        assert not check
        backup = ROOT/'workplace/poster151/previous/docs/internals/data/v160-delivery.json'
        assert sha(backup) == sha(ROOT/'docs/internals/data/v160-delivery.json')
        previous = read(backup)
    built = packages(True); image_index = images(True)
    players = [gui(BATCH/'player-linux', LINUX/'gin3'), gui(WINROOT/'player-windows', WINDOWS/'gin3.exe')]
    assert players[0]['saved'] == players[1]['saved']
    garrisons = [garrison(BATCH/'garrison-linux', LINUX/'gin3'),garrison(WINROOT/'garrison-windows', WINDOWS/'gin3.exe')]
    assert garrisons[0]['saved'] == garrisons[1]['saved']
    diagnostics = []
    for folder,binary in [(BATCH/'diagnostic-linux',LINUX/'gin3'),(WINROOT/'diagnostic-windows',WINDOWS/'gin3.exe')]:
        r = read(folder/'RESULT.json'); assert r['passed'] and r['version'] == VERSION and r['binary_sha256'] == sha(binary) and len(r['runs']) == 3
        assert all(not s['final_frame']['poster_mode'] and s['final_frame']['poster_available'] for s in r['runs'])
        diagnostics.append(record(folder/'RESULT.json'))
    native,_ = captures(BATCH/'native-linux'); assert native['binary_sha256'] == sha(LINUX/'gin3')
    tests = (BATCH/'standard-final.log').read_text().splitlines()
    assert sum(s.startswith('ok ') for s in tests) == 16 and sum(s.startswith('? ') for s in tests) == 4
    assert not any(s.startswith(('FAIL','--- FAIL')) for s in tests)
    promo = OUT/'promo'; audit = read(promo/'AUDIT.json'); plan = read(promo/'PLAN.json')
    assert audit['passed'] and audit['visual_review'].startswith('已目視')
    assert sha(promo/audit['file']) == audit['sha256'] and sha(promo/'PLAN.json') == audit['plan_sha256']
    assert not audit['black_segments'] and not audit['subtitle_clipping'] and audit['audio']['snr_db'] > 25
    for name,digest in plan['input_sha256'].items(): assert sha(ROOT/name) == digest
    for folder in [LINUX,WINDOWS]:
        for name,row in read(folder/'MANIFEST.json')['files'].items(): assert sha(folder/name) == row['sha256']
    files = [OUT/r['path'] for r in built['packages']]+[OUT/'full/BUILD.json',OUT/'full/SHA256SUMS.txt']+[f for f in sorted(promo.iterdir()) if f.is_file()]
    checksum = ''.join(sha(f)+'  '+str(f.relative_to(OUT))+'\n' for f in files)
    if check: assert (OUT/'SHA256SUMS.txt').read_text() == checksum
    else: (OUT/'SHA256SUMS.txt').write_text(checksum)
    result = dict(passed=True, version=VERSION, packages=built['packages'], architectures=built['architectures'],
                  frozen_files=len(built['source_files']),product_files=sum(not n.startswith('vendor/') for n in built['source_files']),shared_files=len(built['common_files']),
                  build=record(OUT/'full/BUILD.json'),players=players,garrisons=garrisons,diagnostics=diagnostics,native=record(BATCH/'native-linux/RESULT.json'),
                  tests=dict(passed=16,no_tests=4,log=record(BATCH/'standard-final.log')),
                  images=dict(index=record(ROOT/'docs/images/readme-v160.json'),files=len(image_index['files'])),
                  movie=dict(file=str((promo/audit['file']).relative_to(ROOT)),sha256=audit['sha256'],seconds=audit['seconds'],audit=record(promo/'AUDIT.json')),
                  catalog=record(ROOT/'internal/localize/ui.tsv'),font=read(ROOT/'internal/localize/SOURCE.json'),
                  art=[record(ROOT/'assets/hd/poster'/n) for n in ['SOURCES.json','SETS.SOURCE.json','NARROW.SOURCE.json','QUOTE-SOURCES.json']],
                  scope='三組海報、五語主要操作介面與名言；原版圖片內文字及未收錄文字保留，無新增DOS parity；Windows Wine、macOS非真機',
                  rights='完整版及含原版音畫影片限本機。使用者授權README新實錄，未發布Release或移動tag。',checksums=record(OUT/'SHA256SUMS.txt'))
    result['generators'] = {str(f.relative_to(ROOT)):sha(f) for f in [Path(__file__).resolve(),ROOT/'tools/scripts/poster_release.py',ROOT/'tools/scripts/poster_sets_gui.py',ROOT/'tools/scripts/garrison_promo_gui.py',ROOT/'tools/scripts/poster_promo_movie.py',ROOT/'tools/scripts/garrison_promo_movie.py']}
    if plan.get('storyboard'):
        result['movie']['storyboard'] = plan['storyboard']
        result['movie']['recording_scope'] = plan['scope']
        result['previous_delivery'] = plan['previous_delivery']
        generator = ROOT/'tools/scripts/poster_story_promo.py'
        result['generators'][str(generator.relative_to(ROOT))] = sha(generator)
    target = ROOT/'docs/internals/data/v160-delivery.json'
    if check: assert read(target) == result
    elif refresh_promo:
        unchanged = set(previous) - {'movie','checksums','generators'}
        assert all(result[k] == previous[k] for k in unchanged), '影片重剪不能改封包、產品、GUI、字型或README身份'
        temporary = target.with_suffix('.json.tmp'); assert not temporary.exists()
        write(temporary, result); os.replace(temporary, target)
    else: assert not target.exists(); write(target,result)
    print(json.dumps(dict(passed=True,packages=result['packages'],images=result['images']['files'],movie=result['movie'],screenshots=sum(r['screenshots'] for r in players+garrisons)),ensure_ascii=False,indent=2))


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__); p.add_argument('mode',choices=['packages','images','receipt']); p.add_argument('--check',action='store_true'); p.add_argument('--refresh-promo',action='store_true'); a=p.parse_args()
    assert not a.refresh_promo or (a.mode == 'receipt' and not a.check)
    if a.mode == 'packages': packages(a.check)
    elif a.mode == 'images': images(a.check)
    else: receipt(a.check, a.refresh_promo)
