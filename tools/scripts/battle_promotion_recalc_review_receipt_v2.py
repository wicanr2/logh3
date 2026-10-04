"""#79：核對凍結來源、IDA bytes與自動合併的語意索引。"""
from pathlib import Path
import hashlib
import json

def sha(b):
    return hashlib.sha256(b).hexdigest()

primary_path = Path('docs/internals/data/battle-promotion-recalc-v2.json')
raw = primary_path.read_bytes()
assert sha(raw) == 'cda14b5b122e4e875a825948c61194c96b534a07c7a563c770b26570b35291e6'
primary = json.loads(raw)
for entry in primary['sources']:
    p = Path(entry['path'])
    b = p.read_bytes()
    assert len(b) == entry['bytes'] and sha(b) == entry['sha256'], str(p)
    assert p.stat().st_uid == 1000, str(p)
annotations_path = Path('docs/internals/data/battle-promotion-recalc-annotations-v2.json')
notes_raw = annotations_path.read_bytes()
assert notes_raw == Path('workplace/ida/promotion79-r85-recalc-annotations-v2.json').read_bytes()
notes = json.loads(notes_raw)
assert notes['source_sha256'] == sha(raw)
index = {r['original_location']:r for r in notes['rows']}
memory = Path('workplace/oracle/ownership79-r82-entry-a/mem-in.bin').read_bytes()
instructions, functions, merged = [], [], set()
sources = []
def source(p):
    p = Path(p)
    b = p.read_bytes()
    assert p.stat().st_uid == 1000
    sources.append({'path':str(p),'bytes':len(b),'sha256':sha(b)})

for p in [primary_path,annotations_path,'tools/scripts/battle_promotion_recalc_annotations_v2.py',
          'tools/scripts/battle_promotion_recalc_review_ida_v2.py',
          'tools/scripts/battle_promotion_recalc_review_receipt_v2.py','tools/battle_promotion_recalc_review_ida_v2.sh']:
    source(p)
for seg in ('297B','090A'):
    p = Path(f'workplace/ida/promotion79-r85-graded-v2-{seg}.json')
    data = json.loads(p.read_text())
    source(p)
    assert data['kernel_version'] == '9.4' and data['uid'] == 1000
    assert data['annotations']['sha256'] == sha(notes_raw)
    assert data['annotations']['source_sha256'] == sha(raw)
    original = Path(f'workplace/ida/promotion79-r85-reviewed-{seg}.com').read_bytes()
    assert data['input_sha256'] == sha(original)
    for ins in data['instructions']:
        segment, off = [int(v,16) for v in ins['original_location'].split(':')]
        b = bytes.fromhex(ins['original_bytes_hex'])
        assert memory[segment*16+off:segment*16+off+len(b)] == b
        if ins['original_location'] in index:
            note = index[ins['original_location']]
            assert all(ins[k] == note[k] for k in ('original_bytes_hex','grade','semantic','source'))
            assert ins['grade'] == 'proven' and not ins['warning']
            merged.add(ins['original_location'])
        else:
            assert ins['grade'] == 'unknown' and ins['warning']
        instructions.append(ins)
    functions.extend(data['functions'])
assert merged == set(index)
assert len(instructions) == 808
boundaries = [('297B:10DE','297B:116B'),('297B:12FB','297B:131B'),('297B:0209','297B:025C'),('297B:05B2','297B:083E'),
              ('297B:0E8B','297B:10DD'),('297B:116C','297B:12FA'),
              ('090A:0E6D','090A:0E87'),('090A:0E88','090A:0F84')]
by_location = {i['original_location']:i for i in instructions}
for start,end in boundaries:
    assert start in by_location and by_location[end]['original_bytes_hex'] == 'cb'
result = {'schema':'logh3.battle-promotion-recalc-review.v2','grade':'proven',
          'source_sha256':sha(raw),'annotations_sha256':sha(notes_raw),
          'scope':'808條原始指令bytes與42條已審查語意；其他766條保持unknown。全局稅收／星域尚未獨立驗算。',
          'boundaries':[{'original_entry':s,'original_last':e,'grade':'proven'} for s,e in boundaries],
          'functions':functions,'instructions':instructions,'sources':sources}
dest = Path('docs/internals/data/battle-promotion-recalc-review-v2.json')
with dest.open('x',encoding='utf-8') as f:
    json.dump(result,f,ensure_ascii=False,indent=2); f.write('\n')
print('來源',len(primary['sources']),'指令',len(instructions),'proven',len(merged))
print(dest,len(dest.read_bytes()),sha(dest.read_bytes()),'sources',len(sources))
