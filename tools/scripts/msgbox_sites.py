"""把所有 `3826:0005`（訊息框）的呼叫端與它們的字串抓出來。

    msgbox_sites.py <程式碼傾印> <起始線性位址hex> <ds 傾印> <輸出.json>

呼叫慣例（`3826:0005` 的 `[bp+32]`–`[bp+38]`）：由後往前推入
「字串段、字串位移、色、種類」。Borland 的碼形固定是 `mov r,imm ; push r`
與 `push ds`，所以往回掃 48 bytes 找那四個推入就夠；**認不出來的就寫
`null`，不猜**。
"""
import json
import sys

CODE, BASE, DSF, OUT = sys.argv[1], int(sys.argv[2], 16), sys.argv[3], sys.argv[4]
code = open(CODE, "rb").read()
ds = open(DSF, "rb").read()

SEGS = sorted([0x090A, 0x0C0A, 0x1373, 0x1588, 0x165F, 0x1727, 0x17A0, 0x1930,
               0x1AAB, 0x1C03, 0x1C61, 0x1CC5, 0x1D00, 0x203A, 0x20A0, 0x2100,
               0x21A0, 0x2264, 0x2350, 0x2376, 0x2583, 0x2703, 0x28E0, 0x297B,
               0x2A00, 0x2AF5, 0x2B99, 0x2BE1, 0x2E54, 0x307D, 0x3310, 0x341B,
               0x34CB, 0x3603, 0x3617, 0x3667, 0x3703, 0x37A9, 0x3826, 0x3870,
               0x3982, 0x39D7, 0x3ABF, 0x3CC9, 0x3CD5])


def where(lin):
    best = None
    for s in SEGS:
        if s * 16 <= lin:
            best = s
    return "%04X:%04X" % (best, lin - best * 16)


def string_at(off):
    end = ds.find(b"\x00", off)
    if end < 0 or end == off:
        return None, None
    raw = ds[off:end]
    try:
        return raw.decode("big5"), raw.hex()
    except UnicodeDecodeError:
        return None, raw.hex()


def pushes(win):
    """把窗口裡的推入序列解出來，只認得 `mov r,imm ; push r` 與 `push ds`。"""
    out, i = [], 0
    reg = {0xB8: "ax", 0xBB: "bx", 0xB9: "cx", 0xBA: "dx", 0xBE: "si", 0xBF: "di"}
    push = {0x50: "ax", 0x53: "bx", 0x51: "cx", 0x52: "dx", 0x56: "si", 0x57: "di"}
    held = {}
    while i < len(win):
        b = win[i]
        if b in reg and i + 2 < len(win):
            held[reg[b]] = int.from_bytes(win[i + 1:i + 3], "little")
            i += 3
        elif b == 0xB0 and i + 1 < len(win):
            held["ax"] = win[i + 1]
            i += 2
        elif b == 0x31 and i + 1 < len(win) and win[i + 1] == 0xC0:
            held["ax"] = 0
            i += 2
        elif b in push:
            out.append(held.get(push[b]))
            i += 1
        elif b == 0x1E:
            out.append("ds")
            i += 1
        elif b == 0x8C and i + 1 < len(win):  # mov r,sreg
            held[{0xD8: "ax", 0xDB: "bx"}.get(win[i + 1], "ax")] = "ds"
            i += 2
        else:
            i += 1
    return out


pat = bytes([0x9A, 0x05, 0x00, 0x26, 0x38])
sites, i = [], 0
while True:
    i = code.find(pat, i)
    if i < 0:
        break
    lin = BASE + i
    seq = pushes(code[i - 48:i])
    kind = colour = soff = None
    if len(seq) >= 4 and isinstance(seq[-1], int) and isinstance(seq[-2], int):
        kind, colour = seq[-1], seq[-2]
        if isinstance(seq[-3], int):
            soff = seq[-3]
    text, raw = string_at(soff) if soff is not None else (None, None)
    sites.append({"call": where(lin), "linear": "%05X" % lin, "kind": kind,
                  "colour": colour,
                  "string": ("ds:%04X" % soff) if soff is not None else None,
                  "text": text, "bytes": raw})
    i += 1

json.dump({"_note": "訊息框（3826:0005）的呼叫端；由 tools/scripts/msgbox_sites.py 掃出",
           "sites": sites}, open(OUT, "w"), ensure_ascii=False, indent=1)
for s in sites:
    print("%-12s 種類 %-4s 色 %-4s %-9s %r" % (
        s["call"], s["kind"], ("%02X" % s["colour"]) if s["colour"] is not None else "?",
        s["string"] or "?", s["text"]))
