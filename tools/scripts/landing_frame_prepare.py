"""第109輪可丟棄單幀輸入；只從原版欄位及呼叫記錄取typed資料。"""
import base64
import hashlib
import json
import os
from pathlib import Path
import struct
import sys

assert Path.cwd() == Path('/src') and os.getuid() == 1000
parent = json.loads(Path('docs/internals/data/landing-scene.json').read_text())
sources = {}


def read(path):
    path = Path(path)
    raw = path.read_bytes()
    identity = {'bytes': len(raw), 'sha256': hashlib.sha256(raw).hexdigest()}
    if str(path) in parent['sources']:
        assert identity == parent['sources'][str(path)], str(path)
    sources[str(path)] = identity
    return raw


def word(m, a):
    return struct.unpack_from('<H', m, a)[0]


def ptr(m, a):
    return word(m, a) + 16 * word(m, a + 2)


def text(m, a):
    return m[a:a+256].split(b'\0', 1)[0]


def b64(b):
    return base64.b64encode(b).decode()


ds, fleet, planet = 0x54b30, 0x61a6a, 0x54560
cases = []
for kind in ('failure', 'capture', 'counterfirst', 'empire', 'multiple'):
    revision = 'r2' if kind == 'multiple' else 'r1'
    stem = f'landing48-r108-r5-{kind}' if kind in ('failure','capture') else f'landing48-r109-branch-{revision}-{kind}'
    a = Path('workplace/oracle')/(stem+'-a')
    b = Path('workplace/oracle')/(stem+'-b')
    report = json.loads(read(a/'report.json'))
    baseline = read(a/'phase6-input-mem.bin')
    calls = [e for e in report['events'] if e['original_location'] == '3958:0041']
    returns = [e for e in report['frames'] if e['name'].startswith('comani-return')]
    assert len(calls) == len(returns)
    names = ['frame-20b4-078f-01', 'frame-20b4-07cb-01', 'frame-20b4-089c-01']
    if kind == 'capture':
        names.append('frame-3765-006c-02')
    if kind == 'multiple':
        names.extend(['frame-20b4-07cb-02','frame-20b4-089c-02'])
    for name in names:
        frame = next(f for f in report['frames'] if f['name'] == name)
        actor = [e for e in report['events'] if e['original_location'] == '20B4:0503' and e['step'] < frame['step']][-1]
        fleet = actor['stack_words'][2] + 16*actor['stack_words'][3]
        landing = [e for e in report['events'] if e['original_location'] == '20B4:0711' and e['step'] < frame['step']][-1]
        attack, damage = landing['stack_words'][4], landing['stack_words'][1]
        if kind in ('failure','capture'):
            assert attack == 844 and damage == (518 if kind == 'failure' else 1)
        elif kind == 'counterfirst':
            assert (attack,damage) == (832,410)
        mem, native = read(a/(name+'-mem.bin')), read(a/(name+'-screen.bin'))
        assert mem == read(b/(name+'-mem.bin')) and native == read(b/(name+'-screen.bin'))
        read(a/(name+'-palette.json'))
        ship = ptr(mem, ds+0xcfb4)
        units, expected = [], []
        for i in range(mem[fleet+15]):
            idx = word(mem, fleet+16+i*4)
            s = ship+idx*78
            units.append(dict(ID=idx, Side=mem[s], X=word(mem,s+52), Y=word(mem,s+54),
                              Squadron=mem[s+3], Detail=mem[s+23], Ships=word(mem,s+8)))
            expected.append(dict(ID=idx, X=word(mem,s+62), Y=word(mem,s+64),
                                 Visible=bool(mem[s+66]), Depth=struct.unpack_from('<h',mem,s+68)[0]))
        side = mem[fleet+1]
        assert len(units) == 7 and all(u['Side'] == side for u in units)
        lower = any(e['original_location'] == '321F:0897' and e['step'] < frame['step'] for e in report['events'])
        after = name.startswith('frame-20b4-089c-') or name == 'frame-3765-006c-02'
        captured = name == 'frame-3765-006c-02'
        owner = baseline[planet+17]
        visibility = dict(Mode=mem[ds+0x8f20], PlayerSide=mem[ds+0x8f25],
                          SideControl=list(mem[ds+0x8f21:ds+0x8f23]))
        control = visibility['SideControl'][0 if side == visibility['PlayerSide'] else 1]
        controlled = visibility['Mode'] != 2 and control == 0
        known = controlled or mem[ship+word(mem,fleet+16)*78+23] == 3
        pictures = []
        for e, ret in zip(calls, returns, strict=True):
            assert e['step'] < ret['step']
            if ret['step'] > frame['step']:
                continue
            r = bytes.fromhex(e['comani_record_hex'])
            static, animate = e['stack_words'][4:6]
            if r[1] in (9, 10, 11):
                prefix = bytes([ord('P'), mem[ds+0x8e1a+r[1]]])
            else:
                second = mem[ds+0x8e1a+r[1]] if r[4] != 0 else mem[ds+0x8e34+r[5]]
                prefix = bytes([mem[ds+0x8e16+r[0]], second])
            if static:
                suffix = b'D0'
            else:
                fourth = mem[ds+0x8e2c+r[3]+(3 if r[2] == 0 else 0)]
                if r[2] == 0 and r[1] in (10, 11) and r[3] == 1:
                    fourth = b'MF'[r[0]]
                suffix = bytes([mem[ds+0x8e28+r[2]], fourth])
            pictures.append(dict(Side=r[0], SegmentHex=(prefix+suffix).hex(), AllFrames=bool(animate),
                                 OriginalRecordHex=e['comani_record_hex'], Static=bool(static)))
        view = dict(Faction=owner, PlanetName=b64(text(mem,planet)),
                    Defence=word(mem,planet+20), DefenceMax=word(mem,planet+22),
                    FleetKnown=known, FleetNumber=mem[fleet+2],
                    AdmiralName=b64(text(mem,ptr(mem,ds+0x6428+4*mem[fleet+4]))),
                    Ships=attack, Label=2 if captured else 1,
                    Destroyed=damage, ShowResult=after and not captured,
                    Gauges=dict(FleetFaction=side, Integrity=mem[fleet+60],
                                Defence=word(mem,planet+20), Attack=attack))
        formats = {hex(x): text(mem,ds+x).hex() for x in
                   (0x483f,0x4882,0x4886,0x488c,0x485b,0x48b1,0x48d1)}
        formats['fleet_prefix'] = ('0c'+text(mem,ds+(0x4830 if controlled else 0x481f)).hex())
        formats['fleet_suffix'] = text(mem,ds+(0x4833 if controlled else 0x4822)).hex()
        formats['fleet_name'] = text(mem,ds+(0x4838 if controlled else 0x4829)).hex()
        order = []
        if lower:
            for i in range(len(units)+1):
                v = ptr(mem,ds+0xd78e+i*4)
                assert v == 0 or (v-ship) % 78 == 0
                order.append(-1 if v == 0 else (v-ship)//78)
        cases.append(dict(Name=kind+'-'+name, Frame=frame, View=view, Formats=formats,
                          FactionText=b64(text(mem,ptr(mem,ds+0x60ae+owner*4))),
                          PlanetKindText=b64(text(mem,ptr(mem,ds+0x619c+mem[planet+18]*4))),
                          Units=units, Visibility=visibility, FleetSide=side,
                          FleetColour=mem[ds+0x4422+side], PlanetColour=mem[ds+0x4422+owner],
                          Facing=mem[fleet+74], Lower=lower,
                          PlanetPoint=dict(X=word(mem,planet+70),Y=word(mem,planet+72)),
                          ExpectedUnits=expected, ExpectedPlanet=[word(mem,planet+76),word(mem,planet+78)],
                          ExpectedOrder=order, Pictures=pictures,
                          NativePath=str(a/(name+'-screen.bin')), NativeSHA256=hashlib.sha256(native).hexdigest()))

for segment in ('20B4','321F'):
    j = json.loads(read(f'workplace/ida/landing48-r109-{segment}.json'))
    assert j['kernel_version'] == '9.4' and j['uid'] == 1000
    raw = read(f'workplace/ida/landing48-r109-{segment}.com')
    assert hashlib.sha256(raw).hexdigest() == j['input_sha256']
    mem = read('workplace/oracle/landing48-r108-r5-failure-a/frame-20b4-078f-01-mem.bin')
    base = int(segment,16)*16
    assert raw[256:] == mem[base:base+65536]
    for r in j['instructions']:
        loc = int(r['runtime_linear_hex'],16)
        raw = bytes.fromhex(r['original_bytes_hex'])
        assert mem[loc:loc+len(raw)] == raw
result = dict(Schema='logh3.landing-frame-probe.v1', Cases=cases, Sources=sources,
              ParentSHA256=hashlib.sha256(Path('docs/internals/data/landing-scene.json').read_bytes()).hexdigest())
dest = Path(sys.argv[1] if len(sys.argv) == 2 else 'workplace/landing48-r109-typed-r4.json')
with dest.open('x') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write('\n')
print('typed cases',len(cases),'sources',len(sources))
