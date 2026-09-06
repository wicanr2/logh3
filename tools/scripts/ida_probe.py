"""三十秒驗證：確認 IDAPython 在這顆 image 上真的跑得起來。

知識庫的硬規則：exit code 不能當證據，唯一可信的訊號是輸出檔。
所以這支只做一件事——數函式、數字串、寫 JSON。
"""
import json, sys, hashlib
import idautils, idaapi, idc, ida_auto, ida_nalt, ida_pro

ida_auto.auto_wait()
out = sys.argv[1] if len(sys.argv) > 1 else '/work/ida_probe.json'
path = ida_nalt.get_input_file_path()
try:
    sha = hashlib.sha256(open(path, 'rb').read()).hexdigest()
except Exception:
    sha = None
data = dict(
    input=path, sha256=sha,
    funcs=len(list(idautils.Functions())),
    segs=[idc.get_segm_name(s) for s in idautils.Segments()],
    strings=sum(1 for _ in idautils.Strings()),
    min_ea=idc.get_inf_attr(idc.INF_MIN_EA),
    max_ea=idc.get_inf_attr(idc.INF_MAX_EA),
)
with open(out, 'w') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
ida_pro.qexit(0)
