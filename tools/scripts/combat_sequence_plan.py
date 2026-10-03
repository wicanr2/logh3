"""第66輪固定原版命令序列與係數域，不覆寫前輪收據。"""
import hashlib
import json
import os
from pathlib import Path
import re
import struct
import sys

name = sys.argv[1]
match = re.fullmatch(r'sequence79-r66-([a-z0-9-]+)-([ab])', name)
assert match
case, side = match.groups()
prefix = '/out/oracle/'+name
own, target, second = 0x68112, 0x687c6, 0x668b2+82*78
rows, clicks = [], []
regs = '1ED2:0097,1ED2:0205,1ED2:02AB,1ED2:036D,1ED2:037C,1ED2:03E9,1DD6:0152,1DD6:030F,1DD6:05D7,1DD6:0C54,1DD6:0C6D,1DD6:0C83,1DD6:0E17,1DD6:0F59,1DD6:0F60,1DD6:026B,181B:0385,203A:012A,203A:017E,203A:01C4,203A:01E9,203A:020B,3CD5:000C'

def poke(addr, data, why, step):
    rows.append({'address': f'lin:{addr:x}', 'step': step, 'bytes_hex': bytes(data).hex(' '), 'reason': why})

curve = re.fullmatch(r'curve([0-9]+)-(zero|near|stack-first)', case)
if curve:
    t, where = int(curve[1]), curve[2]
    profiles = {0:(0,0,0),1:(1,0,0),2:(0,2,0),3:(0,4,0),5:(0,0,1),6:(0,2,1),7:(0,4,1),8:(0,3,1),9:(1,4,1),10:(0,0,2),11:(1,0,2)}
    faction, kind, weapon = profiles[t]
    parent = Path('workplace/oracle/combat79-r63-laser-submit-a/selected.state')
    assert hashlib.sha256(parent.read_bytes()).hexdigest() == 'af0019ce8a29ea2e0026a0ba15ac4f37d870f848000247a07250ee10190de19d'
    start, step, limit = 495599999, 506948195, 508000000
    mem = Path('workplace/oracle/normal79-r61-a/end-mem.bin').read_bytes()
    x,y = struct.unpack_from('<HH', mem, own+0x34)
    for addr, data, why in [(0x5e00c,struct.pack('<I',0x12345678),'係數入口固定seed'),
                            (own,bytes([faction]),'實際艦種表陣營'),
                            (own+4,bytes([kind]),'實際可用武器類型的艦種'),
                            (own+0x2e,bytes([weapon]),'此案的光線砲／飛彈／戰鬥機欄位'),
                            (0x6115c+1,bytes([faction]),'攻方艦隊陣營同步'),
                            (0x603fc+1,bytes([faction]),'攻方提督陣營同步'),
                            (target,bytes([1-faction]),'敵方艦艇陣營同步'),
                            (0x61344+1,bytes([1-faction]),'敵方艦隊陣營同步'),
                            (0x6070c+1,bytes([1-faction]),'敵方提督陣營同步'),
                            (target+0x34,struct.pack('<HH',x+(20 if where=='near' else 0),y),'局部係數診斷：零距離或距離20')]:
        poke(addr,data,why,step)
    expected_type = mem[0x54b30+0x4e3a+faction*208+kind*26+10+weapon*4]
    assert expected_type == t
    input_step, saved = step+1, 'end.state'
    if where == 'stack-first':
        assert t == 0
        poke(0x65a7a, struct.pack('<HH',0x45cc,0x54b3), '局部反例：只把未初始化的前項指標改成曲線首項；不作正常玩家聲明', 506950550)
        input_step = 506950551
    inherited = 'combat79-r63-laser-submit-a-plan.json'
    stage = 'coefficient-domain'
elif case == 'multi-entry':
    parent = Path('workplace/oracle/combat79-r62-natural-a/entry.state')
    assert hashlib.sha256(parent.read_bytes()).hexdigest() == 'c4147b9d310c84824d03d43407cf3b335b950e1aacc52ee79dc37a7563b494cb'
    base = json.loads(Path('workplace/weapon79-r65-near-a-plan.json').read_text())
    rows = base['injections'].copy()
    start, step, limit = 484357727, 484357728, 486000000
    poke(second+0x1e,[127],'第二個一般分隊82可命令，81是獨立偵察池',step)
    poke(second+0x34,struct.pack('<HH',416,320),'第二攻方固定在下方40',step)
    t2 = 0x668b2+104*78
    poke(t2+0x34,struct.pack('<HH',436,320),'第二敵方目標104固定在下方40',step)
    poke(t2+0x17,[2],'第二目標Detail2',step)
    poke(t2+0x1e,[127],'第二目標Out7F',step)
    input_step, saved = step+1, 'prompt.state'
    inherited = 'weapon79-r65-near-a-plan.json'
    stage = 'controlled-entry'
else:
    stages = {
        'multi-command': ('multi-entry','prompt.state',485999999,490000000,'command.state',['486400000:240:230:1:100000']),
        'multi-first': ('multi-command','command.state',489999999,498000000,'first.state',['490300000:268:208:1:100000','495300000:270:250:1:100000']),
        'multi-second': ('multi-first','first.state',497999999,508000000,'submitted.state',['502000000:268:248:1:100000','507000000:270:290:1:100000']),
        'same-target-second': ('multi-first','first.state',497999999,508000000,'submitted.state',['502000000:268:208:1:100000','507000000:270:250:1:100000']),
        'multi-end': ('multi-second','submitted.state',507999999,800000000,'end.state',[]),
        'multi-order': ('multi-second','submitted.state',507999999,800000000,'end.state',[]),
        'multi-ranked': ('multi-second','submitted.state',507999999,800000000,'end.state',[]),
        'same-target-ranked': ('same-target-second','submitted.state',507999999,800000000,'end.state',[]),
        'same-target-measured': ('same-target-second','submitted.state',507999999,800000000,'end.state',[]),
        'same-target-end': ('same-target-second','submitted.state',507999999,800000000,'end.state',[]),
        'cancel-outer': ('EXTERNAL','command.state',489999999,510000000,'end.state',['490400000:400:300:2:100000']),
    }
    parent_case, file, start, limit, saved, clicks = stages[case]
    if parent_case == 'EXTERNAL':
        parent = Path('workplace/oracle/weapon79-r65-command-a/command.state')
        inherited = 'weapon79-r65-command-a-plan.json'
    else:
        parent = Path(f'workplace/oracle/sequence79-r66-{parent_case}-a/{file}')
        inherited = f'sequence79-r66-{parent_case}-a-plan.json'
    assert parent.is_file()
    input_step, stage = start+1, 'original-command-sequence'
    if case in ('multi-order', 'multi-ranked', 'same-target-ranked', 'same-target-measured'):
        regs += ',1DD6:033E,1DD6:0522,1DD6:0539'
flags = ['-load-state','/out/'+str(parent.relative_to('workplace')),'-steps',str(limit),'-tick','20000',
         '-regs-from',str(start),'-regs-max','256','-regs-at',regs,
         '-save-state',f'{limit-1}:{prefix}/{saved}',
         '-dump-mem-at',f'{input_step}:lin:0:1048576:{prefix}/input-mem.bin;{limit-1}:lin:0:1048576:{prefix}/end-mem.bin']
if rows:
    flags += ['-poke',';'.join(f"{r['address']}@{r['step']}={r['bytes_hex']}" for r in rows)]
if clicks:
    flags += ['-clicks',','.join(clicks),'-click-move-lead','200000']
if case in ('multi-order', 'multi-ranked', 'same-target-ranked', 'same-target-measured'):
    flags[flags.index('-dump-mem-at')+1] += f';523550689:lin:0:1048576:{prefix}/damage1-mem.bin;561173442:lin:0:1048576:{prefix}/damage2-mem.bin'
if case in ('multi-ranked', 'same-target-ranked'):
    flags[flags.index('-dump-mem-at')+1] += f';508249335:lin:0:1048576:{prefix}/keys-mem.bin;508251432:lin:0:1048576:{prefix}/sorted-mem.bin'
if case == 'same-target-measured':
    # 同目標分支排序邊界晚13264步；錯相位的ranked收據保留，不覆寫。
    flags[flags.index('-dump-mem-at')+1] += f';508262599:lin:0:1048576:{prefix}/keys-mem.bin;508264696:lin:0:1048576:{prefix}/sorted-mem.bin'
out = Path('workplace')/(name+'-flags.txt')
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
out.write_text('\n'.join(flags)+'\n')
plan = {'name':name,'case':case,'stage':stage,'side':side,'parent':str(parent),
        'parent_sha256':hashlib.sha256(parent.read_bytes()).hexdigest(),'injections':rows,
        'clicks':clicks,'inherited_plan':inherited,'input_dump_step':input_step,'end_dump_step':limit-1,
        'classification':'controlled original; no normal-remake or whole-phase parity claim'}
out.with_name(name+'-plan.json').write_text(json.dumps(plan,ensure_ascii=False,indent=2)+'\n')
print(name,stage,'注入',len(rows),'點擊',len(clicks))
