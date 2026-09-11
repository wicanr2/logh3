# -*- coding: utf-8 -*-
"""找「把某個結構欄位當 unsigned char 讀進運算式」的指令。

    tools/py.sh tools/scripts/loadbyte.py <傾印> <起始線性位址hex> <位移hex> [位移hex ...]

`field_xref.py` 收所有 disp8 存取形式，訊號很吵（同一個位移在十幾種結構上
都會出現）；這一支只收 Borland 讀 `unsigned char` 欄位的固定形狀——
`8A /r disp8` 後面緊接著清高位元組（`xor ah,ah` 家族或 `mov ah,0`）。
代價是漏掉「只拿去跟常數比」的讀法，換來的是幾乎沒有假陽性。

基底收 `si`／`di`／`bx`／`bp`。**`bp` 的那些是堆疊區域變數不是結構欄位**，
判讀時要先把它們分出去（例如艦隊組成畫面把十一個艦隊能力抄到
`[bp+0x3F]`–`[bp+0x49]` 再逐格印，看起來會很像艦隊記錄）。

**下「沒有消費者」的結論之前先做正對照**：同一次跑一個已知有人用的位移，
它要找得到；再拿兩份以上不同階段的傾印跑同一組位移，命中數要一致
（程式影像沒有換頁）。用例見
[`docs/re/287`](../../docs/re/287-admiral-remaining-fields.md)。
"""
import sys

REG = {0: "al", 1: "cl", 2: "dl", 3: "bl", 4: "ah", 5: "ch", 6: "dh", 7: "bh"}
RM = {4: "si", 5: "di", 6: "bp", 7: "bx"}
SEG = {0x26: "es:", 0x2E: "cs:", 0x36: "ss:", 0x3E: "ds:"}
ZERO = {
    b"\x30\xe4": "xor ah,ah", b"\x30\xed": "xor ch,ch",
    b"\x30\xff": "xor bh,bh", b"\x30\xf6": "xor dh,dh",
    b"\xb4\x00": "mov ah,0",
}

mem = open(sys.argv[1], "rb").read()
base = int(sys.argv[2], 16)
for t in sys.argv[3:]:
    disp = int(t, 16)
    print("== +0x%02X ==" % disp)
    n = 0
    for i in range(1, len(mem) - 6):
        if mem[i] != 0x8A:
            continue
        m = mem[i + 1]
        if m >> 6 != 1 or (m & 7) not in RM or mem[i + 2] != disp:
            continue
        tail = mem[i + 3:i + 5]
        if tail not in ZERO:
            continue
        print("    lin %05X  mov %s,%s[%s+%02X]  ; %s" % (
            base + i, REG[(m >> 3) & 7], SEG.get(mem[i - 1], ""),
            RM[m & 7], disp, ZERO[tail]))
        n += 1
    print("    共 %d 處" % n)
