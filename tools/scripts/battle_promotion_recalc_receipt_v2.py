"""#79：獨立驗算升級亂數、能力、速度、完整度與全艦隊費用。"""
from pathlib import Path
import ast
import collections
import hashlib
import json

def sha(b):
    return hashlib.sha256(b).hexdigest()

memory_path = Path('workplace/oracle/ownership79-r82-entry-a/mem-in.bin')
memory = memory_path.read_bytes()
assert sha(memory) == 'b905eda112fd3a78292bc0d1cdc10068703e0291968640cbaa0d150404404900'
ds = 0x54B30
need = [int.from_bytes(memory[ds+0x4DBE+i*2:ds+0x4DC0+i*2], 'little') for i in range(40)]
mobility = list(memory[ds+0x4E2E:ds+0x4E3A])
kind_bytes = memory[ds+0x4E3A:ds+0x4E3A+416]
# 沿第84輪已審查的純成長模型，只載入兩個函式，不執行舊收據產生流程。
tree = ast.parse(Path('tools/scripts/battle_promotion_receipt.py').read_text())
functions = [n for n in tree.body if isinstance(n, ast.FunctionDef) and n.name in ('rank', 'promote')]
assert len(functions) == 2
exec(compile(ast.Module(body=functions, type_ignores=[]), '<frozen-r84-growth>', 'exec'))

def word(b, o):
    return int.from_bytes(b[o:o+2], 'little')

def dword(b, o):
    return int.from_bytes(b[o:o+4], 'little')

def putword(b, o, v):
    b[o:o+2] = (v & 65535).to_bytes(2, 'little')

def putdword(b, o, v):
    b[o:o+4] = (v & 0xFFFFFFFF).to_bytes(4, 'little')

def signed16(n):
    n &= 65535
    return n if n < 32768 else n-65536

def signed32(n):
    n &= 0xFFFFFFFF
    return n if n < 0x80000000 else n-0x100000000

def trunc(n, d):
    return n//d if n >= 0 else -((-n)//d)

def kind(ship, off):
    assert ship[0] < 2 and ship[4] < 8
    return kind_bytes[ship[0]*208+ship[4]*26+off]

def cost(fleet, ships):
    sums = [0, 0]
    for i in range(fleet[15]):
        pool = word(fleet, 0x10+4*i)
        ship = ships[pool*78:(pool+1)*78]
        for j in range(2):
            sums[j] = (sums[j]+kind(ship, 4+j)*word(ship, 8)) & 0xFFFFFFFF
    for j in range(2):
        factor = 200-fleet[0x47]
        value = (((sums[j]//1000)*factor) & 0xFFFFFFFF)//100
        putword(fleet, 0xA+2*j, value)
    return signed16(word(fleet, 0xA))

def recalculate(tables, index):
    admirals, fleets, ships = (tables[n] for n in ('admiral', 'fleet', 'ship'))
    f = bytearray(fleets[index*122:(index+1)*122])
    a = admirals[f[4]*28:(f[4]+1)*28]
    deputy = admirals[f[5]*28:(f[5]+1)*28] if f[5] != 255 else bytes(28)
    f[14] = next((v for limit,v in [(2,1),(6,2),(16,3),(20,4),(24,6),(29,7),(34,8),(39,9)] if a[6]<limit),10)
    for i, off in enumerate([15,16,17,19,18,12,13,20,22,23,14]):
        f[0x3D+i] = a[off] if a[off] >= deputy[off] else ((a[off]+deputy[off])//2 if i<5 else deputy[off])
    minima = [255]*4
    integrity = []
    for i in range(f[15]):
        pool = word(f, 0x10+4*i)
        sh = bytearray(ships[pool*78:(pool+1)*78])
        base = [kind(sh,8)*f[0x43]//100, kind(sh,6)*f[0x42]//80, kind(sh,7)*f[0x42]//60]
        p = trunc(signed16((270-2*f[0x46])*mobility[f[15]]),100)
        for j, off in enumerate([0x1F,0x21,0x22]):
            v = base[j] & 255
            discount = ((v*(p & 65535)) & 65535)//100
            sh[off] = (v-discount) & 255
        for j, off in enumerate([0x1F,0x21,0x22,0x23]):
            minima[j] = min(minima[j], sh[off])
        if word(sh,8):
            full = word(sh,10)
            assert full != 0
            lost = (full-word(sh,8)) & 65535
            loss = (((lost*(200-a[0xB])) & 0xFFFFFFFF)//100) & 65535
            loss = min(loss, full)
            sh[6] = ((full-loss)*100//full) & 255
            integrity.append(sh[6])
        ships[pool*78:(pool+1)*78] = sh
    f[0x3C] = 100
    f[0x4D:0x51] = bytes(minima)
    scout = word(f,0x38)
    assert scout < 352
    sh = bytearray(ships[scout*78:(scout+1)*78])
    for off, ko in [(0x1F,8),(0x21,6),(0x22,7)]:
        sh[off] = kind_bytes[sh[0]*208+7*26+ko]
    ships[scout*78:(scout+1)*78] = sh
    if not integrity:
        f[0] |= 8
    else:
        f[0] &= 247
        f[0x3C] = sum(integrity)//len(integrity)
        old = f[0]
        f[0] = f[0]|1 if f[0x3C]<50 else f[0]&254
        if (old^f[0])&1:
            f[0] |= 4
    fleets[index*122:(index+1)*122] = f
    # 090A:0E88 → 297B:10DE 會更新全陣營每支現役艦隊的兩項费用。
    totals = [0,0]
    moves = [0,0]
    for i in range(32):
        fr = bytearray(fleets[i*122:(i+1)*122])
        if fr[4] == 255:
            continue
        totals[fr[1]] += cost(fr,ships)
        if fr[8] not in (2,5):
            moves[fr[1]] += signed16(word(fr,0xC))
        fleets[i*122:(i+1)*122] = fr
    return totals,moves

paths = [Path(f'workplace/promotion79-r85-recalc-{side}.json') for side in 'ab']
raw = paths[0].read_bytes()
assert raw == paths[1].read_bytes(), '原版雙跑不同'
oracle = json.loads(raw)
assert oracle['schema'] == 'logh3.battle-promotion-recalc-native.v1'
assert oracle['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
assert not oracle['dosgolem_dirty'] and oracle['uid'] == 1000
summary = collections.Counter()
cases = []
returns = {'28E0:098D','28E0:09A0','28E0:070D','28E0:078A','28E0:0857','28E0:08E2'}
for row in oracle['rows']:
    model = {n:bytearray.fromhex(row['before'][n]) for n in ('admiral','fleet','ship')}
    seed = 0x12345678
    active = None
    cursor = 0
    checkpoints = []
    for call in row['calls']:
        where = call['location']
        if where == '28E0:05F9':
            assert active is None
            active = {'index':call['args'][0]&255, 'randoms':[]}
        elif where in returns:
            seed = (seed*0x41C64E6D+0x3039)&0xFFFFFFFF
            value = (seed>>16)&32767
            assert active is not None and value == call['ax'], (row['name'],where,value,call['ax'])
            active['randoms'].append(value)
            summary['random_outputs'] += 1
        elif where in ('185C:065B','185C:06A8'):
            assert active is not None
            i = active['index']
            after, flags = promote(model['admiral'][i*28:(i+1)*28],active['randoms'])
            assert flags == call['ax'], (row['name'],i,flags,call['ax'])
            model['admiral'][i*28:(i+1)*28] = after
            summary['attempts'] += 1
            summary['upgrades'] += bool(flags)
            active = None
        elif where in ('185C:066E','185C:06BB'):
            expected = row['recalculations'][cursor]
            assert expected['location'] == where and expected['step'] == call['step']
            totals,moves = recalculate(model,row['case'][0])
            hashes = {}
            for n, b in model.items():
                original = bytes.fromhex(expected['tables'][n])
                delta = [(hex(i),v,original[i]) for i,v in enumerate(b) if v != original[i]]
                assert not delta, (row['name'],where,n,delta[:20])
                hashes[n] = sha(b)
            work = bytes.fromhex(expected['tables']['work'])
            for faction in range(2):
                assert dword(work,0x5E+4*faction) == (totals[faction] & 0xFFFFFFFF)
                assert dword(work,0x6E+4*faction) == (moves[faction] & 0xFFFFFFFF)
                available = signed32(dword(work,0x3E+4*faction)+dword(work,0x56+4*faction)-totals[faction]-moves[faction])
                budget = 0 if available<0 else ((available//100)*work[0x7E+faction] if available>=21474836 else trunc(available*work[0x7E+faction],100))
                assert dword(work,0x46+4*faction) == (budget & 0xFFFFFFFF), (row['name'],faction,available,budget,dword(work,0x46+4*faction))
            checkpoints.append({'location':where,'step':call['step'],'model_sha256':hashes,'upkeep':totals,'move_cost':moves,
                                'work_sha256':sha(work),'stellar_sha256':sha(bytes.fromhex(expected['tables']['stellar']))})
            cursor += 1
            summary['recalculations'] += 1
    assert active is None and cursor == len(row['recalculations'] or [])
    assert seed.to_bytes(4,'little').hex() == row['seed_after_hex']
    tables = {}
    for n, before_hex in row['before'].items():
        before, after = bytes.fromhex(before_hex),bytes.fromhex(row['after'][n])
        assert len(before) == len(after)
        if n in model:
            assert model[n] == after, (row['name'],n)
        differences = [{'offset_hex':hex(i),'before':v,'after':after[i]} for i,v in enumerate(before) if v != after[i]]
        summary[n+'_changed_bytes'] += len(differences)
        tables[n] = {'bytes':len(before),'before_sha256':sha(before),'after_sha256':sha(after),'differences':differences,
                     'model': '全表獨立驗算' if n in model else '完整原版觀測；費用與造艦預估已核，稅收／星域尚未獨立驗算'}
    cases.append({k:row[k] for k in ('name','index','case','need','injections','seed_after_hex','steps')} |
                 {'tables':tables,'checkpoints':checkpoints,'grade':'proven','comparison':'固定8-4入口診斷；非正常玩家GUI／完整回合同態'})
    summary['cases'] += 1
    summary['reports'] += sum(c['location']=='3826:0405' for c in row['calls'])

sources = []
def source(p):
    p = Path(p)
    assert p.is_file() and p.stat().st_uid == 1000
    b = p.read_bytes()
    sources.append({'path':str(p),'bytes':len(b),'sha256':sha(b)})

for p in paths+[memory_path,Path('workplace/promotion79-r85-recalc.go'),Path('docs/internals/data/battle-promotion.json'),Path('workplace/promotion79-r84-native-v2.go')]:
    source(p)
for p in ['tools/scripts/battle_promotion_recalc_prepare.py','tools/scripts/battle_promotion_receipt.py',
          'tools/scripts/battle_promotion_recalc_receipt_v2.py','tools/scripts/battle_promotion_recalc_ida_prepare.py',
          'tools/scripts/battle_promotion_review_ida.py','tools/battle_promotion_recalc_oracle.sh',
          'tools/battle_promotion_recalc_native_run.sh','tools/battle_promotion_recalc_ida.sh',
          'workplace/oracle/ownership79-r82-entry-a/entry.state','workplace/ida/re79-r76-integrity.json']:
    source(p)
instructions = []
for seg in ('297B','090A'):
    p = Path(f'workplace/ida/promotion79-r85-reviewed-{seg}.json')
    d = json.loads(p.read_text())
    assert d['kernel_version'] == '9.4' and d['uid'] == 1000
    original = Path(f'workplace/ida/promotion79-r85-reviewed-{seg}.com')
    assert sha(original.read_bytes()) == d['input_sha256']
    source(original); source(p)
    for ins in d['instructions']:
        segment,off = [int(v,16) for v in ins['original_location'].split(':')]
        b = bytes.fromhex(ins['original_bytes_hex'])
        assert memory[segment*16+off:segment*16+off+len(b)] == b
        instructions.append(ins)

legacy_path = Path('docs/internals/data/battle-work-control.json')
legacy_bytes = legacy_path.read_bytes()
assert sha(legacy_bytes) == '96660f9c7c8cea715b484d39cfe0a586f81ecc574d963b5bb3ce5c6b9b06578f'
legacy = json.loads(legacy_bytes)
source(legacy_path)
source('docs/internals/data/battle-promotion-recalc.json')
global_cost_instructions = []
for export in legacy['ida_exports']:
    for ins in export['data']['instructions']:
        location = ins['original_location']
        if '297B:10DE' <= location < '297B:116C' or '297B:12FB' <= location <= '297B:131B':
            segment, off = [int(v,16) for v in location.split(':')]
            b = bytes.fromhex(ins['original_bytes_hex'])
            assert memory[segment*16+off:segment*16+off+len(b)] == b
            global_cost_instructions.append(ins | {'source_export_path':export['path'],'source_export_sha256':export['sha256']})
assert len(global_cost_instructions) == 67
result = {'schema':'logh3.battle-promotion-recalc.v2','grade':'proven',
          'supersedes':'v1漏驗全局移動費；原版與三表成長／速度結果仍有效。','scope':'66個原版8-4入口診斷雙跑與三表逐次重算；全局稅收／星域僅觀測，不宣稱獨立驗算或產品已接線。',
          'oracle_identity':{k:v for k,v in oracle.items() if k!='rows'},'original_json_sha256':sha(raw),
          'summary':dict(summary),'mobility':mobility,'ship_kind_table_sha256':sha(kind_bytes),
          'cases':cases,'instructions':instructions,'global_cost_instructions':global_cost_instructions,'sources':sources}
dest = Path('docs/internals/data/battle-promotion-recalc-v2.json')
with dest.open('x',encoding='utf-8') as f:
    json.dump(result,f,ensure_ascii=False,indent=2); f.write('\n')
print('重算獨立模型通過',dict(summary))
print(dest,len(dest.read_bytes()),sha(dest.read_bytes()),'sources',len(sources))
