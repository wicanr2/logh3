"""第50輪固定原版、正常入口與完整差分；在tools/gui.sh xrun內執行。"""
from collections import Counter
import hashlib
import json
import os
from pathlib import Path
import re
import struct
from pngout import game_palette, write_png


def sha(p):
    return hashlib.sha256(Path(p).read_bytes()).hexdigest()


def pixels(a, b, pal, rect):
    x, y, w, h = rect
    offsets = [j*640+i for j in range(y, y+h) for i in range(x, x+w)]
    diff = [i for i in offsets if a[i] != b[i]]
    return {'rectangle_xywh': rect, 'compared_pixels': len(offsets), 'same_pixels': len(offsets)-len(diff),
            'different_index_pixels': len(diff),
            'different_rgb_pixels': sum(pal[a[i]*3:a[i]*3+3] != pal[b[i]*3:b[i]*3+3] for i in diff),
            'maximum_channel_difference': max((abs(pal[a[i]*3+k]-pal[b[i]*3+k]) for i in diff for k in range(3)), default=0),
            'difference_bbox_inclusive': [min(i%640 for i in diff), min(i//640 for i in diff), max(i%640 for i in diff), max(i//640 for i in diff)] if diff else None}


def runs(prefix, names):
    paths = [Path('workplace/oracle/'+prefix+n) for n in 'ab']
    assert all((paths[0]/n).read_bytes() == (paths[1]/n).read_bytes() for n in names)
    result = []
    for p in paths:
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in (p/'meta.txt').read_text()
        for n in names:
            if n in ('loaded.bin', 'pre-battle.bin', 'entry.bin', 'end.bin'):
                assert (p/n).stat().st_size == 1048576
        result.append({'path': str(p), 'command_argv': (p/'flags.txt').read_text().splitlines(),
                       'hashes': {n: sha(p/n) for n in names+['meta.txt', 'flags.txt', 'report.txt']}})
    return paths[0], result


out = Path('docs/internals/data/fleet-arrival-placement.json')
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
original, originals = runs('normal89-r50-arrival-', ['loaded.bin', 'pre-battle.bin', 'entry.bin', 'screen.bin', 'screen.png', 'pal.bin'])
spot, spotting = runs('normal89-r50-arrival-spot-', ['entry.bin', 'end.bin', 'screen.bin', 'screen.png', 'pal.bin'])
for run in spotting:
    report = (Path(run['path'])/'report.txt').read_text()
    for addr, count in [('181B:0262', 1), ('181B:028C', 1), ('2526:00F8', 2), ('2526:033B', 0),
                        ('2526:04B1', 0), ('3826:0298', 0), ('3765:006C', 0), ('25DB:0E78', 0)]:
        assert f'{addr} 執行時的暫存器（前 {count} 次）' in report
    assert struct.unpack_from('<H', (Path(run['path'])/'end.bin').read_bytes(), 0x5da4a)[0] == 1
ida_path = Path('workplace/ida/normal89-r50-arrival-reviewed.json')
ida = json.loads(ida_path.read_text())
loaded = (original/'loaded.bin').read_bytes()
assert ida['uid'] == 1000 and ida['kernel_version'] == '9.4'
assert ida['input_sha256'] == hashlib.sha256(bytes(256)+loaded[0x297b0:0x2a650]).hexdigest()
assert len(ida['instructions']) == 465 and len(ida['functions']) == 5
for row in ida['instructions']:
    addr = int(row['runtime_linear_hex'], 16)
    raw = bytes.fromhex(row['original_bytes_hex'])
    assert loaded[addr:addr+len(raw)] == raw
players = [Path('workplace/round50-final-'+n) for n in 'ab']
names = sorted(p.relative_to(players[0]).as_posix() for p in players[0].rglob('*') if p.is_file())
assert len(names) == 19
assert names == sorted(p.relative_to(players[1]).as_posix() for p in players[1].rglob('*') if p.is_file())
assert all((players[0]/n).read_bytes() == (players[1]/n).read_bytes() for n in names)
meta = json.loads((players[0]/'probe.json').read_text())
assert meta['date'] == [799, 3, 1] and meta['ticks'] == 460
assert meta['strategy_seed_before'] == 'D3DC167E' and meta['strategy_seed_after'] == 'B6E4B78A'
mem = (original/'entry.bin').read_bytes()
tables = {}
for name, addr, stride, count, expected in [('WORK', 0x5da44, 230, 1, 14), ('FLEET', 0x60ba4, 122, 32, 27),
                                          ('SHIP', 0x668b2, 78, 352, 276), ('STELLAR', 0x5f2cc, 126, 32, 0), ('PLANET', 0x53132, 82, 81, 0)]:
    a = mem[addr:addr+stride*count]
    before = Path('workplace/round49-normal-a/entry/'+name+'.DT0').read_bytes()
    b = (players[0]/'entry'/(name+'.DT0')).read_bytes()
    assert len(a) == len(before) == len(b)
    diff = [i for i, (x, y) in enumerate(zip(a, b)) if x != y]
    assert len(diff) == expected, (name, len(diff))
    tables[name] = {'bytes': len(a), 'different_bytes_before': sum(x != y for x, y in zip(a, before)),
                    'different_bytes_after': len(diff), 'offset_counts': dict(sorted(Counter(i%stride for i in diff).items())),
                    'original_sha256': hashlib.sha256(a).hexdigest(), 'before_sha256': hashlib.sha256(before).hexdigest(), 'after_sha256': hashlib.sha256(b).hexdigest(),
                    'unresolved_byte_differences': [{'offset_hex': hex(i), 'original': a[i], 'remake': b[i]} for i in diff]}
moved, pools = [], set()
for index in (8, 12, 14, 16, 17, 28, 29, 30, 31):
    off = 0x60ba4+index*122
    record, old = mem[off:off+122], loaded[off:off+122]
    got = (players[0]/'entry/FLEET.DT0').read_bytes()[index*122:(index+1)*122]
    for start, size in ((0x4a, 3), (0x5c, 4), (0x64, 4)):
        assert record[start:start+size] == got[start:start+size]
    group_pools = [struct.unpack_from('<H', record, 0x10+4*j)[0] for j in range(record[15])]
    group_pools.append(struct.unpack_from('<H', record, 0x38)[0])
    pools.update(group_pools)
    moved.append({'index': index, 'from': old[6], 'to': record[6], 'heads_before': list(old[0x4a:0x4d]), 'heads_after': list(record[0x4a:0x4d]),
                  'xy_before': list(struct.unpack_from('<HH', old, 0x5c)), 'xy_after': list(struct.unpack_from('<HH', record, 0x5c)), 'pools': group_pools})
assert len(pools) == 70
ship = (players[0]/'entry/SHIP.DT0').read_bytes()
for pool in pools:
    off = 0x668b2+pool*78
    a, b = mem[off:off+78], ship[pool*78:(pool+1)*78]
    assert loaded[off+2] == a[2] == b[2]
    for start, size in ((0x28, 4), (0x34, 4), (0x16, 4), (0x1e, 1)):
        assert a[start:start+size] == b[start:start+size]
a, pal = (original/'screen.bin').read_bytes(), bytes(game_palette(str(original/'screen.png')))
assert len(a) == 307200 and len(pal) == 48
frames = {}
for label, p in [('before', Path('workplace/round49-normal-a/entry')), ('after', players[0]/'entry')]:
    b, base = (p/'screen.bin').read_bytes(), 'workplace/round50-normal-'+label
    assert not Path(base+'.png').exists()
    write_png(base+'.png', 640, 480, b, pal)
    write_png(base+'-diff.png', 640, 480, bytes(int(x != y) for x, y in zip(a, b)), bytes([0, 0, 0, 255, 0, 0]))
    write_png(base+'-pair.png', 1280, 480, b''.join(a[y*640:(y+1)*640]+b[y*640:(y+1)*640] for y in range(480)), pal)
    frames[label] = {'statistics': {n: pixels(a, b, pal, r) for n, r in {'full': (0, 0, 640, 480), 'field': (0, 0, 448, 360), 'right': (448, 0, 192, 480), 'bottom': (0, 360, 448, 120)}.items()},
                     'images_sha256': {s: sha(base+s) for s in ('.png', '-diff.png', '-pair.png')}}
assert frames['after']['statistics']['full']['different_index_pixels'] == 5853
gui = []
for scale in (1, 2):
    base = f'workplace/round50-gui-final-{scale}x'
    state = Path(base+'.txt').read_text()
    assert 'state=StateBattle' in state and f'window={640*scale}x{480*scale}' in state and 'cursor=240,240' in state
    assert Path(base+'.bin').read_bytes() == (players[0]/'entry/screen.bin').read_bytes()
    gui.append({'scale': scale, 'classification': 'normal cmd/gin3; real X11 keyboard and mouse; no direct entry', 'hashes': {s: sha(base+s) for s in ('.bin', '.png', '.txt', '.log')}})
limited, standard = [Path('workplace/round50-'+n+'-retest.log').read_text() for n in ('limited', 'standard')]
assert '--- SKIP:' not in limited and '--- FAIL:' not in limited
assert len(re.findall(r'^(ok\s|\?\s)', standard, re.M)) == 18 and 'FAIL' not in standard
sources = ['internal/game/'+n+'.go' for n in ('game', 'battle', 'rebellion', 'fleetarrival', 'fleetarrival_test', 'battlespotting_test', 'battlemovement', 'battlemovement_test', 'battlephase_test', 'battlesecond_test')]
sources += ['internal/rules/'+n+'.go' for n in ('battle', 'fleetsave', 'fleetarrival', 'battlemovement')]
sources += ['tools/fleet_arrival_oracle.sh', 'tools/turn_counter_gui.sh', 'tools/scripts/fleet_arrival_ida.py', 'tools/scripts/fleet_arrival_receipt.py', 'tools/scripts/battle_normal_probe_seed.py', 'tools/scripts/battle_normal_probe.go', 'workplace/round50-normal-overlay.json', 'workplace/round50-normal-seed.go']
baseline = Path('docs/internals/data/month-seat-gate.json')
assert sha(baseline) == '1a06468e6cf408a3e4bbae057bf6a7111cc1ba1fca548d7e025b9d21d3939217'
data = {'schema': 'logh3.fleet-arrival-placement.v1', 'date': '2026-10-03', 'parent_commit': '70fd008c2bf0c742bdb66685794b57a11d529e8b',
        'status': 'limited-production-validation; full parity unresolved', 'issue': 'https://github.com/wicanr2/logh3/issues/89', 'grade': 'proven; 限明列欄位與控制邊界',
        'original_bundle_sha256': 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
        'original_executable_sha256': '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918',
        'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', 'source_snapshot_sha256': '0155f63e373582b76a3b7ee54eccbef5ae2626481e09a68550b5fddc580146d7',
        'toolchain': json.loads(baseline.read_text())['toolchain'], 'original_runs': originals, 'normal_first_spotting_runs': spotting,
        'normal_first_spotting': {'classification': 'normal original with fixed test RNG; no state injection', 'tactical_seed': '12345678', 'completion_steps': [483778106, 484357728], 'seats_queried': 2, 'human_header_calls': 0, 'highlight_calls': 0},
        'ida': ida, 'ida_export_sha256': sha(ida_path), 'moved_fleets': moved, 'moved_pool_count': 70,
        'normal_player_runs': [{'path': str(p), 'hashes': {n: sha(p/n) for n in names}} for p in players], 'normal_player_metadata': meta,
        'tables': tables, 'frames': frames, 'gui': gui, 'comparison_level': 'nearby-state; 沿第44輪預定等級，未升格',
        'normalization': '原生640x480色號與原版PNG有效PLTE；無裁切、縮放或遮罩', 'sources_sha256': {n: sha(n) for n in sources}, 'previous_receipt_sha256': sha(baseline),
        'verification': {'limited_pass_events': len(re.findall(r'--- PASS:', limited)), 'limited_skip': 0, 'standard_packages': 18, 'command': 'tools/gui.sh xrun go test ./internal/... ./cmd/... -count=1',
                         'logs_sha256': {n: sha('workplace/'+n) for n in ['round50-limited-retest.log', 'round50-standard-retest.log', 'round50-standard-final.log', 'round50-standard-unscoped-failed.log', 'round50-limited.log', 'round50-passive-fixed.log']}},
        'limitations': ['完整088保持DRAFT，#89／#79 OPEN', 'WORK14／FLEET27／SHIP276 bytes與全圖5853 pixels仍不同', 'FF參數／缺入口fallback只保留靜態證據，未新增動態驗收',
                        'StartBattle診斷入口保持舊近似，不當作正常首戰', 'CPU停留為既有061刻意偏離；原版與VS路徑不凍結', '首回合無報告只限本次正常S3；受控151幀未換成此局面',
                        '未新增同盟首戰／雙人全流程／其他八步／音樂聆聽或實機驗收']}
out.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print(out, sha(out), data['verification'])
