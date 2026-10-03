"""第52輪：保存摘要小地圖限定驗收；tools/gui.sh xrun執行。"""
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


def stats(a, b, pal, rect):
    x, y, w, h = rect
    ds = [j*640+i for j in range(y, y+h) for i in range(x, x+w) if a[j*640+i] != b[j*640+i]]
    return {'rectangle_xywh': rect, 'compared_pixels': w*h, 'same_pixels': w*h-len(ds),
            'different_index_pixels': len(ds),
            'different_rgb_pixels': sum(pal[a[i]*3:a[i]*3+3] != pal[b[i]*3:b[i]*3+3] for i in ds),
            'maximum_channel_difference': max((abs(pal[a[i]*3+k]-pal[b[i]*3+k]) for i in ds for k in range(3)), default=0),
            'difference_bbox_inclusive': [min(i%640 for i in ds), min(i//640 for i in ds), max(i%640 for i in ds), max(i//640 for i in ds)] if ds else None,
            'remaining_pixels': [{'xy': [i%640, i//640], 'original_index': a[i], 'remake_index': b[i]} for i in ds] if len(ds) <= 10 else []}


out = Path('docs/internals/data/battle-hover-marks.json')
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
previous = Path('docs/internals/data/battle-input-display.json')
assert sha(previous) == 'c1f387c1e3729450f4b7b3f9f8babb21b0b8f49d1b0762b1f6e0cca648cfb5c6'
old = json.loads(previous.read_text())
entry_path = Path('workplace/oracle/normal89-r51-base-a/entry.bin')
entry = entry_path.read_bytes()
assert sha(entry_path) == 'ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97'
players = [Path('workplace/round52-pass-'+s) for s in 'ab']
names = sorted(p.relative_to(players[0]).as_posix() for p in players[0].rglob('*') if p.is_file())
assert len(names) == 55
assert names == sorted(p.relative_to(players[1]).as_posix() for p in players[1].rglob('*') if p.is_file())
assert all((players[0]/n).read_bytes() == (players[1]/n).read_bytes() for n in names)
metadata = json.loads((players[0]/'probe.json').read_text())
assert metadata['date'] == [799, 3, 1] and metadata['strategy_seed_after'] == 'B6E4B78A'
originals = {}
for kind in ('flagship', 'squad', 'fleet-target', 'squad-target'):
    roots = [Path(f'workplace/oracle/normal89-r52-{kind}-{s}') for s in 'ab']
    files = ('end.bin', 'screen.bin', 'screen.png', 'pal.bin')
    assert all((roots[0]/n).read_bytes() == (roots[1]/n).read_bytes() for n in files)
    originals[kind] = []
    for side, p in zip('ab', roots):
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in (p/'meta.txt').read_text()
        flags = (p/'flags.txt').read_text().splitlines()
        assert '-mouse-at' in flags and flags[flags.index('-mouse-at')+1] == '460000000'
        if 'target' in kind:
            address = '611C0' if kind == 'fleet-target' else '681D6'
            assert f'lin:{address}@457000000=40 02 90 01' in flags
            assert f'改記憶體 lin:{address}（{address}）4 bytes' in (p/'report.txt').read_text()
        m = (p/'end.bin').read_bytes()
        assert len(m) == 1048576 and struct.unpack_from('<I', m, 0x5e00c)[0] == 0xb6e4b78a
        parent = Path(f'workplace/oracle/normal89-r51-empty-{side}/machine.state')
        originals[kind].append({'path': str(p), 'command_argv': flags,
                               'parent_snapshot': str(parent), 'parent_snapshot_sha256': sha(parent),
                               'hashes': {n: sha(p/n) for n in files+('machine.state', 'flags.txt', 'meta.txt', 'report.txt')},
                               'diagnostic_injection': 'target' in kind})
spec = json.loads(Path('docs/internals/data/screen-battle.json').read_text())
mini = spec['minimap']
rectangles = {'full': (0, 0, 640, 480), 'field': (0, 0, 448, 360),
              'right': (448, 0, 192, 480), 'bottom': (0, 360, 448, 120),
              'unit_panel': (480, 242, 128, 160),
              'minimap': tuple(mini[k] for k in ('x', 'y', 'w', 'h'))}
frames = {}
for display, kind in (('entry', 'base'), ('flagship', 'flagship'), ('squad', 'squad'), ('empty', 'empty'), ('panel', 'panel')):
    op = Path(f'workplace/oracle/normal89-r51-{kind}-a')
    a = (op/'screen.bin').read_bytes()
    b = (players[0]/display/'screen.bin').read_bytes()
    before = (Path('workplace/round52-before')/display/'screen.bin').read_bytes()
    assert len(a) == len(b) == len(before) == 307200
    assert hashlib.sha256(before).hexdigest() == old['normal_player_runs'][0]['hashes'][display+'/screen.bin']
    pal = bytes(game_palette(str(op/'screen.png')))
    regions = {n: stats(a, b, pal, r) for n, r in rectangles.items()}
    assert regions['minimap']['different_index_pixels'] == 0
    assert regions['unit_panel']['different_index_pixels'] == 0
    assert regions['full']['different_index_pixels'] == (1 if display == 'entry' else 0)
    prefix = 'workplace/round52-'+display
    assert not Path(prefix+'.png').exists()
    write_png(prefix+'.png', 640, 480, b, pal)
    write_png(prefix+'-diff.png', 640, 480, bytes(int(x != y) for x, y in zip(a, b)), bytes([0, 0, 0, 255, 0, 0]))
    write_png(prefix+'-pair.png', 1280, 480, b''.join(a[y*640:(y+1)*640]+b[y*640:(y+1)*640] for y in range(480)), pal)
    frames[display] = {'regions': regions, 'before_regions': {n: stats(a, before, pal, r) for n, r in rectangles.items()},
                       'original_path': str(op), 'original_hashes': {n: sha(op/n) for n in ('screen.bin', 'screen.png', 'pal.bin', 'flags.txt', 'meta.txt')},
                       'original_command_argv': (op/'flags.txt').read_text().splitlines(),
                       'before_frame_sha256': hashlib.sha256(before).hexdigest(),
                       'images_sha256': {s: sha(prefix+s) for s in ('.png', '-diff.png', '-pair.png')},
                       'classification': 'nearby-state; static frames bounded, complete record parity unresolved'}
    for n in ('FLEET.DT0', 'SHIP.DT0', 'WORK.DT0'):
        assert (players[0]/display/n).read_bytes() == (players[0]/'entry'/n).read_bytes()
tables = {}
for name, addr, stride, count in [('WORK', 0x5da44, 230, 1), ('FLEET', 0x60ba4, 122, 32),
                                ('SHIP', 0x668b2, 78, 352), ('STELLAR', 0x5f2cc, 126, 32), ('PLANET', 0x53132, 82, 81)]:
    a = entry[addr:addr+stride*count]
    b = (players[0]/'entry'/(name+'.DT0')).read_bytes()
    ds = [i for i, (x, y) in enumerate(zip(a, b)) if x != y]
    tables[name] = {'bytes': len(a), 'different_bytes': len(ds), 'offset_counts': dict(Counter(i%stride for i in ds)),
                    'original_sha256': hashlib.sha256(a).hexdigest(), 'remake_sha256': hashlib.sha256(b).hexdigest()}
assert [tables[n]['different_bytes'] for n in ('WORK', 'FLEET', 'SHIP', 'STELLAR', 'PLANET')] == [14, 27, 276, 0, 0]
ida_path = Path('workplace/ida/normal89-r52-1727-graded.json')
ida = json.loads(ida_path.read_text())
assert ida['kernel_version'] == '9.4' and ida['uid'] == 1000 and len(ida['functions']) == 4
assert ida['input_sha256'] == sha(ida['input'].replace('/work/', 'workplace/ida/'))
for row in ida['instructions']:
    address, raw = int(row['runtime_linear_hex'], 16), bytes.fromhex(row['original_bytes_hex'])
    assert entry[address:address+len(raw)] == raw
assert len(ida['instructions']) == 105 and sum(r['grade'] == 'proven' for r in ida['instructions']) == 12
objects = []
count = struct.unpack_from('<H', entry, 0x54b30+0xeade)[0]
for index in range(count):
    off, seg = struct.unpack_from('<HH', entry, 0x54b30+0xeae0+index*4)
    address = seg*16+off
    raw = entry[address:address+40]
    assert len(raw) == 40
    if raw[3] >= 0xfe:
        objects.append({'list_index': index, 'runtime_linear_hex': hex(address),
                        'original_far_pointer_offset_segment': [hex(off), hex(seg)],
                        'original_bytes_hex': raw.hex(), 'priority_byte_plus_3': raw[3],
                        'unsigned_word_plus_4': struct.unpack_from('<H', raw, 4)[0],
                        'logical_xy_plus_6_8': list(struct.unpack_from('<HH', raw, 6)),
                        'tile_byte_plus_C': raw[12], 'grade': 'proven',
                        'source': '研究390；原版完整RAM及35D1:02AB比較器，FF字鍵初始化仍unknown'})
current = next(o for o in objects if o['runtime_linear_hex'] == '0x74c20')
direction = next(o for o in objects if o['runtime_linear_hex'] == '0x7101e')
assert current['unsigned_word_plus_4'] == 56319 and direction['unsigned_word_plus_4'] == 64
assert current['list_index'] < direction['list_index']
gui = []
for scale in (1, 2):
    prefix = f'workplace/round52-gui-final-{scale}x'
    state = Path(prefix+'.txt').read_text()
    assert 'state=StateBattle' in state and f'window={640*scale}x{480*scale}' in state
    assert Path(prefix+'.bin').read_bytes() == (players[0]/'entry/screen.bin').read_bytes()
    gui.append({'scale': scale, 'path': prefix, 'normal_player_path': True,
                'hashes': {s: sha(prefix+s) for s in ('.bin', '.png', '.txt', '.log')}})
logs = {n: Path('workplace/'+n).read_text() for n in ('round52-limited-final2.log', 'round52-standard-final2.log')}
assert all('FAIL' not in s for s in logs.values())
assert '--- SKIP:' not in logs['round52-limited-final2.log']
assert len(re.findall(r'^(ok\s|\?\s)', logs['round52-standard-final2.log'], re.M)) == 18
sources = ['internal/game/game.go', 'internal/game/battlecmd.go', 'internal/game/battlecmd_test.go',
           'internal/screen/battle.go', 'internal/screen/minimap_test.go',
           'tools/battle_hover_marks_oracle.sh', 'tools/scripts/battle_hover_marks_ida.py',
           'tools/scripts/battle_hover_marks_receipt.py', 'tools/scripts/battle_normal_probe.go',
           'tools/scripts/battle_normal_probe_seed.py', 'tools/turn_counter_gui.sh',
           'workplace/round52-normal-overlay.json', 'workplace/round52-normal-seed.go']
data = {'schema': 'logh3.battle-hover-marks.v1', 'date': '2026-10-03',
        'parent_commit': '8eab9ba896189159d437e974453ed1c02d19cc78',
        'remake_head': '8eab9ba896189159d437e974453ed1c02d19cc78',
        'remake_worktree': 'dirty; 正式來源身份記在source_hashes',
        'status': 'limited-production-validation; first entry retains one draw-order pixel',
        'comparison_level': 'nearby-state', 'issue': 'https://github.com/wicanr2/logh3/issues/89',
        'dosgolem_head': old['dosgolem_head'], 'dosgolem_dirty': False,
        'original_bundle_sha256': old['original_bundle_sha256'], 'original_executable_sha256': old['original_executable_sha256'],
        'toolchain': old['toolchain'], 'previous_receipt_sha256': sha(previous),
        'original_runs': originals, 'normal_player_metadata': metadata,
        'normal_player_runs': [{'path': str(p), 'hashes': {n: sha(p/n) for n in names}} for p in players],
        'ida': {'export_path': str(ida_path), 'export_sha256': sha(ida_path), 'evidence': ida},
        'display_order_diagnostic': {'objects': objects, 'remaining_pixel_xy': [255, 255],
                                     'classification': 'proven observed direction covers current highlight; general FF key contract unknown'},
        'frames': frames, 'tables': tables, 'gui': gui,
        'verification': {'limited_pass_events': len(re.findall(r'^\s*--- PASS:', logs['round52-limited-final2.log'], re.M)),
                         'standard_product_packages': 18, 'log_sha256': {n: sha('workplace/'+n) for n in logs}},
        'source_hashes': {p: sha(p) for p in sources},
        'normalization': '640x480原生色號與原版有效RGB色盤；無裁切、縮放或遮罩',
        'limitations': ['初始(255,255)方向物件覆蓋差1 pixel，所有FF物件排序契約未審，不改像素常數',
                       '其餘四張靜態全圖零差異不證明完整記錄／正常索敵／八步parity',
                       '四案原版RAM疊圖回歸只驗正式primitive；正常Game五案另列，不混稱',
                       '敵方／行星／特殊艦種、模態退出及非法座標未新增原版驗收',
                       '原版輸入用快照、明示target注入診斷；正常Game只以測試期Go覆蓋固定策略seed']}
out.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print(out, sha(out), data['verification'])
