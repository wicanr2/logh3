"""規格 076：Windows FM／GM 固定工具合併；只建立本機技術驗證包。

bash tools/music_cross.sh python3 /tools/scripts/music_windows_bundle.py。
預設無網路，拒絕覆蓋；正式版號、完整原始碼與發行驗收另行處理。
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess

SYSTEM = {'kernel32.dll', 'msvcrt.dll', 'advapi32.dll', 'ole32.dll',
          'shell32.dll', 'user32.dll', 'ws2_32.dll', 'bcrypt.dll'}
ROOT = Path('/out')
META = Path('/metadata')


def sha(path):
    if not path.is_file() or path.is_symlink() or (path.stat().st_uid, path.stat().st_gid) != (os.getuid(), os.getgid()):
        raise ValueError('輸入不是本人擁有的普通檔案：'+str(path))
    with path.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def merge_records(groups):
    result, overlaps = {}, []
    for group, records in groups:
        for name, row in records.items():
            if not re.fullmatch(r'[A-Za-z0-9_.+\-]+\.(exe|dll)', name):
                raise ValueError('不接受非工具檔名：'+name)
            key = name.lower()
            if key in result:
                previous = result[key]
                if any(previous[k] != row[k] for k in ['sha256', 'bytes', 'imports']):
                    raise ValueError('同名工具／DLL 內容或引用不同：'+name)
                previous['groups'].append(group)
                overlaps.append(key)
            else:
                result[key] = dict(row, filename=name, groups=[group])
    for name, row in result.items():
        for dll in row['imports']:
            if dll.lower() not in result and dll.lower() not in SYSTEM:
                raise ValueError('未閉合非系統引用：'+name+' → '+dll)
    for name in ['fluidsynth.exe', 'ffmpeg.exe', 'opl2wav.exe', 'dosgolem-probe.exe']:
        if name not in result:
            raise ValueError('缺正式工具：'+name)
    return result, overlaps


def build(utf8=False):
    package = ROOT/('combined-utf8-stage/封裝 測試' if utf8 else 'combined-stage/封裝 測試')
    ancestor = package.parent
    while ancestor != ROOT.parent:
        if ancestor.is_symlink():
            raise ValueError('輸出路徑含連結：'+str(ancestor))
        if ancestor.exists() and (not ancestor.is_dir() or (ancestor.stat().st_uid, ancestor.stat().st_gid) != (os.getuid(), os.getgid())):
            raise ValueError('輸出祖先目錄形態／擁有權不符：'+str(ancestor))
        ancestor = ancestor.parent
    if package.exists() or package.is_symlink():
        raise ValueError('拒絕覆蓋既有合併工具包')
    if (ROOT.stat().st_uid, ROOT.stat().st_gid) != (os.getuid(), os.getgid()):
        raise ValueError('輸出目錄擁有權不符')
    gm_path, fm_path = META/'windows-gm-source-tools.json', META/'windows-fm-pcm.json'
    if sha(gm_path) != '49c07dac366671e95ec5b6589f83efaecf4ff2ac8df88ee0d45023a14cd8dded':
        raise ValueError('GM 技術收據身份改變，先重審規格')
    if sha(fm_path) != '4fc93a59c4a9b1f65b9e1a0114bc919126471b0d13fcef9102fdc29df93bac5a':
        raise ValueError('FM 技術收據身份改變，先重審規格')
    gm, fm = json.loads(gm_path.read_text()), json.loads(fm_path.read_text())
    gm_package = ROOT/'gm-source-utf8-probe/封裝 測試'
    groups = []
    for group, folder, expected in [
        ('gm', gm_package/'tools', {r['filename']: r for r in gm['package']['files'].values()}),
        ('fm', ROOT/'pcm-gate/封裝 測試/tools', {n: r for n, r in fm['package_files'].items() if n != 'music-import.exe'}),
    ]:
        if {p.name for p in folder.iterdir()} != set(expected):
            raise ValueError('工具目錄與固定清單不同：'+group)
        records = {}
        for name, row in expected.items():
            p = folder/name
            if sha(p) != row['sha256'] or p.stat().st_size != row['bytes']:
                raise ValueError('固定工具大小／雜湊不符：'+name)
            output = subprocess.check_output(['x86_64-w64-mingw32-objdump', '-p', str(p)], text=True)
            imports = re.findall(r'DLL Name: (\S+)', output)
            if 'file format pei-x86-64' not in output or imports != row['imports']:
                raise ValueError('工具架構／引用與已驗收據不同：'+name)
            records[name] = dict(sha256=sha(p), bytes=p.stat().st_size, imports=imports,
                                 source=str(p.relative_to(ROOT)))
        groups.append((group, records))
    tools, overlaps = merge_records(groups)
    renderer = None
    if utf8:
        renderer_path = ROOT/'opl-unicode-production/BUILD.json'
        renderer = json.loads(renderer_path.read_text())
        path = ROOT/'opl-unicode-production/opl2wav.exe'
        if sha(path) != renderer['program_sha256'] or sha(Path('/tools/opl2wav.cpp')) != renderer['source_sha256']:
            raise ValueError('Unicode OPL 正式來源與建置身份不符')
        output = subprocess.check_output(['x86_64-w64-mingw32-objdump', '-p', str(path)], text=True)
        if 'file format pei-x86-64' not in output:
            raise ValueError('Unicode OPL 架構不符')
        tools['opl2wav.exe'] = {'filename': path.name, 'sha256': sha(path), 'bytes': path.stat().st_size,
                               'imports': re.findall(r'DLL Name: (\S+)', output),
                               'source': str(path.relative_to(ROOT)), 'groups': ['fm-unicode']}
        for dll in tools['opl2wav.exe']['imports']:
            if dll.lower() not in tools and dll.lower() not in SYSTEM:
                raise ValueError('Unicode OPL 新引用未閉合：'+dll)
    importer = ROOT/'music-import-combined.exe'
    program_hash = sha(importer)
    buildinfo_path = ROOT/'importer-combined-buildinfo.txt'
    buildinfo = buildinfo_path.read_text()
    for text in ['path\tgithub.com/wicanr2/logh3/cmd/music-import', 'vcs.modified=false',
                 'vcs.revision=9a2ce97df1e222a4928dd782ddf210033213a96e', 'GOOS=windows', 'GOARCH=amd64']:
        if text not in buildinfo:
            raise ValueError('原生入口須從指定乾淨提交建置')
    font = gm_package/'soundfont/FluidR3_GM.sf2'
    if sha(font) != gm['package']['soundfont_sha256']:
        raise ValueError('固定音色庫不符')
    source = ROOT/'source-delivery/windows-music-fixed-sources.tar'
    if sha(source) != gm['source_bundle']['sha256'] or source.stat().st_size != gm['source_bundle']['bytes']:
        raise ValueError('完整固定來源 tar 不符')
    inputs = {'music-import.exe': importer, 'soundfont/FluidR3_GM.sf2': font,
              'sources/windows-music-fixed-sources.tar': source,
              'metadata/GM-TOOLS.json': gm_path, 'metadata/FM-TOOLS.json': fm_path,
              'metadata/IMPORTER-BUILDINFO.txt': buildinfo_path}
    inputs |= {'tools/'+r['filename']: ROOT/r['source'] for r in tools.values()}
    if utf8:
        inputs |= {'sources/opl2wav.cpp': Path('/tools/opl2wav.cpp'),
                   'metadata/OPL-UNICODE-BUILD.json': renderer_path,
                   'metadata/OPL-UNICODE-WAVE.json': ROOT/'opl-unicode-probe/WAVE.json'}
    for name, digest in gm['package']['licenses_sha256'].items():
        path = gm_package/'licenses'/name
        if sha(path) != digest:
            raise ValueError('完整授權與固定清單不符：'+name)
        inputs['licenses/'+name] = path
    for name in ['adplug', 'libbinio']:
        folder = ROOT/'src'/name
        for p in [folder/'COPYING', folder/'AUTHORS', folder/'debian/copyright']:
            sha(p)
            inputs['licenses/'+name+'/'+str(p.relative_to(folder))] = p
    # 所有來源先核對，再建立新包；失敗包不能被後續呼叫默默覆蓋。
    hashes = {name: {'sha256': sha(p), 'bytes': p.stat().st_size} for name, p in inputs.items()}
    for name, p in inputs.items():
        target = package/name
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(p, target)
        if sha(target) != hashes[name]['sha256']:
            raise ValueError('複製後雜湊不同：'+name)
    note = ('銀河英雄傳說III SP — Windows 音樂工具技術驗證包\n\n'
            '請在本目錄開啟終端，使用自備的合法原版遊戲目錄：\n'
            '.\\music-import.exe -source "C:\\原版遊戲" -output "C:\\我的曲庫"\n'
            '預設轉換 GM／FM 各十一首，音色庫與工具已包含，無須另外安裝。\n'
            '保留 tools、soundfont、sources、licenses 與 metadata 的相對位置。\n'
            '第三方工具依各自完整授權，來源與本地補丁在 sources 內；不以 remake 授權取代。\n'
            '生成曲庫源自原版，只能留在本機；這個包沒有原版檔案或生成音樂。\n'
            '目前僅完成容器技術準備；不是正式發行包，實機、聆聽與循環仍待驗收。\n')
    if utf8:
        note += ('OPL 工具另修正 Windows Unicode 檔名；更新來源為 sources/opl2wav.cpp，\n'
                 '實際建置與三路波形比對在 metadata/OPL-UNICODE-*.json，舊來源 tar 保留。\n')
    readme = package/'README.txt'
    readme.write_bytes(b'\xef\xbb\xbf'+note.replace('\n', '\r\n').encode('utf-8'))
    hashes['README.txt'] = {'sha256': sha(readme), 'bytes': readme.stat().st_size}
    record = {'schema': 'logh3.music-tools.windows.bundle.v1', 'status': 'staged-for-technical-validation',
              'program_sha256': program_hash, 'source_head': '9a2ce97df1e222a4928dd782ddf210033213a96e',
              'source_dirty': False, 'builder_sha256': sha(Path(__file__)), 'tools': tools,
              'identical_overlaps': overlaps, 'renderer_update': renderer,
              'files': hashes, 'formal_delivery_complete': False}
    (package/'BUNDLE.json').write_text(json.dumps(record, ensure_ascii=False, indent=2)+'\n')
    print('Windows 合併驗證包完成', len(tools), '個工具／DLL，入口 SHA-256', program_hash, flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--utf8', action='store_true')
    build(parser.parse_args().utf8)
