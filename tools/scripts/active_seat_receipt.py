"""第48輪#89活動控制；只重生本輪收據，保留46／47輪原始雜湊。"""
import hashlib
import json
from pathlib import Path
import struct

from pngout import game_palette, write_png


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def stats(a, b, pal, rect):
    x, y, w, h = rect
    offsets = [j * 640 + i for j in range(y, y + h) for i in range(x, x + w)]
    diff = [i for i in offsets if a[i] != b[i]]
    return {'rectangle_xywh': rect, 'compared_pixels': len(offsets),
            'same_pixels': len(offsets) - len(diff), 'different_index_pixels': len(diff),
            'different_rgb_pixels': sum(pal[a[i]*3:a[i]*3+3] != pal[b[i]*3:b[i]*3+3] for i in diff),
            'maximum_channel_difference': max((abs(pal[a[i]*3+k]-pal[b[i]*3+k]) for i in diff for k in range(3)), default=0),
            'difference_bbox_inclusive': [min(i % 640 for i in diff), min(i // 640 for i in diff),
                                          max(i % 640 for i in diff), max(i // 640 for i in diff)] if diff else None}


def ram(path):
    raw = Path(path).read_bytes()
    assert len(raw) == 1048576, path
    return raw


def state(raw):
    return {'active_faction': raw[0x5da4d], 'active_seat': raw[0x5da4e],
            'adjacent_unknown_0b': raw[0x5da4f], 'turn': struct.unpack_from('<H', raw, 0x5da48)[0],
            'strategy_seed_hex': f'{struct.unpack_from("<I", raw, 0x5e00c)[0]:08X}'}


originals = []
cases = {
    'clock': {'initial': (0, 0, 0), 'thinking': (1, 1, 0), 'ai-before': (1, 1, 0),
              'ai-after': (1, 1, 0), 'processing': (255, 1, 0), 'command': (0, 0, 1)},
    'alliance': {'loaded': (1, 0, 380), 'entry': (255, 1, 380)},
    'versus': {'loaded': (0, 0, 380), 'empire': (0, 0, 380),
               'handoff': (1, 1, 380), 'alliance': (1, 1, 380)},
}
for kind, stages in cases.items():
    a, b = [Path('workplace/oracle/normal89-r48-' + kind + '-' + n) for n in 'ab']
    names = [n + '.bin' for n in stages] + ['screen.bin', 'screen.png', 'pal.bin']
    assert all((a/n).read_bytes() == (b/n).read_bytes() for n in names), kind
    for p in (a, b):
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in (p/'meta.txt').read_text()
        argv = (p/'flags.txt').read_text().splitlines()
        assert argv[argv.index('-tick')+1] == '20000'
        assert ('-poke' in argv) == (kind == 'alliance')
        if kind == 'alliance':
            assert argv[argv.index('-poke')+1] == 'lin:5E00C@300000000=7E 16 DC D3'
        states = {}
        for name, expected in stages.items():
            raw = ram(p/(name+'.bin'))
            s = state(raw)
            assert (s['active_faction'], s['active_seat'], s['turn']) == expected
            assert s['adjacent_unknown_0b'] == 0
            states[name] = s
        originals.append({'path': str(p), 'command_argv': argv, 'states': states,
                          'classification': 'normal player continuation; own source snapshot',
                          'entry_filename_note': 'alliance entry.bin is processing, not battle' if kind == 'alliance' else None,
                          'hashes': {n: sha(p/n) for n in names+['meta.txt', 'flags.txt', 'report.txt']}})

aux = Path('workplace/explore/normal89-r48-alliance-long2-explore')
s = state(ram(aux/'mem.bin'))
assert (s['active_faction'], s['active_seat'], s['turn']) == (1, 0, 380)
assert ram(aux/'mem.bin')[0x5da46:0x5da48] == bytes([1, 26])
auxiliary = {'path': str(aux), 'grade': 'proven values at single auxiliary checkpoint; prefix not independently repeated',
             'comparison_level': 'auxiliary observation only; no full battle parity claim', 'state': s,
             'command_argv': (aux/'flags.txt').read_text().splitlines(),
             'hashes': {n: sha(aux/n) for n in ['mem.bin', 'screen.bin', 'screen.png', 'pal.bin',
                                               'final.state', 'meta.txt', 'flags.txt', 'report.txt']}}

ida_path = Path('workplace/ida/normal89-r48-active-seat.json')
ida = json.loads(ida_path.read_text())
assert sha(ida_path) == 'a7311ebbfe317274177c31a1a7df7eaf269d9c63b8ef55ddc7ababeb1c098f28'
assert ida['uid'] == 1000 and ida['kernel_version'] == '9.4'
assert len(ida['functions']) == 4 and len(ida['instructions']) == 102
assert sum(row['grade'] == 'proven' for row in ida['instructions']) == 7
initial = ram('workplace/oracle/normal89-r47-clock-a/initial.bin')
assert hashlib.sha256(bytes(256)+initial[0x4f80:0x56a0]).hexdigest() == ida['input_sha256']
for row in ida['instructions']:
    pos = int(row['runtime_linear_hex'], 16)
    raw = bytes.fromhex(row['original_bytes_hex'])
    assert initial[pos:pos+len(raw)] == raw


def double_players(prefix, count, command):
    a, b = [Path('workplace/'+prefix+'-'+n) for n in 'ab']
    names = sorted(p.relative_to(a).as_posix() for p in a.rglob('*') if p.is_file())
    assert names == sorted(p.relative_to(b).as_posix() for p in b.rglob('*') if p.is_file())
    assert len(names) == count
    assert all((a/n).read_bytes() == (b/n).read_bytes() for n in names)
    return [{'path': str(p), 'command_argv': command + ['-out', str(p)],
             'normal_input_metadata': {str(n.relative_to(p)): json.loads(n.read_text()) for n in p.rglob('probe.json')},
             'hashes': {n: sha(p/n) for n in names}} for p in (a, b)]


players = double_players('round48-player-final', 55, ['tools/gui.sh', 'xrun', 'go', 'run', 'tools/scripts/load_seat_player.go'])
normal = double_players('round48-normal-final', 19, ['tools/gui.sh', 'xrun', 'go', 'run', '-overlay',
                           'workplace/round44-normal-overlay.json', 'tools/scripts/battle_normal_probe.go'])
meta = json.loads(Path('workplace/round48-normal-final-a/probe.json').read_text())
assert meta['date'] == [799, 3, 1] and meta['ticks'] == 460
assert meta['strategy_seed_before'] == meta['strategy_seed_observed_before'] == 'D3DC167E'
assert meta['strategy_seed_after'] == 'B6E4B78A'
baseline = 'docs/internals/data/turn-counter-boundary.json'
assert sha(baseline) == '09d28a9e6d2b5c5cc49fd4f1ffa8d7c057170e4eb7f9e022baf6c9593d3a5a7b'
baseline46 = 'docs/internals/data/load-seat-selection.json'
assert sha(baseline46) == 'f8ba6dacc701418f1fe88c0c785fd871af8677916e4a54ed0ad1f9c19debf129'

rects = {'full': (0, 0, 640, 480), 'field': (0, 0, 448, 360),
         'right': (448, 0, 192, 480), 'bottom': (0, 360, 448, 120)}
comparisons, image_hashes, controls = {}, {}, []
for mode, stage, kind in [('empire', 'loaded', 'empire'), ('alliance', 'loaded', 'alliance'),
                          ('versus', 'loaded', 'versus'), ('versus', 'confirmed', 'vsconfirm'),
                          ('versus', 'handoff', 'vshandoff')]:
    p = Path(f'workplace/round48-player-final-a/{mode}/{stage}')
    old = Path(f'workplace/round46-player-a/{mode}/{stage}')
    op = Path('workplace/oracle/normal89-r46-'+kind+'-a')
    m = ram(op/'loaded-mem.bin')
    w, before = (p/'WORK.DT0').read_bytes(), (old/'WORK.DT0').read_bytes()
    assert w[9:11] == m[0x5da4d:0x5da4f]
    assert all(x == y for i, (x, y) in enumerate(zip(w, before)) if i not in (9, 10))
    for name in ('FLEET', 'SHIP', 'STELLAR', 'PLANET'):
        assert (p/(name+'.DT0')).read_bytes() == (old/(name+'.DT0')).read_bytes()
    controls.append({'mode': mode, 'stage': stage, 'before': list(before[9:11]), 'after': list(w[9:11]),
                     'original': list(m[0x5da4d:0x5da4f]), 'unknown_0b': w[11], 'other_saved_bytes_changed': 0})
    a, b = (op/'screen.bin').read_bytes(), (p/'screen.bin').read_bytes()
    assert len(a) == len(b) == 307200 and b == (old/'screen.bin').read_bytes()
    pal = bytes(game_palette(str(op/'screen.png')))
    base = f'workplace/round48-{mode}-{stage}'
    write_png(base+'.png', 640, 480, b, pal)
    write_png(base+'-diff.png', 640, 480, bytes(int(x != y) for x, y in zip(a, b)), bytes([0, 0, 0, 255, 0, 0]))
    comparisons[mode+'/'+stage] = {n: stats(a, b, pal, r) for n, r in rects.items()}
    assert comparisons[mode+'/'+stage]['full']['different_index_pixels'] == (2841 if kind == 'versus' else 0)
    image_hashes.update({base+n: sha(base+n) for n in ('.png', '-diff.png')})

op = Path('workplace/oracle/normal89-r47-battle-a')
entry = Path('workplace/round48-normal-final-a/entry')
mem = ram(op/'entry.bin')
tables = {}
for name, addr, length, expected in [('WORK', 0x5da44, 230, 32), ('FLEET', 0x60ba4, 3904, 120),
                                      ('SHIP', 0x668b2, 27456, 865), ('STELLAR', 0x5f2cc, 4032, 44)]:
    original = mem[addr:addr+length]
    before = Path('workplace/round47-normal-a/entry/'+name+'.DT0').read_bytes()
    after = (entry/(name+'.DT0')).read_bytes()
    assert before == after and len(after) == length
    assert sum(x != y for x, y in zip(original, after)) == expected
    tables[name] = {'bytes': length, 'remaining_different_bytes': expected, 'changed_since_round47': 0,
                    'original_sha256': hashlib.sha256(original).hexdigest(), 'remake_sha256': sha(entry/(name+'.DT0'))}
a, b = (op/'screen.bin').read_bytes(), (entry/'screen.bin').read_bytes()
assert b == Path('workplace/round47-normal-a/entry/screen.bin').read_bytes()
pal = bytes(game_palette(str(op/'screen.png')))
base = 'workplace/round48-first-battle'
write_png(base+'.png', 640, 480, b, pal)
write_png(base+'-diff.png', 640, 480, bytes(int(x != y) for x, y in zip(a, b)), bytes([0, 0, 0, 255, 0, 0]))
comparisons['empire/first-battle'] = {n: stats(a, b, pal, r) for n, r in rects.items()}
assert comparisons['empire/first-battle']['full']['different_index_pixels'] == 5853
image_hashes.update({base+n: sha(base+n) for n in ('.png', '-diff.png')})

gui = []
for name in ('empire', 'alliance', 'versus', 'versus-handoff', 'empire2x'):
    mode = 'empire' if name == 'empire2x' else 'versus' if name == 'versus-handoff' else name
    stage = 'handoff' if name == 'versus-handoff' else 'loaded'
    base = 'workplace/round48-seat-checked-gui-'+name
    assert Path(base+'.bin').read_bytes() == Path(f'workplace/round48-player-final-a/{mode}/{stage}/screen.bin').read_bytes()
    gui.append({'prefix': base, 'command_argv': ['tools/gui.sh', 'xrun', 'env', 'LOGH3_ACTIVE_GUI_PREFIX=round48-seat-checked-gui', 'bash', 'tools/active_seat_gui.sh'],
                'normal_keyboard_and_mouse': True, 'full_different_pixels_from_normal_Game': 0,
                'hashes': {base+n: sha(base+n) for n in ('.bin', '.png', '.txt', '.log')}})
for scale in (1, 2):
    base = f'workplace/round48-final-gui-{scale}x'
    assert Path(base+'.bin').read_bytes() == b
    gui.append({'prefix': base, 'command_argv': ['tools/gui.sh', 'xrun', 'env', 'LOGH3_TURN_GUI_PREFIX=round48-final-gui',
                                               'bash', 'tools/turn_counter_gui.sh'],
                'normal_keyboard_and_mouse': True, 'full_different_pixels_from_normal_Game': 0,
                'hashes': {base+n: sha(base+n) for n in ('.bin', '.png', '.txt', '.log')}})

logs = ['workplace/round48-active-limited.log', 'workplace/round48-standard-tests.log']
limited, standard = [Path(p).read_text() for p in logs]
assert '\nPASS\n' in limited and '--- FAIL:' not in limited and '--- SKIP:' not in limited
assert 'FAIL' not in standard and len([s for s in standard.splitlines() if s.startswith(('ok ', '? '))]) == 18
sources = ['internal/game/'+n for n in ('activeseat.go', 'activeseat_test.go', 'game.go', 'turnthink.go',
                                      'battle.go', 'loadfaction_test.go')]
sources += ['tools/active_seat_oracle.sh', 'tools/active_seat_gui.sh', 'tools/turn_counter_gui.sh',
            'tools/scripts/active_seat_ida.py', 'tools/scripts/active_seat_receipt.py',
            'tools/scripts/load_seat_player.go', 'tools/scripts/battle_normal_probe.go',
            'workplace/round44-normal-overlay.json', 'workplace/round44-normal-seed.go']
data = {'schema': 'logh3.active-seat.v1', 'grade': 'proven captured control boundaries; complete normal entry parity unknown',
        'classification': 'nearby-state whole frame comparisons; no masks, cropping or exclusions',
        'parent_head': '9bbc20cb2690057075052cafd0e7942719da70dc', 'remake_worktree': 'dirty; source hashes below',
        'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', 'dosgolem_worktree': 'clean',
        'program_sha256': sha('org_game/extracted/GIN3.COM'), 'bundle_sha256': sha('org_game/银河英雄传说3SP.zip'),
        'toolchain': json.loads(Path(baseline).read_text())['images'],
        'prior_receipts': {p: sha(p) for p in (baseline46, baseline)},
        'original_runs': originals, 'alliance_battle_auxiliary': auxiliary,
        'parent_snapshot_hashes': {'menu': sha('workplace/oracle/normal89-r46-menu-a/menu.state'),
            'fcmd_base': 'c55de581a6393e0e39e304dd0f191e9d1cfa8457b86b93f2a9e3368b5751ba26'},
        'ida_export': {'path': str(ida_path), 'sha256': sha(ida_path), 'data': ida},
        'normal_player_runs': players, 'normal_first_battle_runs': normal, 'first_battle_metadata': meta,
        'saved_controller_fields': controls, 'remaining_first_battle_tables': tables,
        'screen_comparisons': comparisons, 'comparison_image_hashes': image_hashes, 'actual_x11_gui': gui,
        'source_hashes': {p: sha(p) for p in sources}, 'test_logs': {p: sha(p) for p in logs},
        'limited_pass_events': limited.count('--- PASS:'), 'limited_skip_events': 0, 'standard_packages': 18,
        'failed_gui_input': {'prefix': 'workplace/round48-seat-final-gui-versus-handoff',
            'classification': 'StateCommandPhase passed but actual faction remained Empire; 7526 different pixels; validation script failure',
            'short_hold_script_sha256': sha('workplace/round48-active-seats-short-hold.sh'),
            'hashes': {'workplace/round48-seat-final-gui-versus-handoff'+n: sha('workplace/round48-seat-final-gui-versus-handoff'+n) for n in ('.bin', '.png', '.txt', '.log')}},
        'preserved_failed_logs': {p: sha(p) for p in ['workplace/round48-active-limited-fixture-failed.log',
                                                     'workplace/round48-standard-fixture-failed.log']},
        'limits': ['088完整規格仍DRAFT；雙人首戰／戰術內座位／退場／終局活動值尚未新增原版證據',
                   '同盟435M是處理畫面；650M首戰為單次輔助，不冒稱全流程雙跑',
                   '+0B語意unknown，沒有填入首都；helpers暫時陣營2不逐指令模擬',
                   '雙人未確認2841 pixels與STELLAR+7D22 bytes、首戰全表差異保留',
                   '只驗Linux X11鍵鼠；不宣稱macOS、觸控、人耳或完整發行驗收']}
target = Path('docs/internals/data/active-seat.json')
target.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print('原版3組雙跑；正常55與19檔雙跑；GUI7案；限定PASS', data['limited_pass_events'])
print('首戰5853 pixels及WORK32/FLEET120/SHIP865/STELLAR44保持；完整088未通過')
print('收據SHA-256', sha(target))
