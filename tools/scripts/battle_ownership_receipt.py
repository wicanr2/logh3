"""#79 子步8-2：審查双跑、完整記錄、報告等待與原指令，凍結typed證據。"""
import base64
import hashlib
import json
import os
from pathlib import Path
import re

root = Path('.')
head = 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
sources = {}


def read(path):
    p = root / path
    b = p.read_bytes()
    sources[str(p)] = {'sha256': hashlib.sha256(b).hexdigest(), 'size': len(b)}
    return b


def call_count(text, entry):
    m = re.search(re.escape(entry) + r' 執行時的暫存器（前 (\d+) 次）', text)
    assert m, entry
    return int(m.group(1))


a = read('workplace/ownership79-r82-probe-v2-a.json')
assert a == read('workplace/ownership79-r82-probe-v2-b.json')
probe = json.loads(a)
assert probe['dosgolem_head'] == head and len(probe['rows']) == 156
base_memory = read('workplace/oracle/battle79-move-entry-a/mem-in.bin')
reporters = list(base_memory[0x54B30+0x8F4D:0x54B30+0x8F4F])
assert reporters == [13, 58]
cases = []
reports = 0
for row in probe['rows']:
    owners = list(base64.b64decode(row['planet_owners']))
    counts = [owners.count(i) for i in range(3)]
    own = 0 if counts[1] == counts[2] == 0 else 1 if counts[0] == counts[2] == 0 else 2
    before = bytearray.fromhex(row['before_record_hex'])
    want = before.copy()
    want[18] = own
    want[22:25] = bytes(counts)
    assert want.hex() == row['after_record_hex'], row['name']
    assert row['seed_after_hex'] == '78563412'
    calls = row['calls'] or []
    notify = [x for x in calls if x['entry'] == '185C:03C6']
    shown = [x for x in calls if x['entry'] == '3826:0405']
    seats = [f for f in range(2) if row['human_mask'] & (1 << f)] if row['old_owner'] != own else []
    assert [x['args'][2] for x in shown] == seats
    assert [x['args'][0] & 255 for x in notify] == ([0, 1] if row['old_owner'] != own else [])
    pending = int(row['region'] == 14 and own < 2 and row['old_owner'] != own
                  and bool(row['human_mask'] & (1 << own)) and row['occurred'] == 0)
    assert row['pending_after'] == pending
    assert all(bytes.fromhex(x['region_record_hex'])[18] == row['old_owner'] for x in calls)
    reports += len(shown)
    cases.append({k: row[k] for k in ('name', 'region', 'old_owner', 'human_mask', 'occurred',
                                     'injections', 'before_record_hex', 'after_record_hex',
                                     'pending_after', 'seed_after_hex', 'steps')} | {
        'planet_owners': owners, 'new_owner': own, 'counts': counts,
        'reports': [{'seat': x['args'][2], 'text_hex': x['text_hex'],
                     'pending_before_display': x['pending_before_display'],
                     'region_record_hex': x['region_record_hex'],
                     'face': reporters[x['args'][2]]} for x in shown],
    })
assert reports == 104

natural = []
state_compare = json.loads(read('workplace/ownership79-r82-state-compare-v4.json'))
assert len(state_compare['rows']) == 14
assert all(x['decoded_all_fields_equal'] and x['negative_control_passed']
           and x['machine_fields'] == 46 and x['dos_fields'] == 24 for x in state_compare['rows'])
read('workplace/ownership79-r82-state-compare-v4.go')
read('workplace/ownership79-r82-state-schema-v4.txt')
names = ('entry', 'control', 'capture0', 'capture1', 'partial0', 'partial1',
         'capture0-second', 'capture1-second', 'capture0-finish', 'capture1-finish',
         'capture0-complete', 'capture1-complete', 'partial0-complete', 'partial1-complete')
for name in names:
    dirs = [f'workplace/oracle/ownership79-r82-{name}-{side}' for side in 'ab']
    files = ('mem-in.bin', 'entry.state', 'screen.bin', 'pal.bin', 'screen.png') if name == 'entry' else (
        'mem-out.bin', 'end.state', 'screen.bin', 'pal.bin', 'screen.png')
    for file in files:
        left, right = read(dirs[0] + '/' + file), read(dirs[1] + '/' + file)
        if not file.endswith('.state'):
            assert left == right, (name, file)
        else:
            state = next(x for x in state_compare['rows'] if x['name'] == name)
            assert hashlib.sha256(left).hexdigest() == state['raw_sha256_a']
            assert hashlib.sha256(right).hexdigest() == state['raw_sha256_b']
    for d in dirs:
        meta = read(d + '/meta.txt').decode()
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in meta
        read(d + '/report.txt')
        read(d + '/flags.txt')
    flags = read(dirs[0] + '/flags.txt').decode().splitlines()
    report = read(dirs[0] + '/report.txt').decode()
    if name == 'entry':
        raw = read(dirs[0] + '/mem-in.bin')
        assert len(raw) == 1 << 20
        assert raw[0x5E00C:0x5E010] == bytes.fromhex('78563412')
        assert raw[0x5FFAA] == 2
    else:
        raw = read(dirs[0] + '/mem-out.bin')
        count = call_count(report, '3826:0405')
        continued = call_count(report, '185C:0491')
        if name == 'control':
            assert count == 0 and continued == 1 and raw[0x5FFAA] == 2
        elif name.endswith('-finish'):
            assert count == 2 and continued == 1
            assert raw[0x5FFAA] == int(name[7])
            assert call_count(report, '185C:03BD') == 1
        else:
            assert count == (2 if name.endswith('-second') else 1) and continued == 0
            old = int(name[7]) if name.startswith('partial') else 2
            assert raw[0x5FFAA] == old
            assert call_count(report, '185C:03BD') == 0
            if name.endswith('-complete'):
                assert call_count(report, '3958:0215') == 1
                assert call_count(report, '3765:006C') == 1
                assert 'AX=000D' in report
        assert raw[0x5E00C:0x5E010] == bytes.fromhex('78563412')
    natural.append({'name': name, 'dirs': dirs, 'flags': flags,
                    'classification': 'controlled ownership/seat diagnostic via natural call; not full normal-player parity',
                    'comparison': '完整RAM／framebuffer／色盤／PNG雙跑逐byte相同；保存全70欄位解碼相同，原檔gob map順序可不同'})

ida = []
for segment, name in ((0x185C, '185C-v2'), (0x1213, '1213'), (0x3982, '3982'), (0x3826, '3826')):
    data = json.loads(read(f'workplace/ida/ownership79-r82-{name}.json'))
    assert data['kernel_version'] == '9.4' and data['uid'] == os.getuid()
    binary = read(f'workplace/ida/ownership79-r82-{segment:04X}.com')
    assert data['input_sha256'] == hashlib.sha256(binary).hexdigest()
    original = read('workplace/oracle/battle79-move-entry-a/mem-in.bin')
    assert binary[256:] == original[segment * 16:segment * 16 + 4096]
    for insn in data['instructions']:
        location = int(insn['original_location'].split(':')[1], 16)
        b = bytes.fromhex(insn['original_bytes_hex'])
        assert binary[256 + location:256 + location + len(b)] == b
    ida.append(data)

semantics = [
    ('185C:0007', '9ada025c18', '8-1返回後呼叫8-2'),
    ('185C:0308', '26884416', '清三方行星數的最後一格'),
    ('185C:0330', '268a4c11', '讀行星+11歸屬'),
    ('185C:033A', '26fe07', '對應陣營byte行星數加一'),
    ('185C:0340', '83460652', '行星指標前進82bytes'),
    ('185C:0357', 'c6460100', '無同盟與無主行星時新歸屬0'),
    ('185C:036E', 'c6460101', '無帝國與無主行星時新歸屬1'),
    ('185C:0374', 'c6460102', '其餘新歸屬2'),
    ('185C:038A', '7434', '歸屬不变跳過通知與歸屬寫入'),
    ('185C:0398', '9ac6035c18', '先通知陣營0'),
    ('185C:03AC', '9ac6035c18', '再通知陣營1'),
    ('185C:03BD', '268807', '兩次通知返回後才寫星域+12'),
    ('185C:03D2', '9a2d028239', '查陣營映射座位的控制值'),
    ('185C:03DB', '7407', '只有人類控制值0續行'),
    ('185C:040E', '9a7b001312', '人類自己拿下分支登記費沙事件'),
    ('185C:0422', '268a4712', '部分歸屬分支仍讀舊星域+12'),
    ('185C:0483', '9a05042638', '以陣營為第二參數呼叫戰鬥報告，頭像取該陣營情報參謀'),
    ('3826:040C', '8b7630', '讀第二參數陣營／FFFF'),
    ('3826:045E', '83f8ff', '第二參數非FFFF時使用陣營參謀'),
    ('3826:0465', '8a874d8f', '讀DS8F4D陣營情報參謀索引'),
    ('3826:046D', '8a4632', '僅第二參數FFFF時消費第三參數提督低byte'),
    ('3826:0480', '9a15025839', '畫頭像；本批兩陣營分別13與58'),
    ('3826:048B', '9a6c006537', '頭像完成後才進入訊息等待'),
    ('1213:007E', '803e488f00', '費沙事件已發生則不再登記'),
    ('1213:0088', '26807f110e', '星域+11編號必須14'),
    ('1213:008F', 'c6068e9801', '登記同一戰略回合的費沙待辦'),
    ('3982:023C', 'a0258f', '讀陣營對座位映射'),
    ('3982:024A', '8a87218f', '讀所屬座位控制值'),
]
lookup = {x['original_location']: x for data in ida for x in data['instructions']}
for location, signature, meaning in semantics:
    insn = lookup[location]
    assert insn['original_bytes_hex'] == signature
    insn.update({'semantic': meaning, 'grade': 'proven', 'warning': None,
                 'source': '原始bytes／caller／156案常式双跑及自然續行10組；範圍見研究416'})

original = read('workplace/oracle/ownership79-r82-entry-a/mem-in.bin')
formats = {}
for key, off in (('battle_region_taken', 0x3D58), ('battle_region_grip_lost', 0x3D71),
                 ('battle_region_partly_landed', 0x3D92), ('battle_region_lost', 0x3DAF)):
    formats[key] = original[0x54B30+off:0x54B30+off+80].split(b'\0')[0].hex()
for path in ('tools/scripts/battle_ownership_probe.go', 'tools/scripts/battle_ownership_ida.py',
             'tools/battle_ownership_oracle.sh', 'tools/scripts/battle_ownership_receipt.py',
             'tools/scripts/battle_ownership_state_prepare.py'):
    read(path)
legacy = read('docs/internals/data/battle-ownership.json')
assert hashlib.sha256(legacy).hexdigest() == 'c4b36a1280a713f818716ce67b69b25a7c5e65ba4abdfe79667e3aa46e081ec1'
result = {'schema': 'logh3.battle-ownership.v2', 'parent_commit': '58291ecad91f4efbfdfce26a3988707200044e51',
          'dosgolem_head': head, 'program_sha256': probe['program_sha256'],
          'scope': probe['setup'], 'cases': cases, 'report_count': reports, 'reporters': reporters,
          'supersedes': {'path': 'docs/internals/data/battle-ownership.json',
                        'reason': 'v1 face=-1與無頭像語意被完整報告及3826原指令推翻；規則記錄與原始收據保留'},
          'natural_receipts': natural, 'state_comparison': state_compare,
          'ida': ida, 'report_formats': formats,
          'sources': sources}
out = Path('docs/internals/data/battle-ownership-v2.json')
assert out.parent.stat().st_uid == os.getuid()
with out.open('x', encoding='utf-8') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write('\n')
print('156案／104報告／14組完整原版收據／28原指令，來源', len(sources),
      'SHA-256', hashlib.sha256(out.read_bytes()).hexdigest())
