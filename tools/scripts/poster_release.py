"""150：同一凍結來源重建含三組海報的v1.6.0三平台完整版。Docker內執行。"""
import argparse
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import zipfile

import v140_platform_delivery as package
from local_test_macos_package import universal

ROOT = Path('/src'); VERSION = 'v.1.6.0-20261024'; OUT = ROOT/'dist-all'/VERSION
STAGE = ROOT/'workplace'/('package92-'+VERSION+'-release150')
MAC = ROOT/'workplace/macos-release150'
BASE = ROOT/'dist-all/v.1.5.0-20261019/full/LOGH3-v.1.5.0-20261019-windows-amd64'
sha = package.sha; write = package.write
def read(path): return json.loads(Path(path).read_text())


def notes(platform, full=True):
    text = f'''銀河英雄傳說III SP {VERSION} {platform} 本機完整版

已含原版資料、HD圖像、配樂、音效、音色庫及三組海報。請完整解壓後使用。
海報預設關閉。在系統指令→設定右上切換：關閉→v1→v2→v3→關閉。
v1左萊因哈特、右楊威利，底部斜體名言；v2左右兩軍將領；v3左右雙方親友。
中央保持完整4:3，海報開啟後外框16:9；F7切換原版／HD，兩側不接受遊戲操作。
設定左上Lang按鈕切換繁體中文、简体中文、日本語、English、한국어。
介面與v1名言共用語系；語系、海報組別及開關保存於preferences.json，重啟保留。
可用-locale zh-TW|zh-CN|ja|en|ko、-poster=true及-poster-set v1|v2|v3明確指定。
譯文唯一來源為data/ui-locales.tsv，可修改後重啟。圖片內商標、故事圖及未收錄原文不冒稱已翻譯。
海報依官方經典動畫參考生成，名言其他語言為本專案翻譯；來源及提示見art/poster。
前線駐留後再出擊從提督所在地領船、退船及編成；改善參謀／副官選人視窗延遲。
「非戰艦旗艦」預設關閉。開啟後B／H／C／A／D可當旗艦，修護艦E與運補艦T仍排除。
擊破轉移另須開啟「轉移旗艦」，成功率沿提督設定；不增加艦艇或彈藥。
編輯提督可調十三項能力、四項特殊能力；選項不改原版六表格式。
原版AI／強化AI啟動器均保留，強化預設難度1。F1說明，F7原版／HD，防拷題138。
不附玩家偏好或存檔，原使用者進度不覆蓋。戰術讀回重建仍由#48追蹤。
完整版含原版與衍生素材，限合法持有人本機使用，不公開散布。
引擎RRSAL-1.0；第三方、字型及素材來源見licenses、music、soundfont及art/poster。
'''
    if platform == 'windows':
        text += '\n雙擊gin3.exe，或start-original.cmd／start-enhanced.cmd。\n保留original、data、art、music、soundfont在同一目錄。\n存檔：%APPDATA%\\logh3\\save。Windows採Wine驗證，未真機。\n'
    elif platform == 'linux':
        text += '\n執行./gin3或兩個.sh啟動器，需X11／XWayland、OpenGL與ALSA。\n存檔：$XDG_CONFIG_HOME/logh3/save，未設定時為~/.config/logh3/save。\n解壓若未保留權限，將gin3與.sh設為可執行。\n'
    else:
        text += '\nIntel與Apple Silicon通用，最低macOS 11。開啟.app或兩個.command。\n存檔：~/Library/Application Support/logh3/save。\n紀錄：~/Library/Logs/logh3/game.log。\n已核對架構、系統依賴與ARM簽章，未真機；無Developer ID或Apple公證。\n'
    return b'\xef\xbb\xbf'+(text.replace('\n', '\r\n') if platform == 'windows' else text).encode()


def source_check():
    inputs = read(STAGE/'INPUTS.json')
    for name, row in inputs['source_files'].items():
        assert sha(STAGE/'source'/name) == row['sha256'], name
        if not name.startswith('vendor/'): assert sha(ROOT/name) == row['sha256'], name
    for name, row in inputs['common_files'].items(): assert sha(STAGE/'common'/name) == row['sha256'], name
    return inputs


def prepare():
    assert os.getuid() == os.getgid() == 1000
    assert not STAGE.exists() and not OUT.exists(), '拒絕覆寫既有版號或批次'
    assert (ROOT/'workplace').stat().st_uid == (ROOT/'dist-all').stat().st_uid == os.getuid()
    assert shutil.disk_usage(ROOT).free > 8*1024**3
    package.verify(BASE)
    STAGE.mkdir(); source = STAGE/'source'; source.mkdir()
    for name in ['cmd', 'internal']: shutil.copytree(ROOT/name, source/name)
    for name in ['go.mod', 'go.sum', 'LICENSE']: shutil.copyfile(ROOT/name, source/name)
    lock = source/'docs/internals/data/fm-import-lock.json'; lock.parent.mkdir(parents=True)
    shutil.copyfile(ROOT/'docs/internals/data/fm-import-lock.json', lock)
    subprocess.run(['go', 'mod', 'vendor'], cwd=source, env=dict(os.environ, GOPROXY='off', GOSUMDB='off', GOWORK='off'), check=True, timeout=180)
    common = STAGE/'common'; common.mkdir()
    for name in ['original', 'art', 'music', 'soundfont', 'licenses']: shutil.copytree(BASE/name, common/name)
    shutil.copyfile(ROOT/'LICENSE', common/'LICENSE')
    for name, target in [('internal/font/settings_font.LICENSE.txt', 'licenses/Noto-Settings.txt'), ('internal/localize/LICENSE.font.txt', 'licenses/Noto-Locales.txt')]: shutil.copyfile(ROOT/name, common/target)
    poster = common/'art/poster'; assert not poster.exists(); shutil.copytree(ROOT/'assets/hd/poster', poster)
    poster_notes = '三組海報及來源\nv1：左萊因哈特、右楊威利。v2：兩軍將領。v3：雙方親友。\n在遊戲設定右上切換；左上Lang切五語，v1底部名言共用語系。\nSOURCES.json保存v1；SETS.SOURCE.json保存v2、v3完整提示及官方參考。\nQUOTE-SOURCES.json保存名言出處及翻譯分類。僅本機使用，不公開散布。\n'
    (poster/'README.txt').write_bytes(b'\xef\xbb\xbf'+poster_notes.replace('\n', '\r\n').encode())
    data = common/'data'; data.mkdir()
    names = set(re.findall(r'"([^"/]+\.json)"', (source/'internal/bootstrap/load.go').read_text()))
    names.update(f.name for f in (ROOT/'docs/internals/data').glob('screen-*.json'))
    for name in sorted(names): shutil.copyfile(ROOT/'docs/internals/data'/name, data/name)
    shutil.copyfile(ROOT/'internal/localize/ui.tsv', data/'ui-locales.tsv')
    inputs = dict(version=VERSION, source_files=package.inventory(source), common_files=package.inventory(common),
                  base=str(BASE.relative_to(ROOT)), base_manifest_sha256=sha(BASE/'MANIFEST.json'),
                  head=subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
                  working_tree=subprocess.check_output(['git', 'status', '--porcelain'], text=True),
                  generators={str(f.relative_to(ROOT)): sha(f) for f in [Path(__file__).resolve(), ROOT/'tools/scripts/v140_platform_delivery.py', ROOT/'tools/scripts/local_test_macos_build.sh', ROOT/'tools/macos.sh']})
    write(STAGE/'INPUTS.json', inputs); (OUT/'full').mkdir(parents=True)
    print('READY', len(inputs['source_files']), '凍結來源，', len(inputs['common_files']), '共用檔案', flush=True)


def build(platform):
    source_check()
    target = STAGE/('gin3.exe' if platform == 'windows' else 'gin3'); assert not target.exists()
    command = ['go', 'build', '-mod=vendor', '-trimpath', '-buildvcs=false', '-ldflags=-X main.buildVersion='+VERSION, '-o', str(target), './cmd/gin3']
    subprocess.run(command, cwd=STAGE/'source', env=dict(os.environ, GOPROXY='off', GOSUMDB='off', GOWORK='off', GOOS=platform, GOARCH='amd64', CGO_ENABLED='0' if platform == 'windows' else '1'), check=True, timeout=600)
    if platform == 'linux': folder = package.unix_folder('linux-amd64', True)
    else:
        folder = OUT/'full'/('LOGH3-'+VERSION+'-windows-amd64'); assert not folder.exists(); shutil.copytree(STAGE/'common', folder)
        shutil.copyfile(target, folder/'gin3.exe'); (folder/'README.txt').write_bytes(notes('windows')); write(folder/'VERSION.json', package.metadata('windows-amd64'))
        for mode, level in [('original', -1), ('enhanced', 1)]:
            quality = ' -music-quality hd' if mode == 'enhanced' else ''
            (folder/('start-'+mode+'.cmd')).write_bytes(('@echo off\r\ncd /d "%~dp0"\r\ngin3.exe -ai '+str(level)+' -scale 1.5'+quality+' %*\r\nif errorlevel 1 pause\r\n').encode('ascii'))
    row = package.archive(folder, '本機完整版，含原版及衍生素材，不公開')
    write(STAGE/(platform.upper()+'.json'), dict(version=VERSION, package=row, binary_sha256=sha(target), command=command, go_version=subprocess.check_output(['go', 'version'], text=True).strip()))


def macos():
    source_check(); slices = universal(MAC/'gin3-universal')
    row = package.archive(package.unix_folder('macos-universal', True), '本機完整版，含原版及衍生素材，不公開')
    write(STAGE/'MACOS.json', dict(version=VERSION, package=row, slices=slices, binary_sha256=sha(MAC/'gin3-universal'), toolchain=(MAC/'TOOLCHAIN.txt').read_text(), runtime='未真機'))


def extract(platform):
    parent = ROOT/'workplace/poster150/package-linux' if platform == 'linux' else ROOT/'workplace/music69-windows/poster150/installed'
    assert not parent.exists(); parent.mkdir(parents=True)
    file = OUT/'full'/('LOGH3-'+VERSION+('-linux-amd64.zip' if platform == 'linux' else '-windows-amd64.zip'))
    with zipfile.ZipFile(file) as archive:
        assert archive.testzip() is None
        for entry in archive.infolist():
            target = parent/entry.filename; assert target.resolve().is_relative_to(parent.resolve())
            archive.extract(entry, parent); target.chmod((entry.external_attr >> 16)&0o777)
    package.verify(parent/file.stem)
    print('PASS 實際ZIP解壓與manifest', parent, flush=True)


def configure():
    package.VERSION, package.OUT, package.STAGE, package.MACOS, package.readme = VERSION, OUT, STAGE, MAC, notes


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__); p.add_argument('mode', choices=['prepare', 'windows', 'linux', 'macos', 'extract-linux', 'extract-windows'])
    args = p.parse_args(); configure()
    if args.mode in ['windows', 'linux']: build(args.mode)
    elif args.mode.startswith('extract-'): extract(args.mode.split('-')[1])
    else: globals()[args.mode]()
