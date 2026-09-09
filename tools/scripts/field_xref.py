"""在整份程式碼傾印裡找「以 disp8 存取某個結構欄位」的指令。

    field_xref.py <傾印> <起始線性位址hex> <位移hex> [位移hex ...]

只收 mod=01（disp8）且基底是 bx／si／di 的 modrm——bp 是堆疊框，
排除掉才看得見結構欄位。輸出的線性位址要自己換算回段:位移。
"""
import sys

OPS = {
    0x00: "add", 0x01: "add", 0x02: "add", 0x03: "add",
    0x08: "or", 0x09: "or", 0x0A: "or", 0x0B: "or",
    0x10: "adc", 0x11: "adc", 0x12: "adc", 0x13: "adc",
    0x18: "sbb", 0x19: "sbb", 0x1A: "sbb", 0x1B: "sbb",
    0x20: "and", 0x21: "and", 0x22: "and", 0x23: "and",
    0x28: "sub", 0x29: "sub", 0x2A: "sub", 0x2B: "sub",
    0x30: "xor", 0x31: "xor", 0x32: "xor", 0x33: "xor",
    0x38: "cmp", 0x39: "cmp", 0x3A: "cmp", 0x3B: "cmp",
    0x80: "grp1b", 0x81: "grp1w", 0x83: "grp1ws",
    0x84: "test", 0x85: "test",
    0x88: "mov", 0x89: "mov", 0x8A: "mov", 0x8B: "mov",
    0x8D: "lea", 0x8C: "movsr", 0x8E: "movsr",
    0xC4: "les", 0xC5: "lds", 0xC6: "movib", 0xC7: "moviw",
    0xF6: "grp3b", 0xF7: "grp3w", 0xFE: "grp4", 0xFF: "grp5",
}
BASE = {4: "si", 5: "di", 7: "bx"}
GRP1 = ["add", "or", "adc", "sbb", "and", "sub", "xor", "cmp"]
GRP3 = ["test", "?", "not", "neg", "mul", "imul", "div", "idiv"]
GRP5 = ["inc", "dec", "call", "callf", "jmp", "jmpf", "push", "?"]

dump = open(sys.argv[1], "rb").read()
base = int(sys.argv[2], 16)
for t in sys.argv[3:]:
    disp = int(t, 16)
    print("== +0x%02X ==" % disp)
    n = 0
    for i in range(len(dump) - 4):
        op = dump[i]
        if op not in OPS:
            continue
        m = dump[i + 1]
        if m >> 6 != 1 or (m & 7) not in BASE:
            continue
        if dump[i + 2] != disp:
            continue
        pfx = ""
        if i and dump[i - 1] in (0x26, 0x2E, 0x36, 0x3E):
            pfx = {0x26: "es:", 0x2E: "cs:", 0x36: "ss:", 0x3E: "ds:"}[dump[i - 1]]
        name = OPS[op]
        reg = (m >> 3) & 7
        if name == "grp1b" or name == "grp1w" or name == "grp1ws":
            name = GRP1[reg]
        elif name in ("grp3b", "grp3w"):
            name = GRP3[reg]
        elif name == "grp5":
            name = GRP5[reg]
        tail = " ".join("%02X" % b for b in dump[i:i + 6])
        print("    lin %05X  %-6s %s[%s+%02X]  %s" % (
            base + i, name, pfx, BASE[m & 7], disp, tail))
        n += 1
    print("    共 %d 處" % n)
