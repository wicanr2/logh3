"""引擎包沿啟動器及玩家自備資料啟動，不把原版檔案複製進發行目錄。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import zipfile

VERSION = 'v.1.0.0-20261010'


def main(windows=False, original_override=None):
    assert os.getuid() == 1000
    root = Path('/src/dist-all') / VERSION
    output = Path('/out/release-smoke-' + VERSION) if windows else root / 'smoke/release-linux'
    output.mkdir()
    platform = 'windows-amd64' if windows else 'linux-amd64'
    if windows:
        with zipfile.ZipFile('/out/release-' + VERSION + '-windows-amd64.zip') as archive:
            assert archive.testzip() is None
            archive.extractall(output)
        package = output / f'LOGH3-{VERSION}-{platform}'
    else:
        package = root / 'release' / f'LOGH3-{VERSION}-{platform}'
    original = Path('/out/package92-smoke-' + VERSION) / f'LOGH3-{VERSION}-windows-amd64/original' if windows else Path('/orig')
    if original_override:
        original = Path(original_override)
    assert original.is_dir() and not (package / 'original').exists()

    def path(value):
        return 'Z:' + str(value.resolve()).replace('/', '\\') if windows else str(value.resolve())

    binary = package / ('gin3.exe' if windows else 'gin3')
    query = subprocess.check_output((['wine'] if windows else []) + [str(binary), '-version'], text=True, timeout=45)
    assert query.strip() == VERSION
    runs = []
    for mode in ('original', 'enhanced'):
        for name, script, expected in (
            ('main', 'move 300 240', 'StateMainMenu'),
            ('confirm-budget', 'click 300 210;type 138;click 432 270;click 448 350', 'StateBudgetPlan'),
        ):
            prefix = output / (mode + '-' + name)
            launcher = package / ('start-' + mode + ('.cmd' if windows else '.sh'))
            command = (['wine', 'cmd', '/c', path(launcher)] if windows else [str(launcher)]) + [
                '-orig', path(original), '-script', script, '-shot', path(prefix.with_suffix('.png')),
                '-state-out', path(prefix.with_suffix('.txt')), '-save', path(output / ('save-' + mode)), '-shot-after', '4']
            with prefix.with_suffix('.log').open('x') as log:
                subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, timeout=60, check=True)
            state = prefix.with_suffix('.txt').read_text()
            assert 'state=' + expected in state, state
            runs.append({'command': command, 'state': state, 'png_sha256': hashlib.sha256(prefix.with_suffix('.png').read_bytes()).hexdigest()})
    (output / 'SMOKE.json').write_text(json.dumps({'version': VERSION, 'platform': platform, 'runs': runs,
        'external_original_only': True, 'classification': 'Windows Wine啟動器' if windows else 'Linux正式啟動器'}, ensure_ascii=False, indent=2) + '\n')
    print('PASS 引擎包兩種AI啟動器／自備資料／確認與財政變更', platform, flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--windows', action='store_true')
    parser.add_argument('--version', default=VERSION)
    parser.add_argument('--original', help='唯讀的外部合法原版資料目錄')
    args = parser.parse_args()
    VERSION = args.version
    main(args.windows, args.original)
