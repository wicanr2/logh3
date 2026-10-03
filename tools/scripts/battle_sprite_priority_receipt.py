"""第54輪圖示順位收據；tools/gui.sh xrun執行，不覆寫既有JSON。"""
from collections import Counter
import hashlib
import json
import os
from pathlib import Path
import re
import struct


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


out = Path('docs/internals/data/battle-sprite-priority.json')
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
previous_path = 'docs/internals/data/battle-scout-cache.json'
assert sha(previous_path) == '1203dc687f6f2481ac53acfaa3a2da8866d8d82638ebeda6a9dd452d875bf3d9'
previous = json.loads(Path(previous_path).read_text())
original_runs = []
roots = [Path('workplace/oracle/normal89-r54-init-'+side) for side in 'ab']
for root in roots:
    assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in (root/'meta.txt').read_text()
    files = ('entry.bin', 'screen.bin', 'screen.png', 'pal.bin', 'machine.state', 'flags.txt', 'meta.txt', 'report.txt')
    original_runs.append({'path': str(root), 'hashes': {n: sha(root/n) for n in files},
                          'command_argv': (root/'flags.txt').read_text().splitlines()})
for name in ('entry.bin', 'screen.bin', 'screen.png', 'pal.bin'):
    assert (roots[0]/name).read_bytes() == (roots[1]/name).read_bytes()
m = (roots[0]/'entry.bin').read_bytes()
assert sha(roots[0]/'entry.bin') == 'ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97'
assert len(m) == 1048576 and struct.unpack_from('<I', m, 0x5e00c)[0] == 0xb6e4b78a
traces = {}
for addr in ('25DB:0007', '25DB:0178', '25DB:01A8', '25DB:01B0', '25DB:0915'):
    captures = []
    for root in roots:
        text = (root/'report.txt').read_text()
        match = re.search(re.escape(addr)+r' 執行時的暫存器（前 (\d+) 次）：\n((?:  #.*\n      \[bp\+\].*\n)+)', text)
        assert match is not None
        captures.append({'count': int(match[1]), 'raw_registers_and_stack': match[2]})
    assert captures[0] == captures[1]
    traces[addr] = captures[0]
assert [traces[a]['count'] for a in traces] == [1, 55, 55, 55, 80]
pool_records = []
for ax, si, es in re.findall(r'AX=([0-9A-F]+).*SI=([0-9A-F]+).*ES=([0-9A-F]+)', traces['25DB:01B0']['raw_registers_and_stack']):
    address = int(es, 16)*16+int(si, 16)
    pool = (address-0x668b2)//78
    assert address == 0x668b2+pool*78 and int(ax, 16)&255 == 128
    r = m[address:address+78]
    offset, segment = struct.unpack_from('<HH', r, 0x38)
    object_address = segment*16+offset
    obj = m[object_address:object_address+40]
    assert r[0x30] == obj[3] == 128 and struct.unpack_from('<H', obj, 4)[0] == r[1]+(r[3]<<8)
    pool_records.append({'pool': pool, 'runtime_linear_hex': hex(address), 'ship_record_hex': r.hex(),
                          'object_far_pointer_offset_segment': [hex(offset), hex(segment)],
                          'object_runtime_linear_hex': hex(object_address), 'object_40_bytes_hex': obj.hex()})
assert len(pool_records) == len(set(r['pool'] for r in pool_records)) == 55
assert sorted(r['pool'] for r in pool_records) == previous['battle_pools']
ida_path = Path('workplace/ida/normal89-r54-25DB-graded.json')
ida = json.loads(ida_path.read_text())
assert ida['kernel_version'] == '9.4' and ida['uid'] == 1000 and len(ida['functions']) == 3
assert len(ida['instructions']) == 873 and sum(r['grade'] == 'proven' for r in ida['instructions']) == 19
assert sha(ida['input'].replace('/work/', 'workplace/ida/')) == ida['input_sha256']
for row in ida['instructions']:
    a = int(row['runtime_linear_hex'], 16)
    raw = bytes.fromhex(row['original_bytes_hex'])
    assert m[a:a+len(raw)] == raw
players = [Path('workplace/round54-pass-'+side) for side in 'ab']
names = sorted(p.relative_to(players[0]).as_posix() for p in players[0].rglob('*') if p.is_file())
assert len(names) == 55
assert names == sorted(p.relative_to(players[1]).as_posix() for p in players[1].rglob('*') if p.is_file())
assert all((players[0]/n).read_bytes() == (players[1]/n).read_bytes() for n in names)
metadata = json.loads((players[0]/'probe.json').read_text())
assert metadata == previous['normal_player_metadata']
before = Path('workplace/round53-pass-a')
for n, expected in previous['normal_player_runs'][0]['hashes'].items():
    assert sha(before/n) == expected
changes = sorted(r['pool']*78+0x30 for r in pool_records)
for n in names:
    a, b = (before/n).read_bytes(), (players[0]/n).read_bytes()
    expected = changes if n.endswith('/SHIP.DT0') and not n.startswith('loaded/') else []
    assert [i for i, (x, y) in enumerate(zip(a, b)) if x != y] == expected, n
    assert len(a) == len(b)
tables = {}
for name, address, stride, count in [('WORK', 0x5da44, 230, 1), ('FLEET', 0x60ba4, 122, 32),
                                    ('SHIP', 0x668b2, 78, 352), ('STELLAR', 0x5f2cc, 126, 32), ('PLANET', 0x53132, 82, 81)]:
    a, b = m[address:address+stride*count], (players[0]/'entry'/(name+'.DT0')).read_bytes()
    assert len(a) == len(b)
    ds = [i for i, (x, y) in enumerate(zip(a, b)) if x != y]
    tables[name] = {'bytes': len(a), 'different_bytes': len(ds), 'offset_counts': dict(Counter(i % stride for i in ds)),
                    'original_sha256': hashlib.sha256(a).hexdigest(), 'remake_sha256': hashlib.sha256(b).hexdigest()}
assert [tables[n]['different_bytes'] for n in tables] == [14, 27, 214, 0, 0]
assert set(tables['SHIP']['offset_counts']).issubset({0x38, 0x39, 0x3a, 0x3b})
frames = previous['frames']
for display, frame in frames.items():
    for suffix, expected in frame['unchanged_images_sha256'].items():
        assert sha(frame['unchanged_images_prefix']+suffix) == expected
    for name, expected in frame['original_hashes'].items():
        assert sha(Path(frame['original'])/name) == expected
    frame['before_screen_sha256'] = sha(before/display/'screen.bin')
    frame['after_screen_sha256'] = sha(players[0]/display/'screen.bin')
    assert frame['before_screen_sha256'] == frame['after_screen_sha256']
gui = []
for scale in (1, 2):
    prefix = f'workplace/round54-gui-{scale}x'
    assert Path(prefix+'.bin').read_bytes() == (players[0]/'entry/screen.bin').read_bytes()
    gui.append({'scale': scale, 'prefix': prefix, 'hashes': {s: sha(prefix+s) for s in ('.bin', '.png', '.txt', '.log')},
                'classification': '正式cmd/gin3真實X11正常入口，邏輯畫面與Game相同'})
verification = {}
for label in ('priority', 'limited', 'standard'):
    p = Path(f'workplace/round54-{label}.log')
    text = p.read_text()
    assert 'FAIL' not in text and 'SKIP' not in text
    tested = len(re.findall(r'^ok\s', text, re.M))
    no_tests = len(re.findall(r'^\?\s.*\[no test files\]$', text, re.M))
    assert tested+no_tests == (18 if label == 'standard' else 3 if label == 'limited' else 2)
    verification[label] = {'path': str(p), 'sha256': sha(p), 'pass_events': len(re.findall(r'^\s*--- PASS:', text, re.M)),
                           'packages': tested+no_tests, 'tested_packages': tested, 'no_test_packages': no_tests, 'skips': 0}
baseline = Path('workplace/round54-before-priority.log')
assert '--- FAIL: TestBattleUnitPriorityPrecedesOrdinal/priority' in baseline.read_text()
assert len(re.findall(r'^    --- FAIL:', baseline.read_text(), re.M)) == 1
sources = ['internal/rules/battle.go', 'internal/game/battle.go', 'internal/game/battlemarker.go',
           'internal/game/fleetarrival_test.go', 'internal/screen/battle.go', 'internal/screen/battleunits_test.go',
           'tools/battle_sprite_priority_oracle.sh', 'tools/scripts/battle_sprite_priority_ida.py',
           'tools/scripts/battle_sprite_priority_receipt.py', 'tools/scripts/battle_normal_probe.go',
           'tools/scripts/battle_normal_probe_seed.py', 'tools/turn_counter_gui.sh',
           'workplace/round54-normal-overlay.json', 'workplace/round54-normal-seed.go',
           'workplace/round54-before-screen.go', 'workplace/round54-before-screen-overlay.json']
probe = Path('workplace/explore/normal89-r54-low-first-explore')
data = {k: previous[k] for k in ('date', 'dosgolem_head', 'dosgolem_dirty', 'original_bundle_sha256', 'original_executable_sha256', 'toolchain')}
data.update({'schema': 'logh3.battle-sprite-priority.v1', 'issue': 89,
             'parent_commit': '1ce5c23ef2bab576eb21971fa81e1896dbe49d6a', 'remake_worktree': 'dirty; 本輪順位修正及回歸',
             'status': '限定初始化／順位來源／保存已驗，完整088仍DRAFT', 'comparison_level': 'nearby-state; 不升格前輪',
             'previous_receipts': {previous_path: sha(previous_path)},
             'original_parent': json.loads(Path('workplace/round54-parent.json').read_text()),
             'original_runs': original_runs, 'initialization_traces': traces, 'pool_records': pool_records,
             'ida': {'path': str(ida_path), 'sha256': sha(ida_path), 'export': ida},
             'normal_player_metadata': metadata,
             'normal_player_runs': [{'path': str(p), 'hashes': {n: sha(p/n) for n in names}} for p in players],
             'changed_ship_byte_offsets': changes, 'tables': tables, 'frames': frames, 'gui': gui, 'verification': verification,
             'before_regression': {'path': str(baseline), 'sha256': sha(baseline), 'failed_subcases': 1},
             'diagnostic_attempt': {'path': str(probe), 'classification': 'unknown presentation; 未觸發重畫，不算像素對拍',
                                    'command_argv': (probe/'flags.txt').read_text().splitlines(),
                                    'hashes': {n: sha(probe/n) for n in ('end.bin', 'screen.bin', 'screen.png', 'pal.bin', 'flags.txt', 'report.txt')}},
             'source_hashes': {p: sha(p) for p in sources},
             'limitations': ['SHIP214及FLEET27剩餘差異仍是原版顯示物件指標，未造DOS指標',
                             'WORK14與初始全圖一像素、一般FF字鍵仍待證',
                             '非128順位回歸只證明remake比較器實作，不稱新增原版same-state畫面',
                             '五張PNG／差分圖沿用逐byte核對相同的第52輪影像；正式GUI另行重跑']})
out.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print('收據SHA-256', sha(out), '限定驗證', verification['limited']['pass_events'], 'PASS')
