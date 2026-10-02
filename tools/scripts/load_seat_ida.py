"""第46輪一次性042F切段；原始名稱與bytes保留，語意附加。入口第46輪goal。"""
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
assert path.endswith('/normal89-r46-042F.com')
base = ida_segment.getnseg(0).start_ea + 0x100
assert ida_bytes.get_byte(base + 5) == 0xCB
ida_bytes.del_items(base + 6, ida_bytes.DELIT_SIMPLE, 0x13A)
cur = base + 6
rows = []
while cur < base + 0x180:
    size = ida_ua.create_insn(cur)
    assert size > 0
    off = cur - base
    raw = ida_bytes.get_bytes(cur, size)
    op = idc.generate_disasm_line(cur, 0)
    # 分級索引只附在原始定位，不改IDA名稱。範圍由三分支動態watch審查。
    proven = '8F2' in op and (0x9F <= off <= 0x111)
    rows.append({'original_location': '042F:%04X' % off, 'ida_ea_hex': hex(cur),
                 'runtime_linear_hex': hex(0x42F0 + off), 'original_bytes_hex': raw.hex(),
                 'operand': op, 'semantic': '座位設定區寫入，三分支watch核對' if proven else '未附加語意',
                 'grade': 'proven' if proven else 'unknown',
                 'warning': '' if proven else '⚠ 未附加已證實語意',
                 'source': 'normal89-r46帝國／同盟／對戰原版收據；第46輪goal'})
    cur += size
    if raw == b'\xcb':
        break
assert rows[-1]['original_bytes_hex'] == 'cb'
ida_funcs.add_func(base + 6, cur)
data = {'schema': 'logh3.load-seat-ida.v1', 'tool_version': idaapi.IDA_SDK_VERSION,
        'kernel_version': idaapi.get_kernel_version(), 'uid': __import__('os').getuid(),
        'input': path, 'input_sha256': hashlib.sha256(open(path, 'rb').read()).hexdigest(),
        'ida_base_hex': hex(base), 'runtime_segment': '042F',
        'address_space': 'IDA COM位址、原版CS:IP及dosgolem線性分欄',
        'original_name': idc.get_name(base + 6), 'function_boundary': [hex(base+6), hex(cur)],
        'functions': len(list(idautils.Functions())), 'instructions': rows}
with open(idc.ARGV[1], 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
