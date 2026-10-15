"""驗證 Windows 封包從任意目錄直接啟動；由 tools/player_wine.sh 執行。"""
import hashlib
import json
import os
from pathlib import Path
import subprocess

ROOT = Path('/out/windows138-paths')
PACKAGE = ROOT / '銀河 英雄'
assert PACKAGE.is_dir() and os.getuid() == os.getgid() == 1000
os.environ['LP_NUM_THREADS'] = '1'


def win(path):
    return 'Z:' + str(path).replace('/', '\\')


def run(name, binary, extra, expected):
    prefix = ROOT / name
    command = ['wine', win(binary), '-save', win(ROOT / 'save'),
               '-shot', win(prefix.with_suffix('.png')),
               '-state-out', win(prefix.with_suffix('.txt')),
               '-shot-after', '4'] + extra
    with prefix.with_suffix('.log').open('w') as log:
        result = subprocess.run(command, cwd='/tmp', stdout=log, stderr=subprocess.STDOUT, timeout=90)
    state = prefix.with_suffix('.txt').read_text() if prefix.with_suffix('.txt').exists() else ''
    assert result.returncode == expected[0], prefix.with_suffix('.log').read_text()
    observed = state if expected[0] == 0 else prefix.with_suffix('.log').read_text().replace('\\', '/')
    assert expected[1] in observed, observed
    return dict(name=name, command=command, exit=result.returncode, state=state,
                binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                png_sha256=hashlib.sha256(prefix.with_suffix('.png').read_bytes()).hexdigest() if state else None)


runs = []
runs.append(run('before', PACKAGE / 'old.exe', ['-script', 'move 300 240'], (1, 'org_game/extracted')))
runs.append(run('direct-script', PACKAGE / 'gin3.exe', ['-script', 'move 300 240'], (0, 'StateMainMenu')))
runs.append(run('direct-intro', PACKAGE / 'gin3.exe', [], (0, 'StateIntro')))
runs.append(run('explicit-missing', PACKAGE / 'gin3.exe', ['-orig', 'missing-data', '-script', 'move 300 240'], (1, 'missing-data')))
(ROOT / 'RESULT.json').write_text(json.dumps(dict(runs=runs, classification='Windows Wine，非 Windows 真機',
    external_cwd='/tmp', resources_flags_omitted=['orig', 'data', 'art', 'music', 'graphics']), ensure_ascii=False, indent=2)+'\n')
print('PASS Windows：舊版重現、直接啟動、正常片頭、明示錯誤路徑保持')
