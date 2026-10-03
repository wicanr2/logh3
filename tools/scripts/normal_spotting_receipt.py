"""#79正常首次索敵：保存原版雙跑、正式玩家輸出與未豁免完整差異。"""
import hashlib
import json
import os
from pathlib import Path
import re
import struct

from pngout import game_palette, write_png


def sha(p):
    return hashlib.sha256(Path(p).read_bytes()).hexdigest()


def owned(p):
    assert Path(p).stat().st_uid == os.getuid(), str(p)


def compare_files(a, b, names):
    for n in names:
        owned(a/n)
        owned(b/n)
        assert (a/n).read_bytes() == (b/n).read_bytes(), n


def fingerprint(p):
    return {q.relative_to(p).as_posix(): {'sha256': sha(q), 'bytes': q.stat().st_size, 'uid': q.stat().st_uid}
            for q in sorted(p.rglob('*')) if q.is_file()}


def table_diff(a, b, stride):
    assert len(a) == len(b)
    diff = [{'record': i//stride, 'offset_hex': hex(i % stride), 'original': x, 'remake': y}
            for i, (x, y) in enumerate(zip(a, b)) if x != y]
    return {'bytes': len(a), 'different_bytes': len(diff), 'differences': diff,
            'original_sha256': hashlib.sha256(a).hexdigest(), 'remake_sha256': hashlib.sha256(b).hexdigest()}


def pixels(a, b, pal, rect):
    x, y, w, h = rect
    indices = [j*640+i for j in range(y, y+h) for i in range(x, x+w)]
    diff = [i for i in indices if a[i] != b[i]]
    return {'rectangle_xywh': rect, 'compared_pixels': len(indices), 'different_index_pixels': len(diff),
            'different_rgb_pixels': sum(pal[a[i]*3:a[i]*3+3] != pal[b[i]*3:b[i]*3+3] for i in diff),
            'maximum_channel_difference': max((abs(pal[a[i]*3+k]-pal[b[i]*3+k]) for i in diff for k in range(3)), default=0),
            'difference_bbox_inclusive': [min(i % 640 for i in diff), min(i//640 for i in diff), max(i % 640 for i in diff), max(i//640 for i in diff)] if diff else None}


target = Path('docs/internals/data/normal-spotting.json')
assert not target.exists(), '拒絕覆寫固定收據'
owned(target.parent)
originals = [Path('workplace/oracle/normal79-r61-'+s) for s in 'ab']
original_names = [p.name for p in originals[0].iterdir() if p.name not in ('meta.txt', 'flags.txt', 'report.txt')]
assert set(original_names) == {'battle.bin', 'pal.bin', 'screen.bin', 'screen.png'} | {
    stage+suffix for stage in ('entry','label','seat1','end') for suffix in ('-mem.bin','.bin','.png')}
compare_files(*originals, original_names)
for p in originals:
    report = (p/'report.txt').read_text()
    assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in (p/'meta.txt').read_text()
    for addr, n in [('181B:0262', 1), ('181B:028C', 1), ('2526:00F8', 2),
                    ('2526:033B', 0), ('2526:04B1', 0), ('3826:0298', 0), ('3765:006C', 0), ('25DB:0E78', 0)]:
        assert f'{addr} 執行時的暫存器（前 {n} 次）' in report
    for step in (483778106, 483798430, 484103337, 484357728):
        assert f'#{step} AX=' in report
    for n in ['battle.bin', 'entry-mem.bin', 'label-mem.bin', 'seat1-mem.bin', 'end-mem.bin']:
        assert (p/n).stat().st_size == 1048576
players = [Path('workplace/round61-normal-checked-'+s) for s in 'ab']
player_names = sorted(fingerprint(players[0]))
assert len(player_names) == 24, len(player_names)
assert sorted(fingerprint(players[1])) == player_names
compare_files(*players, player_names)
for s in 'ab':
    log = Path(f'workplace/round61-normal-checked-{s}.log').read_text()
    assert '--- PASS: TestNormalSpottingInformationAndSave' in log and '--- FAIL:' not in log and '--- SKIP:' not in log
standard = Path('workplace/round61-standard.log').read_text()
assert len(re.findall(r'^(ok\s|\?\s)', standard, re.M)) == 18 and 'FAIL' not in standard

tables = [('WORK', 0x5da44, 230, 1), ('FLEET', 0x60ba4, 122, 32), ('SHIP', 0x668b2, 78, 352),
          ('STELLAR', 0x5f2cc, 126, 32), ('PLANET', 0x53132, 82, 81)]
comparisons = {}
pal = bytes(game_palette(str(originals[0]/'label.png')))
assert len(pal) == 48
# Game.Palette保存6-bit DAC；兩側PNG已經走正式8-bit RGB轉換。
# 直接核PNG有效色盤，不把不同單位的DAC數值與RGB相比。
assert bytes(game_palette(str(players[0]/'label/screen.png'))) == pal
for stage in ['label', 'end']:
    mem = (originals[0]/(stage+'-mem.bin')).read_bytes()
    assert bytes(game_palette(str(players[0]/stage/'screen.png'))) == pal
    original_seed = struct.unpack_from('<I', mem, 0x5e00c)[0]
    assert original_seed == struct.unpack_from('<I', (originals[0]/'entry-mem.bin').read_bytes(), 0x5e00c)[0]
    assert list(struct.unpack_from('<HH', mem, 0x5e724)) == [192, 100]
    differences = {}
    for n, addr, stride, count in tables:
        a = mem[addr:addr+stride*count]
        b = (players[0]/stage/(n+'.bin')).read_bytes()
        differences[n] = table_diff(a, b, stride)
    ship = (players[0]/stage/'SHIP.bin').read_bytes()
    for pool in range(352):
        a = mem[0x668b2+pool*78:]
        b = ship[pool*78:]
        assert a[0x16]&1 == b[0x16]&1 and a[0x17:0x19] == b[0x17:0x19]
    a, b = (originals[0]/(stage+'.bin')).read_bytes(), (players[0]/stage/'screen.bin').read_bytes()
    assert len(a) == len(b) == 307200
    base = f'workplace/round61-normal-{stage}'
    for suffix in ('-diff.png', '-pair.png'):
        assert not Path(base+suffix).exists()
    write_png(base+'-diff.png', 640, 480, bytes(int(x != y) for x, y in zip(a, b)), bytes([0, 0, 0, 255, 0, 0]))
    write_png(base+'-pair.png', 1280, 480, b''.join(a[y*640:(y+1)*640]+b[y*640:(y+1)*640] for y in range(480)), pal)
    comparisons[stage] = {'classification': 'nearby-state; 正常玩家路徑，入口移動狀態及顯示指標未同態',
                          'original_seed': f'{original_seed:08X}',
                          'remake_seed':json.loads((players[0]/'states.json').read_text())[stage]['seed'],
                          'tables': differences,
                          'frames': {n: pixels(a,b,pal,r) for n,r in {'full':[0,0,640,480], 'field':[0,0,448,360], 'right':[448,0,192,480], 'bottom':[0,360,448,120]}.items()},
                          'images': {suffix: sha(base+suffix) for suffix in ('-diff.png','-pair.png')}}
entry, end = [(originals[0]/(s+'-mem.bin')).read_bytes() for s in ('label','end')]
for n, addr, stride, count in tables:
    assert entry[addr:addr+stride*count] == end[addr:addr+stride*count]
    assert (players[0]/'label'/(n+'.bin')).read_bytes() == (players[0]/'end'/(n+'.bin')).read_bytes()

sources = ['tools/normal_spotting_oracle.sh', 'tools/scripts/normal_spotting_receipt.py',
           'internal/game/normalspotting_test.go', 'internal/game/normalspotting_player_test.go',
           'internal/game/battlespotting.go', 'internal/game/battlemovement.go', 'internal/game/battle.go',
           'internal/game/passivebattle.go', 'internal/rules/fleetmove.go', 'internal/rules/battle.go']
toolchain = [{'name': d['RepoTags'], 'id': d['Id']} for d in map(json.loads,Path('workplace/round61-toolchain-inspect.jsonl').read_text().splitlines())]
data = {'schema':'logh3.normal-spotting.v1', 'date':'2026-10-03',
        'parent_commit':'eb6814ab39daef76f2531e4020f64c15bf50679d',
        'issue':'https://github.com/wicanr2/logh3/issues/79', 'spec':'docs/spec/087-battle-spotting-phase.md',
        'status':'limited-normal-player-information-validation; whole-phase parity unresolved',
        'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
        'source_snapshot_sha256':'0155f63e373582b76a3b7ee54eccbef5ae2626481e09a68550b5fddc580146d7',
        'original_executable_sha256':sha('org_game/extracted/GIN3.COM'),
        'original_bundle_sha256':sha('org_game/银河英雄传说3SP.zip'),
        'toolchain':toolchain, 'source_sha256':{p:sha(p) for p in sources},
        'original_runs':[{'path':str(p),'argv':(p/'flags.txt').read_text().splitlines(),'files':fingerprint(p)} for p in originals],
        'normal_runs':[{'path':str(p),'files':fingerprint(p),'states':json.loads((p/'states.json').read_text())} for p in players],
        'input_contract':{'original':'正常S3父狀態→系統544,85→實行560,85→報告240,200→戰術選單100,60→實行100,70',
                          'remake':'正常bootstrap→主選單→138→S3→帝國→系統544,125→實行560,125→報告240,240→戰術選單100,100→實行100,110',
                          'seeds':'實行前策略D3DC167E，正常首戰B6E4B78A，第一次戰術實行前12345678；未重擲',
                          'policy':'正式061被動CPU保持；沒有原版RAM注入。只讀after觀測先捕獲再呼叫原續接。'},
        'limited_result':{'grade':'proven; 限固定正常S3首回合', 'information_records':352,
                          'fields':'SHIP+16 bit0 Seen；+17 Detail；+18 Hold', 'five_full_table_phase_delta_bytes':0,
                          'reports':0, 'resume_count':1, 'save_reload':'complete SHIP and FLEET identical'},
        'comparisons':comparisons,
        'limitations':['正常上游位置／移動旗標不同，061被動CPU保持，近似狀態不升格。','策略／战术seed只在實行前固定；索敵前原版AFCFD1BC與remake12345678不同，最早消費者未定位。','顯示物件指標及FF排序鍵仍未裁決。','未新增cmd/gin3真實X11、正常發現報告、完整087／八步或發行驗收。'],
        'logs':{p:sha(p) for p in ['workplace/round61-normal-a.log','workplace/round61-normal-final-a.log','workplace/round61-normal-checked-a.log','workplace/round61-normal-checked-b.log','workplace/round61-standard.log']}}
target.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
print('收據SHA-256',sha(target))
print('正常階段完整差異', {s:{'tables':{n:v['different_bytes'] for n,v in c['tables'].items()},'pixels':c['frames']['full']['different_index_pixels']}for s,c in comparisons.items()})
