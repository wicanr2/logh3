"""第49輪月底閘門；保留原始名稱、位址、bytes及分級。"""
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
import idc


def error(kind, value, tb):
    with open(idc.ARGV[1]+'.error.txt', 'w', encoding='utf-8') as f:
        f.write(''.join(traceback.format_exception(kind, value, tb)))
    ida_pro.qexit(1)


sys.excepthook = error
ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
assert path.endswith('/normal89-r49-0654.com')
base = ida_segment.getnseg(0).start_ea+0x100
assert ida_bytes.get_byte(base+0x33) == 0xCB
semantics = {0x3B: '比較WORK+E0[活動陣營]是否非零',
             0x42: '旗標0直接跳共同尾巴',
             0x4D: '回合除以10，DX為餘數',
             0x53: '餘數非零直接跳共同尾巴',
             0x129: '共同尾巴將本陣營WORK+E0旗標清0'}
start, end = 0x34, 0x130
ida_bytes.del_items(base+start, ida_bytes.DELIT_SIMPLE, end-start)
rows, cur = [], base+start
while cur < base+end:
    size = ida_ua.create_insn(cur)
    assert size > 0
    off = cur-base
    proven = off in semantics
    rows.append({'original_location': f'0654:{off:04X}', 'ida_ea_hex': hex(cur),
                 'runtime_linear_hex': hex(0x6540+off),
                 'original_bytes_hex': ida_bytes.get_bytes(cur, size).hex(),
                 'operand': idc.generate_disasm_line(cur, 0),
                 'semantic': semantics.get(off, '未附加語意'),
                 'grade': 'proven' if proven else 'unknown',
                 'warning': '' if proven else '⚠ 未附加已證實語意',
                 'source': '研究255及normal89-r49自然0／受控1原版雙跑'})
    cur += size
assert cur == base+end and rows[-1]['original_bytes_hex'] == 'cb'
ida_funcs.add_func(base+start, base+end)
data = {'schema': 'logh3.month-seat-gate-ida.v1', 'tool_version': idaapi.IDA_SDK_VERSION,
        'kernel_version': idaapi.get_kernel_version(), 'uid': os.getuid(),
        'input': path, 'input_sha256': hashlib.sha256(open(path, 'rb').read()).hexdigest(),
        'ida_base_hex': hex(base), 'runtime_segment': '0654',
        'address_space': 'IDA COM位址、原版CS:IP及dosgolem線性分欄',
        'previous_bytes_hex': ida_bytes.get_bytes(base+0x2F, 5).hex(),
        'original_name': idc.get_name(base+start),
        'function_boundary': [hex(base+start), hex(base+end)], 'instructions': rows}
with open(idc.ARGV[1], 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
