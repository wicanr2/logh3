"""固定原版DS字串寫入118正式畫面規格；所有來源身份另由收據核對。"""
import hashlib
import json
import os
from pathlib import Path
import struct
import sys

assert Path.cwd() == Path('/src') and os.getuid() == 1000
path = Path('workplace/oracle/landing48-r108-r5-failure-a/frame-20b4-078f-01-mem.bin')
m = path.read_bytes()
assert hashlib.sha256(m).hexdigest() == '3df4f9a27014ca83a6194df9e6cc248eda3b945fcb9a68a9043012f2df140274'
ds = 0x54b30


def text(a):
    return m[a:a+256].split(b'\0',1)[0].hex()


def at(a):
    return text(ds+a)


def far(a):
    off, seg = struct.unpack_from('<HH',m,ds+a)
    return text(off+16*seg)


p = Path('docs/internals/data/engagement-panel.json')
j = json.loads(p.read_text())
j['planet_frame'] = dict(Factions=[far(0x60ae+4*i) for i in range(3)],
    Kinds=[far(0x619c+4*i) for i in range(3)],Colours=list(m[ds+0x4422:ds+0x4425]),
    Attack=at(0x483f),Ships=at(0x4867),Unknown=at(0x4801),
    ControlledPrefix='0c'+at(0x4830),ControlledSuffix=at(0x4833),ControlledName=at(0x4838),
    KnownPrefix='0c'+at(0x481f),KnownSuffix=at(0x4822),KnownName=at(0x4829),
    Title=at(0x4882),Name=at(0x4886),Defence=at(0x488c),
    Labels=[at(x) for x in (0x48a7,0x485b,0x48d1,0x48db)],Destroyed=at(0x48b1))
if len(sys.argv) == 3 and sys.argv[1] == '--output':
    p = Path(sys.argv[2])
    assert str(p).startswith('workplace/') and not p.exists()
elif len(sys.argv) != 1:
    raise SystemExit('formats [--output workplace/<new-file>.json]')
p.write_text(json.dumps(j,ensure_ascii=False,indent=1)+'\n')
print('118 formats generated',len(j['planet_frame']))
