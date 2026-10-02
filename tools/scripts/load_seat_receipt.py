"""第46輪088讀檔座位；入口研究384與internals索引。只驗原始收據，不猜補差異。"""
import hashlib
import json
from pathlib import Path
import struct

from pngout import game_palette, write_png


def sha(p):
    return hashlib.sha256(Path(p).read_bytes()).hexdigest()


def stats(a, b, pal, rect):
    x, y, w, h = rect
    offsets = [j * 640 + i for j in range(y, y + h) for i in range(x, x + w)]
    different = [i for i in offsets if a[i] != b[i]]
    rgb = [i for i in different if pal[a[i]*3:a[i]*3+3] != pal[b[i]*3:b[i]*3+3]]
    return {'rectangle_xywh': rect, 'compared_pixels': len(offsets),
            'same_pixels': len(offsets)-len(different), 'different_index_pixels': len(different),
            'different_rgb_pixels': len(rgb),
            'maximum_channel_difference': max((abs(pal[a[i]*3+k]-pal[b[i]*3+k]) for i in different for k in range(3)), default=0),
            'difference_bbox_inclusive': [min(i%640 for i in different), min(i//640 for i in different), max(i%640 for i in different), max(i//640 for i in different)] if different else None}


root = Path('workplace/oracle')
program = '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918'
assert sha('org_game/extracted/GIN3.COM') == program
bundle = Path('org_game/银河英雄传说3SP.zip')
assert sha(bundle) == 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673'
cases = ['menu', 'empire', 'alliance', 'versus', 'hover0', 'hover1', 'hover2', 'vsconfirm', 'vshandoff']
originals = []
for kind in cases:
    a, b = [root / ('normal89-r46-'+kind+'-'+n) for n in 'ab']
    memname = 'menu-mem.bin' if kind == 'menu' else 'loaded-mem.bin'
    names = [memname, 'screen.bin', 'screen.png', 'pal.bin']
    for n in names:
        assert (a/n).read_bytes() == (b/n).read_bytes(), (kind, n)
    assert (a/memname).stat().st_size == 1048576
    for p in (a, b):
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in (p/'meta.txt').read_text()
        argv = (p/'flags.txt').read_text().splitlines()
        assert argv[argv.index('-tick')+1] == '20000'
        assert not any(v in argv for v in ('-poke', '-patch'))
        more = ['menu.state'] if kind == 'menu' else []
        originals.append({'path': str(p), 'classification': 'cold normal player input' if kind == 'menu' else 'continuation of own cold menu snapshot; normal selection',
                          'command_argv': argv, 'hashes': {n: sha(p/n) for n in names+more+['meta.txt', 'flags.txt', 'report.txt']}})


def ram(kind):
    return (root/('normal89-r46-'+kind+'-a')/('menu-mem.bin' if kind == 'menu' else 'loaded-mem.bin')).read_bytes()


menu = ram('menu')
assert struct.unpack_from('<H', menu, 0x5da48)[0] == 0
controls = {'empire': bytes.fromhex('00000100010001001f'),
            'alliance': bytes.fromhex('010001000101001f00'), 'versus': bytes.fromhex('02000000000001001f')}
for kind, control in controls.items():
    assert ram(kind)[0x5da50:0x5da59] == control
    assert struct.unpack_from('<H', ram(kind), 0x5da48)[0] == 380
for kind in ('hover0', 'hover1', 'hover2'):
    assert ram(kind)[0x5da44:0x5db2a] == menu[0x5da44:0x5db2a]

desc = menu[0x54b30+0x22e:0x54b30+0x246]
assert desc.hex() == '01001102b35401001802b35401001f02b35401002602b354'
captions = []
for i in range(4):
    flags, off, seg = struct.unpack_from('<HHH', desc, i*6)
    assert flags == 1
    loc = seg*16+off
    end = menu.index(0, loc)
    captions.append({'original_descriptor_ds_offset_hex': hex(0x22e+i*6), 'far_pointer': f'{seg:04X}:{off:04X}',
                     'linear_hex': hex(loc), 'raw_hex': menu[loc:end+1].hex(), 'shown': i<3, 'grade': 'proven'})
assert [c['raw_hex'] for c in captions[:3]] == ['abd22020b0ea00', 'a6502020b7f900', 'b9ef2020bed400']

ida_path = Path('workplace/ida/normal89-r46-seat.json')
ida = json.loads(ida_path.read_text())
assert ida['kernel_version'] == '9.4' and ida['uid'] == 1000
assert ida['original_name'] == 'sub_10206' and ida['function_boundary'] == ['0x10206', '0x10342']
assert hashlib.sha256(bytes(0x100)+menu[0x42f0:0x4f80]).hexdigest() == ida['input_sha256'] == 'dc7d9ecd9fbe218720ecc85e29f46b858be9935b0aaa2b7d3c2487b8c879180a'

edges = []
for name, inside, xy in [('x302', False, [302,208]), ('x303',False,[303,208]), ('x304',True,[304,208]),
                         ('x334',True,[334,208]), ('x335',True,[335,208]), ('x336',False,[336,208]),
                         ('y199',False,[318,199]), ('y200',True,[318,200]), ('y259',True,[318,259]), ('y260',False,[318,260])]:
    p=Path('workplace/explore/normal89-r46-edge-'+name)
    m=(p/'loaded-mem.bin').read_bytes()
    assert struct.unpack_from('<H',m,0x5da48)[0] == (380 if inside else 0)
    edges.append({'screen_xy':xy,'loaded':inside,'grade':'proven measured single continuation; auxiliary edge receipt',
                  'path':str(p),'command_argv':(p/'flags.txt').read_text().splitlines(),
                  'hashes':{n:sha(p/n) for n in ['meta.txt','flags.txt','report.txt','loaded-mem.bin','screen.bin','screen.png','pal.bin']}})

players = [Path('workplace/round46-player-'+s) for s in 'cd']
files = sorted(p.relative_to(players[0]).as_posix() for p in players[0].rglob('*') if p.is_file())
assert files == sorted(p.relative_to(players[1]).as_posix() for p in players[1].rglob('*') if p.is_file())
for n in files:
    assert (players[0]/n).read_bytes() == (players[1]/n).read_bytes(), n
remakes = [{'path':str(p),'command_argv':['tools/gui.sh','xrun','go','run','tools/scripts/load_seat_player.go','-out',str(p)],
            'hashes':{n:sha(p/n) for n in files}} for p in players]

comparisons = {}
tables = [('WORK',0x5da44,230),('FLEET',0x60ba4,3904),('SHIP',0x668b2,27456),('STELLAR',0x5f2cc,4032)]
for kind, control in controls.items():
    work=(players[0]/kind/'loaded/WORK.DT0').read_bytes()
    assert work[0xc:0x15] == control
    assert struct.unpack_from('<H',work,4)[0] == 380
    for stage, original in [('loaded',kind)]+([('confirmed','vsconfirm'),('handoff','vshandoff')] if kind=='versus' else []):
        p=players[0]/kind/stage
        o=root/('normal89-r46-'+original+'-a')
        a,b=(o/'screen.bin').read_bytes(),(p/'screen.bin').read_bytes()
        pal=bytes(game_palette(str(o/'screen.png')))
        assert len(a)==len(b)==307200 and len(pal)==48
        base='workplace/round46-'+kind+'-'+stage
        write_png(base+'-remake.png',640,480,b,pal)
        write_png(base+'-diff.png',640,480,bytes(int(x!=y) for x,y in zip(a,b)),bytes([0,0,0,255,0,0]))
        data={}
        for name,addr,length in tables:
            raw=ram(original)[addr:addr+length]
            remake=(p/(name+'.DT0')).read_bytes()
            assert len(raw)==len(remake)==length
            different=[i for i in range(length) if raw[i]!=remake[i]]
            data[name]={'bytes':length,'different_bytes':len(different),'offsets':different,
                        'original_hex':raw.hex(),'remake_hex':remake.hex()}
        comparisons[kind+'/'+stage]={'classification':'nearby-state; full images and complete byte differences retained',
                 'frame':stats(a,b,pal,[0,0,640,480]),'tables':data,
                 'images':{base+suffix:sha(base+suffix) for suffix in ('-remake.png','-diff.png')}}

menus = {}
for sel in range(-1,3):
    kind='menu' if sel==-1 else 'hover'+str(sel)
    p=root/('normal89-r46-'+kind+'-a')
    a=(p/'screen.bin').read_bytes()
    b=(players[0]/'empire'/('menu'+str(sel)+'.bin')).read_bytes()
    proto=Path('workplace/round46-menu-proto2/menu'+str(sel)+'.bin').read_bytes()
    pal=bytes(game_palette(str(p/'screen.png')))
    reg={'full':stats(a,b,pal,[0,0,640,480]),'popup':stats(a,b,pal,[289,181,62,98])}
    assert reg['popup']['different_index_pixels'] == 0
    assert stats(a,proto,pal,[289,181,62,98])['different_index_pixels'] == 0
    base='workplace/round46-menu'+str(sel)
    write_png(base+'-remake.png',640,480,b,pal)
    write_png(base+'-diff.png',640,480,bytes(int(x!=y) for x,y in zip(a,b)),bytes([0,0,0,255,0,0]))
    menus[str(sel)]={'classification':'nearby-state whole frame; proven static full-popup reproduction',
                    'regions':reg,'prototype_full':stats(a,proto,pal,[0,0,640,480]),
                    'images':{base+suffix:sha(base+suffix) for suffix in ('-remake.png','-diff.png')}}

for name,addr,length in tables[1:3]:
    assert ram('versus')[addr:addr+length] == ram('vshandoff')[addr:addr+length]
    assert (players[0]/('versus/loaded/'+name+'.DT0')).read_bytes() == (players[0]/('versus/handoff/'+name+'.DT0')).read_bytes()

normal = [Path('workplace/round46-normal-'+s) for s in 'ab']
normal_files = ['probe.json']+[stage+'/'+n for stage in ('loaded','entry') for n in
    ('STELLAR.DT0','PLANET.DT0','ADMIRAL.DT0','FLEET.DT0','SHIP.DT0','WORK.DT0','SAVEDATA.IDX','screen.bin')]
for n in normal_files:
    assert (normal[0]/n).read_bytes() == (normal[1]/n).read_bytes(), n
normal_meta = json.loads((normal[0]/'probe.json').read_text())
assert normal_meta['strategy_seed_before'] == 'D3DC167E' and normal_meta['strategy_seed_after'] == 'B6E4B78A'
assert normal_meta['date'] == [799,3,4]
fleet=(normal[0]/'entry/FLEET.DT0').read_bytes()
ships=(normal[0]/'entry/SHIP.DT0').read_bytes()
assert [i for i in range(32) if fleet[i*122+4] != 255 and fleet[i*122+6] == 26] == [12,16,17,28,29,30,31]
assert sum(ships[i*78+2] == 26 for i in range(352)) == 55
normal_tables={}
orig_entry=Path('workplace/oracle/battle79-normal-r44-a/entry-mem.bin').read_bytes()
for name,addr,length in tables:
    a=orig_entry[addr:addr+length]
    b=(normal[0]/'entry'/(name+'.DT0')).read_bytes()
    diff=[i for i in range(length) if a[i]!=b[i]]
    normal_tables[name]={'different_bytes':len(diff),'offsets':diff,'original_hex':a.hex(),'remake_hex':b.hex()}
    if name!='WORK':
        assert b == Path('workplace/round45-normal-fixed-a/entry',name+'.DT0').read_bytes()
normal_orig=Path('workplace/oracle/battle79-normal-r44-a/screen.bin').read_bytes()
normal_game=(normal[0]/'entry/screen.bin').read_bytes()
normal_pal=bytes(game_palette('workplace/oracle/battle79-normal-r44-a/screen.png'))
normal_images={}
for suffix,px,pal in [('-remake.png',normal_game,normal_pal),('-diff.png',bytes(int(a!=b) for a,b in zip(normal_orig,normal_game)),bytes([0,0,0,255,0,0]))]:
    p='workplace/round46-normal-entry'+suffix
    write_png(p,640,480,px,pal)
    normal_images[p]=sha(p)
gui={}
for name in ('hover','empire','alliance','versus','empire2x'):
    base='workplace/round46-gui-'+name
    state=Path(base+'.txt').read_text()
    expected='StateLoadFaction' if name=='hover' else 'StateVersusInput' if name=='versus' else 'StateCommandPhase'
    assert 'state='+expected+'\n' in state
    gui[name]={'command_argv':['tools/gui.sh','xrun','bash','tools/load_seat_gui.sh'],'state_receipt':state,
               'hashes':{base+s:sha(base+s) for s in ('.bin','.png','.txt','.log')}}
source_files=['internal/game/game.go','internal/game/loadfaction.go','internal/game/loadfaction_test.go',
              'internal/screen/commandmenu.go','docs/internals/data/battle-command-menus.json',
              'tools/load_seat_oracle.sh','tools/scripts/load_seat_ida.py','tools/scripts/load_seat_probe.go',
              'tools/scripts/load_seat_player.go','tools/scripts/load_seat_receipt.py','tools/load_seat_gui.sh',
              'tools/scripts/battle_normal_probe.go','tools/scripts/battle_normal_probe_seed.py','internal/game/rebellion.go',
              'internal/bootstrap/load.go','internal/bootstrap/saveload_test.go','internal/game/game_test.go',
              'internal/game/aimode_test.go','internal/game/fleetsave_test.go','internal/game/normal_ai_gate_test.go',
              'internal/game/planscreens_test.go','internal/game/region_supply_timing_test.go','internal/game/runtime_fields_test.go']
logs=['workplace/round46-load-tests.log','workplace/round46-load-retest.log','workplace/round46-load-final.log',
      'workplace/round46-standard-tests.log','workplace/round46-standard-retest.log']
final=Path(logs[-1]).read_text()
assert 'FAIL' not in final and sum(line.startswith(('ok  ','?   ')) for line in final.splitlines()) == 18
assert 'FAIL' not in Path(logs[2]).read_text() and 'SKIP' not in Path(logs[2]).read_text()
data={'schema':'logh3.load-seat-selection.v1','grade':'proven selection popup and nine controller bytes; whole entry unknown',
      'parent_head':'04ae7d659321b0b568287852b7afc1dba30121f1','remake_worktree':'dirty; production source hashes recorded',
      'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c','dosgolem_dirty':False,
      'program_sha256':program,'bundle_sha256':sha(bundle),
      'toolchain':{'go_image':'golang:1.24-bookworm','go_image_id':'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac',
                   'gui_image':'logh3-ebiten:2','gui_image_id':'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
                   'ida_image':'ida-pro-9.4-idapython:locked-v1','ida_image_id':'sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780'},
      'original_runs':originals,'edge_auxiliary_receipts':edges,'remake_runs':remakes,'menu_comparisons':menus,
      'normal_loaded_comparisons':comparisons,'captions':captions,'descriptor_hex':desc.hex(),
      'normal_first_battle':{'classification':'nearby-state; prior fixed seed overlay only provides seed setup and observation',
                            'metadata':normal_meta,'tables':normal_tables,
                            'frame':stats(normal_orig,normal_game,normal_pal,[0,0,640,480]),'images':normal_images,
                            'runs':[{'path':str(p),'command_argv':['tools/gui.sh','xrun','go','run','-overlay','workplace/round44-normal-overlay.json','tools/scripts/battle_normal_probe.go','-out',str(p)],
                                     'hashes':{n:sha(p/n) for n in normal_files}} for p in normal],
                            'original_receipt_metadata':{'path':'workplace/oracle/battle79-normal-r44-a/meta.txt','sha256':sha('workplace/oracle/battle79-normal-r44-a/meta.txt'),
                               'command_argv':Path('workplace/oracle/battle79-normal-r44-a/flags.txt').read_text().splitlines(),'memory_sha256':sha('workplace/oracle/battle79-normal-r44-a/entry-mem.bin')},
                            'seed_overlay_hashes':{p:sha(p) for p in ['workplace/round44-normal-overlay.json','workplace/round44-normal-seed.go']}},
      'actual_x11_gui':gui,
      'controller_fields':{'structure':'WORK','offset_hex':'0x0C','bytes':9,'meaning':['mode','configured_cpu_0','configured_cpu_1','live_cpu_0','live_cpu_1','seat_faction_0','seat_faction_1','seat_capital_0','seat_capital_1'],
                           'mode_values':{k:v.hex() for k,v in controls.items()},'grade':'proven'},
      'ida_export':{'path':str(ida_path),'sha256':sha(ida_path),'input_sha256':ida['input_sha256'],'instructions':ida['instructions']},
      'original_save_sources':{str(p):sha(p) for p in Path('org_game/extracted').glob('*.DTA')},
      'source_hashes':{p:sha(p) for p in source_files},'test_logs':{p:sha(p) for p in logs},
      'limits':['完整088仍DRAFT；日期380/381與戰場全表差異另續',
                '正常冷讀檔背景初始世界不同，保留整圖與差分，不靠popup零差升格',
                '雙人只驗已載入這一輪帝國與同盟交接、同一全域FLEET/SHIP池；跨月仍未知',
                'WORK九個控制bytes已驗；活動陣營、座位、runtime指標等其他bytes完整列出，未猜補']}
target=Path('docs/internals/data/load-seat-selection.json')
target.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
print('formal original pairs',len(cases),'remake double identical files',len(files))
print('menu',json.dumps({k:v['regions'] for k,v in menus.items()},ensure_ascii=False))
print('loaded',json.dumps({k:{name:d['different_bytes'] for name,d in v['tables'].items()} for k,v in comparisons.items()}))
print('receipt SHA256',sha(target))
