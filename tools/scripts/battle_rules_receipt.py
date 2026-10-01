"""#53 的受控收據與經人工審查語意索引；入口 docs/playtest/038。

原版記憶體與完整單位 bytes 留在 workplace；版控只存解出的欄位與身份。
"""
import argparse
import base64
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
assert args.golem_dirty == '', '原版執行器須為固定乾淨版本'
root = Path('/ref')
out = Path('/work/docs/internals/data')
local = Path('/work/workplace/oracle/battle53-routines')
def digest(b):
    return hashlib.sha256(b).hexdigest()
def file_hash(path):
    return digest(path.read_bytes())
def write(path, value):
    assert path.parent.stat().st_uid == 1000
    if path.exists():
        assert path.stat().st_uid == 1000
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')
a, b = [local.joinpath(name).read_bytes() for name in ('run-c.json','run-d.json')]
assert a == b, '兩輪原版收據須逐位元組相同'
r = json.loads(a)
assert r['seed'] == '0x12345678' and r['tick'] == 0
seed_bytes = list(struct.pack('<I',0x12345678))
assert all(list(base64.b64decode(c['rng_after'])) == seed_bytes
           for c in r['losses'] + r['spreads']), '受控常式不應消費亂數'
def fields(value):
    raw = bytes.fromhex(value)
    assert len(raw) == 78
    values = struct.unpack_from('<HHHII',raw,8)
    return dict(zip(('ships','full_ships','repair_cap','missiles','fighters'),values))
losses = [{k:c[k] for k in ('kind','amount','return','kind_flags')} |
          {'before':fields(c['before_hex']),'after':fields(c['after_hex'])} for c in r['losses']]
spreads = [{k:c[k] for k in ('amount','flag','return','display_percent')} |
           {'before':[fields(x) for x in c['before_hex']], 'after':[fields(x) for x in c['after_hex']]}
           for c in r['spreads']]
receipt = {'schema':'logh3.battle-rule-corrections.v1','date':'2026-10-02','issue':53,
           'grade':'proven','scope':r['scope'], 'input_sha256':r['input_sha256'],
           'exe_sha256':r['exe_sha256'], 'bundle_sha256':'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
           'golem_head':args.golem_head,'golem_dirty':False,
           'remake_parent':args.remake_head,'remake_worktree':args.remake_dirty,
           'go_image':'golang:1.24-bookworm','go_image_id':args.go_image_id,
           'original_command':'DOSGOLEM_ORIG=/home/anr2/cht/logh3/org_game/extracted DOSGOLEM_EXTRA_MOUNT=/home/anr2/cht/logh3/workplace:/out bash /home/anr2/cht/dosgolem-logh3/tools/go.sh run /out/battle_rules_probe.go -output /out/oracle/battle53-routines/run-{c,d}.json',
           'seed':r['seed'],'seed_location':r['seed_linear'],'remake_seed':r['seed'],
           'rng_method':'兩側事前固定 0x12345678；三支受控常式無亂數判定；原版執行後種子 bytes 不變。remake 同初始欄位，不以相同 seed 數字證明完整骰序。',
           'injection':r['register_setup']+'；扣損五案及分攤兩案的全部注入 bytes 與位址保留在兩份本機收據',
           'runs':[{'path':'workplace/oracle/battle53-routines/'+n,'sha256':digest(a)} for n in ('run-c.json','run-d.json')],
           'baseline_meta_sha256':file_hash(Path('/work/workplace/oracle/s3assault/meta.txt')),
           'baseline_flags_sha256':file_hash(Path('/work/workplace/oracle/s3assault/flags.txt')),
           'atan_table':r['atan_table'],'atan_table_sha256':r['atan_table_sha256'],
           'rows':r['rows'],'losses':losses,'spreads':spreads,
           'source_sha256':{name:file_hash(root/name) for name in
                            ['tools/scripts/battle_rules_probe.go','tools/scripts/battle_rules_receipt.py','tools/scripts/battle_rules_ida.py',
                             'internal/rules/battle.go','internal/rules/fortressgun.go','internal/rules/planetassault.go',
                             'internal/rules/strategicassault.go','internal/rules/battle_rules_oracle_test.go',
                             'tools/scripts/battle_rules_gui_check.py']}}
assert root.joinpath('tools/scripts/battle_rules_probe.go').read_bytes() == Path('/work/workplace/battle_rules_probe.go').read_bytes()
events=[json.loads(line) for line in (local/'battle-regressions.jsonl').read_text().splitlines() if line.startswith('{')]
assert not any(e['Action'] in ('skip','fail') for e in events)
passed=[e['Test'] for e in events if e['Action']=='pass' and 'Test' in e]
required={'TestBattleRulesMatchesOracle','TestPlanetAssaultMatchesOracle','TestFortressGunMatchesOracle',
          'TestStrategicAssaultMatchesOracle','TestBattleMoveMatchesOracle','TestBattleResultPersists'}
assert required.issubset(passed)
gui=json.loads((local/'gui-final/gui-check.json').read_text())
assert [c['name'] for c in gui['cases']]==['normal-entry','normal-turn','normal-save']
assert gui['save']['different']==0 and gui['save']['fleet_counts']==[1,3]
for name in ['internal/rules/battle.go','internal/rules/planetassault.go','internal/rules/fortressgun.go','internal/rules/strategicassault.go']:
    assert gui['source_sha256'][name]==file_hash(root/name)
receipt['validation']={'battle_test_pass_events':len(passed),'battle_tests':passed,'skipped':0,'failed':0,
                       'normal_gui':gui,'go_version':(local/'go-version.txt').read_text().strip(),
                       'commands':['bash tools/gui.sh test -count=1 ./internal/... ./cmd/...',
                                   'bash tools/gui.sh test -count=1 -json ./internal/rules ./internal/game ./internal/screen -run <對拍 038 的固定測試名稱>',
                                   'bash tools/gui.sh xrun python3 tools/scripts/battle_rules_gui_check.py --head <Git HEAD> --dirty <工作樹狀態> --image-id <Docker image ID>'],
                       'files':{n:file_hash(local/n) for n in ['before-tests.log','after-tests.log','after-verified-tests.log','product-tests.log','battle-regressions.jsonl','gui-final/gui-check.json']}}
ida_files=['re53-loss-reviewed-ida.json','re53-spread-reviewed-ida.json','re53-atan-reviewed-ida.json']
receipt['ida']={'image':'ida-pro-9.4-idapython:locked-v1',
                'image_id':'sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780',
                'exports':[{'path':'workplace/ida/'+n,'sha256':file_hash(Path('/work/workplace/ida')/n),
                            'tool':json.loads(Path('/work/workplace/ida',n).read_text())['tool']} for n in ida_files]}
write(out/'battle-rule-corrections.json',receipt)

# 下列項目已逐項人工對照原始指令及受控實驗，並非從散文自動擷取。
review = {
    '203A:0620':'種類 5 跳到飛彈縮減分支，繞過表旗標 FF 排除。',
    '203A:0668':'種類 5 跳到戰機縮減分支，繞過表旗標 FF 排除。',
    '20B4:11C9':'尾額夾到艘數減一後寫到 [bp+2]；後續 call 沒有推入這個 local。',
    '20B4:11D2':'尾額 call 的量來自 [bp+6] 完整剩餘量，未套用 [bp+2] 夾值。',
    '20B4:1215':'返回 AX 載入 [bp+8] 實扣艘數；百分比先另寫 ds:A768。',
    '2BE1:01EE':'反正切查表起點為原版資料段偏移 593A。',
    '2BE1:0202':'查表索引上限 64；表範圍 593A..59B9 的 64 個 word 已逐值核對。',
}
exports = [json.loads(Path('/work/workplace/ida',n).read_text()) for n in
           ('re53-loss-ida.json','re53-spread-full-ida.json','re53-atan-ida.json')]
entries=[]
for d in exports:
    for f in d['functions']:
        for row in f['instructions']:
            at=row['original_location']
            if at in review:
                entries.append({'kind':'instruction','original_location':at,'original_bytes_hex':row['original_bytes_hex'],
                                'semantic':review[at],'grade':'proven',
                                'source':'docs/re/366-battle-rule-corrections.md；原版受控兩輪 SHA-256 '+digest(a)})
assert len(entries)==len(review)
index={'schema':'logh3.battle-rule-semantics.v1','source':'研究 366 的逐指令人工審查',
       'inputs':{d['functions'][0]['original_location'].split(':')[0]:d['input_sha256'] for d in exports},
       'address_contract':'逐項 CS:IP 為本次 dosgolem runtime；必須另保留 IDA .COM 位址及原始 bytes。長期資料以切段雜湊與原始指令簽章定位。',
       'entries':entries}
write(out/'battle-rule-semantics.json',index)
write(Path('/work/workplace/ida/re53-semantics.json'),index)
print('正式原版兩輪 SHA-256',digest(a))
print('解出的欄位摘要 SHA-256',file_hash(out/'battle-rule-corrections.json'))
print('語意索引 SHA-256',file_hash(out/'battle-rule-semantics.json'))
