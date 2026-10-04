"""核對分級IDA索引及原始bytes，保留未審查指令，不覆寫舊收據。"""
from pathlib import Path
import hashlib
import json

def sha(b):
    return hashlib.sha256(b).hexdigest()

primary_path = Path('docs/internals/data/battle-promotion.json')
primary_bytes = primary_path.read_bytes()
primary = json.loads(primary_bytes)
for item in primary['sources']:
    b = Path(item['path']).read_bytes()
    assert len(b) == item['bytes'] and sha(b) == item['sha256'], item['path']
memory = Path('workplace/oracle/ownership79-r82-entry-a/mem-in.bin').read_bytes()
review_path = Path('docs/internals/data/battle-promotion-annotations.json')
review_bytes = review_path.read_bytes()
review = json.loads(review_bytes)
assert review['source_sha256'] == sha(primary_bytes)
sources = []
def source(p):
    p = Path(p)
    b = p.read_bytes()
    assert (p.stat().st_uid, p.stat().st_gid) == (1000, 1000)
    sources.append({'path': str(p), 'bytes': len(b), 'sha256': sha(b)})

for p in [primary_path, review_path, 'tools/scripts/battle_promotion_review_ida.py',
          'tools/scripts/battle_promotion_review_receipt.py']:
    source(p)
exports = []
for name in ['185C', '28E0']:
    p = Path(f'workplace/ida/promotion79-r84-{name}-reviewed.json')
    data = json.loads(p.read_bytes())
    assert data['kernel_version'] == '9.4' and data['uid'] == 1000
    assert data['annotations']['sha256'] == sha(review_bytes)
    assert data['input_sha256'] == sha(Path(f'workplace/ida/promotion79-r84-{name}.com').read_bytes())
    for row in data['instructions']:
        seg, off = (int(n, 16) for n in row['original_location'].split(':'))
        b = bytes.fromhex(row['original_bytes_hex'])
        assert memory[seg*16+off:seg*16+off+len(b)] == b
        assert row['runtime_linear_hex'] == hex(seg*16+off)
        assert row['grade'] in {'proven', 'unknown'}
        if row['grade'] != 'proven':
            assert row['warning']
    exports.append(data)
    source(p)
ranks = []
for faction, base in enumerate([0x6830, 0x689C]):
    names = []
    for i in range(12):
        entry = 0x54B30+base+i*4
        pointer = memory[entry:entry+4]
        off = int.from_bytes(pointer[:2], 'little')
        seg = int.from_bytes(pointer[2:], 'little')
        start = seg*16+off
        b = memory[start:memory.index(0, start)]
        names.append({'rank_index': i, 'pointer_ds_offset_hex': hex(base+i*4),
                      'pointer_bytes_hex': pointer.hex(), 'string_segment_offset': f'{seg:04X}:{off:04X}',
                      'big5_hex': b.hex(), 'text': b.decode('big5'), 'grade': 'proven',
                      'source': '固定父RAM；28E0:004F直接消費者'})
    ranks.append({'faction': faction, 'names': names})
result = {'schema': 'logh3.battle-promotion-review.v1', 'primary_sha256': sha(primary_bytes),
          'primary_sources_verified': len(primary['sources']),
          'grade': 'proven', 'scope': '原指令bytes、21項直接語意及兩陣營12級文字；其他指令保持unknown。非正式remake或GUI對拍。',
          'annotation_rows': len(review['rows']), 'rank_names': ranks, 'exports': exports, 'sources': sources}
p = Path('docs/internals/data/battle-promotion-review.json')
with p.open('x', encoding='utf-8') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write('\n')
print('53原始來源、完整指令bytes、21項分級索引通過')
print(p, len(p.read_bytes()), sha(p.read_bytes()), 'sources', len(sources))
