"""第47輪原版0374主迴圈，保留原始名稱／定位與分級附加語意。"""
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
    with open(idc.ARGV[1] + '.error.txt', 'w', encoding='utf-8') as f:
        f.write(''.join(traceback.format_exception(kind, value, tb)))
    ida_pro.qexit(1)


sys.excepthook = error
ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
assert path.endswith('/normal89-r47-0374.com')
base = ida_segment.getnseg(0).start_ea + 0x100
assert ida_bytes.get_byte(base + 0x342) == 0xCB
assert ida_bytes.get_bytes(base + 0x3D9, 4) == bytes.fromhex('ff06188f')
ida_bytes.del_items(base + 0x343, ida_bytes.DELIT_SIMPLE, 0xA5)
cur = base + 0x343
rows = []
semantics = {0x3D9: '全部戰場與回合尾段返回且未結局後，WORK+04戰略回合加1'}
while cur < base + 0x3E8:
    size = ida_ua.create_insn(cur)
    assert size > 0
    off = cur - base
    raw = ida_bytes.get_bytes(cur, size)
    proven = off in semantics
    rows.append({'original_location': '0374:%04X' % off,
                 'ida_ea_hex': hex(cur), 'runtime_linear_hex': hex(0x3740 + off),
                 'original_bytes_hex': raw.hex(),
                 'operand': idc.generate_disasm_line(cur, 0),
                 'semantic': semantics.get(off, '未附加語意'),
                 'grade': 'proven' if proven else 'unknown',
                 'warning': '' if proven else '⚠ 未附加已證實語意',
                 'source': '研究251與normal89-r47原版雙跑watch／完整RAM'})
    cur += size
assert cur == base + 0x3E8
assert rows[-1]['original_bytes_hex'] == 'e97dff'
ida_funcs.add_func(base + 0x343, cur)
data = {'schema': 'logh3.turn-counter-ida.v1', 'tool_version': idaapi.IDA_SDK_VERSION,
        'kernel_version': idaapi.get_kernel_version(), 'uid': os.getuid(),
        'input': path, 'input_sha256': hashlib.sha256(open(path, 'rb').read()).hexdigest(),
        'ida_base_hex': hex(base), 'runtime_segment': '0374',
        'address_space': 'IDA COM位址、原版CS:IP及dosgolem線性分欄',
        'original_name': idc.get_name(base + 0x343),
        'function_boundary': [hex(base + 0x343), hex(cur)],
        'previous_bytes_hex': ida_bytes.get_bytes(base + 0x33E, 5).hex(),
        'instructions': rows}
with open(idc.ARGV[1], 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
