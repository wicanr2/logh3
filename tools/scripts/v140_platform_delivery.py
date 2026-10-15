"""v.1.4.0：沿已驗Windows凍結輸入補齊Linux／macOS，保留舊包與收據。Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import plistlib
import re
import shutil
import subprocess
import zipfile

from local_test_macos_package import universal

ROOT = Path('/src')
VERSION = 'v.1.4.0-20261015'
OUT = ROOT/'dist-all'/VERSION
STAGE = ROOT/'workplace'/('package92-'+VERSION+'-release139')
WIN = OUT/'full'/('LOGH3-'+VERSION+'-windows-amd64')
MACOS = ROOT/'workplace/macos-release139'


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1024*1024), b''):
            h.update(block)
    return h.hexdigest()


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')


def inventory(folder):
    return {str(p.relative_to(folder)): dict(bytes=p.stat().st_size, sha256=sha(p),
            mode=oct(p.stat().st_mode & 0o777)) for p in sorted(folder.rglob('*')) if p.is_file()}


def verify(folder):
    manifest = json.loads((folder/'MANIFEST.json').read_text())
    for name, row in manifest['files'].items():
        p = folder/name
        assert p.stat().st_size == row['bytes'] and sha(p) == row['sha256'], name


def archive(folder, classification):
    assert folder.stat().st_uid == os.getuid()
    files = inventory(folder)
    write(folder/'MANIFEST.json', dict(version=VERSION, classification=classification, files=files))
    target = folder.parent/(folder.name+'.zip')
    assert not target.exists(), target
    with zipfile.ZipFile(target, 'x', compression=zipfile.ZIP_DEFLATED, compresslevel=6) as z:
        for p in sorted(folder.rglob('*')):
            if p.is_file():
                z.write(p, str(Path(folder.name)/p.relative_to(folder)))
    with zipfile.ZipFile(target) as z:
        assert z.testzip() is None
        for name, expected in files.items():
            entry = folder.name+'/'+name
            value = z.read(entry)
            assert len(value) == expected['bytes'] and hashlib.sha256(value).hexdigest() == expected['sha256']
            assert (z.getinfo(entry).external_attr >> 16) & 0o777 == int(expected['mode'], 8)
        assert all(i.filename.isascii() or i.flag_bits & 0x800 for i in z.infolist())
    row = dict(path=str(target.relative_to(OUT)), bytes=target.stat().st_size, sha256=sha(target))
    print('PASS ZIP逐檔／權限', row, flush=True)
    return row


def readme(platform, full):
    details = ('完整包已含原版、HD、兩種音色及音效，只限合法持有人本機使用，不公開散布。\n'
               if full else '引擎包不含原版、HD、音訊、SF2或ROM。請匯入合法持有的原版資料。\n')
    location = ('gin3旁的original' if platform == 'linux' else '銀河英雄傳說III.app/Contents/Resources/original')
    details += f'資料目錄為{location}，須包含GIN3.COM、GRAPH.IMG、FACE.AMG及其餘原版檔案。\n'
    details += ('start-original使用原版AI，start-enhanced使用強化AI；F1說明、F7原版／HD、防拷題138。\n'
	            '系統設定使用緊湊字型，顯示敵軍預設關閉，可在設定開啟。AI強化可選原版及強化0至3，數字越大越強。\n'
                '設定頁使用專用清晰字型，HD直接繪製高解析字形；視窗金屬外緣保留內側陰影。\n'
                '系統設定可選Roland MT-32／GeneralUser GS；轉移旗艦及名將移動加成預設開啟，可分別關閉並保存偏好。\n'
                '旗艦擊毀後謬拉100%、其他提督約25%，須有存活戰艦或高速戰艦，轉往最多艘數分隊並保留提督、副官。\n'
                '名將移動只限強化模式，帶隊楊威利／萊因哈特參戰時全陣營新移動額度1.5倍，關閉仍保留戰術AI。\n'
                '戰術保存可返回戰場，正式戰術讀回重建仍由#48追蹤；不宣稱DOS整場同態。\n')
    if platform == 'linux':
        details += '存檔預設為$XDG_CONFIG_HOME/logh3/save，未設定時為~/.config/logh3/save。\n完整解壓後執行./gin3或兩個.sh啟動器；如解壓工具未保留權限，將gin3及.sh標為可執行。\n'
    else:
        details += ('Intel／Apple Silicon通用程式，最低macOS11；可開.app或兩個.command。\n'
                    '存檔：~/Library/Application Support/logh3/save；錯誤紀錄：~/Library/Logs/logh3/game.log。\n'
                    '已核對雙架構、系統依賴、Apple Silicon ad-hoc程式頁簽章及ZIP權限，尚未在macOS真機執行。\n'
                    '沒有Developer ID或Apple公證；本包不變更系統安全設定。\n')
    return b'\xef\xbb\xbf'+('銀河英雄傳說III SP '+VERSION+' '+platform+'\n\n'+details).encode()


def prepare():
    assert not STAGE.exists()
    verify(WIN)
    build = json.loads((OUT/'BUILD.json').read_text())
    frozen = ROOT/'workplace'/('windows-package-'+VERSION)/'source'
    for name, row in build['source_files'].items():
        assert sha(frozen/name) == row['sha256'], name
        if not name.startswith('vendor/'):
            assert sha(ROOT/name) == row['sha256'], ('現行產品與凍結來源不同', name)
    STAGE.mkdir()
    shutil.copytree(frozen, STAGE/'source')
    common = STAGE/'common'
    common.mkdir()
    for name in ['original', 'data', 'art', 'music', 'licenses', 'soundfont']:
        shutil.copytree(WIN/name, common/name)
    shutil.copyfile(WIN/'LICENSE', common/'LICENSE')
    write(STAGE/'INPUTS.json', dict(version=VERSION, windows_build_sha256=sha(OUT/'BUILD.json'),
          windows_manifest_sha256=sha(WIN/'MANIFEST.json'), source_files=build['source_files'],
          head=subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
          working_tree=subprocess.check_output(['git', 'status', '--porcelain'], text=True)))
    print('READY 凍結來源與Windows正式素材逐檔一致', flush=True)


def metadata(platform):
    return dict(version=VERSION, platform=platform, frozen_inputs_sha256=sha(STAGE/'INPUTS.json'),
                source_files=json.loads((STAGE/'INPUTS.json').read_text())['source_files'])


def unix_folder(platform, full):
    folder = OUT/('full' if full else 'release')/('LOGH3-'+VERSION+'-'+platform)
    assert not folder.exists()
    folder.mkdir()
    assets = folder
    if platform == 'macos-universal':
        contents = folder/'銀河英雄傳說III.app/Contents'
        assets = contents/'Resources'
        assets.mkdir(parents=True)
        (contents/'MacOS').mkdir()
        short = VERSION.split('-')[0][2:]
        info = dict(CFBundleName='銀河英雄傳說III', CFBundleDisplayName='銀河英雄傳說III',
                    CFBundleIdentifier='org.logh3.remake', CFBundleExecutable='launch', CFBundlePackageType='APPL',
                    CFBundleShortVersionString=short, CFBundleVersion=short, LSMinimumSystemVersion='11.0',
                    NSHighResolutionCapable=True, LOGH3FullVersion=VERSION)
        (contents/'Info.plist').write_bytes(plistlib.dumps(info))
        shutil.copyfile(MACOS/'gin3-universal', contents/'MacOS/gin3')
        (contents/'MacOS/gin3').chmod(0o755)
        launcher = contents/'MacOS/launch'
        launcher.write_text('#!/bin/sh\nset -eu\n'
            'LOGH3_APP_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)\n'
            'LOGH3_RES_DIR="$LOGH3_APP_DIR/../Resources"\n'
            'mkdir -p "$HOME/Library/Logs/logh3"\n'
            'exec "$LOGH3_APP_DIR/gin3" -orig "$LOGH3_RES_DIR/original" -data "$LOGH3_RES_DIR/data" '
            + ('-art "$LOGH3_RES_DIR/art" -graphics hd -music "$LOGH3_RES_DIR/music" ' if full else '-music "" ')
            + '-ai -1 -scale 1.5 "$@" 2>>"$HOME/Library/Logs/logh3/game.log"\n')
        launcher.chmod(0o755)
    else:
        shutil.copyfile(STAGE/'gin3', folder/'gin3')
        (folder/'gin3').chmod(0o755)
    for name in (['original', 'data', 'art', 'music', 'licenses', 'soundfont'] if full else ['data', 'licenses']):
        shutil.copytree(STAGE/'common'/name, assets/name)
    shutil.copyfile(STAGE/'common/LICENSE', assets/'LICENSE')
    for mode, level in [('original', -1), ('enhanced', 1)]:
        quality = ' -music-quality hd' if full and mode == 'enhanced' else ''
        suffix = '.command' if platform == 'macos-universal' else '.sh'
        launcher = folder/('start-'+mode+suffix)
        launch = ('"./銀河英雄傳說III.app/Contents/MacOS/launch"' if platform == 'macos-universal' else './gin3')
        launcher.write_text('#!/bin/sh\nset -eu\ncd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"\n'
                            f'exec {launch} -ai {level} -scale 1.5{quality} "$@"\n')
        launcher.chmod(0o755)
    (folder/'README.txt').write_bytes(readme('macos' if platform == 'macos-universal' else 'linux', full))
    write(folder/'VERSION.json', metadata(platform))
    if not full:
        assert not any(p.is_file() and p.suffix.lower() in {'.com','.img','.amg','.ogg','.sf2','.rom','.dat','.png','.mid'} for p in folder.rglob('*'))
    return folder


def linux():
    assert not (STAGE/'gin3').exists()
    env = dict(os.environ, GOPROXY='off', GOSUMDB='off', GOWORK='off', GOOS='linux', GOARCH='amd64', CGO_ENABLED='1')
    subprocess.run(['go', 'build', '-mod=vendor', '-trimpath', '-buildvcs=false',
                    '-ldflags=-X main.buildVersion='+VERSION, '-o', str(STAGE/'gin3'), './cmd/gin3'],
                   cwd=STAGE/'source', env=env, check=True)
    rows = [archive(unix_folder('linux-amd64', full), '本機完整包' if full else '可散布引擎包') for full in [True, False]]
    write(STAGE/'LINUX.json', dict(version=VERSION, packages=rows,
          go_version=subprocess.check_output(['go','version'], text=True).strip()))


def macos():
    slices = universal(MACOS/'gin3-universal')
    rows = [archive(unix_folder('macos-universal', full), '本機完整包' if full else '可散布引擎包') for full in [True, False]]
    smoke = OUT/'smoke/macos'
    assert not smoke.exists()
    smoke.mkdir()
    for p in MACOS.glob('*.txt'):
        shutil.copyfile(p, smoke/p.name)
    write(smoke/'MACOS.json', dict(version=VERSION, slices=slices, runtime_test='未在macOS真機執行',
                                  packages=rows, zip_integrity=True, executable_permissions_preserved=True))


def finalize():
    files = []
    for name in ['full', 'release']:
        for p in sorted((OUT/name).glob('*')):
            if p.suffix == '.zip' or p.name.endswith('.tar.gz'):
                files.append(dict(path=str(p.relative_to(OUT)), bytes=p.stat().st_size, sha256=sha(p)))
    assert len(files) == 7 and len([r for r in files if r['path'].startswith('release/')]) == 4
    write(OUT/'SHA256SUMS-platforms.json', dict(version=VERSION, files=files))
    listing = ''.join(r['sha256']+'  '+Path(r['path']).name+'\n' for r in files if r['path'].startswith('release/'))
    target = OUT/'release/SHA256SUMS.txt'
    assert not target.exists()
    target.write_text(listing)
    print('PASS 七封包，Release三平台引擎／來源及SHA256SUMS.txt', flush=True)


def extract(name='package', extract_root=None):
    parent=Path(extract_root) if extract_root else ROOT/'workplace/promo139'
    assert parent.resolve().is_relative_to((ROOT/'workplace').resolve())
    assert parent.is_dir() and parent.stat().st_uid==os.getuid()
    target=parent/name
    assert not target.exists()
    target.mkdir()
    source=OUT/'full'/('LOGH3-'+VERSION+'-linux-amd64.zip')
    with zipfile.ZipFile(source) as z:
        assert z.testzip() is None
        for info in z.infolist():
            p=target/info.filename
            assert p.resolve().is_relative_to(target.resolve())
            z.extract(info,target)
            if not info.is_dir():p.chmod((info.external_attr>>16)&0o777)
    package=target/('LOGH3-'+VERSION+'-linux-amd64')
    verify(package)
    print('PASS 實際Linux完整ZIP解壓，逐檔及執行權限核對',flush=True)


def archive_pending():
    destination=ROOT/'workplace/promo139/precompact'
    assert not destination.exists()
    metadata=json.loads((OUT/'SHA256SUMS-platforms.json').read_text())
    for row in metadata['files']:
        p=OUT/row['path']
        assert p.stat().st_uid==os.getuid() and sha(p)==row['sha256'],p
    sources=[OUT,ROOT/'workplace'/('windows-package-'+VERSION),STAGE,ROOT/'workplace/macos-release139']
    assert all(p.is_dir() and p.stat().st_uid==os.getuid() for p in sources)
    assert destination.parent.stat().st_uid==os.getuid()
    destination.mkdir()
    mapping=[]
    for source in sources:
        target=destination/source.relative_to(ROOT)
        target.parent.mkdir(parents=True,exist_ok=True)
        shutil.move(str(source),target)
        mapping.append(dict(original=str(source.relative_to(ROOT)),archive=str(target.relative_to(ROOT))))
    write(destination/'MAPPING.json',dict(version=VERSION,reason='使用者指定未發布v.1.4.0設定頁修正後重建，舊候選完整保留',
          previous_packages=metadata,mapping=mapping))
    print('ARCHIVED 四個舊候選目錄，七封包SHA-256一致；未覆寫已發布tag或Release',flush=True)


if __name__ == '__main__':
    assert Path.cwd() == ROOT and os.getuid() == os.getgid() == 1000
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('mode', choices=['prepare', 'linux', 'macos', 'finalize', 'extract', 'archive_pending'])
    p.add_argument('--extract-name',choices=['package','package-compact'],default='package')
    p.add_argument('--extract-root',help='既有workplace子目錄，每次解壓建立獨立目錄')
    p.add_argument('--version',default=VERSION)
    p.add_argument('--batch',default='release139')
    args = p.parse_args()
    assert re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}',args.version)
    assert re.fullmatch(r'release[0-9]+',args.batch)
    VERSION=args.version
    OUT=ROOT/'dist-all'/VERSION
    STAGE=ROOT/'workplace'/('package92-'+VERSION+'-'+args.batch)
    WIN=OUT/'full'/('LOGH3-'+VERSION+'-windows-amd64')
    MACOS=ROOT/'workplace'/('macos-'+args.batch)
    if args.mode=='extract':extract(args.extract_name,args.extract_root)
    else:globals()[args.mode]()
