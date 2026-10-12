"""v1：由已核對的本機完整版建立不含原版素材的引擎包。於GUI Docker執行。"""
import argparse
import hashlib
import json
import os
import re
from pathlib import Path
import shutil
import zipfile

ROOT = Path('/src')


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()


def record(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def main(version):
    assert re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}', version)
    root = ROOT / 'dist-all' / version
    build = json.loads((root / 'BUILD.json').read_text())
    assert build['version'] == version and build['full_directory'] == 'full'
    assert len([p for p in build['packages'] if p['path'].endswith('.zip')]) == 3
    output = root / 'release'
    output.mkdir()
    assert output.stat().st_uid == os.getuid()
    packages = []
    for platform in ('linux-amd64', 'windows-amd64', 'macos-universal'):
        name = f'LOGH3-{version}-{platform}'
        source = root / 'full' / name
        target = output / name
        if platform == 'macos-universal':
            contents = source / '銀河英雄傳說III.app/Contents'
            destination = target / '銀河英雄傳說III.app/Contents'
            destination.mkdir(parents=True)
            shutil.copyfile(contents / 'Info.plist', destination / 'Info.plist')
            shutil.copytree(contents / 'MacOS', destination / 'MacOS')
            assets_source = contents / 'Resources'
            assets_target = destination / 'Resources'
            assets_target.mkdir()
            for mode in ('original', 'enhanced'):
                shutil.copyfile(source / f'start-{mode}.command', target / f'start-{mode}.command')
                (target / f'start-{mode}.command').chmod(0o755)
            launcher = destination / 'MacOS/launch'
            launcher.write_text(launcher.read_text().replace('-graphics hd -art "$LOGH3_RES_DIR/art" ', '').replace('-music "$LOGH3_RES_DIR/music"', '-music ""'))
            launcher.chmod(0o755)
        else:
            target.mkdir()
            assets_source, assets_target = source, target
            binary = 'gin3.exe' if platform == 'windows-amd64' else 'gin3'
            shutil.copyfile(source / binary, target / binary)
            (target / binary).chmod(0o755)
            for mode, level in [('original', -1), ('enhanced', 1)]:
                if platform == 'windows-amd64':
                    text = ('@echo off\r\ncd /d "%~dp0"\r\n'
                            f'gin3.exe -orig "original" -data "data" -music "" -ai {level} -scale 1.5 %*\r\n'
                            'if errorlevel 1 pause\r\n')
                    (target / f'start-{mode}.cmd').write_bytes(text.encode('ascii'))
                else:
                    launcher = target / f'start-{mode}.sh'
                    launcher.write_text('#!/bin/sh\nset -eu\ncd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"\n'
                                        f'exec ./gin3 -orig original -data data -music "" -ai {level} -scale 1.5 "$@"\n')
                    launcher.chmod(0o755)
        for folder in ('data', 'licenses'):
            shutil.copytree(assets_source / folder, assets_target / folder)
        shutil.copyfile(assets_source / 'LICENSE', assets_target / 'LICENSE')
        shutil.copyfile(source / 'VERSION.json', target / 'VERSION.json')
        readme = (f'銀河英雄傳說III SP remake {version} 引擎發行包\n\n'
                  '本包不含原版遊戲、字型、音樂、音效、人物圖或HD衍生素材。引擎授權RRSAL-1.0，第三方授權見licenses。\n'
                  '先取得自己合法持有的PC DOS中文版原版資料。解開原版ZIP後，找到同時包含GIN3.COM、GRAPH.IMG、FACE.AMG的目錄。\n'
                  'Windows／Linux：在本包的gin3旁建立original資料夾，把該原版目錄內的檔案複製進去，再啟動start-original或start-enhanced。\n'
                  'macOS：對銀河英雄傳說III.app按右鍵，顯示套件內容；在Contents/Resources建立original資料夾，把原版目錄內的檔案複製進去。再開.app或start-enhanced.command。\n'
                  '這是本機資料匯入；不需要Docker、不上傳遊戲檔，不修改原版來源。\n'
                  '也可用啟動器的-orig參數指定既有資料夾；Linux例：./start-original.sh -orig "/path/to/game"。\n'
                  '防拷畫面輸入138後可點確認或按Enter。兩種AI共用存檔，原圖模式為預設。\n'
                  '音訊為選配。若已有自己匯出的22首配樂與28段音效OGG，可加-music "/path/to/music"；未匯入時保持安靜。\n'
                  'Windows存檔：%APPDATA%\\logh3\\save；Linux：$XDG_CONFIG_HOME/logh3/save或~/.config/logh3/save；macOS：~/Library/Application Support/logh3/save。\n'
                  '滑鼠移到頂端顯示功能列；F1說明、F7原版／HD。顯示敵軍請到系統指令的設定選單。\n'
                  'Linux需X11或XWayland、OpenGL與ALSA；Windows需OpenGL驅動；macOS最低11.0，支援Intel及Apple Silicon。\n'
                  'macOS未經Developer ID簽署或Apple公證。Windows／macOS真機與人耳仍待驗收；完整登陸分段演出尚未完成。\n'
                  '本版本是功能發行，不宣稱全遊戲原版逐像素、逐位元組或流程完全一致。\n'
                  '可另外以-graphics hd -art "/path/to/art"指定自己持有的HD素材資料夾。\n')
        if platform == 'windows-amd64':
            readme = readme.replace('\n', '\r\n')
        (target / 'README.txt').write_bytes(b'\xef\xbb\xbf' + readme.encode())
        (target / 'RIGHTS.txt').write_text('引擎及重建版面資料依RRSAL-1.0。原版及HD衍生素材均未包含，須由合法持有人在本機提供。\n')
        files = {str(p.relative_to(target)): {'bytes': p.stat().st_size, 'sha256': sha(p)}
                 for p in sorted(target.rglob('*')) if p.is_file()}
        forbidden = {'.amg', '.img', '.bmg', '.com', '.mds', '.ogg', '.sf2', '.dat', '.bit', '.rgn', '.spr'}
        assert not any(Path(p).suffix.lower() in forbidden or 'art/face/' in p for p in files)
        record(target / 'MANIFEST.json', {'version': version, 'classification': '引擎發行包；無原版或HD衍生素材', 'files': files})
        archive = output / (name + '.zip')
        with zipfile.ZipFile(archive, 'x', compression=zipfile.ZIP_DEFLATED, compresslevel=6) as z:
            for p in sorted(target.rglob('*')):
                if p.is_file():
                    z.write(p, str(Path(name) / p.relative_to(target)))
        with zipfile.ZipFile(archive) as z:
            assert z.testzip() is None
            assert set(z.namelist()) == {name + '/MANIFEST.json', *(name + '/' + p for p in files)}
            for path, expected in files.items():
                data = z.read(name + '/' + path)
                assert len(data) == expected['bytes'] and hashlib.sha256(data).hexdigest() == expected['sha256']
        packages.append({'path': str(archive.relative_to(root)), 'bytes': archive.stat().st_size, 'sha256': sha(archive)})
        print('PASS 引擎發行包', platform, archive.stat().st_size, flush=True)
    source = root / 'full' / f'LOGH3-{version}-engine-source.tar.gz'
    target = output / source.name
    shutil.copyfile(source, target)
    assert sha(source) == sha(target)
    packages.append({'path': str(target.relative_to(root)), 'bytes': target.stat().st_size, 'sha256': sha(target)})
    build['packages'].extend(packages)
    build['release_directory'] = 'release'
    build['not_claimed'] = ['原版全遊戲parity', 'Windows／macOS真機', '人耳驗收', '完整登陸分段演出']
    record(root / 'BUILD.json', build)
    record(root / 'SHA256SUMS.json', {'version': version, 'files': build['packages']})


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--version', required=True)
    main(parser.parse_args().version)
