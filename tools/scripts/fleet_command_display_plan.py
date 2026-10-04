"""第77輪固定命令等待；只控制敵方情報，保留完整父狀態與輸入。"""
import hashlib
import json
import os
from pathlib import Path
import re
import sys

name = sys.argv[1]
m = re.fullmatch(r'combat79-r77-(wait|hover[123]|popup[123])-([ab])', name)
assert m
case = m[1]
parent = Path('workplace/oracle/combat79-r62-command-a/command.state')
sha = '108d2cac6ab7cb278c932f97542a9476ef6a9da683eb87976be694fceae6af95'
assert hashlib.sha256(parent.read_bytes()).hexdigest() == sha
prefix = '/out/oracle/' + name
rows = []
if case != 'wait':
    rows.append({'address': f'lin:{0x668b2+102*78+0x17:x}', 'step': 490050000,
                 'bytes_hex': f'{int(case[-1]):02x}', 'reason': '控制同一敵方池102的情報等級；不改候選、位置或seed'})
flags = ['-load-state', '/out/' + str(parent.relative_to('workplace')),
         '-steps', '495000000', '-tick', '20000',
         '-regs-from', '489999999', '-regs-max', '64',
         '-regs-at', '1ED2:06E1,1ED2:0717,1ED2:0781,1ED2:07DF,1ED2:0847,1ED2:0912,1ED2:0960,1ED2:03E9',
         '-save-state', f'494999999:{prefix}/end.state',
         '-dump-mem-at', f'490050001:lin:0:1048576:{prefix}/input-mem.bin;494999999:lin:0:1048576:{prefix}/end-mem.bin']
if case.startswith('popup'):
    flags += ['-clicks', '490300000:348:208', '-click-hold', '100000', '-click-move-lead', '200000']
elif case.startswith('hover'):
    flags += ['-mouse-at', '490100000', '-mouse-x', '348', '-mouse-y', '208']
if rows:
    flags += ['-poke', ';'.join(f"{r['address']}@{r['step']}={r['bytes_hex']}" for r in rows)]
p = Path('workplace') / (name + '-flags.txt')
assert not p.exists() and p.parent.stat().st_uid == os.getuid()
p.write_text('\n'.join(flags) + '\n')
plan = {'name': name, 'case': case, 'parent_state': str(parent), 'parent_sha256': sha,
        'parent_plan': 'combat79-r62-command-a-plan.json',
        'inherited_injections': 'combat79-r62-gate-near-human-a-plan.json',
        'comparison_level': 'same-state controlled diagnostic; not normal player path',
        'mouse_coordinate_space': 'original virtual640x400; framebuffer Y=mouse Y+40',
        'seed_before': '12345678', 'injections': rows, 'flags': flags}
with p.with_name(name + '-plan.json').open('x', encoding='utf-8') as f:
    json.dump(plan, f, ensure_ascii=False, indent=2)
    f.write('\n')
print(name, '固定輸入及情報', rows)
