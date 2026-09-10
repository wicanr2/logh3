"""在記憶體傾印裡找「存取某個全域（disp16）」的指令。

    tools/py.sh tools/scripts/global_xref.py <傾印> <起始線性位址hex> <disp16hex> [更多...]

`field_xref.py` 掃的是 disp8 的結構欄位；全域變數走的是 **disp16**，
形狀不同：絕對定址（mod=00 rm=110）、基底加位移（mod=10，這個遊戲
一律是 `bx ＝ ds:8F1D` 的陣營下標），以及累加器專用的 `A0`–`A3`。

輸出標「讀」還是「寫」——找「誰改了這一格」時只有寫的那幾行有意義。
**這是掃位元組不是反組譯**：同樣的位元組也可能落在資料裡，命中之後
要拿 IDA 對那一段線性解一次確認（`disasm_funcs.py`）。

正對照：拿 `docs/re/206` 已經逐處列過的 `8FD4` 跑一次，應該重現
「15 處候選、13 處是指令」那張表。
"""
import pathlib
import sys

# opcode → (助憶, 對 r/m 是讀還是寫)。「改」＝ 讀出來改完寫回去。
OPS = {
    0x88: ("mov r/m8,r8", "寫"), 0x89: ("mov r/m16,r16", "寫"),
    0x8A: ("mov r8,r/m8", "讀"), 0x8B: ("mov r16,r/m16", "讀"),
    0xC6: ("mov r/m8,imm8", "寫"), 0xC7: ("mov r/m16,imm16", "寫"),
    0x84: ("test r/m8,r8", "讀"), 0x85: ("test r/m16,r16", "讀"),
    0x8D: ("lea", "取址"),
}
ALU = {0x00: "add", 0x08: "or", 0x10: "adc", 0x18: "sbb",
       0x20: "and", 0x28: "sub", 0x30: "xor", 0x38: "cmp"}
GRP1 = {0: "add", 1: "or", 2: "adc", 3: "sbb", 4: "and", 5: "sub", 6: "xor", 7: "cmp"}
GRP_FE = {0: "inc", 1: "dec"}
GRP_FF = {0: "inc", 1: "dec", 2: "call", 3: "call far", 4: "jmp", 5: "jmp far", 6: "push"}
RM = {0: "bx+si", 1: "bx+di", 2: "bp+si", 3: "bp+di", 4: "si", 5: "di", 6: "bp", 7: "bx"}
PREFIX = {0x26: "es:", 0x2E: "cs:", 0x36: "ss:", 0x3E: "ds:", 0xF2: "repne ", 0xF3: "rep "}


def decode(mem, i, disp):
    """i 是 disp16 兩個位元組的位置。回 (指令描述, 讀寫) 或 None。"""
    # 累加器專用形式：A0/A1 讀、A2/A3 寫，disp16 直接跟在 opcode 後面。
    op = mem[i - 1]
    if op in (0xA0, 0xA1, 0xA2, 0xA3):
        name = ["mov al,[%04X]", "mov ax,[%04X]", "mov [%04X],al", "mov [%04X],ax"][op - 0xA0]
        return name % disp, ("讀" if op < 0xA2 else "寫")
    if i < 2:
        return None
    modrm = mem[i - 1]
    mod, reg, rm = modrm >> 6, (modrm >> 3) & 7, modrm & 7
    if not (mod == 0 and rm == 6) and mod != 2:
        return None
    where = "[%04X]" % disp if mod == 0 else "[%s+%04X]" % (RM[rm], disp)
    op = mem[i - 2]
    pre = PREFIX.get(mem[i - 3], "") if i >= 3 else ""
    if op in OPS:
        name, rw = OPS[op]
        return "%s%s %s" % (pre, name.split()[0], pre and where or where), rw
    if op & 0xC7 in (0x00, 0x01, 0x02, 0x03) and (op & 0x38) in ALU and op < 0x40:
        mn = ALU[op & 0x38]
        rw = "讀" if (op & 2) else ("讀" if mn == "cmp" else "改")
        return "%s%s %s" % (pre, mn, where), rw
    if op in (0x80, 0x81, 0x83):
        mn = GRP1[reg]
        return "%s%s %s,imm" % (pre, mn, where), ("讀" if mn == "cmp" else "改")
    if op == 0xFE:
        return "%s%s %s" % (pre, GRP_FE.get(reg, "?"), where), "改"
    if op == 0xFF:
        mn = GRP_FF.get(reg, "?")
        return "%s%s %s" % (pre, mn, where), ("改" if reg < 2 else "讀")
    return None


def main():
    mem = pathlib.Path(sys.argv[1]).read_bytes()
    base = int(sys.argv[2], 16)
    for t in sys.argv[3:]:
        disp = int(t, 16)
        pat = bytes([disp & 0xFF, disp >> 8])
        print("== disp16 %04X" % disp)
        n_hit = n_insn = 0
        i = mem.find(pat)
        while i != -1:
            n_hit += 1
            got = decode(mem, i, disp) if i >= 3 else None
            if got:
                n_insn += 1
                txt, rw = got
                # 累加器形式只有一個 opcode 位元組，沒有 modrm。
                if mem[i - 1] in (0xA0, 0xA1, 0xA2, 0xA3):
                    start = i - 1
                else:
                    start = i - (3 if mem[i - 3] in PREFIX else 2)
                print("  %05X  %-8s %-4s %-26s  %s"
                      % (start, mem[start:i + 2].hex(" "), rw, txt,
                         mem[max(0, start - 6):start].hex(" ")))
            i = mem.find(pat, i + 1)
        print("  候選 %d 處，其中 %d 處解得成指令" % (n_hit, n_insn))


main()
