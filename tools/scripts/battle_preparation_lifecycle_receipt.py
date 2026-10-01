"""#79 原版生命週期收據核對及欄位匯出；入口見 docs/re/369。"""
import argparse
import hashlib
import json
import struct
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('--output', required=True)
parser.add_argument('--golem-head', required=True)
parser.add_argument('--golem-dirty', required=True)
args = parser.parse_args()
assert args.golem_head == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
assert args.golem_dirty == ''
root = Path('workplace/oracle')
def digest(raw):
    return hashlib.sha256(raw).hexdigest()
def equal(a, b):
    raw = a.read_bytes()
    assert raw == b.read_bytes(), (a, b, '雙跑不相同')
    return raw
def memory(path):
    raw = path.read_bytes()
    assert len(raw) == 1048576, (path, '不是完整1MiB')
    return raw
def ship(raw, index):
    at = 0x668B2 + 78 * index
    return {'pool_index': index, 'fleet': raw[at+1], 'kind': raw[at+4],
            'ships': struct.unpack_from('<H', raw, at+8)[0],
            'full_ships': struct.unpack_from('<H', raw, at+10)[0],
            'repair_cap': struct.unpack_from('<H', raw, at+12)[0],
            'missiles': struct.unpack_from('<I', raw, at+14)[0],
            'fighters': struct.unpack_from('<I', raw, at+18)[0]}
probe = equal(root/'battle79-life-probe-r2-a.json', root/'battle79-life-probe-r2-b.json')
assert digest(probe) == 'f1a95e05dbdc0aec055c10ae07de89ae93ea1a76f1b4d8a6d09b59c280d82c27'
p = json.loads(probe)
assert p['schema'] == 'logh3.battle-preparation-lifecycle-oracle.v1'
assert p['seed'] == '0x12345678' and p['tick'] == 0
assert p['program_sha256'] == '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918'
assert p['memory_sha256'] == 'f70507aaa771a0b97d425037d06696c90385238c6e294a39a103a836960a6808'
assert len(p['rows']) == 6
probe_rows = []
for row in p['rows']:
    if 'source_after_hex' in row:
        assert row['source_after_hex'] == 'ec008b66'
        assert row['rng_after_hex'] == '78563412'
    else:
        expected = ([[70, 50], [50, 70]] if row['name'] == 'retained-and-reused-pool'
                    else [[50, 70]] if row['name'].endswith('sixteen') else [[70, 50]])
        assert [x['target_ships'] for x in row['phases']] == expected
        assert all(x['rng_after_hex'] == '78563412' for x in row['phases'])
    probe_rows.append({key: value for key, value in row.items() if key != 'injections'})

index_path = Path('docs/internals/data/battle-preparation-lifecycle-semantics.json')
index_raw = index_path.read_bytes()
index = json.loads(index_raw)
assert len(index['entries']) == 15
ida = []
for seg, expected_entries in [('1930', 6), ('1AAB', 9)]:
    path = Path('workplace/ida')/f're79-life-reviewed-{seg}.json'
    raw = path.read_bytes()
    d = json.loads(raw)
    assert d['tool'] == 'IDA Pro 9.4'
    assert d['input_sha256'] == index['inputs'][seg]
    assert d['semantic_index']['sha256'] == digest(index_raw)
    records = [f for f in d['functions']]
    records += [i for f in d['functions'] for i in f['instructions']]
    reviewed = [r for r in records if r['grade'] == 'proven']
    assert len(reviewed) == expected_entries
    assert all(f['instructions'][-1]['original_bytes_hex'] == 'cb' for f in d['functions'])
    assert not any(f['original_location'] == '1AAB:024D' for f in d['functions'])
    ida.append({'segment': seg, 'input_sha256': d['input_sha256'],
                'export_sha256': digest(raw), 'reviewed_entries': len(reviewed),
                'function_locations': [f['original_location'] for f in d['functions']]})

batches = []
for mode in ['supply-panel', 'supply-end', 'repair-panel', 'repair-end', 'morale-panel', 'morale-end']:
    a, b = root/f'battle79-life-{mode}-a', root/f'battle79-life-{mode}-b'
    metadata = [(path/'meta.txt').read_text() for path in (a,b)]
    assert all('dosgolem HEAD: f7523f9（未提交檔案 0 個）' in m for m in metadata)
    inputs = equal(a/'mem-in.bin', b/'mem-in.bin')
    outputs = equal(a/'mem-out.bin', b/'mem-out.bin')
    assert len(inputs) == len(outputs) == 1048576
    assert inputs[0x5E00C:0x5E010] == outputs[0x5E00C:0x5E010] == bytes.fromhex('78563412')
    screen = equal(a/'screen.bin', b/'screen.bin')
    palette = equal(a/'pal.bin', b/'pal.bin')
    assert len(screen) == 640*480
    row = {'mode': mode, 'scope': '原版注入診斷局面，沿既有正常控制流；不是remake同狀態對拍。',
           'memory_in_sha256': digest(inputs), 'memory_out_sha256': digest(outputs),
           'screen_sha256': digest(screen), 'palette_sha256': digest(palette),
           'rng_before_hex': inputs[0x5E00C:0x5E010].hex(), 'rng_after_hex': outputs[0x5E00C:0x5E010].hex(),
           'fleet12_flags_before': inputs[0x6115C], 'fleet12_flags_after': outputs[0x6115C],
           'fleet12_morale_after': outputs[0x61198],
           'supply_source_after_hex': outputs[0x5EF5E:0x5EF62].hex(),
           'repair_source_after_hex': outputs[0x5EF70:0x5EF74].hex(),
           'units_after': [ship(outputs, i) for i in (80,88)],
           'receipts': [{'batch': path.name, 'flags_sha256': digest((path/'flags.txt').read_bytes()),
                         'report_sha256': digest((path/'report.txt').read_bytes()),
                         'png_sha256': digest((path/'screen.png').read_bytes())} for path in (a,b)]}
    if mode.startswith('supply'):
        assert row['units_after'][0]['missiles'] == row['units_after'][0]['fighters'] == 2000
        assert row['units_after'][1]['missiles'] == row['units_after'][1]['fighters'] == 0
    if mode.startswith('repair'):
        assert row['units_after'][0]['ships'] == 70 and row['units_after'][0]['repair_cap'] == 75
        assert row['units_after'][1]['missiles'] == 0
    if mode == 'morale-panel':
        assert row['fleet12_flags_before'] & 1 and row['fleet12_flags_after'] & 4
        assert row['fleet12_morale_after'] == 100
    if mode == 'morale-end':
        assert row['fleet12_flags_before'] & 1 and row['fleet12_flags_after'] & 5 == 0
    batches.append(row)

boot = equal(root/'battle79-source-boot-a/mem.bin', root/'battle79-source-boot-b/mem.bin')
assert len(boot) == 1048576
assert boot[0x5EF5E:0x5EF62] == boot[0x5EF70:0x5EF74] == bytes(4)
assert boot[1] == 0
result = {'schema': 'logh3.battle-preparation-lifecycle.v1', 'golem_head': args.golem_head,
          'golem_dirty': False, 'program_sha256': p['program_sha256'],
          'seed': p['seed'], 'probe_sha256': digest(probe), 'probe_rows': probe_rows,
          'semantic_index_sha256': digest(index_raw), 'ida_exports': ida,
          'boot': {'memory_sha256': digest(boot), 'source_pointer_hex': '00000000',
                   'physical_one': 0, 'scope': '固定執行器主選單狀態；不外推所有真實DOS的中斷向量內容。'},
          'batches': batches}
backlinks = []
evidence = Path('docs/re/369-battle-preparation-lifecycle.md').read_text()
for old in ['docs/re/367-battle-preparation.md', 'docs/re/368-battle-preparation-sort.md',
            'docs/playtest/040-battle-preparation-sort.md']:
    text = Path(old).read_text()
    assert '2026-10-02／研究369續證' in text and '369-battle-preparation-lifecycle.md' in text
    for original in ['1930:05EC', '1AAB:0583', '1AAB:0190']:
        assert original in evidence
    backlinks.append({'old_document': old, 'required_marker': '2026-10-02／研究369續證',
                      'program_sha256': p['program_sha256'],
                      'original_locations': ['1930:05EC', '1AAB:0583', '1AAB:0190'],
                      'scope': '只補已解來源／結果／等待；完整第一階段仍未完成。'})
result['resolution_backlinks'] = backlinks
data = (json.dumps(result, ensure_ascii=False, indent=2)+'\n').encode()
with open(args.output, 'xb') as f:
    f.write(data)
print('六案／六組雙跑與十五項分級核對通過；SHA-256', digest(data))
