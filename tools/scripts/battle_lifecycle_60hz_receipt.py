"""第102輪完整畫面、60 Hz觀測、公開存讀及來源身份核對。"""
import argparse
import hashlib
import json
import os
from pathlib import Path

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('--freeze', action='store_true')
p.add_argument('--check', action='store_true')
a = p.parse_args()
assert a.freeze != a.check
target = Path('docs/internals/data/battle-lifecycle-60hz.json')


def contents(folder):
    return {str(f.relative_to(folder)): f.read_bytes() for f in folder.rglob('*') if f.is_file()}


if a.freeze:
    assert not target.exists()
    first = Path('workplace/lifecycle79-r102-a-plan')
    second = Path('workplace/lifecycle79-r102-b-plan')
    files = contents(first)
    assert files == contents(second)
    plan = json.loads((first/'plan.json').read_text())
    assert len(plan['actions']) == 1032
    assert [r['kind'] for r in plan['lifecycle_reports']] == ['annihilation', 'victory']
    seeds = [r for r in plan['actions'] if r['kind'] == 'seed']
    assert len(seeds) == 1 and seeds[0]['seed'] == 0x12345678
    gui = Path('workplace/lifecycle79-r102-a-gui')
    receipt = json.loads((gui/'receipt.json').read_text())
    assert receipt['test_tps'] == 60 and receipt['natural_annihilation'] and receipt['natural_finish']
    assert receipt['public_save_load_equal'] and receipt['six_diagnostic_saves_equal']
    assert len(receipt['actions']) == len(plan['actions'])
    assert len(receipt['comparisons']) == 63
    assert all(r['pixels'] == 307200 and r['x11_different'] == 0 and r['game_index_equal']
               for r in receipt['comparisons'])
    updates = 0
    for line in (gui/'updates.jsonl').open():
        assert json.loads(line)['tps'] == 60
        updates += 1
    assert updates > 0
    saves = sorted((first/'player-save').glob('*.DT0'))
    assert len(saves) == 6
    for f in saves:
        assert f.read_bytes() == (first/f.name).read_bytes()
        assert f.read_bytes() == (gui/'player-save'/f.name).read_bytes()
        assert f.read_bytes() == (gui/f.name).read_bytes()
    reports = [r['state']['lifecycle_report'] for r in receipt['actions'] if r['kind'] == 'wait_report']
    assert reports == plan['lifecycle_reports']
    # overlay只追加觀測。沒有拿歷史產品正文覆蓋現在的正式程式。
    for revision in ('a', 'b'):
        stem = Path('workplace')/('lifecycle79-r102-'+revision)
        game = Path(str(stem)+'-game.go').read_text()
        assert game.startswith(Path('internal/game/fleetcombat.go').read_text())
        main = Path(str(stem)+'-main.go').read_text()
        assert 'ebiten.SetTPS' not in main
    tests = Path('workplace/round102-standard-gui.log').read_text()
    assert sum(line.split()[:1] == ['ok'] for line in tests.splitlines()) == 13
    assert sum('[no test files]' in line for line in tests.splitlines()) == 5
    assert '\nFAIL' not in tests
    parent_name = 'docs/internals/data/battle-lifecycle-player.json'
    parent = json.loads(Path(parent_name).read_text())
    paths = {parent_name, 'tools/scripts/battle_lifecycle_60hz.py',
             'tools/scripts/battle_lifecycle_60hz_receipt.py',
             'tools/scripts/sound_tail_input.go', 'tools/scripts/fleet_player_gui_capture.go',
             'docs/goal/2026-10-05-battle-lifecycle-60hz-round-102.md',
             'docs/playtest/090-battle-lifecycle-60hz.md', 'docs/spec/110-battle-finish.md',
             'workplace/round102-standard-gui.log', 'workplace/round79-capture',
             'workplace/order79-r93-input'}
    for name in ('workplace/lifecycle79-r95-gui-game.go', 'workplace/lifecycle79-r95-gui-main.go',
                 'workplace/lifecycle79-r95-gui-plan.go', 'workplace/lifecycle79-r95-gui-driver.py',
                 'workplace/lifecycle79-r95-normal-b/normal-1/inputs.json'):
        assert hashlib.sha256(Path(name).read_bytes()).hexdigest() == parent['sources'][name]['sha256']
        paths.add(name)
    for folder in ('internal', 'cmd'):
        paths.update(str(f) for f in Path(folder).rglob('*.go'))
    paths.update(('go.mod', 'go.sum', 'tools/gui.sh'))
    for f in Path('workplace').glob('lifecycle79-r102-*'):
        if f.is_file():
            paths.add(str(f))
        else:
            paths.update(str(v) for v in f.rglob('*') if v.is_file())
    sources = {}
    for name in sorted(paths):
        data = Path(name).read_bytes()
        sources[name] = {'bytes': len(data), 'sha256': hashlib.sha256(data).hexdigest()}
    result = {'schema': 'logh3.battle-lifecycle-60hz.v1', 'date': '2026-10-05', 'issue': 79,
              'parent_commit': 'b5ac935d2e77d45d82e1173d340e26fb116f03d8',
              'engine_worktree': '正式產品未修改，新增驗證工具與文件；完整來源逐項凍結',
              'build_image': 'logh3-ebiten:2', 'logical_canvas': [640, 480], 'window_scale': 1,
              'original': {'bundle': 'org_game/银河英雄传说3SP.zip',
                           'bundle_sha256': 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
                           'executable': 'org_game/extracted/GIN3.COM',
                           'executable_sha256': '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918',
                           'new_oracle_run': False},
              'commands': [
                  'bash tools/gui.sh xrun python3 tools/scripts/battle_lifecycle_60hz.py a',
                  'bash tools/gui.sh xrun python3 tools/scripts/battle_lifecycle_60hz.py b',
                  'bash tools/gui.sh xrun python3 workplace/lifecycle79-r102-a-driver.py --plan workplace/lifecycle79-r102-a-plan/plan.json --out workplace/lifecycle79-r102-a-gui --scale 1 --binary /src/workplace/lifecycle79-r102-a-gin3 --input-binary /src/workplace/order79-r93-input',
                  "bash tools/gui.sh xrun bash -c 'go test ./internal/... ./cmd/... -count=1 > workplace/round102-standard-gui.log 2>&1'"],
              'classification': 'remake正常玩家流程；未新增原版整回合同態',
              'initial_seed': 0x12345678, 'state_injection': False,
              'bootstrap_repeats': 2, 'repeat_files_equal': len(files),
              'gui': {'tps': 60, 'updates_checked': updates, 'actions': len(receipt['actions']),
                      'frames': len(receipt['comparisons']), 'pixels': 63*307200, 'different': 0,
                      'camera_gestures': len(receipt['camera_gestures']),
                      'public_save_load_equal': True, 'storage_bytes': sum(f.stat().st_size for f in saves)},
              'limitations': ['未新增占領、非旗艦自然死亡／空組或自然升級GUI',
                              '額外鏡頭回授有完整記錄，不是相同實體輸入時序',
                              '單人戰術CPU沿061保持停留，沒有改AI設定',
                              '未新增原版正常整回合同態；062 DRAFT及#79 OPEN'],
              'sources': sources}
    for key in ('bundle', 'executable'):
        source = Path(result['original'][key])
        assert hashlib.sha256(source.read_bytes()).hexdigest() == result['original'][key+'_sha256']
    with target.open('x') as stream:
        stream.write(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
index = json.loads(target.read_text())
for name, want in index['sources'].items():
    data = Path(name).read_bytes()
    assert len(data) == want['bytes'] and hashlib.sha256(data).hexdigest() == want['sha256'], name
print(len(index['sources']), '來源身份符合，SHA-256', hashlib.sha256(target.read_bytes()).hexdigest())
