"""交戰下方場景的精確原指令，保留原定位及工具位址空間。"""
import hashlib
import json
import os
import sys
import traceback
import idaapi
import ida_auto
import ida_bytes
import ida_funcs
import ida_nalt
import ida_pro
import ida_segment
import ida_ua
import idautils
import idc


def fail(kind, value, tb):
    with open(idc.ARGV[1]+'.error.txt', 'w', encoding='utf-8') as f:
        f.write(''.join(traceback.format_exception(kind, value, tb)))
    ida_pro.qexit(1)


sys.excepthook = fail
ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
assert '/blocks48-r70-' in path
with open(path, 'rb') as f:
    data = f.read()
raw = data[256:]
base = ida_segment.getnseg(0).start_ea+0x100
assert ida_bytes.get_bytes(base, len(raw)) == raw
segment = int(idc.ARGV[2], 16)
rows = []
for part in idc.ARGV[3:]:
    first, last = part.split(':')
    start = int(first, 16)
    if last == 'ret':
        # 有界線性診斷匯出；返回點不是經審查的整個函式邊界。
        end = min(start+2048, len(raw))
    elif last == 'auto':
        fn = ida_funcs.get_func(base+start)
        if fn is None or fn.start_ea != base+start:
            # 原版執行斷點已確認入口；COM 匯入不會保留原段的函式入口。
            ida_ua.create_insn(base+start)
            ida_funcs.add_func(base+start)
            ida_auto.auto_wait()
            fn = ida_funcs.get_func(base+start)
        assert fn is not None and fn.start_ea == base+start
        end = fn.end_ea-base
    else:
        end = int(last, 16)
    assert 0 <= start < end <= len(raw)
    ida_bytes.del_items(base+start, ida_bytes.DELIT_SIMPLE, end-start)
    ea = base+start
    while ea < base+end:
        size = ida_ua.create_insn(ea)
        if not (size > 0 and ea+size <= base+end):
            with open(idc.ARGV[1]+'.error.txt', 'w', encoding='utf-8') as f:
                f.write(f'original {segment:04X}:{ea-base:04X}, size {size}, requested end {end:04X}\n')
                f.write(ida_bytes.get_bytes(ea, 16).hex()+'\n')
                f.write(json.dumps(rows[-8:], ensure_ascii=False, indent=2))
            ida_pro.qexit(1)
        rows.append({'original_location': f'{segment:04X}:{ea-base:04X}',
                     'original_name': idc.get_name(ea), 'ida_ea_hex': hex(ea),
                     'runtime_linear_hex': hex(segment*16+ea-base),
                     'original_bytes_hex': ida_bytes.get_bytes(ea, size).hex(),
                     'operand': idc.generate_disasm_line(ea, 0),
                     'xrefs': [{'ida_from_hex': hex(x.frm), 'type': x.type}
                               for x in idautils.XrefsTo(ea)],
                     'grade': 'unknown', 'semantic': '未審查',
                     'warning': '區段匯出不代表全部函式邊界已證實', 'source': '第70輪固定原版RAM'})
        ea += size
        if last == 'ret' and rows[-1]['original_bytes_hex'][:2] in ('cb', 'ca', 'c3', 'c2'):
            break
out = {'schema': 'logh3.fleet-blocks-ida.v1', 'kernel_version': idaapi.get_kernel_version(),
       'uid': os.getuid(), 'input': path, 'input_sha256': hashlib.sha256(data).hexdigest(),
       'ida_base_hex': hex(base), 'runtime_segment': f'{segment:04X}',
       'address_space': '原版CS:IP、IDA EA、dosgolem線性位址分欄',
       'ranges': list(idc.ARGV[3:]), 'instructions': rows}
with open(idc.ARGV[1], 'x', encoding='utf-8') as f:
    json.dump(out, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
