"""第64輪原版傷害owner：固定父狀態、唯一輸入與拒絕覆寫。"""
import hashlib
import json
import os
from pathlib import Path
import re
import sys

name = sys.argv[1]
m = re.fullmatch(r'damage79-r64-([a-z0-9-]+)-([ab])', name)
assert m
case, side = m.groups()
observe = case in ('laser-observe', 'missile-observe', 'laser-interpolate', 'missile-interpolate')
kind = case.removesuffix('-observe').removesuffix('-interpolate') if observe else 'laser'
parent = Path(f'workplace/oracle/combat79-r63-{kind}-submit-a/selected.state')
parent_hash = {'laser': 'af0019ce8a29ea2e0026a0ba15ac4f37d870f848000247a07250ee10190de19d',
               'missile': '1b84dd2f0bf76cfd5b526d4df18d7dcbc1f7330478c35e2d97844f6dfaf0457a'}[kind]
assert parent.is_file() and hashlib.sha256(parent.read_bytes()).hexdigest() == parent_hash
limit = 508000000
prefix = '/out/oracle/' + name
ship = 0x668b2
own, target = ship + 80*78, ship + 102*78
target_fleet = 0x60ba4 + 16*122
target_adml = 0x54b30 + 0xb77c + 40*28
changes = {
    'laser-fixed': [],
    'missile-fixed': [(own+0x2e, 1, 1, '攻擊武器改1')],
    'missile-low': [(own+0x2e, 1, 1, '攻擊武器改1'), (own+0x0e, 499, 4, '飛彈499')],
    'missile-zero': [(own+0x2e, 1, 1, '攻擊武器改1'), (own+0x0e, 0, 4, '飛彈0')],
    'fighter-fixed': [(own+0x2e, 2, 1, '攻擊武器改2')],
    'fighter-low': [(own+0x2e, 2, 1, '攻擊武器改2'), (own+0x12, 499, 4, '戰鬥機499')],
    'fighter-zero': [(own+0x2e, 2, 1, '攻擊武器改2'), (own+0x12, 0, 4, '戰鬥機0')],
    'fighter-kind3': [(own+0x2e, 2, 1, '攻擊武器改2'), (own+4, 3, 1, '攻擊艦種3')],
    'defence': [(target+0x2e, 3, 1, '目標武器改防禦3')],
    'nonflagship': [(target+3, 1, 1, '目標分隊序1')],
    'target99': [(target+8, 99, 2, '目標99艘')],
    'target100': [(target+8, 100, 2, '目標100艘')],
    'target1': [(target+8, 1, 2, '目標1艘')],
    'flag-high101': [(target_fleet+0x3d, 20, 1, '目標防禦能力20；此早期名稱誤讀101，實際提督屬性104，保持舊收據旗標')],
    'flag-high104': [(target_fleet+0x3d, 20, 1, '目標防禦能力20；提督屬性原值104')],
    'flag-high-cpu': [(target_fleet+0x3d, 20, 1, '目標防禦能力20'), (target_adml+0x12, 100, 1, '提督屬性100'), (0x54b30+0x8f22, 1, 1, '對手操作模式1')],
    'flag-high-human': [(target_fleet+0x3d, 20, 1, '目標防禦能力20'), (target_adml+0x12, 100, 1, '提督屬性100'), (0x54b30+0x8f22, 0, 1, '對手操作模式0')],
    'angle-front': [(target_fleet+0x4a, 128, 1, '目標朝向128')],
    'angle-back': [(target_fleet+0x4a, 0, 1, '目標朝向0')],
    'curve-near': [(target+0x34, 426, 2, '目標X改426，距離10')],
    'curve-far': [(target+0x34, 666, 2, '目標X改666，距離250')],
    'curve-down': [(target+0x34, 591, 2, '目標X改591，距離175，下降曲線')],
    'low-negative-odd': [(target+8, 99, 2, '目標99艘'), (target_adml+0x12, 130, 1, '129減130，負奇數除二')],
    'low-negative-even': [(target+8, 99, 2, '目標99艘'), (target_adml+0x12, 131, 1, '129減131，負偶數除二')],
    'zero-fallback': [(own+0x2e, 1, 1, '攻擊武器改1'), (own+0x0e, 0, 4, '飛彈0'), (target_adml+0x12, 200, 1, '負傷害夾0，再取第二次亂數')],
    'fallback-fixed': [(own+0x2e, 1, 1, '攻擊武器改1'), (own+0x0e, 0, 4, '飛彈0'), (target_adml+0x12, 200, 1, '在傷害入口固定seed；負傷害夾0再取第二次亂數')],
    'zero-own': [(own+8, 0, 2, '攻擊艘數0，在05D7前注入')],
    'zero-target': [(target+8, 0, 2, '目標艘數0，在05D7前注入')],
}
assert observe or case in changes
step = 495600001 if case.startswith('zero-') else 506948195
injections = []
if not observe:
    for address, value, size, reason in [(0x5e00c, 0x12345678, 4, '局部戰損固定seed，隔離前段排序骰序')] + changes[case]:
        injections.append({'step': step, 'address': f'lin:{address:x}', 'value_hex': value.to_bytes(size, 'little').hex(' '), 'reason': reason})
input_step = 495600000 if observe else step+1
flags = ['-load-state', '/out/' + str(parent.relative_to('workplace')),
         '-steps', str(limit), '-tick', '20000', '-regs-from', '495599999', '-regs-max', '64',
         '-regs-at', '1DD6:030F,1DD6:05D7,1DD6:0C54,1DD6:0C6D,1DD6:0C83,1DD6:0E17,1DD6:0E38,1DD6:0F25,1DD6:0F59,1DD6:0F60,1DD6:0F6D,203A:0002,203A:00E3,203A:012A,203A:01C4,203A:01E9,203A:020B,203A:0214,203A:02C2,203A:02D4,203A:03A7,203A:03B0,203A:0406,203A:0593,203A:0743,3CD5:000C',
         '-save-state', f'{limit-1}:{prefix}/damage.state',
         '-dump-mem-at', f'{input_step}:lin:0:1048576:{prefix}/input-mem.bin;{limit-1}:lin:0:1048576:{prefix}/end-mem.bin']
if injections:
    flags += ['-poke', ';'.join(f"{i['address']}@{i['step']}={i['value_hex']}" for i in injections)]
out = Path('workplace')/(name+'-flags.txt')
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
out.write_text('\n'.join(flags)+'\n')
plan = {'name': name, 'kind': kind, 'case': case, 'parent_state_sha256': parent_hash,
        'classification': 'controlled-original-damage-owner; no normal player parity',
        'parent_plan': f'combat79-r63-{kind}-submit-a-plan.json',
        'inherited_seed': '12345678', 'injections': injections, 'inputs': [],
        'scope': '既有選擇後正常callee；有注入者限定damage-only診斷，不證明武器選單可達或演出形狀',
        'input_dump_step': input_step, 'end_dump_step': limit-1}
with out.with_name(name+'-plan.json').open('x', encoding='utf-8') as f:
    json.dump(plan, f, ensure_ascii=False, indent=2)
    f.write('\n')
print(name, '固定父狀態', parent_hash)
