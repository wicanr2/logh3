"""089：固定已驗證的正式保持成果；完整比較、不覆寫收據。"""
import hashlib
import json
import subprocess
from pathlib import Path
from pngout import game_palette, write_png

ROOT = Path('/src')
OUT = ROOT / 'docs/internals/data/left-button-repeat-game.json'
DIFF = ROOT / 'workplace/round58-hold-differences'
assert not OUT.exists() and not DIFF.exists(), '拒絕覆寫固定收據或差分'
assert ROOT.stat().st_uid == 1000 and OUT.parent.stat().st_uid == 1000


def record(path):
    data = path.read_bytes()
    return {'path': str(path.relative_to(ROOT)), 'bytes': len(data),
            'sha256': hashlib.sha256(data).hexdigest()}


frozen = []
for name, expected in [
        ('left-button-repeat.json', '4873d0449777bd1ae6be2e558c82493b582c32f38f96d8d1e20d96e4715dfc5a'),
        ('battle-data-scroll.json', '0da250722c37fa5c51518f7b932d29e7a800fd60bab54ac2d7bbb8066a279819')]:
    r = record(ROOT / 'docs/internals/data' / name)
    assert r['sha256'] == expected
    frozen.append(r)

monthly = []
for name in ['parent', 'a', 'b', 'close-a', 'close-b']:
    p = ROOT / 'workplace/oracle' / ('hold60-r58-month-clean-' + name)
    assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in (p / 'meta.txt').read_text()
    assert len((p / 'mem.bin').read_bytes()) == 1048576
    assert (p / 'mem.bin').read_bytes()[0x54b30 + 0x1bf2:0x54b30 + 0x1bf2 + 38].hex() == (
        '000111100100000000000000110101010101000000000000111001010102000000000000ffff')
    monthly.append({'name': name, 'flags': (p / 'flags.txt').read_text().splitlines(),
                    'artifacts': [record(p / n) for n in ['meta.txt', 'flags.txt', 'report.txt',
                        'mem.bin', 'machine.state', 'screen.bin', 'screen.png', 'pal.bin', 'writes.txt']]})
for pair in [('a', 'b'), ('close-a', 'close-b')]:
    for n in ['mem.bin', 'screen.bin', 'screen.png', 'pal.bin']:
        a, b = [ROOT / 'workplace/oracle' / ('hold60-r58-month-clean-' + v) / n for v in pair]
        assert a.read_bytes() == b.read_bytes(), (pair, n)

names = ['tab-300000', 'star-300000', 'tax-3000000', 'pct-3000000', 'month', 'last', 'zero']
frames = []
DIFF.mkdir()
for name in names:
    if name in ['last', 'zero']:
        originals = [ROOT / 'workplace/oracle' / ('data67-r57-' + name + '-' + v) for v in ['a', 'b']]
        players = [ROOT / 'workplace' / ('round58-hold-player-final-' + v) / name for v in ['a', 'b']]
        origin = '正式bootstrap／Game，主選單138→S3同盟→實行808 tick→首戰資料'
        inputs = {'cursor_canvas': [440, 390], 'hold_frames': 300, 'tick_ns': 16666666,
                  'arrow': '上' if name == 'zero' else '下', 'target_offset': 0 if name == 'zero' else 26}
    elif name == 'month':
        originals = [ROOT / 'workplace/oracle' / ('hold60-r58-month-clean-' + v) for v in ['a', 'b']]
        players = [ROOT / 'workplace' / ('round58-month-player-final-' + v) for v in ['a', 'b']]
        origin = '實際原版素材的測試資源，正常新遊戲／月初建造計劃；未注入局面'
        inputs = {'original_hold_instructions': 300000, 'game_first_arrow': [608, 390],
                  'game_hold_ns': 350000000, 'cursor_canvas': [617, 390], 'target_offset': 2}
    else:
        originals = [ROOT / 'workplace/oracle' / ('hold60-r58-' + name + '-' + v) for v in ['a', 'b']]
        players = [ROOT / 'workplace' / ('round58-hold-compare-final-' + v) / name for v in ['a', 'b']]
        origin = '實際原版素材的測試資源，正常新遊戲／戰略命令；未注入局面'
        steps = 5 if name.startswith('tab') else 2
        inputs = {'original_hold_instructions': int(name.split('-')[1]),
                  'game_hold_ns': 650000000 if steps == 5 else (
                      350000000 if name.startswith('star') else 8 * 10 * 2048 * 1000000000 // 1193182),
                  'updates': steps}
    original, got = (originals[0] / 'screen.bin').read_bytes(), (players[0] / 'screen.bin').read_bytes()
    assert len(original) == len(got) == 307200 and original == got, name
    for n in ['mem.bin', 'screen.bin', 'screen.png', 'pal.bin']:
        assert (originals[0] / n).read_bytes() == (originals[1] / n).read_bytes(), (name, n)
    files = sorted(p.name for p in players[0].iterdir() if p.is_file())
    assert files == sorted(p.name for p in players[1].iterdir() if p.is_file())
    for f in files:
        assert (players[0] / f).read_bytes() == (players[1] / f).read_bytes(), (name, f)
    assert game_palette(originals[0] / 'screen.png') == game_palette(players[0] / 'screen.png')
    if name in ['last', 'zero']:
        rectangle = json.loads((ROOT / 'docs/internals/data/battle-data.json').read_text())['window']
    elif name == 'month' or name.startswith(('tax', 'pct')):
        source = 'screen-admiral-list' if name == 'month' else 'screen-tax-plan'
        rectangle = json.loads((ROOT / 'docs/internals/data' / (source + '.json')).read_text())['windows']['overlay']
    elif name.startswith('star'):
        rectangle = json.loads((ROOT / 'docs/internals/data/list-screen-panel.json').read_text())['content']
    else:
        source = json.loads((ROOT / 'docs/internals/data/fleet-list-panel.json').read_text())
        rectangle = {'x': source['row']['x'], 'y': source['row']['top'],
                     'w': source['row']['w'], 'h': source['pitch'] * source['rows']}
    x, y, w, h = [rectangle[k] for k in ['x', 'y', 'w', 'h']]
    region = [yy * 640 + xx for yy in range(y, y + h) for xx in range(x, x + w)]
    assert not any(original[i] != got[i] for i in region)
    if name.startswith(('tax', 'pct')):
        work = (players[0] / 'WORK.DT0').read_bytes()
        assert work[0xc0] == (4 if name.startswith('tax') else 2)
        assert work[0x7e] == (27 if name.startswith('pct') else 25)
    # 差分由雙方色號計算；完整畫面零差才產生固定收據。
    diff = bytes(15 if x != y else 0 for x, y in zip(original, got))
    dp = DIFF / (name + '.png')
    write_png(dp, 640, 480, diff, [0, 0, 0] * 15 + [255, 255, 255])
    frames.append({'name': name, 'comparison_level': 'same-state，限核定等價輸入後的靜態全圖',
                   'resource_scope': origin, 'input_mapping': inputs,
                   'original_artifacts': [record(p / n) for p in originals for n in ['screen.bin', 'screen.png']],
                   'game_artifacts': [record(p / n) for p in players for n in files],
                   'statistics': {'region': [0, 0, 640, 480], 'pixels': 307200, 'equal': 307200,
                                  'different': 0, 'max_channel_difference': 0, 'difference_bbox': None},
                   'active_region_statistics': {'region': [x, y, w, h], 'pixels': len(region),
                        'equal': len(region), 'different': 0, 'max_channel_difference': 0, 'difference_bbox': None},
                   'difference_png': record(dp), 'allowed_pixel_regions': []})

gui = []
for scale in [1, 2]:
    base = ROOT / 'workplace' / ('round58-left-hold-gui-final-' + str(scale) + 'x')
    assert base.with_suffix('.bin').read_bytes() == (ROOT / 'workplace/round58-hold-player-final-a/last/screen.bin').read_bytes()
    state = base.with_suffix('.txt').read_text()
    assert 'state=StateBattle\n' in state and f'window={640*scale}x{480*scale}\n' in state
    assert 'cursor=440,390\n' in state
    gui.append({'scale': scale, 'normal_input': '真實X11鍵盤138、S3同盟、實行、就緒觀測、開資料、左鍵保持5秒、放開',
                'artifacts': [record(base.with_suffix(s)) for s in ['.bin', '.png', '.txt', '.log']],
                'different_pixels': 0})

logs = []
for name in ['round58-hold-standard-final.log', 'round58-hold-receipt-final-a.log',
             'round58-hold-receipt-final-b.log', 'round58-left-hold-gui-final.log']:
    p = ROOT / 'workplace' / name
    text = p.read_text()
    assert 'FAIL' not in text and 'SKIP' not in text
    if name == 'round58-hold-standard-final.log':
        assert sum(line.startswith(('ok ', '? ')) for line in text.splitlines()) == 18
    logs.append(record(p))

sources = [
    'internal/game/lefthold.go', 'internal/game/lefthold_test.go', 'internal/game/lefthold_normal_test.go',
    'internal/game/lefthold_oracle_test.go', 'internal/game/game.go', 'internal/game/game_test.go',
    'internal/game/fleetadj.go', 'internal/game/staffpick.go', 'internal/game/listtabs.go',
    'internal/rules/admiral.go', 'cmd/gin3/main.go', 'tools/left_hold_gui.sh',
    'tools/scripts/left_hold_gui_ready.py', 'tools/monthly_admiral_hold_oracle.sh',
    'tools/scripts/monthly_admiral_cursor.py', 'tools/scripts/mainscreen_spec.py',
    'docs/internals/data/screen-admiral-list.json', 'tools/scripts/left_hold_game_receipt.py']
data = {'schema': 'logh3.left-hold-game.v1', 'date': '2026-10-03',
        'parent_git_head': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip(),
        'working_tree_dirty': True, 'go_version': subprocess.check_output(['go', 'version'], text=True).strip(),
        'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
        'gui_image': 'logh3-ebiten:2', 'gui_image_sha256': '6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
        'approved_list_timing': {'first_update': '按下立即一列', 'delay_ms': 350, 'period_ms': 100,
                                 'classification': '使用者核准的刻意差異'},
        'economic_period_ns': 8 * 10 * 2048 * 1000000000 // 1193182,
        'economic_timing_grade': 'hardware-spec approximation；不宣稱原版初始相位或逐週期一致',
        'frozen_research': frozen, 'monthly_original_runs': monthly, 'frames': frames, 'gui': gui,
        'logs': logs, 'sources': [record(ROOT / p) for p in sources],
        'cursor_cleanup': json.loads((ROOT / 'workplace/round58-month-cursor-cleanup.json').read_text()),
        'limitations': ['未比較remake與原版整份RAM；保留既有指標差',
                        '經濟移出／返回的remake回歸不等於原版逐相位對拍',
                        'GUI抽樣限定正常首戰資料窗兩種倍率；其餘清單由Game及既有回歸核對',
                        '月初清單其他控制、BuildPlan十四顆造艦比例箭頭及其他八步不由本輪升格',
                        '第56輪FF字鍵裁決仍pending，音樂聆聽／macOS／正式交付仍缺']}
OUT.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')
print(json.dumps({'output': record(OUT), 'frames': len(frames), 'gui_scales': len(gui)}, ensure_ascii=False))
