"""#79：只保存解出的欄位、固定身份與人工審查索引。入口：研究 367。"""
import argparse
import hashlib
import json
import struct
from pathlib import Path

p = argparse.ArgumentParser()
p.add_argument('--golem-head', required=True)
p.add_argument('--golem-dirty', required=True)
p.add_argument('--remake-head', required=True)
p.add_argument('--remake-dirty', required=True)
p.add_argument('--go-image-id', required=True)
args = p.parse_args()
assert args.golem_head == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
assert args.golem_dirty == '', '原版執行器必須為固定乾淨版本'
root = Path('/src')
local = root/'workplace/oracle/battle79-routines'
target = root/'docs/internals/data'
sha = lambda b: hashlib.sha256(b).hexdigest()
def write(path, value):
    assert path.parent.stat().st_uid == 1000
    assert not path.exists() or path.stat().st_uid == 1000
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')

a, b = [(local/n).read_bytes() for n in ('run-c.json', 'run-d.json')]
assert a == b
r = json.loads(a)
assert len(r['rows']) == 16 and all(c['rng_after_hex'] == '78563412' for c in r['rows'])
assert (root/'tools/scripts/battle_preparation_probe.go').read_bytes() == (root/'workplace/battle_preparation_probe.go').read_bytes()
def unit(h):
    v = bytes.fromhex(h)
    return dict(zip(('ships','full_ships','repair_cap','missiles','fighters'), struct.unpack_from('<HHHII', v, 8))) | {
        'faction':v[0], 'fleet':v[1], 'squadron':v[3], 'kind':v[4],
        'target_x':struct.unpack_from('<H',v,52)[0], 'target_y':struct.unpack_from('<H',v,54)[0]}
def planet(h):
    v = bytes.fromhex(h)
    return dict(zip(('defence','defence_max','recover'), struct.unpack_from('<HHH',v,20))) | {
        'owner':v[17], 'x':struct.unpack_from('<H',v,70)[0], 'y':struct.unpack_from('<H',v,72)[0],
        'ships':list(struct.unpack_from('<7I',v,32)), 'missiles':struct.unpack_from('<I',v,60)[0],
        'fighters':struct.unpack_from('<I',v,64)[0]}
rows = [{'name':c['name'],'entry':c['entry'], 'before':[unit(h) for h in c['units_before_hex']],
         'after':[unit(h) for h in c['units_after_hex']], 'planet_before':planet(c['planet_before_hex']),
         'planet_after':planet(c['planet_after_hex'])} for c in r['rows']]
replays=[]
for name in ('battle79-preparation-b','battle79-preparation-c'):
    d=root/'workplace/oracle'/name
    replays.append({'path':str(d.relative_to(root)),
                    'sha256':{n:sha((d/n).read_bytes()) for n in ('mem-in.bin','mem-out.bin','screen.bin','pal.bin','flags.txt','meta.txt')},
                    'flags':(d/'flags.txt').read_text().splitlines()})
for n in ('mem-in.bin','mem-out.bin','screen.bin','pal.bin'):
    assert replays[0]['sha256'][n] == replays[1]['sha256'][n], n
value={'schema':'logh3.battle-preparation.v1','date':'2026-10-02','issue':79,'grade':'proven',
       'scope':r['scope'],'input_sha256':r['input_sha256'],'program_sha256':r['program_sha256'],
       'bundle_sha256':'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
       'golem_head':args.golem_head,'golem_dirty':False,'remake_parent':args.remake_head,
       'remake_worktree':args.remake_dirty,'go_image':'golang:1.24-bookworm','go_image_id':args.go_image_id,
       'seed':r['seed'],'seed_location':r['seed_linear'],'tick':0,
       'rng_method':'事前固定種子；十六案執行後 bytes 不變；未以相同 seed 數字宣稱整場骰序相同。',
       'command':'DOSGOLEM_ORIG=/home/anr2/cht/logh3/org_game/extracted DOSGOLEM_EXTRA_MOUNT=/home/anr2/cht/logh3/workplace:/out bash /home/anr2/cht/dosgolem-logh3/tools/go.sh run /out/battle_preparation_probe.go -output /out/oracle/battle79-routines/run-{c,d}.json',
       'injection':r['setup']+' 完整注入表及 bytes 在本機 run-c／run-d；版控只存解出的欄位。',
       'run_sha256':sha(a),'runs':['workplace/oracle/battle79-routines/run-c.json','workplace/oracle/battle79-routines/run-d.json'],
       'phase_replays':replays,'rows':rows,
       'source_sha256':{n:sha((root/n).read_bytes()) for n in ('tools/scripts/battle_preparation_probe.go','tools/scripts/battle_preparation_receipt.py','tools/scripts/battle_rules_ida.py')}}
write(target/'battle-preparation.json',value)

index=json.loads((root/'workplace/ida/re79-semantic-index.json').read_text())
index['schema']='logh3.battle-preparation-semantics.v1'
index['source']='研究 367 逐指令人工審查；不從散文自動匯入'
index['entries']=[]
reports={seg:json.loads((root/'workplace/ida'/('re79-'+name+'.json')).read_text())
         for seg,name in [('1930','1930-g'),('1AAB','1AAB-c'),('181B','181B')]}
annotations={
 '1930:023A':'只有種類 5 呼叫運補艦自動補給；按戰場艦艇表順序遍歷。',
 '1930:029D':'有主行星自動補給飛彈與戰機；在修理入口之前。',
 '1AAB:0296':'自動修理初選半徑為修護艦種類列速度乘 6；不是 6.4。',
 '1AAB:033F':'依排序清單呼叫共用修理；運補艦可作為修理目標。',
 '1AAB:0428':'自動行星補充不呼叫排除種類 5／6 的手動篩選函式。',
 '1AAB:0E10':'補充前先記下原修理上限減原艘數的差。',
 '1AAB:0E1B':'增加目標現有艘數；行星該艦種存量相應減少。',
 '1AAB:0E33':'新艘數加原修理差，與滿編上限比較後回寫修理上限。',
 '1AAB:0E4A':'寫入目標 +0x0C 修理上限；滿補與部分補充皆會更新。',
 '181B:00BF':'目前防衛力無號大於等於上限時，整筆跳過；不強制夾回上限。',
 '181B:00CD':'清掉乘積高 16 位；先取低 16 位，再除以 100。',
 '181B:00D4':'防衛力以 16 位加法回繞後，才與上限比較。'}
for location,semantic in annotations.items():
    report=reports[location[:4]]
    ins=next(x for f in report['functions'] for x in f['instructions'] if x['original_location']==location)
    index['entries'].append({'kind':'instruction','original_location':location,'original_bytes_hex':ins['original_bytes_hex'],
                             'semantic':semantic,'grade':'proven','source':'docs/re/367-battle-preparation.md；兩輪常式收據 SHA-256 '+sha(a)})
write(target/'battle-preparation-semantics.json',index)
write(root/'workplace/ida/re79-reviewed-index.json',index)
print('兩份階段擷取及十六案常式雙跑逐位元組相同；種子不變。收據',sha(a))
print('解出欄位 SHA-256',sha((target/'battle-preparation.json').read_bytes()))
print('人工審查索引 SHA-256',sha((target/'battle-preparation-semantics.json').read_bytes()))
