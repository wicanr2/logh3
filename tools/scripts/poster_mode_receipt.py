"""149：核對正常視窗中央像素、海報、偏好、凍結來源與本機Windows更新包。Docker內執行。"""
import hashlib
import json
from pathlib import Path
import shutil
import struct
import zipfile

from PIL import Image, ImageChops

ROOT = Path('/src'); VERSION = 'v.1.5.3-20261024'; OUT = ROOT/'dist-all'/VERSION
read = lambda p: json.loads(Path(p).read_text())
write = lambda p, r: Path(p).write_text(json.dumps(r, ensure_ascii=False, indent=2)+'\n')


def sha(path):
    h = hashlib.sha256()
    with Path(path).open('rb') as f:
        for b in iter(lambda: f.read(1024*1024), b''): h.update(b)
    return h.hexdigest()


build = read(OUT/'BUILD.json'); archive = OUT/build['archive']['path']
assert archive.stat().st_size == build['archive']['bytes'] and sha(archive) == build['archive']['sha256']
frozen = ROOT/'workplace'/('personnel-patch-'+VERSION)/'source'
for name, value in build['source_files'].items():
    assert sha(frozen/name) == value
    if not name.startswith('vendor/'): assert sha(ROOT/name) == value, name
with zipfile.ZipFile(archive) as z:
    assert z.testzip() is None
    prefix = archive.stem+'/'
    manifest = json.loads(z.read(prefix+'MANIFEST.json')); assert manifest['version'] == VERSION
    assert set(z.namelist()) == {prefix+n for n in manifest['files']} | {prefix+'MANIFEST.json'}
    for name, row in manifest['files'].items():
        b = z.read(prefix+name)
        assert len(b) == row['bytes'] and hashlib.sha256(b).hexdigest() == row['sha256']
        assert name in ['gin3.exe', 'LICENSE', 'README.txt'] or name.startswith(('licenses/', 'art/poster/'))
        assert name.isascii() or z.getinfo(prefix+name).flag_bits & 0x800
    text = z.read(prefix+'README.txt'); assert text.startswith(b'\xef\xbb\xbf') and b'\n' not in text.replace(b'\r\n', b'')
    text = z.read(prefix+'art/poster/README.txt'); assert text.startswith(b'\xef\xbb\xbf') and b'\n' not in text.replace(b'\r\n', b'')
    exe = z.read(prefix+'gin3.exe'); pe = struct.unpack_from('<I', exe, 0x3c)[0]
    assert exe[:2] == b'MZ' and exe[pe:pe+4] == b'PE\0\0' and struct.unpack_from('<H', exe, pe+4)[0] == 0x8664
    for name in ['empire.png', 'alliance.png']:
        assert hashlib.sha256(z.read(prefix+'art/poster/'+name)).hexdigest() == sha(ROOT/'assets/hd/poster'/name)
binary_hash = hashlib.sha256(exe).hexdigest()
tests = (ROOT/'workplace/poster149/standard-tests.log').read_text().splitlines()
assert not any(s.startswith(('FAIL', '--- FAIL')) for s in tests)
assert sum(s.startswith('ok ') for s in tests) == 15 and sum(s.startswith('? ') for s in tests) == 4
gui = []
for folder in [ROOT/'workplace/poster149/gui-linux-r1', ROOT/'workplace/music69-windows/poster149/gui-windows-r1']:
    r = read(folder/'RESULT.json')
    assert r['passed'] and r['normal_entry'] and not any(r[k] for k in ['script_flag', 'state_injection', 'seed_injection'])
    if 'wine' in r['command']: assert r['binary_sha256'] == binary_hash
    else: assert r['binary_sha256'] == sha(ROOT/'workplace/poster149/gin3')
    snapshots = {e['name']: e for e in r['events']}; assert len(snapshots) == 13
    for e in snapshots.values(): assert sha(folder/e['screenshot']) == e['sha256']
    off = Image.open(folder/'map-off.png').convert('RGB'); on = Image.open(folder/'map-on.png').convert('RGB')
    assert off.size == (960, 720) and on.size == (1280, 720)
    difference = ImageChops.difference(off, on.crop((160, 0, 1120, 720)))
    assert difference.getbbox() is None, folder
    side_checks = []
    for name in ['settings-on', 'side-clicks-ignored', 'original-on', 'help-on', 'loaded-on', 'restart-on', 'typing-from-side']:
        im = Image.open(folder/(name+'.png')).convert('RGB'); assert im.size == (1280, 720)
        for side, box in [('left', (0, 0, 160, 720)), ('right', (1120, 0, 1280, 720))]:
            assert ImageChops.difference(on.crop(box), im.crop(box)).getbbox() is None, (folder, name, side)
        side_checks.append(name)
    assert not snapshots['settings-off']['observed']['poster_mode']
    assert snapshots['side-clicks-ignored']['observed']['state'] == 'StateSettings'
    assert snapshots['original-on']['observed']['graphics'] == 'original' and snapshots['original-on']['observed']['poster_mode']
    assert snapshots['help-on']['observed']['help_visible']
    assert snapshots['resized-settings']['window'] == [1152, 648] and snapshots['resized-off']['window'] == [864, 648]
    assert snapshots['restart-on']['observed']['poster_mode']
    missing = snapshots['missing-fallback']['observed']; assert not missing['poster_mode'] and not missing['poster_available']
    assert read(folder/'save/preferences.json')['poster_mode']
    assert len(r['saved']) == 6
    for n, digest in r['saved'].items(): assert sha(folder/'save'/n) == digest
    gui.append(dict(path=str(folder.relative_to(ROOT)), result_sha256=sha(folder/'RESULT.json'),
                    binary_sha256=r['binary_sha256'], captures=13, center=dict(pixels=off.width*off.height, different=0),
                    side_checks=side_checks, saved=r['saved']))
native = ROOT/'workplace/poster149/native-diagnostic/RESULT.json'; assert read(native)['passed']
native_gui = ROOT/'workplace/poster149/gui-native-r1/RESULT.json'
if native_gui.exists():
    n = read(native_gui); assert n['passed'] and n['original_only'] and n['center_comparison']['different'] == 0
preview = OUT/'full'/('LOGH3-'+VERSION+'-poster-preview.png')
if not preview.exists(): shutil.copyfile(ROOT/'workplace/poster149/gui-linux-r1/map-on.png', preview)
assert sha(preview) == sha(ROOT/'workplace/poster149/gui-linux-r1/map-on.png')
evidence = dict(passed=True, version=VERSION, archive=build['archive'], binary_sha256=binary_hash,
                build_sha256=sha(OUT/'BUILD.json'), frozen_files=len(build['source_files']),
                product_files=sum(not n.startswith('vendor/') for n in build['source_files']), gui=gui,
                compared_center_pixels=sum(g['center']['pixels'] for g in gui)+(read(native_gui)['center_comparison']['pixels'] if native_gui.exists() else 0),
                preview=dict(path=str(preview.relative_to(ROOT)), sha256=sha(preview), size=list(Image.open(preview).size)),
                art=read(ROOT/'assets/hd/poster/SOURCES.json'),
                prompts=dict(path='workplace/poster149/PROMPTS.json', sha256=sha(ROOT/'workplace/poster149/PROMPTS.json')),
                tests=dict(command='tools/gui.sh test ./internal/... ./cmd/...', passed=15, no_tests=4, sha256=sha(ROOT/'workplace/poster149/standard-tests.log')),
                native_diagnostic=dict(path=str(native.relative_to(ROOT)), sha256=sha(native)),
                native_gui=dict(path=str(native_gui.relative_to(ROOT)), sha256=sha(native_gui)) if native_gui.exists() else None,
                scope='使用者指定顯示外框，中央remake前後像素比較；無新增DOS parity、Windows真機或macOS真機聲明',
                rights='新海報為官方參考經image_gen重構，本機使用；不加入Git或公開引擎包。既有交付保持。',
                generators={str(p.relative_to(ROOT)): sha(p) for p in [Path(__file__).resolve(), ROOT/'tools/scripts/poster_mode_gui.py', ROOT/'tools/scripts/personnel_windows_patch.py']})
write(ROOT/'docs/internals/data/poster-mode149.json', evidence)
files = [archive, OUT/'BUILD.json', preview]
(OUT/'SHA256SUMS.txt').write_text(''.join(sha(p)+'  '+str(p.relative_to(OUT))+'\n' for p in files))
print(json.dumps(dict(passed=True, archive=build['archive'], captures=26+(5 if native_gui.exists() else 0), pixels=evidence['compared_center_pixels'], frozen=evidence['frozen_files'], products=evidence['product_files']), ensure_ascii=False, indent=2))
