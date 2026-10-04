"""#79 子步8-3產品來源索引；控制對拍、正常操作與診斷保存分開。"""
import hashlib
import json
import os
from pathlib import Path

sources = {}


def read(path):
    p = Path(path)
    b = p.read_bytes()
    sources[str(p)] = {'size': len(b), 'sha256': hashlib.sha256(b).hexdigest()}
    return b


original = json.loads(read('docs/internals/data/scout-fuel.json'))
for p, s in original['sources'].items():
    b = read(p)
    assert len(b) == s['size'] and hashlib.sha256(b).hexdigest() == s['sha256']
target = read('workplace/round83-target-final.log').decode()
assert target.count('--- FAIL:') == 0 and target.count('--- SKIP:') == 0
for name in ('OriginalCases', 'WindowOriginal', 'SequentialReports', 'PlayerDeployTurnSave'):
    assert '--- PASS: TestScoutFuel' + name in target
standard = read('workplace/round83-standard.log').decode()
assert 'FAIL' not in standard
passed = [x for x in standard.splitlines() if x.startswith('ok\t') or x.startswith('ok  \t')]
empty = [x for x in standard.splitlines() if '[no test files]' in x]
assert len(passed) == 13 and len(empty) == 5
gui = read('workplace/round83-gui-v2.log').decode()
assert '倍率1' in gui and '倍率2' in gui
entry = read('workplace/scout79-r83-normal-entry/entry/screen.bin')
assert len(entry) == 307200
for scale in (1, 2):
    prefix = f'workplace/scout79-r83-gui-v2-{scale}x'
    assert read(prefix + '.bin') == entry
    state = read(prefix + '.txt').decode()
    assert 'state=StateBattle' in state and f'window={640*scale}x{480*scale}' in state
    read(prefix + '.png')
    read(prefix + '.log')
comparisons = []
for fleet in (12, 16):
    prefix = f'workplace/scout79-r83-game-final/fleet{fleet}'
    d = json.loads(read(prefix + '-comparison.json'))
    assert d['window']['pixels'] == 58240 and d['window']['different'] == 0
    assert d['window']['maximum_channel_difference'] == 0 and d['window']['bbox'] is None
    assert d['full']['different'] == 248681 and d['mask'] is None
    for suffix in ('-remake.png', '-diff.png'):
        read(prefix + suffix)
    comparisons.append(d)
for p in Path('workplace/scout79-r83-game-final/normal-save').iterdir():
    if p.is_file():
        read(p)
for p in ('internal/game/scoutfuel.go', 'internal/game/scoutfuel_test.go',
          'internal/game/battleownership.go', 'internal/game/battlefinal.go',
          'docs/internals/data/turn-window-panel.json', 'tools/scripts/scout_fuel_game_receipt.py',
          'tools/scripts/battle_ownership_player_entry.go', 'tools/turn_counter_gui.sh'):
    read(p)
out = Path('docs/internals/data/scout-fuel-game.json')
assert out.parent.stat().st_uid == os.getuid()
result = {'schema': 'logh3.scout-fuel-game.v1', 'parent_commit': original['parent_commit'],
          'original_receipt_sha256': sources['docs/internals/data/scout-fuel.json']['sha256'],
          'scope': '281原版控制案×兩種typed形態、48報告、兩矩形；正常公開派出及一次實行；後續第8步回收／冷卻／再派出與保存為診斷；首戰GUI兩倍率僅邏輯畫面回歸。',
          'standard': {'passed_packages': 13, 'without_tests': 5},
          'comparisons': comparisons, 'sources': sources}
with out.open('x', encoding='utf8') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write('\n')
print('產品来源', len(sources), 'SHA-256', hashlib.sha256(out.read_bytes()).hexdigest())
