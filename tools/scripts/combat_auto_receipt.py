"""第75輪自動命令原版證據；雙跑、bytes及完整來源驗證後凍結。"""
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

def read(p):
    p = Path(p)
    assert p.is_file() and p.stat().st_uid == os.getuid(), p
    b = p.read_bytes()
    files[str(p)] = {'sha256': hashlib.sha256(b).hexdigest(), 'size': len(b),
                     'uid': p.stat().st_uid, 'gid': p.stat().st_gid}
    return b

def observations(b):
    out = {}
    key = None
    for line in b.decode().splitlines():
        if '執行時的暫存器（前' in line:
            key = line.split()[0]
            out[key] = []
        elif key and re.match(r'  #\d+ ', line):
            out[key].append({'step': int(line.split()[0][1:]), 'original_line': line.strip()})
        elif key and line.startswith('      [bp+]'):
            out[key][-1]['bp_words'] = line.strip()
    return out

w = lambda m, a: struct.unpack_from('<H', m, a)[0]
cases = []
for p in sorted(Path('workplace/oracle').glob('command79-r75-*-a')):
    name = p.name.removeprefix('command79-r75-').removesuffix('-a')
    if name == 'trace':
        continue
    b = p.with_name(p.name[:-1]+'b')
    identical = []
    for f in ['input-mem.bin', 'end-mem.bin', 'screen.bin', 'pal.bin', 'screen.png']:
        assert read(p/f) == read(b/f), (name, f)
        identical.append(f)
    runs = []
    for run in [p, b]:
        plan = json.loads(read(Path('workplace')/(run.name+'-plan.json')))
        meta = read(run/'meta.txt').decode()
        assert 'f7523f9' in meta and '未提交檔案 0 個' in meta
        flags = read(run/'flags.txt').decode()
        argv = flags.splitlines()
        parent = Path('workplace')/argv[argv.index('-load-state')+1].removeprefix('/out/')
        assert hashlib.sha256(read(parent)).hexdigest() == plan['parent_state_sha256']
        obs = observations(read(run/'report.txt'))
        assert len(obs['1F6F:005F']) == 1
        runs.append({'plan': plan, 'flags': flags, 'metadata': meta, 'observations': obs})
    assert runs[0]['observations'] == runs[1]['observations'], name
    m, end = read(p/'input-mem.bin'), read(p/'end-mem.bin')
    assert len(m) == len(end) == 1048576
    own = 0x668b2+80*78
    n = len(runs[0]['observations']['1F6F:00BB'])
    assert n <= 3 and len(runs[0]['observations']['3CD5:000C']) == int(n != 0)
    # 無候選的自動分支立即返回；外層已計一筆手動入口，所以仍可能進0152。
    # 不把開場呼叫當成取樣失敗；本批以RND次數及資料不變證明尚未排序／傷害。
    assert n == 0 or len(runs[0]['observations']['1DD6:0152']) == 0
    for pool in [80, 102, 103, 104, 105]:
        a = 0x668b2+pool*78
        assert m[a+8:a+22] == end[a+8:a+22], (name, pool, '艘數／庫存不應變動')
    assert struct.unpack_from('<I', m, 0x5e00c)[0] == int(runs[0]['plan']['seed_before'], 16)
    candidates = []
    for i in range(n):
        pool = w(end, 0x54b30+0xa49a+i*2)
        a = 0x668b2+pool*78
        candidates.append({'pool': pool, 'mask': end[a+0x3c],
                           'ranking_score': w(end, a+0x46),
                           'weapon_scores': [w(end, a+o) for o in [0x48, 0x4a, 0x4c]]})
    cases.append({'case': name, 'grade': 'proven',
                  'classification': 'controlled-original-repeatability; not remake player parity',
                  'identical_files': identical, 'different_pixels': 0,
                  'original_result': {'target_pool': w(end, own+0x2c), 'weapon': end[own+0x2e],
                                      'own_anticipated': w(end, own+0x3e),
                                      'seed': f'{struct.unpack_from("<I", end, 0x5e00c)[0]:08x}',
                                      'sorted_candidates': candidates}, 'runs': runs})
assert len(cases) == 31, len(cases)

source = read('workplace/oracle/normal79-r61-a/end-mem.bin')
assert hashlib.sha256(source).hexdigest() == '6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4'
annotations = {
    '1F6F:004D': '每次艦隊命令入口將物件SHIP+08拷到+3E，重設預估剩餘量',
    '1F6F:006A': '自動先清攻方武器FF；沒有候選時改防禦3',
    '1F6F:0073': '重新查詢候選，不沿用手動入口的mask',
    '1F6F:00B6': '逐候選計算預估量；061F不還原攻方武器byte',
    '1F6F:00D6': '以01D1比較unsigned SHIP+46降序排列候選',
    '1F6F:0103': '提交所選目標原池word+2C',
    '1F6F:011C': '最佳量初始化FFFF；沒有beam候選時其他word不可能嚴格超過它',
    '1F6F:012D': 'beam可用才重新寫武器0並取+48',
    '1F6F:0149': 'missile預估量嚴格大於目前最佳才更新武器1',
    '1F6F:016E': 'fighter預估量嚴格大於目前最佳才更新武器2',
    '1F6F:018B': '從攻方自己的+3E減最佳量，word回繞；不是扣目標+3E',
    '1F6F:01AF': '自動標示後等待六個已知timer tick',
    '1F6F:01C5': '無候選寫武器3；不寫目標word也不取RND',
    '1F6F:0219': '候選排序unsigned score比較，大者回FFFF，小者回1，相等回0',
    '1F6F:026F': '有候選恰取一次RND',
    '1F6F:027C': 'RND乘FLEET+45除100，截成低word',
    '1F6F:029C': '值小於8192且候選至少三筆，取排序第三筆',
    '1F6F:02AF': '否則值小於16384且至少兩筆，取排序第二筆；其餘第一筆',
    '1F6F:062F': '先清目標三個預估word+48／+4A／+4C',
    '1F6F:064D': '逐可用武器暫寫攻方+2E後呼叫0714，不還原',
    '1F6F:06C4': 'SHIP+46取三個unsigned預估量的最大值',
    '1F6F:06FC': '控制查詢AL為1才加CPU目標種類權重；不把Faction1等同CPU',
    '1F6F:0849': '射擊量乘attack取32位低積',
    '1F6F:0855': '預估百分比使用unsigned32除defence，再除750',
    '1F6F:0871': 'OR已清CF，JNB必跳；此分支不能當signed負數夾零',
    '1F6F:088D': '目標武器3將百分比除二；不扣彈藥、不取傷害RND',
    '1F6F:0899': '傷害預估讀目標+3E，不讀旗艦ADML+12修正',
    '1F6F:08EA': 'CPU目標旗艦score乘3後保留低word',
    '1F6F:08FA': 'CPU目標非旗艦kind6 score乘2，保留低word',
    '1F6F:0915': 'CPU目標非旗艦kind5 score乘3取低word後再unsigned除2',
}
exports = []
for label in ['auto', 'evaluate']:
    p = Path('workplace/ida')/('command79-r75-'+label+'.json')
    data = json.loads(read(p))
    raw = read(p.with_suffix('.com'))
    assert data['kernel_version'] == '9.4' and data['uid'] == os.getuid()
    seg = int(data['runtime_segment'], 16)*16
    assert hashlib.sha256(raw).hexdigest() == data['input_sha256']
    assert raw[256:] == source[seg:seg+len(raw)-256]
    for r in data['instructions']:
        offset = int(r['original_location'].split(':')[1], 16)
        original = bytes.fromhex(r['original_bytes_hex'])
        assert source[seg+offset:seg+offset+len(original)] == original
        if r['original_location'] in annotations:
            r.update(semantic=annotations.pop(r['original_location']), grade='proven',
                     source='本收據固定RAM bytes、direct caller及31案雙跑，限列出數值域')
            r.pop('warning', None)
    exports.append(data)
assert not annotations, annotations
ancestry = []
for name in ['combat79-r62-gate-near-human-a', 'combat79-r62-command-a', 'combat79-r63-target-a']:
    ancestry.append(json.loads(read(Path('workplace')/(name+'-plan.json'))))
data = {'schema': 'logh3.fleet-auto-command.v1',
        'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
        'remake_base_commit': '529b61a5deb64e7affbb8b66f43b70b069d1e7a2',
        'gin3_sha256': '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918',
        'bundle_sha256': 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
        'cases': cases, 'ida_exports': exports, 'ancestry': ancestry,
        'limitations': ['受控原版自動，不代替remake正常命令／演出／保存',
                        '保留原版缺beam時沿最後預估武器，以及攻方+3E低word回繞',
                        '061被動CPU政策保持；原版CPU診斷不是啟用正式攻擊的授權',
                        '零距離沿已核准首點修正，本次原版案例沒有零距離'],
        'files': files}
target.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print(len(cases), '組雙跑，', len(files), '份來源；SHA-256', hashlib.sha256(target.read_bytes()).hexdigest())
