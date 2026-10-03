"""核雙跑及IDA原始bytes，凍結文字、色盤與診斷畫面；拒絕異內容覆寫。"""
import base64
import hashlib
import json
import struct
import zlib
from pathlib import Path

root = Path('/work')
sources = {}


def read(p):
    b = p.read_bytes()
    assert (p.stat().st_uid, p.stat().st_gid) == (1000, 1000)
    sources[str(p.relative_to(root))] = hashlib.sha256(b).hexdigest()
    return b


def pack(b):
    return base64.b64encode(zlib.compress(b, 9)).decode()


a, b = [root / f'workplace/oracle/presentation48-r71-{x}' for x in ('a', 'b')]
ra, rb = [json.loads(read(p/'report.json')) for p in (a, b)]
ca, cb = dict(ra), dict(rb)
ca.pop('command')
cb.pop('command')
assert ca == cb
assert len(ra['cases']) == 92
for name in sorted(p.name for p in a.iterdir() if p.name != 'report.json'):
    assert read(a/name) == read(b/name), name

parent = root/'workplace/oracle/blocks48-r70-a/scene-01-01-entry-mem.bin'
mem = read(parent)
ds = 0x54b30


def word(m, a):
    return struct.unpack_from('<H', m, a)[0]


def ptr(m, a):
    return word(m, a)+16*word(m, a+2)


def string(m, a):
    return m[a:a+256].split(b'\0', 1)[0].hex()


windows = []
for i in range(3):
    w = ptr(mem, ds+0xa44a+4*i)
    windows.append(dict(x=mem[w+5], y=mem[w+6], w=mem[w+7], h=mem[w+8], colour=mem[w+1]))
formats = {key: string(mem, ds+addr) for key, addr in (
    ('clear', 0x442e), ('unknown', 0x4430), ('prefix', 0x4461),
    ('suffix', 0x4464), ('squadron', 0x446b), ('admiral', 0x4477),
    ('ships', 0x447e), ('weapon', 0x4495), ('goto', 0x44a3), ('destroyed', 0x44ac))}
assert string(mem, ds+0x4444) == formats['prefix']
assert string(mem, ds+0x4447) == formats['suffix']
assert string(mem, ds+0x444e) == formats['squadron']
assert string(mem, ds+0x445a) == formats['admiral']
palette = json.loads(read(a/'case-00-palette.json'))
assert all(v == ((v>>2)<<2 | (v>>2)>>4) for rgb in palette for v in rgb)
spec = dict(schema='logh3.fleet-presentation-spec.v1', windows=windows,
            formats=formats, weapon_names=[string(mem, ptr(mem, ds+0x50a8+4*i)) for i in range(13)],
            palette=[[v>>2 for v in rgb] for rgb in palette])
pool = ptr(mem, ds+0xcfb4)
# 082F:0875/0877取DS:C074表本體，並非遠指標。
fleet_base = ds+0xc074
cases = []
for c in ra['cases']:
    m = read(a/(c['name']+'-mem.bin'))
    ship = pool+c['params'][0]*78
    fleet = fleet_base+m[ship+1]*122
    kind = ds+0x4e3a+m[ship]*0xd0+m[ship+4]*0x1a
    w = ptr(m, ds+0xa44a+4*(m[ship] if c['params'][3] < 0 else 2))
    content = ptr(m, w+0xe)
    n = m[w+7]*m[w+8]
    cells = m[content:content+4*n]
    inp = dict(side=m[ship], squadron=m[ship+3], fleet_number=word(m, fleet+2),
               admiral_name=string(m, ptr(m, ds+0x6428+4*m[fleet+4])),
               kind_name=string(m, ptr(m, ds+0x613a+4*m[ship+4])),
               ships=word(m, ship+8), ships_max=word(m, ship+10), detail=m[ship+0x17],
               mode=m[ds+0x8f20], side_control=list(m[ds+0x8f21:ds+0x8f23]),
               player_side=m[ds+0x8f25], weapon=m[ship+0x2e],
               weapon_types=[m[kind+x] for x in (10,14,18)], destroyed=c['params'][4])
    cases.append(dict(metadata=c, input=inp, native_cells_hex=cells.hex(),
                      native_cursor=list(m[w+0x1c:w+0x1e]), native_colour=m[w+1],
                      native_zlib_base64=pack(read(a/(c['name']+'-screen.bin')))))
exports = []
for filename in ('fleet48-r71-1dd6', 'fleet48-r71-3982', 'fleet48-r71-3982-body', 'fleet48-r71-3982-number', 'scene48-r69-1DD6'):
    p = root/f'workplace/ida/{filename}.json'
    e = json.loads(read(p))
    binary = read(p.with_suffix('.com'))
    assert hashlib.sha256(binary).hexdigest() == e['input_sha256']
    seg = int(e['runtime_segment'], 16)
    assert binary[256:] == mem[seg*16:seg*16+65536]
    for row in e['instructions']:
        address = int(row['runtime_linear_hex'], 16)
        raw = bytes.fromhex(row['original_bytes_hex'])
        assert mem[address:address+len(raw)] == raw
    exports.append(e)
for p in ('tools/fleet_presentation_oracle.sh', 'tools/scripts/fleet_presentation_native.go',
          'tools/scripts/fleet_presentation_ida.py', 'tools/scripts/fleet_presentation_receipt.py'):
    read(root/p)
planet_pal = read(root/'workplace/oracle/eng-open/pal.bin')
assert list(planet_pal[:48]) == [v for rgb in palette for v in rgb]
asset_inputs = {name: hashlib.sha256((Path('/orig')/name).read_bytes()).hexdigest()
                for name in ('GRAPH.IMG', 'ANI_PNL1.A98', 'COMANI.BMG')}
result = dict(schema='logh3.fleet-presentation-receipt.v1', report=ra, spec=spec, asset_inputs=asset_inputs,
              base_zlib_base64=pack(read(root/'workplace/oracle/blocks48-r70-a/scene-01-01-entry-screen.bin')),
              logical_palette_rgb8=palette, cases=cases, ida_exports=exports, sources=sources,
              classification='same-state direct-entry presentation; other layers borrowed from native entry')
for p, obj in ((root/'docs/internals/data/fleet-presentation-spec.json', spec),
               (root/'docs/internals/data/fleet-presentation.json', result)):
    payload=(json.dumps(obj, ensure_ascii=False, indent=2)+'\n').encode()
    if p.exists():
        assert p.read_bytes() == payload, p
    else:
        p.write_bytes(payload)
    print(p, len(payload), hashlib.sha256(payload).hexdigest())
print('paired cases',len(cases),'paired outputs',276,'sources',len(sources))
