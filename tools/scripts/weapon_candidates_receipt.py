"""第65輪原版收據凍結。核a/b、IDA bytes與模型，拒絕覆寫。"""
import hashlib
import json
import os
from pathlib import Path
import re
import sys

root = Path('/ref')
work = Path('/work/workplace')
output = Path(sys.argv[1])
assert not output.exists() and output.parent.stat().st_uid == os.getuid()
sources = {}

def sha(data):
    return hashlib.sha256(data).hexdigest()

def record(p):
    assert p.is_file() and p.stat().st_uid == os.getuid(), p
    data = p.read_bytes()
    key = str(p.relative_to(root)) if p.is_relative_to(root) else 'workplace/'+str(p.relative_to(work))
    sources[key] = {'sha256': sha(data), 'bytes': len(data), 'uid': p.stat().st_uid, 'gid': p.stat().st_gid}
    return data

source_mem = record(work/'oracle/normal79-r61-a/end-mem.bin')
assert sha(source_mem) == '6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4'
record(work/'oracle/combat79-r62-natural-a/entry.state')
record(work/'oracle/combat79-r62-natural-a/entry-mem.bin')
model = json.loads(record(work/'round65-probe-final.json'))
assert len(model['candidates']) == 48
pairs = []
for a in sorted((work/'oracle').glob('weapon79-r65-*-a')):
    b = a.with_name(a.name[:-1]+'b')
    assert b.is_dir(), b
    plans = [json.loads(record(work/(p.name+'-plan.json'))) for p in (a, b)]
    assert plans[0]['injections'] == plans[1]['injections']
    assert plans[0]['clicks'] == plans[1]['clicks']
    assert plans[0]['parent_sha256'] == plans[1]['parent_sha256']
    parent = record(root/plans[0]['parent'])
    assert sha(parent) == plans[0]['parent_sha256']
    flags = [record(p/'flags.txt').decode() for p in (a, b)]
    assert flags[0].replace(a.name, '<run>') == flags[1].replace(b.name, '<run>')
    comparisons = []
    for file in ('input-mem.bin', 'end-mem.bin', 'screen.bin', 'pal.bin', 'screen.png'):
        da, db = (record(p/file) for p in (a, b))
        assert da == db, (a.name, file)
        if file.endswith('mem.bin'):
            assert len(da) == 1048576
        comparisons.append({'file': file, 'equal': True, 'sha256': sha(da), 'bytes': len(da)})
    for p in (a, b):
        meta = record(p/'meta.txt').decode()
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in meta
        assert '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918' in meta
        record(p/'report.txt')
        record(work/(p.name+'-flags.txt'))
        for state in p.glob('*.state'):
            record(state)
    pairs.append({'case': plans[0]['case'], 'stage': plans[0]['stage'], 'plans': plans,
                  'comparisons': comparisons, 'classification': plans[0]['classification']})
assert len(pairs) == 53
semantics = {
    '17A0:00FC': '候選陣營讀SHIP+00，與要求的敵方陣營比較',
    '17A0:0106': 'SHIP+1E bit7設立則排除，不要求目標Out恰為7F',
    '17A0:010F': '距離查詢使用SHIP+36目前Y',
    '17A0:0115': '距離查詢使用SHIP+34目前X',
    '17A0:0133': 'FixedDist大於300才排除，等於300保留',
    '1F6F:0351': '目標Detail為0直接返回，不清既有+3C',
    '1F6F:035A': '目標艦種7直接返回，不清既有+3C',
    '1F6F:0382': '寫入目標SHIP+3C候選mask；此欄位可被後續分隊重新查詢覆寫',
    '1F6F:046F': '光線砲艦種旗標FF跳過',
    '1F6F:04B3': '只有光線砲候選續查091C射線',
    '1F6F:04D4': '飛彈庫存兩word OR為0才排除',
    '1F6F:0524': '戰鬥機庫存兩word OR為0才排除',
    '1F6F:0615': '距離小於或等於方向橢圓半徑回1',
    '1F6F:094A': '射線起點X使用SHIP+34加7；其餘三座標同樣加7',
    '1F6F:0BBA': '遮蔽物枚舉對所有陣營適用，不過濾艦種或艘數',
    '1F6F:0BBC': '遮蔽物Out bit7設立則排除',
    '1F6F:0C1C': '射線只在步數為8的倍數檢查地形與物件',
    '1F6F:0C85': '遮蔽物矩形X半開區間為原點至原點加16',
    '1F6F:0C92': '遮蔽物矩形Y半開區間為原點至原點加16',
    '1588:057C': '世界Y除10轉地形子格',
    '1588:0588': '世界X右移3轉地形子格',
    '1588:063F': '子格座標各除2取得地形tile編號',
    '1588:0663': '子格奇偶查DS:3C3E四項bitmask',
    '1588:0684': '地形tile讀DS:3C1A farptr，row stride128',
    '1588:069D': 'tile編號查DS:9BB4碰撞mask表',
}
exports = []
for p in sorted((work/'ida').glob('battle79-r65-*.json')):
    d = json.loads(record(p))
    assert d['kernel_version'] == '9.4' and d['uid'] == os.getuid()
    com = record(p.with_suffix('.com'))
    assert sha(com) == d['input_sha256']
    segment = int(d['runtime_segment'], 16)
    assert com[256:] == source_mem[segment*16:segment*16+65536]
    for row in d['instructions']:
        off = int(row['original_location'].split(':')[1], 16)
        raw = bytes.fromhex(row['original_bytes_hex'])
        assert source_mem[segment*16+off:segment*16+off+len(raw)] == raw
        key = row['original_location']
        if key in semantics:
            row['semantic'] = semantics[key]
            row['grade'] = 'proven'
            row['warning'] = None
            row['source'] = str(p.relative_to(work))+'；固定bytes及第65輪48案候選probe'
    exports.append(d)
assert set(semantics) <= {r['original_location'] for d in exports for r in d['instructions']}
for file in ('tools/weapon_candidates_oracle.sh', 'tools/scripts/weapon_candidates_plan.py',
             'tools/scripts/weapon_candidates_ida.py', 'tools/scripts/weapon_candidates_probe.go',
             'tools/scripts/weapon_candidates_receipt.py'):
    record(root/file)
record(work/'round65-standard-tests.log')
images = record(work/'round65-images.txt').decode().splitlines()
assert len(images) == 4 and all('sha256:' in line for line in images)
data = {'schema': 'logh3.weapon-candidates.v1', 'date': '2026-10-04',
        'remake_parent': '124f4ee33f4593a3605574152d251ad155cbb925',
        'remake_worktree_dirty': True, 'images': images,
        'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', 'dosgolem_dirty': False,
        'original_bundle_sha256': 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
        'pairs': pairs, 'model': model, 'ida_exports': exports, 'sources': sources,
        'limits': ['原版受控診斷，不替代正常remake玩家路徑', 'state序列化未作逐byte一致聲明',
                   'search-wide400圓的b平方截斷；搜尋邊界只採search-open兩案',
                   'coincident已進可用候選，但傷害curve前項仍unknown，不猜fallback',
                   '多候選／多攻擊者、完整演出、正式Game及保存未完成，092 DRAFT']}
output.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print('pairs', len(pairs), 'equal files', sum(len(p['comparisons']) for p in pairs),
      'sources', len(sources), 'IDA rows', sum(len(d['instructions']) for d in exports),
      'proven', len(semantics), 'SHA-256', sha(output.read_bytes()))
