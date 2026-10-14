"""144 DRAFT：只由原版記錄與呼叫順序建完整單幀 typed input。

容器內於/src執行。原版像素只存路徑及雜湊，絕不進呈現輸入。
"""
import base64
import hashlib
import json
import struct
import sys
from pathlib import Path

DS = 0x54B30


def word(m, a):
    return struct.unpack_from("<H", m, a)[0]


def ptr(m, a):
    return word(m, a) + 16 * word(m, a + 2)


def text(m, a):
    return m[a:a + 256].split(b"\0", 1)[0]


def b64(b):
    return base64.b64encode(b).decode()


def digest(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def prepare(stem):
    root = Path("workplace/oracle") / (stem + "-a")
    r = json.loads((root / "report.json").read_text())
    assert r["dosgolem_head"] == "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c" and not r["dosgolem_dirty"]
    mem = lambda f: (root / (f["name"] + "-mem.bin")).read_bytes()
    baseline = (root / "gun-before-mem.bin").read_bytes()
    pool = ptr(baseline, DS + 0xCFB4)
    planet = ptr(baseline, DS + 0xCFF8)
    calls = [e for e in r["events"] if "comani_record_hex" in e]
    returns = [f for f in r["frames"] if f["name"].startswith("comani-return")]
    assert len(calls) == len(returns)
    damage = [e for e in r["events"] if e["original_location"] == "2264:0EE3"]
    starts = [e for e in calls if bytes.fromhex(e["comani_record_hex"])[1] in (9, 10, 11) and e["stack_words"][4] == 1]
    assert len(starts) == len(damage)
    bmg = Path("/orig/COMANI.BMG").read_bytes()
    counts = {}
    for a in range(2, 2 + word(bmg, 0), 10):
        offset = struct.unpack_from("<I", bmg, a + 6)[0] + 2
        counts[bmg[a:a + 4].hex()] = word(bmg, offset)

    def visibility(m):
        return dict(Mode=m[DS + 0x8F20], PlayerSide=m[DS + 0x8F25], SideControl=list(m[DS + 0x8F21:DS + 0x8F23]))

    def fleet(m, i):
        return DS + 0xC074 + i * 122

    def lower(m, fi):
        f = fleet(m, fi)
        units = []
        for i in range(m[f + 15]):
            idx = word(m, f + 16 + 4 * i)
            u = pool + idx * 78
            units.append(dict(ID=idx, Side=m[u], X=word(m, u + 52), Y=word(m, u + 54),
                              Squadron=m[u + 3], Detail=m[u + 23], Ships=word(m, u + 8)))
        return dict(Units=units, Visibility=visibility(m), Facing=m[f + 74],
                    X=word(m, planet + 70), Y=word(m, planet + 72), Owner=m[planet + 17], Defence=word(m, planet + 20))

    stages = []
    for i, start in enumerate(starts):
        k = calls.index(start)
        before = mem(returns[k])
        shot = damage[i]
        unit = shot["stack_words"][4] + 16 * shot["stack_words"][5]
        assert (unit - pool) % 78 == 0
        end = starts[i + 1]["step"] if i + 1 < len(starts) else r["end_step"]
        damaged_call = next(j for j in range(k + 1, len(calls)) if calls[j]["step"] > shot["step"] and calls[j]["step"] < end)
        after = mem(returns[damaged_call])
        wait = next(e for e in r["events"] if e["original_location"] == "3765:006C" and returns[damaged_call]["step"] < e["step"] < end)
        gauge = next(e for e in r["events"] if e["original_location"] == "321F:0DF7" and (starts[i - 1]["step"] if i else 0) < e["step"] < start["step"])
        stages.append(dict(start=start["step"], end=end, before=before, after=after, unit=unit,
                           shot=shot["step"], result=wait["step"], attack=gauge["stack_words"][6]))

    def picture(call, last):
        rec = bytes.fromhex(call["comani_record_hex"])
        m = baseline
        if rec[1] in (9, 10, 11):
            prefix = bytes([ord("P"), m[DS + 0x8E1A + rec[1]]])
        else:
            prefix = bytes([m[DS + 0x8E16 + rec[0]], m[DS + 0x8E1A + rec[1]] if rec[4] != 0 else m[DS + 0x8E34 + rec[5]]])
        if call["stack_words"][4]:
            suffix = b"D0"
        else:
            fourth = m[DS + 0x8E2C + rec[3] + (3 if rec[2] == 0 else 0)]
            if rec[2] == 0 and rec[1] in (10, 11) and rec[3] == 1:
                fourth = b"MF"[rec[0]]
            suffix = bytes([m[DS + 0x8E28 + rec[2]], fourth])
        segment = (prefix + suffix).hex()
        return dict(Side=rec[0], SegmentHex=segment, LastFrame=counts[segment] - 1 if last is None else last)

    formats = {k: text(baseline, DS + a).hex() for k, a in dict(clear=0x4957, unknown=0x4959,
        prefix=0x496D, suffix=0x4970, squadron=0x4975, admiral=0x4984, ships=0x49A9,
        title=0x49D3, name=0x49D7, defence=0x49DD, label=0x49F8, goto=0x49C0, destroyed=0x49C9).items()}
    assert text(baseline, DS + 0x498B).hex() == formats["prefix"]
    assert text(baseline, DS + 0x498E).hex() == formats["suffix"]
    assert text(baseline, DS + 0x4993).hex() == formats["squadron"]
    assert text(baseline, DS + 0x49A2).hex() == formats["admiral"]
    factions = [text(baseline, ptr(baseline, DS + 0x60AE + 4 * i)).hex() for i in range(3)]
    kinds = [text(baseline, ptr(baseline, DS + 0x619C + 4 * i)).hex() for i in range(3)]
    cases = []
    close = next(e["step"] for e in r["events"] if e["original_location"] == "1DD6:026B")
    for frame in r["frames"]:
        step = frame["step"]
        if step < stages[0]["start"] or step >= close:
            continue
        stage = next(s for s in reversed(stages) if s["start"] <= step)
        result = step >= stage["result"]
        m = stage["after"] if result else stage["before"]
        u = stage["unit"]
        f = fleet(m, m[u + 1])
        target = dict(Side=m[u], Squadron=m[u + 3], Detail=m[u + 23], FleetNumber=word(m, f + 2),
                      AdmiralName=b64(text(m, ptr(m, DS + 0x6428 + 4 * m[f + 4]))),
                      KindName=b64(text(m, ptr(m, DS + 0x613A + 4 * m[u + 4]))),
                      Ships=word(m, u + 8), ShipsMax=word(m, u + 10), Visibility=visibility(m))
        pictures = []
        cleared = {}
        for e in r["events"]:
            if e["original_location"] == "3958:01CC" and e["step"] < step:
                cleared[e["stack_words"][2] & 255] = e["step"]
        if baseline[DS + 0x8F36]:
            for call, ret in zip(calls, returns):
                side = bytes.fromhex(call["comani_record_hex"])[0]
                if call["step"] < cleared.get(side, 0):
                    continue
                if ret["step"] <= step:
                    pictures.append(picture(call, None if call["stack_words"][5] else 0))
                elif call["step"] < step and call["stack_words"][5]:
                    draws = sum(f["name"].startswith("draw-frame-return") and call["step"] < f["step"] <= step for f in r["frames"])
                    if draws > 1:
                        pictures.append(picture(call, draws - 2))
        low = [e for e in r["events"] if e["original_location"] == "321F:0897" and e["step"] < step]
        scene = None
        if low:
            e = low[-1]
            ls = next(s for s in reversed(stages) if s["start"] <= e["step"])
            lm = ls["before"] if e["step"] < ls["shot"] else ls["after"]
            fp = e["stack_words"][2] + 16 * e["stack_words"][3]
            scene = lower(lm, (fp - DS - 0xC074) // 122)
        native = root / (frame["name"] + "-screen.bin")
        cases.append(dict(Name=stem + "-" + frame["name"],
            Planet=dict(Owner=m[planet + 17], Kind=m[planet + 18], Name=b64(text(m, planet)),
                        Defence=word(m, planet + 20), DefenceMax=word(m, planet + 22)),
            Target=target, Gauges=dict(FleetFaction=m[u], Integrity=m[f + 60],
                                      Defence=word(m, planet + 20), Attack=stage["attack"]),
            Pictures=pictures, ClearSides=sorted(cleared), Scene=scene, ShowResult=result,
            Destroyed=word(stage["before"], u + 8) - word(stage["after"], u + 8),
            NativePath=str(native), NativeSHA256=digest(native), Frame=frame))
    return dict(Cases=cases, Formats=formats, Factions=factions, Kinds=kinds,
                Colours=list(baseline[DS + 0x4422:DS + 0x4425]),
                FortressCues=list(baseline[DS + 0x47B4:DS + 0x47BA]),
                ReportSHA256=digest(root / "report.json"))


result = prepare(sys.argv[1])
with Path(sys.argv[2]).open("x") as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write("\n")
print("typed complete frames", len(result["Cases"]))
