"""第55輪WORK控制原始指令匯出；定位、bytes與語意分欄。"""
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
    '0374:03FB': '星域掃描從參數寫WORK+03',
    '0374:0415': 'WORK+03到32時停止掃描',
    '0374:0447': '當前星域及戰鬥回合參數進12AA:000E',
    '0374:0476': '完成一個星域後WORK+03增加1',
    '12AA:001B': '戰場進入時WORK+02戰鬥旗標設1',
    '12AA:0064': '新戰場的回合FF參數走初始化分支',
    '12AA:006A': '新戰場FF參數轉零起算回合0',
    '12AA:0078': '回合參數保存至WORK+06 word',
    '12AA:00E2': '完整戰術回合返回後WORK+06增加1',
    '12AA:0103': '非零系統命令回傳保留戰鬥旗標',
    '12AA:0107': '正常戰鬥結束清WORK+02旗標',
    '12AA:0279': '战場配置後初始化兩座位鏡頭',
    '12AA:02D9': 'Input選取活動座位，非固定陣營',
    '12AA:030B': '活動座位乘4索引WORK+26鏡頭陣列',
    '12AA:0317': '讀活動座位已保存的鏡頭到目前視窗',
    '12AA:034C': 'Input返回後讀目前鏡頭X',
    '12AA:0359': '目前鏡頭X保存到活動座位WORK+26',
    '12AA:035D': 'Input返回後讀目前鏡頭Y',
    '12AA:0373': '目前鏡頭Y保存到活動座位WORK+28',
    '1588:04AF': '鏡頭欄小於0時限制到0',
    '1588:04BD': '鏡頭列小於0時限制到0',
    '1588:04CC': '鏡頭欄大於100時限制到100',
    '1588:04DB': '鏡頭列大於84時限制到84',
    '1588:0933': '初始化座位0鏡頭',
    '1588:093D': '初始化座位1鏡頭',
    '1588:0951': '座位參數查WORK+11對應陣營',
    '1588:0985': '參戰艦隊FLEET+01與座位陣營比較',
    '1588:098C': '找第一筆對應陣營的參戰艦隊；無匹配尾端語意未知',
    '1588:09B2': '讀FLEET+5C世界X',
    '1588:09B6': 'X減224像素',
    '1588:09BD': 'signed除16朝零截斷得到欄',
    '1588:09C4': '讀FLEET+5E世界Y',
    '1588:09C8': 'Y減180像素',
    '1588:09CF': 'signed除20朝零截斷得到列',
    '1588:09DE': '呼叫04A0限制欄列範圍',
    '1588:09F6': '座位乘4保存欄乘16至WORK+26 X',
    '1588:0A0B': '座位乘4保存列乘20至WORK+28 Y',
    '1261:0078': '先按艦隊編制重算士氣',
    '1261:008C': '按該艦隊陣營重算全方維持費與移動經費',
    '1261:00A8': '維持費高word保存到WORK+60加陣營乘4',
    '1261:00AC': '維持費低word保存到WORK+5E加陣營乘4',
    '1261:00BA': '經費保存後才顯示士氣報告',
    '297B:110A': '跳過FLEET+04為FF的未使用艦隊',
    '297B:1115': '只合計指定陣營FLEET+01',
    '297B:111E': '合計艦隊維持費',
    '297B:1131': '合計艦隊移動經費',
    '297B:1158': '移動經費高word保存到WORK+70加陣營乘4',
    '297B:115C': '移動經費低word保存到WORK+6E加陣營乘4',
    '297B:12B7': '維持費公式結果低word寫FLEET+0A',
    '297B:12E2': '移動經費公式結果低word寫FLEET+0C',
    '297B:12EC': '回讀FLEET+0A維持費word作為回傳值',
    '297B:12EF': '維持費word以signed方式延伸至32位元',
    '297B:1305': '艦隊移動狀態2不列入移動經費',
    '297B:1309': '艦隊移動狀態5不列入移動經費',
    '297B:130D': '其他狀態讀FLEET+0C保存的移動經費',
    '297B:1315': '移動經費word以signed方式延伸至32位元',
}
BOUNDARIES = {
    0x12AA: ((0xE, 0x126), (0x126, 0x1E3), (0x26C, 0x286), (0x2D4, 0x37C)),
    0x1588: ((0x4A0, 0x4EA), (0x930, 0x945), (0x945, 0xA15)),
    0x1261: ((0x67, 0xC8),),
    0x297B: ((0x10DE, 0x116C), (0x116C, 0x12FB), (0x12FB, 0x131C)),
}
INPUT_HASHES = {
    0x12AA: '663a028ded7fdc37bf4bb77429f7d19e6a5348fc4ffdb178d64938cfec5647eb',
    0x1588: '399da320a9429566b9106d915dd0b38fd664fddb13e721e70b82c2392d5b0318',
    0x1261: 'fe397d2987a5659e99e77aa79be64754c3668282bc6ebda282928af0067188c9',
    0x297B: '457a52282d8e4fa72e1f3614220c49f417e6dcc54008ee5163b306ebca429657',
    0x0374: 'f353d6c5922f4392dca2b129b41082089d8166e5aa84fad503bfa3f4406a0203',
}

def error(kind, value, tb):
    with open(idc.ARGV[1]+'.error.txt', 'w', encoding='utf-8') as f:
        f.write(''.join(traceback.format_exception(kind, value, tb)))
    ida_pro.qexit(1)

sys.excepthook = error
ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
assert '/normal89-r55-' in path
raw_file = open(path, 'rb').read()
raw = raw_file[256:]
base = ida_segment.getnseg(0).start_ea+0x100
assert ida_bytes.get_bytes(base, len(raw)) == raw
segment = int(idc.ARGV[2], 16)
assert hashlib.sha256(raw_file).hexdigest() == INPUT_HASHES[segment]
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
                 'source': '研究393；第55輪固定S3原版雙跑、寫入監看及原指令資料流'})
    ea += size
functions = []
if graded:
    for lo, hi in BOUNDARIES.get(segment, ()):
        if not (start <= lo < hi <= end):
            continue
        assert raw[lo-1] == 0xcb and raw[hi-1] == 0xcb
        ida_funcs.add_func(base+lo, base+hi)
        functions.append({'original_location': f'{segment:04X}:{lo:04X}',
                          'original_name': idc.get_name(base+lo),
                          'ida_start_hex': hex(base+lo), 'ida_end_hex': hex(base+hi),
                          'grade': 'proven',
                          'boundary_source': '原前支retf、prologue及本支retf／資料流審查'})
data = {'schema': 'logh3.battle-work-control-ida.v1',
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
