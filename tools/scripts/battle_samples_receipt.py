"""第35輪正常玩家抽樣收據；入口為 docs/playtest/045-battle-movement-render.md。

僅核對 remake 正常路徑，沒有原版同狀態比較或對拍通過聲明。
"""
import hashlib
import json
from pathlib import Path
import re
import sys

from pngout import write_png

ROOT = Path('/ref')
WORK = Path('/work/workplace')


def identity(path):
    data = path.read_bytes()
    return {'bytes': len(data), 'sha256': hashlib.sha256(data).hexdigest()}


def main():
    head, dirty = sys.argv[1:]
    if not re.fullmatch(r'[0-9a-f]{40}', head) or dirty not in ('clean', 'dirty'):
        raise ValueError('須記錄完整父提交及工作樹狀態')
    target = (WORK / 'battle79-round35-normal.log').read_text()
    if 'FAIL' in target or 'SKIP' in target:
        raise ValueError('正常玩家目標驗證失敗或跳過')
    pattern = r'正常S(\d)：星域(\d+)、戰略回合(\d+)、(\d+)艦隊／(\d+)分隊；受損(\d+)／玩家受損(\d+)／補給艦(\d+)；結果幀補給(\d+)／修理(\d+)／移動(\d+)；Input還原，SHIP (\d+)／FLEET (\d+) bytes存讀相同；seed戰略([0-9A-F]+)、戰術12345678'
    rows = re.findall(pattern, target)
    if len(rows) != 3 or [r[0] for r in rows] != ['1', '2', '3']:
        raise ValueError('三份隨附存檔正常驗證不完整')
    standard = (WORK / 'battle79-round35-standard-final.log').read_text()
    lines = [s for s in standard.splitlines() if s.startswith(('ok ', '? '))]
    if len(lines) != 18 or 'FAIL' in standard:
        raise ValueError('18個標準產品套件未完成')
    tool_test = (WORK / 'battle79-round35-evidence-final.log').read_text()
    if 'FAIL' in tool_test or 'SKIP' in tool_test or '--- PASS: TestIssueEvidencePaths ' not in tool_test:
        raise ValueError('共用證據解析器回歸不完整')
    receipt = {
        'scope': 'remake-normal-player-path-only',
        'comparison_grade': None,
        'original_parity_pass': False,
        'parent_git_head': head, 'worktree': dirty,
        'canvas': [640, 480], 'scaling': 'none', 'masking': 'none',
        'images': {'gui': 'logh3-ebiten:2@sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
                   'go': 'golang:1.24-bookworm@sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac'},
        'seeds': {'strategy_before_execute': 'D3DC167E', 'battle_before_execute': '12345678',
                  'method': '在各自首次實行前 NewRand 固定，不重擲；GUI沿產品預設，另列'},
        'cases': [], 'frames': {}, 'logs': {}, 'sources': {}, 'original_inputs': {},
        'limitations': ['原版第二戰場同狀態未驗', '正常補給與修理結果幀均為零',
                        '不是全部玩家存檔驗收', '不是完整八步或硬體時長驗收'],
    }
    for r in rows:
        numbers = list(map(int, r[:13]))
        keys = ['save', 'region', 'strategy_turn', 'fleets', 'units', 'damaged',
                'human_damaged', 'supply_ships', 'supply_frames', 'repair_frames',
                'movement_frames', 'ship_bytes', 'fleet_bytes']
        receipt['cases'].append(dict(zip(keys, numbers)))
    for label in ['s1-entry', 's1-movement', 's1-input', 's2-entry', 's2-input', 's3-entry', 's3-input']:
        stem = 'battle79-round35-normal-' + label
        pix = (WORK / (stem + '.bin')).read_bytes()
        dac = (WORK / (stem + '.pal')).read_bytes()
        if len(pix) != 640*480 or len(dac) != 16*3 or max(pix) > 15 or max(dac) > 63:
            raise ValueError('原生色號或DAC形狀不符：' + stem)
        write_png(WORK / (stem + '.png'), 640, 480, pix, [v << 2 | v >> 4 for v in dac])
        receipt['frames'][label] = {ext: identity(WORK / (stem + '.' + ext)) for ext in ['bin', 'pal', 'png']}
    for name in ['battle79-round35-s1-gui-movement', 'battle79-round35-s2-gui-input-corrected']:
        receipt['frames'][name] = {ext: identity(WORK / (name + '.' + ext)) for ext in ['bin', 'png']}
    if 'state=StateBattle' not in (WORK / 'battle79-round35-s2-gui-input-corrected.state').read_text():
        raise ValueError('正式S2未在戰場')
    if (WORK / 'battle79-round35-s2-gui-input-corrected.bin').read_bytes() == (WORK / 'battle79-round35-s2-gui-input.bin').read_bytes():
        raise ValueError('S2修正腳本仍停在舊第1回合畫面')
    for name in ['normal', 'standard', 's1-gui', 's2-gui', 's2-gui-final', 's2-gui-corrected',
                 'production-probe', 'production-probe-fixed', 'resource-probe', 'later-probe',
                 'standard-final', 'evidence-before', 'evidence-after', 'evidence-final']:
        receipt['logs'][name] = identity(WORK / ('battle79-round35-' + name + '.log'))
    receipt['s2_script_correction'] = '正式bootstrap先停OccupationReport，第一次點擊至TurnReport，第二次至Battle第1回合；另外兩次選戰術實行才到第2回合。失敗截圖與探測log保留；探測uint8誤讀只屬腳本失敗。'
    for path in ['internal/game/battlesamples_test.go', 'internal/game/battle.go',
                 'internal/game/battlepreparation.go', 'internal/game/battlemovement.go',
                 'internal/game/game.go', 'cmd/gin3/main.go', 'internal/bootstrap/load.go',
                 'cmd/worklist/issues.go', 'cmd/worklist/issues_test.go',
                 'docs/internals/data/battle-menu.json', 'docs/internals/data/screen-battle.json',
                 'docs/spec/079-battle-preparation.md', 'docs/spec/080-battle-movement-phase.md',
                 'docs/spec/083-battle-phase-panel.md', 'tools/scripts/battle_samples_receipt.py']:
        receipt['sources'][path] = identity(ROOT / path)
    for ext in ['DT8', 'DT9', 'DTA']:
        for name in ['STELLAR', 'PLANET', 'ADMIRAL', 'FLEET', 'SHIP', 'WORK']:
            receipt['original_inputs'][name + '.' + ext] = identity(Path('/orig') / (name + '.' + ext))
    for name in ['PLADAT.DAT', 'GRAPH.IMG', 'FRAME.BLK', 'ANI_PNL0.A98', 'BLOCK_2.BLK']:
        receipt['original_inputs'][name] = identity(Path('/orig') / name)
    original = WORK / 'oracle/battle79-round35-s1-explore'
    receipt['original_exploration'] = {
        'scope': '探索，仍在戰略移動過場；不登錄第二戰場或same-state通過',
        'executor_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
        'parent_state': json.loads((WORK / 'battle79-round35-original-parent.json').read_text()),
        'files': {n: identity(original / n) for n in ['meta.txt', 'flags.txt', 'report.txt', 'screen.png', 'screen.bin', 'pal.bin', 'mem.bin']},
        'flags': (original / 'flags.txt').read_text().splitlines(),
    }
    out = WORK / 'battle79-round35-normal-receipt.json'
    out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps({'receipt': out.name, **identity(out), 'cases': receipt['cases']}, ensure_ascii=False))


if __name__ == '__main__':
    main()
