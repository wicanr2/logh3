"""在記憶體 dump 裡找「會寫某個結構位移」的指令。

    tools/py.sh tools/scripts/find_writers.py <dump> <位移十進位> [更多位移...]

掃的是 `mov es:[reg+disp8], al/imm` 這一類常見形狀。**這是掃位元組，
不是反組譯**：同樣的位元組也可能落在資料裡，命中之後要自己看一眼。
比「跑一份收據設監看」便宜非常多，適合先縮小範圍。
"""
import pathlib
import sys

mem = pathlib.Path(sys.argv[1]).read_bytes()
offs = [int(v) for v in sys.argv[2:]]

# 26 = es 前綴；88 /r = mov r/m8,r8；C6 /0 = mov r/m8,imm8
# modrm 的低三位：4=si 5=di 6=bp 7=bx（disp8 形式是 0x40|reg）
REGS = {0x44: 'si', 0x45: 'di', 0x46: 'bp', 0x47: 'bx'}

for off in offs:
    print('== 位移 +0x%02X' % off)
    n = 0
    for i in range(len(mem) - 5):
        if mem[i] != 0x26:
            continue
        op, modrm = mem[i+1], mem[i+2]
        if op == 0x88 and modrm in REGS and mem[i+3] == off:
            print('  %05X  mov es:[%s+%d], al' % (i, REGS[modrm], off))
            n += 1
        elif op == 0xC6 and (modrm & 0xF8) == 0x40 and (modrm & 7) in (4, 5, 6, 7) \
                and mem[i+3] == off:
            print('  %05X  mov byte es:[%s+%d], %02X'
                  % (i, REGS[0x40 | (modrm & 7)], off, mem[i+4]))
            n += 1
    print('  共 %d 處' % n)
