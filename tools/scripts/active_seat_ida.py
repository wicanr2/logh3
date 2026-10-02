"""第48輪座位迴圈原始定位；語意附加不取代IDA原始名稱。"""
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
assert path.endswith('/normal89-r48-04F8.com')
base = ida_segment.getnseg(0).start_ea+0x100
assert ida_bytes.get_byte(base+0x18F) == 0xCB
semantics = {0x191: '人類座位迴圈設活動座位0', 0x1A3: '人類座位迴圈設活動座位1',
             0x1C8: '從活動座位的陣營映射寫入WORK+09', 0x2E7: '進電腦座位迴圈前WORK+09設FF',
             0x2F7: '電腦座位迴圈設活動座位0', 0x301: '電腦座位迴圈設活動座位1',
             0x319: '從活動座位的陣營映射寫入WORK+09'}
rows, functions = [], []
for start, end in ((0x190, 0x1B9), (0x1B9, 0x237), (0x2E7, 0x30E), (0x30E, 0x355)):
    ida_bytes.del_items(base+start, ida_bytes.DELIT_SIMPLE, end-start)
    cur = base+start
    while cur < base+end:
        size = ida_ua.create_insn(cur)
        assert size > 0
        off = cur-base
        raw = ida_bytes.get_bytes(cur, size)
        proven = off in semantics
        rows.append({'original_location': f'04F8:{off:04X}', 'ida_ea_hex': hex(cur),
                     'runtime_linear_hex': hex(0x4F80+off), 'original_bytes_hex': raw.hex(),
                     'operand': idc.generate_disasm_line(cur, 0),
                     'semantic': semantics.get(off, '未附加語意'),
                     'grade': 'proven' if proven else 'unknown',
                     'warning': '' if proven else '⚠ 未附加已證實語意',
                     'source': '研究186／187及normal89-r48正常原版雙跑watch／完整RAM'})
        cur += size
    assert cur == base+end and rows[-1]['original_bytes_hex'] == 'cb'
    ida_funcs.add_func(base+start, base+end)
    functions.append({'original_location': f'04F8:{start:04X}',
                      'original_name': idc.get_name(base+start),
                      'ida_boundary': [hex(base+start), hex(base+end)]})
data = {'schema': 'logh3.active-seat-ida.v1', 'tool_version': idaapi.IDA_SDK_VERSION,
        'kernel_version': idaapi.get_kernel_version(), 'uid': os.getuid(),
        'input': path, 'input_sha256': hashlib.sha256(open(path, 'rb').read()).hexdigest(),
        'ida_base_hex': hex(base), 'runtime_segment': '04F8',
        'address_space': 'IDA COM位址、原版CS:IP及dosgolem線性分欄',
        'previous_human_loop_bytes_hex': ida_bytes.get_bytes(base+0x18B, 5).hex(),
        'functions': functions, 'instructions': rows}
with open(idc.ARGV[1], 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
