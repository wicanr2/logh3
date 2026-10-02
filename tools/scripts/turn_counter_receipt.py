"""第47輪#89；驗完整RAM／原生全圖，重生已索引的計數器收據。"""
from collections import Counter
import hashlib
import json
from pathlib import Path
import struct

from pngout import game_palette, write_png


def sha(p):
    return hashlib.sha256(Path(p).read_bytes()).hexdigest()


def stats(a, b, pal, rect):
    x, y, w, h = rect
    offsets = [j*640+i for j in range(y, y+h) for i in range(x, x+w)]
    diff = [i for i in offsets if a[i] != b[i]]
    rgb = [i for i in diff if pal[a[i]*3:a[i]*3+3] != pal[b[i]*3:b[i]*3+3]]
    return {'rectangle_xywh': rect, 'compared_pixels': len(offsets),
            'same_pixels': len(offsets)-len(diff), 'different_index_pixels': len(diff),
            'different_rgb_pixels': len(rgb),
            'maximum_channel_difference': max((abs(pal[a[i]*3+k]-pal[b[i]*3+k]) for i in diff for k in range(3)), default=0),
            'difference_bbox_inclusive': [min(i%640 for i in diff), min(i//640 for i in diff), max(i%640 for i in diff), max(i//640 for i in diff)] if diff else None}


originals = []
for kind, stages, turns in [('clock', ['initial', 'before', 'after', 'command'], [0, 0, 1, 1]),
                            ('battle', ['loaded', 'entry'], [380, 380])]:
    paths = [Path('workplace/oracle/normal89-r47-'+kind+'-'+n) for n in 'ab']
    names = [s+'.bin' for s in stages]+['screen.bin', 'screen.png', 'pal.bin']
    assert all((paths[0]/n).read_bytes() == (paths[1]/n).read_bytes() for n in names)
    for p in paths:
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in (p/'meta.txt').read_text()
        argv = (p/'flags.txt').read_text().splitlines()
        assert argv[argv.index('-tick')+1] == '20000'
        if kind == 'clock':
            assert '-poke' not in argv
            assert '#435711456 5DA48: 00→01' in (p/'report.txt').read_text()
        else:
            assert argv[argv.index('-poke')+1] == 'lin:5E00C@300000000=7E 16 DC D3'
            assert '[watch]' not in (p/'report.txt').read_text()
        state = {}
        for stage, turn in zip(stages, turns):
            mem = (p/(stage+'.bin')).read_bytes()
            assert len(mem) == 1048576
            assert struct.unpack_from('<H', mem, 0x5da48)[0] == turn
            state[stage] = {'turn': turn, 'strategy_seed_hex': f'{struct.unpack_from("<I", mem, 0x5e00c)[0]:08X}'}
        originals.append({'path': str(p), 'classification': 'normal player continuation; fixed source snapshot',
                          'command_argv': argv, 'states': state,
                          'hashes': {n: sha(p/n) for n in names+['meta.txt', 'flags.txt', 'report.txt']}})

ida_path = Path('workplace/ida/normal89-r47-counter.json')
ida = json.loads(ida_path.read_text())
assert ida['uid'] == 1000 and ida['kernel_version'] == '9.4'
assert ida['original_name'] == 'sub_10543' and ida['function_boundary'] == ['0x10543', '0x105e8']
assert ida['input_sha256'] == '4d3f044955ce8e52c79b5b0c52060cd1b0474b57bddc2aa76f871f283799d72d'
initial = Path('workplace/oracle/normal89-r47-clock-a/initial.bin').read_bytes()
assert hashlib.sha256(bytes(256)+initial[0x3740:0x42f0]).hexdigest() == ida['input_sha256']
assert initial[0x3b19:0x3b1d] == bytes.fromhex('ff06188f')

players = [Path('workplace/round47-normal-'+n) for n in 'ab']
names = sorted(p.relative_to(players[0]).as_posix() for p in players[0].rglob('*') if p.is_file())
assert names == sorted(p.relative_to(players[1]).as_posix() for p in players[1].rglob('*') if p.is_file())
assert len(names) == 19
assert all((players[0]/n).read_bytes() == (players[1]/n).read_bytes() for n in names)
meta = json.loads((players[0]/'probe.json').read_text())
assert meta['date'] == [799, 3, 1]
assert meta['strategy_seed_before'] == meta['strategy_seed_observed_before'] == 'D3DC167E'
assert meta['strategy_seed_after'] == 'B6E4B78A'
for stage in ('loaded', 'entry'):
    raw = (players[0]/stage/'WORK.DT0').read_bytes()
    assert struct.unpack_from('<H', raw, 4)[0] == 380

mem = Path('workplace/oracle/normal89-r47-battle-a/entry.bin').read_bytes()
old = Path('workplace/round46-normal-a/entry')
baseline = 'docs/internals/data/load-seat-selection.json'
assert sha(baseline) == 'f8ba6dacc701418f1fe88c0c785fd871af8677916e4a54ed0ad1f9c19debf129'
tables = {}
for name, addr, stride, count, expected in [('WORK', 0x5da44, 230, 1, 32), ('FLEET', 0x60ba4, 122, 32, 120),
                                          ('SHIP', 0x668b2, 78, 352, 865), ('STELLAR', 0x5f2cc, 126, 32, 44)]:
    length = stride*count
    original = mem[addr:addr+length]
    before = (old/(name+'.DT0')).read_bytes()
    after = (players[0]/'entry'/(name+'.DT0')).read_bytes()
    assert len(original) == len(before) == len(after) == length
    diffs = [i for i, (a, b) in enumerate(zip(original, after)) if a != b]
    assert len(diffs) == expected
    tables[name] = {'bytes': length, 'different_bytes_before': sum(a != b for a, b in zip(original, before)),
                    'different_bytes_after': len(diffs), 'offset_counts': dict(sorted(Counter(i%stride for i in diffs).items())),
                    'original_sha256': hashlib.sha256(original).hexdigest(),
                    'before_sha256': hashlib.sha256(before).hexdigest(), 'after_sha256': hashlib.sha256(after).hexdigest(),
                    'unresolved_byte_differences': [{'offset_hex': hex(i), 'original': original[i], 'remake': after[i]} for i in diffs]}
fleet = (players[0]/'entry/FLEET.DT0').read_bytes()
ship = (players[0]/'entry/SHIP.DT0').read_bytes()
assert [i for i in range(32) if fleet[i*122+4] != 255 and fleet[i*122+6] == 26] == [12, 16, 17, 28, 29, 30, 31]
assert sum(ship[i*78+2] == 26 for i in range(352)) == 55

orig_dir = Path('workplace/oracle/normal89-r47-battle-a')
a = (orig_dir/'screen.bin').read_bytes()
before, after = [(p/'screen.bin').read_bytes() for p in (old, players[0]/'entry')]
pal = bytes(game_palette(str(orig_dir/'screen.png')))
assert len(a) == len(before) == len(after) == 307200 and len(pal) == 48
rects = {'full': (0, 0, 640, 480), 'field': (0, 0, 448, 360),
         'right': (448, 0, 192, 480), 'bottom': (0, 360, 448, 120)}
images, comparisons = {}, {}
for label, raw in [('before', before), ('after', after)]:
    base = 'workplace/round47-normal-'+label
    write_png(base+'.png', 640, 480, raw, pal)
    write_png(base+'-diff.png', 640, 480, bytes(int(x != y) for x, y in zip(a, raw)), bytes([0, 0, 0, 255, 0, 0]))
    pair = b''.join(a[y*640:(y+1)*640]+raw[y*640:(y+1)*640] for y in range(480))
    write_png(base+'-pair.png', 1280, 480, pair, pal)
    comparisons[label] = {n: stats(a, raw, pal, r) for n, r in rects.items()}
    images.update({base+s: sha(base+s) for s in ('.png', '-diff.png', '-pair.png')})
assert comparisons['before']['full']['different_index_pixels'] == 5863
assert comparisons['after']['full']['different_index_pixels'] == 5853

gui = []
for scale in (1, 2):
    base = 'workplace/round47-gui-pass-'+str(scale)+'x'
    text = Path(base+'.txt').read_text()
    assert 'state=StateBattle\n' in text and f'window={640*scale}x{480*scale}\n' in text
    assert 'cursor=240,240\n' in text and 'cursor_mode=1\n' in text
    assert Path(base+'.bin').read_bytes() == after
    gui.append({'path_prefix': base, 'normal_x11_input': True, 'full_pixels_different_from_normal_game': 0,
                'command_argv': ['tools/gui.sh', 'xrun', 'bash', 'tools/turn_counter_gui.sh'],
                'product_arguments': ['-orig', '/orig', '-data', '/src/docs/internals/data', '-intro=false',
                    '-scale', str(scale), '-save', '/tmp/logh3-turn-counter-save', '-shot-bin', '/src/'+base+'.bin',
                    '-shot', '/src/'+base+'.png', '-state-out', '/src/'+base+'.txt', '-shot-after', '1200'],
                'hashes': {base+s: sha(base+s) for s in ('.bin', '.png', '.txt', '.log')}})

logs = ['workplace/round47-counter-limited.log', 'workplace/round47-standard-tests.log']
limited = Path(logs[0]).read_text()
assert '\nPASS\n' in limited and '--- SKIP:' not in limited and '--- FAIL:' not in limited
standard = Path(logs[1]).read_text()
assert 'FAIL' not in standard
assert len([s for s in standard.splitlines() if s.startswith(('ok ', '? '))]) == 18
sources = ['internal/game/'+n for n in ('game.go', 'originalbrain.go', 'battle.go', 'regionturn.go', 'phezzan.go',
           'turncounter_test.go', 'battlesamples_test.go', 'regionturn_test.go', 'victoryreport_test.go', 'builder_test.go')]
sources += ['tools/turn_counter_oracle.sh', 'tools/turn_counter_gui.sh', 'tools/scripts/turn_counter_ida.py',
            'tools/scripts/turn_counter_receipt.py', 'tools/scripts/battle_normal_probe.go',
            'workplace/round44-normal-overlay.json', 'workplace/round44-normal-seed.go']
data = {'schema': 'logh3.turn-counter-boundary.v1', 'grade': 'proven counter boundary; complete normal entry parity unknown',
        'classification': 'nearby-state full-screen comparison; no masking or cropping',
        'parent_head': '3c7d2126562c2a81e2519fe0bab8ac09496647a9', 'remake_worktree': 'dirty; source hashes below',
        'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', 'dosgolem_worktree': 'clean',
        'program_sha256': sha('org_game/extracted/GIN3.COM'), 'bundle_sha256': sha('org_game/银河英雄传说3SP.zip'),
        'images': {'go': {'tag': 'golang:1.24-bookworm', 'id': 'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac'},
                   'gui': {'tag': 'logh3-ebiten:2', 'id': 'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8'},
                   'ida': {'tag': 'ida-pro-9.4-idapython:locked-v1', 'id': 'sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780'}},
        'parent_snapshots': json.loads(Path('workplace/round47-parent-snapshots.json').read_text()),
        'originals': originals, 'ida_export': {'path': str(ida_path), 'sha256': sha(ida_path), 'data': ida},
        'normal_players': [{'path': str(p), 'command_argv': ['tools/gui.sh', 'xrun', 'go', 'run', '-overlay',
           'workplace/round44-normal-overlay.json', 'tools/scripts/battle_normal_probe.go', '-out', str(p)],
           'hashes': {n: sha(p/n) for n in names}} for p in players],
        'normal_metadata': meta, 'date_before': [799, 3, 4], 'date_after': [799, 3, 1],
        'before_baseline': {'path': baseline, 'sha256': sha(baseline), 'remake_committed_head': '3c7d2126562c2a81e2519fe0bab8ac09496647a9'},
        'original_and_remake_turn_after': 380, 'table_differences': tables,
        'screen_comparisons': comparisons, 'comparison_image_hashes': images, 'actual_x11_gui': gui,
        'source_hashes': {p: sha(p) for p in sources}, 'test_logs': {p: sha(p) for p in logs},
        'limited_pass_events': limited.count('--- PASS:'), 'limited_skip_events': 0, 'standard_packages': 18,
        'limits': ['088整份DRAFT，活動座位、runtime欄位及戰場資料仍不同',
                   '全部戰場／9→10／費沙模態／終局及兩AI的邊界是remake接線抽驗，不冒稱這些整段原版同態',
                   '真實X11只驗帝國S3與兩視窗尺寸；未聲稱觸控或macOS',
                   '兩次外部resize未通過；第一回點錯命令、第二回維持640x480。改用既有-scale並以PID選視窗後重跑，不當產品缺陷']}
target = Path('docs/internals/data/turn-counter-boundary.json')
target.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print('Original pairs 2; normal player files 19; GUI full frame equal 2; limited PASS', data['limited_pass_events'])
print('Full pixel differences 5863 -> 5853; WORK 33 -> 32; unresolved tables retained')
print('receipt SHA256', sha(target))
