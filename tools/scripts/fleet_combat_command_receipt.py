"""第63輪原版目標／武器／交戰收據；禁止覆寫已凍結資料。"""
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
    files[str(path)] = {"sha256": hashlib.sha256(data).hexdigest(),
                        "size": len(data), "uid": path.stat().st_uid,
                        "gid": path.stat().st_gid}
    return data


def observations(report):
    headers = list(re.finditer(r'([0-9A-F]{4}:[0-9A-F]{4}) 執行時的暫存器（前 (\d+) 次）：', report))
    out = {}
    for i, m in enumerate(headers):
        text = report[m.end():headers[i+1].start() if i+1 < len(headers) else len(report)]
        lines = text.splitlines()
        rows = []
        for line in lines:
            if re.match(r'  #\d+ ', line):
                rows.append({"step": int(line.split()[0][1:]), "original_line": line.strip(),
                             "registers": dict(re.findall(r'(?<![:A-Z])([A-Z]{2})=([0-9A-F]{4})', line))})
            elif line.startswith('      [bp+]') and rows:
                rows[-1]['bp_words_hex'] = line.strip()
        assert len(rows) == int(m[2]), m[1]
        out[m[1]] = {"observed_count": len(rows), "cap": 64,
                     "count_is_capped": len(rows) == 64, "rows": rows}
    return out


def record(mem, pool):
    data = mem[0x668b2+pool*78:0x668b2+(pool+1)*78]
    return {"pool": pool, "raw_hex": data.hex(),
            "ships": struct.unpack_from('<H', data, 8)[0],
            "repair_cap": struct.unpack_from('<H', data, 12)[0],
            "missiles": struct.unpack_from('<I', data, 14)[0],
            "fighters": struct.unpack_from('<I', data, 18)[0],
            "target": struct.unpack_from('<H', data, 44)[0],
            "weapon": data[46], "candidate_mask": data[60], "morale": data[6]}


cases = []
kinds = ('target-mapped laser missile disabled cancel defence auto '
         'laser-submit missile-submit disabled-submit cancel-submit defence-submit '
         'laser-phase-end missile-phase-end laser-finish missile-finish').split()
for kind in kinds:
    name = 'combat79-r63-' + kind
    a, b = (Path('workplace/oracle') / (name+'-'+side) for side in 'ab')
    same = []
    for filename in ('end-mem.bin', 'screen.bin', 'pal.bin', 'screen.png'):
        assert read(a/filename) == read(b/filename), (kind, filename)
        same.append(filename)
    mem = read(a/'end-mem.bin')
    assert len(mem) == 1048576 and len(read(a/'screen.bin')) == 640*480
    runs = []
    for side in 'ab':
        run = Path('workplace/oracle')/(name+'-'+side)
        plan = json.loads(read(Path('workplace')/(name+'-'+side+'-plan.json')))
        flags = read(run/'flags.txt').decode()
        meta = read(run/'meta.txt').decode()
        assert 'f7523f9' in meta and '未提交檔案 0 個' in meta
        argv = flags.splitlines()
        parent = Path('workplace')/argv[argv.index('-load-state')+1].removeprefix('/out/')
        assert hashlib.sha256(read(parent)).hexdigest() == plan['parent_state_sha256']
        runs.append({"plan": plan, "flags": flags, "metadata": meta,
                     "calls": observations(read(run/'report.txt').decode())})
    common = runs[0]['calls'].keys() & runs[1]['calls'].keys()
    assert all(runs[0]['calls'][loc] == runs[1]['calls'][loc] for loc in common), kind
    own, enemy = record(mem, 80), record(mem, 102)
    seed = struct.unpack_from('<I', mem, 0x5e00c)[0]
    if kind.endswith('-submit'):
        expected = {'laser-submit': (102, 0), 'missile-submit': (102, 1),
                    'disabled-submit': (65535, 255), 'cancel-submit': (65535, 255),
                    'defence-submit': (65535, 3)}[kind]
        assert (own['target'], own['weapon']) == expected
        assert seed == 0x12345678 and own['ships'] == enemy['ships'] == 3000
        assert own['missiles'] == 12000 and own['fighters'] == 36000
        for loc in ('1DD6:0152','1DD6:030F','3CD5:000C'):
            assert runs[0]['calls'][loc]['observed_count'] == 0
    if kind.endswith(('-phase-end', '-finish')):
        laser = kind.startswith('laser-')
        assert own['ships'] == 3000 and enemy['ships'] == (2274 if laser else 2424)
        assert own['missiles'] == (12000 if laser else 6000)
        assert enemy['missiles'] == (9096 if laser else 9696)
        assert enemy['fighters'] == (20466 if laser else 21816)
        assert seed == 0xe06870c1
        assert runs[0]['calls']['1DD6:026B']['observed_count'] == 1
        assert runs[0]['calls']['181B:0385']['observed_count'] == int(kind.endswith('-finish'))
    states_same = read(a/'selected.state') == read(b/'selected.state')
    cases.append({"case": kind, "grade": "proven",
                  "classification": "controlled-original-repeatability; no remake parity",
                  "comparison": {"identical_files": same, "pixels_compared": 307200,
                                 "different_pixels": 0, "exclusions": [],
                                 "serialized_states_identical": states_same,
                                 "state_scope": "比較RAM與framebuffer；共同固定a父快照，完整序列化state不作完成門檻"},
                  "seed_end_hex": f'{seed:08X}', "own": own, "enemy": enemy,
                  "runs": runs})

source = read('workplace/oracle/normal79-r61-a/end-mem.bin')
assert hashlib.sha256(source).hexdigest() == '6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4'
annotations = {
    '1ED2:0408': '呼叫049D填武器選單旗標與數量',
    '1ED2:0478': '選單選擇返回AL；此處不写SHIP',
    '1ED2:0507': '第一武器不可用時清DS44B6',
    '1ED2:0551': '第二武器不可用時清DS44BC',
    '1ED2:05BA': '第三武器不可用時清DS44C2',
    '1ED2:06AA': '自動選項設定DSA458為1',
    '1ED2:06B6': '防禦選項設定DSA458為0並返回1',
    '1ED2:0781': '候選命中回呼；讀SHIP+34/+36，16乘16半開區間',
    '1DD6:0329': '03A1收所有參戰艦隊的一般分隊池，含未提交者',
    '1DD6:0339': '0421生成優先值後排序池列表',
    '1DD6:037F': '武器FF跳過該分隊',
    '1DD6:0383': '武器3防禦亦跳過攻擊owner',
    '1DD6:038C': '呼叫05D7处理有效攻擊者',
    '1DD6:04CB': '每個收集池優先值生成的第一個亂數',
    '1DD6:04E7': '每個收集池優先值生成的第二個亂數',
    '1DD6:04DE': '寫SHIP+42優先值；非演出位置亂數',
    '1DD6:055C': '比較SHIP+42，較大者優先',
    '1DD6:07AA': '呼叫0C54，持有攻擊與目標遠指標',
    '1DD6:0D67': '飛彈攻擊扣SHIP+0E低word，接sbb高word',
    '1DD6:0DCD': '戰鬥機分支扣SHIP+12低word，接sbb高word',
    '1DD6:0E2E': '目標武器3時將傷害百分比減半',
    '1DD6:0F59': '呼叫203A:0593扣艘數與後續資料更新',
}
exports = []
for label in ('1ED2','1DD6-phase','1DD6-loss','1DD6-damage'):
    path = Path('workplace/ida')/('battle79-r63-'+label+'.json')
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
                       source='固定RAM bytes、direct caller與本收據原版分支觀測；限定已列位移及控制流')
            row.pop('warning', None)
    exports.append(data)

# 共同的祖先注入全文保留，不能只引用會消失的本機檔。
ancestry = []
for name in ('combat79-r62-gate-near-human-a','combat79-r62-command-a',
             'combat79-r63-target-a','combat79-r63-target-mapped-a'):
    ancestry.append(json.loads(read(Path('workplace')/(name+'-plan.json'))))
data = {"schema": "logh3.fleet-combat-command.v1", "status": "092 DRAFT",
        "remake_base_commit": "3873bf01dfc15508c1395cd98074e5d59dd3b21a",
        "production_changes": [], "images": read('workplace/round63-images.txt').decode(),
        "bundle_sha256": "de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673",
        "gin3_sha256": "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918",
        "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c",
        "ancestry": ancestry, "cases": cases, "ida_exports": exports,
        "limitations": ["原版受控入口双跑不是remake same-state驗收",
                         "可用戰鬥機、候選完整條件及多攻擊者順序尚待驗",
                         "203A:0002/0214係數、0C54全部傷害分支與分數仍未形成READY契約",
                         "原版防禦的16次亂數來自030F内0421生成8筆排序鍵；不稱演出亂數",
                         "061單人CPU被動與#89未決忠實度保持"], "files": files}
with target.open('x', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
print(f'{len(cases)}組雙跑、{len(files)}份來源、{sum(len(x["instructions"])for x in exports)}條IDA指令；092 DRAFT')
print('SHA-256', hashlib.sha256(target.read_bytes()).hexdigest())
