"""第64輪戰損診斷收據；保存原始定位與所有注入，拒絕覆寫。"""
import hashlib
import json
import os
from pathlib import Path
import re
import struct
import sys

target = Path(sys.argv[1])
assert not target.exists() and target.parent.stat().st_uid == os.getuid()
files = {}


def read(path):
    path = Path(path)
    assert path.is_file() and path.stat().st_uid == os.getuid(), path
    data = path.read_bytes()
    files[str(path)] = {'sha256': hashlib.sha256(data).hexdigest(), 'size': len(data),
                        'uid': path.stat().st_uid, 'gid': path.stat().st_gid}
    return data


def calls(report):
    headers = list(re.finditer(r'([0-9A-F]{4}:[0-9A-F]{4}) 執行時的暫存器（前 (\d+) 次）：', report))
    out = {}
    for i, h in enumerate(headers):
        rows = []
        block = report[h.end():headers[i+1].start() if i+1 < len(headers) else len(report)]
        for line in block.splitlines():
            if re.match(r'  #\d+ ', line):
                rows.append({'step': int(line.split()[0][1:]), 'raw_line': line.strip(),
                             'registers': dict(re.findall(r'(?<![:A-Z])([A-Z]{2})=([0-9A-F]{4})', line))})
            elif line.startswith('      [bp+]') and rows:
                rows[-1]['bp_words_hex'] = line.strip()
        assert len(rows) == int(h[2]), h[1]
        out[h[1]] = {'count': len(rows), 'capped': len(rows) == 64, 'rows': rows}
    return out


def records(mem):
    out = {}
    for name, pool in (('own', 80), ('target', 102)):
        address = 0x668b2 + pool*78
        row = mem[address:address+78]
        fleet = 0x54b30+0xc074+row[1]*122
        admiral = 0x54b30+0xb77c+mem[fleet+4]*28
        out[name] = {'pool': pool, 'ship_raw_hex': row.hex(),
                     'fleet_raw_hex': mem[fleet:fleet+122].hex(),
                     'admiral_raw_hex': mem[admiral:admiral+28].hex()}
    out['seed_hex'] = f'{struct.unpack_from("<I", mem, 0x5e00c)[0]:08x}'
    out['scores_raw_hex'] = mem[0x54b30+0x8fe4:0x54b30+0x8ff4].hex()
    return out


results = json.loads(read('workplace/round64-damage-probe.json'))
assert len(results) == 31
cases = []
for result in results:
    name = result['case'][:-2]
    a, b = (Path('workplace/oracle')/(name+'-'+side) for side in 'ab')
    same = []
    for filename in ('input-mem.bin', 'end-mem.bin', 'screen.bin', 'pal.bin', 'screen.png'):
        assert read(a/filename) == read(b/filename), (name, filename)
        same.append(filename)
    runs = []
    for side, run in (('a', a), ('b', b)):
        plan = json.loads(read(Path('workplace')/(name+'-'+side+'-plan.json')))
        flags = read(run/'flags.txt').decode()
        meta = read(run/'meta.txt').decode()
        assert 'f7523f9' in meta and '未提交檔案 0 個' in meta
        argv = flags.splitlines()
        parent = Path('workplace')/argv[argv.index('-load-state')+1].removeprefix('/out/')
        assert hashlib.sha256(read(parent)).hexdigest() == plan['parent_state_sha256']
        mem = read(run/'input-mem.bin')
        assert len(mem) == 1048576
        for injection in plan['injections']:
            assert injection['address'].startswith('lin:')
            address = int(injection['address'].split(':')[1], 16)
            value = bytes.fromhex(injection['value_hex'])
            assert mem[address:address+len(value)] == value, injection
            assert f"#{injection['step']} 改記憶體 lin:{address:X}（{address:X}）{len(value)} bytes" in read(run/'report.txt').decode()
        runs.append({'plan': plan, 'flags': flags, 'metadata': meta,
                     'calls': calls(read(run/'report.txt').decode())})
    ka, kb = (set(run['calls']) for run in runs)
    common = ka & kb
    assert all(runs[0]['calls'][loc] == runs[1]['calls'][loc] for loc in common), name
    if ka != kb:
        assert name.endswith(('laser-observe', 'missile-observe'))
        assert not (ka-kb) and kb-ka == {'203A:012A', '203A:01C4', '203A:01E9', '203A:020B', '203A:02C2', '203A:02D4'}
    observation_comparison = {'common_points_identical': sorted(common),
                              'only_a': sorted(ka-kb), 'only_b': sorted(kb-ka),
                              'reason': '早期observe-b增加六處唯讀觀測；兩份完整flags及calls保持。interpolate兩跑用相同觀測點。'}
    end = read(a/'end-mem.bin')
    assert len(end) == 1048576 and len(read(a/'screen.bin')) == 640*480
    cases.append({'case': name.removeprefix('damage79-r64-'), 'grade': 'proven',
                  'classification': 'controlled-original-damage-diagnosis; no production parity',
                  'comparison': {'identical_files': same, 'pixels': 307200, 'different_pixels': 0,
                                 'exclusions': [], 'serialized_state_identical': read(a/'damage.state') == read(b/'damage.state'),
                                 'scope': 'RAM、framebuffer及calls重現；序列化state差異未裁決，不作整機state相同聲明'},
                  'input_records': records(read(a/'input-mem.bin')), 'end_records': records(end),
                  'observation_comparison': observation_comparison,
                  'probe_projection': result, 'runs': runs})

source = read('workplace/oracle/normal79-r61-a/end-mem.bin')
assert hashlib.sha256(source).hexdigest() == '6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4'
annotations = {
    '1DD6:0C68': '先求攻擊係數203A:0002',
    '1DD6:0C7E': '再求目標方向防禦203A:0214',
    '1DD6:0D67': '飛彈消耗min(發射座乘艘數,庫存)，低word接sbb',
    '1DD6:0DCD': '戰鬥機扣min(每艘架數乘艘數,庫存)的20%，低word接sbb',
    '1DD6:0DD7': '艦種3在庫存消耗後把damage用架數減半',
    '1DD6:0E0F': '攻擊乘射擊量除防禦後再除750，取低word百分比',
    '1DD6:0E20': '百分比無號上限100；前方or/jnb不能當負數夾零證據',
    '1DD6:0E2E': '目標武器3時百分比減半',
    '1DD6:0E3B': '只有目標分隊序0才走旗艦百分比減傷',
    '1DD6:0E4B': '提督+12大於100才按100/屬性縮減百分比',
    '1DD6:0E73': '屬性不大於100且百分比至少80時查目標陣營操作模式',
    '1DD6:0E7E': '操作模式返回1時百分比再減10',
    '1DD6:0E88': '目標少於100艘走亂數差額除二分支',
    '1DD6:0EEF': '負數除二先減signword再sar，往零截斷',
    '1DD6:0F1D': '有號16位元requested負值夾零',
    '1DD6:0F2B': 'requested為0時再取亂數mod10',
    '1DD6:0F59': '委派0593處理實扣、庫存、修理上限、分數及功績',
    '1DD6:0F6D': '損失數字顯示owner；不是計分函式，功能邊界到0FC1',
    '203A:01C4': '保留AX:BX定點比例，供後續signed斜率乘積高word截斷',
    '203A:01E9': '已插值曲線值，再乘武器power除100',
    '203A:020B': '武器power返回前再由0002乘攻方相應艦隊能力除100',
    '203A:02C2': '以目標看向攻擊者的方位角',
    '203A:02D4': '與目標FLEET+4A朝向取正差，不折回0到128',
    '203A:03A7': '目標裝甲橢圓半徑取高word，再乘目標對應武器能力除100',
    '203A:06F5': '實扣除20，依對手陣營加入DS8FE4及8FEC兩個u32分數',
    '203A:0729': '實扣除二，非空攻方艦隊指標委派0743計功',
    '203A:076D': '攻方提督ADML+02低word加功績，接高wordcarry',
    '203A:0797': '副官若非FF則再除二，低word加功績，接高wordcarry',
    '3982:0233': '陣營2返回FF，其餘依DS8F25選座位',
    '3982:024A': '返回DS8F21座位操作模式；本次0/1兩分支已核',
}
exports = []
for label in ('203A', '1DD6', '3982', 'merit-full'):
    path = Path('workplace/ida')/('battle79-r64-'+label+'.json')
    data = json.loads(read(path))
    assert data['kernel_version'] == '9.4' and data['uid'] == os.getuid()
    segment = int(data['runtime_segment'], 16)
    raw = read(path.with_suffix('.com'))
    assert hashlib.sha256(raw).hexdigest() == data['input_sha256']
    assert raw[256:] == source[segment*16:segment*16+len(raw)-256]
    for row in data['instructions']:
        offset = int(row['original_location'].split(':')[1], 16)
        original = bytes.fromhex(row['original_bytes_hex'])
        assert source[segment*16+offset:segment*16+offset+len(original)] == original
        if row['original_location'] in annotations:
            row.update(semantic=annotations[row['original_location']], grade='proven',
                       source='固定RAM bytes、direct caller及第64輪原版受控返回／欄位觀測，範圍限此契約')
            row.pop('warning', None)
    exports.append(data)

ancestry = []
for name in ('combat79-r62-gate-near-human-a', 'combat79-r62-command-a',
             'combat79-r63-target-a', 'combat79-r63-target-mapped-a',
             'combat79-r63-laser-submit-a', 'combat79-r63-missile-submit-a'):
    ancestry.append(json.loads(read(Path('workplace')/(name+'-plan.json'))))
for path in ('tools/fleet_combat_damage_oracle.sh', 'tools/scripts/fleet_combat_damage_plan.py',
             'tools/scripts/fleet_combat_damage_probe.go', 'tools/scripts/fleet_combat_damage_ida.py',
             'tools/scripts/fleet_combat_damage_receipt.py', 'internal/rules/fixed.go',
             'internal/rules/fortressgun.go', 'internal/rules/rand.go',
             'internal/rules/attackcoef.go', 'workplace/round64-standard-tests.log'):
    read(Path('/ref')/path)
tables = []
for weapon_type in (0, 5, 10):
    stat = 0x54b30+0x4fda+weapon_type*6
    ptr = 0x54b30+0x477c+weapon_type*4
    offset, segment = struct.unpack_from('<HH', source, ptr)
    entries = []
    for i in range(16):
        row = source[segment*16+offset+i*6:segment*16+offset+(i+1)*6]
        x, y = struct.unpack('<IH', row)
        entries.append({'x_fixed16_16_hex': f'{x:08x}', 'y_word': y, 'raw_hex': row.hex()})
        if x == 0x7fffffff:
            break
    else:
        raise AssertionError('曲線上限未取得')
    tables.append({'weapon_type': weapon_type, 'grade': 'proven',
                   'source': 'normal79-r61-a/end-mem.bin；原版DS表與0002/00EC直接讀取',
                   'stat_ds_offset_hex': f'{0x4fda+weapon_type*6:04x}',
                   'stat_raw_hex': source[stat:stat+6].hex(),
                   'curve_pointer_ds_offset_hex': f'{0x477c+weapon_type*4:04x}',
                   'curve_pointer_raw_hex': source[ptr:ptr+4].hex(), 'entries': entries})
data = {'schema': 'logh3.fleet-combat-damage.v1', 'status': '092 DRAFT',
        'remake_base_commit': '5c2049510890f50b52986976a520fb6f5319499d', 'production_changes': [],
        'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
        'bundle_sha256': 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
        'gin3_sha256': '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918',
        'image_identities': read('workplace/round63-images.txt').decode(),
        'ancestry': ancestry, 'cases': cases, 'weapon_tables': tables, 'ida_exports': exports,
        'limitations': ['原版双跑及可丟棄計算probe不等於正式Game、UI或保存對拍',
                        '戰鬥機為傷害入口注入；正式可用2選單、候選條件與多攻擊者仍缺',
                        'distance0曲線前項、其他武器表及極端整數溢位不外推',
                        '完整0152/026B演出、正常玩家路徑與保存仍缺；092 DRAFT',
                        'zero-fallback在排序前注入，明列16次前段乱數；fallback-fixed在傷害入口固定seed',
                        'flag-high101為早期误讀名稱；實際屬性104，舊收據保留，現行集合改用flag-high104',
                        '序列化state差異未裁決；RAM、framebuffer及calls一致',
                        '061單人CPU被動與#89待裁決差異保持'], 'files': files}
with target.open('x', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
print(f'{len(cases)}組雙跑，{len(files)}份來源，{sum(len(x["instructions"]) for x in exports)}條原始IDA指令')
print('SHA-256', hashlib.sha256(target.read_bytes()).hexdigest())
