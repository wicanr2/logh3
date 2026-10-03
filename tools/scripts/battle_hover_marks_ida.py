"""第52輪小地圖摘要標記；探針保持unknown，graded才合併審查索引。"""
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

SEMANTICS = {
    '1727:067F': '旗艦航線讀FLEET+5C現在X',
    '1727:068E': '旗艦航線讀FLEET+5E現在Y',
    '1727:069B': '旗艦航線讀FLEET+64集結X',
    '1727:06A9': '旗艦航線讀FLEET+66集結Y',
    '1727:06D1': '旗艦航線XOR色號4',
    '1727:06E5': '呼叫39D7:0936畫實線，前一參數FFFF',
    '1727:071E': '退出時同參數再XOR擦除旗艦航線',
    '1727:073C': '一般分隊讀SHIP+28集結X',
    '1727:074B': '一般分隊讀SHIP+2A集結Y',
    '1727:0758': '一般分隊集結點XOR色號4',
    '1727:075E': '呼叫39D7:08F4點集結點',
    '1727:0780': '退出時同參數再XOR擦除分隊集結點',
}


def error(kind, value, tb):
    with open(idc.ARGV[1]+'.error.txt', 'w', encoding='utf-8') as f:
        f.write(''.join(traceback.format_exception(kind, value, tb)))
    ida_pro.qexit(1)


sys.excepthook = error
ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
assert '/normal89-r52-' in path
raw_file = open(path, 'rb').read()
raw = raw_file[256:]
base = ida_segment.getnseg(0).start_ea+0x100
assert ida_bytes.get_bytes(base, len(raw)) == raw
segment = int(idc.ARGV[2], 16)
start, end = int(idc.ARGV[3], 16), int(idc.ARGV[4], 16)
ida_bytes.del_items(base+start, ida_bytes.DELIT_SIMPLE, end-start)
rows = []
graded = len(idc.ARGV) > 5 and idc.ARGV[5] == 'graded'
ea = base+start
while ea < base+end:
    size = ida_ua.create_insn(ea)
    assert size > 0
    location = f'{segment:04X}:{ea-base:04X}'
    proven = graded and location in SEMANTICS
    rows.append({'original_location': location,
                 'ida_ea_hex': hex(ea), 'runtime_linear_hex': hex(segment*16+ea-base),
                 'original_bytes_hex': ida_bytes.get_bytes(ea, size).hex(),
                 'operand': idc.generate_disasm_line(ea, 0),
                 'semantic': SEMANTICS[location] if proven else '未附加語意',
                 'grade': 'proven' if proven else 'unknown',
                 'warning': '' if proven else '⚠ 原始指令，語意待審查',
                 'source': '研究390；第52輪原版八組獨立雙跑、原指令資料流'})
    ea += size
# 探針終點只是掃描上限，不能將一個指令中間的值當函式邊界。
functions = []
if graded:
    assert segment == 0x1727 and start == 0x677 and end == 0x792 and ea == base+end
    for lo, hi in ((0x677, 0x6f0), (0x6f0, 0x736), (0x736, 0x767), (0x767, 0x792)):
        assert raw[lo-1] == 0xcb and raw[hi-1] == 0xcb
        ida_funcs.add_func(base+lo, base+hi)
        functions.append({'original_location': f'1727:{lo:04X}',
                          'original_name': idc.get_name(base+lo),
                          'ida_boundary': [hex(base+lo), hex(base+hi)],
                          'previous_bytes_hex': raw[lo-5:lo].hex()})
data = {'schema': 'logh3.battle-hover-marks-ida.v1',
        'kernel_version': idaapi.get_kernel_version(), 'uid': os.getuid(),
        'input': path, 'input_sha256': hashlib.sha256(raw_file).hexdigest(),
        'ida_base_hex': hex(base), 'runtime_segment': f'{segment:04X}',
        'scan_start_hex': hex(start), 'requested_limit_hex': hex(end),
        'actual_instruction_end_hex': hex(ea-base),
        'address_space': 'IDA COM位址、原版CS:IP及dosgolem線性分欄',
        'functions': functions, 'instructions': rows}
with open(idc.ARGV[1], 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
