"""第55輪固定WORK原版雙跑、完整保存及未變畫面；拒絕覆寫收據。"""
from collections import Counter
import hashlib
import json
from pathlib import Path
import re


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


previous_path = Path('docs/internals/data/battle-sprite-priority.json')
assert sha(previous_path) == '4540cb1002bf37cf7f927cd91349d3910cf59b66b4d06ee24c893a84768b5e29'
previous = json.loads(previous_path.read_text())
out = Path('docs/internals/data/battle-work-control.json')
assert not out.exists(), '固定收據不以新來源重生'
original_runs = []
watches = []
memory = None
for side in 'ab':
    p = Path('workplace/oracle/normal89-r55-work-'+side)
    m = (p/'entry.bin').read_bytes()
    assert len(m) == 1048576
    assert hashlib.sha256(m).hexdigest() == 'ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97'
    if memory is None:
        memory = m
    else:
        assert m == memory
        for n in ('screen.bin', 'screen.png', 'pal.bin'):
            assert (p/n).read_bytes() == (Path(original_runs[0]['path'])/n).read_bytes()
    assert 'f7523f9（未提交檔案 0 個）' in (p/'meta.txt').read_text()
    report = (p/'report.txt').read_text()
    rows = []
    for line in report.splitlines():
        match = re.fullmatch(r'\[watch\] #(\d+) ([0-9A-F]+): ([0-9A-F]{2}) → ([0-9A-F]{2})  ← ([0-9A-F]{4}):([0-9A-F]{4})', line)
        if match:
            step, a, old, new, cs, ip = match.groups()
            rows.append({'step': int(step), 'linear_hex': '0x'+a, 'old_hex': old, 'new_hex': new,
                         'post_write_ip': cs+':'+ip, 'grade': 'proven',
                         'source': str(p/'report.txt'), 'warning': 'IP是執行後位置，非寫入指令起點'})
    assert rows
    comparable = [{k: v for k, v in r.items() if k != 'source'} for r in rows]
    if watches:
        assert comparable == watches
    else:
        watches = comparable
    flags = (p/'flags.txt').read_text().splitlines()
    assert flags[flags.index('-poke')+1] == 'lin:5E00C@300000000=7E 16 DC D3'
    assert flags[flags.index('-dump-mem-at')+1].startswith('434999999:lin:0:1048576:')
    original_runs.append({'path': str(p), 'flags': flags,
                          'hashes': {n: sha(p/n) for n in ('entry.bin', 'machine.state', 'screen.bin', 'screen.png', 'pal.bin', 'report.txt', 'meta.txt', 'flags.txt')},
                          'writes': rows,
                          'register_capture_note': '12AA入口／Input、297B經費及1588鏡頭捕獲上限40；不由上限推定總次數'})
assert original_runs[0]['flags'] == [x.replace('normal89-r55-work-b', 'normal89-r55-work-a') for x in original_runs[1]['flags']]
ida_exports = []
all_rows = []
functions = []
for p in sorted(Path('workplace/ida').glob('normal89-r55-*-graded.json')):
    d = json.loads(p.read_text())
    assert d['kernel_version'] == '9.4' and d['uid'] == 1000 and d['ida_base_hex'] == '0x10200'
    for row in d['instructions']:
        a = int(row['runtime_linear_hex'], 16)
        b = bytes.fromhex(row['original_bytes_hex'])
        assert memory[a:a+len(b)] == b
    ida_exports.append({'path': str(p), 'sha256': sha(p), 'data': d})
    all_rows.extend(d['instructions'])
    functions.extend(d['functions'])
assert len(ida_exports) == 8 and len(all_rows) == 785 and len(functions) == 11
assert sum(r['grade'] == 'proven' for r in all_rows) == 56
players = [Path('workplace/round55-pass-'+side) for side in 'ab']
names = sorted(str(p.relative_to(players[0])) for p in players[0].rglob('*') if p.is_file())
assert len(names) == 55
assert names == sorted(str(p.relative_to(players[1])) for p in players[1].rglob('*') if p.is_file())
assert all((players[0]/n).read_bytes() == (players[1]/n).read_bytes() for n in names)
metadata = json.loads((players[0]/'probe.json').read_text())
assert metadata == previous['normal_player_metadata']
before = Path('workplace/round54-pass-a')
for n, expected in previous['normal_player_runs'][0]['hashes'].items():
    assert sha(before/n) == expected
changed_offsets = [2, 3, 6, 7, 0x26, 0x28, 0x2A, 0x2B, 0x2C, 0x2D, 0x6E, 0x6F, 0x72, 0x73]
for n in names:
    a, b = (before/n).read_bytes(), (players[0]/n).read_bytes()
    expected = changed_offsets if n.endswith('/WORK.DT0') and not n.startswith('loaded/') else []
    assert len(a) == len(b)
    assert [i for i, (x, y) in enumerate(zip(a, b)) if x != y] == expected, n
work = memory[0x5DA44:0x5DA44+230]
for phase in ('entry', 'flagship', 'squad', 'empty', 'panel'):
    assert (players[0]/phase/'WORK.DT0').read_bytes() == work
tables = {}
for name, address, stride, count in [('WORK', 0x5DA44, 230, 1), ('FLEET', 0x60BA4, 122, 32),
                                    ('SHIP', 0x668B2, 78, 352), ('STELLAR', 0x5F2CC, 126, 32), ('PLANET', 0x53132, 82, 81)]:
    a, b = memory[address:address+stride*count], (players[0]/'entry'/(name+'.DT0')).read_bytes()
    assert len(a) == len(b)
    ds = [i for i, (x, y) in enumerate(zip(a, b)) if x != y]
    tables[name] = {'bytes': len(a), 'different_bytes': len(ds), 'offset_counts': dict(Counter(i % stride for i in ds)),
                    'original_sha256': hashlib.sha256(a).hexdigest(), 'remake_sha256': hashlib.sha256(b).hexdigest()}
assert [tables[n]['different_bytes'] for n in tables] == [0, 27, 214, 0, 0]
assert set(tables['SHIP']['offset_counts']).issubset({0x38, 0x39, 0x3A, 0x3B})
frames = previous['frames']
for display, frame in frames.items():
    for suffix, expected in frame['unchanged_images_sha256'].items():
        assert sha(frame['unchanged_images_prefix']+suffix) == expected
    for n, expected in frame['original_hashes'].items():
        assert sha(Path(frame['original'])/n) == expected
    assert sha(players[0]/display/'screen.bin') == frame['after_screen_sha256']
gui = []
for scale in (1, 2):
    prefix = Path(f'workplace/round55-gui-{scale}x')
    assert Path(str(prefix)+'.bin').read_bytes() == (players[0]/'entry/screen.bin').read_bytes()
    gui.append({'scale': scale, 'prefix': str(prefix),
                'hashes': {s: sha(str(prefix)+s) for s in ('.bin', '.png', '.txt', '.log')},
                'classification': '正式cmd/gin3正常主選單、鍵盤138、S3帝國及真實滑鼠；不使用direct-entry'})
verification = {}
for label, packages in [('work', 1), ('limited', 3), ('standard', 18)]:
    p = Path(f'workplace/round55-{label}.log')
    s = p.read_text()
    assert 'FAIL' not in s and 'SKIP' not in s and 'no tests to run' not in s
    tested = len(re.findall(r'^ok\s', s, re.M))
    no_tests = len(re.findall(r'^\?\s.*\[no test files\]$', s, re.M))
    assert tested+no_tests == packages
    verification[label] = {'path': str(p), 'sha256': sha(p), 'pass_events': len(re.findall(r'^\s*--- PASS:', s, re.M)),
                           'packages': packages, 'tested_packages': tested, 'no_test_packages': no_tests, 'skips': 0}
baseline = Path('workplace/round55-before-work.log')
assert len(re.findall(r'^    --- FAIL:', baseline.read_text(), re.M)) == 2
assert baseline.read_text().count('位移：[2 3 6 7 26 28 2A 2B 2C 2D 6E 6F 72 73]') == 2
sources = ['internal/game/battlework.go', 'internal/game/battlework_test.go', 'internal/game/battle.go',
           'internal/game/rebellion.go', 'internal/game/game.go', 'internal/game/originalbrain.go',
           'internal/game/fleetarrival_test.go', 'internal/game/turncounter_test.go', 'internal/rules/save.go',
           'tools/battle_work_control_oracle.sh', 'tools/scripts/battle_work_control_ida.py',
           'tools/scripts/battle_work_control_receipt.py', 'tools/scripts/battle_normal_probe.go',
           'tools/scripts/battle_normal_probe_seed.py', 'tools/turn_counter_gui.sh',
           'workplace/round55-normal-overlay.json', 'workplace/round55-normal-seed.go',
           'workplace/round55-limited-command.sh']
data = {k: previous[k] for k in ('date', 'dosgolem_head', 'dosgolem_dirty', 'original_bundle_sha256', 'original_executable_sha256', 'toolchain')}
data.update({'schema': 'logh3.battle-work-control.v1', 'issue': 89,
             'parent_commit': 'b25ee53ca5cac21aa06bd8328162cca100b5e68b',
             'remake_worktree': 'dirty; 本輪WORK及經費修正、回歸與文件',
             'status': '限定正常首戰完整WORK／保存／重讀已驗，完整088仍DRAFT',
             'comparison_level': 'nearby-state; 不升格原畫面收據',
             'previous_receipts': {str(previous_path): sha(previous_path)},
             'original_parent': previous['original_parent'], 'original_runs': original_runs,
             'writes': watches, 'ida_exports': ida_exports, 'original_functions': functions,
             'instruction_count': len(all_rows), 'proven_semantics_count': 56,
             'original_work': {'linear_hex': '0x5DA44', 'bytes': 230, 'hex': work.hex(), 'sha256': hashlib.sha256(work).hexdigest()},
             'normal_player_metadata': metadata,
             'normal_player_runs': [{'path': str(p), 'hashes': {n: sha(p/n) for n in names}} for p in players],
             'changed_work_offsets': changed_offsets, 'table_comparison': tables, 'frames': frames, 'normal_gui': gui,
             'verification': verification,
             'before_regression': {'path': str(baseline), 'sha256': sha(baseline), 'failed_subcases': 2},
             'sources_sha256': {n: sha(n) for n in sources},
             'limits': ['鏡頭互換／狀態2／5及signed溢位測試屬指令契約的診斷回歸，不稱自然局面原版動態收據',
                        '同盟完整正常首戰雙跑與續存續戰仍未新增驗收',
                        'FLEET27及SHIP214仍為顯示指標差，不捏造DOS指標；FF字鍵及初始一像素待驗',
                        '完整088及八步#79維持原門檻；原版硬體時序與音樂裁決不重開']})
out.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print('收據SHA-256', sha(out), '限定驗證', verification['limited']['pass_events'], 'PASS，WORK 0／230')
