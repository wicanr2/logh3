"""第135輪：沿118已驗寫入者快照，補PF／PZ登陸完整幀。"""
import base64
import hashlib
import json
import os
from pathlib import Path
import struct
import sys

assert Path.cwd() == Path('/src') and os.getuid() == 1000
ds, planet = 0x54b30, 0x54560
sources = {}


def read(path):
    p = Path(path)
    b = p.read_bytes()
    sources[str(p)] = dict(bytes=len(b), sha256=hashlib.sha256(b).hexdigest())
    return b


def word(m, a):
    return struct.unpack_from('<H', m, a)[0]


def ptr(m, a):
    return word(m, a) + 16 * word(m, a+2)


def text(m, a):
    return m[a:a+256].split(b'\0', 1)[0]


def b64(b):
    return base64.b64encode(b).decode()


cases = []
for kind in ('fortress-failure', 'fortress-capture', 'iserlohn-failure', 'iserlohn-capture'):
    stem = f'landing48-r135-{kind}'
    a = Path('workplace/oracle')/(stem+'-a')
    b = Path('workplace/oracle')/(stem+'-b')
    report = json.loads(read(a/'report.json'))
    frames = report['frames']
    events = report['events']
    cache = {}

    def memory(frame):
        name = frame['name']
        if name not in cache:
            m = read(a/(name+'-mem.bin'))
            assert m == read(b/(name+'-mem.bin')), name
            cache[name] = m
        return cache[name]

    def after(event):
        # 此幀是寫入者事件之後的第一份完整RAM；各欄位沒有跨越下一次修改。
        frame = next(f for f in frames if f['step'] >= event['step'])
        return frame, memory(frame)

    def latest(location, frame, predicate=lambda e: True):
        found = [e for e in events if e['original_location'] == location
                 and e['step'] < frame['step'] and predicate(e)]
        return found[-1] if found else None

    def lower(event):
        frame, m = after(event)
        fleet = event['stack_words'][2] + 16*event['stack_words'][3]
        ship = ptr(m, ds+0xcfb4)
        units, expected = [], []
        for i in range(m[fleet+15]):
            idx = word(m, fleet+16+i*4)
            s = ship+idx*78
            units.append(dict(ID=idx, Side=m[s], X=word(m,s+52), Y=word(m,s+54),
                              Squadron=m[s+3], Detail=m[s+23], Ships=word(m,s+8)))
            expected.append(dict(ID=idx, X=word(m,s+62), Y=word(m,s+64),
                                 Visible=bool(m[s+66]), Depth=struct.unpack_from('<h',m,s+68)[0]))
        order = []
        for i in range(len(units)+1):
            v = ptr(m,ds+0xd78e+i*4)
            assert v == 0 or (v-ship) % 78 == 0
            order.append(-1 if v == 0 else (v-ship)//78)
        return dict(Units=units, Visibility=dict(Mode=m[ds+0x8f20],PlayerSide=m[ds+0x8f25],
                    SideControl=list(m[ds+0x8f21:ds+0x8f23])), Facing=m[fleet+74],
                    View=dict(Faction=m[planet+17],Defence=word(m,planet+20)),
                    PlanetPoint=dict(X=word(m,planet+70),Y=word(m,planet+72)),
                    ExpectedUnits=expected,ExpectedPlanet=[word(m,planet+76),word(m,planet+78)],
                    ExpectedOrder=order, SnapshotFrame=frame['name'], Writer=event)

    calls = [e for e in events if e['original_location'] == '3958:0041']
    returns = [f for f in frames if f['name'].startswith('comani-return')]
    assert len(calls) == len(returns)
    stop = next(f['step'] for f in frames if f['name'] == 'frame-1dd6-026b-01')
    for frame in frames:
        if not calls[0]['step'] < frame['step'] < stop:
            continue
        m = memory(frame)
        tf = latest('20B4:0A26', frame)
        tp = latest('20B4:0C4E', frame)
        gf = latest('321F:0D11', frame)
        gp = latest('321F:0DF7', frame)
        assert all((tf,tp,gf,gp)), frame['name']
        _, fm = after(tf)
        _, pm = after(tp)
        _, gm = after(gf)
        fleet = tf['stack_words'][2]+16*tf['stack_words'][3]
        ship = ptr(fm,ds+0xcfb4)
        side, owner = fm[fleet+1], pm[planet+17]
        visibility = dict(Mode=fm[ds+0x8f20], PlayerSide=fm[ds+0x8f25],
                          SideControl=list(fm[ds+0x8f21:ds+0x8f23]))
        control = visibility['SideControl'][0 if side == visibility['PlayerSide'] else 1]
        controlled = visibility['Mode'] != 2 and control == 0
        known = controlled or fm[ship+word(fm,fleet+16)*78+23] == 3
        landing = tf['stack_words'][4] & 255
        ships = ships_max = 0
        for i in range(fm[fleet+15]):
            s = ship+word(fm,fleet+16+i*4)*78
            ships += word(fm,s+8)
            ships_max += word(fm,s+10)
        if landing:
            # 0681保留本次攻擊力；DF7收到同一個值。
            ships = gp['stack_words'][6]
        writers = [e for e in events if e['step'] < frame['step'] and (
            e['original_location'] in ('20B4:0E45','20B4:1004','20B4:121F') or
            e['original_location'] == '20B4:0A26' and e['stack_words'][4]&255 or
            e['original_location'] == '20B4:0C4E' and e['stack_words'][6]&255)]
        middle = dict(Label=0, Destroyed=0, ShowResult=False)
        title = next((e for e in reversed(writers) if e['original_location'] not in
                      ('20B4:0E45','20B4:1004')), None)
        assert title
        loc = title['original_location']
        middle['Label'] = 1 if loc == '20B4:0A26' else (2 if title['stack_words'][2] else 3) if loc == '20B4:121F' else 0
        label_colour = m[ds+0x4422+(side if middle['Label']==1 else owner)] if middle['Label']<2 else 4
        result_colour = 4
        last = writers[-1]
        if last['original_location'] in ('20B4:0E45','20B4:1004'):
            middle.update(Destroyed=last['stack_words'][2],ShowResult=True)
            result_colour = m[ds+0x4422+(last['stack_words'][3]&255)]
        pictures = []
        for e, ret in zip(calls,returns,strict=True):
            if e['step'] >= frame['step']:
                continue
            draws = [f for f in frames if e['step'] < f['step'] <= min(frame['step'],ret['step'])
                     and f['name'].startswith('draw-frame-return')]
            if frame['name'].startswith('draw-frame-return') and frame['step'] < ret['step']:
                # 348A:0281計時等待返回；此格尚未寫入framebuffer。
                draws = draws[:-1]
            if not draws:
                continue
            r = bytes.fromhex(e['comani_record_hex'])
            static, animate = e['stack_words'][4:6]
            if r[1] in (9,10,11):
                prefix = bytes([ord('P'), m[ds+0x8e1a+r[1]]])
            else:
                second = m[ds+0x8e1a+r[1]] if r[4] else m[ds+0x8e34+r[5]]
                prefix = bytes([m[ds+0x8e16+r[0]],second])
            if static:
                suffix = b'D0'
            else:
                fourth = m[ds+0x8e2c+r[3]+(3 if r[2]==0 else 0)]
                if r[2]==0 and r[1] in (10,11) and r[3]==1:
                    fourth = b'MF'[r[0]]
                suffix = bytes([m[ds+0x8e28+r[2]],fourth])
            pictures.append(dict(Side=r[0],SegmentHex=(prefix+suffix).hex(),Count=len(draws),
                                 OriginalRecordHex=e['comani_record_hex'],Static=bool(static),Animate=bool(animate)))
        formats = {hex(x): text(m,ds+x).hex() for x in
                   (0x483f,0x4867,0x4882,0x4886,0x488c,0x485b,0x48a7,0x48b1,0x48d1,0x48db)}
        formats.update(fleet_prefix='0c'+text(m,ds+(0x4830 if controlled else 0x481f)).hex(),
                       fleet_suffix=text(m,ds+(0x4833 if controlled else 0x4822)).hex(),
                       fleet_name=text(m,ds+(0x4838 if controlled else 0x4829)).hex())
        _, gpm = after(gp)
        gptr = gp['stack_words'][2]+16*gp['stack_words'][3]
        view = dict(Faction=owner, PlanetName=b64(text(pm,planet)), Defence=word(pm,planet+20),
                    DefenceMax=word(pm,planet+22), FleetKnown=known,FleetNumber=fm[fleet+2],
                    AdmiralName=b64(text(fm,ptr(fm,ds+0x6428+4*fm[fleet+4]))),
                    Ships=ships&65535,ShipsMax=ships_max&65535,**middle,
                    Gauges=dict(FleetFaction=side,Integrity=gm[fleet+60],
                                Defence=word(gpm,gptr+20),Attack=gp['stack_words'][6]))
        le = latest('321F:0897',frame)
        native = read(a/(frame['name']+'-screen.bin'))
        assert native == read(b/(frame['name']+'-screen.bin'))
        read(a/(frame['name']+'-palette.json'))
        cases.append(dict(Name=kind+'-'+frame['name'],Frame=frame,View=view,Formats=formats,
             FactionText=b64(text(pm,ptr(pm,ds+0x60ae+owner*4))),
             PlanetKindText=b64(text(pm,ptr(pm,ds+0x619c+pm[planet+18]*4))),
             FleetSide=side,PlanetColour=m[ds+0x4422+owner],Landing=bool(landing),
             Visibility=visibility, FleetDetail=fm[ship+word(fm,fleet+16)*78+23],
             PlanetKind=pm[planet+18],
             LabelColour=label_colour,ResultColour=result_colour,Lower=bool(le),
             LowerView=lower(le) if le else None,Pictures=pictures,
             Writers=dict(Fleet=tf,Planet=tp,FleetGauge=gf,PlanetGauge=gp,Middle=last),
             NativePath=str(a/(frame['name']+'-screen.bin')),NativeSHA256=hashlib.sha256(native).hexdigest()))
result = dict(Schema='logh3.landing-caller-probe.v1', Cases=cases, Sources=sources)
with Path(sys.argv[1]).open('x') as f:
    json.dump(result,f,ensure_ascii=False,indent=2)
    f.write('\n')
print('typed cases',len(cases),'sources',len(sources))
