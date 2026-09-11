# -*- coding: utf-8 -*-
"""從一份 probe 報告數出「兩個斷點之間抽了幾次亂數」。

    tools/py.sh tools/scripts/rng_watch_draws.py <report.txt> [種子線性位址hex]

報告要同時有：

- `-watch 5E00C-5E00F`：原版亂數種子 `ds:94DC`（`docs/re/87`）的每一次寫入；
- `-regs-at <位址,...>`：想比對的那幾個斷點（每一次命中的步數）。

做法是照 `[watch]` 那幾行把 4 個位元組的種子一路重建出來，取每一次斷點
命中那一刻的種子，再用 LCG 往前走去數兩個種子之間差幾步。斷點落在
`3CD5:000C` 執行中途的機會是零（那支沒有斷點會停的地方），所以每一刻的
種子都是完整的。

用途：`docs/re/297` 拿它量 `2703:1789` 一格抽幾次（收據 `rbx297w`）。
這支只讀報告，不改任何東西。
"""
import re
import sys

MUL, ADD = 0x41C64E6D, 0x3039


def main():
    lines = open(sys.argv[1], encoding="utf-8").read().splitlines()
    base = int(sys.argv[2], 16) if len(sys.argv) > 2 else 0x5E00C

    writes = []
    for l in lines:
        m = re.match(r"\[watch\] #(\d+) ([0-9A-F]{5}): ([0-9A-F]{2}) → ([0-9A-F]{2})", l)
        if m:
            writes.append((int(m.group(1)), int(m.group(2), 16), int(m.group(4), 16)))
    if not writes:
        raise SystemExit("報告裡沒有 [watch] 那幾行——要加 -watch 監看種子")

    events, sec = [], None
    for l in lines:
        m = re.match(r"(\S+) 執行時的暫存器", l)
        if m:
            sec = m.group(1)
            continue
        m = re.match(r"\s+#(\d+) AX", l)
        if m and sec:
            events.append((int(m.group(1)), sec))
        elif sec and not l.startswith(" "):
            sec = None
    events.sort()

    def seed_at(step):
        b = [0, 0, 0, 0]
        for s, a, v in writes:
            if s >= step:
                break
            if base <= a < base + 4:
                b[a - base] = v
        return b[0] | b[1] << 8 | b[2] << 16 | b[3] << 24

    def dist(a, b, limit=100000):
        n = 0
        while a != b and n < limit:
            a = (a * MUL + ADD) & 0xFFFFFFFF
            n += 1
        return n if a == b else None

    prev = None
    count = {}
    print("步數          斷點        種子      與上一個斷點之間抽了幾次")
    for step, sec in events:
        s = seed_at(step)
        d = dist(prev, s) if prev is not None else None
        count[sec] = count.get(sec, 0) + 1
        print("#%-11d  %-10s  %08X  %s" % (step, "%s#%d" % (sec, count[sec] - 1), s,
                                            "—" if d is None else d))
        prev = s
    print("報告結束時的種子 %08X" % seed_at(1 << 62))


main()
