"""由 IDA 資料庫匯出函式清單與機械特徵，供台帳分類。

只讀資料庫，不改名、不改型別。輸出 /work/functions.json。
IDA 線性位址 = 映像內位移 + 0x10000；dosgolem 線性位址 = 映像內位移 + 0x3490。

每個函式的特徵（全部由指令掃描得出，不含語意推論）：
    callees_far / callees_near   呼叫目標（IDA 位址，去重）
    indirect_calls               以暫存器或記憶體間接呼叫的次數
    ints                         出現的 int 編號（去重）
    io                           in／out 指令數
    vram_imm                     立即值為 A000／A800／B000／B800 的次數（視訊記憶體段）
    stores_nonlocal              目的運算元是記憶體、且不是以 BP 為底的區域變數的指令數
    string_stores                stos／movs 數（寫 es:di）
    flag_ops                     cli／sti／pushf／popf 數
"""
import hashlib
import json
import sys
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_idp
import ida_pro
import ida_ua
import idautils

OUT = "/work/functions.json"
IDA_BASE = 0x10000
GOLEM_BASE = 0x3490
VRAM_SEGS = {0xA000, 0xA800, 0xB000, 0xB800}
# 第一個運算元只讀的助記詞；其餘有記憶體目的運算元者視為寫入
READ_FIRST = {"cmp", "test", "push", "call", "jmp", "mul", "imul", "div", "idiv", "bt", "out", "lea",
              "les", "lds", "fld", "fcom", "fcomp", "fst0", "xlat", "lods", "scas", "cmps"}
# 16 位元定址：phrase 2、3、6 是以 BP 為底（預設 SS 段）
BP_PHRASES = {2, 3, 6}


def fail(kind, value, tb):
    with open(OUT + ".error.txt", "w", encoding="utf-8") as f:
        f.write("".join(traceback.format_exception(kind, value, tb)))
    ida_pro.qexit(1)


sys.excepthook = fail
ida_auto.auto_wait()

funcs_start = set(idautils.Functions())


def func_start(ea):
    f = ida_funcs.get_func(ea)
    return f.start_ea if f else ea


rows = []
for start in sorted(funcs_start):
    f = ida_funcs.get_func(start)
    end = f.end_ea
    data = ida_bytes.get_bytes(start, end - start) or b""
    callers = sum(1 for x in idautils.XrefsTo(start, 0) if x.type in (16, 17, 18, 19))
    far_c, near_c = set(), set()
    insn_offsets = []
    ints, io, vram, ind, stores, sstores, flagops = set(), 0, 0, 0, 0, 0, 0
    for ea in idautils.FuncItems(start):
        insn = ida_ua.insn_t()
        if ida_ua.decode_insn(insn, ea) == 0:
            continue
        insn_offsets.append(ea - start)
        mn = insn.get_canon_mnem()
        ops = [insn.ops[i] for i in range(6) if insn.ops[i].type != ida_ua.o_void]
        if mn == "call" or mn == "callfi" or mn == "callni":
            tgt = [r for r in idautils.CodeRefsFrom(ea, 0)]
            if not tgt or ops[0].type in (ida_ua.o_reg, ida_ua.o_phrase, ida_ua.o_displ, ida_ua.o_mem):
                ind += 1
            for t in tgt:
                (far_c if "callfi" in mn or ida_bytes.get_byte(ea) == 0x9A else near_c).add(func_start(t))
        elif mn == "int":
            ints.add(ops[0].value & 0xFF if ops else -1)
        elif mn in ("in", "out", "insb", "insw", "outsb", "outsw"):
            io += 1
        elif mn in ("cli", "sti", "pushf", "popf"):
            flagops += 1
        elif mn in ("stos", "movs", "stosb", "stosw", "movsb", "movsw"):
            sstores += 1
        for op in ops:
            if op.type == ida_ua.o_imm and (op.value & 0xFFFF) in VRAM_SEGS:
                vram += 1
        if ops and mn not in READ_FIRST and ops[0].type in (ida_ua.o_mem, ida_ua.o_phrase, ida_ua.o_displ):
            if not (ops[0].type in (ida_ua.o_phrase, ida_ua.o_displ) and ops[0].phrase in BP_PHRASES):
                stores += 1
    rows.append({
        "ida_ea": start,
        "image_offset": start - IDA_BASE,
        "golem_linear": start - IDA_BASE + GOLEM_BASE,
        "size": end - start,
        "name": ida_funcs.get_func_name(start),
        "far": bool(f.flags & ida_funcs.FUNC_FAR),
        "lib": bool(f.flags & ida_funcs.FUNC_LIB),
        "thunk": bool(f.flags & ida_funcs.FUNC_THUNK),
        "tail_chunks": len(list(idautils.Chunks(start))) - 1,
        "callers": callers,
        "sha256": hashlib.sha256(data).hexdigest(),
        "callees_far": sorted(far_c),
        "callees_near": sorted(near_c),
        "indirect_calls": ind,
        "ints": sorted(ints),
        "io": io,
        "vram_imm": vram,
        "stores_nonlocal": stores,
        "string_stores": sstores,
        "flag_ops": flagops,
        "insn_offsets": insn_offsets,
    })

with open(OUT, "w", encoding="utf-8") as fh:
    json.dump({"schema": "logh3.functions.v2", "count": len(rows), "functions": rows}, fh)
ida_pro.qexit(0)
