"""拿收據的星域表傾印驗「民心每回合往擁有者的目標值走一格」（`docs/re/295`）。

    tools/py.sh tools/scripts/morale_drift_check.py <收據目錄> <前綴> r0p r1 r2 r3

每個名字對應兩個檔：`<前綴>_<名字>.bin`（`ds:A79C` 那張 32 × 126 的星域表）
與 `<前綴>_<名字>_t.bin`（回合計數器 `ds:8F18`，2 bytes）。相鄰兩份傾印
之間要**正好跑過一次** `056A:0AA5`（收據的 `-regs-at` 會列出每一次）。

規則（`056A:0AA5`，反組譯見 `docs/re/295`）：

    目標 = 擁有者 0 → 0、擁有者 2 → 100、其餘 → 200
    d = 民心 − 目標
    d < 0 → 民心 + 1；d > 0 → 民心 − 1；d == 0 → 不動

擁有者取**前一份**傾印的 `+0x12`：兩份傾印都停在命令階段，`056A:0AA5`
跑在按下「實行」之後那一串的第五支，那時候這一回合還沒有任何一格易主
（擁有者在回合尾巴才依行星重算）。
"""
import sys

REC = 126
N = 32
OWNER, MORALE = 0x12, 0x7C


def target(owner):
    if owner == 0:
        return 0
    if owner == 2:
        return 100
    return 200


def step(morale, owner):
    d = morale - target(owner)
    if d < 0:
        return morale + 1
    if d > 0:
        return morale - 1
    return morale


def load(d, prefix, name):
    tab = open("%s/%s_%s.bin" % (d, prefix, name), "rb").read()
    t = open("%s/%s_%s_t.bin" % (d, prefix, name), "rb").read()
    if len(tab) != REC * N:
        raise SystemExit("%s 長度 %d，不是 %d" % (name, len(tab), REC * N))
    return tab, t[0] | t[1] << 8


def main():
    d, prefix, names = sys.argv[1], sys.argv[2], sys.argv[3:]
    snaps = [(n,) + load(d, prefix, n) for n in names]
    total = bad = moved = 0
    for (n0, a, t0), (n1, b, t1) in zip(snaps, snaps[1:]):
        print("== %s（回合 %d）→ %s（回合 %d）==" % (n0, t0, n1, t1))
        if t1 != t0 + 1:
            print("   ⚠ 回合計數器差 %d，不是 1" % (t1 - t0))
        for i in range(N):
            o = a[i * REC + OWNER]
            m0 = a[i * REC + MORALE]
            m1 = b[i * REC + MORALE]
            p = step(m0, o)
            total += 1
            if p != m0:
                moved += 1
            flag = "" if p == m1 else "   ✗"
            if p != m1:
                bad += 1
            if m0 != target(o) or m1 != m0 or p != m1:
                print("   星域 %2d 擁有者 %d 目標 %3d：%3d → 預測 %3d，實際 %3d%s"
                      % (i, o, target(o), m0, p, m1, flag))
    print("逐格比對 %d 格（其中預測會動的 %d 格），不符 %d 格" % (total, moved, bad))
    sys.exit(1 if bad else 0)


main()
