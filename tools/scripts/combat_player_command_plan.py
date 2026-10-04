"""第75輪固定自動選擇入口；每項注入逐一保存，不以重擲取得局面。"""
import hashlib
import json
import os
from pathlib import Path
import re
import struct
import sys

name = sys.argv[1]
m = re.fullmatch(r'command79-r75-([a-z0-9-]+)-([ab])', name)
assert m
case = m[1]
cases = {
    'base': {}, 'beam-absent': {'absent': [10]},
    'missile-only': {'absent': [10, 18]}, 'fighter-only': {'absent': [10, 14]},
    'defence': {'defence': True}, 'anticipated-zero': {'anticipated': 0},
    'none': {'none': True}, 'two-low': {'count': 2, 'rnd': 8191},
    'three-low': {'count': 3, 'rnd': 8191},
    'three-middle': {'count': 3, 'rnd': 8192},
    'three-middle-end': {'count': 3, 'rnd': 16383},
    'three-high': {'count': 3, 'rnd': 16384},
    'three-overflow': {'count': 3, 'rnd': 32767, 'ability': 255},
    'three-no-beam': {'count': 3, 'absent': [10], 'rnd': 16384},
    'computer-flagship': {'computer': True},
    'computer-kind5': {'computer': True, 'kind': 5, 'squadron': 1},
    'computer-kind6': {'computer': True, 'kind': 6, 'squadron': 1},
    'computer-kind0': {'computer': True, 'squadron': 1},
    'computer-word': {'computer': True, 'anticipated': 65535},
    'near': {'distance': 20},
    'missile-stock1': {'absent': [10, 18], 'missiles': 1},
    'fighter-stock1': {'distance': 20, 'own_kind': 3, 'fighters': 1},
    'fighter-half': {'distance': 20, 'own_kind': 3},
    'general-two': {'count': 2, 'rnd': 8191, 'general': True},
    'general-three-low': {'count': 3, 'rnd': 8191, 'general': True},
    'general-three-middle': {'count': 3, 'rnd': 8192, 'general': True},
    'general-three-high': {'count': 3, 'rnd': 16384, 'general': True},
    'general-three-no-beam': {'count': 3, 'rnd': 16384, 'absent': [10], 'general': True},
    'computer-general-kind0': {'computer': True, 'kind': 0, 'target_general': True},
    'computer-general-kind5': {'computer': True, 'kind': 5, 'target_general': True},
    'computer-general-kind6': {'computer': True, 'kind': 6, 'target_general': True},
}
assert case == 'trace' or case in cases
parent = Path('workplace/oracle/combat79-r63-target-a/selected.state')
parent_hash = 'db488d25851ef2af76f599b7057cb34df3931aa34e06fdb2283682f816727e13'
assert hashlib.sha256(parent.read_bytes()).hexdigest() == parent_hash
prefix = '/out/oracle/' + name
limit = 497000000 if case == 'trace' else 495900000
rows = []
if case != 'trace':
    c = cases[case]
    m = Path('workplace/oracle/command79-r75-trace-a/end-mem.bin').read_bytes()
    own, target = 0x668b2 + 80*78, 0x668b2 + 102*78
    def poke(address, data, why):
        rows.append({'address': f'lin:{address:x}', 'step': 495000000,
                     'bytes_hex': bytes(data).hex(' '), 'reason': why})
    if c.get('target_general'):
        rec = bytearray(m[target:target+78])
        rec[3] = 1
        poke(0x668b2+104*78, rec, '真正一般分隊pool104，保持編制第1組身分')
        poke(target+0x17, [0], '旗艦102不作候選，隔離非旗艦權重')
        target = 0x668b2+104*78
    seed = 0x12345678
    if 'rnd' in c:
        # 依已證實LCG反解唯一預設種子，控制首次RND值；不試種子或重擲。
        seed = (((c['rnd'] << 16) - 0x3039) * pow(0x41c64e6d, -1, 1 << 32)) & 0xffffffff
        poke(0x6115c+0x45, [c.get('ability', 100)], '選擇倍率FLEET+45；保持原byte語意')
    poke(0x5e00c, struct.pack('<I', seed), '固定首次RND條件，無重擲')
    poke(own+0x3e, struct.pack('<H', 3000), '清前一診斷暫存，自動入口前anticipated=艘數')
    poke(target+0x3e, struct.pack('<H', c.get('anticipated', 3000)), '固定目標預估剩餘艘數')
    if 'distance' in c:
        x, y = struct.unpack_from('<HH', m, own+0x34)
        poke(target+0x34, struct.pack('<HH', x+c['distance'], y), '固定非零近距離')
    if 'own_kind' in c:
        poke(own+4, [c['own_kind']], '攻方艦種3架數除二對照')
    for field, offset in [('missiles', 14), ('fighters', 18)]:
        if field in c:
            poke(own+offset, struct.pack('<I', c[field]), '預估量夾庫存 '+field)
    if c.get('none'):
        poke(target+0x17, [0], '自動重新查詢沒有情報候選')
    if c.get('defence'):
        poke(target+0x2e, [3], '目標已有防禦命令')
    for offset in c.get('absent', []):
        poke(0x54b30+0x4e3a+offset, [255], '攻方艦種缺指定武器')
    if c.get('computer'):
        poke(0x54b30+0x8f21, [1], '原版攻方座位control=電腦；正式061政策不因此改動')
    for field, offset in [('kind', 4), ('squadron', 3)]:
        if field in c:
            poke(target+offset, [c[field]], 'CPU評分對照目標'+field)
    for i in range(1, c.get('count', 1)):
        pool = (103+i) if c.get('general') else (102+i)
        address = 0x668b2 + pool*78
        rec = bytearray(m[target:target+78])
        rec[3] = i
        rec[0x2e] = 255
        x, y = struct.unpack_from('<HH', m, own+0x34)
        struct.pack_into('<HH', rec, 0x34, x+20, y+20 if i == 1 else y-20)
        struct.pack_into('<H', rec, 0x3e, 1000 if i == 1 else 2000)
        poke(address, rec, '受控新增候選，保留原池身分；固定座標／預估艘數')
flags = ['-load-state', '/out/' + str(parent.relative_to('workplace')),
         '-steps', str(limit), '-tick', '20000',
         '-clicks', '495300000:350:250', '-click-hold', '100000',
         '-click-move-lead', '200000', '-regs-from', '494999999', '-regs-max', '96',
         '-regs-at', '1F6F:005F,1F6F:0078,1F6F:00BB,1F6F:00DB,1F6F:00F0,1F6F:0185,1F6F:01CA,1F6F:0714,1F6F:072F,1F6F:0745,1F6F:086C,1F6F:08CC,1F6F:0299,1ED2:01CE,1ED2:01DF,1DD6:0152,3CD5:000C',
         '-save-state', f'{limit-1}:{prefix}/end.state',
         '-dump-mem-at', f'495000001:lin:0:1048576:{prefix}/input-mem.bin;{limit-1}:lin:0:1048576:{prefix}/end-mem.bin']
if rows:
    flags += ['-poke', ';'.join(f"{r['address']}@{r['step']}={r['bytes_hex']}" for r in rows)]
p = Path('workplace') / (name + '-flags.txt')
assert not p.exists() and p.parent.stat().st_uid == os.getuid()
p.write_text('\n'.join(flags) + '\n')
plan = {'name': name, 'case': case, 'parent_state_sha256': parent_hash,
        'classification': 'controlled-original diagnostic; not normal remake parity',
        'inherited_plan': 'combat79-r63-target-a-plan.json', 'injections': rows,
        'seed_before': f'{seed:08x}' if case != 'trace' else '12345678',
        'controlled_first_rnd': cases[case].get('rnd') if case != 'trace' else None,
        'inputs': [{'step': 495300000, 'x': 350, 'y': 250, 'button': 1,
                    'hold_steps': 100000, 'move_lead_steps': 200000}],
        'stop_step': limit - 1}
p.with_name(name + '-plan.json').write_text(json.dumps(plan, ensure_ascii=False, indent=2) + '\n')
print(name, '固定父快照及輸入')
