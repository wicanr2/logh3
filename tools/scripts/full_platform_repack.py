"""147：沿既有跨平台封包器，補建v1.5.2三平台本機完整版，不覆寫修正包或影片。Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import zipfile

import v140_platform_delivery as package
from local_test_macos_package import universal

ROOT = Path('/src')
VERSION = 'v.1.5.2-20261024'
OUT = ROOT/'dist-all'/VERSION
STAGE = ROOT/'workplace'/('package92-'+VERSION+'-release147')
MAC = ROOT/'workplace/macos-release147'
BASE = ROOT/'dist-all/v.1.5.0-20261019/full/LOGH3-v.1.5.0-20261019-windows-amd64'


def read(path): return json.loads(Path(path).read_text())
def write(path, value): Path(path).write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')
def sha(path): return package.sha(Path(path))


def notes(platform, full=True):
    text = f'''銀河英雄傳說III SP {VERSION} {platform} 本機完整版

已含原版資料、HD圖像、配樂、音效及音色庫。完整解壓後使用，不用另裝修正包。
修正前線駐留後再出擊：從提督所在地領船、退船及編成，母星庫存保持。
改善參謀及副官選人視窗重複繪製造成的延遲。
系統設定新增「非戰艦旗艦」，預設關閉，開啟後編成及擊破轉移可選B／H／C／A／D。
修護艦E與運補艦T始終不能當旗艦。轉移仍需開啟「轉移旗艦」，成功率沿提督設定。
新開關保存於preferences.json，重啟保留；原版六表存檔格式保持。
原版AI與強化AI啟動器均保留；強化預設難度1。F1說明，F7切換原版／HD，防拷題138。
系統設定可調AI、顯示敵軍、字型及音色，並可編輯十三項提督能力與四項特殊能力。
現有存檔仍沿各平台使用者目錄。新版不附偏好或玩家存檔，不覆蓋原有進度。
戰術保存返回已驗，正式戰術讀回重建仍由#48追蹤，不宣稱DOS整場同態。
完整版含原版及衍生素材，只限合法持有人本機使用，不公開散布。
引擎依RRSAL-1.0，第三方與素材來源見licenses、music及soundfont內聲明。
'''
    if platform == 'windows':
        text += '''
直接雙擊gin3.exe，或start-original.cmd／start-enhanced.cmd。
保留同一資料夾的original、data、art、music、soundfont；程式自動從執行檔旁尋找。
存檔：%APPDATA%\\logh3\\save。
Windows使用Wine驗證，未作Windows真機驗收。
'''
    elif platform == 'linux':
        text += '''
執行./gin3、./start-original.sh或./start-enhanced.sh，需要可用的圖形與音訊環境。
存檔：$XDG_CONFIG_HOME/logh3/save；未設定時為~/.config/logh3/save。
解壓工具若未保留權限，將gin3與兩個.sh設為可執行。
'''
    else:
        text += '''
Intel與Apple Silicon共用通用程式，最低macOS 11。
開啟銀河英雄傳說III.app，或start-original.command／start-enhanced.command。
存檔：~/Library/Application Support/logh3/save。
錯誤紀錄：~/Library/Logs/logh3/game.log。
已檢查雙架構、系統依賴及ARM程式頁簽章，未在macOS真機執行。
本機包沒有Developer ID或Apple公證，不變更系統安全設定。
'''
    return b'\xef\xbb\xbf'+(text.replace('\n', '\r\n') if platform == 'windows' else text).encode()


def source_check():
    inputs = read(STAGE/'INPUTS.json')
    for name, row in inputs['source_files'].items():
        assert sha(STAGE/'source'/name) == row['sha256'], name
        if not name.startswith('vendor/'): assert sha(ROOT/name) == row['sha256'], name
    for name, row in inputs['preserved'].items(): assert sha(OUT/name) == row, name
    return inputs


def prepare():
    assert os.getuid() == os.getgid() == 1000
    assert not STAGE.exists() and not (OUT/'full').exists()
    assert STAGE.parent.stat().st_uid == OUT.stat().st_uid == os.getuid()
    assert shutil.disk_usage(ROOT).free > 6*1024**3, '需至少6GiB工作空間'
    package.verify(BASE)
    patch = read(OUT/'BUILD.json')
    frozen = ROOT/'workplace'/('personnel-patch-'+VERSION)/'source'
    for name, digest in patch['source_files'].items():
        assert sha(frozen/name) == digest, name
        if not name.startswith('vendor/'): assert sha(ROOT/name) == digest, name
    STAGE.mkdir()
    shutil.copytree(frozen, STAGE/'source')
    common = STAGE/'common'; common.mkdir()
    for name in ['original', 'art', 'music', 'soundfont', 'licenses']:
        shutil.copytree(BASE/name, common/name)
    shutil.copyfile(ROOT/'LICENSE', common/'LICENSE')
    shutil.copyfile(ROOT/'internal/font/settings_font.LICENSE.txt', common/'licenses/Noto-Settings.txt')
    data = common/'data'; data.mkdir()
    names = set(re.findall(r'"([^"/]+\.json)"', (STAGE/'source/internal/bootstrap/load.go').read_text()))
    names.update(p.name for p in (ROOT/'docs/internals/data').glob('screen-*.json'))
    for name in sorted(names): shutil.copyfile(ROOT/'docs/internals/data'/name, data/name)
    preserved = {str(p.relative_to(OUT)): sha(p) for p in [OUT/'BUILD.json', OUT/patch['archive']['path'], *sorted((OUT/'promo').glob('*'))] if p.is_file()}
    shutil.copyfile(OUT/'SHA256SUMS.txt', STAGE/'previous-SHA256SUMS.txt')
    inputs = dict(version=VERSION, source_files=package.inventory(STAGE/'source'),
                  common_files=package.inventory(common), base=str(BASE.relative_to(ROOT)),
                  base_manifest_sha256=sha(BASE/'MANIFEST.json'), patch_build_sha256=sha(OUT/'BUILD.json'),
                  preserved=preserved, previous_checksums_sha256=sha(STAGE/'previous-SHA256SUMS.txt'),
                  head=subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
                  working_tree=subprocess.check_output(['git', 'status', '--porcelain'], text=True),
                  generators={str(p.relative_to(ROOT)): sha(p) for p in [Path(__file__).resolve(), ROOT/'tools/scripts/v140_platform_delivery.py', ROOT/'tools/scripts/local_test_macos_build.sh', ROOT/'tools/macos.sh']})
    write(STAGE/'INPUTS.json', inputs)
    (OUT/'full').mkdir()
    print('READY', len(inputs['source_files']), '凍結來源，共用素材已核對', flush=True)


def build(platform):
    source_check()
    target = STAGE/('gin3.exe' if platform == 'windows' else 'gin3')
    assert not target.exists()
    command = ['go', 'build', '-mod=vendor', '-trimpath', '-buildvcs=false', '-ldflags=-X main.buildVersion='+VERSION, '-o', str(target), './cmd/gin3']
    env = dict(os.environ, GOPROXY='off', GOSUMDB='off', GOWORK='off', GOOS=platform, GOARCH='amd64', CGO_ENABLED='0' if platform == 'windows' else '1')
    subprocess.run(command, cwd=STAGE/'source', env=env, check=True, timeout=600)
    if platform == 'linux': folder = package.unix_folder('linux-amd64', True)
    else:
        folder = OUT/'full'/('LOGH3-'+VERSION+'-windows-amd64')
        assert not folder.exists(); shutil.copytree(STAGE/'common', folder)
        shutil.copyfile(target, folder/'gin3.exe')
        (folder/'README.txt').write_bytes(notes('windows'))
        write(folder/'VERSION.json', package.metadata('windows-amd64'))
        for mode, level in [('original', -1), ('enhanced', 1)]:
            quality = ' -music-quality hd' if mode == 'enhanced' else ''
            text = '@echo off\r\ncd /d "%~dp0"\r\ngin3.exe -ai '+str(level)+' -scale 1.5'+quality+' %*\r\nif errorlevel 1 pause\r\n'
            (folder/('start-'+mode+'.cmd')).write_bytes(text.encode('ascii'))
    row = package.archive(folder, '含原版及衍生素材，僅本機完整版')
    write(STAGE/(platform.upper()+'.json'), dict(version=VERSION, package=row, binary_sha256=sha(target), command=command, go_version=subprocess.check_output(['go', 'version'], text=True).strip()))


def macos():
    source_check()
    slices = universal(MAC/'gin3-universal')
    folder = package.unix_folder('macos-universal', True)
    row = package.archive(folder, '含原版及衍生素材，僅本機完整版')
    write(STAGE/'MACOS.json', dict(version=VERSION, package=row, slices=slices, binary_sha256=sha(MAC/'gin3-universal'), runtime='未在macOS真機執行', toolchain=(MAC/'TOOLCHAIN.txt').read_text()))


def extract():
    """Linux抽驗必須取實際ZIP；保留ZIP中的可執行權限。"""
    parent = STAGE/'unpacked'
    assert not parent.exists(); parent.mkdir()
    file = OUT/'full'/('LOGH3-'+VERSION+'-linux-amd64.zip')
    with zipfile.ZipFile(file) as z:
        assert z.testzip() is None
        for item in z.infolist():
            path = parent/item.filename
            assert path.resolve().is_relative_to(parent.resolve())
            z.extract(item, parent); path.chmod((item.external_attr >> 16) & 0o777)
    package.verify(parent/('LOGH3-'+VERSION+'-linux-amd64'))
    print('PASS Linux實際ZIP解壓、manifest及權限', flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode', choices=['prepare', 'windows', 'linux', 'macos', 'extract'])
    args = parser.parse_args()
    package.VERSION, package.OUT, package.STAGE, package.MACOS, package.readme = VERSION, OUT, STAGE, MAC, notes
    if args.mode in ['windows', 'linux']: build(args.mode)
    else: globals()[args.mode]()
