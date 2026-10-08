"""由 IDA 資料庫匯出函式清單（原始位址、大小、near／far、位元組雜湊、呼叫者數）。

只讀資料庫，不改名、不改型別。輸出 /work/functions.json。
IDA 線性位址 = 映像內位移 + 0x10000；dosgolem 線性位址 = 映像內位移 + 0x3490。
"""
import hashlib
import json
import sys
import traceback

import ida_bytes
import ida_auto
import ida_funcs
import ida_pro
import idautils
import idc

OUT = "/work/functions.json"
IDA_BASE = 0x10000
GOLEM_BASE = 0x3490


def fail(kind, value, tb):
    with open(OUT + ".error.txt", "w", encoding="utf-8") as f:
        f.write("".join(traceback.format_exception(kind, value, tb)))
    ida_pro.qexit(1)


sys.excepthook = fail
ida_auto.auto_wait()

rows = []
for start in idautils.Functions():
    f = ida_funcs.get_func(start)
    end = f.end_ea
    data = ida_bytes.get_bytes(start, end - start) or b""
    callers = sum(1 for x in idautils.XrefsTo(start, 0) if x.type in (16, 17, 18, 19))  # fl_CF fl_CN fl_JF fl_JN
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
    })

with open(OUT, "w", encoding="utf-8") as fh:
    json.dump({"schema": "logh3.functions.v1", "count": len(rows), "functions": rows}, fh)
ida_pro.qexit(0)
