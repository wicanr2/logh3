"""第62輪原版艦隊戰鬥入口收據；完整注入、原始定位、雙跑與未知項分開保存。"""
import hashlib
import json
import os
from pathlib import Path
import re
import struct
import sys

target=Path(sys.argv[1])
assert not target.exists(), '收據已凍結，禁止覆寫'
assert target.parent.stat().st_uid==os.getuid()
root=Path('workplace/oracle')
files={}


def digest(data):
    return hashlib.sha256(data).hexdigest()


def read(path):
    path=Path(path)
    assert path.is_file() and path.stat().st_uid==os.getuid(),path
    data=path.read_bytes()
    files[str(path)]={'sha256':digest(data),'size':len(data),'uid':path.stat().st_uid,'gid':path.stat().st_gid}
    return data


def observations(text,cap=128):
    matches=list(re.finditer(r'([0-9A-F]{4}:[0-9A-F]{4}) 執行時的暫存器（前 (\d+) 次）：',text))
    result={}
    for i,m in enumerate(matches):
        block=text[m.end():matches[i+1].start() if i+1<len(matches) else len(text)]
        rows=[]
        for line in block.splitlines():
            if re.match(r'  #\d+ ',line):
                registers=dict(re.findall(r'(?<![:A-Z])([A-Z]{2})=([0-9A-F]{4})',line))
                stack=re.search(r'SS:SP=([0-9A-F]{4}:[0-9A-F]{4})',line)
                assert stack
                registers['SS:SP']=stack[1]
                rows.append({'step':int(line.split()[0][1:]),'registers':registers,'original_line':line.strip()})
            elif line.startswith('      [bp+]') and rows:
                rows[-1]['bp_words_hex']=line.strip()
        assert len(rows)==int(m[2])
        result[m[1]]={'observed_count':len(rows),'cap':cap,'rows':rows}
    return result


def image_pair(p,q,names):
    checks=[]
    for name in names:
        a,b=read(p/name),read(q/name)
        assert a==b,(p,name)
        if name=='screen.bin':
            assert len(a)==640*480
        checks.append(name)
    return {'identical_files':checks,'pixels_compared':640*480,'different_pixels':0,
            'max_channel_difference':0,'difference_bbox':None,'exclusions':[]}


ids=[12,16,17,28,29,30,31]
natural=root/'combat79-r62-natural-a'
mem=read(natural/'entry-mem.bin')
assert len(mem)==1048576 and digest(mem)=='fbe061f125a0554bb6513612c45cba422d26441338e7b431c6b5316473bf6ec1'
general=[]
scouts=[]
for index in ids:
    record=mem[0x60ba4+index*122:0x60ba4+(index+1)*122]
    general.extend(struct.unpack_from('<H',record,0x10+j*4)[0]for j in range(record[15]))
    scouts.append(struct.unpack_from('<H',record,0x38)[0])
assert len(general)==48 and len(scouts)==7
natural_comparison=image_pair(natural,root/'combat79-r62-natural-b',
    ('entry-mem.bin','end-mem.bin','screen.bin','pal.bin','screen.png'))
state_a=read(natural/'entry.state')
state_b=read(root/'combat79-r62-natural-b/entry.state')
natural_comparison['serialized_states_identical']=state_a==state_b
natural_comparison['state_limitation']='兩份序列化快照不同；只使用固定a作父狀態，不宣稱完整state逐byte相同。RAM與framebuffer雙跑相同。'
natural_runs=[]
for side in 'ab':
    run=root/('combat79-r62-natural-'+side)
    natural_runs.append({'actual_flags':read(run/'flags.txt').decode(),
        'metadata':read(run/'meta.txt').decode(),'calls':observations(read(run/'report.txt').decode(),512)})
assert natural_runs[0]['calls']==natural_runs[1]['calls']
assert struct.unpack_from('<I',mem,0x5e00c)[0]==0xafcfd1bc
cases=[]
for kind in 'none action1 action11 action21 actionf1 action0 action4 actionff morale1 morale2 invalid two reverse-faction reverse-number near-outs near-human'.split():
    name='combat79-r62-gate-'+kind
    p,q=root/(name+'-a'),root/(name+'-b')
    comparison=image_pair(p,q,('input-mem.bin','end-mem.bin','screen.bin','pal.bin','screen.png'))
    plans=[]
    for side in 'ab':
        run=root/(name+'-'+side)
        plans.append(json.loads(read(Path('workplace')/(name+'-'+side+'-plan.json'))))
        flags=read(run/'flags.txt').decode()
        metadata=read(run/'meta.txt').decode()
        assert 'f7523f9' in metadata and '未提交檔案 0 個' in metadata
        argv=flags.splitlines()
        assert argv[argv.index('-tick')+1]=='20000'
        report=read(run/'report.txt').decode()
        for injection in plans[-1]['injections']:
            address=int(injection['linear_hex'],16)
            token=f'改記憶體 lin:{address:X}（{address:05X}）'
            assert token.upper() in report.upper(),(kind,token)
        plans[-1]['actual_flags']=flags
        plans[-1]['metadata']=metadata
        plans[-1]['calls']=observations(report)
    a=plans[0]['calls']
    b=plans[1]['calls']
    # 新增觀測旗標不影響舊a收據；共同的呼叫觀測必須一致。
    for loc in a.keys()&b.keys():
        assert a[loc]==b[loc],(kind,loc)
    count=lambda loc:a.get(loc,{}).get('observed_count',0)
    for loc,n in plans[0]['required_counts'].items():
        assert count(loc)==n,(kind,loc,count(loc),n)
    order=[(int(r['registers']['SI'],16)-0xc074)//122 for r in a['1ED2:0001']['rows']]
    assert sorted(order)==sorted(plans[0]['expected_collected_before_sort'])
    for loc in ('1DD6:0152','1DD6:030F','3CD5:000C'):
        assert count(loc)==0,(kind,loc)
    inp=read(p/'input-mem.bin')
    end=read(p/'end-mem.bin')
    assert len(inp)==len(end)==1048576
    assert struct.unpack_from('<I',inp,0x5e00c)[0]==struct.unpack_from('<I',end,0x5e00c)[0]==0x12345678
    records=[]
    for pool in general+scouts:
        off=0x668b2+pool*78
        before,after=inp[off:off+78],end[off:off+78]
        skipped=kind=='invalid' and pool in general[:8]
        expected=0x7d if pool in scouts or skipped else 0xff
        assert after[0x2e]==expected,(kind,pool,after[0x2e],expected)
        records.append({'pool':pool,'type':'scout' if pool in scouts else 'general',
            'input_hex':before.hex(),'end_hex':after.hex(),
            'changed_offsets_hex':[hex(i)for i in range(78)if before[i]!=after[i]]})
    if kind=='near-outs':
        assert a['1ED2:0097']['rows'][0]['registers']['AX']=='0001'
        assert count('181B:0385')==1
    if kind=='near-human':
        assert count('181B:0385')==count('1ED2:0205')==0
        assert a['1ED2:0097']['rows'][0]['registers']['AX']=='0001'
        for side in 'ab':
            read(root/(name+'-'+side)/'prompt.state')
    cases.append({'case':kind,'classification':'controlled-entry-original-repeatability',
        'comparison':comparison,'plans':plans,'observed_fleet_order':order,
        'seed_input_hex':'12345678','seed_end_hex':'12345678','records':records})

p,q=root/'combat79-r62-command-a',root/'combat79-r62-command-b'
command_comparison=image_pair(p,q,('end-mem.bin','screen.bin','pal.bin','screen.png'))
commands=[]
for side in 'ab':
    run=root/('combat79-r62-command-'+side)
    plan=json.loads(read(Path('workplace')/('combat79-r62-command-'+side+'-plan.json')))
    plan['actual_flags']=read(run/'flags.txt').decode()
    plan['metadata']=read(run/'meta.txt').decode()
    plan['calls']=observations(read(run/'report.txt').decode())
    assert plan['calls']['1ED2:0205']['observed_count']==1
    assert plan['calls']['165F:034F']['observed_count']==1
    for loc in ('1ED2:02AB','1ED2:03E9','1DD6:0152','1DD6:030F','181B:0385','3CD5:000C'):
        assert plan['calls'][loc]['observed_count']==0
    read(run/'command.state')
    commands.append(plan)
assert commands[0]['calls']==commands[1]['calls']

annotations={
    '181B:02EA':'FLEET+04提督byte等於FF時跳過；不初始化該艦隊一般分隊',
    '181B:032C':'寫編制內一般分隊SHIP+2E為FF',
    '181B:033D':'讀FLEET+48行動byte',
    '181B:0341':'只取行動低四位',
    '181B:0345':'只有低四位等於1收進參戰清單',
    '181B:0349':'排除FLEET+00 bit0已設的艦隊',
    '181B:0376':'名單非空才交給1DD6:0009',
    '1DD6:0017':'先彙總可進命令流程的一般分隊數；AL為0則返回',
    '1DD6:0025':'選擇結果非零後才開演出',
    '1DD6:0031':'演出後呼叫030F損失流程',
    '1DD6:008D':'以排序後的FLEET指標呼叫1ED2:0001',
    '1DD6:0135':'比較器載入右側FLEET+02到AL',
    '1DD6:0139':'把AX複製到CX；其後AL與CL比較無法形成大於分支',
    '1ED2:00A2':'交戰查詢非零仍另驗一般分隊Out等於7F',
    '1ED2:0218':'人類命令操作先呼叫165F:034F',
    '1ED2:022F':'人類命令操作循環呼叫1629:0005',
    '1ED2:036D':'將選擇返回AL寫入SHIP+2E；完整玩家語意仍待驗',
    '1ED2:037C':'成功選擇後將目標池索引寫入SHIP+2C',
    '1ED2:03AA':'另一分支把SHIP+2E寫成3；玩家語意尚未確認',
    '1F6F:02F8':'敵方候選查詢使用固定300座標單位範圍',
    '1F6F:04BE':'第一武器合格時設候選mask bit0',
    '1F6F:050E':'第二武器合格時設候選mask bit1',
    '1F6F:055E':'第三武器合格時設候選mask bit2',
}
ida=[]
ram=read('workplace/oracle/normal79-r61-a/end-mem.bin')
for name in ('181B','1DD6','1ED2','1ED2-full','1F6F','1F6F-command','1ED2-command'):
    data=json.loads(read('workplace/ida/battle79-r62-'+name+'.json'))
    assert data['kernel_version']=='9.4' and data['uid']==1000
    for row in data['instructions']:
        address=int(row['runtime_linear_hex'],16)
        raw=bytes.fromhex(row['original_bytes_hex'])
        assert ram[address:address+len(raw)]==raw,row['original_location']
        if row['original_location'] in annotations:
            row.update(semantic=annotations[row['original_location']],grade='proven',warning='',
                       source='研究400：固定原版RAM／IDA9.4原指令及受控雙跑；範圍限原始讀寫與呼叫')
    ida.append(data)
for path in ('tools/fleet_combat_entry_oracle.sh','tools/scripts/fleet_combat_entry_plan.py',
             'tools/scripts/fleet_combat_phase_ida.py','tools/scripts/fleet_combat_entry_receipt.py'):
    read(path)
data={'schema':'logh3.fleet-combat-entry.v1','source_commit':'5662aed0892fc2e413befec535a40a580d80451d',
    'source_dirty':True,'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c','dosgolem_dirty':False,
    'bundle_sha256':'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
    'gin3_sha256':'770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918',
    'tool_images':{'golang:1.24-bookworm':'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac',
        'logh3-ebiten:2':'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
        'python:3.12-slim':'sha256:229a2c5bfa27522db7815ea81f9bed70af17ccb9de9fc7ad142b1877b5830d36',
        'ida-pro-9.4-idapython:locked-v1':'sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780'},
    'natural_comparison':natural_comparison,'natural_runs':natural_runs,'cases':cases,'command_comparison':command_comparison,
    'command_runs':commands,'ida_exports':ida,'files':files,
    'limitations':['原版受控入口雙跑，不是原版／remake same-state收據',
        '自然S3無參戰艦隊；接敵局面明示注入，不能替代正常玩家命令路徑',
        '門檻案例末態包含後續執行；初始化語意另綁原始寫入指令',
        'command只到165F／1629操作等待，未選目標／武器，無演出／損失／存讀驗收',
        '排序反例有限，不推定所有排列結果；完整第④步092維持DRAFT',
        '序列化state不同的原因未裁決，不宣稱完整機器快照逐byte相同']}
with target.open('x',encoding='utf-8') as f:
    json.dump(data,f,ensure_ascii=False,indent=2)
    f.write('\n')
print(target,digest(target.read_bytes()),len(files),'檔案',len(cases),'門檻案例',
      sum(len(x['instructions'])for x in ida),'原指令',
      sum(r['grade']=='proven' for x in ida for r in x['instructions']),'proven')
