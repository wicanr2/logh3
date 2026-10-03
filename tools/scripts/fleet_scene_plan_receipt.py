"""由既有原版兩個06FC入口與完整結果凍結typed場景來源。"""
import hashlib
import json
import struct
from pathlib import Path

root = Path('/src')
sources = []


def read(p):
    raw = p.read_bytes()
    sources.append({'path': str(p.relative_to(root)), 'bytes': len(raw),
                    'sha256': hashlib.sha256(raw).hexdigest()})
    return raw


raw = read(root / 'docs/internals/data/fleet-blocks.json')
assert hashlib.sha256(raw).hexdigest() == '20874578ab46987b007f746efa0e180ff0906e5cc8c23ca3afdd0f1eadc9afeb'
blocks = json.loads(raw)
cases = []
for scene in (1, 2):
    frames = [f for f in blocks['frames'] if f['metadata']['scene'] == scene]
    entry, complete = frames[0], frames[-1]
    assert entry['metadata']['stage'] == 'entry' and complete['metadata']['stage'] == 'complete'
    ram = read(root / ('workplace/oracle/blocks48-r70-a/'+entry['metadata']['name']+'-mem.bin'))
    other = read(root / ('workplace/oracle/blocks48-r70-b/'+entry['metadata']['name']+'-mem.bin'))
    assert ram == other
    ds = 0x54b30
    def word(a): return struct.unpack_from('<H', ram, a)[0]
    def pointer(a): return word(a)+16*word(a+2)
    pool = pointer(ds+0xcfb4)
    sides, facing = [], []
    for fleet in (12, 16):
        f = ds+0xc074+fleet*122
        facing.append(ram[f+0x4a])
        units = []
        for i in range(ram[f+15]):
            idx = word(f+16+i*4)
            a = pool+idx*78
            units.append({'id': idx, 'side': ram[a], 'x': word(a+0x34), 'y': word(a+0x36),
                          'squadron': ram[a+3], 'detail': ram[a+0x17], 'weapon': ram[a+0x2e],
                          'ships': word(a+8), 'raw_hex': ram[a:a+78].hex()})
        sides.append(units)
    arrow = [list(struct.unpack_from('<4h', ram, ds+0x6072+i*8)) for i in range(4)]
    assert arrow == [[16,0,-16,-16],[16,0,-16,16],[-16,-16,0,0],[-16,16,0,0]]
    cases.append({'name': f'scene-{scene}', 'sides': sides, 'facing': facing,
                  'visibility': {'mode': ram[ds+0x8f20], 'player_side': ram[ds+0x8f25],
                                 'side_control': [ram[ds+0x8f21], ram[ds+0x8f22]]},
                  'attacker_id': 80, 'target_id': 102, 'arrow_segments': arrow,
                  'operations': [op for f in frames for op in f['operations']],
                  'entry_native_zlib_base64': entry['native_zlib_base64'],
                  'complete_native_zlib_base64': complete['native_zlib_base64']})
read(root / 'tools/scripts/fleet_scene_plan_receipt.py')
out = {'schema': 'logh3.fleet-scene-plan.v1', 'date': '2026-10-04',
       'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
       'classification': 'same-state lower-scene compositor; above borrowed from native entry; not Game player path',
       'palette': blocks['logical_palette_rgb8'], 'sprites': blocks['sprites'],
       'sources': sources, 'cases': cases}
data = (json.dumps(out, ensure_ascii=False, indent=2)+'\n').encode()
dest = root / 'docs/internals/data/fleet-scene-plan.json'
if dest.exists(): assert dest.read_bytes() == data, '拒絕改寫凍結收據'
else: dest.write_bytes(data)
print(len(sources), 'sources', len(data), 'bytes', hashlib.sha256(data).hexdigest())
