"""#69：Windows 公開保存／退出／主選單讀回；Linux 腳本僅作對照。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import time

LINUX_SHA = '48c89ba6e9b7bf069ccba7f6345228af4ffc23a7baf63fd87da02aa44ff70754'
WINDOWS_SHA = '6af940c72f340a5ecc49b2c6dc092e0e6dca077c1136abada3cad09b49e8ffa9'
STEMS = ['STELLAR', 'PLANET', 'ADMIRAL', 'FLEET', 'SHIP', 'WORK']
BASE = 'click 300 210;type 138;enter;rclick 300 240;rclick 300 240'
CREATED = BASE + ';click 352 125;click 400 125;click 400 85;click 488 330;click 480 350;click 510 350'
SAVE = CREATED + ';click 544 125;click 560 165'
LOAD = 'click 300 270;type 138;enter'
RESTORED = LOAD + ';click 300 122;click 318 208'
SCRIPTS = {
    'main': 'move 300 240', 'copy': 'click 300 210',
    'month': 'click 300 210;type 138;enter',
    'ship': 'click 300 210;type 138;enter;rclick 300 240', 'input': BASE,
    'created': CREATED, 'save': SAVE, 'saved': SAVE + ';click 300 142',
    'quit-new': SAVE + ';click 300 142;rclick 300 142;click 544 125;click 560 205',
    'copy-load': 'click 300 270', 'load': LOAD, 'faction': LOAD + ';click 300 122',
    'restored': RESTORED, 'resave': RESTORED + ';click 544 125;click 560 165;click 300 162',
    'quit-load': RESTORED + ';click 544 125;click 560 165;click 300 162;rclick 300 162;click 544 125;click 560 205',
}


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def record(path, value):
    with Path(path).open('x') as f:
        json.dump(value, f, ensure_ascii=False, indent=2)
        f.write('\n')


def win(path):
    return 'Z:' + str(path).replace('/', '\\')


def owned(path):
    for p in [path, *path.rglob('*')]:
        assert p.stat().st_uid == os.getuid() and p.stat().st_gid == os.getgid(), p


def build_players():
    assert Path.cwd() == Path('/src') and os.getuid() == 1000
    out = Path('workplace/music69-windows/saveload69-r106-build')
    out.mkdir()
    source = out / 'source'
    source.mkdir()
    for name in ['cmd', 'internal']:
        shutil.copytree(name, source / name)
    for name in ['go.mod', 'go.sum']:
        shutil.copyfile(name, source / name)
    shutil.copytree('workplace/source69-r103-d/extracted/logh3/vendor', source / 'vendor')
    subprocess.run(['go', 'build', '-mod=vendor', '-o', str((out / 'capture').resolve()),
                    str(Path('tools/scripts/fleet_player_gui_capture.go').resolve())], cwd=source, check=True)
    rows = []
    env = dict(os.environ, GOPROXY='off', GOSUMDB='off', GOWORK='off', GOFLAGS='', GOARCH='amd64')
    for platform in ['linux', 'windows']:
        for repeat in ['a', 'b']:
            name = platform + '-' + repeat + ('.exe' if platform == 'windows' else '')
            command = ['go', 'build', '-mod=vendor', '-trimpath', '-buildvcs=false',
                       '-o', str((out / name).resolve()), './cmd/gin3']
            subprocess.run(command, cwd=source, env=dict(env, GOOS=platform,
                            CGO_ENABLED='0' if platform == 'windows' else '1'), check=True)
            rows.append({'platform': platform, 'repeat': repeat, 'file': name, 'sha256': sha(out / name)})
    assert rows[0]['sha256'] == rows[1]['sha256'] and rows[2]['sha256'] == rows[3]['sha256']
    record(out / 'BUILD.json', {'binaries': rows, 'go': subprocess.check_output(['go', 'version'], text=True),
                              'flags': ['-mod=vendor', '-trimpath', '-buildvcs=false'],
                              'original_source_head': '91d85785b937145f060066cf9bb29edc643a7c05',
                              'change': 'only production cmd/gin3/main.go consumes Game.Quit; new quit_test.go'})
    owned(out)
    print('PASS independent offline Linux / Windows builds', rows, flush=True)


def references(revision):
    assert Path.cwd() == Path('/src') and os.getuid() == 1000
    out = Path('workplace/music69-windows') / ('saveload69-r106-' + revision)
    out.mkdir()
    build = Path('/src/workplace/music69-windows/saveload69-r106-build')
    binary = build / 'linux-a'
    identity = json.loads((build / 'BUILD.json').read_text())
    assert sha(binary) == identity['binaries'][0]['sha256']
    runs = []
    for name, script in SCRIPTS.items():
        save = out / ('保存 空間-' + name)
        save.mkdir()
        if name in ['copy-load', 'load', 'faction', 'restored', 'resave', 'quit-load']:
            for src in (out / '保存 空間-saved').iterdir():
                shutil.copyfile(src, save / src.name)
        command = [str(binary), '-orig', '/orig', '-data', '/src/docs/internals/data',
                   '-save', str(save.resolve()), '-script', script,
                   '-shot', str(out.resolve() / (name + '.png')),
                   '-shot-bin', str(out.resolve() / (name + '.bin')),
                   '-state-out', str(out.resolve() / (name + '.txt')), '-shot-after', '4']
        with (out / (name + '.log')).open('x') as log:
            subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=60)
        runs.append({'name': name, 'command': command, 'state': (out / (name + '.txt')).read_text(),
                     'png_sha256': sha(out / (name + '.png')), 'index_sha256': sha(out / (name + '.bin'))})
    record(out / 'REFERENCES.json', {'binary_sha256': sha(binary), 'script_diagnostic': True,
                                   'scope': '同來源Linux公開輸入／存檔對照，非新增原版同態', 'runs': runs})
    owned(out)
    print('PASS Linux references', len(runs), flush=True)


def gui(revision, reference_revision, windows=True):
    assert Path.cwd() == Path('/player' if windows else '/src') and os.getuid() == 1000
    root = Path('/out' if windows else '/src/workplace/music69-windows')
    out = root / ('saveload69-r106-' + revision)
    out.mkdir()
    assert out.parent.stat().st_uid == os.getuid() and out.parent.stat().st_gid == os.getgid()
    shutil.copyfile(__file__, out / 'driver.py')
    save = out / '玩家 保存 空間'
    save.mkdir()
    build = Path('/player') if windows else root / 'saveload69-r106-build'
    identity = json.loads((build / 'BUILD.json').read_text())
    binary = build / ('windows-a.exe' if windows else 'linux-a')
    assert sha(binary) == identity['binaries'][2 if windows else 0]['sha256']
    reference = root / ('saveload69-r106-' + reference_revision)
    refs = json.loads((reference / 'REFERENCES.json').read_text())
    for row in refs['runs']:
        assert sha(reference / (row['name'] + '.png')) == row['png_sha256']
    phases = []
    for phase in ['new', 'load']:
        folder = out / phase
        folder.mkdir()
        command = (['wine'] if windows else []) + [str(binary), '-orig', win('/orig') if windows else '/orig',
                   '-data', win('/data') if windows else '/src/docs/internals/data',
                   '-save', win(save) if windows else str(save)]
        start, actions, frames = time.monotonic(), [], []
        record(folder / 'START.json', {'command': command, 'binary_sha256': sha(binary),
                                      'platform': 'windows/amd64' if windows else 'linux/amd64',
                                      'wine': subprocess.check_output(['wine', '--version'], text=True).strip() if windows else None,
                                      'default_intro': True, 'script_used': False, 'state_injection': False,
                                      'working_directory': str(Path.cwd()),
                                      'readonly_inputs': ['/player', '/orig', '/data'] if windows else ['/orig'],
                                      'save_directory': str(save), 'reference_sha256': sha(reference / 'REFERENCES.json')})
        with (folder / 'player.log').open('x') as log:
            process = subprocess.Popen(command, stdout=log, stderr=subprocess.STDOUT)
            try:
                window = None
                while window is None:
                    assert process.poll() is None, (folder / 'player.log').read_text()
                    assert time.monotonic() - start < 20
                    r = subprocess.run(['xdotool', 'search', '--onlyvisible', '--name', '^銀河英雄傳說III SP$'],
                                       capture_output=True, text=True)
                    if r.returncode == 0:
                        window = r.stdout.strip().splitlines()[-1]
                    else:
                        time.sleep(.1)
                subprocess.run(['xdotool', 'windowfocus', window], check=True)
                time.sleep(.5)

                def event(args):
                    subprocess.run(['xdotool'] + args, check=True)
                    actions.append({'elapsed': time.monotonic() - start, 'command': ['xdotool'] + args})

                def click(button, x, y):
                    event(['mousemove', '--window', window, str(x), str(y)])
                    time.sleep(.08)
                    event(['mousedown', str(button)]); time.sleep(.08)
                    event(['mouseup', str(button)]); time.sleep(.15)

                def frame(name):
                    deadline, attempt = time.monotonic() + 3, 0
                    while True:
                        stem = name if attempt == 0 else name + '-wait-%02d' % attempt
                        png = folder / (stem + '.png')
                        if windows:
                            subprocess.run(['import', '-window', window, str(png)], check=True)
                            result = subprocess.run(['compare', '-metric', 'AE', str(reference / (name + '.png')),
                                                 str(png), str(folder / (stem + '-diff.png'))],
                                                capture_output=True, text=True)
                            different = result.stderr.strip()
                        else:
                            shutil.copyfile(reference / (name + '.png'), png)
                            result = subprocess.run([str(build / 'capture'), '-window', window,
                                                     '-base', str(folder / stem)], capture_output=True, text=True)
                            stats = json.loads((folder / (stem + '-x11-comparison.json')).read_text())
                            different = str(stats['different'])
                            png = folder / (stem + '-x11.png')
                        frames.append({'name': name, 'path': str(png), 'elapsed': time.monotonic() - start,
                                       'different_pixels': different, 'exit': result.returncode})
                        if result.returncode == 0: break
                        assert result.returncode == 1 and time.monotonic() < deadline, (name, result.stderr)
                        time.sleep(.1); attempt += 1

                click(1, 300, 240); frame('main')
                click(1, 300, 210 if phase == 'new' else 270)
                frame('copy' if phase == 'new' else 'copy-load')
                for key in ['1', '3', '8', 'Return']:
                    event(['keydown', '--window', window, key]); time.sleep(.08)
                    event(['keyup', '--window', window, key]); time.sleep(.15)
                if phase == 'new':
                    frame('month')
                    click(3, 300, 240); frame('ship')
                    click(3, 300, 240); frame('input')
                    for x, y in [(352, 125), (400, 125), (400, 85), (488, 330), (480, 350), (510, 350)]:
                        click(1, x, y)
                    frame('created')
                    click(1, 544, 125); click(1, 560, 165); frame('save')
                    click(1, 300, 142); frame('saved')
                    ext, expected = '0', reference / '保存 空間-saved'
                else:
                    frame('load')
                    click(1, 300, 122); frame('faction')
                    click(1, 318, 208); frame('restored')
                    click(1, 544, 125); click(1, 560, 165); click(1, 300, 162); frame('resave')
                    ext, expected = '1', reference / '保存 空間-resave'
                comparisons = []
                for name in [stem + '.DT' + ext for stem in STEMS] + ['SAVEDATA.IDX']:
                    a, b = (save / name).read_bytes(), (expected / name).read_bytes()
                    assert len(a) == len(b) and a == b, name
                    comparisons.append({'file': name, 'bytes': len(a), 'sha256': sha(save / name), 'different': 0})
                click(3, 300, 142 if phase == 'new' else 162)
                click(1, 544, 125); click(1, 560, 205); frame('quit-' + phase)
                click(1, 430, 270)
                rc = process.wait(timeout=10)
                assert rc == 0, (folder / 'player.log').read_text()
                phases.append({'phase': phase, 'exit': rc, 'public_quit': True,
                               'actions': len(actions), 'frames': frames, 'files': comparisons})
            finally:
                record(folder / 'TRACE.json', {'actions': actions, 'frames': frames})
                if process.poll() is None:
                    process.kill(); process.wait()
    record(out / 'CHECK.json', {'phases': phases, 'script_used': False, 'state_injection': False,
                               'fresh_process_load': True,
                               'scope': ('Windows Wine' if windows else 'Linux') + '正常公開存讀；非實機或新增原版同態'})
    owned(out)
    print('PASS', 'Windows' if windows else 'Linux', 'save / public quit / restart load / resave', revision, flush=True)


if __name__ == '__main__':
    p = argparse.ArgumentParser()
    p.add_argument('mode', choices=['build', 'references', 'gui', 'native-gui'])
    p.add_argument('--revision', default='build')
    p.add_argument('--reference', default='refa')
    a = p.parse_args()
    assert a.revision.isascii() and a.revision.isalnum()
    assert a.reference.isascii() and a.reference.isalnum()
    if a.mode == 'build': build_players()
    elif a.mode == 'references': references(a.revision)
    else: gui(a.revision, a.reference, a.mode == 'gui')
