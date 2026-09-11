# -*- coding: utf-8 -*-
"""戰場索敵：把「誰看得到誰」照 `2526:00F8` 那條鏈重算一次，跟傾印裡的
艦艇 `+0x17`／`+0x18` 並排。

    tools/py.sh tools/scripts/scout_vision.py <1MB 記憶體傾印> [更多傾印...]

觀測者（`2526:0198`／`2526:022C`）：

- 己方每一支在場艦隊的**旗艦分隊**（艦隊 `+0x10` 那一組的池索引），
  半徑 200；分隊艦種（`+0x04`）為 7 時 250（`2526:01FE`）。
- 已經放出去的**索敵分遣隊**（艦隊 `+0x38`，條件 `+0x3A != 0xFF`），同一組半徑。
- 己方**在這個星域的每一顆行星**，半徑 ＝ `clamp(行星[+0x14], 1500, 6000) ÷ 25`
  （`2AF5:01F7`）。

被看到的候選（`17A0:00AC`）：在場分隊清單（`ds:98BA`、筆數 `ds:9896`）裡
陣營等於**敵方**（`1 − 觀測方`）、`+0x1E` bit7 沒設、且距離 ≤ 半徑的。

**這支是對照工具不是證據**：`+0x17` 是跨回合累積上去的（每命中一輪 ＋1、
上限 3），傾印只是某一個瞬間，所以「現在在範圍內」與「`+0x17` 現在是多少」
不會逐筆相等。要看的是兩個方向的極端：範圍內卻 `+0x17 == 0`、
或 `+0x17 != 0` 卻離最近的敵方觀測者很遠。
"""
import math
import pathlib
import struct
import sys

FLEET, FSTRIDE, SREC = 0x60BA4, 122, 78
DS = FLEET - 0xC074


def u16(m, a):
    return struct.unpack_from("<H", m, a)[0]


def s16(m, a):
    return struct.unpack_from("<h", m, a)[0]


def dist(a, b):
    return math.hypot(a[0] - b[0], a[1] - b[1])


def report(path):
    m = pathlib.Path(path).read_bytes()
    off, seg = struct.unpack_from("<HH", m, DS + 0xCFB4)
    pool = seg * 16 + off
    poff, pseg = struct.unpack_from("<HH", m, DS + 0xCFF8)
    planets = pseg * 16 + poff
    nplanet = m[DS + 0xCFFC]
    nfleet = u16(m, DS + 0x9894)
    nunit = u16(m, DS + 0x9896)
    region, turn, side = m[DS + 0x8F17], u16(m, DS + 0x8F1A), m[DS + 0x8F1D]
    print("== %s" % path)
    print("   星域 %d、回合 %d（ds:8F1A）、目前陣營 %d；在場艦隊 %d、在場分隊 %d、行星 %d"
          % (region, turn, side, nfleet, nunit, nplanet))

    fleets = [m[DS + 0x9898 + i] for i in range(nfleet)]
    units = [u16(m, DS + 0x98BA + i * 2) for i in range(nunit)]

    # 觀測者：(陣營, x, y, 半徑, 標籤)
    eyes = []
    for f in fleets:
        r = FLEET + f * FSTRIDE
        # 艦隊記錄的 +0x01 只用來挑「這一輪輪到誰」，觀測方看分隊自己的 +0x00
        idx = u16(m, r + 0x10)
        if idx != 0xFFFF:
            u = pool + idx * SREC
            rad = 250 if m[u + 4] == 7 else 200
            # 觀測方取「這一個分隊自己的 +0x00」（`2526:01FE` push es:[bx]），
            # 不是艦隊記錄的 +0x01——注入過的收據兩者會不一樣。
            eyes.append((m[u], s16(m, u + 0x34), s16(m, u + 0x36), rad,
                         "艦隊%d 旗艦" % f))
        if m[r + 0x3A] != 0xFF:
            idx = u16(m, r + 0x38)
            u = pool + idx * SREC
            rad = 250 if m[u + 4] == 7 else 200
            eyes.append((m[u], s16(m, u + 0x34), s16(m, u + 0x36), rad,
                         "艦隊%d 偵察隊" % f))
    for i in range(nplanet):
        p = planets + i * 82
        owner = m[p + 0x11]
        if owner > 1:
            continue
        rad = min(max(u16(m, p + 0x14), 1500), 6000) // 25
        eyes.append((owner, s16(m, p + 0x46), s16(m, p + 0x48), rad,
                     "行星%d" % i))

    rows = []
    for idx in units:
        u = pool + idx * SREC
        fac = m[u]
        pos = (s16(m, u + 0x34), s16(m, u + 0x36))
        best, who = None, ""
        for efac, ex, ey, rad, tag in eyes:
            if efac == fac:
                continue
            d = dist(pos, (ex, ey)) - rad
            if best is None or d < best:
                best, who = d, tag
        rows.append((idx, fac, m[u + 1], m[u + 3], m[u + 4], m[u + 0x16],
                     m[u + 0x17], m[u + 0x18], m[u + 0x1E], pos, best, who))

    inrange_seen = [r for r in rows if r[10] is not None and r[10] <= 0 and r[6] != 0]
    inrange_blind = [r for r in rows if r[10] is not None and r[10] <= 0 and r[6] == 0]
    out_seen = [r for r in rows if (r[10] is None or r[10] > 0) and r[6] != 0]
    print("   觀測者 %d；在敵方視野內 %d（其中 +0x17 != 0 的 %d）、"
          "視野外但 +0x17 != 0 的 %d"
          % (len(eyes), len(inrange_seen) + len(inrange_blind),
             len(inrange_seen), len(out_seen)))

    def dump(title, rs):
        if not rs:
            return
        print("   %s" % title)
        print("     池# 陣營 艦隊 組 艦種 +16 +17 +18 +1E   位置        "
              "離最近敵方觀測者（負＝在範圍內）")
        for (idx, fac, fl, grp, kind, f16, f17, f18, f1e, pos, best, who) in rs:
            b = "－" if best is None else "%+8.1f %s" % (best, who)
            print("     %3d  %d  %3d %2d  %d   %02X  %d   %-3d %02X  (%5d,%5d) %s"
                  % (idx, fac, fl, grp, kind, f16, f17, f18, f1e,
                     pos[0], pos[1], b))

    dump("在敵方視野內且 +0x17 != 0", inrange_seen)
    dump("在敵方視野內但 +0x17 == 0", inrange_blind)
    dump("視野外但 +0x17 != 0", out_seen)
    print()


for p in sys.argv[1:]:
    report(p)
