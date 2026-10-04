"""第77輪正式顯示收據；完整全圖零差閘門保持失敗，不裁切未核准像素。"""
import hashlib
import json
import os
from pathlib import Path
import re
import struct
import subprocess
import sys

root = Path('/src')
os.chdir(root)
def digest(b): return hashlib.sha256(b).hexdigest()
def create(path, content):
    p = Path(path)
    assert not p.exists() and p.parent.stat().st_uid == os.getuid(), str(p)
    p.write_text(content)
def overlay(replace, name):
    path = f'workplace/{name}.json'
    create(path, json.dumps({'Replace': {str(root/a):str(root/b) for a,b in replace.items()}}, indent=2)+'\n')
    print(path)

if sys.argv[1] == '--overlay':
    source = 'internal/game/fleetcombat.go'
    target = 'workplace/round77-display-replay-final.go'
    create(target, Path(source).read_text()+'''\n// 僅供重播：固定seed及唯讀命令stage；不注入局面。
func (g *Game) FleetChainProbeSeed(seed uint32) { g.battleRand = rules.NewRand(seed) }
func (g *Game) FleetCommandProbeStage() int { if !g.fleetCombat.active {return -1};return int(g.fleetCombat.stage) }
''')
    subprocess.run(['gofmt','-w',target], check=True)
    overlay({source:target}, 'round77-display-replay-final')
    sys.exit(0)

if sys.argv[1] == '--before':
    source = 'internal/game/fleetcombat.go'
    old = Path('workplace/round77-v76-fleetcombat.go').read_text()
    s = Path(source).read_text()
    start, end = 'func (g *Game) drawFleetCombatCommand(', 'func (g *Game) drawFleetCombatRange('
    oldfunc = old[old.index(start):old.index('func (g *Game) finishFleetCombat(')]
    s = s[:s.index(start)]+oldfunc+s[s.index(end):]
    at, after = s.index(end), s.index('func (g *Game) finishFleetCombat(')
    s = s[:at]+'func (g *Game) drawFleetCombatRange(fb *gfx.Framebuffer, view screen.BattleView) {}\n\n'+s[after:]
    target='workplace/round77-before-production.go'
    create(target,s)
    test='internal/game/fleetcommanddisplay_test.go'
    t=Path(test).read_text().replace('t.Fatalf("出現既有亮標順位以外的差異：','t.Logf("第76輪正式繪法的完整差異：')
    t=t.replace('full["remaining_classification"] = ', 'full["before_rendering_source"] = "第76輪drawFleetCombatCommand與無射程亮標；其餘用已修正診斷fixture"\n\tfull["remaining_classification"] = ')
    testtarget='workplace/round77-before-test.go'
    create(testtarget,t)
    subprocess.run(['gofmt','-w',target,testtarget],check=True)
    overlay({source:target,test:testtarget},'round77-before-production-overlay')
    sys.exit(0)

assert sys.argv[1]=='--freeze'
out=Path('docs/internals/data/fleet-command-display-game.json')
assert not out.exists() and out.parent.stat().st_uid==os.getuid()
files={}
def read(path):
    p=Path(path)
    assert p.stat().st_uid==os.getuid(),str(p)
    b=p.read_bytes()
    files[str(path)]={'sha256':digest(b),'bytes':len(b)}
    return b
native_path='docs/internals/data/fleet-command-display-native.json'
assert digest(read(native_path))=='d56ab5516e494b43987aab0d72210b71d2ff93f1b56c9c3576dfba08c255b1db'
assert digest(read('docs/internals/data/fleet-command-display-spec.json'))=='5bf0789c6c62484b49ac18f757d4ee246647f75b860017897d4c227885637fb4'
comparisons=[]
for p in sorted(Path('workplace/round77-final-d').glob('*')):
    a=read(p);b=read(Path('workplace/round77-final-e')/p.name)
    assert a==b,p.name
    if p.name.endswith('-comparison.json'):
        c=json.loads(a)
        assert c['different']==c['unapproved_differences']==1 and not c['whole_frame_gate_passed']
        assert c['effective_palette_equal'] and c['bbox']==[255,255,255,255]
        assert all(x['different']==0 for x in c['regions'][1:])
        comparisons.append(c)
assert len(comparisons)==7
before=[]
for p in sorted(Path('workplace/round77-before-corrected').glob('*')):
    b=read(p)
    if p.name.endswith('-comparison.json'):before.append(json.loads(b))
assert len(before)==7 and all(c['different']>1 for c in before)
normal=[]
for side in [0,1]:
    first,second=[Path(f'workplace/round77-chain-{s}/normal-{side}')for s in ['c','d']]
    names=sorted(p.name for p in first.iterdir()if p.is_file())
    assert names==sorted(p.name for p in second.iterdir()if p.is_file())
    for name in names:assert read(first/name)==read(second/name),name
    commands=[n for n in names if n.startswith('command-') and n.endswith('.bin')]
    assert commands and any(n.startswith('command-1-')for n in commands) and any(n.startswith('command-2-')for n in commands)
    paths=[Path(f'workplace/round77-bootstrap-{s}-{side}')for s in ['final-a','final-b']]
    ns=sorted(p.name for p in paths[0].iterdir()if p.is_file())
    assert ns==sorted(p.name for p in paths[1].iterdir()if p.is_file())
    for name in ns:assert read(paths[0]/name)==read(paths[1]/name),name
    for name in commands:assert read(paths[0]/name)==read(first/name)
    plan=json.loads((first/'inputs.json').read_text());replay=json.loads((paths[0]/'replay.json').read_text())
    assert replay['six_save_files_equal'] and plan['cpu_position_unchanged'] and plan['scenes']>0
    normal.append({'faction':side,'commands_compared':len(commands),'orders':plan['orders'],'hits':plan['scenes'],
        'inputs':replay['inputs'],'six_save_files_equal':True,'cpu_position_unchanged':True,
        'bootstrap_commands_equal':True,'double_replay_equal':True,
        'classification':'正常Game與正式bootstrap；原版整段same-state仍未知'})
ram=read('workplace/oracle/normal79-r61-a/end-mem.bin')
supplementary=[]
semantic_index=[]
semantics={
    '39D7:0098':'讀半徑作誤差與x初值；y初值0',
    '39D7:0136':'取舊y；先計算2*y-1再遞增y',
    '39D7:0147':'誤差非負保持x；負值時減x並補2*(舊x-1)',
    '3982:003E':'unsigned編號小於10使用全形格式，其他用两半形',
    '3982:0046':'DS8E86全形格式，沿研究143／338',
    '3982:0060':'DS8E8B的兩半形格式',
    '25DB:0E8D':'配40-byte亮標物件，沿研究394',
    '25DB:0F17':'亮標priority byte+3寫FF；本函式沒有word+4寫入',
    '25DB:0F33':'亮標tile+0C寫10',
    '25DB:0F4E':'由SHIP+34讀目前X；後續+36讀目前Y',
}
for seg in ['39D7','3982','25DB']:
    p=f'workplace/ida/combat79-r77-{seg}.json';d=json.loads(read(p))
    assert d['kernel_version']=='9.4' and d['uid']==1000
    assert digest(read(f'workplace/ida/combat79-r77-{seg}.com'))==d['input_sha256']
    for x in d['instructions']:
        at=int(x['runtime_linear_hex'],16);b=bytes.fromhex(x['original_bytes_hex']);assert ram[at:at+len(b)]==b
        if x['original_location']in semantics:
            semantic_index.append({**x,'semantic':semantics[x['original_location']],'grade':'proven',
                'warning':None,'source':'docs/re/414-fleet-command-display.md；固定原始bytes與第77輪原版雙跑'})
    supplementary.append({'path':p,'instructions':len(d['instructions']),'input_sha256':d['input_sha256'],
        'grade':'unknown逐指令；審查過的語意另列研究414','ranges':d['ranges']})
objects=[]
m=read('workplace/oracle/combat79-r77-wait-a/end-mem.bin')
for i in range(struct.unpack_from('<H',m,0x54b30+0xeade)[0]):
    off,seg=struct.unpack_from('<HH',m,0x54b30+0xeae0+i*4);at=seg*16+off;b=m[at:at+40]
    if b[3]>=254:objects.append({'list_index':i,'runtime_linear_hex':hex(at),'bytes_hex':b.hex(),
        'priority':b[3],'word_plus_4':struct.unpack_from('<H',b,4)[0],'grade':'proven本次物件；初始化未知'})
assert next(o for o in objects if o['runtime_linear_hex']=='0x74c20')['word_plus_4']==8224
assert next(o for o in objects if o['runtime_linear_hex']=='0x7101e')['word_plus_4']==64
logs=['round77-limited-c.log','round77-limited-d.log','round77-standard.log','round77-before-corrected.log']
logs+=[f'round77-bootstrap-{s}-{side}.log'for s in['final-a','final-b']for side in[0,1]]
for p in logs:read('workplace/'+p)
standard=Path('workplace/round77-standard.log').read_text()
assert not re.search(r'^(FAIL|--- FAIL)',standard,re.M)
assert sum(l.startswith('ok ')for l in standard.splitlines())==13
assert sum('[no test files]'in l for l in standard.splitlines())==5
for p in ['internal/game/fleetcombat.go','internal/game/game.go','internal/game/fleetcommanddisplay_test.go',
    'internal/game/fleetplayerchain_test.go','internal/bootstrap/load.go','internal/rules/fleetcommandrange.go',
    'internal/screen/fleet_command_display.go','internal/screen/unitpanel.go',
    'tools/scripts/fleet_command_display_replay.go','tools/scripts/fleet_command_display_game_receipt.py',
    'workplace/round77-display-replay-final.go','workplace/round77-display-replay-final.json',
    'workplace/round77-v76-fleetcombat.go','workplace/round77-before-production.go',
    'workplace/round77-before-test.go','workplace/round77-before-production-overlay.json']:
    read(p)
out.write_text(json.dumps({'schema':'logh3.fleet-command-display-game.v1','date':'2026-10-04',
    'parent_commit':'e30f3869b58db2d484f6109931e5181e784026b7','worktree':'本輪修改；來源逐檔雜湊如下',
    'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c','dosgolem_worktree':'clean; unchanged',
    'image':'logh3-ebiten:2','seed':0x12345678,'comparison_level':'same-state限定受控顯示，正常全流程不外推',
    'whole_frame_gate_passed':False,'unapproved_differences_each':1,
    'comparisons':comparisons,'before_comparisons':before,'normal_cases':normal,
    'supplementary_ida':supplementary,'semantic_index':semantic_index,'cached_display_objects':objects,'files':files,
    'standard_tests':{'pass_packages':13,'no_test_packages':5},
    'limits':['七張完整圖各一像素未核准；#89既有排序鍵裁決pending','103保持READY；#48／#79保持OPEN',
              '原版注入後既有顯示物件仍在；fixture明示还原快取，不宣稱全部typed欄位same-state',
              '補給／修理／偵察範圍、實際X11命令與音效未新增驗收','舊76輪及更早收據未覆寫']},ensure_ascii=False,indent=2)+'\n')
print(out.name,digest(out.read_bytes()),len(files),'sources')
