"""#79 子步8-4：從固定父RAM建立一次性IDA輸入。"""
from pathlib import Path
import hashlib

for p in [Path('tools/scripts'), Path('workplace/ida')]:
    assert (p.stat().st_uid, p.stat().st_gid) == (1000, 1000)
src = Path('tools/scripts/scout_fuel_ida.py').read_text()
src = src.replace('scout79-r83-', 'promotion79-r84-')
src = src.replace('logh3.scout-fuel-ida.v1', 'logh3.battle-promotion-ida.v1')
src = src.replace('#79 子步8-3', '#79 子步8-4')
src = src.replace("start, end = (int(x, 16) for x in part.split(':'))",
                  "left, right = part.split(':')\n    start = int(left,16)\n"
                  "    until_ret = right == 'ret'\n"
                  "    end = len(raw) if until_ret else int(right,16)")
src = src.replace('ea += size\n    ida_funcs.add_func',
                  "ea += size\n        if until_ret and idc.print_insn_mnem(ea-size) in ('retf', 'ret'):\n"
                  "            end = ea-base\n            break\n    ida_funcs.add_func")
with Path('tools/scripts/battle_promotion_ida.py').open('x') as f:
    f.write(src)
raw = Path('workplace/oracle/ownership79-r82-entry-a/mem-in.bin').read_bytes()
assert hashlib.sha256(raw).hexdigest() == 'b905eda112fd3a78292bc0d1cdc10068703e0291968640cbaa0d150404404900'
for seg in [0x185C, 0x28E0, 0x297B]:
    p = Path(f'workplace/ida/promotion79-r84-{seg:04X}.com')
    with p.open('xb') as f:
        f.write(bytes(256) + raw[seg*16:seg*16+4096])
    print(p, hashlib.sha256(p.read_bytes()).hexdigest())
