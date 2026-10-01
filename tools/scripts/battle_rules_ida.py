"""#53 的一次性 IDA 原始指令證據；只接受 re53- 切段，禁止改名。

tools/ida.sh raw idat -A -S/idascripts/battle_rules_ida.py... /work/re53-....com
每個輸出同列保留 CS:IP、IDA 位址、線性位址、bytes、原運算元與證據等級。
"""
import hashlib
import json
import sys
import traceback
import ida_auto
import idaapi
import ida_bytes
import ida_funcs
import ida_nalt
import ida_pro
import ida_segment
import ida_ua
import idautils
import idc

out = idc.ARGV[1]
def error(exc_type, exc, tb):
    with open(out+'.error.txt', 'w', encoding='utf-8') as f:
        f.write(''.join(traceback.format_exception(exc_type, exc, tb)))
    ida_pro.qexit(1)
sys.excepthook = error
ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
if '/re53-' not in path:
    raise ValueError('僅接受 #53 一次性切段資料庫')
raw = open(path, 'rb').read()
source_hash = hashlib.sha256(raw).hexdigest()
base = ida_segment.getnseg(0).start_ea + 0x100
segment = int(idc.ARGV[2], 16)
report = {'schema': 'logh3.battle-rules-ida.v1', 'input': path, 'input_sha256': source_hash,
          'tool': 'IDA Pro '+idaapi.get_kernel_version(), 'ida_base_hex': hex(base),
          'address_space': 'IDA .COM 載入位址；CS:IP 與 dosgolem 線性位址另列',
          'functions': []}
args = idc.ARGV[3:]
index_paths = [a.removeprefix('--index=') for a in args if a.startswith('--index=')]
if len(index_paths) != 1:
    raise ValueError('必須指定經審查語意索引 --index=，不得從散文自動提升語意')
index_raw = open(index_paths[0], 'rb').read()
index = json.loads(index_raw)
if index['inputs'].get('%04X' % segment) != source_hash:
    raise ValueError('語意索引與切段雜湊不符')
report['semantic_index'] = {'path': index_paths[0], 'sha256': hashlib.sha256(index_raw).hexdigest()}
entries = {(e['kind'], e['original_location']):e for e in index['entries']}
def annotate(record, kind):
    entry = entries.get((kind, record['original_location']))
    if entry is None:
        return
    if entry['grade'] not in ('proven', 'strong inference', 'hypothesis', 'unknown'):
        raise ValueError('未分級語意')
    if kind == 'instruction' and entry['original_bytes_hex'] != record['original_bytes_hex']:
        raise ValueError('語意索引與原始指令不符')
    for key in ('semantic', 'grade', 'source'):
        record[key] = entry[key]
    record['warning'] = '' if entry['grade'] == 'proven' else '⚠ 未證實語意'
for query in (a for a in args if not a.startswith('--index=')):
    off, length = [int(v,16) for v in query.split(':')]
    start, end = base+off, base+off+length
    ida_bytes.del_items(start, ida_bytes.DELIT_SIMPLE, length)
    cur = start
    while cur < end:
        instruction = ida_ua.insn_t()
        size = ida_ua.decode_insn(instruction,cur)
        if not size or ida_ua.create_insn(cur) != size:
            raise ValueError('指令解碼失敗 %X' % cur)
        cur += size
    ida_funcs.add_func(start, end)
    ida_auto.auto_wait()
    f = ida_funcs.get_func(start)
    record = {'original_location': '%04X:%04X' % (segment,off),
              'original_name': idc.get_name(start), 'ida_ea_hex': hex(start),
              'runtime_linear_hex': hex(segment*16+off),
              'boundary_ida': [hex(f.start_ea), hex(f.end_ea)] if f else None,
              'boundary_origin': '明示查詢範圍，僅導覽；須以原始 retf 及控制流審查',
              'semantic': '未審查', 'grade': 'unknown', 'warning': '⚠ 未證實語意',
              'source': source_hash, 'instructions': [],
              'xrefs_to': [{'ida_ea_hex': hex(x.frm),'type':x.type} for x in idautils.XrefsTo(start)]}
    annotate(record, 'function')
    for ea in idautils.Heads(start,end):
        size = idc.get_item_size(ea)
        record['instructions'].append({'original_location':'%04X:%04X' % (segment,ea-base),
            'ida_ea_hex':hex(ea), 'runtime_linear_hex':hex(segment*16+ea-base),
            'original_bytes_hex': (ida_bytes.get_bytes(ea,size) or b'').hex(' '),
            'original_operand':idc.GetDisasm(ea), 'semantic':'未審查', 'grade':'unknown',
            'warning':'⚠ 未證實語意', 'source':source_hash})
        annotate(record['instructions'][-1], 'instruction')
    report['functions'].append(record)
with open(out,'w',encoding='utf-8') as f:
    json.dump(report,f,ensure_ascii=False,indent=2)
ida_pro.qexit(0)
