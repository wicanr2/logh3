"""以 apt 已驗證的套件索引取得精確 URI，有限並行下載並核對雜湊。"""
import concurrent.futures
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import shutil
import time
import urllib.request
import urllib.parse
import sys

PACKAGES = ['chromium=154.0.8037.92-1~deb12u1', 'fonts-noto-cjk=1:20220127+repack1-1']
empty_archives = Path('/tmp/logh3-hd-preview-empty-archives')
empty_archives.mkdir(parents=True, exist_ok=True)
listing = subprocess.check_output([
    'apt-get', '-o', f'Dir::Cache::archives={empty_archives}',
    '--print-uris', '--yes', '--download-only', '--no-install-recommends',
    'install', *PACKAGES
], text=True)
archives = Path('/var/cache/apt/archives')
archives.mkdir(parents=True, exist_ok=True)
chunk_size = 2 * 1024 * 1024
jobs = []
packages = []
unparsed = []
for line in listing.splitlines():
    match = re.fullmatch(r"'([^']+)' (\S+) (\d+)(?: (\w+):([0-9a-fA-F]+))?\s*", line)
    if not match:
        if line.startswith("'"):
            unparsed.append(line)
        continue
    uri, filename, size, algorithm, checksum = match.groups()
    size = int(size)
    path = archives / urllib.parse.unquote(filename)
    package_name, version, architecture = path.name.rsplit('_', 2)
    records = subprocess.check_output(['apt-cache', 'show', f'{package_name}={version}'], text=True)
    candidates = []
    for block in records.split('\n\n'):
        fields = dict(row.split(': ', 1) for row in block.splitlines() if ': ' in row and not row.startswith(' '))
        if fields.get('Version') == version and fields.get('Size') == str(size) and fields.get('SHA256'):
            candidates.append(fields)
    assert candidates and len({row['SHA256'] for row in candidates}) == 1, (package_name, version)
    algorithm, checksum = 'SHA256', candidates[0]['SHA256']
    packages.append((path, size, algorithm, checksum))
    hash_name = algorithm.lower().replace('sum', '')
    if path.exists() and path.stat().st_size == size:
        if hashlib.new(hash_name, path.read_bytes()).hexdigest() == checksum.lower():
            continue
    path.write_bytes(b'')
    with path.open('r+b') as f:
        f.truncate(size)
    for start in range(0, size, chunk_size):
        jobs.append((uri.replace('http://', 'https://', 1), path, start, min(start+chunk_size, size)-1, size))


def download(job):
    uri, path, start, end, total = job
    for attempt in range(3):
        try:
            request = urllib.request.Request(uri, headers={'Range': f'bytes={start}-{end}'})
            with urllib.request.urlopen(request, timeout=25) as response:
                if response.status != 206:
                    # Small files may be returned in full by the CDN.
                    if not (start == 0 and end == total-1 and response.status == 200):
                        raise ValueError(f'Range ignored: {path.name} {response.status}')
                data = response.read(end-start+2)
            if len(data) != end-start+1:
                raise ValueError(f'Length mismatch: {path.name}')
            fd = os.open(path, os.O_WRONLY)
            try:
                os.pwrite(fd, data, start)
            finally:
                os.close(fd)
            return
        except Exception:
            if attempt == 2:
                raise
            time.sleep(1)


assert not unparsed, f'Unsupported apt URI records: {unparsed!r}'
print(f'{len(packages)} archives, {len(jobs)} bounded segments', flush=True)
assert packages, 'apt did not provide the complete installation manifest'
with concurrent.futures.ThreadPoolExecutor(max_workers=8) as pool:
    pending = [pool.submit(download, job) for job in jobs]
    for count, result in enumerate(concurrent.futures.as_completed(pending), 1):
        result.result()
        print(f'Downloaded {count}/{len(jobs)} segments', flush=True)
for path, size, algorithm, checksum in packages:
    actual = hashlib.new(algorithm.lower().replace('sum', ''), path.read_bytes()).hexdigest()
    assert actual == checksum.lower(), (path.name, actual, checksum)
    print(f'Verified {path.name}: {size} bytes', flush=True)
Path('/opt/debian-downloads-receipt.json').write_text(json.dumps([
    {'filename': p.name, 'bytes': size, 'apt_hash_algorithm': algo, 'apt_hash': digest,
     'sha256': hashlib.sha256(p.read_bytes()).hexdigest()}
    for p, size, algo, digest in packages
], indent=2))
assert '--download-only' in sys.argv, '安裝由Dockerfile的離線apt排序處理'
