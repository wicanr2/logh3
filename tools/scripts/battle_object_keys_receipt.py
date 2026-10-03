"""第56輪：顯示物件關係、殘留排序鍵反例與可丟棄呈現探針。"""
import hashlib
import json
import re
import struct
from pathlib import Path

ROOT = Path.cwd()
OUTPUT = ROOT / 'docs/internals/data/battle-object-keys.json'
assert not OUTPUT.exists(), '不可覆寫凍結收據'


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def word(mem, address):
    return struct.unpack_from('<H', mem, address)[0]


def pointer(mem, address):
    return word(mem, address)+16*word(mem, address+2)


def files(root, names):
    return {n: sha(root/n) for n in names}


def run(kind, side):
    directory = ROOT/f'workplace/oracle/normal89-r56-{kind}-{side}'
    hashes = files(directory, ['entry.bin', 'screen.bin', 'screen.png', 'pal.bin',
                              'flags.txt', 'meta.txt', 'report.txt', 'machine.state'])
    mem = (directory/'entry.bin').read_bytes()
    assert len(mem) == 1048576
    flags = (directory/'flags.txt').read_text().splitlines()
    assert '-tick' in flags and flags[flags.index('-tick')+1] == '20000'
    meta = (directory/'meta.txt').read_text()
    assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in meta
    report = (directory/'report.txt').read_text()
    assert 'lin:5E00C（5E00C）' in report
    assert flags.count('-poke') == 1
    assert mem[0x5E00C:0x5E010].hex() == '8ab7e4b6'
    if kind == 'key-high-joined':
        assert 'lin:71022（71022）2 bytes' in report
        assert word(mem, 0x71022) == 65535
    else:
        assert word(mem, 0x71022) == 64
    writes = []
    for m in re.finditer(r'^\[watch\] #(\d+) ([0-9A-F]+): ([0-9A-F]+) → ([0-9A-F]+)  ← ([0-9A-F]+:[0-9A-F]+)', report, re.M):
        writes.append({'step': int(m[1]), 'linear_hex': m[2], 'before_hex': m[3],
                       'after_hex': m[4], 'post_instruction_cs_ip': m[5]})
    objects = []
    count = word(mem, 0x54B30+0xEADE)
    assert count == 64
    for i in range(count):
        address = pointer(mem, 0x54B30+0xEAE0+4*i)
        raw = mem[address:address+40]
        assert len(raw) == 40
        objects.append({'sorted_position': i, 'linear_hex': hex(address),
                        'raw_hex': raw.hex(), 'priority': raw[3],
                        'word_key': word(mem, address+4), 'sprite': raw[12],
                        'view_relative_xy': [word(mem, address+6), word(mem, address+8)]})
    keys = [(o['priority'], -o['word_key']) for o in objects]
    assert keys == sorted(keys)
    assert len({o['linear_hex'] for o in objects}) == 64
    return {'kind': kind, 'side': side, 'path': str(directory.relative_to(ROOT)),
            'classification': 'normal-player-route' if kind == 'base' else 'explicit-key-injection-diagnostic',
            'hashes': hashes, 'flags': flags, 'writes': writes, 'objects': objects}, mem


runs = []
memories = {}
for kind in ['base', 'key-high-joined']:
    a, ma = run(kind, 'a')
    b, mb = run(kind, 'b')
    for name in ['entry.bin', 'screen.bin', 'screen.png', 'pal.bin']:
        assert a['hashes'][name] == b['hashes'][name], (kind, name)
    assert a['objects'] == b['objects'] and a['writes'] == b['writes']
    runs.extend([a, b])
    memories[kind] = ma
mem = memories['base']
assert sha(ROOT/'workplace/oracle/normal89-r56-base-a/entry.bin') == 'ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97'
original = ROOT/'workplace/oracle/normal89-r56-base-a'
changed = ROOT/'workplace/oracle/normal89-r56-key-high-joined-a'
screen = (original/'screen.bin').read_bytes()
other = (changed/'screen.bin').read_bytes()
changed_pixels = [{'xy': [i % 640, i // 640], 'base': a, 'key_high': b}
                  for i, (a, b) in enumerate(zip(screen, other)) if a != b]
assert changed_pixels == [{'xy': [255, 255], 'base': 12, 'key_high': 5}]

table_shapes = {'WORK': (0x5DA44, 230, None), 'FLEET': (0x60BA4, 32*122, (122, 0x6C, 0x70)),
                'SHIP': (0x668B2, 352*78, (78, 0x38, 0x3C)),
                'STELLAR': (0x5F2CC, 32*126, None), 'PLANET': (0x53132, 81*82, None)}
tables = {}
for name, (address, length, pointer_lane) in table_shapes.items():
    raw = mem[address:address+length]
    assert raw == memories['key-high-joined'][address:address+length]
    saved = (ROOT/f'workplace/round55-pass-a/entry/{name}.DT0').read_bytes()
    assert len(saved) == length
    diffs = [i for i, (a, b) in enumerate(zip(raw, saved)) if a != b]
    if pointer_lane:
        stride, lo, hi = pointer_lane
        assert all(lo <= i % stride < hi for i in diffs)
    else:
        assert not diffs
    tables[name] = {'bytes': length, 'original_sha256': hashlib.sha256(raw).hexdigest(),
                    'remake_sha256': hashlib.sha256(saved).hexdigest(), 'different_byte_offsets': diffs,
                    'status': 'raw-differences-retained; pointer-comparison-contract-pending' if diffs else 'all-bytes-equal'}
assert len(tables['FLEET']['different_byte_offsets']) == 27
assert len(tables['SHIP']['different_byte_offsets']) == 214

relationships = []
registered = {int(o['linear_hex'], 16): o for o in runs[0]['objects']}
for fleet in [12, 16, 17, 28, 29, 30, 31]:
    fa = 0x60BA4+fleet*122
    direction = pointer(mem, fa+0x6C)
    assert direction in registered and registered[direction]['sprite'] == mem[fa+0x6A]
    pools = [word(mem, fa+0x10+4*i) for i in range(mem[fa+0x0F])]+[word(mem, fa+0x38)]
    units = []
    for pool in pools:
        sa = 0x668B2+pool*78
        obj = pointer(mem, sa+0x38)
        assert obj in registered
        row = registered[obj]
        assert row['priority'] == mem[sa+0x30] == 128
        assert row['word_key'] == mem[sa+1] | (mem[sa+3] << 8)
        assert row['sprite'] == mem[sa+0x2F]
        assert row['view_relative_xy'] == [word(mem, sa+0x34)-192, word(mem, sa+0x36)-100]
        units.append({'pool': pool, 'ship_pointer_hex': mem[sa+0x38:sa+0x3C].hex(), 'object': row})
    relationships.append({'fleet': fleet, 'fleet_pointer_hex': mem[fa+0x6C:fa+0x70].hex(),
                          'direction_object': registered[direction], 'unit_objects': units})
assert sum(len(f['unit_objects']) for f in relationships) == 55

exports = []
semantics = {
    '35D1:0029': '註冊物件segment至DS:EAE0清單',
    '35D1:002D': '註冊物件offset至DS:EAE0清單',
    '35D1:0031': '物件數DS:EADE增加1',
    '35D1:0053': '註冊完成後呼叫028D排序同一清單',
    '35D1:028D': '比較器segment為35D1',
    '35D1:0290': '比較器offset為02AB',
    '35D1:0299': '排序長度取DS:EADE物件數',
    '35D1:029E': '排序來源為DS:EAE0遠指標清單',
    '35D1:02A2': '把清單、4-byte元素寬、長度及比較器交給排序服務',
    '35D1:02B4': '第一物件讀+3 unsigned byte順位',
    '35D1:02BE': '第二物件讀+3 unsigned byte順位',
    '35D1:02C4': '第一順位較小時回傳負值，先畫',
    '35D1:02E1': '第一順位較大時回傳正值，後畫',
    '35D1:02EE': '相同順位讀第一物件+4 unsigned word鍵',
    '35D1:02F8': '相同順位讀第二物件+4 unsigned word鍵',
    '35D1:02FE': '第一字鍵較大時回傳負值，先畫',
    '35D1:031B': '第一字鍵較小時回傳正值，後畫',
    '35D1:0322': '兩鍵相等回傳0；相等鍵置換仍未驗',
    '25DB:0109': '方向物件segment保存FLEET+6E',
    '25DB:010D': '方向物件offset保存FLEET+6C',
    '25DB:012A': '方向建立器初始化+3為F0，沒有設定+4字鍵',
    '25DB:0165': '方向物件交給35D1:0009註冊',
    '25DB:0198': '分隊物件segment保存SHIP+3A',
    '25DB:019C': '分隊物件offset保存SHIP+38',
    '25DB:01C9': '分隊物件+3初始化80',
    '25DB:01D0': '分隊序號構成物件+4 word高byte',
    '25DB:01DC': '艦隊索引構成物件+4 word低byte',
    '25DB:01E7': '分隊物件明確寫入+4排序鍵',
    '25DB:06B5': '方向位置更新經FLEET+6C遠指標取得物件',
    '25DB:06E4': '方向物件X由艦隊X減視角加方向位移',
    '25DB:0709': '方向物件Y由艦隊Y減視角加方向位移',
    '25DB:0C0C': '目前位置反白物件+3設FF；本建立器沒有初始化+4字鍵',
    '25DB:0C8E': '集結點反白物件+3設FE；本建立器沒有初始化+4字鍵',
}
for name, lo, hi in [('normal89-r56-35D1-register.json', 9, 0x333),
                     ('normal89-r54-25DB-graded.json', 0xFB, 0x221),
                     ('normal89-r54-25DB-graded.json', 0x670, 0x714),
                     ('normal89-r51-25DB-graded.json', 0xBD2, 0xD11)]:
    path = ROOT/'workplace/ida'/name
    export = json.loads(path.read_text())
    rows = [r for r in export['instructions'] if lo <= int(r['original_location'].split(':')[1], 16) < hi]
    for row in rows:
        address = int(row['runtime_linear_hex'], 16)
        raw = bytes.fromhex(row['original_bytes_hex'])
        assert mem[address:address+len(raw)] == raw
        row['semantic'] = semantics.get(row['original_location'], '未審查')
        row['grade'] = 'proven' if row['original_location'] in semantics else 'unknown'
        row['warning'] = '' if row['grade'] == 'proven' else '未證實語意；掃描範圍不是函式邊界'
        row['source'] = '研究394；原始bytes資料流、雙跑物件清單及字鍵反例'
    exports.append({'path': str(path.relative_to(ROOT)), 'sha256': sha(path),
                    'input_sha256': export['input_sha256'], 'kernel_version': export['kernel_version'],
                    'ida_base_hex': export['ida_base_hex'], 'range_hex': [hex(lo), hex(hi)],
                    'instructions': rows})

frames = {}
previous = json.loads((ROOT/'docs/internals/data/battle-work-control.json').read_text())
for name, old in previous['frames'].items():
    a = ROOT/f'workplace/round55-pass-a/{name}/screen.bin'
    b = ROOT/f'workplace/round56-render-probe/{name}/screen.bin'
    orig = ROOT/old['original']/'screen.bin'
    aa, bb, oo = a.read_bytes(), b.read_bytes(), orig.read_bytes()
    assert len(aa) == len(bb) == len(oo) == 307200
    before = sum(x != y for x, y in zip(aa, oo))
    after = sum(x != y for x, y in zip(bb, oo))
    assert before == (1 if name == 'entry' else 0) and after == 0
    frames[name] = {'original': str(orig.relative_to(ROOT)), 'original_sha256': sha(orig),
                    'product': str(a.relative_to(ROOT)), 'product_sha256': sha(a),
                    'prototype': str(b.relative_to(ROOT)), 'prototype_sha256': sha(b),
                    'comparison': 'disposable-render-order-probe; not-production',
                    'before_different_pixels': before, 'prototype_different_pixels': after,
                    'compared_pixels': 307200, 'mask': None}

sources = ['tools/battle_object_keys_oracle.sh', 'tools/scripts/battle_object_keys_ida.py',
           'tools/scripts/battle_object_keys_receipt.py', 'internal/game/game.go',
           'internal/screen/battle.go', 'workplace/round56-input-before-directions.go',
           'workplace/round56-normal-overlay.json', 'workplace/round56-render-overlay.json',
           'workplace/round56-normal-seed.go', 'tools/scripts/battle_normal_probe.go']
data = {'schema': 'logh3.battle-object-keys.v1', 'date': '2026-10-03', 'issue': 89,
        'parent_commit': 'c9d4ce3fd8b5fbbb37743684bcbf13449f7d7dfd',
        'dosgolem_head': previous['dosgolem_head'], 'dosgolem_dirty': False,
        'original_bundle_sha256': previous['original_bundle_sha256'],
        'original_executable_sha256': previous['original_executable_sha256'],
        'toolchain': previous['toolchain'], 'status': 'evidence-proven; fidelity-decision-pending',
        'previous_receipts': {'docs/internals/data/battle-work-control.json': sha(ROOT/'docs/internals/data/battle-work-control.json')},
        'original_parent': previous['original_parent'], 'original_runs': runs,
        'explicit_key_diagnostic_full_screen': {'compared_pixels': 307200, 'changed_pixels': changed_pixels},
        'tables': tables, 'pointer_relationships': relationships,
        'ida_exports': exports, 'frames': frames, 'sources_sha256': {n: sha(ROOT/n) for n in sources},
        'proven_semantics_count': sum(row['grade'] == 'proven' for e in exports for row in e['instructions']),
        'limits': ['固定來源雙跑只證明指定局面', '沒有改production繪製次序',
                   '任意殘留鍵局面的固定次序不保證原版像素一致', '指標raw差異完整保留，未核准排除',
                   '完整088仍DRAFT；正常索敵與八步尚未完成']}
OUTPUT.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print('原版兩條分支各雙跑、64物件／55原池／7方向關係與一像素反例已保存', sha(OUTPUT))
