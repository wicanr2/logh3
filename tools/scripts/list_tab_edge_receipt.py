"""#59：凍結原版雙跑／正式全圖／GUI／存檔與指令，拒絕覆寫。"""
import hashlib
import json
from pathlib import Path
from pngout import game_palette, write_png

root = Path('/src')
target = root / 'docs/internals/data/list-tab-edges.json'
diffdir = root / 'workplace/round59-list-tab-differences'
assert not target.exists() and not diffdir.exists(), '拒絕覆寫凍結收據或差分'
names = ['empire-counts', 'empire-hidden', 'ally-visible-counts', 'ally-visible-hidden',
         'star-battle', 'morale49', 'morale50', 'battle', 'occupied']


def item(path):
    p = root / path
    b = p.read_bytes()
    assert p.stat().st_uid == 1000 and p.stat().st_gid == 1000
    return {'path': str(path), 'bytes': len(b), 'sha256': hashlib.sha256(b).hexdigest()}


def original(name):
    base = Path('workplace/oracle/list59-r59-' + name)
    assert (root/base/'mem.bin').stat().st_size == 1048576
    assert (root/base/'screen.bin').stat().st_size == 307200
    meta = (root/base/'meta.txt').read_text()
    assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in meta
    assert '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918' in meta
    flags = (root/base/'flags.txt').read_text().splitlines()
    return {'name': name, 'flags': flags, 'metadata': meta,
            'rng_94dc_hex': (root/base/'mem.bin').read_bytes()[0x5E00C:0x5E010].hex(),
            'files': [item(base/f) for f in ['mem.bin', 'screen.bin', 'screen.png',
                       'pal.bin', 'machine.state', 'flags.txt', 'meta.txt', 'report.txt']]}


orig = []
for name in names + ['ally-normal', 'ally-counts', 'ally-hidden']:
    a, b = original(name+'-a'), original(name+'-b')
    for f in ['mem.bin', 'screen.bin', 'screen.png', 'pal.bin']:
        assert (root/f'workplace/oracle/list59-r59-{name}-a'/f).read_bytes() == \
               (root/f'workplace/oracle/list59-r59-{name}-b'/f).read_bytes()
    orig.append({'case': name, 'grade': 'diagnostic-blind' if name in ['ally-counts', 'ally-hidden'] else 'proven',
                 'full_ram_and_screen_and_png_and_palette_equal': True, 'runs': [a, b]})

diffdir.mkdir()
comparisons = []
for name in names + ['normal-ally']:
    a, b = (root/f'workplace/round59-game-final-{v}'/name for v in ['a', 'b'])
    assert sorted(p.name for p in a.iterdir()) == sorted(p.name for p in b.iterdir())
    for p in a.iterdir():
        assert p.read_bytes() == (b/p.name).read_bytes(), p.name
    oname = 'ally-normal' if name == 'normal-ally' else name
    o = root/f'workplace/oracle/list59-r59-{oname}-b'
    want, got = (o/'screen.bin').read_bytes(), (b/'screen.bin').read_bytes()
    assert len(want) == len(got) == 307200
    opal, gpal = game_palette(o/'screen.png'), game_palette(b/'screen.png')
    assert opal == gpal
    bad = [i for i, (x, y) in enumerate(zip(want, got)) if x != y]
    assert not bad, name
    write_png(diffdir/(name+'.png'), 640, 480, [int(x != y) for x, y in zip(want, got)],
              [0, 0, 0, 255, 0, 0] + [0]*762)
    comparisons.append({'case': name, 'grade': 'same-state',
                        'input_kind': 'normal-player' if name == 'normal-ally' else 'explicit-poke-equivalent',
                        'original': str(o.relative_to(root)),
                        'remake_runs': [[item(p.relative_to(root)) for p in sorted(d.iterdir())] for d in [a, b]],
                        'full_frame': {'rect': [0, 0, 640, 480], 'total': 307200, 'equal': 307200,
                                       'different': 0, 'max_channel_difference': 0, 'bounding_box': None},
                        'list_panel': {'rect': [320, 40, 320, 380], 'total': 121600, 'different': 0},
                        'active_rows': {'rect': [336, 80, 272, 320], 'total': 87040, 'different': 0},
                        'allowed_regions': [], 'diff': item((diffdir/(name+'.png')).relative_to(root))})

gui = []
for scale in [1, 2]:
    base = Path(f'workplace/round59-list-tab-gui-{scale}x')
    assert (root/str(base.with_suffix('.bin'))).read_bytes() == \
           (root/'workplace/round59-game-final-b/normal-ally/screen.bin').read_bytes()
    state = (root/base.with_suffix('.txt')).read_text()
    assert 'state=StateStarList\n' in state and f'window={640*scale}x{480*scale}\n' in state
    gui.append({'scale': scale, 'state': state, 'full_frame_different': 0,
                'files': [item(base.with_suffix(ext)) for ext in ['.bin', '.png', '.txt', '.log']]})

annotations = {
    '0D05:057F': '觀看座位分流，座位0走+28/+2B，非零座位走+2A/+29',
    '0D05:0589': '座位0帝國數讀星域+28', '0D05:0590': '座位0同盟數讀星域+2B',
    '0D05:059C': '座位1帝國數讀星域+2A', '0D05:05A3': '座位1同盟數讀星域+29',
    '0D05:05B5': '左數FF走占位字串', '0D05:0612': '右數FF走占位字串',
    '0D05:05E3': '星域+13 bit2控制冒號色號',
    '0C7F:0648': '士氣與50比較，低於50才換色2',
    '0C7F:0652': '低士氣分支寫顏色屬性82',
    '0C7F:06DB': '先判星域+13 bit2，印戰鬥',
    '0C7F:06F2': '無bit2再判bit4，印占領',
    '2703:0872': '星域+29清零', '2703:0876': '星域+28清零',
    '2703:08B9': '取星域+28地址', '2703:08C1': '讀艦隊+01陣營',
    '2703:08C9': '以陣營偏移選+28或+29', '2703:08CB': '遞增選中的星域實數格',
}
ida = []
for name in ['star', 'fleet', 'recount']:
    path = Path(f'workplace/ida/list59-r59-{name}.json')
    d = json.loads((root/path).read_text())
    assert d['kernel_version'] == '9.4' and d['uid'] == 1000
    inp = Path('workplace/ida')/Path(d['input']).name
    assert item(inp)['sha256'] == d['input_sha256']
    for row in d['instructions']:
        if row['original_location'] in annotations:
            row.update(semantic=annotations[row['original_location']], grade='proven',
                       warning=None, source='研究397；原指令與第59輪完整雙跑')
    ida.append({'export': item(path), 'input': item(inp), 'data': d})
assert sum(row['grade'] == 'proven' for d in ida for row in d['data']['instructions']) == len(annotations)

sources = ['internal/game/game.go', 'internal/game/listscreen.go', 'internal/game/listtabedges_test.go',
           'internal/rules/region.go', 'internal/rules/region_list_test.go', 'cmd/gin3/main.go',
           'tools/list_tab_edge_oracle.sh', 'tools/list_tab_edge_gui.sh',
           'tools/scripts/list_tab_edge_ida.py', 'tools/scripts/list_tab_edge_receipt.py', 'tools/scripts/pngout.py']
standard = (root/'workplace/round59-standard.log').read_text()
assert '\nFAIL' not in standard and len([l for l in standard.splitlines() if l.startswith(('ok  ', '?   '))]) == 18
for run in ['a', 'b']:
    log = (root/f'workplace/round59-game-final-{run}.log').read_text()
    assert 'FAIL' not in log and 'SKIP' not in log and log.count('--- PASS:') == 11
data = {
    'schema': 'logh3.list-tab-edges.v1', 'date': '2026-10-03', 'issue': 59,
    'parent_git_head': '2d3cac5071a3519135b718bf8a48d0a5e8cf9b9f', 'working_tree': 'dirty; sources逐檔固定',
    'dosgolem': {'head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', 'dirty': False},
    'images': {'gui': 'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
               'go': 'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac',
               'ida': 'sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780'},
    'parents': [original('ally-base'),
                {'path': 'workplace/oracle/hold60-r58-base',
                 'files': [item(Path('workplace/oracle/hold60-r58-base')/f) for f in
                           ['mem.bin', 'machine.state', 'flags.txt', 'meta.txt']]}],
    'original_pairs': orig, 'ida': ida, 'comparisons': comparisons, 'gui': gui,
    'sources': [item(p) for p in sources],
    'logs': [item('workplace/'+n) for n in ['round59-before.log', 'round59-after-first.log',
             'round59-game-final-a.log', 'round59-game-final-b.log', 'round59-standard.log', 'round59-gui.log']],
    'limits': ['四格寫入生命週期仍未驗，不關#59', '不對原版與remake完整RAM相等作聲明',
               '九個局面為明示注入，正常玩家路徑只限第十張同盟星系清單',
               '兩組同盟星域0盲探針不算分頁驗收', '不改原版格式，不發明情報揭露機制'],
}
with target.open('x', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
    f.write('\n')
print(item(target.relative_to(root)))
