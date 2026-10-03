"""第49輪原版閘門與正式正常入口；在專案Docker內執行。"""
import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
import re
import struct

from pngout import game_palette, write_png


def sha(p):
    return hashlib.sha256(Path(p).read_bytes()).hexdigest()


def work(raw):
    w = raw[0x5da44:0x5da44+230]
    return {'hex': w.hex(), 'turn': struct.unpack_from('<H', w, 4)[0],
            'month_flags': list(w[224:226]),
            'military': list(struct.unpack_from('<II', w, 0x3e)),
            'tax': list(struct.unpack_from('<II', w, 0x56)),
            'building': list(struct.unpack_from('<14I', w, 0x88))}


def pixels(a, b, pal, rect):
    x, y, w, h = rect
    offsets = [j*640+i for j in range(y, y+h) for i in range(x, x+w)]
    diff = [i for i in offsets if a[i] != b[i]]
    rgb = [i for i in diff if pal[a[i]*3:a[i]*3+3] != pal[b[i]*3:b[i]*3+3]]
    return {'rectangle_xywh': rect, 'compared_pixels': len(offsets),
            'same_pixels': len(offsets)-len(diff), 'different_index_pixels': len(diff),
            'different_rgb_pixels': len(rgb),
            'maximum_channel_difference': max((abs(pal[a[i]*3+k]-pal[b[i]*3+k]) for i in diff for k in range(3)), default=0),
            'difference_bbox_inclusive': [min(i%640 for i in diff), min(i//640 for i in diff), max(i%640 for i in diff), max(i//640 for i in diff)] if diff else None}


parser = argparse.ArgumentParser()
parser.add_argument('--final', action='store_true')
args = parser.parse_args()
originals = []
for mode in ('natural', 'due'):
    paths = [Path('workplace/oracle/normal89-r49-'+mode+'-'+n) for n in 'ab']
    names = ['loaded.bin', 'entry.bin', 'screen.bin', 'screen.png', 'pal.bin']
    assert all((paths[0]/n).read_bytes() == (paths[1]/n).read_bytes() for n in names)
    for p in paths:
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in (p/'meta.txt').read_text()
        argv = (p/'flags.txt').read_text().splitlines()
        poke = 'lin:5E00C@300000000=7E 16 DC D3'
        if mode == 'due':
            poke += ';lin:5DB25@300000000=01'
        assert argv[argv.index('-poke')+1] == poke
        report = (p/'report.txt').read_text()
        assert ('2703:03AF 執行時的暫存器（前 1 次）' in report) == (mode == 'due')
        assert ('2B5F:0103 執行時的暫存器（前 2 次）' in report) == (mode == 'due')
        states = {}
        for stage in ('loaded', 'entry'):
            raw = (p/(stage+'.bin')).read_bytes()
            assert len(raw) == 1048576
            states[stage] = work(raw)
            assert states[stage]['turn'] == 380
        assert states['loaded']['month_flags'] == [0, int(mode == 'due')]
        assert states['entry']['month_flags'] == [0, 0]
        originals.append({'path': str(p), 'grade': 'proven',
                          'classification': 'normal-player-continuation' if mode == 'natural' else 'controlled-counterfactual; only WORK+E1 changed',
                          'command_argv': argv, 'states': states,
                          'hashes': {n: sha(p/n) for n in names+['meta.txt', 'flags.txt', 'report.txt']}})
natural = Path('workplace/oracle/normal89-r49-natural-a')
due = Path('workplace/oracle/normal89-r49-due-a')
a, b = [(p/'loaded.bin').read_bytes() for p in (natural, due)]
assert [i for i, (x, y) in enumerate(zip(a, b)) if x != y] == [0x5db25]
assert (natural/'screen.bin').read_bytes() == (due/'screen.bin').read_bytes()
ida_path = Path('workplace/ida/normal89-r49-gate.json')
ida = json.loads(ida_path.read_text())
assert ida['uid'] == 1000 and ida['kernel_version'] == '9.4'
assert ida['original_name'] == 'sub_10234' and ida['function_boundary'] == ['0x10234', '0x10330']
assert ida['input_sha256'] == hashlib.sha256(bytes(256)+a[0x6540:0x6e80]).hexdigest()
for row in ida['instructions']:
    addr = int(row['runtime_linear_hex'], 16)
    raw = bytes.fromhex(row['original_bytes_hex'])
    assert a[addr:addr+len(raw)] == raw
counter = json.loads(Path('workplace/ida/normal89-r47-counter.json').read_text())
reset = []
sem = {'0374:03DD': '回合進位後AL設1', '0374:03DF': 'WORK+E1設1', '0374:03E2': 'WORK+E0設1'}
for old in counter['instructions']:
    if old['original_location'] not in sem:
        continue
    row = dict(old)
    addr = int(row['runtime_linear_hex'], 16)
    raw = bytes.fromhex(row['original_bytes_hex'])
    assert a[addr:addr+len(raw)] == raw
    row.update(semantic=sem[row['original_location']], grade='proven', warning='',
               source='第47輪IDA原始匯出、本輪完整RAM bytes核對及第48輪正常0→1回合旗標0／1')
    reset.append(row)
assert len(reset) == 3
baseline = Path('docs/internals/data/active-seat.json')
assert sha(baseline) == '40ce3fbedda34a073f80f0799a820e27f567b9cb3210cf6c282ef67d66f87bac'
data = {'schema': 'logh3.month-seat-gate.v1', 'date': '2026-10-03',
        'parent_commit': '5f94ca9799886800920a6cf5ac464ed4f71ef25c',
        'status': 'original-evidence-reviewed; production not yet verified',
        'issue': 'https://github.com/wicanr2/logh3/issues/89',
        'original_bundle_sha256': 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
        'original_executable_sha256': '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918',
        'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
        'toolchain': {'go_image': 'golang:1.24-bookworm',
                      'go_image_id': 'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac',
                      'gui_image': 'logh3-ebiten:2',
                      'gui_image_id': 'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
                      'ida_image': 'ida-pro-9.4-idapython:locked-v1',
                      'ida_image_id': 'sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780'},
        'source_snapshot_sha256': '0155f63e373582b76a3b7ee54eccbef5ae2626481e09a68550b5fddc580146d7',
        'original_runs': originals, 'ida': ida, 'ida_export_sha256': sha(ida_path),
        'main_loop_reset': reset, 'controlled_loaded_memory_difference': ['lin:5DB25'],
        'same_battle_frame_does_not_prove_same_finances': True,
        'previous_receipt_sha256': sha(baseline),
        'comparison_level': 'nearby-state; 沿第44輪預定等級，未升格',
        'limitations': ['完整088保持DRAFT', 'WORK／FLEET／SHIP／全圖仍有未分類差異',
                        '雙人跨月處理保持016既有範圍', '受控旗標1分支不當作自然S3',
                        '未新增任意人類旗標1讀檔月報的原版驗收聲明']}
if args.final:
    players = [Path('workplace/round49-normal-'+n) for n in 'ab']
    names = sorted(p.relative_to(players[0]).as_posix() for p in players[0].rglob('*') if p.is_file())
    assert len(names) == 19
    assert names == sorted(p.relative_to(players[1]).as_posix() for p in players[1].rglob('*') if p.is_file())
    assert all((players[0]/n).read_bytes() == (players[1]/n).read_bytes() for n in names)
    meta = json.loads((players[0]/'probe.json').read_text())
    assert meta['date'] == [799, 3, 1] and meta['ticks'] == 460
    assert meta['strategy_seed_before'] == 'D3DC167E' and meta['strategy_seed_after'] == 'B6E4B78A'
    mem = (natural/'entry.bin').read_bytes()
    tables = {}
    for name, addr, stride, count, expected in [('WORK', 0x5da44, 230, 1, 14), ('FLEET', 0x60ba4, 122, 32, 120),
                                              ('SHIP', 0x668b2, 78, 352, 865), ('STELLAR', 0x5f2cc, 126, 32, 0),
                                              ('PLANET', 0x53132, 82, 81, 0)]:
        original = mem[addr:addr+stride*count]
        before = Path('workplace/round48-normal-final-a/entry/'+name+'.DT0').read_bytes()
        after = (players[0]/'entry'/(name+'.DT0')).read_bytes()
        assert len(original) == len(before) == len(after)
        diffs = [i for i, (x, y) in enumerate(zip(original, after)) if x != y]
        assert len(diffs) == expected, (name, len(diffs))
        tables[name] = {'bytes': len(original), 'different_bytes_before': sum(x != y for x, y in zip(original, before)),
                        'different_bytes_after': len(diffs), 'offset_counts': dict(sorted(Counter(i%stride for i in diffs).items())),
                        'original_sha256': hashlib.sha256(original).hexdigest(), 'before_sha256': hashlib.sha256(before).hexdigest(),
                        'after_sha256': hashlib.sha256(after).hexdigest(),
                        'unresolved_byte_differences': [{'offset_hex': hex(i), 'original': original[i], 'remake': after[i]} for i in diffs]}
    a = (natural/'screen.bin').read_bytes()
    pal = bytes(game_palette(str(natural/'screen.png')))
    assert len(a) == 307200 and len(pal) == 48
    frames = {}
    for label, p in [('before', Path('workplace/round48-normal-final-a/entry')), ('after', players[0]/'entry')]:
        b = (p/'screen.bin').read_bytes()
        base = 'workplace/round49-normal-'+label
        write_png(base+'.png', 640, 480, b, pal)
        write_png(base+'-diff.png', 640, 480, bytes(int(x != y) for x, y in zip(a, b)), bytes([0, 0, 0, 255, 0, 0]))
        write_png(base+'-pair.png', 1280, 480, b''.join(a[y*640:(y+1)*640]+b[y*640:(y+1)*640] for y in range(480)), pal)
        frames[label] = {'statistics': {n: pixels(a, b, pal, r) for n, r in {'full': (0, 0, 640, 480), 'field': (0, 0, 448, 360), 'right': (448, 0, 192, 480), 'bottom': (0, 360, 448, 120)}.items()},
                         'images_sha256': {s: sha(base+s) for s in ('.png', '-diff.png', '-pair.png')}}
    assert frames['after']['statistics']['full']['different_index_pixels'] == 5853
    data.update(status='limited-production-validation; full parity unresolved', tables=tables, frames=frames,
                normal_player_runs=[{'path': str(p), 'hashes': {n: sha(p/n) for n in names}} for p in players],
                normal_player_metadata=meta)
    gui = []
    for scale in (1, 2):
        base = f'workplace/round49-gui-{scale}x'
        state = Path(base+'.txt').read_text()
        assert 'state=StateBattle' in state and f'window={640*scale}x{480*scale}' in state
        assert 'cursor=240,240' in state
        assert Path(base+'.bin').read_bytes() == (players[0]/'entry/screen.bin').read_bytes()
        gui.append({'scale': scale, 'classification': 'normal cmd/gin3; real X11 keyboard and mouse; no direct entry',
                    'visual_inspection': '7990301／第1回合／帝國輸入已目視確認',
                    'hashes': {s: sha(base+s) for s in ('.bin', '.png', '.txt', '.log')}})
    limited = Path('workplace/round49-limited.log').read_text()
    standard = Path('workplace/round49-standard.log').read_text()
    assert len(re.findall(r'--- PASS:', limited)) == 37
    assert '--- SKIP:' not in limited and '--- FAIL:' not in limited
    assert len(re.findall(r'^(ok\s|\?\s)', standard, re.M)) == 18 and 'FAIL' not in standard
    source_names = ['internal/game/game.go', 'internal/game/monthseat.go', 'internal/game/monthseat_test.go',
                    'internal/rules/save.go', 'tools/month_seat_gate_oracle.sh', 'tools/turn_counter_gui.sh',
                    'tools/scripts/month_seat_gate_ida.py', 'tools/scripts/month_seat_gate_receipt.py',
                    'tools/scripts/battle_normal_probe.go', 'workplace/round44-normal-overlay.json',
                    'workplace/round44-normal-seed.go']
    data.update(gui=gui, verification={'limited_pass_events': 37, 'limited_skip': 0, 'standard_packages': 18,
                                     'command': 'tools/gui.sh xrun go test ./internal/... ./cmd/... -count=1',
                                     'logs_sha256': {n: sha('workplace/'+n) for n in ['round49-limited.log', 'round49-standard.log', 'round49-limited-load-baseline-failed.log']},
                                     'failed_test_classification': '讀檔重算未安定度的比較點錯誤；改取正常載入後狀態，產品與旗標判準未改'},
                sources_sha256={n: sha(n) for n in source_names},
                remake_source_state='parent Git HEAD; modified month-seat implementation; exact sources hashed',
                capture_parameters={'original_mode': '640x480 planar VGA', 'remake_canvas': [640, 480],
                                    'normalization': '未裁切或縮放；原生色號與有效原版PLTE；全畫面及四區域',
                                    'seed_overlay': '只固定測試期RNG，未匯入原版RAM或旗標',
                                    'gui_command': 'tools/gui.sh xrun env LOGH3_TURN_GUI_PREFIX=round49-gui LOGH3_TURN_PLAYER_DIR=round49-normal-a bash tools/turn_counter_gui.sh'})
out = Path('docs/internals/data/month-seat-gate.json')
out.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print(out, sha(out), data['status'])
