"""第51輪原版雙跑、正常入口全圖與原始IDA收據；tools/gui.sh xrun執行。"""
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
            'unresolved_pixels': [{'xy': [i%640, i//640], 'original_index': a[i], 'remake_index': b[i]} for i in ds] if len(ds) <= 10 else []}


out = Path('docs/internals/data/battle-input-display.json')
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
baseline = Path('docs/internals/data/fleet-arrival-placement.json')
assert sha(baseline) == '58fd315475e4756f671078e71044d319078689a53b2c6f7d25ce2f24a8516a2a'
originals, frames = {}, {}
entry = Path('workplace/oracle/normal89-r51-base-a/entry.bin').read_bytes()
assert len(entry) == 1048576
assert hashlib.sha256(entry).hexdigest() == 'ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97'
players = [Path('workplace/round51-final-'+s) for s in 'ab']
names = sorted(p.relative_to(players[0]).as_posix() for p in players[0].rglob('*') if p.is_file())
assert len(names) == 55
assert names == sorted(p.relative_to(players[1]).as_posix() for p in players[1].rglob('*') if p.is_file())
assert all((players[0]/n).read_bytes() == (players[1]/n).read_bytes() for n in names)
meta = json.loads((players[0]/'probe.json').read_text())
assert meta['date'] == [799, 3, 1] and meta['strategy_seed_after'] == 'B6E4B78A'
for kind, display, expected, hover in [('base', 'entry', 2, 80), ('flagship', 'flagship', 1, 80),
                                      ('squad', 'squad', 1, 82), ('empty', 'empty', 0, 65535), ('panel', 'panel', 0, 65535)]:
    ps = [Path(f'workplace/oracle/normal89-r51-{kind}-{s}') for s in 'ab']
    files = ['entry.bin' if kind == 'base' else 'end.bin', 'screen.bin', 'screen.png', 'pal.bin']
    assert all((ps[0]/n).read_bytes() == (ps[1]/n).read_bytes() for n in files)
    originals[kind] = []
    for p in ps:
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in (p/'meta.txt').read_text()
        m = (p/files[0]).read_bytes()
        assert len(m) == 1048576
        assert all(entry[a:a+n] == m[a:a+n] for a, n in [(0x60ba4, 3904), (0x668b2, 27456), (0x5da44, 230)])
        assert struct.unpack_from('<I', m, 0x5e00c)[0] == 0xb6e4b78a
        assert struct.unpack_from('<H', m, 0x54b30+0x9b7a)[0] == hover
        originals[kind].append({'path': str(p), 'command_argv': (p/'flags.txt').read_text().splitlines(),
                               'hashes': {n: sha(p/n) for n in files+['machine.state', 'meta.txt', 'flags.txt', 'report.txt']}})
    a = (ps[0]/'screen.bin').read_bytes()
    b = (players[0]/display/'screen.bin').read_bytes()
    assert len(a) == len(b) == 307200
    pal = bytes(game_palette(str(ps[0]/'screen.png')))
    regions = {n: stats(a, b, pal, r) for n, r in {'full': (0, 0, 640, 480), 'field': (0, 0, 448, 360),
               'right': (448, 0, 192, 480), 'bottom': (0, 360, 448, 120),
               'unit_panel': (480, 242, 128, 160)}.items()}
    assert regions['full']['different_index_pixels'] == expected
    assert regions['unit_panel']['different_index_pixels'] == 0
    base = 'workplace/round51-'+display
    assert not Path(base+'.png').exists()
    write_png(base+'.png', 640, 480, b, pal)
    write_png(base+'-diff.png', 640, 480, bytes(int(x != y) for x, y in zip(a, b)), bytes([0, 0, 0, 255, 0, 0]))
    write_png(base+'-pair.png', 1280, 480, b''.join(a[y*640:(y+1)*640]+b[y*640:(y+1)*640] for y in range(480)), pal)
    frames[display] = {'regions': regions, 'images_sha256': {s: sha(base+s) for s in ('.png', '-diff.png', '-pair.png')},
                       'classification': 'nearby-state; 完整資料仍有差；非全入口same-state'}
    for name in ('FLEET.DT0', 'SHIP.DT0', 'WORK.DT0'):
        assert (players[0]/display/name).read_bytes() == (players[0]/'entry'/name).read_bytes()
tables = {}
for name, addr, stride, count in [('WORK', 0x5da44, 230, 1), ('FLEET', 0x60ba4, 122, 32),
                                ('SHIP', 0x668b2, 78, 352), ('STELLAR', 0x5f2cc, 126, 32), ('PLANET', 0x53132, 82, 81)]:
    a = entry[addr:addr+stride*count]
    b = (players[0]/'entry'/(name+'.DT0')).read_bytes()
    ds = [i for i, (x, y) in enumerate(zip(a, b)) if x != y]
    tables[name] = {'bytes': len(a), 'different_bytes': len(ds), 'offset_counts': dict(Counter(i%stride for i in ds)),
                    'original_sha256': hashlib.sha256(a).hexdigest(), 'remake_sha256': hashlib.sha256(b).hexdigest()}
assert [tables[n]['different_bytes'] for n in ('WORK', 'FLEET', 'SHIP', 'STELLAR', 'PLANET')] == [14, 27, 276, 0, 0]
ida = []
for seg in ('12F1', '1CC5', '25DB'):
    p = Path(f'workplace/ida/normal89-r51-{seg}-graded.json')
    j = json.loads(p.read_text())
    assert j['kernel_version'] == '9.4' and j['uid'] == 1000
    assert j['input_sha256'] == sha(Path(j['input'].replace('/work/', 'workplace/ida/')))
    for r in j['instructions']:
        a, raw = int(r['runtime_linear_hex'], 16), bytes.fromhex(r['original_bytes_hex'])
        assert entry[a:a+len(raw)] == raw
    ida.append({'export_path': str(p), 'export_sha256': sha(p), 'evidence': j})
assert sum(len(j['evidence']['functions']) for j in ida) == 5
assert sum(len(j['evidence']['instructions']) for j in ida) == 535
assert sum(r['grade'] == 'proven' for j in ida for r in j['evidence']['instructions']) == 22
gui = []
for scale in (1, 2):
    base = f'workplace/round51-gui-{scale}x'
    st = Path(base+'.txt').read_text()
    assert 'state=StateBattle' in st and f'window={640*scale}x{480*scale}' in st
    assert Path(base+'.bin').read_bytes() == (players[0]/'entry/screen.bin').read_bytes()
    gui.append({'scale': scale, 'path': base, 'normal_player_path': True,
                'hashes': {s: sha(base+s) for s in ('.bin', '.png', '.txt', '.log')}})
logs = {n: Path('workplace/'+n).read_text() for n in ['round51-limited.log', 'round51-standard-product.log', 'round51-hover-test.log']}
assert all('FAIL' not in s for s in logs.values())
assert '--- SKIP:' not in logs['round51-limited.log']+logs['round51-hover-test.log']
assert len(re.findall(r'^(ok\s|\?\s)', logs['round51-standard-product.log'], re.M)) == 18
sources = ['internal/game/'+n+'.go' for n in ['game', 'battle', 'battlecmd', 'battlecmd_test', 'rebellion']]
sources += ['internal/screen/battle.go', 'tools/battle_input_display_oracle.sh', 'tools/turn_counter_gui.sh',
            'tools/scripts/battle_input_display_ida.py', 'tools/scripts/battle_input_display_receipt.py',
            'tools/scripts/battle_normal_probe.go', 'tools/scripts/battle_normal_probe_seed.py',
            'workplace/round51-normal-overlay.json', 'workplace/round51-normal-seed.go']
data = {'schema': 'logh3.battle-input-display.v1', 'date': '2026-10-03',
        'parent_commit': '4be3b1b8f37dd1b9d16331cc2aaaafa2285bce87',
        'remake_head': '4be3b1b8f37dd1b9d16331cc2aaaafa2285bce87',
        'remake_worktree': 'dirty; 本輪正式來源身份逐檔記在source_hashes',
        'status': 'limited-production-validation; remaining full-frame pixels unknown',
        'issue': 'https://github.com/wicanr2/logh3/issues/89', 'comparison_level': 'nearby-state',
        'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', 'dosgolem_dirty': False,
        'original_bundle_sha256': 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
        'original_executable_sha256': sha('org_game/extracted/GIN3.COM'),
        'toolchain': json.loads(baseline.read_text())['toolchain'], 'previous_receipt_sha256': sha(baseline),
        'source_snapshot_sha256': json.loads(baseline.read_text())['source_snapshot_sha256'],
        'original_runs': originals, 'normal_player_metadata': meta,
        'normal_player_runs': [{'path': str(p), 'hashes': {n: sha(p/n) for n in names}} for p in players],
        'ida': ida, 'frames': frames, 'tables': tables, 'gui': gui,
        'verification': {'limited_pass_events': len(re.findall(r'^\s*--- PASS:', logs['round51-limited.log'], re.M)),
                         'hover_pass_events': 1, 'standard_product_packages': 18,
                         'log_sha256': {n: sha('workplace/'+n) for n in logs}},
        'source_hashes': {p: sha(p) for p in sources},
        'normalization': '640x480原生色號與原版有效RGB色盤；無裁切、縮放或遮罩',
        'limitations': ['初始圖2 pixels、旗艦及分隊各1 pixel仍unknown，完整首戰未通過',
                       '敵方／行星／特殊艦種、重疊命中及模態退出尚未新增對拍',
                       '完整原池far指標、顯示快取及八步仍由#89／#79接續',
                       'normal-player Go覆蓋只固定策略seed；產品正式RNG不鎖定',
                       'GUI未固定策略seed，既有S3自然分支與Game結果相同，非原版跨流程骰序聲明']}
out.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print(out, sha(out), data['verification'])
