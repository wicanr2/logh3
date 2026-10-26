"""在Docker內核對Windows修正包、凍結來源與三條正常GUI收據。"""
import hashlib
import json
import os
from pathlib import Path
import struct
import zipfile

root = Path('/src')
out = root/'dist-all/v.1.5.1-20261024'
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
build = json.loads((out/'BUILD.json').read_text())
archive = out/build['archive']['path']
assert sha(archive) == build['archive']['sha256']
assert archive.stat().st_size == build['archive']['bytes']
frozen = root/'workplace/personnel-patch-v.1.5.1-20261024/source'
for name,digest in build['source_files'].items():
    assert sha(frozen/name) == digest, name
    if not name.startswith('vendor/'):
        assert sha(root/name) == digest, ('現行產品與封包來源不同',name)
with zipfile.ZipFile(archive) as z:
    assert z.testzip() is None
    prefix = 'LOGH3-v.1.5.1-20261024-windows-amd64-patch/'
    manifest = json.loads(z.read(prefix+'MANIFEST.json'))
    assert manifest['version'] == build['version']
    assert all(i.filename.isascii() or i.flag_bits & 0x800 for i in z.infolist())
    readme = z.read(prefix+'README.txt')
    assert readme.startswith(b'\xef\xbb\xbf')
    readme.decode('utf-8-sig')
    for name,row in manifest['files'].items():
        b = z.read(prefix+name)
        assert len(b) == row['bytes'] and hashlib.sha256(b).hexdigest() == row['sha256']
        assert name in ('gin3.exe','LICENSE','README.txt') or name.startswith('licenses/')
    binary = z.read(prefix+'gin3.exe')
    assert binary[:2] == b'MZ'
    pe = struct.unpack_from('<I',binary,0x3c)[0]
    assert binary[pe:pe+4] == b'PE\0\0'
    assert struct.unpack_from('<H',binary,pe+4)[0] == 0x8664
binary_hash = hashlib.sha256(binary).hexdigest()
assert binary_hash == sha(root/'workplace/music69-windows/personnel145/gin3-final.exe')
version_file = root/'workplace/music69-windows/personnel145/version.txt'
assert version_file.read_text().strip() == build['version']
gui = []
for folder in [root/'workplace/personnel145-alliance-verified',
               root/'workplace/personnel145-empire-verified',
               root/'workplace/music69-windows/personnel145/alliance-verified']:
    result = json.loads((folder/'RESULT.json').read_text())
    assert result['passed'] and result['normal_entry'] and not result['state_injection'] and not result['script_flag']
    assert result['button_ms'] == 80
    if 'wine' in result['command']: assert result['binary_sha256'] == binary_hash
    captures = [e for e in result['events'] if e['Kind'] == 'capture']
    assert len(captures) == 28
    comparisons = []
    for e in captures:
        name = e['Name']+'-x11-comparison.json'
        row = json.loads((folder/name).read_text())
        assert row['different'] == 0 and row['pixels'] == 691200
        comparisons.append(dict(name=name,sha256=sha(folder/name)))
    gui.append(dict(path=str(folder.relative_to(root)),result_sha256=sha(folder/'RESULT.json'),
                    binary_sha256=result['binary_sha256'],acknowledged=result.get('acknowledged',False),captures=comparisons))
receipt = dict(version=build['version'],platform='Windows amd64，Wine驗證，未作Windows真機驗收',
    scope='remake前線來源修正、人事靜態背景／字模／游標快取，不新增DOS同狀態聲明',
    archive=build['archive'],binary_sha256=binary_hash,build_sha256=sha(out/'BUILD.json'),
    version_output_sha256=sha(version_file),
    source_files=len(build['source_files']),gui=gui,total_compared_pixels=3*28*691200,
    tests='tools/gui.sh test ./internal/... ./cmd/...：19套件，15 PASS／4無測試',
    installation='只用本包gin3.exe替換既有v1.5同名檔案；原資料與存檔格式保持')
target = root/'docs/internals/data/personnel-frontline145.json'
assert target.parent.stat().st_uid == os.getuid()
target.write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(dict(passed=True,archive=receipt['archive'],pixels=receipt['total_compared_pixels']),indent=2),flush=True)
