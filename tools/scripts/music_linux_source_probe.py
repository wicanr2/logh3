"""規格 072 的三個主要工具固定來源描述探針；不是完整 157 套件來源交付器。

入口見 docs/playtest/032-linux-music-tools.md，網路只在建置容器明示開放。
"""
from pathlib import Path
import os
import urllib.request
import urllib.error
import json
import hashlib
from urllib.parse import quote

work = Path('/work/workplace/music69-linux')
build = json.loads((work/'封裝 測試/BUILD.json').read_text())
sources = {(x['source_package'], x['source_version']) for x in build['packages'].values()}
out = work/'source-index'
out.mkdir(exist_ok=True)
assert not out.is_symlink() and (out.stat().st_uid, out.stat().st_gid) == (os.getuid(), os.getgid())
result = {}
for package, version in sorted(sources):
    if package not in ('ffmpeg', 'fluidsynth', 'adplug'):
        continue
    prefix = package[:4] if package.startswith('lib') else package[0]
    name = f'{package}_{version.split(":")[-1]}.dsc'
    dest = out/name
    if dest.exists():
        assert not dest.is_symlink() and (dest.stat().st_uid, dest.stat().st_gid) == (os.getuid(), os.getgid())
    candidates = [f'https://deb.debian.org/debian/pool/main/{prefix}/{package}/{quote(name)}', f'https://deb.debian.org/debian-security/pool/updates/main/{prefix}/{package}/{quote(name)}']
    errors = []
    for url in candidates:
        try:
            with urllib.request.urlopen(url, timeout=15) as response:
                data = response.read(2*1024*1024)
            assert len(data) < 2*1024*1024
            text = data.decode()
            assert f'Version: {version}\n' in text
            dest.write_bytes(data)
            result[package] = {'version': version, 'url': url, 'sha256': hashlib.sha256(data).hexdigest(), 'bytes': len(data)}
            print(package, version, '固定版本 dsc 已取得', flush=True)
            break
        except Exception as error:
            errors.append(str(error))
    else:
        result[package] = {'version': version, 'errors': errors}
        print(package, '未取得', errors, flush=True)
(out/'probe.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
assert len(result) == 3 and all('sha256' in x for x in result.values()), '主要工具來源描述尚未齊全'
