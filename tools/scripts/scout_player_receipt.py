"""正常索敵、報告差分及完整來源身份。Docker內執行。"""
import hashlib
import json
import os
from pathlib import Path
import struct
import sys
import zlib

assert Path.cwd() == Path('/src') and os.getuid() == 1000
mode = sys.argv[1]
if mode == '--refresh-editorial':
    # 首次快照保持原身份；只修文件文字及錯誤訊息，驗收數值不變。
    old = Path('docs/internals/data/scout-player.json')
    assert hashlib.sha256(old.read_bytes()).hexdigest() == 'daa2e17efc0d106b727e4c98b4ab52f32f27caecfde8f111d2902f7105caef8a'
    archived = Path('workplace/scout79-r101-first-index.json')
    assert not archived.exists()
    old.rename(archived)
    mode = '--freeze'
native = json.loads(Path('docs/internals/data/scout-fuel.json').read_text())

def identity(path):
    data = path.read_bytes()
    return {'bytes': len(data), 'sha256': hashlib.sha256(data).hexdigest()}

def create(path, data):
    assert path.parent.stat().st_uid == os.getuid()
    with path.open('xb') as stream:
        stream.write(data)

def write_json(path, value):
    create(path, (json.dumps(value, ensure_ascii=False, indent=2)+'\n').encode())

def png_palette(path):
    data = path.read_bytes()
    assert data[:8] == b'\x89PNG\r\n\x1a\n'
    at = 8
    while at < len(data):
        size = struct.unpack('>I', data[at:at+4])[0]
        if data[at+4:at+8] == b'PLTE':
            return data[at+8:at+8+size]
        at += size+12
    raise AssertionError('缺PNG palette')

if mode == '--compare':
    prefix = 'workplace/scout79-r83-fleet12-original'
    original = Path(prefix+'.screen.bin').read_bytes()
    palette = Path(prefix+'.palette.bin').read_bytes()
    for suffix in ('screen.bin', 'palette.bin', 'png'):
        name = prefix+'.'+suffix
        assert identity(Path(name))['sha256'] == native['sources'][name]['sha256']
    rectangle = json.loads(Path('docs/internals/data/turn-window-panel.json').read_text())['battle_report']
    window = (rectangle['cell_x']*16, 40+rectangle['cell_y']*20, (rectangle['cols']+2)*16, (rectangle['rows']+2)*20)
    results = {}
    for label, folder in [('before','scout79-r101-normal-i'), ('after','scout79-r101-normal-j')]:
        p = Path('workplace',folder,'report-4.bin')
        remake = p.read_bytes()
        colours = png_palette(p.with_suffix('.png'))
        assert colours[:48] == palette[:48] and len(remake) == len(original) == 307200
        stats = {}
        for name, (x,y,w,h) in [('full',(0,0,640,480)), ('report',window)]:
            different = []
            maximum = 0
            for yy in range(y,y+h):
                for xx in range(x,x+w):
                    at = yy*640+xx
                    if original[at] == remake[at]:
                        continue
                    different.append((xx,yy))
                    maximum = max(maximum, *(abs(palette[original[at]*3+c]-colours[remake[at]*3+c]) for c in range(3)))
            bbox = None
            if different:
                bbox = [min(xx for xx,yy in different),min(yy for xx,yy in different),max(xx for xx,yy in different),max(yy for xx,yy in different)]
            stats[name] = {'rect':[x,y,w,h], 'pixels':w*h,'same':w*h-len(different),'different':len(different),'maximum_channel_difference':maximum,'bbox':bbox}
        assert stats['report']['different'] == (137 if label=='before' else 0)
        # 標準PNG串流；未裁切全圖，所有差異保留。
        rows = b''.join(b'\0'+bytes(255 if original[y*640+x]!=remake[y*640+x] else 0 for x in range(640)) for y in range(480))
        def chunk(kind, data):
            return struct.pack('>I',len(data))+kind+data+struct.pack('>I',zlib.crc32(kind+data)&0xffffffff)
        png = b'\x89PNG\r\n\x1a\n'+chunk(b'IHDR',struct.pack('>IIBBBBB',640,480,8,0,0,0,0))+chunk(b'IDAT',zlib.compress(rows))+chunk(b'IEND',b'')
        create(Path(f'workplace/scout79-r101-report-{label}-diff.png'),png)
        results[label] = stats
    results.update({'classification':'nearby-state；原版為第83輪控制報告，正常玩家背景不同，不宣稱完整同態',
                    'original_source':prefix, 'cursor_screen':[240,240], 'mask':None})
    write_json(Path('workplace/scout79-r101-report-comparison.json'),results)
    print(json.dumps(results,ensure_ascii=False))
elif mode in ('--freeze','--check'):
    target = Path('docs/internals/data/scout-player.json')
    if mode == '--freeze':
        assert not target.exists()
        a, b = (Path(f'workplace/scout79-r101-normal-{r}') for r in ('j','k'))
        for p in a.rglob('*'):
            if p.is_file():
                assert p.read_bytes() == (b/p.relative_to(a)).read_bytes(),str(p)
        normal = json.loads((a/'receipt.json').read_text())
        assert len([i for i in normal['inputs'] if i['kind']=='seed']) == 1
        observations = {o['name']:o['state'] for o in normal['observations']}
        for turn in range(1,8):
            s = observations[f'turn-{turn}']
            fleet = bytes.fromhex(s['Tables']['fleet'])[12*122:13*122]
            assert (fleet[0x3a],fleet[0x3b]) == ((0,turn) if turn<4 else (255,7-turn))
        for s in observations.values():
            f = bytes.fromhex(s['Tables']['fleet'])
            assert int.from_bytes(f[12*122+0x38:12*122+0x3a],'little') == 81
        report = observations['report-4']
        case = next(c for c in native['cases'] if c['fleet']==12 and c['mode']==0 and c['fuel']==3 and c['human_mask']==1)
        expected = next(c for c in case['calls'] if c['entry']=='3826:0405')
        assert report['ReportText'] == expected['text_hex'] and report['ReportFace']==12
        assert len(observations['redeployed']['Units']) == len(observations['deployed']['Units']) == 49
        for folder in (a,b):
            cooldown = (folder/'cooldown-save/FLEET.DT0').read_bytes()
            scout = (folder/'cooldown-save/SHIP.DT0').read_bytes()[81*78:82*78]
            assert cooldown[12*122+0x3a:12*122+0x3c] == bytes([255,3]) and scout[0x1e]==255
            deltas = {}
            for name in ('STELLAR','PLANET','ADMIRAL','FLEET','SHIP','WORK'):
                old = (folder/'player-save'/f'{name}.DT0').read_bytes()
                new = (folder/f'{name}.DT0').read_bytes()
                deltas[name]=[[i,x,y] for i,(x,y) in enumerate(zip(old,new)) if x!=y]
                assert len(old)==len(new)
            assert deltas['STELLAR']==[[3318,0,1],[3527,22,25],[3653,37,39]]
            assert all(not v for k,v in deltas.items() if k!='STELLAR')
        gui = json.loads(Path('workplace/scout79-r101-gui-d/receipt.json').read_text())
        assert gui['test_tps']==60 and gui['normal_scout_expired'] and gui['public_save_load_equal_plan']
        assert len(gui['comparisons'])==14 and all(c['x11_different']==0 and c['game_index_equal'] for c in gui['comparisons'])
        log = Path('workplace/round101-standard-fixed-gui.log').read_text()
        assert len([l for l in log.splitlines() if l.startswith('ok  ')])==13 and log.count('[no test files]')==5 and 'FAIL' not in log
        paths = {Path(p) for p in ('tools/scripts/scout_player_prepare.py','tools/scripts/scout_player_probe.go','tools/scripts/scout_player_gui.py','tools/scripts/scout_player_receipt.py',
                 'internal/game/game.go','internal/game/scoutfuel_test.go','internal/game/scoutfuel.go','internal/game/scoutcmd.go','internal/game/battlemovement.go',
                 'internal/game/fleetcombat.go','cmd/gin3/main.go','tools/gui.sh','tools/scripts/fleet_player_gui_prepare.py','tools/scripts/fleet_player_gui_driver.py',
                 'tools/scripts/fleet_player_gui_capture.go','tools/scripts/sound_tail_input.go','docs/spec/107-scout-fuel.md','docs/spec/006-unrest-at-load.md',
                 'docs/internals/data/scout-fuel.json','docs/internals/data/turn-window-panel.json',
                 'docs/goal/2026-10-05-scout-player-round-101.md','docs/re/435-scout-player-report-cursor.md','docs/playtest/089-scout-player.md')}
        paths.update(Path(p) for p in native['sources'] if p.startswith('workplace/scout79-r83-fleet12-original.'))
        for p in Path('workplace').glob('scout79-r101*'):
            if p.is_file():paths.add(p)
            else:paths.update(f for f in p.rglob('*') if f.is_file())
        paths.update(Path(p) for p in ('workplace/round101-report-before.log','workplace/round101-report-after.log','workplace/round101-standard-gui.log','workplace/round101-standard-fixed-gui.log','workplace/order79-r93-input','workplace/round79-capture'))
        paths.update(Path(p) for p in Path('workplace').glob('round79-fleet-gui-*-scout101c.*'))
        value = {'schema':'logh3.scout-player.v1','date':'2026-10-05','issue':79,'parent_commit':'e2a1d01890bc608de8e987277d31925d3a2847ec',
                 'classification':'remake正常玩家及60Hz真實GUI；原版完整正常回合同態未驗',
                 'scenario':'S3','faction':0,'seed':0x12345678,'seed_method':normal['seed_method'],
                 'normal':{'inputs':len(normal['inputs']),'repeats':2,'milestones':14,'fleet':12,'scout_pool':81,'deployed_units':49,'scout_reports':1,'cooldown_turns':3,'same_pool_redeployed':True,'public_save_load':True,'load_differences':deltas},
                 'gui':{'test_tps':60,'actions':len(gui['actions']),'frames':14,'pixels':4300800,'different':0,'camera_gestures':len(gui['camera_gestures'])},
                 'report_comparison':json.loads(Path('workplace/scout79-r101-report-comparison.json').read_text()),
                 'limitations':['正常原版整回合同態未新增','原版報告為控制狀態，完整背景差異保持','只抽帝國S3索敵路徑，其他自然分支未完成','載入回命令階段，不稱戰場續戰','061被動戰術CPU保持','MIDI與一像素裁決仍待答'],
                 'sources':{str(p):identity(p) for p in sorted(paths)}}
        write_json(target,value)
    value = json.loads(target.read_text())
    for name, expected in value['sources'].items():
        assert identity(Path(name))==expected,name
    print(len(value['sources']),'來源符合，索引SHA-256',identity(target)['sha256'])
else:
    raise SystemExit('未知模式')
