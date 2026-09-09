"""比對整補前後兩份 1 MB 傾印：艦隊編制與星域待命部隊。

    tools/py.sh tools/scripts/refit_diff.py <pre.bin> <post.bin> <陣營> [艦隊槽...]

艦隊表 `ds:C074`、星域表 `ds:A79C`、艦艇池遠指標 `ds:CFB4`
（`docs/internals/33`、`docs/re/215`）。
"""
import sys

DS = 0x54B30
FLEET = DS + 0xC074
REGION = DS + 0xA79C
POOL = DS + 0xCFB4


def u16(d, a):
    return d[a] | d[a + 1] << 8


def u32(d, a):
    return u16(d, a) | u16(d, a + 2) << 16


def pool(d):
    return u16(d, POOL + 2) * 16 + u16(d, POOL)


def fleet_desc(d, slot):
    f = FLEET + slot * 122
    if d[f + 4] == 0xFF:
        return None
    p = pool(d)
    groups = []
    for i in range(d[f + 0x0F]):
        g = f + 0x10 + i * 4
        idx = u16(d, g)
        s = p + idx * 78
        groups.append((idx, d[s + 4], d[s + 5], u16(d, s + 8), u16(d, s + 10)))
    return dict(admiral=d[f + 4], adjutant=d[f + 5], region=d[f + 6],
                want=d[f + 0x0E], have=d[f + 0x0F], ops=d[f + 0x47],
                upkeep=u16(d, f + 0x0A), groups=groups)


def stock(d, region, faction):
    r = REGION + region * 126
    return [u32(d, r + 0x30 + faction * 28 + k * 4) for k in range(7)]


pre = open(sys.argv[1], "rb").read()
post = open(sys.argv[2], "rb").read()
faction = int(sys.argv[3])
slots = [int(x) for x in sys.argv[4:]] or range(32)

for slot in slots:
    a, b = fleet_desc(pre, slot), fleet_desc(post, slot)
    if a is None and b is None:
        continue
    print("== 艦隊槽 %d" % slot)
    for tag, v in (("前", a), ("後", b)):
        if v is None:
            print("  %s：空槽" % tag)
            continue
        print("  %s：提督 %d 副官 %d 星域 %d 編制 %d/%d 運用 %d 維持 %d"
              % (tag, v["admiral"], v["adjutant"], v["region"],
                 v["have"], v["want"], v["ops"], v["upkeep"]))
        for i, g in enumerate(v["groups"]):
            print("     組%d 艦艇#%-3d 艦種%d 隊形槽%-3d %5d/%5d艘"
                  % (i, g[0], g[1], g[2], g[3], g[4]))
    if a and b and a["region"] == b["region"]:
        r = a["region"]
        print("  星域 %d 陣營 %d 待命部隊 前 %s" % (r, faction, stock(pre, r, faction)))
        print("  星域 %d 陣營 %d 待命部隊 後 %s" % (r, faction, stock(post, r, faction)))
