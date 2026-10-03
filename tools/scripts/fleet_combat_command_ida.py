"""第63輪固定原版RAM的人類目標／武器選單；原始定位與unknown分欄。"""
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


def fail(kind,value,tb):
    with open(idc.ARGV[1]+'.error.txt','w',encoding='utf-8') as f:
        f.write(''.join(traceback.format_exception(kind,value,tb)))
    ida_pro.qexit(1)


sys.excepthook=fail
ida_auto.auto_wait()
path=ida_nalt.get_input_file_path()
assert '/battle79-r63-' in path
with open(path,'rb') as f:
    file=f.read()
raw=file[256:]
base=ida_segment.getnseg(0).start_ea+0x100
assert ida_bytes.get_bytes(base,len(raw))==raw
segment=int(idc.ARGV[2],16)
rows=[]
functions=[]
for part in idc.ARGV[3:]:
    start,end=(int(x,16)for x in part.split(':'))
    assert 0<=start<end<=len(raw)
    ida_bytes.del_items(base+start,ida_bytes.DELIT_SIMPLE,end-start)
    ea=base+start
    while ea<base+end:
        size=ida_ua.create_insn(ea)
        assert size>0 and ea+size<=base+len(raw)
        rows.append({'original_location':f'{segment:04X}:{ea-base:04X}',
            'ida_ea_hex':hex(ea),'runtime_linear_hex':hex(segment*16+ea-base),
            'original_bytes_hex':ida_bytes.get_bytes(ea,size).hex(),
            'operand':idc.generate_disasm_line(ea,0),'semantic':'未審查',
            'grade':'unknown','warning':'附加語意未審查',
            'source':'第63輪沿normal79-r61-a固定RAM，來源SHA由收據綁定'})
        ea+=size
    fn=ida_funcs.get_func(base+start)
    functions.append({'requested_entry':f'{segment:04X}:{start:04X}',
        'preceding_bytes_hex':ida_bytes.get_bytes(base+max(0,start-8),min(8,start)).hex(),
        'original_name':idc.get_func_name(base+start),
        'ida_start_hex':hex(fn.start_ea)if fn else None,'ida_end_hex':hex(fn.end_ea)if fn else None,
        'decoded_end_offset_hex':hex(ea-base),'grade':'unknown','warning':'原始邊界需caller及動態觀測確認'})
data={'schema':'logh3.fleet-combat-command-ida.v1','kernel_version':idaapi.get_kernel_version(),
    'uid':os.getuid(),'input':path,'input_sha256':hashlib.sha256(file).hexdigest(),
    'ida_base_hex':hex(base),'runtime_segment':f'{segment:04X}',
    'address_space':'原版CS:IP、IDA EA、dosgolem線性分欄',
    'ranges':list(idc.ARGV[3:]),'functions':functions,'instructions':rows}
with open(idc.ARGV[1],'x',encoding='utf-8') as f:
    json.dump(data,f,ensure_ascii=False,indent=2)
    f.write('\n')
ida_pro.qexit(0)
