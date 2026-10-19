"""143：正式封包驗證資料、README原尺寸截圖及本機交付身份；Docker內執行。"""
import argparse
import ast
import hashlib
import json
import os
from pathlib import Path
import shutil
import zipfile

from PIL import Image

ROOT = Path('/src')
VERSION = 'v.1.5.0-20261019'
OUT = ROOT/'dist-all'/VERSION
BATCH = ROOT/'workplace/admiral143'
WINE = ROOT/'workplace/music69-windows/windows143-delivery'


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1024*1024), b''):
            h.update(block)
    return h.hexdigest()


def read(path): return json.loads(path.read_text())
def write(path, value): path.write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')


def prepare():
    assert not BATCH.exists() and BATCH.parent.stat().st_uid == os.getuid()
    BATCH.mkdir()
    for name in ['admiral_editor_gui.py', 'admiral_delivery.py', 'promo_v150.py']:
        ast.parse((ROOT/'tools/scripts'/name).read_text())
    assert not WINE.exists() and WINE.parent.stat().st_uid == os.getuid()
    WINE.mkdir()
    for kind, target in [('full', 'full.zip'), ('release', 'engine.zip')]:
        source = OUT/kind/f'LOGH3-{VERSION}-windows-amd64.zip'
        shutil.copyfile(source, WINE/target)
        assert sha(source) == sha(WINE/target)
    # 專案既有權限保留解壓器之後建立BATCH/package，這裡只準備輸出根。
    print('READY 正式Windows完整／引擎ZIP獨立副本與本輪輸出根', flush=True)


def images(check=False):
    folder = BATCH/'editor-linux'
    result = read(folder/'RESULT.json')
    assert result['passed'] and result['normal_entry'] and not result['script_flag'] and not result['state_injection']
    binary = BATCH/'package'/f'LOGH3-{VERSION}-linux-amd64/gin3'
    assert result['binary_sha256'] == sha(binary)
    captures = {row['capture']: row for row in result['events'] if 'capture' in row}
    mapping = {'settings-empire': '0-settings', 'settings-alliance': '1-settings',
               'editor-empire': '0-special', 'editor-alliance': '1-special',
               'editor-basic': '0-basic', 'editor-internal': '0-hidden',
               'editor-original': '0-original',
               'editor-applied': '0-applied',
               'editor-kemp': '0-kemp-fighter', 'editor-poplan': '1-poplan-fighter'}
    entries = []
    for label, capture in mapping.items():
        row = captures[capture]
        source = folder/row['file']
        assert sha(source) == row['sha256']
        with Image.open(source) as im: assert im.size == (960, 720)
        target = ROOT/'docs/images'/f'readme-v150-{label}.png'
        if not check:
            assert not target.exists() and target.parent.stat().st_uid == os.getuid()
            shutil.copyfile(source, target)
        assert sha(source) == sha(target)
        entries.append(dict(file=str(target.relative_to(ROOT)), source=str(source.relative_to(ROOT)),
                            sha256=sha(target), bytes=target.stat().st_size, observed=row['observed']))
    for faction, named in [(0, {'map-empire': 'strategy-current', 'battle-empire': 'entry', 'scenarios': 'load', 'ai': 'ai-menu'}),
                           (1, {'map-alliance': 'strategy-current', 'battle-alliance': 'turn-02'})]:
        source_folder = BATCH/f'gameplay-f{faction}-final'
        trace = read(source_folder/'TRACE.json')
        assert trace['normal_x11_input'] and not trace['state_injection'] and trace['binary_sha256'] == sha(binary)
        events = {e['name']: e for e in trace['events'] if e.get('kind') == 'capture'}
        for label, capture in named.items():
            assert capture in events
            source = source_folder/(capture+'.png')
            target = ROOT/'docs/images'/f'readme-v150-{label}.png'
            if not check:
                assert not target.exists()
                shutil.copyfile(source, target)
            with Image.open(target) as im: assert im.size == (960, 720)
            assert sha(target) == sha(source)
            entries.append(dict(file=str(target.relative_to(ROOT)), source=str(source.relative_to(ROOT)), sha256=sha(target),
                                bytes=target.stat().st_size, source_trace_sha256=sha(source_folder/'TRACE.json'), capture_event=events[capture]))
    plan = read(BATCH/'movie/PLAN.json')
    for weapon, label in [(0, 'beam'), (1, 'missile'), (2, 'fighters')]:
        scene = next(s for s in plan['scenes'] if s.get('battle_animation') and s['weapon'] == weapon)
        target = ROOT/'docs/images'/f'readme-v150-{label}.png'
        seconds = scene['start']+scene['duration']/2
        if not check:
            assert not target.exists()
            import subprocess
            subprocess.run(['ffmpeg', '-nostdin', '-v', 'error', '-threads', '2', '-ss', str(seconds), '-i', scene['file'],
                            '-frames:v', '1', str(target)], check=True, timeout=30)
        with Image.open(target) as im: assert im.size == (960, 720)
        entries.append(dict(file=str(target.relative_to(ROOT)), sha256=sha(target), bytes=target.stat().st_size,
                            source=scene['file'], source_sha256=sha(Path(scene['file'])), source_seconds=seconds, playback=scene['playback_id']))
    data = dict(version=VERSION, source='實際Linux完整版ZIP解壓後正常片頭至S1操作，未縮放完整視窗',
                zip_sha256=sha(OUT/'full'/f'LOGH3-{VERSION}-linux-amd64.zip'), binary_sha256=sha(binary),
                receipt_sha256=sha(folder/'RESULT.json'), files=entries,
                generator_sha256=sha(Path(__file__)))
    manifest = ROOT/'docs/images/readme-v150.json'
    if check: assert read(manifest) == data
    else:
        assert not manifest.exists()
        write(manifest, data)
    print('PASS README'+str(len(entries))+'張正式包原尺寸實錄'+('獨立核對' if check else '已匯出'), flush=True)


def receipt(check=False):
    package_record = read(OUT/'SHA256SUMS-platforms.json')
    assert len(package_record['files']) == 7
    for row in package_record['files']:
        path = OUT/row['path']
        assert path.stat().st_size == row['bytes'] and sha(path) == row['sha256']
    for kind in ['full', 'release']:
        with zipfile.ZipFile(OUT/kind/f'LOGH3-{VERSION}-windows-amd64.zip') as archive:
            assert all(i.filename.isascii() or i.flag_bits & 0x800 for i in archive.infolist())
            prefix = f'LOGH3-{VERSION}-windows-amd64/'
            notes = archive.read(prefix+'README.txt')
            assert notes.startswith(b'\xef\xbb\xbf') and b'\n' not in notes.replace(b'\r\n', b'')
            for mode in ['original', 'enhanced']:
                script = archive.read(prefix+f'start-{mode}.cmd')
                assert script.isascii() and b'\n' not in script.replace(b'\r\n', b'')
    build = read(OUT/'BUILD.json')
    frozen = ROOT/'workplace'/f'windows-package-{VERSION}/source'
    for name, row in build['source_files'].items():
        assert sha(frozen/name) == row['sha256']
        if not name.startswith('vendor/'): assert sha(ROOT/name) == row['sha256'], name
    windows = read(WINE/'RESULT.json')
    assert len(windows['runs']) == 8 and all(r['exit'] == (1 if 'missing' in r['name'] else 0) for r in windows['runs'])
    gui = [BATCH/'editor-linux/RESULT.json', ROOT/'workplace/music69-windows/windows143-editor-r2/RESULT.json']
    for path in gui:
        result = read(path)
        assert result['passed'] and result['normal_entry'] and not result['state_injection'] and not result['script_flag']
    diagnostics = [OUT/'smoke/full-linux/SMOKE.json', OUT/'smoke/macos/MACOS.json', WINE/'RESULT.json']
    assert len(read(diagnostics[0])['runs']) == 11
    assert {s['architecture'] for s in read(diagnostics[1])['slices']} == {'amd64', 'arm64'}
    promo = read(OUT/'promo/SHA256SUMS.json')
    for row in promo['files']:
        path = ROOT/row['path']
        assert path.stat().st_size == row['bytes'] and sha(path) == row['sha256']
    images(True)
    tested = read(ROOT/'docs/images/admiral-editor.json')
    for name, digest in tested['product_sources'].items(): assert sha(ROOT/name) == digest, name
    assert sha(ROOT/'workplace/admiral-editor-tests.log') == tested['tests']['sha256']
    record = dict(version=VERSION, rights='完整包與影片僅本機；引擎包候選未上傳',
                  packages=package_record, build_sha256=sha(OUT/'BUILD.json'),
                  frozen_sources=build['source_files'], source_head=build['head'], working_tree=build['working_tree'],
                  receipts={str(p.relative_to(ROOT)): sha(p) for p in gui+diagnostics},
                  readme_sha256=sha(ROOT/'docs/images/readme-v150.json'), promo=promo,
                  generators={str(p.relative_to(ROOT)): sha(p) for p in [
                      Path(__file__), ROOT/'tools/scripts/promo_v150.py', ROOT/'tools/scripts/admiral_editor_gui.py',
                      ROOT/'tools/scripts/windows_hotfix_package.py', ROOT/'tools/scripts/v140_platform_delivery.py',
                      ROOT/'tools/macos.sh']},
                  test_receipt_sha256=sha(ROOT/'docs/images/admiral-editor.json'),
                  tests=tested['tests'],
                  limits=['Windows為Wine，macOS未真機', '未新增人耳或DOS整場同態', '#48保持OPEN'])
    target = ROOT/'docs/internals/data/v150-delivery.json'
    if check: assert read(target) == record
    else:
        assert not target.exists()
        write(target, record)
    print('PASS 三平台七包、正式GUI、README及推廣片交付身份', flush=True)


if __name__ == '__main__':
    assert os.getuid() == os.getgid() == 1000 and Path.cwd() == ROOT
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode', choices=['prepare', 'images', 'receipt'])
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    if args.mode == 'prepare': prepare()
    else: globals()[args.mode](args.check)
