"""#59：固定RAM的可見艦隊數寫入與玩家入口，保留原始指令。"""
import hashlib
import json
import os
import sys
import traceback
import idaapi
import ida_auto
import ida_bytes
import ida_nalt
import ida_pro
import ida_segment
import ida_ua
import idc


def fail(kind, value, tb):
    with open(idc.ARGV[1] + '.error.txt', 'w', encoding='utf-8') as f:
        f.write(''.join(traceback.format_exception(kind, value, tb)))
    ida_pro.qexit(1)


sys.excepthook = fail
ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
assert '/intel59-r60-' in path
with open(path, 'rb') as f:
    raw_file = f.read()
raw = raw_file[256:]
base = ida_segment.getnseg(0).start_ea + 0x100
assert ida_bytes.get_bytes(base, len(raw)) == raw
segment = int(idc.ARGV[2], 16)
rows = []
for part in idc.ARGV[3:]:
    start, end = (int(x, 16) for x in part.split(':'))
    assert 0 <= start < end <= len(raw)
    ida_bytes.del_items(base + start, ida_bytes.DELIT_SIMPLE, end - start)
    ea = base + start
    while ea < base + end:
        size = ida_ua.create_insn(ea)
        assert size > 0
        rows.append({
            'original_location': f'{segment:04X}:{ea-base:04X}',
            'ida_ea_hex': hex(ea), 'runtime_linear_hex': hex(segment*16+ea-base),
            'original_bytes_hex': ida_bytes.get_bytes(ea, size).hex(),
            'operand': idc.generate_disasm_line(ea, 0),
            'semantic': '未審查', 'grade': 'unknown',
            'warning': '附加語意未經審查',
            'source': '研究398；固定原版RAM，來源由凍結收據綁定',
        })
        ea += size
    assert ea == base + end
data = {
    'schema': 'logh3.region-visible-fleets-ida.v1',
    'kernel_version': idaapi.get_kernel_version(), 'uid': os.getuid(),
    'input': path, 'input_sha256': hashlib.sha256(raw_file).hexdigest(),
    'ida_base_hex': hex(base), 'runtime_segment': f'{segment:04X}',
    'ranges': list(idc.ARGV[3:]),
    'address_space': 'IDA COM位址、原版CS:IP與dosgolem線性分欄',
    'instructions': rows,
}
with open(idc.ARGV[1], 'x', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
