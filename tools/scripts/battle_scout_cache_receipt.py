"""第53輪快取限定收據；以tools/gui.sh xrun執行，既有收據保持。

python3 tools/scripts/battle_scout_cache_receipt.py
"""
from collections import Counter
import hashlib
import json
import os
from pathlib import Path
import re
import struct


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def checked_json(path, expected):
    assert sha(path) == expected, path
    return json.loads(Path(path).read_text())


out = Path('docs/internals/data/battle-scout-cache.json')
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
previous_path = 'docs/internals/data/battle-hover-marks.json'
previous = checked_json(previous_path, '91ec1901d3232702c3612d4edd459d72ff60c4633483bc578339b0e3f5c56250')
original_path = 'docs/internals/data/battle-input-display.json'
original = checked_json(original_path, 'c1f387c1e3729450f4b7b3f9f8babb21b0b8f49d1b0762b1f6e0cca648cfb5c6')
assert original['dosgolem_head'] == previous['dosgolem_head']
original_runs = original['original_runs']
for kind, runs in original_runs.items():
    for run in runs:
        for name, expected in run['hashes'].items():
            assert sha(Path(run['path'])/name) == expected
    for name in ('screen.bin', 'screen.png', 'pal.bin'):
        assert (Path(runs[0]['path'])/name).read_bytes() == (Path(runs[1]['path'])/name).read_bytes()
ram = Path(original_runs['base'][0]['path'])/'entry.bin'
assert sha(ram) == 'ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97'
m = ram.read_bytes()
assert len(m) == 1048576
assert m == (Path(original_runs['base'][1]['path'])/'entry.bin').read_bytes()
assert struct.unpack_from('<I', m, 0x5e00c)[0] == 0xb6e4b78a
players = [Path('workplace/round53-pass-'+s) for s in 'ab']
names = sorted(p.relative_to(players[0]).as_posix() for p in players[0].rglob('*') if p.is_file())
assert len(names) == 55
assert names == sorted(p.relative_to(players[1]).as_posix() for p in players[1].rglob('*') if p.is_file())
assert all((players[0]/n).read_bytes() == (players[1]/n).read_bytes() for n in names)
metadata = json.loads((players[0]/'probe.json').read_text())
assert metadata == previous['normal_player_metadata']
before = Path('workplace/round52-pass-a')
for n, expected in previous['normal_player_runs'][0]['hashes'].items():
    assert sha(before/n) == expected
tables = {}
for name, address, stride, count in [('WORK', 0x5da44, 230, 1), ('FLEET', 0x60ba4, 122, 32),
                                    ('SHIP', 0x668b2, 78, 352), ('STELLAR', 0x5f2cc, 126, 32),
                                    ('PLANET', 0x53132, 82, 81)]:
    a = m[address:address+stride*count]
    b = (players[0]/'entry'/(name+'.DT0')).read_bytes()
    assert len(a) == len(b)
    ds = [i for i, (x, y) in enumerate(zip(a, b)) if x != y]
    tables[name] = {'bytes': len(a), 'different_bytes': len(ds),
                    'offset_counts': dict(Counter(i % stride for i in ds)),
                    'original_sha256': hashlib.sha256(a).hexdigest(),
                    'remake_sha256': hashlib.sha256(b).hexdigest()}
assert [tables[n]['different_bytes'] for n in ('WORK', 'FLEET', 'SHIP', 'STELLAR', 'PLANET')] == [14, 27, 269, 0, 0]
scouts = []
pools = set()
new_ship = (players[0]/'entry/SHIP.DT0').read_bytes()
old_ship = (before/'entry/SHIP.DT0').read_bytes()
for fleet in (12, 16, 17, 28, 29, 30, 31):
    f = m[0x60ba4+fleet*122:0x60ba4+(fleet+1)*122]
    pool = struct.unpack_from('<H', f, 0x38)[0]
    for j in range(f[0x0f]):
        pools.add(struct.unpack_from('<H', f, 0x10+j*4)[0])
    pools.add(pool)
    r = m[0x668b2+pool*78:0x668b2+(pool+1)*78]
    assert f[0x3a] == 255 and r[4] == 7 and r[0x17] == 0 and r[0x2f] == 9
    assert old_ship[pool*78+0x2f] == 255 and new_ship[pool*78+0x2f] == 9
    scouts.append({'fleet': fleet, 'pool': pool, 'original_record_hex': r.hex(),
                   'before_record_hex': old_ship[pool*78:(pool+1)*78].hex(),
                   'after_record_hex': new_ship[pool*78:(pool+1)*78].hex()})
assert len(pools) == 55
assert all(new_ship[p*78+0x2f] == m[0x668b2+p*78+0x2f] for p in pools)
changes = [i for i, (a, b) in enumerate(zip(old_ship, new_ship)) if a != b]
assert changes == sorted(s['pool']*78+0x2f for s in scouts)
for n in names:
    if not n.endswith('/SHIP.DT0'):
        assert (players[0]/n).read_bytes() == (before/n).read_bytes(), n
    else:
        a, b = (before/n).read_bytes(), (players[0]/n).read_bytes()
        assert [i for i, (x, y) in enumerate(zip(a, b)) if x != y] == (changes if not n.startswith('loaded/') else [])
frames = {}
for display in ('entry', 'flagship', 'squad', 'empty', 'panel'):
    frame = previous['frames'][display]
    for suffix, expected in frame['images_sha256'].items():
        assert sha('workplace/round52-'+display+suffix) == expected
    assert (players[0]/display/'screen.bin').read_bytes() == (before/display/'screen.bin').read_bytes()
    frames[display] = {'original': frame['original_path'], 'original_hashes': frame['original_hashes'],
                       'after_screen_sha256': sha(players[0]/display/'screen.bin'),
                       'before_screen_sha256': sha(before/display/'screen.bin'),
                       'before_after_different_pixels': 0, 'regions': frame['regions'],
                       'unchanged_images_prefix': 'workplace/round52-'+display,
                       'unchanged_images_sha256': frame['images_sha256']}
semantics = {
    '26E5:016F': 'FLEET+38讀取偵察池', '26E5:0173': '偵察池乘SHIP stride 78',
    '26E5:0188': 'FLEET+3A與FF比較', '26E5:018F': '未出擊偵察隊force為0',
    '26E5:01B6': '未出擊偵察隊仍呼叫25DB:03E9', '26E5:01C4': '同一池呼叫標記刷新1727:02E2',
    '25DB:0408': 'force檢查', '25DB:040C': '非強制比較SHIP+17與3',
    '25DB:0491': '比較SHIP+04艦種與7', '25DB:0498': '非強制艦種7寫SHIP+2F為9',
}
ida = []
for name, expected, loc in [('re79-marker-r33-reviewed-26E5', '67c9e061d877dcbc91d59500450d7f651e0521aa023b1e072e29e0839e29bd2a', '26E5:00AC'),
                            ('re79-second-r36-25DB', 'b2abee692687909ff67958d75cc8bf8a7241f363767e439f6293cd57b912153b', '25DB:03E9')]:
    path = 'workplace/ida/'+name+'.json'
    d = checked_json(path, expected)
    assert d['tool'] == 'IDA Pro 9.4'
    assert sha(d['input'].replace('/work/', 'workplace/ida/')) == d['input_sha256']
    f = next(f for f in d['functions'] if f['original_location'] == loc)
    for row in f['instructions']:
        address = int(row['runtime_linear_hex'], 16)
        raw = bytes.fromhex(row['original_bytes_hex'])
        assert m[address:address+len(raw)] == raw
        semantic = semantics.get(row['original_location'])
        row['additional_semantic'] = {'text': semantic or '未審查', 'grade': 'proven' if semantic else 'unknown',
                                       'source': '研究391；原指令資料流及固定正常首戰雙跑RAM',
                                       'warning': '' if semantic else '⚠ 未證實語意'}
    ida.append({'export_path': path, 'export_sha256': expected, 'input': d['input'],
                'input_sha256': d['input_sha256'], 'tool': d['tool'],
                'address_space': d['address_space'], 'ida_base_hex': d['ida_base_hex'], 'function': f})
assert sum(len(d['function']['instructions']) for d in ida) == 209
gui = []
for scale in (1, 2):
    prefix = Path(f'workplace/round53-gui-{scale}x')
    assert prefix.with_suffix('.bin').read_bytes() == (players[0]/'entry/screen.bin').read_bytes()
    gui.append({'scale': scale, 'prefix': str(prefix), 'hashes': {s: sha(str(prefix)+s) for s in ('.bin', '.png', '.txt', '.log')},
                'classification': '正式cmd/gin3真實X11正常入口，邏輯畫面與Game相同'})
verification = {}
for label in ('arrival', 'limited', 'standard'):
    p = Path(f'workplace/round53-{label}.log')
    text = p.read_text()
    assert 'FAIL' not in text and 'SKIP' not in text
    passes = len(re.findall(r'^\s*--- PASS:', text, re.M))
    tested_packages = len(re.findall(r'^ok\s', text, re.M))
    no_test_packages = len(re.findall(r'^\?\s.*\[no test files\]$', text, re.M))
    packages = tested_packages+no_test_packages
    assert packages == (18 if label == 'standard' else 3 if label == 'limited' else 1)
    verification[label] = {'path': str(p), 'sha256': sha(p), 'pass_events': passes,
                           'packages': packages, 'tested_packages': tested_packages,
                           'no_test_packages': no_test_packages, 'skips': 0}
sources = ['internal/game/battlemarker.go', 'internal/game/fleetarrival_test.go',
           'internal/rules/battle.go', 'internal/rules/battlemarker.go', 'tools/scripts/battle_scout_cache_receipt.py',
           'tools/scripts/battle_normal_probe.go', 'tools/scripts/battle_normal_probe_seed.py', 'tools/turn_counter_gui.sh',
           'workplace/round53-normal-overlay.json', 'workplace/round53-normal-seed.go']
data = {k: previous[k] for k in ('date', 'dosgolem_head', 'dosgolem_dirty', 'original_bundle_sha256', 'original_executable_sha256', 'toolchain')}
data.update({'schema': 'logh3.battle-scout-cache.v1', 'issue': 89,
             'parent_commit': '517fecc21f8ad0868e26e49448e7edea70e3ee25', 'remake_worktree': 'dirty; 本輪快取修正及回歸',
             'status': '限定快取及保存已驗，完整088仍DRAFT', 'comparison_level': 'nearby-state; 不升格第52輪',
             'previous_receipts': {previous_path: sha(previous_path), original_path: sha(original_path)},
             'original_runs_reused': original_runs, 'original_ram_sha256': sha(ram), 'ida': ida,
             'normal_player_metadata': metadata,
             'normal_player_runs': [{'path': str(p), 'hashes': {n: sha(p/n) for n in names}} for p in players],
             'scouts': scouts, 'battle_pools': sorted(pools), 'changed_ship_byte_offsets': changes,
             'tables': tables, 'frames': frames, 'gui': gui, 'verification': verification,
             'source_hashes': {p: sha(p) for p in sources},
             'limitations': ['SHIP+30、顯示物件far指標及WORK14尚未對齊',
                             '首戰全圖仍一像素差；一般FF物件word+4初始化unknown',
                             '原版雙跑沿用固定f7523f9收據，沒有宣稱本輪重新執行原版',
                             '此輪不修改顯示，五張PNG／差分圖沿用逐byte核對相同的第52輪影像']})
out.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print('收據SHA-256', sha(out), '限定驗證', verification['limited']['pass_events'], 'PASS')
