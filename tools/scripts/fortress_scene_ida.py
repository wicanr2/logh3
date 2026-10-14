"""第133輪：唯讀來源DB的一次性副本，保留原始名稱、位址與bytes。"""
import hashlib
import json
import os
from pathlib import Path
import idaapi
import ida_auto
import ida_bytes
import ida_funcs
import ida_pro
import idautils
import idc

ida_auto.auto_wait()
source = Path('/input/GIN3PS_UNPACKED.EXE')
database = Path('/input/GIN3PS_UNPACKED.EXE.i64')
ram = Path('/native/gun-before-mem.bin').read_bytes()
assert os.getuid() == 1000 and len(ram) == 1048576
rows = []
for start in idautils.Functions():
    linear = start - 0x10000 + 0x3490
    if not 0x22640 <= linear < 0x22640 + 0x0EC3:
        continue
    off = linear - 0x22640
    if not (0x0248 <= off < 0x02D9 or 0x06F5 <= off < 0x0EC3):
        continue
    f = ida_funcs.get_func(start)
    instructions = []
    for ea in idautils.FuncItems(start):
        n = ida_bytes.get_item_size(ea)
        raw = ida_bytes.get_bytes(ea, n)
        address = ea - 0x10000 + 0x3490
        instructions.append({'original_location': f'2264:{address-0x22640:04X}',
            'ida_ea': hex(ea), 'runtime_linear': hex(address),
            'original_bytes_hex': raw.hex(), 'native_bytes_hex': ram[address:address+n].hex(),
            'native_equal': raw == ram[address:address+n],
            'operand': idc.generate_disasm_line(ea, 0), 'grade': 'unknown'})
    rows.append({'original_location': f'2264:{off:04X}', 'original_name': idc.get_name(start),
                 'ida_start': hex(start), 'ida_end': hex(f.end_ea), 'instructions': instructions})
out = {'schema': 'logh3.fortress-scene-ida.v1', 'kernel_version': idaapi.get_kernel_version(),
       'uid': os.getuid(), 'function_count': sum(1 for _ in idautils.Functions()),
       'input_sha256': hashlib.sha256(source.read_bytes()).hexdigest(),
       'source_database_sha256': hashlib.sha256(database.read_bytes()).hexdigest(),
       'native_input_sha256': hashlib.sha256(ram).hexdigest(),
       'address_space': 'IDA EA=image_offset+0x10000; runtime linear=image_offset+0x3490; original CS:IP separate',
       'functions': rows}
with open('/out/fortress-scene-ida.json', 'x') as f:
    json.dump(out, f, ensure_ascii=False, indent=2)
    f.write('\n')
ida_pro.qexit(0)
