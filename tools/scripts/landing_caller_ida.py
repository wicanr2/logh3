"""第110輪：計時等待返回與單格顯示提交的原始指令。"""
import hashlib
import json
import os
import idaapi
import ida_auto
import ida_bytes
import ida_nalt
import ida_pro
import ida_segment
import ida_ua
import idc

ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
assert '/landing48-r110-' in path
data = open(path, 'rb').read()
raw = data[256:]
base = ida_segment.getnseg(0).start_ea + 0x100
assert ida_bytes.get_bytes(base, len(raw)) == raw
segment = int(idc.ARGV[2], 16)
rows = []
for span in idc.ARGV[3:]:
    first, last = (int(x, 16) for x in span.split(':'))
    assert 0 <= first < last <= len(raw)
    ida_bytes.del_items(base + first, ida_bytes.DELIT_SIMPLE, last - first)
    ea = base + first
    while ea < base + last:
        n = ida_ua.create_insn(ea)
        assert n > 0 and ea + n <= base + last, hex(ea)
        rows.append({'original_location': f'{segment:04X}:{ea-base:04X}',
                     'original_name': idc.get_name(ea), 'ida_ea_hex': hex(ea),
                     'runtime_linear_hex': hex(segment * 16 + ea - base),
                     'original_bytes_hex': ida_bytes.get_bytes(ea, n).hex(),
                     'operand': idc.generate_disasm_line(ea, 0),
                     'grade': 'unknown', 'warning': '有界匯出，函式語意另審'})
        ea += n
out = {'schema': 'logh3.landing-frame-ida.v1', 'kernel_version': idaapi.get_kernel_version(),
       'uid': os.getuid(), 'input': path, 'input_sha256': hashlib.sha256(data).hexdigest(),
       'ida_base_hex': hex(base), 'runtime_segment': f'{segment:04X}',
       'address_space': '原版CS:IP、IDA EA、dosgolem線性位址分欄',
       'ranges': list(idc.ARGV[3:]), 'instructions': rows}
with open(idc.ARGV[1], 'x', encoding='utf-8') as f:
    json.dump(out, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
