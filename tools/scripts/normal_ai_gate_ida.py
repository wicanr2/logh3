"""#89 一次性切段資料庫，輸出原始定位與非破壞性語意。入口：第45輪 goal。"""
import hashlib
import json
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

def error(kind, value, tb):
    with open(idc.ARGV[1] + '.error.txt', 'w', encoding='utf-8') as f:
        f.write(''.join(traceback.format_exception(kind, value, tb)))
    ida_pro.qexit(1)
sys.excepthook = error
ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
assert path.endswith('/normal89-r45-2D10-check.com')
base = ida_segment.getnseg(0).start_ea + 0x100
ida_bytes.del_items(base + 5, ida_bytes.DELIT_SIMPLE, 0x33)
cur = base + 5
rows = []
semantics = {
    5: ('WORK+32／DS:8F46 是否非零', 'proven'),
    0xA: ('非零直接跳到清除旗標，略過所有 AI 呼叫', 'proven'),
    0x30: ('AI 正常完成或跳過時均清除旗標', 'proven'),
}
while cur < base + 0x38:
    size = ida_ua.create_insn(cur)
    assert size > 0
    off = cur - base
    semantic, grade = semantics.get(off, ('保留指令，未附加語意', 'unknown'))
    rows.append({'original_location': '2D10:%04X' % off,
                 'ida_ea_hex': hex(cur), 'runtime_linear_hex': hex(0x2D100 + off),
                 'original_bytes_hex': ida_bytes.get_bytes(cur, size).hex(),
                 'operand': idc.generate_disasm_line(cur, 0),
                 'semantic': semantic, 'grade': grade,
                 'warning': '' if grade == 'proven' else '⚠ 未附加已證實語意',
                 'source': 'docs/re/175；本次切段 bytes；normal89-r45-natural／zero 動態收據'})
    cur += size
ida_funcs.add_func(base + 5, base + 0x38)
data = {'schema': 'logh3.normal-ai-gate-ida.v1', 'tool_version': idaapi.IDA_SDK_VERSION,
        'kernel_version': idaapi.get_kernel_version(), 'uid': __import__('os').getuid(),
        'input': path, 'input_sha256': hashlib.sha256(open(path, 'rb').read()).hexdigest(),
        'ida_base_hex': hex(base), 'runtime_segment': '2D10',
        'address_space': 'IDA COM 位址；原版 CS:IP 與線性位址分欄',
        'functions': len(list(idautils.Functions())),
        'original_function_name': idc.get_name(base + 5),
        'boundary': [hex(base + 5), hex(base + 0x38)],
        'boundary_evidence': '前一支 0004 retf；本支 0037 retf；0038 是下一支 push bp',
        'xrefs_to_entry': [{'ida_ea_hex': hex(x.frm), 'type': x.type} for x in idautils.XrefsTo(base + 5)],
        'instructions': rows}
with open(idc.ARGV[1], 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
