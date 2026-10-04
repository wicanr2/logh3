"""第75輪原版交戰自動命令的有界原指令匯出。"""
import hashlib
import json
import os
import idaapi
import ida_auto
import ida_bytes
import ida_nalt
import ida_pro
import ida_segment
import ida_ua
import idautils
import idc

ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
assert '/command79-r75-' in path
data = open(path, 'rb').read()
raw = data[256:]
base = ida_segment.getnseg(0).start_ea+0x100
assert ida_bytes.get_bytes(base, len(raw)) == raw
segment = int(idc.ARGV[2], 16)
rows = []
for part in idc.ARGV[3:]:
    first, last = part.split(':')
    start = int(first, 16)
    end = min(start+4096, len(raw)) if last == 'ret' else int(last, 16)
    assert 0 <= start < end <= len(raw)
    ida_bytes.del_items(base+start, ida_bytes.DELIT_SIMPLE, end-start)
    ea = base+start
    while ea < base+end:
        size = ida_ua.create_insn(ea)
        if not (size > 0 and ea+size <= base+end):
            with open(idc.ARGV[1]+'.error.txt', 'w') as f:
                f.write(f'{segment:04X}:{ea-base:04X} size {size} end {end:04X}\n')
            ida_pro.qexit(1)
        b = ida_bytes.get_bytes(ea, size)
        rows.append({'original_location': f'{segment:04X}:{ea-base:04X}',
                     'original_name': idc.get_name(ea), 'ida_ea_hex': hex(ea),
                     'runtime_linear_hex': hex(segment*16+ea-base),
                     'original_bytes_hex': b.hex(), 'operand': idc.generate_disasm_line(ea, 0),
                     'xrefs': [{'ida_from_hex': hex(x.frm), 'type': x.type} for x in idautils.XrefsTo(ea)],
                     'grade': 'unknown', 'semantic': '未審查', 'warning': '有界線性匯出，非完整函式邊界聲明',
                     'source': '第75輪沿第61輪固定原版RAM'})
        ea += size
        if last == 'ret' and b[0] in (0xcb, 0xca, 0xc3, 0xc2):
            break
out = {'schema': 'logh3.combat-command-ida.v1', 'kernel_version': idaapi.get_kernel_version(),
       'input': path, 'input_sha256': hashlib.sha256(data).hexdigest(), 'uid': os.getuid(),
       'ida_base_hex': hex(base), 'runtime_segment': f'{segment:04X}',
       'address_space': '原版CS:IP、IDA EA與dosgolem線性位址分欄',
       'ranges': list(idc.ARGV[3:]), 'instructions': rows}
with open(idc.ARGV[1], 'x', encoding='utf-8') as f:
    json.dump(out, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
