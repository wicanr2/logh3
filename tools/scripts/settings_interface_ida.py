"""第98輪設定選單的一次性 COM 資料庫匯出，不改正式 i64。"""
import hashlib
import json
import os

import ida_auto
import ida_bytes
import ida_funcs
import ida_loader
import ida_ua
import idaapi
import idc

ida_auto.auto_wait()
base = ida_loader.get_fileregion_ea(0)
manifest = json.load(open('/work/settings66-r98-input.json', encoding='utf-8'))
raw = ida_bytes.get_bytes(base, manifest['size'])
assert raw and hashlib.sha256(raw).hexdigest() == manifest['sha256']
rows = []
for lo, hi in manifest['ranges']:
    ida_bytes.del_items(base + lo, ida_bytes.DELIT_SIMPLE, hi - lo)
    offset = lo
    while offset < hi:
        ea = base + offset
        ins = ida_ua.insn_t()
        size = ida_ua.decode_insn(ins, ea)
        if not size:
            raise ValueError('cannot decode at file offset %X' % offset)
        if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
            assert ida_ua.create_insn(ea), 'cannot create instruction at %X' % offset
        rows.append({
            'ida_ea': '%X' % ea,
            'runtime_cs_ip': '%04X:%04X' % (manifest['segment'], offset),
            'file_offset': '%04X' % offset,
            'bytes': ida_bytes.get_bytes(ea, size).hex(),
            'text': idc.generate_disasm_line(ea, 0) or '',
            'function_original_name': ida_funcs.get_func_name(ea),
            'semantic_level': 'unknown',
            'warning': '未證實：自動名稱只作導覽',
        })
        offset += size
result = {
    'schema': 'logh3.settings-interface-ida.v1',
    'ida_version': idaapi.get_kernel_version(), 'uid': os.getuid(),
    'input': manifest, 'file_offset_zero_ea': '%X' % base,
    'address_space': 'IDA EA = file offset + %Xh; runtime CS separately recorded' % base,
    'instructions': rows,
}
with open('/work/' + manifest['output'], 'x', encoding='utf-8') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
idc.qexit(0)
