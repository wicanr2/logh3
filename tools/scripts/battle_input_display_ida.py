"""第51輪：保存純移動摘要回呼的原始定位；tools/ida.sh raw 執行。"""
import hashlib
import json
import os
import traceback
import sys
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

# 受版控位址索引；保留原運算元，語意另附。依研究389審原bytes及動態雙跑。
SEMANTICS = {
    '12F1:04FA': '比較原摘要物件快取與事件物件編號',
    '12F1:04FC': '同物件跳過摘要重畫',
    '12F1:051F': '呼叫分隊摘要包裝函式1CC5:0009',
    '12F1:0530': '記錄新的摘要物件編號',
    '12F1:07D1': '清除現在及集結點反白物件',
    '12F1:07D6': '摘要物件快取恢復FFFF',
    '1CC5:0030': '比較SHIP陣營與活動陣營',
    '1CC5:0049': '己方分支呼叫完整分隊摘要008E',
    '1CC5:0061': '另一分支呼叫04CC，動態敵方文字未驗',
    '1CC5:0068': '另一分支不建立集結點物件',
    '1CC5:0071': '呼叫現在位置與選擇性的集結點反白建立器',
    '25DB:0BF2': '現在位置物件優先FF',
    '25DB:0C22': '現在位置物件圖塊10',
    '25DB:0C43': '現在位置讀SHIP+34',
    '25DB:0C4A': '現在位置讀SHIP+36',
    '25DB:0C51': '判斷是否建立第二個集結點物件',
    '25DB:0C8E': '集結點物件優先FE',
    '25DB:0CA5': '集結點物件圖塊10',
    '25DB:0CCD': '非旗艦集結點讀SHIP+28',
    '25DB:0CD4': '非旗艦集結點讀SHIP+2A',
    '25DB:0CFA': '旗艦集結點讀FLEET+64',
    '25DB:0D03': '旗艦集結點讀FLEET+66',
}


def error(kind, value, tb):
    with open(idc.ARGV[1]+'.error.txt', 'w', encoding='utf-8') as f:
        f.write(''.join(traceback.format_exception(kind, value, tb)))
    ida_pro.qexit(1)


sys.excepthook = error
ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
segment = idc.ARGV[2]
assert segment in ('12F1', '1CC5', '25DB')
assert '/normal89-r51-' in path
raw = open(path, 'rb').read()[256:]
# 依DOS loader的首段基底加前置256 bytes，再逐byte核對，不能猜固定基底。
base = ida_segment.getnseg(0).start_ea+0x100
assert ida_bytes.get_bytes(base, len(raw)) == raw
runtime = int(segment, 16)*16
rows, functions = [], []
boundaries = {'12F1': [(0x4c3, 0x539), (0x539, 0x7e3)],
              '1CC5': [(9, 0x8e)], '25DB': [(0xaf8, 0xbd2), (0xbd2, 0xd11)]}[segment]
for start, end in boundaries:
    assert raw[start] == 0x55 and raw[end-1] == 0xcb
    assert raw[start-1] == 0xcb
    ida_bytes.del_items(base+start, ida_bytes.DELIT_SIMPLE, end-start)
    ea = base+start
    while ea < base+end:
        size = ida_ua.create_insn(ea)
        assert size > 0
        location = f'{segment}:{ea-base:04X}'
        proven = location in SEMANTICS
        rows.append({'original_location': location,
                     'ida_ea_hex': hex(ea), 'runtime_linear_hex': hex(runtime+ea-base),
                     'original_bytes_hex': ida_bytes.get_bytes(ea, size).hex(),
                     'operand': idc.generate_disasm_line(ea, 0),
                     'semantic': SEMANTICS.get(location, '未附加語意'),
                     'grade': 'proven' if proven else 'unknown',
                     'warning': '' if proven else '⚠ 原始指令，語意待資料流審查',
                     'source': '研究389；第51輪純移動雙跑RAM、原指令資料流'})
        ea += size
    assert ea == base+end
    ida_funcs.add_func(base+start, base+end)
    functions.append({'original_location': f'{segment}:{start:04X}',
                      'original_name': idc.get_name(base+start),
                      'ida_boundary': [hex(base+start), hex(base+end)],
                      'previous_bytes_hex': ida_bytes.get_bytes(base+start-5, 5).hex()})
data = {'schema': 'logh3.battle-input-display-ida.v1',
        'kernel_version': idaapi.get_kernel_version(), 'uid': os.getuid(),
        'input': path, 'input_sha256': hashlib.sha256(open(path, 'rb').read()).hexdigest(),
        'ida_base_hex': hex(base), 'runtime_segment': segment,
        'address_space': 'IDA COM位址、原版CS:IP及dosgolem線性分欄',
        'functions': functions, 'instructions': rows}
with open(idc.ARGV[1], 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
