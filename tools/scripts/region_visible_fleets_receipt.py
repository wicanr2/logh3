"""#59：凍結可見艦隊數的原版、正式資料鏈、正常GUI與IDA證據。"""
import hashlib
import json
from pathlib import Path
from pngout import game_palette, write_png

root = Path('/src')
target = root / 'docs/internals/data/region-visible-fleets.json'
diffdir = root / 'workplace/round60-visible-differences'
assert not target.exists() and not diffdir.exists(), '拒絕覆寫凍結收據或差分'


def item(path):
    p = root / path
    assert p.stat().st_uid == 1000 and p.stat().st_gid == 1000, path
    b = p.read_bytes()
    return {'path': str(path), 'bytes': len(b), 'sha256': hashlib.sha256(b).hexdigest()}


originals = []
files = ['initial.bin', 'input.bin', 'command.bin', 'mem.bin', 'screen.bin', 'screen.png', 'pal.bin']
for name in ['normal', 'late1', 'late255', 'clear', 'peace']:
    runs = []
    for side in ['a', 'b']:
        base = Path(f'workplace/oracle/intel59-r60-{name}-{side}')
        meta = (root/base/'meta.txt').read_text()
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in meta
        assert '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918' in meta
        inp, post = ((root/base/f).read_bytes() for f in ['input.bin', 'command.bin'])
        assert len(inp) == len(post) == 1048576
        assert len((root/base/'screen.bin').read_bytes()) == 307200
        fields = []
        for r in range(32):
            o = 0x5F2CC + r*126 + 0x2A
            fields.append({'region': r, 'input_hex': inp[o:o+2].hex(), 'output_hex': post[o:o+2].hex()})
        runs.append({'metadata': meta, 'flags': (root/base/'flags.txt').read_text().splitlines(),
                     'rng_input_94dc_hex': inp[0x5E00C:0x5E010].hex(), 'fields': fields,
                     'files': [item(base/f) for f in files+(['clear-input.bin', 'clear-output.bin'] if name == 'peace' else [])+['flags.txt', 'meta.txt', 'report.txt']]})
    for f in files+(['clear-input.bin', 'clear-output.bin'] if name == 'peace' else []):
        assert (root/f'workplace/oracle/intel59-r60-{name}-a'/f).read_bytes() == \
               (root/f'workplace/oracle/intel59-r60-{name}-b'/f).read_bytes(), (name, f)
    originals.append({'case': name, 'input_kind': 'normal-player' if name == 'normal' else 'explicit-poke-before-writer',
                      'grade': 'proven', 'paired_outputs_equal': True, 'runs': runs})

diagnostic = Path('workplace/oracle/intel59-r60-level1-a')
early = {'grade': 'proven-single-run', 'scope': '到達先清索敵值，玩家入口再清過期可見數；非雙跑驗收',
         'metadata': (root/diagnostic/'meta.txt').read_text(),
         'flags': (root/diagnostic/'flags.txt').read_text().splitlines(),
         'files': [item(diagnostic/f) for f in ['initial.bin', 'command.bin', 'mem.bin', 'screen.bin', 'report.txt', 'flags.txt', 'meta.txt']]}

game_runs = []
for side in ['a', 'b']:
    base = Path(f'workplace/round60-visible-game-final-{side}')
    game_runs.append([item(p.relative_to(root)) for p in sorted((root/base).iterdir())])
assert [x['sha256'] for x in game_runs[0]] == [x['sha256'] for x in game_runs[1]]
orig = root/'workplace/oracle/intel59-r60-normal-b'
game = root/'workplace/round60-visible-game-final-b'
want, got = ((p/'screen.bin').read_bytes() for p in [orig, game])
assert len(want) == len(got) == 307200 and want == got
assert game_palette(orig/'screen.png') == game_palette(game/'screen.png')
for r in range(32):
    saved = (game/'STELLAR.DT0').read_bytes()
    assert saved[r*126+0x2A:r*126+0x2C] == b'\0\0'

gui = []
for scale in [1, 2]:
    base = Path(f'workplace/round60-visible-gui-final-{scale}x')
    state = (root/(str(base)+'.txt')).read_text()
    assert 'state=StateStarList\n' in state and f'window={640*scale}x{480*scale}\n' in state
    assert 'cursor=460,125\n' in state
    assert (root/(str(base)+'.bin')).read_bytes() == want
    assert game_palette(root/(str(base)+'.png')) == game_palette(orig/'screen.png')
    gui.append({'scale': scale, 'input_kind': 'normal-player-real-X11',
                'files': [item(Path(str(base)+suffix)) for suffix in ['.bin', '.png', '.txt', '.log']]})

semantics = {
    '2703:0214': ('55', '0214函式入口，前一支213為retf'),
    '2703:0237': ('2688442b', '指定星域同盟可見數清零'),
    '2703:023D': ('2688442a', '指定星域帝國可見數清零'),
    '2703:0258': ('26807f04ff', '艦隊+04為FF則排除無效槽'),
    '2703:025F': ('268a4706', '讀艦隊所在星域以比較參數'),
    '2703:0280': ('268a4f0f', '讀編制組數，獨立偵察隊不在此範圍'),
    '2703:028B': ('268b07', '讀編制槽的艦艇池索引'),
    '2703:028E': ('b94e00', '艦艇池記錄步距78'),
    '2703:02A5': ('26807c1700', '索敵等級只與零比較'),
    '2703:02AC': ('bf0100', '任一非零組使布林符合，不按組數累加'),
    '2703:02BF': ('83c32a', '取得星域可見數兩格起點'),
    '2703:02C7': ('268a4c01', '依艦隊陣營選兩格'),
    '2703:02D1': ('26fe07', '符合的每支艦隊遞增一次'),
    '2703:02E4': ('cb', '0214函式返回'),
    '2703:0455': ('9a14020327', '043B迴圈逐星域呼叫0214'),
    '2703:01F0': ('9a14020327', '新遊戲初始化迴圈呼叫0214'),
    '04F8:000F': ('9a3e00f804', '玩家命令流程走共同003E入口'),
    '04F8:0038': ('9a3e00f804', '載入後流程亦走共同003E入口'),
    '04F8:0055': ('9a53080327', '共同入口先重算實數'),
    '04F8:005A': ('9a3b040327', '共同入口接著重算可見數'),
    '0374:041C': ('26f6471304', '讀星域戰鬥旗標'),
    '0374:0421': ('7439', '沒有戰鬥走0556分支'),
    '0374:0462': ('9a56057403', '無戰鬥星域呼叫0556'),
    '0374:057B': ('9a08067403', '先處理耗損與全滅'),
    '0374:058F': ('9af20a7403', '接著清理無戰鬥艦隊'),
    '0374:0599': ('26f6471310', '清理後才判占據'),
    '0374:0B0E': ('26807f04ff', '清理跳過無效槽'),
    '0374:0B15': ('268a4706', '清理只處理指定星域'),
    '0374:0B1E': ('26c6474804', '艦隊行動完整byte設4'),
    '0374:0B38': ('268a4f0f', '只處理編制組數，不包含偵察隊'),
    '0374:0B5D': ('26c6441700', '索敵等級清零，其他艦艇欄位保持'),
}
ida = []
reviewed = set()
for name in ['writer', 'seat', 'clear']:
    path = Path(f'workplace/ida/intel59-r60-{name}.json')
    data = json.loads((root/path).read_text())
    assert data['kernel_version'] == '9.4' and data['uid'] == 1000
    source = Path('workplace/ida')/Path(data['input']).name
    assert item(source)['sha256'] == data['input_sha256']
    for row in data['instructions']:
        loc = row['original_location']
        if loc in semantics:
            raw, meaning = semantics[loc]
            assert row['original_bytes_hex'] == raw, loc
            row.update(semantic=meaning, grade='proven', warning='', source='研究398；原始bytes與資料流審查／原版重播')
            reviewed.add(loc)
    ida.append({'export': item(path), 'input': item(source), 'data': data})
assert reviewed == set(semantics)

logs = ['round60-standard-final.log', 'round60-focused-final.log', 'round60-normal-final-a.log',
        'round60-normal-final-b.log', 'round60-gui-final.log']
for name in logs:
    text = (root/'workplace'/name).read_text()
    assert text and 'FAIL' not in text, name
assert (root/'workplace/round60-standard-final.log').read_text().count('github.com/wicanr2/logh3/') == 18
assert 'SKIP' not in (root/'workplace/round60-focused-final.log').read_text()
sources = ['internal/rules/region.go', 'internal/rules/region_visible_test.go',
           'internal/game/game.go', 'internal/game/regionvisible.go', 'internal/game/regionvisible_test.go',
           'internal/game/listtabedges_test.go', 'internal/game/listscreen.go', 'internal/game/rebellion.go',
           'tools/oracle.sh', 'tools/gui.sh', 'tools/ida.sh', 'tools/region_visible_fleets_oracle.sh',
           'tools/region_visible_fleets_gui.sh', 'tools/scripts/region_visible_fleets_ida.py',
           'tools/scripts/region_visible_fleets_receipt.py', 'tools/scripts/pngout.py']
assert item(Path('docs/internals/data/list-tab-edges.json'))['sha256'] == \
       '9a0e7a220fa8b1a91217c3f06b5c8daa0887d36af0c31951e15b70398693c2b8'
diffdir.mkdir()
write_png(diffdir/'normal.png', 640, 480, [0]*307200, [0]*768)
data = {
    'schema': 'logh3.region-visible-fleets.v1', 'date': '2026-10-03', 'round': 60, 'issue': 59,
    'remake_parent': '22af102d8049f19d54699dd24bf88fb35f988eaf', 'remake_tree_at_tests': 'dirty',
    'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', 'dosgolem_clean': True,
    'original_bundle': item(Path('org_game/银河英雄传说3SP.zip')),
    'parent_ram': item(Path('workplace/oracle/hold60-r58-base/mem.bin')),
    'parent_state': item(Path('workplace/oracle/hold60-r58-base/machine.state')),
    'originals': originals, 'early_diagnostic': early, 'ida': ida, 'reviewed_semantics': len(reviewed),
    'game_runs': game_runs, 'normal_gui': gui,
    'comparison': {'grade': 'same-state', 'input_kind': 'normal-player', 'rect': [0, 0, 640, 480],
                   'total': 307200, 'equal': 307200, 'different': 0, 'maximum_channel_difference': 0,
                   'difference_bbox': None, 'exclusions': [], 'regions': [[320,40,320,380], [336,80,272,320]],
                   'difference': item((diffdir/'normal.png').relative_to(root))},
    'sources': [item(Path(p)) for p in sources], 'logs': [item(Path('workplace')/name) for name in logs],
    'limits': ['晚注入只驗64格導出值，不宣稱正常局面或全圖通過',
               '雙人換邊驗資料接線，未宣稱雙人完整正常GUI', '索敵規則仍由087／#79驗收',
               '全RAM雙跑相同是原版可重現，不是原版與remake全RAM相同'],
}
with target.open('x', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
print(item(target.relative_to(root)))
