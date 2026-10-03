"""#60：由固定原版RAM切段，保存窄幅指令與原始定位。"""
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
assert '/hold60-r58-' in path
with open(path, 'rb') as f:
    raw_file = f.read()
raw = raw_file[256:]
base = ida_segment.getnseg(0).start_ea + 0x100
assert ida_bytes.get_bytes(base, len(raw)) == raw
segment = int(idc.ARGV[2], 16)
start, end = int(idc.ARGV[3], 16), int(idc.ARGV[4], 16)
assert 0 <= start < end <= len(raw)
ida_bytes.del_items(base + start, ida_bytes.DELIT_SIMPLE, end - start)
rows = []
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
        'warning': '掃描範圍不是函式邊界',
        'source': '研究396；固定原版RAM切段，來源由外部收據綁定',
    })
    ea += size
data = {
    'schema': 'logh3.left-hold-ida.v1',
    'kernel_version': idaapi.get_kernel_version(), 'uid': os.getuid(),
    'input': path, 'input_sha256': hashlib.sha256(raw_file).hexdigest(),
    'ida_base_hex': hex(base), 'runtime_segment': f'{segment:04X}',
    'scan_start_hex': hex(start), 'requested_limit_hex': hex(end),
    'actual_instruction_end_hex': hex(ea-base),
    'address_space': 'IDA COM位址、原版CS:IP與dosgolem線性分欄',
    'instructions': rows,
}
with open(idc.ARGV[1], 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
