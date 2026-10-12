"""本次尚未發布的包更正美術說明，保留舊ZIP並從完整檔案集合重封。"""
import json
import os
from pathlib import Path
import shutil
import zipfile
from local_test_package import digest, hd_readme, record

VERSION = 'v.1.0.0-20261010'
root = Path('/src/dist-all') / VERSION
assert os.getuid() == 1000
build = json.loads((root / 'BUILD.json').read_text())
assert build['version'] == VERSION
backup = Path('/src/workplace/release100-20261010/before-art-notes')
backup.mkdir()
notes = hd_readme(VERSION, 83, 1)
corrections = []
for platform in ('linux-amd64', 'windows-amd64', 'macos-universal'):
    name = f'LOGH3-{VERSION}-{platform}'
    folder = root / 'full' / name
    relative = '銀河英雄傳說III.app/Contents/Resources/art/README.md' if platform.startswith('macos') else 'art/README.md'
    path = folder / relative
    assert '目前首批 5 張' in path.read_text() and path.stat().st_uid == 1000
    binary = folder / ('銀河英雄傳說III.app/Contents/MacOS/gin3' if platform.startswith('macos') else 'gin3.exe' if platform.startswith('windows') else 'gin3')
    binary_hash = digest(binary)
    manifest = json.loads((folder / 'MANIFEST.json').read_text())
    original = manifest['files'][relative]
    archive = root / 'full' / (name + '.zip')
    old_hash = digest(archive)
    shutil.move(archive, backup / archive.name)
    (backup / (platform + '-README.md')).write_bytes(path.read_bytes())
    path.write_text(notes)
    original.update(bytes=path.stat().st_size, sha256=digest(path))
    record(folder / 'MANIFEST.json', manifest)
    with zipfile.ZipFile(archive, 'x', compression=zipfile.ZIP_DEFLATED, compresslevel=6) as z:
        for p in sorted(folder.rglob('*')):
            if p.is_file():
                z.write(p, str(Path(name) / p.relative_to(folder)))
    with zipfile.ZipFile(archive) as z:
        assert z.testzip() is None
        assert set(z.namelist()) == {name + '/MANIFEST.json', *(name + '/' + f for f in manifest['files'])}
        import hashlib
        for f, expected in manifest['files'].items():
            data = z.read(name + '/' + f)
            assert len(data) == expected['bytes'] and hashlib.sha256(data).hexdigest() == expected['sha256']
            if 'mode' in expected:
                assert ((z.getinfo(name + '/' + f).external_attr >> 16) & 0o777) == int(expected['mode'], 8)
    assert digest(binary) == binary_hash
    result = {'path': str(archive.relative_to(root)), 'bytes': archive.stat().st_size, 'sha256': digest(archive)}
    build['packages'] = [result if p['path'] == result['path'] else p for p in build['packages']]
    if platform.startswith('macos'):
        build['macos']['archive'] = result
        record(root / 'smoke/macos/MACOS.json', build['macos'])
    corrections.append({'platform': platform, 'old_archive_sha256': old_hash, 'new_archive': result,
                        'binary_sha256_unchanged': binary_hash, 'note_sha256': digest(path)})
    print('PASS 更正美術說明與ZIP逐檔', platform, flush=True)
common = Path('/src/workplace/package92-' + VERSION + '/common/art/README.md')
common.write_text(notes)
build['documentation_correction'] = corrections
record(root / 'BUILD.json', build)
record(root / 'SHA256SUMS.json', {'version': VERSION, 'files': build['packages']})
