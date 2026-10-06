"""#92：目前程式的本機完整版測試包。僅在既有GUI Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tarfile
import zipfile

ROOT = Path('/src')


def digest(path):
    h = hashlib.sha256()
    with Path(path).open('rb') as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()


def owned(path):
    st = path.lstat()
    assert not path.is_symlink() and st.st_uid == os.getuid() and st.st_gid == os.getgid(), path


def copy(source, target):
    owned(source)
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(source, target)
    target.chmod(0o644)
    assert digest(source) == digest(target), target


def record(path, data):
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')


def build(version, revision, art_dir=None):
    assert Path.cwd() == ROOT and os.getuid() == 1000
    assert re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}', version)
    owned(ROOT)
    destination = ROOT/'dist-all'/version
    assert not destination.exists(), '拒絕覆寫版本'
    assert re.fullmatch('[a-z0-9]+', revision)
    stage = ROOT/'workplace'/('package92-'+version+('-'+revision if revision != 'a' else ''))
    assert not stage.exists(), '拒絕覆寫建置批次'
    stage.mkdir()
    source = stage/'source'
    source.mkdir()
    for folder in ('cmd', 'internal'):
        shutil.copytree(ROOT/folder, source/folder)
    for name in ('go.mod', 'go.sum', 'LICENSE'):
        copy(ROOT/name, source/name)
    lock = 'docs/internals/data/fm-import-lock.json'
    copy(ROOT/lock, source/lock)
    env = dict(os.environ, GOPROXY='off', GOSUMDB='off', GOWORK='off')
    subprocess.run(['go', 'mod', 'vendor'], cwd=source, env=env, check=True)
    source_files = {str(p.relative_to(source)): {'bytes': p.stat().st_size, 'sha256': digest(p)}
                    for p in sorted(source.rglob('*')) if p.is_file()}
    full = destination/'full-local'
    full.mkdir(parents=True)
    for name in ('promo', 'smoke'):
        (destination/name).mkdir()
    head = subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip()
    dirty = subprocess.check_output(['git', 'status', '--porcelain'], text=True)
    metadata = {'version': version, 'classification': '本機完整版測試版，含原版素材；禁止公開散布',
                'head': head, 'working_tree': dirty,
                'go_version': subprocess.check_output(['go', 'version'], text=True).strip(),
                'build_image': 'logh3-ebiten:2', 'source_files': source_files,
                'ai_scope': '原版／強化戰略AI；戰術模式以試玩說明及使用者裁決為準',
                'not_claimed': ['原版戰場同態', '人耳驗收', 'Windows實機', 'macOS', '完整登陸分段演出']}
    common = stage/'common'
    common.mkdir()
    graphics_args = ''
    hd_count = 0
    if art_dir is not None:
        art = Path(art_dir)
        assert art.is_dir() and (art/'face').is_dir()
        hd_count = len(list((art/'face').glob('face-*.png')))
        assert hd_count > 0
        for p in sorted(art.rglob('*')):
            if p.is_file():
                copy(p, common/'art'/p.relative_to(art))
        graphics_args = ' -graphics hd -art "art"'
        metadata.update(graphics='hd', hd_portraits=hd_count)
    # 原始資料只讀；本機完整版明確納入合法持有的資料，不作公開包。
    for p in sorted(Path('/orig').rglob('*')):
        if p.is_file():
            copy(p, common/'original'/p.relative_to('/orig'))
    names = set(re.findall(r'"([^"/]+\.json)"', (ROOT/'internal/bootstrap/load.go').read_text()))
    # commandphase依公開畫面名字選規格。
    names.update(p.name for p in (ROOT/'docs/internals/data').glob('screen-*.json'))
    for name in sorted(names):
        copy(ROOT/'docs/internals/data'/name, common/'data'/name)
    library = ROOT/'workplace/round80-library'
    receipt = json.loads((library/'LIBRARY-SOURCES.json').read_text())
    assert len(receipt['files']) == 50
    for entry in receipt['files']:
        p = library/entry['file']
        assert p.stat().st_size == entry['bytes'] and digest(p) == entry['sha256'], p
        copy(p, common/'music'/p.name)
    copy(library/'LIBRARY-SOURCES.json', common/'music/LIBRARY-SOURCES.json')
    sf = json.loads((ROOT/'docs/internals/data/player-music-library.json').read_text())['soundfont']
    prepared = ROOT/'workplace/music69-source/prepared'
    assert (prepared/sf['file']).stat().st_size == sf['bytes'] and digest(prepared/sf['file']) == sf['sha256']
    for name in (sf['file'], 'COPYING', 'README'):
        copy(prepared/name, common/'soundfont'/name)
    copy(ROOT/'docs/internals/data/player-music-library.json', common/'soundfont/SOURCE.json')
    copy(ROOT/'LICENSE', common/'LICENSE')
    (common/'licenses').mkdir(exist_ok=True)
    (common/'licenses/Go-LICENSE').write_bytes(Path('/usr/local/go/LICENSE').read_bytes())
    for p in (source/'vendor').rglob('*'):
        if p.is_file() and (p.name.lower().startswith(('license', 'notice', 'copying'))):
            copy(p, common/'licenses'/p.relative_to(source/'vendor'))
    (common/'RIGHTS.txt').write_text('本包僅供原版合法持有人本機試玩。原版素材及其衍生音訊的權利仍屬原權利人。\n'
                                  '引擎為RRSAL-1.0；第三方與FluidR3_GM授權全文見licenses與soundfont/COPYING。\n'
                                  '音色庫已隨包；播放使用預先合成的OGG，不需另裝合成器或下載素材。\n')
    commands = []
    packages = []
    for platform in ('windows', 'linux'):
        folder = full/('LOGH3-'+version+'-'+platform+'-amd64')
        shutil.copytree(common, folder)
        executable = folder/('gin3.exe' if platform == 'windows' else 'gin3')
        command = ['go', 'build', '-mod=vendor', '-trimpath', '-buildvcs=false',
                   '-ldflags=-X main.buildVersion='+version, '-o', str(executable), './cmd/gin3']
        subprocess.run(command, cwd=source, env=dict(env, GOOS=platform, GOARCH='amd64',
                                                   CGO_ENABLED='0' if platform == 'windows' else '1'), check=True)
        commands.append({'platform': platform, 'command': command, 'binary_sha256': digest(executable)})
        if platform == 'linux':
            executable.chmod(0o755)
        for name, level in [('original', -1), ('enhanced', 1)]:
            if platform == 'windows':
                launcher = folder/('start-'+name+'.cmd')
                launcher.write_bytes(('@echo off\r\ncd /d "%~dp0"\r\n'
                    'gin3.exe -orig "original" -data "data" -music "music" -ai '+str(level)+' -scale 1.5'+graphics_args+'\r\n'
                    'if errorlevel 1 pause\r\n').encode('ascii'))
            else:
                launcher = folder/('start-'+name+'.sh')
                launcher.write_text('#!/bin/sh\nset -eu\ncd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"\n'
                    'exec ./gin3 -orig original -data data -music music -ai '+str(level)+' -scale 1.5'+graphics_args+' "$@"\n')
                launcher.chmod(0o755)
        launch = 'start-original.cmd／start-enhanced.cmd' if platform == 'windows' else './start-original.sh／./start-enhanced.sh'
        readme = ('銀河英雄傳說III SP remake '+version+' 本機測試版\n\n'
                  '解壓整個資料夾，執行 '+launch+'。原版模式與強化模式共用存檔；強化預設難度1，命令列-ai可選0至3。\n'
                  '預設播放片頭；主選單可以開始新局或讀S0至S3。防拷題可用138進入。右鍵離開面板。\n'
                  '遊戲為640×480，啟動視窗960×720，可拖曳改大小。\n'
                  '配樂22首、音效28段已包含。預設片頭為GM，遊戲FM／GM與曲目由設定選擇。主選單安靜是正常行為。\n'
                  'Windows存檔在%APPDATA%\\logh3\\save；Linux在$XDG_CONFIG_HOME/logh3/save或~/.config/logh3/save。\n'
                  'Linux需X11或XWayland、OpenGL與ALSA；Windows需可用OpenGL顯示驅動。\n\n'
                  '本版為可試玩的實作快照，尚未完成人耳、Windows實機及完整登陸分段演出驗收。\n'
                  '雙玩家命令與戰場原版差異不再作本次交付閘門；未宣稱全遊戲逐像素或逐位元組一致。\n'
                  '請測：新局與S3載入、艦隊指令與回合、兩种AI差異、存檔退出再讀回、FM／GM換曲及戰鬥音效。\n'
                  '回報請附版本、平台、AI模式、存檔、操作步驟及畫面；聲音問題另附曲目。\n\n'
                  '戰術AI政策：原版與強化模式均已啟用，電腦會下令、移動及造成傷害。\n'
                  '原版戰術的目標參考點仍有強推論，不宣稱完整原版AI逐步相同。\n'
                  '含原版素材與衍生音訊，僅限本機合法持有人測試，不上傳或公開散布。\n')
        if hd_count:
            readme = readme.replace('遊戲為640×480，啟動視窗960×720，可拖曳改大小。',
                'HD顯示為1920×1440，啟動視窗960×720，可拖曳改大小。\n'
                '本版包含'+str(hd_count)+'張原畫風HD人物；其他圖片與字型沿用原圖，尚未完成整套HD美術。\n'
                '原版與強化AI啟動器都使用HD圖片；遊戲規則與存檔保持。')
        if platform == 'windows':
            readme = readme.replace('\n', '\r\n')
        (folder/'README.txt').write_bytes(b'\xef\xbb\xbf'+readme.encode())
        record(folder/'VERSION.json', {'version': version, 'platform': platform, 'architecture': 'amd64',
                                      'head': head, 'source_files': source_files})
        files = {str(p.relative_to(folder)): {'bytes': p.stat().st_size, 'sha256': digest(p)}
                 for p in sorted(folder.rglob('*')) if p.is_file()}
        record(folder/'MANIFEST.json', {'version': version, 'classification': metadata['classification'], 'files': files})
        archive = full/(folder.name+'.zip')
        with zipfile.ZipFile(archive, 'x', compression=zipfile.ZIP_DEFLATED, compresslevel=6) as z:
            for p in sorted(folder.rglob('*')):
                if p.is_file():
                    z.write(p, str(Path(folder.name)/p.relative_to(folder)))
        with zipfile.ZipFile(archive) as z:
            assert z.testzip() is None
            assert z.read(folder.name+'/README.txt').startswith(b'\xef\xbb\xbf')
            for item in z.infolist():
                assert item.filename.isascii() or item.flag_bits & 0x800
            for name, info in files.items():
                data = z.read(folder.name+'/'+name)
                assert len(data) == info['bytes'] and hashlib.sha256(data).hexdigest() == info['sha256']
        packages.append({'path': str(archive.relative_to(destination)), 'bytes': archive.stat().st_size, 'sha256': digest(archive)})
        print('封包逐檔核對通過', platform, archive.stat().st_size, flush=True)
    source_archive = full/('LOGH3-'+version+'-engine-source.tar.gz')
    with tarfile.open(source_archive, 'x:gz') as tar:
        tar.add(source, arcname='logh3-source')
    packages.append({'path': str(source_archive.relative_to(destination)), 'bytes': source_archive.stat().st_size, 'sha256': digest(source_archive)})
    metadata.update(build_commands=commands, packages=packages, music_tracks=22, sound_effects=28, runtime_json=len(names))
    record(destination/'BUILD.json', metadata)
    record(destination/'SHA256SUMS.json', {'version': version, 'files': packages})
    print('完成本機完整版', destination, flush=True)


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--version', required=True)
    p.add_argument('--revision', default='a')
    p.add_argument('--art', help='可選HD圖片目錄，只加入本機完整版')
    args = p.parse_args()
    build(args.version, args.revision, args.art)
