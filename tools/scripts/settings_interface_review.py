"""在tools/gui.sh xrun容器內核本輪設定來源；--freeze一次、其後--check。"""
from pathlib import Path
import ast
import hashlib
import json
import sys

target = Path('docs/internals/data/settings-interface.json')


def sha(b):
    return hashlib.sha256(b).hexdigest()


native = Path('workplace/oracle')
for a, b in [('bug-b', 'bug-c'), ('midi-b', 'midi-c'),
             ('midi-none-a', 'midi-none-b'), ('midi-330-a', 'midi-330-b'),
             ('midi-cancel-a', 'midi-cancel-b')]:
    for f in ['screen.bin', 'pal.bin']:
        assert (native/('settings66-r98-'+a)/f).read_bytes() == (native/('settings66-r98-'+b)/f).read_bytes(), (a, b, f)
    for name in [a, b]:
        root = native/('settings66-r98-'+name)
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in (root/'meta.txt').read_text()
        assert '沒實作的服務（0 種）' in (root/'report.txt').read_text()

pa, pb = [Path('workplace/settings66-r98-player-'+r) for r in ['a', 'b']]
names = sorted(p.name for p in pa.iterdir())
assert len(names) == 7 and names == sorted(p.name for p in pb.iterdir())
assert all((pa/n).read_bytes() == (pb/n).read_bytes() for n in names)

mem = Path('workplace/settings66-r98-base-v2.mem').read_bytes()
bugmem = Path('workplace/settings66-r98-bug-v2.mem').read_bytes()
assert len(mem) == len(bugmem) == 1048576
assert sha(mem) == '24f40493c0e464034256f185fa5131290667378d19ab3b179e1acd6544b5aba5'
ds, work = 0x54b30, 0x54b30+0x8f14
assert mem[ds+0x210] == mem[ds+0x4cc] == 0
assert mem[work:work+230] == bugmem[work:work+230]
assert mem[ds+0x564:ds+0x572].hex() == 'ffff000320033003320334033603'

ini_results = {}
for name, want in [('none', '000001ffffff'), ('330', '000001ff3003')]:
    a, b = [Path('workplace/settings66-r98-midi-'+name+'-'+r+'.mem').read_bytes() for r in ['a', 'b']]
    assert len(a) == 1048576 and a == b
    offset = int.from_bytes(a[ds+0xf138:ds+0xf13a], 'little')
    segment = int.from_bytes(a[ds+0xf13a:ds+0xf13c], 'little')
    ini = a[segment*16+offset:segment*16+offset+6].hex()
    assert ini == want and a[ds+0x934e] == 2 and a[ds+0x8f33] == 1
    ini_results[name] = dict(ini=ini, device_mask=2, device=1, memory_equal=True)

cancel_a, cancel_b = [Path('workplace/settings66-r98-midi-cancel-'+r+'.mem').read_bytes() for r in ['a', 'b']]
assert cancel_a == cancel_b and cancel_a[ds+0x934e] == 2 and cancel_a[ds+0x8f33] == 1
menu_mem = Path('workplace/settings66-r98-midi-v2.mem').read_bytes()
assert cancel_a[work:work+230] == menu_mem[work:work+230]
for r in ['a', 'b']:
    report = (native/('settings66-r98-midi-cancel-'+r)/'report.txt').read_text()
    assert '開過的檔（0）：（無）' in report and '042F:0971 執行時的暫存器（前 0 次）' in report
ini_results['cancel'] = dict(file_opened=False, write_entry_hits=0, work_unchanged_bytes=230,
                             device_mask=2, device=1, memory_equal=True,
                             ini_pointer_lifetime='not interpreted; selector cancel does not load INI buffer')

spec = json.loads(Path('docs/internals/data/screen-settings.json').read_text())
assert next(h for h in spec['hotspots'] if h['action'] == 'bug')['disabled'] is True
comparison = json.loads(Path('workplace/settings66-r98-comparison/comparison.json').read_text())
assert len(comparison) == 2
for c in comparison:
    assert c['index_different'] == 0 and len(c['regions']) == 3
    assert all(r['different'] == 0 and r['max_channel_difference'] == 0 and r['bbox'] is None for r in c['regions'])
    n = 'base-c' if c['name'] == 'settings' else 'bug-b'
    assert (pa/(c['name']+'-rgb16.bin')).read_bytes() == (native/('settings66-r98-'+n)/'pal.bin').read_bytes()[:48]

guiroot = Path('workplace/settings66-r98-gui-a')
gui = json.loads((guiroot/'receipt.json').read_text())
assert gui['events'][0]['tps'] == 60 and gui['events'][0]['injection'] is False
assert gui['events'][-1]['state']['state'] == 'StateSettings' and gui['events'][-1]['different'] == 0
gui_files = list(guiroot.glob('*-x11-comparison.json'))
assert len(gui_files) == 2
for p in gui_files:
    r = json.loads(p.read_text())
    assert r['pixels'] == 307200 and r['different'] == 0 and r['max_channel_difference'] == 0 and r['bbox'] is None

ida = json.loads(Path('workplace/ida/settings66-r98-042f-ida-v3.json').read_text())
raw = Path('workplace/ida/settings66-r98-042f.com').read_bytes()
assert ida['ida_version'] == '9.4' and ida['uid'] == 1000 and len(ida['instructions']) == 269
assert sha(raw) == ida['input']['sha256'] == 'e087ec02a6aa9b4a3483dbf7f2754ee3319c4232001b30d92da6e3505f100537'
annotations = {
    '01B1': ('a01002', '讀DS:0210的原版啟用值'),
    '01B4': ('a2cc04', '把啟用值寫入DS:04CC熱區+4'),
    '0923': ('9a04006736', '呼叫3667:0004建立選單'),
    '0940': ('9af2006736', '呼叫3667:00F2取得選擇列號'),
    '094F': ('9acb006736', '呼叫3667:00CB關閉選單'),
    '0957': ('807e24ff', '列號FF為取消'),
    '0967': ('d1e3', '列號乘2以索引word埠表'),
    '0969': ('8b876405', '從DS:0564查埠號'),
    '096D': ('c41e38f1', '讀DS:F138的INI緩衝區遠指標'),
    '0971': ('26894704', '寫入INI+4；無／330H雙跑核RAM值'),
    '0984': ('f6064e9301', '測934E的MIDI可用位元'),
    '0990': ('9a09022f04', '重新呼叫設定繪製042F:0209'),
}
rows = []
for original in ida['instructions']:
    r = dict(original)
    offset = int(r['file_offset'], 16)
    assert raw[offset:offset+len(bytes.fromhex(r['bytes']))].hex() == r['bytes']
    assert not r['text'].startswith('db ')
    if r['file_offset'] in annotations:
        signature, semantic = annotations[r['file_offset']]
        assert r['bytes'] == signature
        r.update(semantic=semantic, semantic_level='proven', semantic_source='docs/re/432-settings-interface.md; 固定原始bytes與本輪正常收據', warning='')
    rows.append(r)
assert sum(r['semantic_level'] == 'proven' for r in rows) == len(annotations)

standard = Path('workplace/round98-standard-gui.log').read_text()
assert standard.count('\nok  ')+int(standard.startswith('ok  ')) == 13
assert standard.count('[no test files]') == 5 and 'FAIL' not in standard
for name in ['settings_interface_ida.py', 'settings_interface_gui.py', 'settings_interface_review.py']:
    ast.parse(Path('tools/scripts', name).read_text())

if sys.argv[1] == '--freeze':
    assert not target.exists()
    paths = {
        'docs/goal/2026-10-05-settings-interface-round-98.md', 'docs/re/432-settings-interface.md',
        'docs/spec/113-settings-bug-unavailable.md', 'docs/spec/114-midi-interface.md',
        'docs/playtest/087-settings-interface.md', 'docs/internals/23-hotspot-lists.md',
        'docs/internals/data/screen-settings.json', 'tools/scripts/settings_hotspots.py',
        'tools/scripts/settings_interface_ida.py', 'tools/scripts/settings_interface_gui.py',
        'tools/scripts/settings_interface_compare.go', 'tools/scripts/settings_interface_review.py',
        'tools/scripts/midi_interface_prototype.html', 'tools/scripts/fleet_player_gui_capture.go',
        'tools/oracle.sh', 'tools/go.sh', 'tools/gui.sh', 'tools/py.sh', 'tools/ida.sh',
        'workplace/round98-standard-gui.log', 'docs/internals/data/title-player.json',
        'docs/internals/data/title-asset-receipts.json', 'docs/internals/data/battle-lifecycle-player.json',
    }
    for folder in ['internal/game', 'internal/bootstrap', 'internal/audio', 'internal/gfx', 'internal/screen',
                   'internal/rules', 'internal/asset', 'internal/font', 'internal/ai', 'internal/ui', 'cmd/gin3']:
        paths.update(str(p) for p in Path(folder).glob('*.go'))
    for root in Path('workplace').glob('settings66-r98-*'):
        if root.is_file():
            paths.add(str(root))
        else:
            paths.update(str(p) for p in root.rglob('*') if p.is_file())
    for root in native.glob('settings66-r98-*'):
        paths.update(str(p) for p in root.rglob('*') if p.is_file())
    # 排除可能含授權環境內容的IDA log；只保存可審閱的原始輸入及JSON。
    paths.update(str(p) for p in Path('workplace/ida').glob('settings66-r98-*') if p.suffix in ['.json', '.com'])
    sources = {}
    for name in sorted(paths):
        p = Path(name)
        b = p.read_bytes()
        assert p.stat().st_uid == p.stat().st_gid == 1000, name
        sources[name] = dict(bytes=len(b), sha256=sha(b))
    originals = {}
    for name in ['GIN3.COM', 'MUSIC.INI', 'MUSIC.ORG', 'GRAPH.IMG']:
        b = Path('/orig', name).read_bytes()
        originals[name] = dict(bytes=len(b), sha256=sha(b))
    result = dict(schema='logh3.settings-interface-receipts.v1', date='2026-10-05',
                  parent_commit='62fba4601656c53d2f6c3a88320b1fbb1c3e65ea', remake_worktree_dirty=True,
                  dosgolem_head='f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', dosgolem_dirty=False,
                  go_image='golang:1.24-bookworm', gui_image='logh3-ebiten:2', ida_image='ida-pro-9.4-idapython:locked-v1',
                  spec113='CONFORMED; normal shipped BUG unavailable', spec114='DRAFT; software MIDI mapping pending',
                  original_work_unchanged_bytes=230, player_repeated_files=7, comparison=comparison,
                  midi_original=ini_results, gui_x11_captures=2, gui_x11_pixels=614400,
                  standard_tests=dict(packages=18, passed=13, no_tests=5),
                  ida=dict(version='9.4', file_offset_zero_ea=ida['file_offset_zero_ea'], instructions=269, proven=len(annotations)),
                  instructions=rows, original_inputs=originals, sources=sources)
    with target.open('x') as f:
        f.write(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
elif sys.argv[1] != '--check':
    raise SystemExit('用法：--freeze／--check')

receipt = json.loads(target.read_text())
for name, expected in receipt['sources'].items():
    b = Path(name).read_bytes()
    assert len(b) == expected['bytes'] and sha(b) == expected['sha256'], name
for name, expected in receipt['original_inputs'].items():
    b = Path('/orig', name).read_bytes()
    assert len(b) == expected['bytes'] and sha(b) == expected['sha256'], name
print('正常BUG!全圖、7檔雙跑、兩張X11、269原指令及', len(receipt['sources']), '來源符合；索引SHA-256', sha(target.read_bytes()))
