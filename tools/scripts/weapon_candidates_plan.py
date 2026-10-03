"""第65輪武器候選與正式選單診斷；唯一入口及完整注入可回查。"""
import hashlib
import json
import os
from pathlib import Path
import re
import struct
import sys

name = sys.argv[1]
match = re.fullmatch(r'weapon79-r65-([a-z0-9-]+)-([ab])', name)
assert match
case, side = match.groups()
prefix = '/out/oracle/' + name
rows = []
own, target = 0x68112, 0x687c6
stage = case if case in ('command', 'selected', 'fighter-submit', 'fighter-end', 'fighter-phase-end') else 'entry'
parent_case = {'command': 'near', 'selected': 'command', 'fighter-submit': 'selected',
               'fighter-end': 'fighter-submit', 'fighter-phase-end': 'fighter-submit'}
state_file = {'command': 'prompt.state', 'selected': 'command.state',
              'fighter-submit': 'selected.state', 'fighter-end': 'submitted.state', 'fighter-phase-end': 'submitted.state'}
clicks = []
if stage == 'entry':
    parent = Path('workplace/oracle/combat79-r62-natural-a/entry.state')
    expected_hash = 'c4147b9d310c84824d03d43407cf3b335b950e1aacc52ee79dc37a7563b494cb'
    assert hashlib.sha256(parent.read_bytes()).hexdigest() == expected_hash
    mem = Path('workplace/oracle/combat79-r62-natural-a/entry-mem.bin').read_bytes()
    assert hashlib.sha256(mem).hexdigest() == 'fbe061f125a0554bb6513612c45cba422d26441338e7b431c6b5316473bf6ec1'
    step, limit = 484357728, 486000000
    cases = {'near': {}, 'fighter60': {'distance': 60}, 'fighter61': {'distance': 61},
             'missile132': {'distance': 132}, 'missile133': {'distance': 133},
             'beam191': {'distance': 191}, 'beam192': {'distance': 192},
             'search300': {'distance': 300}, 'search301': {'distance': 301},
             'rear20': {'facing': 128}, 'rear21': {'facing': 128, 'distance': 21},
             'side40': {'facing': 64, 'distance': 40}, 'side41': {'facing': 64, 'distance': 41},
             'missile-zero': {'missile': 0}, 'fighter-zero': {'fighter': 0},
             'stock-one': {'missile': 1, 'fighter': 1}, 'detail-zero': {'detail': 0},
             'scout-kind': {'kind': 7}, 'target-zero': {'ships': 0},
             'target-hidden': {'out': 255}, 'rear8': {'facing': 128, 'distance': 8},
             'rear9': {'facing': 128, 'distance': 9}, 'rear27': {'facing': 128, 'distance': 27},
             'rear28': {'facing': 128, 'distance': 28}, 'side16': {'facing': 64, 'distance': 16},
             'side17': {'facing': 64, 'distance': 17}, 'side45': {'facing': 64, 'distance': 45},
             'side46': {'facing': 64, 'distance': 46}, 'beam-absent': {'absent': 10},
             'missile-absent': {'absent': 14}, 'fighter-absent': {'absent': 18},
             'coincident': {'distance': 0}, 'block-visible': {'block': 127},
             'block-hidden': {'block': 255}, 'block-offset': {'block': 127, 'block_y': 17},
             'terrain-block': {'terrain': 2}, 'terrain-clear': {'terrain': 0},
             'search-wide300': {'distance': 300, 'wide': True},
             'search-wide301': {'distance': 301, 'wide': True},
             'left20': {'distance': -20, 'facing': 128},
             'up20': {'distance': 0, 'dy': -20, 'facing': 192},
             'down20': {'distance': 0, 'dy': 20, 'facing': 64},
             'diagonal20': {'distance': 20, 'dy': 20, 'facing': 32},
             'steep20': {'distance': 10, 'dy': -20, 'facing': 211},
             'shallow20': {'distance': -20, 'dy': 10, 'facing': 109}}
    cases.update({'search-open300': {'distance': 300, 'open': True},
                  'search-open301': {'distance': 301, 'open': True},
                  'stock-highword': {'missile': 65536, 'fighter': 65536}})
    assert case in cases
    c = cases[case]

    def poke(addr, data, why):
        rows.append({'address': f'lin:{addr:x}', 'step': step, 'bytes_hex': bytes(data).hex(' '), 'reason': why})

    poke(0x5e00c, struct.pack('<I', 0x12345678), '第④步唯一入口固定seed，不重擲')
    for index in [12, 16, 17, 28, 29, 30, 31]:
        off = 0x60ba4 + index*122
        r = mem[off:off+122]
        poke(off+0x48, [4], '非指定艦隊待命')
        for j in range(r[15]):
            pool = struct.unpack_from('<H', r, 0x10+j*4)[0]
            poke(0x668b2+pool*78+0x1e, [255], '除自身80及目標102外一般分隊Out=FF')
            poke(0x668b2+pool*78+0x2e, [125], '觀測原版初始化清武器欄位')
        scout = struct.unpack_from('<H', r, 0x38)[0]
        poke(0x668b2+scout*78+0x2e, [125], '保持獨立偵察池對照')
    poke(0x6115c+0x48, [1], '只讓艦隊12進戰鬥命令')
    poke(0x6115c+0x4a, struct.pack('<H', c.get('facing', 0)), '艦隊12朝向，驗射程角度')
    poke(own+0x1e, [127], '自身80可進命令')
    x, y = struct.unpack_from('<HH', mem, own+0x34)
    poke(target+0x34, struct.pack('<HH', x+c.get('distance', 20), y+c.get('dy', 0)), '目標102固定相對座標')
    poke(target+0x17, [c.get('detail', 2)], '候選情報門檻')
    poke(target+0x1e, [c.get('out', 127)], '候選Out門檻')
    for key, offset, size in [('missile', 0x0e, 4), ('fighter', 0x12, 4)]:
        if key in c:
            poke(own+offset, c[key].to_bytes(size, 'little'), '候選庫存 '+key)
    for key, offset, size in [('kind', 4, 1), ('ships', 8, 2)]:
        if key in c:
            poke(target+offset, c[key].to_bytes(size, 'little'), '候選目標 '+key)
    if 'absent' in c:
        kind_base = 0x54b30+0x4e3a+mem[own]*208+mem[own+4]*26
        poke(kind_base+c['absent'], [255], '攻方艦種沒有指定武器')
    if 'block' in c:
        block = 0x668b2+81*78
        poke(block+0x34, struct.pack('<HH', x+10, y+c.get('block_y', 0)), '分隊81置於射線中段或側邊')
        poke(block+0x1e, [c['block']], '射線分隊遮蔽Out bit7對照')
    if 'terrain' in c:
        ptr = struct.unpack_from('<HH', mem, 0x54b30+0x3c1a)
        linear = ptr[1]*16+ptr[0]
        cell_x, cell_y = (x+15)//8, (y+7)//10
        cell = linear+(cell_y//2)*128+cell_x//2
        poke(cell, [255], '第8個射線取樣點的原版地形格改255')
        poke(0x54b30+0x9bb4+255, [c['terrain']], '地形255碰撞mask清零或設右上子格bit1')
    if c.get('wide'):
        poke(0x54b30+0x4fda+2, struct.pack('<HH', 400, 400), '早期射程probe：400圓造成b平方16.16截斷，不能隔離300搜尋上限')
    if c.get('open'):
        poke(0x54b30+0x4fda+2, struct.pack('<HH', 400, 100), '光線砲前向射程大於300；b平方保持signed16.16可表示')
    save = f'{limit-1}:{prefix}/prompt.state'
    input_step = step+1
else:
    parent = Path(f'workplace/oracle/weapon79-r65-{parent_case[stage]}-a/{state_file[stage]}')
    assert parent.is_file()
    step, limit, save_name = {
        'command': (485999999, 490000000, 'command.state'),
        'selected': (489999999, 495000000, 'selected.state'),
        'fighter-submit': (494999999, 495600000, 'submitted.state'),
        'fighter-end': (495599999, 700000000, 'end.state'),
        'fighter-phase-end': (495599999, 535946393, 'end.state')
    }[stage]
    if stage == 'command':
        clicks = ['486400000:240:230:1:100000']
    elif stage == 'selected':
        clicks = ['490300000:268:208:1:100000']
    elif stage == 'fighter-submit':
        clicks = ['495300000:270:250:1:100000']
    input_step = step+1
    save = f'{limit-1}:{prefix}/{save_name}'
flags = ['-load-state', '/out/'+str(parent.relative_to('workplace')), '-steps', str(limit), '-tick', '20000',
         '-regs-from', str(step), '-regs-max', '256',
         '-regs-at', '1F6F:02D2,1F6F:032D,1F6F:0382,1F6F:0398,1F6F:056C,1F6F:091C,1ED2:0097,1ED2:0205,1ED2:036D,1DD6:0152,1DD6:030F,1DD6:0C54,1DD6:0F6D,1DD6:026B,181B:0385,3CD5:000C',
         '-save-state', save,
         '-dump-mem-at', f'{input_step}:lin:0:1048576:{prefix}/input-mem.bin;{limit-1}:lin:0:1048576:{prefix}/end-mem.bin']
if rows:
    flags += ['-poke', ';'.join(f"{r['address']}@{r['step']}={r['bytes_hex']}" for r in rows)]
if clicks:
    flags += ['-clicks', ','.join(clicks), '-click-move-lead', '200000']
out = Path('workplace')/(name+'-flags.txt')
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
out.write_text('\n'.join(flags)+'\n')
plan = {'name': name, 'case': case, 'side': side, 'stage': stage,
        'parent': str(parent), 'parent_sha256': hashlib.sha256(parent.read_bytes()).hexdigest(),
        'classification': 'controlled-original-candidate-and-command; no normal-remake-player claim',
        'seed': '12345678', 'injections': rows, 'clicks': clicks,
        'input_dump_step': input_step,
        'inherited_plan': None if stage == 'entry' else f'weapon79-r65-{parent_case[stage]}-a-plan.json'}
out.with_name(name+'-plan.json').write_text(json.dumps(plan, ensure_ascii=False, indent=2)+'\n')
print(name, '注入', len(rows), 'stage', stage)
