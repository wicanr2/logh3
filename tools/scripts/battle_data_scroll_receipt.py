#!/usr/bin/env python3
"""第57輪固定收據；在tools/gui.sh xrun內執行，不覆寫既有JSON。"""
import hashlib
import json
import re
import struct
import zlib
from collections import Counter
from pathlib import Path
from pngout import game_palette


def sha(p):
    return hashlib.sha256(Path(p).read_bytes()).hexdigest()


def png(path, rgb):
    def chunk(kind, body):
        return struct.pack('>I', len(body)) + kind + body + struct.pack('>I', zlib.crc32(kind + body))
    raw = b''.join(b'\0' + rgb[y * 1920:(y + 1) * 1920] for y in range(480))
    Path(path).write_bytes(b'\x89PNG\r\n\x1a\n' + chunk(b'IHDR', struct.pack('>IIBBBBB', 640, 480, 8, 2, 0, 0, 0)) + chunk(b'IDAT', zlib.compress(raw)) + chunk(b'IEND', b''))


out = Path('docs/internals/data/battle-data-scroll.json')
assert not out.exists(), '固定收據不以新來源重生'
image_dir = Path('workplace/round57-comparison')
assert not image_dir.exists(), '差分輸出不覆寫'
assert image_dir.parent.stat().st_uid == 1000
image_dir.mkdir()
expected = {
    'zero': '8fdb9d1b89c905dc50ed638aa19dd69a63d3949c5519cc8e35790aa1e154eecb',
    'one': 'c62da2e5eb1b406bb6b5d418a75d795b9252ac63c82801a86244c197a7941550',
    'last': 'ed8e1a575a412db0fc95acffbe124c773bbd7cf3978c32278de4459159db0782',
}
parent = Path('workplace/oracle/data67-r57-parent')
assert 'f7523f9（未提交檔案 0 個）' in (parent / 'meta.txt').read_text()
parent_flags = (parent / 'flags.txt').read_text().splitlines()
assert parent_flags[parent_flags.index('-poke') + 1] == 'lin:5E00C@300000001=7E 16 DC D3'
assert len((parent / 'mem.bin').read_bytes()) == 1048576
players = [Path('workplace/round57-player-' + side) for side in 'ab']
player_files = sorted(str(p.relative_to(players[0])) for p in players[0].rglob('*') if p.is_file())
assert len(player_files) == 27
assert player_files == sorted(str(p.relative_to(players[1])) for p in players[1].rglob('*') if p.is_file())
assert all((players[0] / n).read_bytes() == (players[1] / n).read_bytes() for n in player_files)
frames = []
memories = {}
regions = {'whole': [0, 0, 640, 480], 'table_window': [16, 60, 448, 360],
           'minimap': [480, 75, 128, 128], 'status': [464, 220, 160, 200]}
for name, offset in [('zero', 0), ('one', 1), ('last', 26)]:
    originals = [Path('workplace/oracle/data67-r57-' + name + '-' + side) for side in 'ab']
    for n in ('mem.bin', 'screen.bin', 'screen.png', 'pal.bin'):
        assert (originals[0] / n).read_bytes() == (originals[1] / n).read_bytes(), (name, n)
    original = (originals[0] / 'screen.bin').read_bytes()
    assert len(original) == 307200 and sha(originals[0] / 'screen.bin') == expected[name]
    remake = (players[0] / name / 'screen.bin').read_bytes()
    assert remake == original
    mem = (originals[0] / 'mem.bin').read_bytes()
    memories[name] = mem
    assert len(mem) == 1048576
    assert struct.unpack_from('<HH', mem, 0x5E6B6) == (offset, 40)
    assert struct.unpack_from('<HH', mem, 0x5E724) == (1312, 1660)
    assert struct.unpack_from('<I', mem, 0x5E00C)[0] == 0xB6E4B78A
    assert len((originals[0] / 'pal.bin').read_bytes()) == 768
    # pal.bin是原始VGA色盤；實際16色PNG另經屬性映射，不由DAC位深猜色。
    pal = bytes(game_palette(str(originals[0] / 'screen.png')))
    assert len(pal) == 48 and any(pal) and max(remake) < 16
    rgb = bytes(v for i in remake for v in pal[i * 3:i * 3 + 3])
    png(image_dir / (name + '-remake.png'), rgb)
    png(image_dir / (name + '-diff.png'), bytes(640 * 480 * 3))
    before_path = Path('workplace/round57-ally-player-header-explore') / {'zero': 'data0', 'one': 'data1', 'last': 'data26'}[name]
    before = (before_path / 'screen.bin').read_bytes()
    before_diff = [i for i, (a, b) in enumerate(zip(original, before)) if a != b]
    assert len(before_diff) == {'zero': 1189, 'one': 1094, 'last': 962}[name]
    png(image_dir / (name + '-before.png'), bytes(v for i in before for v in pal[i * 3:i * 3 + 3]))
    png(image_dir / (name + '-before-diff.png'), bytes(v for a, b in zip(original, before) for v in ((255, 0, 0) if a != b else (0, 0, 0))))
    tables = {}
    for n, address, size, stride in [('WORK', 0x5DA44, 230, 230), ('FLEET', 0x60BA4, 3904, 122),
                                     ('SHIP', 0x668B2, 27456, 78), ('STELLAR', 0x5F2CC, 4032, 126), ('PLANET', 0x53132, 6642, 82)]:
        a, b = mem[address:address + size], (players[0] / name / (n + '.DT0')).read_bytes()
        assert len(a) == len(b)
        assert (before_path / (n + '.DT0')).read_bytes() == b, '純呈現修正不得改資料'
        differences = [i for i, (x, y) in enumerate(zip(a, b)) if x != y]
        if n == 'FLEET':
            assert len(differences) == 27 and set(i % stride for i in differences) <= set(range(0x6C, 0x70))
        elif n == 'SHIP':
            assert len(differences) == 214 and set(i % stride for i in differences) <= set(range(0x38, 0x3C))
        else:
            assert not differences
        tables[n] = {'linear_hex': hex(address), 'bytes': size, 'different_bytes': len(differences),
                     'offset_counts': dict(Counter(i % stride for i in differences)),
                     'original_sha256': hashlib.sha256(a).hexdigest(), 'remake_sha256': hashlib.sha256(b).hexdigest()}
    runs = []
    for p in originals:
        assert 'f7523f9（未提交檔案 0 個）' in (p / 'meta.txt').read_text()
        flags = (p / 'flags.txt').read_text().splitlines()
        assert '-load-state' in flags and '-tick' in flags
        runs.append({'path': str(p), 'flags': flags, 'hashes': {n: sha(p / n) for n in ('mem.bin', 'machine.state', 'screen.bin', 'screen.png', 'pal.bin', 'flags.txt', 'meta.txt', 'report.txt', 'writes.txt')}})
    frames.append({'name': name, 'offset': offset, 'total': 40, 'camera_world_pixels': [1312, 1660],
                   'seed_after_hex': 'B6E4B78A', 'original_runs': runs,
                   'remake_screen_sha256': sha(players[0] / name / 'screen.bin'),
                   'before_diagnostic': {'path': str(before_path), 'screen_sha256': sha(before_path / 'screen.bin'),
                                         'different_pixels': len(before_diff), 'classification': '正常Game探針，測試覆蓋只固定相同seed；本輪五表與正式路徑相同'},
                   'regions': {n: {'rect': r, 'compared_pixels': r[2] * r[3], 'same_pixels': r[2] * r[3],
                                   'different_pixels': 0, 'max_channel_difference': 0, 'difference_bbox': None} for n, r in regions.items()},
                   'images': {str(image_dir / (name + suffix)): sha(image_dir / (name + suffix)) for suffix in ('-remake.png', '-diff.png', '-before.png', '-before-diff.png')},
                   'table_comparison': tables})
ida = []
semantics = {
    ('113A', 0x21D5): '玩家陣營分支', ('113A', 0x21E3): '帝國基準低word減同盟',
    ('113A', 0x2243): '同盟基準低word減帝國', ('113A', 0x226A): '同盟玩家先印同盟字串',
    ('113A', 0x2286): '同盟玩家後印帝國字串', ('113A', 0x2872): '表列指向呼叫鏡頭置中',
    ('1588', 0x032D): '實際鏡頭X寫入', ('1588', 0x0338): '實際鏡頭Y寫入',
    ('1727', 0x0363): '讀鏡頭世界X', ('1727', 0x0369): 'X右移四位', ('1727', 0x036B): '小地圖X基底480',
    ('1727', 0x0370): '讀鏡頭世界Y', ('1727', 0x0373): 'Y右移四位', ('1727', 0x0375): 'DOS小地圖Y基底36',
    ('1727', 0x03E7): '右邊界X加27', ('1727', 0x043B): '底邊界Y加21',
}
proven = set()
for p in sorted(Path('workplace/ida').glob('data67-r57-*.json')):
    d = json.loads(p.read_text())
    assert d['kernel_version'] == '9.4' and d['uid'] == 1000 and d['ida_base_hex'] == '0x10200'
    for r in d['instructions']:
        address = int(r['runtime_linear_hex'], 16)
        raw = bytes.fromhex(r['original_bytes_hex'])
        assert memories['zero'][address:address + len(raw)] == raw, (p, r['original_location'])
        key = (d['runtime_segment'], int(r['original_location'].split(':')[1], 16))
        if key in semantics:
            r.update({'grade': 'proven', 'semantic': semantics[key], 'warning': ''})
            proven.add(key)
    ida.append({'path': str(p), 'sha256': sha(p), 'data': d})
assert len(ida) == 6 and proven == set(semantics)
gui = []
for scale in (1, 2):
    prefix = f'workplace/round57-data-gui-{scale}x'
    assert Path(prefix + '.bin').read_bytes() == (players[0] / 'one/screen.bin').read_bytes()
    state = Path(prefix + '.txt').read_text()
    assert 'state=StateBattle\n' in state and 'cursor=440,390\n' in state
    assert f'window={640 * scale}x{480 * scale}\n' in state
    gui.append({'scale': scale, 'prefix': prefix, 'hashes': {s: sha(prefix + s) for s in ('.bin', '.png', '.txt', '.log')}})
verification = {}
for label in ('player-a', 'player-b', 'standard'):
    p = Path('workplace/round57-' + label + '.log')
    s = p.read_text()
    assert 'FAIL' not in s and 'SKIP' not in s and 'no tests to run' not in s
    tested = len(re.findall(r'^ok\s', s, re.M))
    empty = len(re.findall(r'^\?\s.*\[no test files\]$', s, re.M))
    assert tested + empty == (18 if label == 'standard' else 1)
    verification[label] = {'path': str(p), 'sha256': sha(p), 'tested_packages': tested, 'no_test_packages': empty}
sources = ['internal/game/battledata.go', 'internal/game/game.go', 'internal/game/battledata_wholeframe_test.go',
           'internal/game/rebellion.go', 'internal/game/originalbrain.go', 'internal/screen/battle.go', 'internal/screen/battledata.go',
           'tools/battle_data_scroll_oracle.sh', 'tools/battle_data_scroll_gui.sh', 'tools/scripts/battle_data_scroll_ida.py',
           'tools/scripts/battle_data_scroll_receipt.py', 'tools/scripts/pngout.py', 'tools/gui.sh', 'tools/ida.sh', 'tools/oracle.sh']
data = {'schema': 'logh3.battle-data-scroll.v1', 'date': '2026-10-03', 'issue': 67,
        'parent_commit': 'd070f01ee90ff7c4bf914f62edf80b657a8be633', 'remake_worktree': 'dirty; 本輪呈現修正及驗證',
        'comparison_level': 'same-state; 限正常S3同盟四十列資料表三個靜態畫面，不稱完整戰場同態',
        'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', 'dosgolem_dirty': False,
        'toolchain': {'GUI': {'tag': 'logh3-ebiten:2', 'image_id': 'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8'},
                      'Go': {'tag': 'golang:1.24-bookworm', 'image_id': 'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac'},
                      'IDA': {'tag': 'ida-pro-9.4-idapython:locked-v1', 'image_id': 'sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780', 'kernel': '9.4'}},
        'original_bundle_sha256': 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
        'original_executable_sha256': '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918',
        'original_parent': {'path': str(parent), 'flags': parent_flags, 'hashes': {n: sha(parent / n) for n in ('mem.bin', 'machine.state', 'screen.bin', 'screen.png', 'pal.bin', 'flags.txt', 'meta.txt', 'report.txt')}},
        'seed_setup': {'original': '單一-poke lin:5E00C@300000001=7E 16 DC D3', 'remake': '正式預設strategyRandSeed D3DC167E；無測試覆蓋或注入', 'after_hex': 'B6E4B78A'},
        'normal_inputs': ['主選單取進度300,270', '138; enter', 'S3 300,322', '同盟318,228', '實行544,125及560,125', '808個60Hz tick', '選單400,80；資料400,100', '游標440,390', '短點0／1／27次下箭頭'],
        'coordinate_mapping': '原版滑鼠DOS y加40為640×480畫面y；邏輯畫布相同，無裁切／縮放／遮罩',
        'date_and_phase': {'date': [799, 3, 1], 'strategic_turn': 380, 'battle_phase': 'Input'},
        'frames': frames, 'normal_player_runs': [{'path': str(p), 'hashes': {n: sha(p / n) for n in player_files}} for p in players],
        'charplane_cells': [{'cell_xy': [x, y], 'runtime_linear_hex': hex(0x54B30 + 0xD81E + (y * 40 + x) * 6),
                             'raw_hex': memories['zero'][0x54B30 + 0xD81E + (y * 40 + x) * 6:0x54B30 + 0xD81E + (y * 40 + x) * 6 + 6].hex(),
                             'grade': 'proven', 'source': 'zero-a/mem.bin；字碼word +2、顏色word +4'} for x, y in [(4, 4), (27, 4), (27, 17)]],
        'ida_exports': ida, 'instruction_count': sum(len(d['data']['instructions']) for d in ida), 'proven_semantics_count': len(proven),
        'normal_gui': gui, 'verification': verification, 'sources_sha256': {n: sha(n) for n in sources},
        'limits': ['FLEET27／SHIP214只剩顯示指標raw差，完整表仍記錄；不稱DOS指標一致',
                   '本輪表頭輸入避開尚未實作的表列指向置中，該行為不由畫面通過升格',
                   '按住連捲#60、其他頁籤及正常首戰一像素排序取捨仍待；087／088與#79／#89保持原閘門',
                   'Linux Xvfb GUI不代表Windows／macOS實機或音訊聆聽驗收']}
out.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')
print('固定收據', sha(out), '三張完整畫面零差，IDA', data['instruction_count'], '原指令，proven', len(proven))
