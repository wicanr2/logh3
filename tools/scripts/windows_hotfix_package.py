"""從受驗完整版素材與目前程式乾淨建立 Windows 修正版；僅在 GUI Docker 內執行。"""
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


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def record(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')


def inventory(root):
    return {str(p.relative_to(root)): dict(bytes=p.stat().st_size, sha256=sha(p))
            for p in sorted(root.rglob('*')) if p.is_file()}


def archive(folder, version, classification):
    files = inventory(folder)
    record(folder/'MANIFEST.json', dict(version=version, classification=classification, files=files))
    target = folder.parent/(folder.name+'.zip')
    assert not target.exists()
    with zipfile.ZipFile(target, 'x', compression=zipfile.ZIP_DEFLATED, compresslevel=6) as z:
        for p in sorted(folder.rglob('*')):
            if p.is_file():
                z.write(p, str(Path(folder.name)/p.relative_to(folder)))
    with zipfile.ZipFile(target) as z:
        assert z.testzip() is None
        for name, expected in files.items():
            b = z.read(folder.name+'/'+name)
            assert len(b) == expected['bytes'] and hashlib.sha256(b).hexdigest() == expected['sha256']
    return dict(path=str(target.relative_to(target.parents[1])), bytes=target.stat().st_size, sha256=sha(target))


def main(version, base_version):
    assert Path.cwd() == ROOT and os.getuid() == os.getgid() == 1000
    for v in [version, base_version]:
        assert re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}', v)
    base = ROOT/'dist-all'/base_version/'full'/f'LOGH3-{base_version}-windows-amd64'
    assert base.is_dir()
    original_manifest = json.loads((base/'MANIFEST.json').read_text())
    for name, info in original_manifest['files'].items():
        assert sha(base/name) == info['sha256'], name
    output = ROOT/'dist-all'/version
    batch = ROOT/'workplace'/('windows-package-'+version)
    assert not output.exists() and not batch.exists(), '拒絕覆寫既有版本／批次'
    assert ROOT.stat().st_uid == os.getuid() and (ROOT/'dist-all').stat().st_uid == os.getuid()
    batch.mkdir()
    source = batch/'source'
    source.mkdir()
    for name in ['cmd', 'internal']:
        shutil.copytree(ROOT/name, source/name)
    for name in ['go.mod', 'go.sum', 'LICENSE']:
        shutil.copyfile(ROOT/name, source/name)
    lock = source/'docs/internals/data/fm-import-lock.json'
    lock.parent.mkdir(parents=True)
    shutil.copyfile(ROOT/'docs/internals/data/fm-import-lock.json', lock)
    env = dict(os.environ, GOPROXY='off', GOSUMDB='off', GOWORK='off')
    subprocess.run(['go', 'mod', 'vendor'], cwd=source, env=env, check=True)
    source_files = inventory(source)
    for kind in ['full', 'release', 'promo', 'smoke']:
        (output/kind).mkdir(parents=True)
    full = output/'full'/f'LOGH3-{version}-windows-amd64'
    full.mkdir()
    for name in ['original', 'art', 'licenses', 'soundfont']:
        shutil.copytree(base/name, full/name)
    shutil.copyfile(ROOT/'LICENSE', full/'LICENSE')
    shutil.copytree(ROOT/'workplace/round132/music', full/'music')
    library = json.loads((full/'music/LIBRARY-SOURCES.json').read_text())
    hd = json.loads((full/'music/hd/MANIFEST.json').read_text())
    mt32 = json.loads((full/'music/mt32/SOURCES.json').read_text())
    for row in library['files']:
        assert sha(full/'music'/row['file']) == row['sha256']
    for row in hd['files']:
        assert sha(full/'music/hd'/row['file']) == row['sha256']
    for row in mt32['tracks']:
        assert sha(full/'music/mt32'/(row['name']+'.ogg')) == row['ogg_sha256']
    data = full/'data'
    data.mkdir()
    names = set(re.findall(r'"([^"/]+\.json)"', (source/'internal/bootstrap/load.go').read_text()))
    names.update(p.name for p in (ROOT/'docs/internals/data').glob('screen-*.json'))
    for name in sorted(names):
        shutil.copyfile(ROOT/'docs/internals/data'/name, data/name)
    binary = full/'gin3.exe'
    command = ['go', 'build', '-mod=vendor', '-trimpath', '-buildvcs=false',
               '-ldflags=-X main.buildVersion='+version, '-o', str(binary), './cmd/gin3']
    subprocess.run(command, cwd=source, env=dict(env, GOOS='windows', GOARCH='amd64', CGO_ENABLED='0'), check=True)
    text = (f'銀河英雄傳說III SP {version}\n\n'
            '完整解壓資料夾後，可直接雙擊 gin3.exe。程式自動讀取執行檔旁的 original、data、art 與 music。\n'
            '原版規則使用 start-original.cmd；強化 AI 使用 start-enhanced.cmd。\n'
            '系統設定可選 GeneralUser GS 或 Roland MT-32，選擇會保存在玩家存檔目錄的 preferences.json。\n'
            '轉移旗艦與名將移動加成預設開啟，可在系統設定各自關閉；關閉與開啟會保存在偏好檔。\n'
            '旗艦被擊毀後，謬拉必定轉移，其他提督約25%；須有存活戰艦或高速戰艦，轉往艘數最多的分隊。\n'
            '名將移動加成只在強化模式啟用，楊威利／萊因哈特參戰時全陣營新移動額度1.5倍；關閉仍保留名將戰術AI。\n'
            'F1 說明、F7 原版／HD；防拷題輸入 138。存檔在 %APPDATA%\\logh3\\save。\n'
            '本包含原版及衍生素材，僅供合法持有人本機使用，不公開散布。\n'
            'Windows Wine 已驗證；Windows 真機待使用者驗收。戰術讀回重建仍由 #48 追蹤。\n')
    (full/'README.txt').write_bytes(b'\xef\xbb\xbf'+text.replace('\n','\r\n').encode())
    (full/'RIGHTS.txt').write_text('僅本機完整包，含原版及衍生素材，禁止公開散布。引擎 RRSAL-1.0，第三方見 licenses、soundfont 與 music 內聲明。\n')
    metadata = dict(version=version, platform='windows', architecture='amd64',
                    head=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
                    working_tree=subprocess.check_output(['git','status','--porcelain'],text=True),
                    source_files=source_files, build_command=command,
                    go_version=subprocess.check_output(['go','version'],text=True).strip(),
                    base_material_version=base_version, base_manifest_sha256=sha(base/'MANIFEST.json'),
                    music_sources_sha256={name:sha(full/'music'/name) for name in ['LIBRARY-SOURCES.json','hd/MANIFEST.json','mt32/SOURCES.json']})
    engine = output/'release'/full.name
    engine.mkdir()
    shutil.copyfile(binary, engine/'gin3.exe')
    shutil.copytree(data, engine/'data')
    shutil.copytree(full/'licenses', engine/'licenses')
    shutil.copyfile(ROOT/'LICENSE', engine/'LICENSE')
    (engine/'README.txt').write_bytes(b'\xef\xbb\xbf'+(f'銀河英雄傳說III SP {version} 引擎包\r\n'
        '不含原版、HD 或音訊素材。請將合法取得的原版資料放入 gin3.exe 旁的 original 資料夾。\r\n'
        '該資料夾須含 GRAPH.IMG、GIN3.COM、FACE.AMG 及其他原版檔案。然後雙擊 gin3.exe 或啟動器。\r\n'
        '可另放入自有 art 與 music；程式會依封包位置尋找，亦接受明確的命令列路徑。\r\n').encode())
    packages = []
    for folder, classification in [(full,'含原版與衍生素材，僅本機完整包'), (engine,'引擎包，不含原版／HD／音訊素材')]:
        record(folder/'VERSION.json', metadata)
        for mode, level in [('original',-1),('enhanced',1)]:
            options = ' -music-quality hd' if mode == 'enhanced' and folder == full else ''
            launcher = ('@echo off\r\ncd /d "%~dp0"\r\ngin3.exe -ai '+str(level)+' -scale 1.5'+options+' %*\r\nif errorlevel 1 pause\r\n')
            (folder/('start-'+mode+'.cmd')).write_bytes(launcher.encode('ascii'))
        if folder == engine:
            assert not any(p.is_file() and p.suffix.lower() in {'.com','.img','.amg','.ogg','.sf2','.rom','.dat'} for p in folder.rglob('*'))
        packages.append(archive(folder,version,classification))
    source_archive = output/'release'/f'LOGH3-{version}-engine-source.tar.gz'
    with tarfile.open(source_archive,'x:gz') as tar:
        tar.add(source,arcname='logh3-source')
    packages.append(dict(path=str(source_archive.relative_to(output)),bytes=source_archive.stat().st_size,sha256=sha(source_archive)))
    record(output/'BUILD.json',metadata)
    record(output/'SHA256SUMS.json',dict(version=version,files=packages))
    print(json.dumps(packages,ensure_ascii=False,indent=2),flush=True)


if __name__ == '__main__':
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--version',required=True)
    p.add_argument('--base-version',default='v.1.3.0-20261013')
    a=p.parse_args()
    main(a.version,a.base_version)
