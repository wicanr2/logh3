"""#85 的一次性 IDA 資料庫查詢；入口 docs/re/354。

只在由記憶體切出的新 .COM 上建指令與函式；保留原始名稱／位址／bytes。
段值及查詢範圍由命令列指定，語意未審查時醒目標示 unknown。
"""
import hashlib
import json
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_nalt
import ida_pro
import ida_segment
import ida_ua
import idautils
import idc

out = idc.ARGV[1]
# 發生腳本錯誤時仍留下可判讀的診斷，不能把無輸出當成功。
def error(exc_type, exc, tb):
    with open(out + '.error.txt', 'w', encoding='utf-8') as fh:
        fh.write(''.join(traceback.format_exception(exc_type, exc, tb)))
    ida_pro.qexit(1)
import sys
sys.excepthook = error
ida_auto.auto_wait()
segment = int(idc.ARGV[2], 16)
base = ida_segment.getnseg(0).start_ea + 0x100
args = idc.ARGV[3:]
index_paths = [a.removeprefix('--index=') for a in args if a.startswith('--index=')]
if len(index_paths) != 1:
    raise ValueError('必須明示經審查的語意索引 --index=；不可從散文自動提升語意')
index_raw = open(index_paths[0], 'rb').read()
index = json.loads(index_raw)
queries = [tuple(int(x, 16) for x in a.split(':')) for a in args if not a.startswith('--index=')]
for off, length in queries:
    cur, end = base + off, base + off + length
    # 只針對本腳本新建的 re85 切段資料庫重新定義資料；原始 bytes 保留。
    # 不接受正式資料庫，避免資料陣列擋住 create_insn 卻輸出空結果。
    if '/re85-' not in ida_nalt.get_input_file_path():
        raise ValueError('只允許 #85 一次性切段輸入')
    ida_bytes.del_items(cur, ida_bytes.DELIT_SIMPLE, length)
    while cur < end:
        insn = ida_ua.insn_t()
        size = ida_ua.decode_insn(insn, cur)
        if not size:
            raise ValueError('無法建立原始指令 %X' % cur)
        if ida_ua.create_insn(cur) != size:
            raise ValueError('原始 bytes 可解碼但無法建立指令 %X' % cur)
        cur += size
    ida_funcs.add_func(base + off, end)
ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
report = {'input': path, 'sha256': hashlib.sha256(open(path, 'rb').read()).hexdigest(),
          'tool': 'IDA Pro 9.4（locked-v1 image；版本由來源工具契約限定）',
          'ida_base_hex': hex(base), 'runtime_segment_hex': hex(segment),
          'address_space': 'IDA .COM 載入位址；CS:IP 與 dosgolem 線性位址另列，不混用',
          'functions': []}
if index['inputs'].get('%04X' % segment) != report['sha256']:
    raise ValueError('語意索引與原始輸入雜湊不符')
report['semantic_index'] = {'path': index_paths[0], 'sha256': hashlib.sha256(index_raw).hexdigest(),
                            'source': index['source']}
entries = {(e['kind'], e['original_location']): e for e in index['entries']}
def annotate(record, kind):
    entry = entries.get((kind, record['original_location']))
    if entry is None:
        return
    if entry['grade'] not in ('proven', 'strong inference', 'hypothesis', 'unknown'):
        raise ValueError('未分級的語意')
    if kind == 'instruction' and entry['original_bytes_hex'] != record['original_bytes_hex']:
        raise ValueError('原始指令與語意索引 bytes 不符')
    for key in ('semantic', 'grade', 'source'):
        record[key] = entry[key]
    record['warning'] = '' if entry['grade'] == 'proven' else '⚠ 未證實語意'
for off, length in queries:
    f = ida_funcs.get_func(base + off)
    record = {'original_location': '%04X:%04X' % (segment, off),
              'ida_ea_hex': hex(base + off), 'runtime_linear_hex': hex(segment * 16 + off),
              'original_name': idc.get_name(base + off),
              'boundary_ida': [hex(f.start_ea), hex(f.end_ea)] if f else None,
              'boundary_origin': '一次性資料庫的指定查詢範圍；僅導覽，須以 retf／控制流審查',
              'semantic': '尚未審查；只保存原始指令與資料庫關係', 'grade': 'unknown',
              'warning': '⚠ 未證實語意', 'source': '本檔固定雜湊的原始切段', 'instructions': [],
              'xrefs_to': [{'ida_ea_hex': hex(x.frm), 'type': x.type} for x in idautils.XrefsTo(base + off)]}
    annotate(record, 'function')
    for ea in idautils.Heads(base + off, base + off + length):
        size = idc.get_item_size(ea)
        record['instructions'].append({'original_location': '%04X:%04X' % (segment, ea - base),
            'ida_ea_hex': hex(ea), 'runtime_linear_hex': hex(segment * 16 + ea - base),
            'original_bytes_hex': (ida_bytes.get_bytes(ea, size) or b'').hex(' '),
            'original_operand': idc.GetDisasm(ea), 'semantic': '未審查', 'grade': 'unknown',
            'warning': '⚠ 未證實語意', 'source': report['sha256']})
        annotate(record['instructions'][-1], 'instruction')
    report['functions'].append(record)
with open(out, 'w', encoding='utf-8') as fh:
    json.dump(report, fh, ensure_ascii=False, indent=2)
ida_pro.qexit(0)
