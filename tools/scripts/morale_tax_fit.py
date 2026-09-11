"""拿收據 `mrx295` 驗：同盟 AI 回合重算稅率時，讀的是**這一回合收斂之前**的民心（`docs/re/295`）。

    tools/py.sh tools/scripts/morale_tax_fit.py <收據目錄> <前綴> r0p r1 r2 r3

稅率算式照 `090A:0B93`（`docs/internals/55`）逐位元組照做，包含一個
remake 在合法值域裡碰不到的細節：`m` 是 **16 位元無號**算的——
`200 − 民心` 零延伸成 word、`mul`（無號）之後**丟掉高位 DX** 再 `div 100`、
用 `jbe` 比。民心 > 200 的同盟星域因此會得到一個很大的 `m`。

收據沒有倒提督表與 `Blend`，所以兩個未知數用擬合：對 `Base ∈ {50,75,100,150,200}`
× `ADMIRAL.DTX` 的 82 位提督，找出能讓**沒被改過民心的**同盟星域稅率全部
重現的組合（要唯一），再拿它去預測被改過的四格——分別代入「前一份傾印的民心」
（收斂之前）與「這一份的民心」（收斂之後），看哪一種對得上。
"""
import sys

REC = 126
OWNER, GROWTH, TAX, MORALE = 0x12, 0x1E, 0x26, 0x7C
POKED = (15, 16, 17, 18)


def tax(growth, morale, base, adm):
    v = base * 5 // 100 if growth <= 0 else growth * base // 100
    v = v * 100 // adm[0x0E]
    v = v * adm[0x15] // 100
    w = adm[0x14]
    m = (((200 - morale) & 0xFFFF) * w & 0xFFFF) // 100
    if m > 150:
        v = v * 30 // w
    elif m > 100:
        v = v * 50 // w
    elif m > 50:
        v = v * 75 // w
    return max(1, min(200, v))


def field(t, i, off, n=1, signed=False):
    return int.from_bytes(t[i * REC + off:i * REC + off + n], "little", signed=signed)


def main():
    d, prefix, names = sys.argv[1], sys.argv[2], sys.argv[3:]
    admirals = open("/orig/ADMIRAL.DTX", "rb").read()
    tabs = {n: open("%s/%s_%s.bin" % (d, prefix, n), "rb").read() for n in names}
    ally = [i for i in range(32) if field(tabs[names[0]], i, OWNER) == 1]
    plain = [i for i in ally if i not in POKED]
    for a, b in zip(names, names[1:]):
        ta, tb = tabs[a], tabs[b]
        fits = []
        for base in (50, 75, 100, 150, 200):
            for k in range(len(admirals) // 28):
                adm = admirals[k * 28:(k + 1) * 28]
                if adm[0x0E] == 0 or adm[0x14] == 0:
                    continue
                if all(tax(field(tb, i, GROWTH, 2, True), field(tb, i, MORALE), base, adm)
                       == field(tb, i, TAX) for i in plain):
                    fits.append((base, k, adm))
        print("== %s → %s：沒改過的同盟星域 %d 格，擬合出 %d 組" % (a, b, len(plain), len(fits)))
        for base, k, adm in fits:
            print("   Base %d、提督 %d（+0x0E %d、+0x14 %d、+0x15 %d）"
                  % (base, k, adm[0x0E], adm[0x14], adm[0x15]))
            for i in POKED:
                g = field(tb, i, GROWTH, 2, True)
                pre = tax(g, field(ta, i, MORALE), base, adm)
                post = tax(g, field(tb, i, MORALE), base, adm)
                print("   星域 %d 民心 %d→%d：收斂前算 %d、收斂後算 %d，收據 %d"
                      % (i, field(ta, i, MORALE), field(tb, i, MORALE), pre, post,
                         field(tb, i, TAX)))


main()
