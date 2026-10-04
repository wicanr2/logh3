"""第77輪原版命令顯示收據及格式；只從固定RAM取資料。"""
import hashlib
import json
import os
from pathlib import Path
import sys

root = Path('/src') if Path('/src/internal/game').is_dir() else Path('/work')
files = {}
def record(path):
    p = root/path
    b = p.read_bytes()
    assert p.stat().st_uid == os.getuid(), str(p)
    files[str(path)] = {'sha256': hashlib.sha256(b).hexdigest(), 'size': len(b)}
    return b

def write(path, data):
    p = root/path
    assert not p.exists() and p.parent.stat().st_uid == os.getuid(), str(p)
    p.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')
    print(path, hashlib.sha256(p.read_bytes()).hexdigest(), '來源', len(files))

ram = record('workplace/oracle/normal79-r61-a/end-mem.bin')
assert files['workplace/oracle/normal79-r61-a/end-mem.bin']['sha256'] == '6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4'
ds = 0x54b30
offsets = {'enemy_unknown_fleet': 0x4534, 'enemy_unknown_name': 0x4541,
           'enemy_known_ships': 0x4528, 'enemy_unknown_ships': 0x455e,
           'enemy_distance': 0x456f}
formats = {k: ram[ds+o:ds+o+100].split(b'\0',1)[0].hex() for k,o in offsets.items()}
assert formats['enemy_distance'] == 'b65ac2f7202025366c7530303030'
if sys.argv[1] == '--spec':
    write('docs/internals/data/fleet-command-display-spec.json',
          {'schema': 'logh3.fleet-command-display-spec.v1',
           'source_ram_sha256': files['workplace/oracle/normal79-r61-a/end-mem.bin']['sha256'],
           'formats_hex': formats})
elif sys.argv[1] == '--native':
    cases = []
    for case in ['wait', 'hover1', 'hover2', 'hover3', 'popup1', 'popup2', 'popup3']:
        for name in ['input-mem.bin','end-mem.bin','screen.bin','screen.png','pal.bin']:
            a,b = [f'workplace/oracle/combat79-r77-{case}-{s}/{name}' for s in 'ab']
            assert record(a) == record(b), case+' '+name
        plans = []
        for s in 'ab':
            prefix = f'workplace/oracle/combat79-r77-{case}-{s}/'
            assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in record(prefix+'meta.txt').decode()
            record(prefix+'flags.txt')
            record(prefix+'report.txt')
            record(prefix+'end.state')
            plans.append(json.loads(record(f'workplace/combat79-r77-{case}-{s}-plan.json')))
        assert plans[0]['injections'] == plans[1]['injections']
        end = (root/f'workplace/oracle/combat79-r77-{case}-a/end-mem.bin').read_bytes()
        assert int.from_bytes(end[0x5e00c:0x5e010], 'little') == 0x12345678
        cases.append({'case': case, 'plan': plans[0], 'two_runs_equal': True,
                      'detail': end[0x668b2+102*78+0x17],
                      'seed_after': '12345678'})
    ida = []
    for seg in ['165F','165F-shapes','1ED2','1629']:
        p = f'workplace/ida/combat79-r77-{seg}.json'
        data = json.loads(record(p))
        binary = f'workplace/ida/combat79-r77-{seg}.com'
        raw = record(binary)
        assert hashlib.sha256(raw).hexdigest() == data['input_sha256']
        assert data['kernel_version'] == '9.4' and data['uid'] == os.getuid()
        for i in data['instructions']:
            a = int(i['runtime_linear_hex'], 16)
            b = bytes.fromhex(i['original_bytes_hex'])
            assert ram[a:a+len(b)] == b
        ida.append({'path':p, 'input':binary, 'instructions':len(data['instructions']),
                    'address_space':data['address_space']})
    record('docs/internals/data/fleet-command-display-spec.json')
    state_check = json.loads(record('workplace/round77-state-comparison.json'))
    assert len(state_check) == 7 and all(r['all_machine_fields_equal'] and r['dos_section_bytes_equal'] for r in state_check)
    record('tools/scripts/fleet_command_state_check.go')
    for p in ['tools/fleet_command_display_oracle.sh','tools/scripts/fleet_command_display_plan.py',
              'tools/scripts/fleet_command_display_ida.py','tools/scripts/fleet_command_display_receipt.py',
              'docs/internals/data/fleet-combat-command.json',
              'workplace/combat79-r62-command-a-plan.json',
              'workplace/combat79-r62-gate-near-human-a-plan.json',
              'workplace/oracle/combat79-r62-command-a/command.state']:
        record(p)
    write('docs/internals/data/fleet-command-display-native.json',
          {'schema':'logh3.fleet-command-display-native.v1',
           'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
           'remake_parent':'e30f3869b58db2d484f6109931e5181e784026b7',
           'classification':'controlled original reproducibility; not normal player path',
           'cases':cases, 'state_comparison':state_check,
           'state_encoding_limit':'gzip/gob map順序不同；完整機器欄位相同，DOS段逐byte相同',
           'ida_exports':ida, 'formats_hex':formats, 'files':files})
else:
    raise ValueError(sys.argv[1])
