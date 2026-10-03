"""第66輪原版序列收據：雙跑、完整注入、IDA bytes及模型，拒覆寫。"""
import hashlib
import json
import os
from pathlib import Path
import sys

root, work = Path('/ref'), Path('/work/workplace')
output = Path(sys.argv[1])
assert not output.exists() and output.parent.stat().st_uid == os.getuid()
sources = {}

def sha(data):
    return hashlib.sha256(data).hexdigest()

def record(p):
    assert p.is_file() and p.stat().st_uid == os.getuid(), p
    data = p.read_bytes()
    key = str(p.relative_to(root)) if p.is_relative_to(root) else 'workplace/'+str(p.relative_to(work))
    sources[key] = {'sha256':sha(data), 'bytes':len(data), 'uid':p.stat().st_uid, 'gid':p.stat().st_gid}
    return data

ancestors = {}
def plan(p):
    d = json.loads(record(p))
    for key in ('inherited_plan','inherited_injections','parent_plan'):
        inherited = d.get(key)
        if inherited and inherited not in ancestors:
            ancestors[inherited] = None
            ancestors[inherited] = plan(work/inherited)
    if 'parent' in d:
        assert sha(record(root/d['parent'])) == d['parent_sha256']
    elif 'parent_state_sha256' in d:
        flags = record(work/'oracle'/d['name']/'flags.txt').decode().splitlines()
        parent = flags[flags.index('-load-state')+1]
        assert parent.startswith('/out/')
        assert sha(record(work/parent[5:])) == d['parent_state_sha256']
    return d

source_mem = record(work/'oracle/normal79-r61-a/end-mem.bin')
assert sha(source_mem) == '6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4'
model = json.loads(record(work/'round66-sequence-probe.json'))
assert len(model['curves']) == 23 and len(model['sequences']) == 2
pairs = []
for a in sorted((work/'oracle').glob('sequence79-r66-*-a')):
    b = a.with_name(a.name[:-1]+'b')
    assert b.is_dir(), b
    plans = [plan(work/(p.name+'-plan.json')) for p in (a,b)]
    for key in ('injections','clicks','parent_sha256','input_dump_step','end_dump_step'):
        assert plans[0][key] == plans[1][key], (a.name,key)
    flags = [record(p/'flags.txt').decode() for p in (a,b)]
    assert flags[0].replace(a.name,'<run>') == flags[1].replace(b.name,'<run>')
    files = ['input-mem.bin','end-mem.bin','screen.bin','pal.bin','screen.png']
    files += sorted(p.name for p in a.glob('*-mem.bin') if p.name not in files)
    comparisons = []
    for file in files:
        da,db = (record(p/file) for p in (a,b))
        assert da == db, (a.name,file)
        if file.endswith('mem.bin'): assert len(da) == 1048576
        comparisons.append({'file':file,'equal':True,'sha256':sha(da),'bytes':len(da)})
    for p in (a,b):
        meta = record(p/'meta.txt').decode()
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in meta
        assert '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918' in meta
        record(p/'report.txt')
        record(work/(p.name+'-flags.txt'))
        for state in p.glob('*.state'): record(state)
    pairs.append({'case':plans[0]['case'],'plans':plans,'comparisons':comparisons,
                  'classification':plans[0]['classification'],
                  'phase_limit':'sorting dumps too early; excluded from model' if plans[0]['case']=='same-target-ranked' else None})
assert len(pairs) == 35
zero = record(work/'oracle/sequence79-r66-curve0-zero-a/input-mem.bin')
changed = record(work/'oracle/sequence79-r66-curve0-stack-first-a/input-mem.bin')
for addr,size in [(0x68112,78),(0x687c6,78),(0x6115c,122),(0x61344,122),(0x603fc,28),(0x6070c,28),(0x5e00c,4)]:
    assert zero[addr:addr+size] == changed[addr:addr+size]

semantics = {
    '203A:0161':'首點X大於或等於q時跳到017E；q=0跳過前項指標初始化',
    '203A:0163':'僅currentX小於q的分支寫入前項指標offset',
    '203A:017E':'使用local前項指標；q=0讀未初始化stack，並非curve[-1]',
    '203A:0181':'使用前項far pointer的segment',
    '203A:070E':'按目標敵方陣營累加DS8FE4分數表',
    '203A:0716':'按目標敵方陣營累加DS8FEC分數表',
    '1DD6:04BC':'SHIP艦種表+09查DS4426排序基值',
    '1DD6:04CB':'每個一般分隊第一個排序RND',
    '1DD6:04E7':'每個一般分隊第二個排序RND',
    '1DD6:04EF':'提督ADML+0B參與第二個排序項',
    '1DD6:04FC':'SHIP+06原始byte參與排序鍵',
    '1DD6:051D':'呼叫297B083F取得機動力加值',
    '1DD6:052C':'機動力加到SHIP+42字鍵',
    '1DD6:0597':'比較左側SHIP+42 unsigned word',
    '1DD6:05A6':'左鍵較大時回FFFF，產生降序比較',
    '297B:0866':'副官FF時不讀副官記錄',
    '297B:0894':'比較提督與副官ADML+17，選較大者',
    '297B:08AE':'機動力中間項270減兩倍能力',
    '297B:08BB':'編制數查DS4E2E係數byte',
    '297B:08C7':'乘積低word以signed idiv100往零截斷',
    '297B:08CE':'100減中間項，以AL byte返回機動力',
}
exports = []
for filename in ('battle79-r63-1DD6-loss.json','battle79-r63-1DD6-damage.json',
                 'battle79-r64-203A.json','battle79-r66-297B.json'):
    p = work/'ida'/filename
    d = json.loads(record(p))
    assert d['kernel_version']=='9.4' and d['uid']==os.getuid()
    com = record(p.with_suffix('.com'))
    assert sha(com)==d['input_sha256']
    segment = int(d['runtime_segment'],16)
    assert com[256:]==source_mem[segment*16:segment*16+65536]
    for row in d['instructions']:
        off = int(row['original_location'].split(':')[1],16)
        raw = bytes.fromhex(row['original_bytes_hex'])
        assert source_mem[segment*16+off:segment*16+off+len(raw)]==raw
        if row['original_location'] in semantics:
            row.update(semantic=semantics[row['original_location']],grade='proven',warning=None,
                       source=filename+'；原始bytes及第66輪固定原版觀測／probe')
    exports.append(d)
assert set(semantics)<={r['original_location'] for d in exports for r in d['instructions']}
for filename in ('tools/combat_sequence_oracle.sh','tools/scripts/combat_sequence_plan.py',
                 'tools/scripts/combat_sequence_probe.go','tools/scripts/combat_sequence_ida.py',
                 'tools/scripts/combat_sequence_receipt.py','docs/internals/56-fleet-composition-panel.md'):
    record(root/filename)
record(work/'round66-standard-tests.log')
images = record(work/'round66-images.txt').decode().splitlines()
assert len(images)==4 and all('sha256:' in x for x in images)
data = {'schema':'logh3.combat-sequence.v1','date':'2026-10-04',
        'remake_parent':'b61764b4e524e9885ffb799b0be870a726e405b9','remake_worktree_dirty':True,
        'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c','dosgolem_dirty':False,
        'original_bundle_sha256':'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
        'images':images,'pairs':pairs,'ancestor_plans':ancestors,'model':model,'ida_exports':exports,
        'sources':sources,'decisions':{'zero_distance':{'approved':'2026-10-04','rule':'取曲線首點',
              'classification':'user-approved original difference; repair undefined stack behavior',
              'production_implemented':False}},
        'limits':['原版受控診斷，不替代正常remake玩家路徑','序列化state不作逐byte相同聲明',
                  'zero曲線模型使用原版觀測stack，不是正式typed規則',
                  'same-target-ranked排序dump錯相位；只用measured核排序',
                  '本輪多擊限同艦隊两筆戰鬥機命令，未覆蓋死亡／取消全部分支',
                  '完整演出、正式Game與保存未完成，092 DRAFT']}
output.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
print('pairs',len(pairs),'equal files',sum(len(x['comparisons']) for x in pairs),
      'sources',len(sources),'IDA rows',sum(len(x['instructions']) for x in exports),
      'proven',len(semantics),'SHA-256',sha(output.read_bytes()))
