# -*- coding: utf-8 -*-
"""看一個存取點附近還碰了同一個結構的哪些位移——用來判「這是哪一張表」。

    tools/py.sh tools/scripts/struct_window.py <傾印> <線性位址hex> [半徑hex] ...

`field_xref.py` 只回答「哪裡出現過 disp8 ＝ N」，回答不了「那個基底指向誰」。
判準改成鄰居：同一支函式會反覆 `les bx,[bp+X]` 再連續讀同一張表，所以把
視窗內所有帶段前綴（`es:`／`ds:`）的 `mod=01` 存取列出來，看那組位移像哪
一張表——艦隊 122 bytes（`+0x04` 提督、`+0x0F` 編制數、`+0x5C` 座標）、
艦艇 78 bytes（`+0x08` 艘數、`+0x34` 座標）、星域 126 bytes、提督 28 bytes。

位移集合是線索不是證據；要下語意結論仍然要把那一段讀出來。
"""
import collections
import pathlib
import sys

OPS = {0x00: "add", 0x01: "add", 0x02: "add", 0x03: "add",
       0x08: "or", 0x09: "or", 0x0A: "or", 0x0B: "or",
       0x20: "and", 0x21: "and", 0x22: "and", 0x23: "and",
       0x28: "sub", 0x29: "sub", 0x2A: "sub", 0x2B: "sub",
       0x30: "xor", 0x31: "xor", 0x32: "xor", 0x33: "xor",
       0x38: "cmp", 0x39: "cmp", 0x3A: "cmp", 0x3B: "cmp",
       0x80: "grp1b", 0x81: "grp1w", 0x83: "grp1ws",
       0x84: "test", 0x85: "test",
       0x88: "mov r/m8,r", 0x89: "mov r/m16,r",
       0x8A: "mov r,r/m8", 0x8B: "mov r,r/m16",
       0xC4: "les", 0xC5: "lds", 0xC6: "mov r/m8,imm", 0xC7: "mov r/m16,imm",
       0xF6: "grp3b", 0xF7: "grp3w", 0xFE: "grp4", 0xFF: "grp5"}
BASE = {4: "si", 5: "di", 6: "bp", 7: "bx"}
SEG = {0x26: "es:", 0x2E: "cs:", 0x36: "ss:", 0x3E: "ds:"}

mem = pathlib.Path(sys.argv[1]).read_bytes()
radius = int(sys.argv[3], 16) if len(sys.argv) > 3 else 0xC0
for t in sys.argv[2].split(","):
    at = int(t, 16)
    lo, hi = max(0, at - radius), min(len(mem) - 4, at + radius)
    print("==== 線性 %05X（±%X）====" % (at, radius))
    seen = collections.Counter()
    i = lo
    while i < hi:
        if mem[i] in SEG and mem[i + 1] in OPS:
            m = mem[i + 2]
            if m >> 6 == 1 and (m & 7) in BASE and (m & 7) != 6:
                print("    %05X  %-12s %s[%s+%02X]" % (
                    i, OPS[mem[i + 1]], SEG[mem[i]], BASE[m & 7], mem[i + 3]))
                seen[mem[i + 3]] += 1
        i += 1
    print("    位移集合：%s" % " ".join(
        "+%02X×%d" % (o, n) for o, n in sorted(seen.items())))
