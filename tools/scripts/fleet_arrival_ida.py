"""第50輪：保留原名、函式邊界、原始運算元與分級定位。"""
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
assert path.endswith('/normal89-r50-297B-reviewed.com')
base = ida_segment.getnseg(0).start_ea+0x100
semantics = {
    0x8FA: '移動後呼叫0ABF，以來源與目的星域定位',
    0x921: '重算各組的隊形集結點',
    0x96A: '每組沿集結點向中心尋找可站位置',
    0x9B3: '偵察隊集結X取艦隊X', 0x9B9: '偵察隊現在X取艦隊X',
    0x9C6: '偵察隊集結Y取艦隊Y', 0x9CC: '偵察隊現在Y取艦隊Y',
    0xACD: '來源FF跳過入口查找，但仍執行避讓與航向',
    0xAD5: '目的FF跳過入口查找，但仍執行避讓與航向',
    0xAE7: '用來源／目的星域查入口表',
    0xAFF: '入口欄乘16寫FLEET+5C', 0xB0D: '入口列乘20寫FLEET+5E',
    0xB26: '入口不存在則選己方經濟力最高待命行星',
    0xB73: '按地形與同星域其他分隊避讓艦隊中心',
    0xB9D: '以1016-X與1012-Y求查表航向',
    0xBA7: 'FLEET+4A寫現在航向', 0xBAB: 'FLEET+4C寫抵達航向',
    0xBAF: 'FLEET+4B寫目的航向', 0xBBC: 'FLEET+64複製現在X',
    0xBCC: 'FLEET+66複製現在Y',
    0xBED: '先測原中心可站', 0xC51: '依八方向乘圈數環搜候選中心',
    0xC62: '避讓後中心X', 0xC69: '避讓後中心Y',
    0xCA5: '讀地圖可站值', 0xCD0: '只檢查在用艦隊',
    0xCE5: '只檢查同星域艦隊', 0xCF7: '比較原far指標以排除自己',
    0xD49: 'X-16與其他組X做無號比較', 0xD53: 'X+16與其他組X做無號比較',
    0xD63: 'Y-16與其他組Y做無號比較', 0xD73: 'Y+16與其他組Y做無號比較',
    0xDA2: '以SHIP+28集結X作避讓起點', 0xDA9: '以SHIP+2A集結Y作避讓起點',
    0xE57: '沿指向艦隊中心方位每次走4像素',
    0xE7B: '最後可站X寫SHIP+34', 0xE82: '最後可站Y寫SHIP+36',
}
rows, functions = [], []
for start, end in ((0x8D8, 0xA19), (0xABF, 0xBD6), (0xBD6, 0xC74),
                   (0xC74, 0xD99), (0xD99, 0xE8B)):
    assert ida_bytes.get_byte(base+start) == 0x55
    ida_bytes.del_items(base+start, ida_bytes.DELIT_SIMPLE, end-start)
    cur = base+start
    while cur < base+end:
        size = ida_ua.create_insn(cur)
        assert size > 0
        off = cur-base
        proven = off in semantics
        rows.append({'original_location': f'297B:{off:04X}', 'ida_ea_hex': hex(cur),
                     'runtime_linear_hex': hex(0x297B0+off),
                     'original_bytes_hex': ida_bytes.get_bytes(cur, size).hex(),
                     'operand': idc.generate_disasm_line(cur, 0),
                     'semantic': semantics.get(off, '未附加語意'),
                     'grade': 'proven' if proven else 'unknown',
                     'warning': '' if proven else '⚠ 未附加已證實語意',
                     'source': '研究84／235／335及第50輪固定S3完整RAM與watch'})
        cur += size
    assert cur == base+end and rows[-1]['original_bytes_hex'] == 'cb'
    ida_funcs.add_func(base+start, base+end)
    functions.append({'original_location': f'297B:{start:04X}',
                      'original_name': idc.get_name(base+start),
                      'ida_boundary': [hex(base+start), hex(base+end)],
                      'previous_bytes_hex': ida_bytes.get_bytes(base+start-5, 5).hex()})
assert set(semantics) <= {int(r['original_location'].split(':')[1], 16) for r in rows}
data = {'schema': 'logh3.fleet-arrival-ida.v1', 'tool_version': idaapi.IDA_SDK_VERSION,
        'kernel_version': idaapi.get_kernel_version(), 'uid': os.getuid(),
        'input': path, 'input_sha256': hashlib.sha256(open(path, 'rb').read()).hexdigest(),
        'ida_base_hex': hex(base), 'runtime_segment': '297B',
        'address_space': 'IDA COM位址、原版CS:IP及dosgolem線性分欄',
        'functions': functions, 'instructions': rows}
with open(idc.ARGV[1], 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
