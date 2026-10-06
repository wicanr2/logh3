"""規格121：核新封包、正常X11完整畫面、舊交付與歷史來源身份。容器內執行。"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tarfile
import zipfile

ROOT = Path('/src')
VERSION = 'v.0.1.3-20261006'


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for b in iter(lambda: f.read(1024*1024), b''):
            h.update(b)
    return h.hexdigest()


def entry(path):
    assert path.is_file() and path.stat().st_uid == os.getuid(), path
    return dict(path=str(path), bytes=path.stat().st_size, sha256=sha(path))


def record(path, data):
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')


def checked(path, expected):
    assert path.stat().st_size == expected['bytes'] and sha(path) == expected['sha256'], path


def main():
    assert Path.cwd() == ROOT and os.getuid() == 1000
    root = Path('dist-all')/VERSION
    build = json.loads((root/'BUILD.json').read_text())
    assert build['version'] == VERSION and len(build['packages']) == 4
    packages = []
    for package in build['packages']:
        path = root/package['path']
        checked(path, package)
        if path.suffix == '.zip':
            with zipfile.ZipFile(path) as z:
                assert z.testzip() is None
                top = path.stem
                mf = json.loads(z.read(top+'/MANIFEST.json'))
                assert mf['version'] == VERSION
                for name, expected in mf['files'].items():
                    b = z.read(top+'/'+name)
                    assert len(b) == expected['bytes'] and hashlib.sha256(b).hexdigest() == expected['sha256']
                assert set(z.namelist()) == {top+'/MANIFEST.json', *(top+'/'+name for name in mf['files'])}
        else:
            with tarfile.open(path) as t:
                members = [m for m in t if m.isfile()]
                assert len(members) == len(build['source_files'])
                for member in members:
                    name = member.name.removeprefix('logh3-source/')
                    b = t.extractfile(member).read()
                    expected = build['source_files'][name]
                    assert len(b) == expected['bytes'] and hashlib.sha256(b).hexdigest() == expected['sha256']
        packages.append(package)
    record(root/'SHA256SUMS.json', {'version': VERSION, 'files': packages})
    binary = root/'full-local'/('LOGH3-'+VERSION+'-linux-amd64/gin3')
    assert subprocess.check_output([str(binary.resolve()), '-version'], text=True).strip() == VERSION
    cases = []
    for name, folder in [('empire', 'input96-r120-package-gui-empire'), ('alliance', 'input96-r120-package-gui-alliance')]:
        source = Path('workplace')/folder
        trace = json.loads((source/'TRACE.json').read_text())
        assert trace['normal_input'] and not trace['script_flag'] and trace['command'][0] == str(binary.resolve())
        comparisons = sorted(source.glob('*-x11-comparison.json'))
        assert len(comparisons) == 24, comparisons
        for p in comparisons:
            value = json.loads(p.read_text())
            assert value['different'] == 0 and value['pixels'] == 307200 and value['max_channel_difference'] == 0, p
        destination = root/'smoke'/('gui-'+name)
        assert not destination.exists()
        shutil.copytree(source, destination)
        # 原擷取器沿用的「整數最近鄰」標籤未反映1.5倍中心取樣。
        # 只勘誤副本的metadata，原始PNG、統計結果與來源JSON保持。
        for p in comparisons:
            value = json.loads(p.read_text())
            value['normalization'] = '保存完整原始視窗；按邏輯像素中心取樣640×480，無排除區域'
            value['metadata_corrected_from'] = entry(p)
            record(destination/p.name, value)
        cases.append(dict(name=name, scale=trace['scale'], comparisons=len(comparisons), native_pixels=24*307200, different=0,
                          files=[entry(p) for p in sorted(destination.rglob('*')) if p.is_file()]))
    windows = Path('workplace/music69-windows/package92-smoke-'+VERSION+'/SMOKE.json')
    win = json.loads(windows.read_text())
    assert win['version'] == VERSION and len(win['runs']) == 4
    target = root/'smoke/windows'
    target.mkdir()
    shutil.copyfile(windows, target/'SMOKE.json')
    for p in windows.parent.glob('*'):
        if p.is_file() and p.suffix in ['.txt', '.log', '.png']:
            shutil.copyfile(p, target/p.name)
    test = Path('workplace/round120-standard-tests-final.log').read_text()
    assert 'FAIL' not in test and sum(l.startswith('ok') for l in test.splitlines()) == 13
    assert sum('[no test files]' in l for l in test.splitlines()) == 5
    frozen = []
    for name, key in [('macos-retro-promo.json', 'sources'), ('local-test-package.json', 'source_files')]:
        path = Path('docs/internals/data')/name
        old = json.loads(path.read_text())
        for item in old['delivery_files']:
            base = Path('dist-all')/old['version']
            checked(base/item['path'], item)
        for item in old[key]:
            data = subprocess.check_output(['git', 'show', '66d7ad7:'+item['path']])
            assert len(data) == item['bytes'] and hashlib.sha256(data).hexdigest() == item['sha256'], item['path']
        frozen.append(dict(receipt=entry(path), historical_revision='66d7ad7', sources=len(old[key]), artifacts_unchanged=True))
    source_names = ['cmd/gin3/main.go', 'internal/game/game.go', 'internal/game/playerinput.go',
                    'internal/game/playerinput_test.go', 'internal/game/planscreens_test.go',
                    'internal/game/game_test.go', 'internal/game/faction_test.go', 'internal/game/buildplan_states_test.go',
                    'internal/game/lefthold_test.go', 'internal/screen/maingame.go', 'tools/gui_check.sh',
                    'tools/scripts/player_input_model.go', 'tools/scripts/player_input_gui.py',
                    'tools/scripts/player_input_capture.go', 'tools/player_input_display.sh',
                    'tools/scripts/player_input_receipt.py', 'docs/spec/121-player-input-hotspots.md',
                    'docs/re/441-player-input-hotspot-space.md', 'docs/playtest/107-player-input-hotspots.md',
                    'docs/goal/2026-10-06-player-input-round-120.md', 'README.md']
    metadata = dict(schema='logh3.player-input-hotspots.v1', version=VERSION, date='2026-10-06',
                    source_revision=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
                    working_tree=subprocess.check_output(['git','status','--porcelain'],text=True),
                    scope='remake可見按鈕、正常X11與renderer一致性；無新增原版對拍',
                    gui=cases, delivery_files=packages, frozen_previous=frozen,
                    sources=[entry(Path(n)) for n in source_names],
                    reproduction=[entry(p) for p in sorted(Path('workplace/input96-r120-old').glob('*')) if p.is_file()],
                    artifacts=[entry(p) for p in sorted((root/'smoke').rglob('*')) if p.is_file()],
                    tests=[entry(Path('workplace')/n) for n in ['round120-before-tests.log','round120-target-tests.log',
                            'round120-save-reload-tests.log','round120-settlement-tests.log',
                            'round120-standard-tests-final.log','round120-final-verify.log']],
                    limitations=['Windows僅Wine入口抽驗，未真機','macOS雙架構封包核對，未真機','未新增人耳與戰場對拍',
                                 '引擎來源tar為建置時快照；後補的保存讀回測試與註解另在本索引保存'],
                    issue='建立Issue被自動審核拒絕，未寫GitHub；使用者缺陷修正已完成，不影響封包')
    path = Path('docs/internals/data/player-input-hotspots.json')
    assert not path.exists()
    record(path, metadata)
    print('PASS新三平台封包、48張正式包X11全圖、13套件、前兩版與歷史來源；', entry(path), flush=True)


if __name__ == '__main__':
    main()
