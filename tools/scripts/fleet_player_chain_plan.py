"""第76輪名冊順位的有限排列。只用固定父state，拒絕覆寫。"""
import hashlib
import json
from pathlib import Path
import re
import struct
import sys

name = sys.argv[1]
m = re.fullmatch(r'fleet79-r76-order-(base|reverse|rotate|interleave|number)-([ab])', name)
assert m
case, side = m.groups()
parent = Path('workplace/oracle/combat79-r62-natural-a')
state_sha = 'c4147b9d310c84824d03d43407cf3b335b950e1aacc52ee79dc37a7563b494cb'
ram_sha = 'fbe061f125a0554bb6513612c45cba422d26441338e7b431c6b5316473bf6ec1'
assert hashlib.sha256((parent/'entry.state').read_bytes()).hexdigest() == state_sha
ram = (parent/'entry-mem.bin').read_bytes()
assert hashlib.sha256(ram).hexdigest() == ram_sha
ids = [12,16,17,28,29,30,31]
orders = {'base':ids, 'reverse':ids[::-1], 'rotate':[17,28,29,30,31,12,16],
          'interleave':[31,30,12,29,28,17,16], 'number':[12,31,30,29,28,17,16]}
step = 484357728
rows = []
def poke(addr, data, reason):
    rows.append({'linear_hex':hex(addr),'bytes_hex':bytes(data).hex(),'step':step,'reason':reason})
poke(0x5e00c, struct.pack('<I',0x12345678), '執行前固定seed，不重擲')
poke(0x5e3c8, orders[case], '固定完整原名冊排列')
for index in ids:
    base = 0x60ba4 + index*122
    poke(base+0x48,[1], '七艦隊均收進命令名冊')
    r = ram[base:base+122]
    for j in range(r[15]):
        pool = struct.unpack_from('<H',r,0x10+4*j)[0]
        poke(0x668b2+pool*78+0x1e,[255], '順位診斷一般池不進等待UI')
prefix = '/out/oracle/'+name
flags = ['-load-state','/out/oracle/combat79-r62-natural-a/entry.state',
         '-steps','484700000','-tick','20000','-regs-from','484357727','-regs-max','256',
         '-regs-at','1DD6:0040,1DD6:0067,1ED2:0001,1DD6:0152,181B:0385,3CD5:000C',
         '-poke',';'.join('lin:'+r['linear_hex'][2:]+'@'+str(step)+'='+' '.join(f'{b:02X}' for b in bytes.fromhex(r['bytes_hex'])) for r in rows),
         '-dump-mem-at',f'{step+1}:lin:0:1048576:{prefix}/input-mem.bin;484699999:lin:0:1048576:{prefix}/end-mem.bin']
for suffix, data in [('flags.txt','\n'.join(flags)+'\n'),('plan.json',json.dumps({
    'name':name,'case':case,'side':side,'classification':'controlled roster diagnostic',
    'parent_state_sha256':state_sha,'parent_ram_sha256':ram_sha,'seed':0x12345678,
    'roster':orders[case],'injections':rows},ensure_ascii=False,indent=2)+'\n')]:
    p = Path('workplace')/(name+'-'+suffix)
    assert not p.exists()
    p.write_text(data)
print(name, orders[case])
