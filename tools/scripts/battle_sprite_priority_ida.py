"""第54輪25DB圖示順位；probe保持unknown，graded附加審查索引。"""
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
    '25DB:005F': '每個方向物件佔40 bytes，指標移到下一個物件',
    '25DB:00A5': '一般分隊呼叫0170建立圖示物件及原池順位',
    '25DB:00B8': '最後讀FLEET+38偵察池，不以是否出擊排除',
    '25DB:00CC': '偵察池也呼叫0170初始化',
    '25DB:0178': '建立器參數讀SHIP池索引',
    '25DB:017B': 'SHIP池乘78 bytes',
    '25DB:0198': 'SHIP+3A保存圖示物件segment',
    '25DB:019C': 'SHIP+38保存圖示物件offset',
    '25DB:01A4': '圖示建立先寫SHIP+2F為空白9',
    '25DB:01AC': '圖示建立寫SHIP+30為80順位',
    '25DB:01C9': '對應物件+3也初始化80順位',
    '25DB:01D0': 'SHIP+03分隊byte構成物件word+4高byte',
    '25DB:01DC': 'SHIP+01艦隊索引構成物件word+4低byte',
    '25DB:01E7': '寫物件word+4排序鍵',
    '25DB:0922': '刷新讀SHIP+30順位',
    '25DB:092C': '經SHIP+38遠指標取顯示物件',
    '25DB:0935': '比較原池順位與物件+3',
    '25DB:0939': '順位不同時將SHIP+30寫到物件+3',
    '25DB:0942': '順位變更將物件+1 dirty設FF',
}

def error(kind, value, tb):
    with open(idc.ARGV[1]+'.error.txt', 'w', encoding='utf-8') as f:
        f.write(''.join(traceback.format_exception(kind, value, tb)))
    ida_pro.qexit(1)


sys.excepthook = error
ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
assert '/normal89-r54-' in path
raw_file = open(path, 'rb').read()
raw = raw_file[256:]
base = ida_segment.getnseg(0).start_ea+0x100
assert ida_bytes.get_bytes(base, len(raw)) == raw
segment = int(idc.ARGV[2], 16)
start, end = int(idc.ARGV[3], 16), int(idc.ARGV[4], 16)
graded = len(idc.ARGV) > 5 and idc.ARGV[5] == 'graded'
ida_bytes.del_items(base+start, ida_bytes.DELIT_SIMPLE, end-start)
rows = []
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
                 'semantic': SEMANTICS[location] if proven else '未審查',
                 'grade': 'proven' if proven else 'unknown',
                 'warning': '' if proven else '⚠ 未證實語意；掃描上限不代表函式邊界',
                 'source': '研究392；第54輪原版正常55池建立雙跑及原始指令資料流'})
    ea += size
functions = []
if graded:
    assert segment == 0x25db and start == 7 and end == 0x952 and ea == base+end
    for lo, hi in ((7, 0xfb), (0x170, 0x221), (0x915, 0x952)):
        assert raw[lo-1] == 0xcb and raw[hi-1] == 0xcb
        ida_funcs.add_func(base+lo, base+hi)
        functions.append({'original_location': f'25DB:{lo:04X}',
                          'original_name': idc.get_name(base+lo),
                          'ida_start_hex': hex(base+lo), 'ida_end_hex': hex(base+hi),
                          'grade': 'proven',
                          'boundary_source': '原前支retf、prologue及本支retf／資料流審查'})
for row in rows:
    address = int(row['ida_ea_hex'], 16)
    func = ida_funcs.get_func(address)
    if not graded and func is not None and not any(f['ida_start_hex'] == hex(func.start_ea) for f in functions):
        functions.append({'original_name': idc.get_name(func.start_ea),
                          'ida_start_hex': hex(func.start_ea), 'ida_end_hex': hex(func.end_ea),
                          'grade': 'unknown', 'boundary_source': 'IDA自動邊界；僅供導覽'})
data = {'schema': 'logh3.battle-sprite-priority-ida.v1',
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
