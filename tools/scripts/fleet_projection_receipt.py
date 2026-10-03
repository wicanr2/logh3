"""第72輪：直接凍結原版投影期望值，不讀 remake 或近似模型。"""
import hashlib
import json
from pathlib import Path

root = Path('/src')
a = root / 'workplace/oracle/projection48-r72-a'
b = root / 'workplace/oracle/projection48-r72-b'
sources = []


def source(p):
    data = p.read_bytes()
    sources.append({'path': str(p.relative_to(root)), 'bytes': len(data),
                    'sha256': hashlib.sha256(data).hexdigest()})
    return data


ra = json.loads(source(a / 'report.json'))
rb = json.loads(source(b / 'report.json'))
ra.pop('command')
rb.pop('command')
assert ra == rb
assert len(ra['cases']) == 30
for p in sorted(a.glob('*-mem.bin')):
    assert source(p) == source(b / p.name), p
ram = (a / 'entry-mem.bin').read_bytes()
instructions = []
for name, limits in [('scene48-r69-321F.json', [(10, 138), (458, 914), (1299, 1788)]),
                     ('projection48-r72-3A92-detail.json', None),
                     ('projection48-r72-3C6A-full.json', None)]:
    p = root / 'workplace/ida' / name
    d = json.loads(source(p))
    for row in d['instructions']:
        segment, ip = [int(x, 16) for x in row['original_location'].split(':')]
        if limits and not any(lo <= ip < hi for lo, hi in limits):
            continue
        raw = bytes.fromhex(row['original_bytes_hex'])
        addr = segment * 16 + ip
        assert ram[addr:addr+len(raw)] == raw, row['original_location']
        instructions.append(row)
for p in [root / 'tools/fleet_projection_oracle.sh',
          root / 'tools/scripts/fleet_projection_native.go',
          root / 'tools/scripts/fleet_projection_ida.py',
          root / 'tools/scripts/fleet_projection_receipt.py']:
    source(p)
facts = [
    ('321F:000A', '依兩個FLEET+0F及+10起的編制順序生成完整分隊列表，不過濾艘數或索敵。'),
    ('321F:0223', '兩側編制首筆的SHIP+34／36是投影參考。角度為負的2BE1:011D低byte。'),
    ('321F:02AA', '中點用兩個word相加後無號右移，近距離先取中點距離整數再乘3除2。'),
    ('321F:02EE', '旗艦距離大於200改用170／470，否則左右以320為中心分開。Y均298。'),
    ('321F:0513', '相對位移乘3除2、旋轉、保存旋轉Y為SHIP+44，再把Y除2並由X扣除。'),
    ('321F:060F', '投影可見旗標為signed X8..568及Y220..376，含邊界。'),
    ('321F:0643', '比較unsigned投影Y，再比較unsigned投影X；不讀SHIP+44。'),
    ('3C6A:000D', '梳狀排序間距乘10除13，9／10改11，只在比較器回正數時交換。'),
    ('3A92:0001', '旋轉只消費角度低byte，使用DS:573A正弦表，乘積加512後捨去10位。'),
]
out = {'schema': 'logh3.fleet-projection.v1', 'date': '2026-10-04',
       'remake_parent': '65937900af4559fbf1fcb871fd4694c536dcf8f1',
       'dosgolem_head': ra['dosgolem_head'], 'dosgolem_dirty': False,
       'classification': ra['classification'], 'native': ra,
       'sources': sources, 'instructions': instructions,
       'semantic_index': [{'original_location': loc, 'semantic': desc, 'grade': 'proven',
                           'source': '原始指令與30組雙跑固定投影收據'} for loc, desc in facts],
       'limits': ['原版state的起點已受控，不能稱正常玩家驗收。',
                  'typed來源限合法佈陣座標0..2047；不宣稱任意16位溢位輸入。',
                  '排序空指標的行星分支不在艦隊子規格。',
                  '投影本身不繪圖；完整場景、等待、Game及保存依092。']}
dest = root / 'docs/internals/data/fleet-projection.json'
data = (json.dumps(out, ensure_ascii=False, indent=2)+'\n').encode()
if dest.exists():
    assert dest.read_bytes() == data, '拒絕改寫既有凍結收據'
else:
    dest.write_bytes(data)
print(len(sources), 'sources', len(instructions), 'instructions', len(data), 'bytes',
      hashlib.sha256(data).hexdigest())
