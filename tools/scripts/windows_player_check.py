"""#69：來源重建 Windows 玩家入口的 Wine GUI 抽驗。只在 Docker 執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

LINUX_SHA = '48c89ba6e9b7bf069ccba7f6345228af4ffc23a7baf63fd87da02aa44ff70754'
WINDOWS_SHA = '6af940c72f340a5ecc49b2c6dc092e0e6dca077c1136abada3cad09b49e8ffa9'


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def record(path, value):
    with Path(path).open('x') as stream:
        json.dump(value, stream, ensure_ascii=False, indent=2)
        stream.write('\n')


def win(path):
    return 'Z:' + str(path).replace('/', '\\')


def references(reference_name):
    assert Path.cwd() == Path('/src') and os.getuid() == 1000
    out = Path('workplace/music69-windows') / ('player69-r104-' + reference_name)
    out.mkdir()
    binary = Path('/src/workplace/source69-r103-d/rebuild/gin3-linux-source')
    assert sha(binary) == LINUX_SHA
    base = 'click 300 210;type 138;enter;rclick 300 240;rclick 300 240'
    scripts = {
        'main': 'move 300 240',
        'copy': 'click 300 210',
        'month': 'click 300 210;type 138;enter',
        'ship': 'click 300 210;type 138;enter;rclick 300 240',
        'input': base,
        'system': base + ';click 544 125',
        'load': base + ';click 544 125;click 560 145',
        'fleet': base + ';click 352 125;click 400 125;click 400 85;click 488 330;click 368 350',
    }
    runs = []
    for name, script in scripts.items():
        command = [str(binary), '-orig', '/orig', '-data', '/src/docs/internals/data',
                   '-save', str(out.resolve() / ('save-' + name)), '-script', script,
                   '-shot', str(out.resolve() / (name + '.png')),
                   '-shot-bin', str(out.resolve() / (name + '.bin')),
                   '-state-out', str(out.resolve() / (name + '.txt')), '-shot-after', '4']
        with (out / (name + '.log')).open('x') as log:
            subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=60)
        runs.append({'name': name, 'command': command, 'state': (out / (name + '.txt')).read_text(),
                     'png_sha256': sha(out / (name + '.png')), 'index_sha256': sha(out / (name + '.bin'))})
    record(out / 'REFERENCES.json', {'binary_sha256': sha(binary), 'script_diagnostic': True,
                                   'scope': '目前Linux渲染及公開輸入對照，非新增原版同態', 'runs': runs})
    print('PASS', len(runs), 'Linux references', flush=True)


def scaled_references(revision):
    assert Path.cwd() == Path('/src') and os.getuid() == 1000
    root = Path('workplace/music69-windows')
    out = root / ('player69-r104-' + revision)
    out.mkdir()
    binary = Path('/src/workplace/source69-r103-d/rebuild/gin3-linux-source')
    assert sha(binary) == LINUX_SHA
    original = Path('tools/scripts/fleet_player_gui_capture.go').read_text()
    source = original.replace('flag.Int("scale", 1, "整數倍率")', 'flag.Float64("scale", 1.5, "縮放倍率")')
    before = 'if *scale < 1 || *scale > 2 || int(g.Width) != 640**scale || int(g.Height) != 480**scale {'
    after = 'if *scale != 1.5 || int(g.Width) != 960 || int(g.Height) != 720 {'
    assert source.count(before) == 1
    source = source.replace(before, after)
    source = source.replace('capture.RGBAAt(x**scale, y**scale)',
                            'capture.RGBAAt(int((float64(x)+0.5)*(*scale)), int((float64(y)+0.5)*(*scale)))')
    source = source.replace('整數最近鄰；完整邏輯畫面無排除', '1.5倍邏輯畫素中心；完整邏輯畫面無排除')
    with (out / 'capture.go').open('x') as stream:
        stream.write(source)
    subprocess.run(['gofmt', '-w', str(out / 'capture.go')], check=True)
    subprocess.run(['go', 'build', '-o', str(out / 'capture'), str(out / 'capture.go')], check=True)
    references = json.loads((root / 'player69-r104-reference2/REFERENCES.json').read_text())
    runs = []
    for name in ['main', 'copy', 'month', 'ship', 'input', 'system', 'load']:
        row = next(r for r in references['runs'] if r['name'] == name)
        script = row['command'][row['command'].index('-script') + 1]
        command = [str(binary), '-orig', '/orig', '-data', '/src/docs/internals/data',
                   '-save', str(out.resolve() / ('save-' + name)), '-script', script, '-scale', '1.5',
                   '-shot', str(out.resolve() / (name + '-final.png')),
                   '-shot-bin', str(out.resolve() / (name + '.bin')),
                   '-state-out', str(out.resolve() / (name + '.txt')), '-shot-after', '180']
        with (out / (name + '.log')).open('x') as log:
            process = subprocess.Popen(command, stdout=log, stderr=subprocess.STDOUT)
            try:
                deadline, window = time.monotonic() + 10, None
                while window is None:
                    assert process.poll() is None
                    assert time.monotonic() < deadline
                    result = subprocess.run(['xdotool', 'search', '--onlyvisible', '--pid',
                                             str(process.pid), '--name', '^銀河英雄傳說III SP$'],
                                            capture_output=True, text=True)
                    if result.returncode == 0:
                        window = result.stdout.strip().splitlines()[-1]
                    else:
                        time.sleep(.05)
                time.sleep(.6)
                capture_deadline, attempts = time.monotonic() + 1.5, []
                while True:
                    stem = name if not attempts else name + '-phase-%02d' % len(attempts)
                    with (out / (stem + '.png')).open('xb') as expected:
                        expected.write((root / 'player69-r104-reference2' / (name + '.png')).read_bytes())
                    capture = [str((out / 'capture').resolve()), '-window', window,
                               '-base', str(out.resolve() / stem)]
                    captured = subprocess.run(capture, capture_output=True, text=True, timeout=5)
                    attempts.append({'command': capture, 'exit': captured.returncode,
                                     'stdout': captured.stdout, 'stderr': captured.stderr})
                    if captured.returncode == 0: break
                    assert captured.returncode == 1 and time.monotonic() < capture_deadline, attempts
                    time.sleep(.05)
                rc = process.wait(timeout=15)
            finally:
                if process.poll() is None:
                    process.kill(); process.wait()
        assert rc == 0
        final = (out / (name + '.bin')).read_bytes()
        expected = (root / 'player69-r104-reference2' / (name + '.bin')).read_bytes()
        different = sum(a != b for a, b in zip(final, expected))
        assert different == 0 or name == 'copy' and different == 32
        runs.append({'name': name, 'command': command, 'capture': capture, 'exit': rc,
                     'state': (out / (name + '.txt')).read_text(),
                     'window': stem + '-x11.png', 'window_sha256': sha(out / (stem + '-x11.png')),
                     'capture_attempts': attempts, 'later_final_index_different': different,
                     'capture_phase': '完整640x480邏輯畫素中心符合固定對照；保留防拷閃爍探針'})
    record(out / 'REFERENCES.json', {'binary_sha256': sha(binary), 'script_diagnostic': True,
                                   'scale': 1.5, 'raw_window_pixels': 691200, 'runs': runs})
    print('PASS 7 actual Linux 960x720 windows', flush=True)


def startup(revision):
    assert Path.cwd() == Path('/out') and os.getuid() == 1000
    out = Path('/out') / ('player69-r104-' + revision)
    out.mkdir()
    with (out / 'driver.py').open('xb') as stream:
        stream.write(Path(__file__).read_bytes())
    binary = Path('/out/source103-player/gin3.exe')
    assert sha(binary) == WINDOWS_SHA
    command = ['wine', str(binary), '-orig', win('/orig'), '-data', win('/data'),
               '-save', win(out / 'save'), '-shot', win(out / 'title.png'),
               '-shot-bin', win(out / 'title.bin'), '-state-out', win(out / 'title.txt'),
               '-shot-after', '4998']
    record(out / 'START.json', {'command': command, 'binary_sha256': sha(binary),
                               'wine': subprocess.check_output(['wine', '--version'], text=True).strip(),
                               'default_intro': True, 'script_used': False})
    start = time.monotonic()
    window, screenshots = None, []
    with (out / 'player.log').open('x') as log:
        process = subprocess.Popen(command, stdout=log, stderr=subprocess.STDOUT)
        try:
            while process.poll() is None:
                elapsed = time.monotonic() - start
                if elapsed > 150:
                    raise TimeoutError('Windows正常片頭150秒期限')
                if window is None:
                    result = subprocess.run(['xdotool', 'search', '--onlyvisible', '--name',
                                             '^銀河英雄傳說III SP$'], capture_output=True, text=True)
                    if result.returncode == 0:
                        window = result.stdout.strip().splitlines()[-1]
                if window and elapsed >= 70 + len(screenshots) * 2:
                    png = out / ('window-%02d.png' % len(screenshots))
                    result = subprocess.run(['import', '-window', window, str(png)], capture_output=True)
                    if result.returncode == 0:
                        screenshots.append({'path': png.name, 'elapsed': elapsed})
                time.sleep(.1)
            rc = process.returncode
        finally:
            if process.poll() is None:
                process.kill()
                process.wait()
    record(out / 'PROCESS.json', {'exit': rc, 'window': window,
                                 'elapsed': time.monotonic() - start, 'screenshots': screenshots})
    assert rc == 0, (out / 'player.log').read_text()
    assert 'state=StateTitle\n' in (out / 'title.txt').read_text()
    assert (out / 'title.bin').stat().st_size == 307200
    print('PASS Windows normal intro to title', flush=True)


def gui(revision, target, scale, reference_name):
    assert Path.cwd() == Path('/out') and os.getuid() == 1000
    out = Path('/out') / ('player69-r104-' + revision)
    out.mkdir()
    with (out / 'driver.py').open('xb') as stream:
        stream.write(Path(__file__).read_bytes())
    binary = Path('/out/source103-player/gin3.exe')
    reference = Path('/out') / ('player69-r104-' + reference_name)
    assert sha(binary) == WINDOWS_SHA
    reference_manifest = json.loads((reference / 'REFERENCES.json').read_text())
    for row in reference_manifest['runs']:
        assert sha(reference / (row['name'] + '.png')) == row['png_sha256']
    command = ['wine', str(binary), '-orig', win('/orig'), '-data', win('/data'),
               '-save', win(out / 'save'), '-shot', win(out / 'final.png'),
               '-shot-bin', win(out / 'final.bin'), '-state-out', win(out / 'final.txt'),
               '-shot-after', '1800', '-scale', str(scale)]
    record(out / 'START.json', {'command': command, 'binary_sha256': sha(binary),
                               'wine': subprocess.check_output(['wine', '--version'], text=True).strip(),
                               'default_intro': True, 'script_used': False, 'target': target, 'scale': scale,
                               'reference': str(reference), 'reference_sha256': sha(reference / 'REFERENCES.json')})
    start, actions, frames = time.monotonic(), [], []
    with (out / 'player.log').open('x') as log:
        process = subprocess.Popen(command, stdout=log, stderr=subprocess.STDOUT)
        try:
            window = None
            while window is None:
                assert process.poll() is None, (out / 'player.log').read_text()
                assert time.monotonic() - start < 20, 'Windows視窗20秒期限'
                result = subprocess.run(['xdotool', 'search', '--onlyvisible', '--name',
                                         '^銀河英雄傳說III SP$'], capture_output=True, text=True)
                if result.returncode == 0:
                    window = result.stdout.strip().splitlines()[-1]
                else:
                    time.sleep(.1)
            subprocess.run(['xdotool', 'windowfocus', window], check=True)
            time.sleep(.5)

            def event(args):
                subprocess.run(['xdotool'] + args, check=True)
                actions.append({'elapsed': time.monotonic() - start, 'command': ['xdotool'] + args})

            def click(button, x, y):
                event(['mousemove', '--window', window, str(round(x * scale)), str(round(y * scale))])
                time.sleep(.08)
                event(['mousedown', str(button)])
                time.sleep(.08)
                event(['mouseup', str(button)])
                time.sleep(.15)

            def frame(name):
                # 全視窗比較。縮放樣本只做已知的最近鄰正規化，保存未縮放原圖。
                deadline = time.monotonic() + 3
                attempt = 0
                while True:
                    stem = name if attempt == 0 else name + '-wait-%02d' % attempt
                    raw = out / (stem + '-window.png')
                    normalized = out / (stem + '.png')
                    subprocess.run(['import', '-window', window, str(raw)], check=True)
                    subprocess.run(['convert', str(raw), '-filter', 'point', '-resize', '640x480!',
                                    str(normalized)], check=True)
                    result = subprocess.run(['compare', '-metric', 'AE', str(reference / (name + '.png')),
                                             str(normalized), str(out / (stem + '-diff.png'))],
                                            capture_output=True, text=True)
                    frames.append({'name': name, 'path': stem, 'elapsed': time.monotonic() - start,
                                   'different_pixels': result.stderr.strip(), 'exit': result.returncode})
                    if result.returncode == 0:
                        break
                    assert result.returncode == 1 and time.monotonic() < deadline, (name, result.stderr)
                    time.sleep(.1)
                    attempt += 1

            click(1, 300, 240)  # 正式玩家操作跳過片頭，未使用 -intro=false。
            frame('main')
            click(1, 300, 210)
            frame('copy')
            for key in ['1', '3', '8', 'Return']:
                event(['keydown', '--window', window, key]); time.sleep(.08)
                event(['keyup', '--window', window, key]); time.sleep(.15)
            frame('month')
            click(3, 300, 240); frame('ship')
            click(3, 300, 240)
            frame('input')
            if target == 'load':
                click(1, 544, 125); frame('system')
                click(1, 560, 145); frame('load')
                expected_state = 'StateLoadGame'
            else:
                for x, y in [(352, 125), (400, 125), (400, 85), (488, 330), (368, 350)]:
                    click(1, x, y)
                frame('fleet')
                expected_state = 'StateFleetNumber'
            rc = process.wait(timeout=40)
        finally:
            record(out / 'TRACE.json', {'actions': actions, 'frames': frames})
            if process.poll() is None:
                process.kill(); process.wait()
    assert rc == 0, (out / 'player.log').read_text()
    state = (out / 'final.txt').read_text()
    assert 'state=' + expected_state + '\n' in state, state
    assert 'window=%dx%d\n' % (round(640 * scale), round(480 * scale)) in state, state
    assert 'cursor_mode=1\n' in state, state
    assert (out / 'final.bin').read_bytes() == (reference / (target + '.bin')).read_bytes()
    record(out / 'CHECK.json', {'exit': rc, 'state': state, 'indexed_pixels': 307200,
                               'different': 0, 'frames': frames, 'actions': len(actions),
                               'script_used': False, 'state_injection': False,
                               'scope': 'Windows Wine正常GUI與同來源Linux對照，非實機或新增原版同態'})
    print('PASS Windows GUI', target, scale, 'frames', len(frames), flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('mode', choices=['references', 'scaled-references', 'startup', 'gui'])
    parser.add_argument('--revision', default='repeat')
    parser.add_argument('--target', choices=['load', 'fleet'], default='load')
    parser.add_argument('--scale', type=float, choices=[1, 1.5], default=1)
    parser.add_argument('--reference', default='reference2')
    args = parser.parse_args()
    assert args.revision.isascii() and args.revision.isalnum()
    assert args.reference.isascii() and args.reference.isalnum()
    if args.mode == 'references': references(args.reference)
    elif args.mode == 'scaled-references': scaled_references(args.revision)
    elif args.mode == 'startup': startup(args.revision)
    else: gui(args.revision, args.target, args.scale, args.reference)
