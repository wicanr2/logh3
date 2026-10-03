"""#79第63輪：固定命令父狀態、完整輸入及獨立輸出，不覆寫舊收據。"""
import hashlib
import json
import os
from pathlib import Path
import re
import sys

name=sys.argv[1]
match=re.fullmatch(r'combat79-r63-([a-z0-9-]+)-([ab])',name)
assert match
kind,side=match.groups()
requested_kind=kind
submit=kind.endswith('-submit')
finish=kind.endswith('-finish')
phase_end=kind.endswith('-phase-end')
if phase_end:
    kind=kind.removesuffix('-phase-end')
elif submit or finish:
    kind=kind.rsplit('-',1)[0]
parent=Path('workplace/oracle/combat79-r62-command-a/command.state')
parent_hash='108d2cac6ab7cb278c932f97542a9476ef6a9da683eb87976be694fceae6af95'
cases={'target':[(490100000,348,248)],'blank':[(490300000,400,350)],
       'hover':[],'target-premove':[(490300000,348,248)],
       'target-mapped':[(490300000,348,208)],
       'laser':[(495300000,350,210)],'missile':[(495300000,350,230)],
       'disabled':[(495300000,350,250)],'cancel':[(495300000,350,210)],
       'defence':[(495300000,350,270)],'auto':[(495300000,350,250)]}
assert kind in cases
limit=495000000
start=489999999
parent_plan='combat79-r62-command-a-plan.json'
if kind in ('laser','missile','disabled','cancel'):
    parent=Path('workplace/oracle/combat79-r63-target-mapped-a/selected.state')
    parent_hash='5437f227c6bddf5831e8f305b144a6a2702f85a648b48422b84ad522c8c5d245'
    limit=505000000
    start=494999999
    parent_plan='combat79-r63-target-mapped-a-plan.json'
if kind in ('defence','auto'):
    parent=Path('workplace/oracle/combat79-r63-target-a/selected.state')
    parent_hash='db488d25851ef2af76f599b7057cb34df3931aa34e06fdb2283682f816727e13'
    limit=505000000
    start=494999999
    parent_plan='combat79-r63-target-a-plan.json'
if submit:
    assert kind in ('laser','missile','disabled','cancel','defence')
    limit=495600000
if finish or phase_end:
    assert kind in ('laser','missile')
    parent=Path(f'workplace/oracle/combat79-r63-{kind}-a/selected.state')
    parent_hash={'laser':'d6c65f6e6e856e0274001b7d4fdf5a49093163c4cbd899fd10b3b3253b78a33d',
                 'missile':'e2171ddfc99df2ddc2488f6a3f9b5123e55fb35d3a4d397fe911514b9d5222e6'}[kind]
    start=504999999
    limit=700000000
    if phase_end:
        limit={'laser':533356600,'missile':534956600}[kind]
    parent_plan=f'combat79-r63-{kind}-a-plan.json'
assert hashlib.sha256(parent.read_bytes()).hexdigest()==parent_hash
inputs=[] if finish or phase_end else cases[kind]
prefix='/out/oracle/'+name
flags=['-load-state','/out/'+str(parent.relative_to('workplace')),
       '-steps',str(limit),'-tick','20000',
       '-clicks',','.join(f'{step}:{x}:{y}'+(':2'if kind=='cancel'else '')for step,x,y in inputs),'-click-hold','100000',
       '-regs-from',str(start),'-regs-max','64',
       '-regs-at','1ED2:02AB,1ED2:02E2,1ED2:031A,1ED2:0336,1ED2:036D,1ED2:037C,1ED2:03E9,1ED2:0616,1ED2:06E1,1ED2:0781,1DD6:0152,1DD6:030F,1DD6:026B,181B:0385,3CD5:000C',
       '-save-state',f'{limit-1}:{prefix}/selected.state',
       '-dump-mem-at',f'{limit-1}:lin:0:1048576:{prefix}/end-mem.bin']
if kind=='hover':
    flags += ['-mouse-x','348','-mouse-y','248','-mouse-at','490100000']
else:
    flags += ['-click-move-lead','200000']
out=Path('workplace')/(name+'-flags.txt')
assert not out.exists() and out.parent.stat().st_uid==os.getuid()
out.write_text('\n'.join(flags)+'\n')
plan={'name':name,'kind':requested_kind,'parent_state_sha256':parent_hash,
      'classification':'controlled entry diagnostic; production parity not claimed',
      'injections':[],'inherited_injections':'combat79-r62-gate-near-human-a-plan.json',
      'parent_plan':parent_plan,'mouse_coordinate_space':'原版virtual400；本次PNG原生Y另加40',
      'seed':0x12345678,'inputs':[{'step':step,'x':x,'y':y,'button':2 if kind=='cancel'else 1,'hold_steps':100000}for step,x,y in inputs]}
if kind=='hover':
    plan['movements']=[{'step':490100000,'x':348,'y':248}]
with out.with_name(name+'-plan.json').open('x',encoding='utf-8') as f:
    json.dump(plan,f,ensure_ascii=False,indent=2)
    f.write('\n')
print(name,'固定輸入',inputs,'父狀態',parent_hash)
