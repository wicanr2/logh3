"""第62輪參戰門檻診斷；唯一入口注入完整列在固定計畫，禁止重擲／覆寫。"""
import hashlib
import json
import os
from pathlib import Path
import re
import struct
import sys

name=sys.argv[1]
if re.fullmatch(r'combat79-r62-command-[ab]',name):
    parent=Path('workplace/oracle/combat79-r62-gate-near-human-a/prompt.state')
    parent_hash='8703c01dcf29c68c9212a4c1f590e16382b177e898403fef3894dab9dfda2d04'
    assert hashlib.sha256(parent.read_bytes()).hexdigest()==parent_hash
    out=Path('workplace')/(name+'-flags.txt')
    assert not out.exists() and out.parent.stat().st_uid==os.getuid()
    prefix='/out/oracle/'+name
    out.write_text('\n'.join(['-load-state','/out/oracle/combat79-r62-gate-near-human-a/prompt.state',
        '-steps','490000000','-tick','20000','-clicks','486100000:240:230','-click-hold','100000',
        '-regs-from','485999999','-regs-max','128',
        '-regs-at','1ED2:0205,1ED2:02AB,1ED2:03E9,1629:0005,165F:034F,1DD6:0152,1DD6:030F,181B:0385,3CD5:000C',
        '-save-state',f'489999999:{prefix}/command.state',
        '-dump-mem-at',f'489999999:lin:0:1048576:{prefix}/end-mem.bin'])+'\n')
    out.with_name(name+'-plan.json').write_text(json.dumps({'name':name,'parent_state_sha256':parent_hash,
        'classification':'controlled entry diagnostic; header dismissal only',
        'injections':[],'inherited_injections':'combat79-r62-gate-near-human-a-plan.json',
        'seed':0x12345678,'input':[{'step':486100000,'x':240,'y':230,'hold_steps':100000}]},ensure_ascii=False,indent=2)+'\n')
    print(name,'固定點擊關閉戰鬥命令標頭')
    sys.exit(0)
match=re.fullmatch(r'combat79-r62-gate-([a-z0-9-]+)-([ab])',name)
assert match
kind,side=match.groups()
parent=Path('workplace/oracle/combat79-r62-natural-a')
assert hashlib.sha256((parent/'entry.state').read_bytes()).hexdigest()=='c4147b9d310c84824d03d43407cf3b335b950e1aacc52ee79dc37a7563b494cb'
mem=(parent/'entry-mem.bin').read_bytes()
assert hashlib.sha256(mem).hexdigest()=='fbe061f125a0554bb6513612c45cba422d26441338e7b431c6b5316473bf6ec1'
ids=[12,16,17,28,29,30,31]
cases={
    'none':{},'action1':{'action':1},'action11':{'action':0x11},'action21':{'action':0x21},
    'actionf1':{'action':0xf1},'action0':{'action':0},'action4':{'action':4},'actionff':{'action':0xff},
    'morale1':{'action':1,'morale':1},'morale2':{'action':1,'morale':2},
    'invalid':{'action':1,'admiral':0xff},
    'two':{'others':[16]},
    'reverse-faction':{'others':[16],'order':[16,12,17,28,29,30,31]},
    'reverse-number':{'others':[16,17],'order':[12,17,16,28,29,30,31]},
    'near-outs':{'action':1,'near':True},
    'near-human':{'action':1,'near':True,'human':True},
}
assert kind in cases
rows=[]
step=484357728


def poke(addr,data,why):
    rows.append({'linear_hex':hex(addr),'bytes_hex':bytes(data).hex(),'step':step,'reason':why})


poke(0x5e00c,struct.pack('<I',0x12345678),'第④步唯一入口固定測試seed')
for index in ids:
    off=0x60ba4+index*122
    record=mem[off:off+122]
    poke(off+0x48,[4],'只由本案例指定艦隊參戰，其餘待命')
    pools=[struct.unpack_from('<H',record,0x10+j*4)[0]for j in range(record[15])]
    scout=struct.unpack_from('<H',record,0x38)[0]
    for pool in pools:
        poke(0x668b2+pool*78+0x1e,[0xff],'門檻診斷不進交戰UI；原版嚴格要求一般分隊Out byte為7F')
        poke(0x668b2+pool*78+0x2e,[0x7d],'觀測第④步是否清這一byte，不推定語意')
    poke(0x668b2+scout*78+0x2e,[0x7d],'對照：偵察池是否在初始化清單')
c=cases[kind]
if 'action' in c:
    poke(0x60ba4+12*122+0x48,[c['action']],'艦隊12行動門檻')
if 'morale' in c:
    poke(0x60ba4+12*122,[c['morale']],'艦隊12士氣旗標門檻')
if 'admiral' in c:
    poke(0x60ba4+12*122+4,[c['admiral']],'艦隊12有效提督門檻')
for index in c.get('others',[]):
    poke(0x60ba4+index*122+0x48,[1],'多艦隊排序診斷')
if kind in ('two','reverse-faction'):
    poke(0x60ba4+12*122+0x48,[1],'多艦隊排序診斷')
if 'order' in c:
    poke(0x5e3c8,c['order'],'唯一入口參戰原始名冊次序')
if c.get('near'):
    attacker=0x668b2+80*78
    target=0x668b2+102*78
    x,y=struct.unpack_from('<HH',mem,attacker+0x34)
    poke(target+0x34,struct.pack('<HH',x+100,y),'唯一入口敵方102在分隊80右方100座標單位，完整記錄接敵診斷')
    poke(target+0x17,[2],'敵方102情報Detail，避免無情報候選被過濾')
    poke(target+0x1e,[0x7f],'敵方102原版一般分隊Out byte')
    poke(0x60ba4+12*122+0x4a,struct.pack('<H',0),'唯一入口艦隊12朝向0；角度契約仍待審查')
    if c.get('human'):
        poke(attacker+0x1e,[0x7f],'分隊80可進人類命令；其餘自身一般分隊仍Out=FF')
expected=[i for i in c.get('order',ids)
          if (i==12 and c.get('action',1 if kind in ('two','reverse-faction') else 4)&15==1
              and not c.get('morale',0)&1 and c.get('admiral',mem[0x60ba4+12*122+4])!=255)
          or i in c.get('others',[])]
out=Path('workplace')/(name+'-flags.txt')
assert not out.exists() and out.parent.stat().st_uid==os.getuid()
prefix='/out/oracle/'+name
flags=['-poke',';'.join(f'lin:{r["linear_hex"][2:]}@{step}='+' '.join(f'{b:02X}'for b in bytes.fromhex(r['bytes_hex']))for r in rows),
       '-dump-mem-at',f'{step+1}:lin:0:1048576:{prefix}/input-mem.bin;{485999999 if c.get("human") else 484699999}:lin:0:1048576:{prefix}/end-mem.bin']
if c.get('human'):
    flags += ['-save-state',f'485999999:{prefix}/prompt.state']
out.write_text('\n'.join(flags)+'\n')
plan=out.with_name(name+'-plan.json')
assert not plan.exists()
plan.write_text(json.dumps({'name':name,'kind':kind,'side':side,'parent_state_sha256':'c4147b9d310c84824d03d43407cf3b335b950e1aacc52ee79dc37a7563b494cb',
                           'classification':'controlled entry diagnostic; no normal player or whole-phase claim',
                           'injections':rows,'expected_collected_before_sort':expected,'sort_order':'unknown; derive from actual 1ED2 calls',
                           'required_counts':{'1DD6:0009':int(bool(expected)),'1ED2:0001':len(expected)}},ensure_ascii=False,indent=2)+'\n')
print(name,'預定參戰名單',expected,'注入',len(rows))
