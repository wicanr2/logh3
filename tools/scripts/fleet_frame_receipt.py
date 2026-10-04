"""凍結完整交戰兩幀的typed來源；不借用原版已畫好的上方。"""
import base64
import hashlib
import json
import struct
from pathlib import Path

root = Path('/src')
sources = {}


def read(path):
    raw = path.read_bytes()
    assert (path.stat().st_uid, path.stat().st_gid) == (1000, 1000)
    sources[str(path.relative_to(root))] = hashlib.sha256(raw).hexdigest()
    return raw


def word(m, a):
    return struct.unpack_from('<H', m, a)[0]


def ptr(m, a):
    return word(m, a) + 16 * word(m, a + 2)


def text(m, a):
    return m[a:a+256].split(b'\0', 1)[0]


lower = json.loads(read(root/'docs/internals/data/fleet-scene-plan.json'))
assert sources['docs/internals/data/fleet-scene-plan.json'] == '4988a5b95d2167bc14166a8fb173fb26bc0b6d4b9f6e7f070c86c9bcc3b126f4'
scene = json.loads(read(root/'docs/internals/data/fleet-scene.json'))
presentation = json.loads(read(root/'docs/internals/data/fleet-presentation.json'))
records = scene['pairs'][0]['comani_records']
assert len(records) == 4
assert records[:2] == [dict(step=496556919, record=[0,0,0,0,0,12,88], static=1, animate=0),
                      dict(step=497395564, record=[1,0,1,0,0,40,88], static=1, animate=0)]
ds = 0x54b30
cases = []
for i, c in enumerate(lower['cases']):
    filename = ('scene-01-20-complete-mem.bin', 'scene-02-40-complete-mem.bin')[i]
    a = read(root/f'workplace/oracle/blocks48-r70-a/{filename}')
    b = read(root/f'workplace/oracle/blocks48-r70-b/{filename}')
    assert a == b and len(a) == 1048576
    visibility = dict(Mode=a[ds+0x8f20], PlayerSide=a[ds+0x8f25], SideControl=list(a[ds+0x8f21:ds+0x8f23]))
    units, integrity = [], []
    for idx in (80, 102):
        sh = ptr(a, ds+0xcfb4) + idx*78
        fleet = ds+0xc074+a[sh+1]*122
        units.append(dict(Side=a[sh], Squadron=a[sh+3], Detail=a[sh+0x17],
                          FleetNumber=word(a, fleet+2), Ships=word(a, sh+8), ShipsMax=word(a, sh+10),
                          AdmiralName=base64.b64encode(text(a,ptr(a,ds+0x6428+4*a[fleet+4]))).decode(),
                          KindName=base64.b64encode(text(a,ptr(a,ds+0x613a+4*a[sh+4]))).decode(),
                          Visibility=visibility))
        integrity.append(a[fleet+0x3c])
    attacker = ptr(a, ds+0xcfb4)+80*78
    kind = ds+0x4e3a+a[attacker]*208+a[attacker+4]*26
    pictures = []
    for record in records[:2 if i == 0 else 4]:
        r = record['record']
        assert r[4] == 0
        segment = bytes([b'EF'[r[0]], a[ds+0x8e34+r[5]], ord('D') if record['static'] else b'AD'[r[2]],
                         ord('0') if record['static'] else b'BMF'[r[3]] if r[2] == 0 else ord('0')+r[3]])
        pictures.append(dict(Side=r[0], SegmentHex=segment.hex(), AllFrames=bool(record['animate'])))
    typed = dict(Scene=dict(Sides=c['sides'], AttackerID=80, TargetID=102, Visibility=visibility, Facing=c['facing']),
                 Units=units, Integrity=integrity,
                 Weapon=dict(Side=a[attacker], Weapon=a[attacker+0x2e],
                             WeaponTypes=[a[kind+x] for x in (10,14,18)], ShowResult=bool(i), Destroyed=726 if i else 0))
    assert typed['Units'][0]['Ships'] == 3000
    assert typed['Units'][1]['Ships'] == (3000 if i == 0 else 2274)
    cases.append(dict(name=c['name'], view=typed, pictures=pictures,
                      native_zlib_base64=c['complete_native_zlib_base64']))

ram = read(root/'workplace/oracle/projection48-r72-a/case-00-before-mem.bin')
ida = []
for segment in ('1DD6', '3958', '348A'):
    path = root/f'workplace/ida/opening48-r73-{segment}.json'
    export = json.loads(read(path))
    raw = read(path.with_suffix('.com'))
    assert hashlib.sha256(raw).hexdigest() == export['input_sha256']
    assert raw[256:] == ram[int(segment,16)*16:int(segment,16)*16+len(raw)-256]
    for row in export['instructions']:
        address = int(row['runtime_linear_hex'],16)
        data = bytes.fromhex(row['original_bytes_hex'])
        assert ram[address:address+len(data)] == data
    ida.append(export)
capacity = json.loads(read(root/'docs/internals/data/ship-kind-table.json'))
full_ships = []
for side in capacity['factions']:
    values = []
    for k in side['kinds']:
        address = ds+0x4e3a+side['faction']*208+k['kind']*26
        assert word(ram,address) == k['full_ships'] == k['words'][0]
        values.append(k['full_ships'])
    full_ships.append(values)
assert full_ships == [[3000,3000,4000,2000,3000,3000,1000,1]]*2
read(root/'tools/scripts/fleet_frame_receipt.py')
read(root/'tools/scripts/engagement_opening_ida.py')
out = dict(schema='logh3.fleet-frame.v1', date='2026-10-04',
           dosgolem_head=lower['dosgolem_head'], sources=sources,
           classification='same-state complete presentation from typed sources; fixed original diagnostic; not Game player path',
           assets=presentation['asset_inputs'], sprites=lower['sprites'], cases=cases,
           full_ships=full_ships, ida_exports=ida)
payload = (json.dumps(out,ensure_ascii=False,indent=2)+'\n').encode()
dest = root/'docs/internals/data/fleet-frame.json'
if dest.exists():
    assert dest.read_bytes() == payload, '拒絕改寫凍結收據'
else:
    dest.write_bytes(payload)
print(len(sources), 'sources', len(payload), 'bytes', hashlib.sha256(payload).hexdigest())
