"""凍結第76輪原版名冊與命令UI資料，不能讀remake期望值。"""
import hashlib
import json
from pathlib import Path
import re
import sys

root = Path('/src') if Path('/src/internal/game').is_dir() else Path('/work')
if sys.argv[1] == '--game':
    import os
    out = root/'docs/internals/data/fleet-player-chain-game.json'
    assert not out.exists() and out.parent.stat().st_uid == os.getuid()
    files = {}
    def record(path):
        p = root/path
        assert p.stat().st_uid == os.getuid(), str(p)
        b = p.read_bytes()
        files[str(path)] = {'sha256':hashlib.sha256(b).hexdigest(),'bytes':len(b)}
        return b
    a,b = [root/('workplace/round76-chain-final-'+s) for s in 'de']
    names = sorted(p.relative_to(a) for p in a.rglob('*') if p.is_file())
    assert len(names) == 36
    assert set(names) == set(p.relative_to(b) for p in b.rglob('*') if p.is_file())
    for n in names:
        assert record((a/n).relative_to(root)) == record((b/n).relative_to(root)), str(n)
    comparisons = []
    for p in sorted(a.glob('*-comparison.json')):
        c = json.loads(p.read_bytes())
        assert c['different'] == 0 and c['maximum_channel_difference'] == 0
        assert all(r['different'] == 0 for r in c['regions'])
        comparisons.append(c)
    assert len(comparisons) == 4
    normal = []
    for side in [0,1]:
        first,second = [root/f'workplace/round76-bootstrap-{s}-{side}' for s in ['third','fourth']]
        ns = sorted(p.name for p in first.iterdir() if p.is_file())
        assert set(ns) == {'ADMIRAL.DT0','FLEET.DT0','PLANET.DT0','SAVEDATA.IDX','SHIP.DT0',
                           'STELLAR.DT0','WORK.DT0','replay.json','scene-1.png','screen.bin','screen.png'}
        assert ns == sorted(p.name for p in second.iterdir() if p.is_file())
        for n in ns:
            assert record((first/n).relative_to(root)) == record((second/n).relative_to(root))
        entry = json.loads((a/f'normal-{side}/inputs.json').read_bytes())
        replay = json.loads((first/'replay.json').read_bytes())
        assert entry['scenes'] > 0 and entry['cpu_position_unchanged']
        assert replay['six_save_files_equal'] and replay['engagement_entries'] == 1
        normal.append({'faction':side,'orders':entry['orders'],'hits':entry['scenes'],
                       'input_count':replay['inputs'],'bootstrap_double_replay':True,
                       'cpu_position_unchanged':True,'six_save_files_equal':True,
                       'classification':'normal player input replay; original full-path parity unknown'})
    for n in ['round76-chain-final-d.log','round76-chain-final-e.log','round76-standard-after-refresh.log',
              'ida/re79-r76-integrity-1261.com','ida/re79-r76-integrity.json',
              'ida/re79-r76-menu-3982.com','ida/re79-r76-menu.json']:
        record(Path('workplace')/n)
    standard = (root/'workplace/round76-standard-after-refresh.log').read_text()
    assert 'FAIL' not in standard and sum(l.startswith('ok ') for l in standard.splitlines()) == 13
    assert sum('[no test files]' in l for l in standard.splitlines()) == 5
    for n in ['internal/game/fleetcombat.go','internal/game/fleetcombatscene.go',
              'internal/game/fleetplayerchain_test.go','internal/game/battle.go',
              'internal/game/game.go','internal/game/intro.go','internal/game/battlecmd.go',
              'internal/game/engagement.go','internal/rules/planetassault.go','internal/rules/planetassault_gauges_test.go',
              'internal/bootstrap/load.go','internal/screen/fleet_command.go',
              'internal/screen/commandmenu.go','tools/scripts/fleet_player_chain_replay.go',
              'tools/scripts/fleet_player_chain_receipt.py',
              'docs/internals/data/fleet-player-chain-native.json',
              'docs/internals/data/fleet-command-ui.json','docs/internals/data/fleet-scene.json']:
        record(Path(n))
    out.write_text(json.dumps({'schema':'logh3.fleet-player-chain-game.v1',
       'parent_commit':'02328c82b8fb27e21ac4a9e89b1aa87205c560b6',
       'worktree':'round76 modified; exact source hashes below',
       'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
       'dosgolem_worktree':'clean; unchanged',
       'image':'logh3-ebiten:2','seed':0x12345678,
       'seed_setup':'test-only before each remake battle; original controlled scene has its own recorded seed',
       'commands':['tools/gui.sh xrun env LOGH3_FLEET_CHAIN_OUTPUT=/src/workplace/round76-chain-final-d go test ./internal/game ./internal/rules -run "^(TestFleetCommandRosterNative|TestFleetSceneGameNative|TestFleetPlayerCommandUI|TestNormalFleetPlayerChain|TestRefreshIntegrityKeepsAnnihilatedFleetWithScout)$" -count=1 -v',
                   'repeat with round76-chain-final-e',
                   'tools/gui.sh xrun python3 tools/scripts/fleet_player_chain_receipt.py --replay-overlay',
                   'tools/gui.sh xrun go run -overlay workplace/round76-fleet-replay-overlay.json tools/scripts/fleet_player_chain_replay.go -input workplace/round76-chain-final-d/normal-0/inputs.json -out workplace/round76-bootstrap-third-0',
                   'repeat faction 1; repeat final-e as bootstrap-fourth-0/1',
                   'tools/gui.sh test ./cmd/... ./internal/... -count=1'],
       'files':files,'comparisons':comparisons,'normal_cases':normal,
       'standard_tests':{'pass_packages':13,'no_test_packages':5},
       'limits':['normal original full-path parity unknown','real GUI mouse timing not sampled',
                 'fleet command marker visuals not compared','SFX playback not implemented or listened',
                 '056/092/062 remain DRAFT; issues 48/79 OPEN']},ensure_ascii=False,indent=2)+'\n')
    print(out.name,hashlib.sha256(out.read_bytes()).hexdigest(),len(files),'sources')
    sys.exit(0)
if sys.argv[1] == '--replay-overlay':
    import os
    source = root/'internal/game/battle.go'
    target = root/'workplace/round76-fleet-replay.go'
    overlay = root/'workplace/round76-fleet-replay-overlay.json'
    assert source.stat().st_uid == os.getuid() and target.parent.stat().st_uid == os.getuid()
    assert not target.exists() and not overlay.exists()
    target.write_text(source.read_text()+'''\n// 僅供正常操作重播固定seed，覆蓋檔不進產品。
func (g *Game) FleetChainProbeSeed(seed uint32) { g.battleRand = rules.NewRand(seed) }
''')
    overlay.write_text(json.dumps({'Replace':{str(source):str(target)}},indent=2)+'\n')
    print('覆蓋只新增測試期seed setter；原檔SHA-256',hashlib.sha256(source.read_bytes()).hexdigest())
    sys.exit(0)
if sys.argv[1] == '--ui':
    source = root/'docs/internals/data/fleet-player-chain-native.json'
    native = source.read_bytes()
    assert hashlib.sha256(native).hexdigest() == 'f569a16f863f9e54d3cda5fe0a8fe2c39566ec74864469d05b374fa29017c7eb'
    ram = (root/'workplace/oracle/normal79-r61-a/end-mem.bin').read_bytes()
    digest = hashlib.sha256(ram).hexdigest()
    assert digest == '6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4'
    ui = json.loads(native)['command_ui']
    ui.update(source=source.name, kind_codes_hex=ram[0x54b30+0x8e1a:0x54b30+0x8e1a+12].hex(),
              weapon_codes_hex=ram[0x54b30+0x8e2f:0x54b30+0x8e2f+3].hex(),
              source_ram_sha256=digest,
              source_address_space='dosgolem DS:8E1A/8E2F; IDA 3958:00A0/0114')
    out = root/'docs/internals/data/fleet-command-ui.json'
    encoded = (json.dumps(ui,ensure_ascii=False,indent=2)+'\n').encode()
    if out.exists():
        assert out.read_bytes() == encoded, 'UI資料與固定原版來源不符'
    else:
        out.write_bytes(encoded)
    print(out.name,hashlib.sha256(encoded).hexdigest(),'原版來源重生一致')
    sys.exit(0)
out = root / sys.argv[1]
assert not out.exists()
files, cases = {}, []
def read(path):
    p = root/path
    b = p.read_bytes()
    files[str(path)] = {'sha256':hashlib.sha256(b).hexdigest(),'bytes':len(b)}
    return b
for case in ['base','reverse','rotate','interleave','number']:
    seqs = []
    for side in ['a','b']:
        name = 'fleet79-r76-order-'+case+'-'+side
        folder = Path('workplace/oracle')/name
        plan = json.loads(read(Path('workplace')/(name+'-plan.json')))
        read(Path('workplace')/(name+'-flags.txt'))
        report = read(folder/'report.txt').decode()
        meta = read(folder/'meta.txt').decode()
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in meta
        assert '1DD6:0152 執行時的暫存器（前 0 次）' in report
        assert '3CD5:000C 執行時的暫存器（前 0 次）' in report
        section = report.split('1ED2:0001 執行時')[1].split('1DD6:0152')[0]
        seq = [(int(x,16)-0xc074)//122 for x in re.findall(r'SI=([0-9A-F]{4})',section)]
        assert len(seq)==7 and set(seq)==set(plan['roster'])
        seqs.append(seq)
        for f in ['input-mem.bin','end-mem.bin','screen.bin','screen.png','pal.bin','flags.txt']:
            b = read(folder/f)
            if side=='b' and f != 'flags.txt':
                other = files[str(folder.with_name(folder.name[:-1]+'a')/f)]
                assert hashlib.sha256(b).hexdigest()==other['sha256']
    assert seqs[0]==seqs[1]
    cases.append({'case':case,'roster':plan['roster'],'native_order':seqs[0],
                  'double_replay':True,'classification':'controlled roster diagnostic'})
ram = read(Path('workplace/oracle/normal79-r61-a/end-mem.bin'))
assert hashlib.sha256(ram).hexdigest()=='6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4'
ds = 0x54b30
def string(off):return ram[ds+off:].split(b'\0',1)[0]
ui = {'header_prefix_hex':string(0x44de).hex(),'header_suffix_hex':string(0x44e9).hex(),
      'weapon_formats_hex':[string(x).hex() for x in [0x44f8,0x4505,0x4512]],
      'auto_defence_hex':[string(x).hex() for x in [0x44c8,0x44cd]],
      'tables':[{'ds_offset_hex':hex(x),'bytes_hex':ram[ds+x:ds+x+n].hex()} for x,n in [(0x44b6,18),(0x44d2,12)]]}
read(Path('workplace/ida/battle79-r62-1DD6.json'))
read(Path('workplace/ida/battle79-r63-1ED2.json'))
out.write_text(json.dumps({'schema':'logh3.fleet-player-chain-native.v1',
    'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
    'seed':0x12345678,'files':files,'cases':cases,'command_ui':ui,
    'scope':'名冊有限排列及原版UI來源；正式玩家鏈由獨立Game收據驗收'},ensure_ascii=False,indent=2)+'\n')
print(out.name,hashlib.sha256(out.read_bytes()).hexdigest(),len(files),'sources')
