"""研究 363：Windows 音樂試建來源的固定下載與離線重驗。

入口 tools/py.sh /work/tools/scripts/music_windows_sources.py [--download]。
不替換正式工具；來源簽章信任鏈、逐套重建及正式交付另行驗收。
"""
import argparse
import concurrent.futures
import json
import hashlib
import io
from pathlib import Path
import shutil
import tempfile
import os
import tarfile
import urllib.request

import music_linux_sources as common

ROOT = Path('/work/workplace/music69-windows')
LOCK = Path('/work/docs/internals/data/windows-music-source-lock.json')
LINUX = Path('/work/workplace/music69-linux/source-cache')


def extra_url(row):
    if row['source'] == 'proxy-libintl':
        import re
        if not re.fullmatch(r'[0-9a-f]{40}', row['commit']):
            raise ValueError('fallback commit 不完整')
        expected = 'https://codeload.github.com/frida/proxy-libintl/tar.gz/'+row['commit']
    elif row['source'] == 'zlib-upstream' and row['version'] == '1.2.13':
        expected = 'https://zlib.net/fossils/zlib-1.2.13.tar.gz'
    else:
        raise ValueError('額外來源不在已審範圍')
    if row['url'] != expected:
        raise ValueError('額外來源 URL 不符')
    return expected


def records_for(lock):
    if lock.get('schema') != 'logh3.music-tools.windows.source-lock.v1':
        raise ValueError('Windows 來源鎖版本錯誤')
    rows = lock['sources']
    records = common.files_for({'schema': 1, 'sources': rows,
        'packages': {r['source']: {'source': r['source'], 'version': r['version']} for r in rows}})
    extras = {}
    for row in lock['extra_sources']:
        name = common.basename(row['filename'])
        extra_url(row)
        import re
        if not re.fullmatch(r'[0-9a-f]{64}', row['sha256']) or not 0 < row['bytes'] <= common.MAX_FILE:
            raise ValueError('額外來源雜湊／大小錯誤')
        if name in records or name in extras:
            raise ValueError('額外來源檔名碰撞')
        extras[name] = row
    if len(extras) != 2 or {r['source'] for r in extras.values()} != {'proxy-libintl', 'zlib-upstream'}:
        raise ValueError('額外來源未完整覆蓋試建依賴')
    if sum(r['bytes'] for r in (records | extras).values()) > common.MAX_TOTAL:
        raise ValueError('來源集合超過容量上限')
    return records, extras


def obtain_extra(cache, row, download):
    target = cache/row['filename']
    if target.exists() or target.is_symlink():
        common.verify_file(target, row)
        return
    if not download:
        raise ValueError('額外來源缺少離線快取：'+row['filename'])
    fd, temp = tempfile.mkstemp(prefix='.extra-', dir=cache)
    try:
        with os.fdopen(fd, 'wb') as out, urllib.request.urlopen(extra_url(row), timeout=45) as response:
            if response.geturl() != extra_url(row):
                raise ValueError('額外來源重新導向未審查位置')
            total = 0
            while block := response.read(1024*1024):
                total += len(block)
                if total > row['bytes']:
                    raise ValueError('額外來源下載超出固定大小')
                out.write(block)
        common.verify_file(Path(temp), row)
        os.replace(temp, target)
    finally:
        Path(temp).unlink(missing_ok=True)


def prepare(download=False):
    common.owned(LOCK)
    raw = LOCK.read_bytes()
    lock = json.loads(raw)
    records, extras = records_for(lock)
    common.owned(ROOT, directory=True)
    cache = ROOT/'source-cache'
    cache.mkdir(exist_ok=True)
    common.owned(cache, directory=True)
    # 網路前先重驗所有既有檔案，損壞快取不能被下載覆蓋。
    for name, row in (records | extras).items():
        target = cache/name
        if target.exists() or target.is_symlink():
            common.verify_file(target, row)
        else:
            candidate = LINUX/name if name in records else ROOT/name
            if candidate.exists() or candidate.is_symlink():
                common.verify_file(candidate, row)
                shutil.copyfile(candidate, target)
    with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
        futures = [pool.submit(common.obtain, cache, r, download) for r in records.values()]
        futures += [pool.submit(obtain_extra, cache, r, download) for r in extras.values()]
        for future in futures:
            future.result()
    for row in lock['sources']:
        actual = common.parse_dsc((cache/row['description']['filename']).read_bytes(), row['source'], row['version'])
        expected = {n: {k: r[k] for k in ('sha256', 'bytes')} for n, r in row['files'].items()}
        if actual != expected:
            raise ValueError('描述與固定來源集合不符')
    all_records = records | extras
    receipt = {'status': 'complete-fixed-windows-probe-source-cache',
        'lock_sha256': common.digest(raw), 'debian_sources': len(lock['sources']),
        'extra_sources': len(extras), 'files': all_records,
        'total_bytes': sum(r['bytes'] for r in all_records.values()),
        'signature_trust_verified': False, 'formal_delivery_complete': False}
    target = cache/'VERIFIED.json'
    if target.exists() or target.is_symlink():
        common.owned(target)
    target.write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({k: v for k, v in receipt.items() if k != 'files'}, ensure_ascii=False))
    return receipt


def bundle():
    receipt = prepare()
    package = ROOT/'gm-source-utf8-probe/封裝 測試'
    probe = json.loads((package/'PROBE.json').read_text())
    if probe['status'] != 'isolated-gm-source-probe-only':
        raise ValueError('缺隔離試建包收據')
    for row in probe['files'].values():
        common.verify_file(package/'tools'/row['filename'], row)
    for name, expected in probe['licenses_sha256'].items():
        path = package/'licenses'/name
        common.owned(path)
        if common.digest(path.read_bytes()) != expected:
            raise ValueError('授權文件與試驗包收據不符')
    out = ROOT/'source-delivery'
    out.mkdir(exist_ok=True)
    common.owned(out, directory=True)
    target = out/'windows-music-fixed-sources.tar'
    if target.exists() or target.is_symlink():
        raise ValueError('拒絕覆蓋既有來源封包')
    inputs = {'sources/'+n: ROOT/'source-cache'/n for n in receipt['files']}
    inputs |= {'SOURCE-LOCK.json': LOCK, 'VERIFIED.json': ROOT/'source-cache/VERIFIED.json',
               'GM-BUILD.json': ROOT/'gm-build-utf8/BUILD.json',
               'FM-BUILD.json': ROOT/'fm-pcm-build.json', 'GM-PROBE.json': package/'PROBE.json',
               'SOUNDFONT-SOURCE.json': Path('/work/docs/internals/data/player-music-library.json')}
    for name in probe['licenses_sha256']:
        inputs['licenses/'+name] = package/'licenses'/name
    for name in ['Dockerfile.music-cross', 'music_cross.sh', 'opl2wav.cpp']:
        inputs['build/'+name] = Path('/work/tools')/name
    for name in ['music_windows_sources.py', 'music_windows_sources_test.py', 'music_linux_sources.py',
                 'music_windows_gm_probe.py', 'music_windows_gm_package_probe.py',
                 'music_windows_fluidsynth_utf8.patch', 'music_windows_fm_probe.py']:
        inputs['build/'+name] = Path('/work/tools/scripts')/name
    note = ('Windows 音樂工具固定來源試驗封包，正式 dist-all 尚未交付。\n'
            'sources/ 保存 14 個 Debian 同版本來源集合與 2 個額外上游完整來源，雜湊與 URL 見 SOURCE-LOCK.json。\n'
            'build/ 保存本專案建置配方、opl2wav 與明示修改的 FluidSynth Unicode 路徑補丁。\n'
            'licenses/ 保留完整原文；GM-BUILD／FM-BUILD／GM-PROBE 記錄實際工具身份。\n'
            '沒有原版遊戲、衍生曲庫、SF2 或執行檔；已授權音色庫的固定下載與授權入口見 SOUNDFONT-SOURCE.json。\n'
            '第三方來源依各自授權，不以 remake 授權取代；未逐套重建執行期或驗 Debian 簽章信任鏈。\n').encode()
    fd, temp = tempfile.mkstemp(prefix='.bundle-', dir=out)
    try:
        with os.fdopen(fd, 'wb') as stream, tarfile.open(fileobj=stream, mode='w', format=tarfile.PAX_FORMAT) as archive:
            content = tarfile.TarInfo('README.txt')
            content.size, content.mode, content.mtime = len(note), 0o644, 0
            archive.addfile(content, io.BytesIO(note))
            for name, path in sorted(inputs.items()):
                common.owned(path)
                info = tarfile.TarInfo(name)
                info.size, info.mode, info.mtime = path.stat().st_size, 0o644, 0
                with path.open('rb') as source:
                    archive.addfile(info, source)
        os.replace(temp, target)
    finally:
        Path(temp).unlink(missing_ok=True)
    observed = set()
    with tarfile.open(target, 'r:') as archive:
        for entry in archive:
            if not entry.isfile() or entry.name in observed:
                raise ValueError('來源封包含連結、重複項或特殊檔')
            observed.add(entry.name)
            content = archive.extractfile(entry)
            if entry.name == 'README.txt':
                if content.read() != note:
                    raise ValueError('來源說明不符')
            else:
                path = inputs.get(entry.name)
                if path is None or entry.size != path.stat().st_size:
                    raise ValueError('來源封包項目／大小不符')
                with path.open('rb') as source:
                    if hashlib.file_digest(source, 'sha256').digest() != hashlib.file_digest(content, 'sha256').digest():
                        raise ValueError('來源封包雜湊不符')
    if observed != set(inputs) | {'README.txt'}:
        raise ValueError('來源封包缺項')
    with target.open('rb') as archive:
        result = {'scope': '本機 Windows 固定來源試驗封包；正式交付未完成',
            'path': 'workplace/music69-windows/source-delivery/'+target.name,
            'sha256': hashlib.file_digest(archive, 'sha256').hexdigest(),
            'bytes': target.stat().st_size, 'entries': len(observed),
            'lock_sha256': receipt['lock_sha256'], 'formal_delivery_complete': False}
    (out/'BUNDLE.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(result, ensure_ascii=False))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--download', action='store_true')
    parser.add_argument('--bundle', action='store_true')
    args = parser.parse_args()
    if args.download and args.bundle:
        parser.error('下載與離線封裝分開執行')
    bundle() if args.bundle else prepare(args.download)
