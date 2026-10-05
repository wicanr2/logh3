"""一次性COM資料庫讀取當前0349載入器；檔案映射EA與runtime CS:IP另列。"""
import json,hashlib,os
import idaapi,ida_auto,ida_bytes,ida_funcs,ida_ua,ida_loader,idautils,idc

ida_auto.auto_wait()
base=ida_loader.get_fileregion_ea(0)
raw=ida_bytes.get_bytes(base,4096)
assert raw and hashlib.sha256(raw).hexdigest()=='0138b89c7e5dbd9bcf708630921835806d01e712d5c8ad7c1d9bbb1b78fb3674'
rows=[]
annotations=json.load(open('/work/title68-r97-loader-annotations.json',encoding='utf-8'))
assert annotations['input_sha256']==hashlib.sha256(raw).hexdigest()
index={int(r['file_offset'],16):r for r in annotations['annotations']}
for offset,r in index.items():
    signature=bytes.fromhex(r['bytes'])
    assert raw[offset:offset+len(signature)]==signature
for lo,hi in [(0x380,0x910),(0x910,0xb90)]:
    ea=lo+base
    while ea<hi+base:
        ins=ida_ua.insn_t();size=ida_ua.decode_insn(ins,ea)
        if not size:
            ea+=1;continue
        if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
            ida_ua.create_insn(ea)
        note=index.get(ea-base)
        rows.append({'ida_ea':f'{ea:X}','runtime_cs_ip':f'0349:{ea-base:04X}',
            'file_offset':f'{ea-base:04X}',
            'bytes':ida_bytes.get_bytes(ea,size).hex(),'text':idc.generate_disasm_line(ea,0) or '',
            'function_original_name':ida_funcs.get_func_name(ea),
            'source':'runtime dump at 911000000; IDA loader file-offset mapping',
            'semantic':note['semantic'] if note else '',
            'semantic_level':note['level'] if note else 'unknown',
            'semantic_source':annotations['source'] if note else '',
            'warning':'' if note else '未證實：自動函式名只作導覽'})
        ea+=size
result={'schema':'logh3.title-loader-ida.v1','ida_version':idaapi.get_kernel_version(),
    'uid':os.getuid(),'input_sha256':hashlib.sha256(raw).hexdigest(),
    'function_count':len(list(idautils.Functions())),
    'file_offset_zero_ea':f'{base:X}',
    'address_space':f'IDA EA = runtime offset + {base:X}h; runtime segment0349',
    'instructions':rows}
with open('/work/title68-r97-loader-ida-v5.json','x',encoding='utf-8') as f:json.dump(result,f,ensure_ascii=False,indent=2)
idc.qexit(0)
