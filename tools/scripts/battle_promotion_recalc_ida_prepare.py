"""#79 第85輪重算 helper：固定RAM切段，拒絕覆寫。"""
from pathlib import Path
import hashlib

memory = Path('workplace/oracle/ownership79-r82-entry-a/mem-in.bin').read_bytes()
assert hashlib.sha256(memory).hexdigest() == 'b905eda112fd3a78292bc0d1cdc10068703e0291968640cbaa0d150404404900'
for name, segment, size in [('297B',0x297B,0x1400),('090A',0x090A,0x2000)]:
    raw = bytes(256) + memory[segment*16:segment*16+size]
    p = Path(f'workplace/ida/promotion79-r85-reviewed-{name}.com')
    with p.open('xb') as f:
        f.write(raw)
    print(p, len(raw), hashlib.sha256(raw).hexdigest())
