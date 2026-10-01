"""規格 073：依固定鎖取得／離線重驗 Linux 音樂工具的完整 Debian 來源。

重生入口：docs/playtest/033-linux-music-sources.md。
--lock-from-probe 只供明示審查的新建鎖，不是下載失敗時的替代路徑。
"""
import argparse
import concurrent.futures
import hashlib
import json
import os
from pathlib import Path
import re
import tempfile
import tarfile
import urllib.request
from urllib.parse import quote, urlsplit

ROOT = Path('/work/workplace/music69-linux')
LOCK = Path('/work/docs/internals/data/linux-music-source-lock.json')
MAX_FILE = 512 * 1024 * 1024
MAX_TOTAL = 2 * 1024 * 1024 * 1024


def digest(data):
    return hashlib.sha256(data).hexdigest()


def owned(path, directory=False):
    s = path.lstat()
    if path.is_symlink() or (s.st_uid, s.st_gid) != (os.getuid(), os.getgid()):
        raise ValueError(f'連結或擁有權錯誤：{path}')
    if not (path.is_dir() if directory else path.is_file()):
        raise ValueError(f'檔案形態錯誤：{path}')


def basename(name):
    if not re.fullmatch(r'[A-Za-z0-9][A-Za-z0-9_.+~\-]*', name) or name in ('.', '..'):
        raise ValueError(f'來源檔名不安全：{name}')
    return name


def parse_dsc(data, source, version):
    if len(data) > 2 * 1024 * 1024:
        raise ValueError('描述超過上限')
    lines = data.decode('utf-8').splitlines()
    if lines and lines[0] == '-----BEGIN PGP SIGNED MESSAGE-----':
        if '' not in lines:
            raise ValueError('PGP 包裝缺少標頭分隔')
        lines = lines[lines.index('') + 1:]
    fields = {}
    current = None
    ended = False
    for line in lines:
        if line == '-----BEGIN PGP SIGNATURE-----':
            break
        if line.startswith('- '):
            line = line[2:]
        if not line:
            ended = True
            continue
        if ended:
            raise ValueError('描述含多個段落')
        if line[0] in ' \t':
            if current is None:
                raise ValueError('續行缺少欄位')
            fields[current] += '\n' + line.strip()
            continue
        m = re.fullmatch(r'([A-Za-z0-9][A-Za-z0-9-]*):\s*(.*)', line)
        if not m or m[1].lower() in fields:
            raise ValueError('欄位格式錯誤或重複')
        current = m[1].lower()
        fields[current] = m[2]
    if fields.get('source') != source or fields.get('version') != version:
        raise ValueError('來源身份與固定版本不符')
    parts = {}
    for line in fields.get('checksums-sha256', '').splitlines():
        if not line:
            continue
        m = re.fullmatch(r'([0-9a-f]{64}) ([0-9]+) (\S+)', line)
        if not m:
            raise ValueError('Checksums-Sha256 格式錯誤')
        name = basename(m[3])
        size = int(m[2])
        if name in parts or not 0 < size <= MAX_FILE:
            raise ValueError('來源檔碰撞或大小錯誤')
        parts[name] = {'sha256': m[1], 'bytes': size}
    if not parts:
        raise ValueError('描述缺少來源 SHA-256 清單')
    return parts


def safe_url(url):
    p = urlsplit(url)
    if p.scheme != 'https' or p.netloc != 'deb.debian.org' or p.query or p.fragment:
        raise ValueError('來源 URL 不在固定 Debian HTTPS 範圍')
    if not (p.path.startswith('/debian/pool/main/') or p.path.startswith('/debian-security/pool/updates/main/')):
        raise ValueError('來源 URL 路徑錯誤')
    return url


def packages(build):
    return {name: {'source': p['source_package'], 'version': p['source_version']}
            for name, p in build['packages'].items()}


def create_lock():
    if LOCK.exists():
        raise ValueError('拒絕覆寫已審查固定鎖')
    build_path = ROOT/'封裝 測試/BUILD.json'
    owned(build_path)
    data = build_path.read_bytes()
    mapping = packages(json.loads(data))
    pairs = {(p['source'], p['version']) for p in mapping.values()}
    probe = ROOT/'source-index/all-probe.json'
    owned(probe)
    rows = json.loads(probe.read_text())
    if {(r['source'], r['version']) for r in rows} != pairs or len(rows) != len(pairs):
        raise ValueError('描述探針沒有完整覆蓋實際套件')
    entries = []
    for row in rows:
        name = basename(row['filename'])
        path = ROOT/'source-index'/name
        owned(path)
        raw = path.read_bytes()
        if len(raw) != row['bytes'] or digest(raw) != row['sha256']:
            raise ValueError('描述探針檔案變動')
        parts = parse_dsc(raw, row['source'], row['version'])
        base = safe_url(row['url']).rsplit('/', 1)[0]
        entries.append({'source': row['source'], 'version': row['version'],
                        'description': {k: row[k] for k in ('filename', 'url', 'sha256', 'bytes')},
                        'files': {n: dict(v, url=base+'/'+quote(n)) for n, v in parts.items()}})
    lock = {'schema': 1, 'build_sha256': digest(data), 'packages': mapping,
            'sources': entries, 'signature_trust_verified': False}
    files_for(lock)
    owned(LOCK.parent, directory=True)
    LOCK.write_text(json.dumps(lock, ensure_ascii=False, indent=2, sort_keys=True)+'\n')


def files_for(lock):
    if lock.get('schema') != 1:
        raise ValueError('固定鎖版本錯誤')
    result = {}
    pairs = set()
    for row in lock['sources']:
        pair = (row['source'], row['version'])
        if pair in pairs:
            raise ValueError('來源身份重複')
        pairs.add(pair)
        records = [row['description']] + [dict(v, filename=n) for n, v in row['files'].items()]
        for r in records:
            name = basename(r['filename'])
            safe_url(r['url'])
            if not re.fullmatch(r'[0-9a-f]{64}', r['sha256']) or not 0 < r['bytes'] <= MAX_FILE:
                raise ValueError('固定雜湊或大小錯誤')
            if name in result and any(result[name][k] != r[k] for k in ('sha256', 'bytes')):
                raise ValueError('來源檔名碰撞')
            result[name] = r
    if pairs != {(p['source'], p['version']) for p in lock['packages'].values()}:
        raise ValueError('來源集合未覆蓋二進位套件')
    if sum(r['bytes'] for r in result.values()) > MAX_TOTAL:
        raise ValueError('完整來源集合超過容量上限')
    return result


def verify_file(path, record):
    owned(path)
    if path.stat().st_size != record['bytes']:
        raise ValueError(f'快取大小不符：{path.name}')
    h = hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda: f.read(1024*1024), b''):
            h.update(block)
    if h.hexdigest() != record['sha256']:
        raise ValueError(f'快取 SHA-256 不符：{path.name}')


def obtain(cache, record, download):
    path = cache/record['filename']
    if path.exists() or path.is_symlink():
        verify_file(path, record)
        return
    if not download:
        raise ValueError(f'離線快取缺檔：{path.name}')
    fd, temp = tempfile.mkstemp(prefix='.part-', dir=cache)
    try:
        with os.fdopen(fd, 'wb') as out:
            with urllib.request.urlopen(safe_url(record['url']), timeout=45) as response:
                safe_url(response.geturl())
                count = 0
                while block := response.read(1024*1024):
                    count += len(block)
                    if count > record['bytes']:
                        raise ValueError('下載超出固定大小')
                    out.write(block)
        verify_file(Path(temp), record)
        os.replace(temp, path)
    finally:
        if Path(temp).exists():
            Path(temp).unlink()


def prepare(cache, download=False):
    owned(LOCK)
    raw = LOCK.read_bytes()
    lock = json.loads(raw)
    build = ROOT/'封裝 測試/BUILD.json'
    owned(build)
    if digest(build.read_bytes()) != lock['build_sha256'] or packages(json.loads(build.read_bytes())) != lock['packages']:
        raise ValueError('固定鎖與實際工具包不符')
    records = files_for(lock)
    owned(cache.parent, directory=True)
    cache.mkdir(exist_ok=True)
    owned(cache, directory=True)
    # 開始網路前先拒絕所有已存在的損壞快取。
    for name, r in records.items():
        p = cache/name
        if p.exists() or p.is_symlink():
            verify_file(p, r)
    with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
        futures = [pool.submit(obtain, cache, r, download) for r in records.values()]
        for i, future in enumerate(futures, 1):
            future.result()
            if i % 40 == 0:
                print(f'來源檔驗證 {i}/{len(records)}', flush=True)
    for row in lock['sources']:
        parts = parse_dsc((cache/row['description']['filename']).read_bytes(), row['source'], row['version'])
        expected = {n: {k: r[k] for k in ('sha256', 'bytes')} for n, r in row['files'].items()}
        if parts != expected:
            raise ValueError('描述與固定來源集合不符')
    receipt = {'status': 'complete-fixed-source-cache', 'lock_sha256': digest(raw),
               'build_sha256': lock['build_sha256'], 'source_packages': len(lock['sources']),
               'binary_packages': len(lock['packages']), 'files': records,
               'total_bytes': sum(r['bytes'] for r in records.values()),
               'signature_trust_verified': False}
    dest = cache/'VERIFIED.json'
    if dest.exists():
        owned(dest)
    dest.write_text(json.dumps(receipt, ensure_ascii=False, indent=2, sort_keys=True)+'\n')
    print(json.dumps({k: v for k, v in receipt.items() if k != 'files'}, ensure_ascii=False))
    return receipt


def bundle():
    cache = ROOT/'source-cache'
    receipt = prepare(cache)
    package = ROOT/'封裝 測試'
    build = json.loads((package/'BUILD.json').read_text())
    # 來源對照必須綁住仍符合原收據的二進位實物。
    for name, record in build['files'].items():
        verify_file(package/name, record)
    out = ROOT/'source-delivery'
    owned(out.parent, directory=True)
    out.mkdir(exist_ok=True)
    owned(out, directory=True)
    target = out/'linux-music-debian-sources.tar'
    if target.exists() or target.is_symlink():
        raise ValueError('拒絕覆蓋既有來源封包')
    inputs = {f'sources/{n}': cache/n for n in receipt['files']}
    inputs['SOURCE-LOCK.json'] = LOCK
    inputs['VERIFIED.json'] = cache/'VERIFIED.json'
    inputs['TOOL-BUILD.json'] = package/'BUILD.json'
    for p in sorted((package/'licenses').rglob('*')):
        if p.is_file() and p.suffix != '.cpp' and p.name != 'Dockerfile.audio':
            inputs[str(p.relative_to(package))] = p
    for name in ['Dockerfile.audio', 'opl2wav.cpp']:
        inputs['build/'+name] = Path('/work/tools')/name
    readme = ('此封包保存 Linux 音樂工具 198 個 Debian 二進位套件對應的 157 個同版本來源集合。\n'
              'sources/ 包含完整 dsc、上游來源及 Debian 修補／建置資料；SOURCE-LOCK.json 列明 URL、版本與 SHA-256。\n'
              'licenses/ 保存原套件著作權與授權；build/ 保存本專案 opl2wav 原始碼與 audio image 配方。\n'
              '沒有原版遊戲檔、音樂、SF2、執行檔或共享庫。各來源授權依各自文件，不以 remake 授權取代。\n'
              '此階段已驗來源檔完整與工具身份對照，尚未逐套重建；Debian 簽章信任鏈也未驗。\n'
              '這是本機待交付來源集合，不是正式版號或全部跨平台發行義務完成聲明。\n')
    fd, temp = tempfile.mkstemp(prefix='.bundle-', dir=out)
    try:
        with os.fdopen(fd, 'wb') as stream, tarfile.open(fileobj=stream, mode='w', format=tarfile.PAX_FORMAT) as archive:
            import io
            content = readme.encode('utf-8')
            info = tarfile.TarInfo('README.txt')
            info.size, info.mode, info.mtime = len(content), 0o644, 0
            archive.addfile(info, io.BytesIO(content))
            for name, path in sorted(inputs.items()):
                owned(path)
                info = tarfile.TarInfo(name)
                info.size, info.mode, info.mtime = path.stat().st_size, 0o644, 0
                with path.open('rb') as f:
                    archive.addfile(info, f)
        os.replace(temp, target)
    finally:
        if Path(temp).exists():
            Path(temp).unlink()
    # 不解包；直接重讀每個來源 entry 並驗 SHA-256，拒絕額外項或連結。
    observed = set()
    with tarfile.open(target, mode='r:') as archive:
        for member in archive:
            if not member.isfile() or member.name in observed:
                raise ValueError('來源封包含重複項或非普通檔案')
            observed.add(member.name)
            expected = inputs.get(member.name)
            if member.name == 'README.txt':
                if archive.extractfile(member).read() != readme.encode('utf-8'):
                    raise ValueError('來源說明不符')
                continue
            if expected is None or member.size != expected.stat().st_size:
                raise ValueError('來源封包項目／大小不符')
            with archive.extractfile(member) as f, expected.open('rb') as original:
                if hashlib.file_digest(f, 'sha256').digest() != hashlib.file_digest(original, 'sha256').digest():
                    raise ValueError('來源封包內容不符')
    if observed != set(inputs) | {'README.txt'}:
        raise ValueError('來源封包缺項')
    with target.open('rb') as f:
        archive_sha = hashlib.file_digest(f, 'sha256').hexdigest()
    result = {'schema': 'logh3.music-tools.linux.sources.v1', 'date': '2026-10-01',
              'status': '完整固定 Debian 來源集合取得、離線重驗及本機封裝通過',
              'lock_sha256': receipt['lock_sha256'], 'build_sha256': receipt['build_sha256'],
              'source_packages': receipt['source_packages'], 'binary_packages': receipt['binary_packages'],
              'source_files': len(receipt['files']), 'source_bytes': receipt['total_bytes'],
              'archive': {'path': 'workplace/music69-linux/source-delivery/'+target.name,
                          'sha256': archive_sha, 'bytes': target.stat().st_size, 'entries': len(observed)},
              'scripts_sha256': {n: digest((Path('/work/tools/scripts')/n).read_bytes())
                                 for n in ['music_linux_sources.py', 'music_linux_sources_test.py']},
              'validation_image': 'sha256:229a2c5bfa27522db7815ea81f9bed70af17ccb9de9fc7ad142b1877b5830d36',
              'signature_trust_verified': False, 'rebuilt_all_sources': False,
              'formal_delivery_complete': False}
    dest = Path('/work/docs/internals/data/linux-music-sources.json')
    owned(dest.parent, directory=True)
    if dest.exists():
        owned(dest)
    dest.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(result, ensure_ascii=False))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--lock-from-probe', action='store_true')
    parser.add_argument('--download', action='store_true')
    parser.add_argument('--bundle', action='store_true')
    args = parser.parse_args()
    if args.lock_from_probe:
        if args.download or args.bundle:
            parser.error('新建固定鎖與下載分開執行')
        create_lock()
    elif args.bundle:
        if args.download:
            parser.error('來源下載與離線封裝分開執行')
        bundle()
    else:
        prepare(ROOT/'source-cache', args.download)
