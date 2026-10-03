"""#60：驗證原版保持補證，凍結來源／原始指令，不覆寫既有收據。"""
import hashlib
import json
import pathlib
import re

ROOT = pathlib.Path('/src')
OUT = ROOT / 'docs/internals/data/left-button-repeat.json'
assert not OUT.exists(), '凍結收據已存在，拒絕覆寫'
ORACLE = ROOT / 'workplace/oracle'
IDA = ROOT / 'workplace/ida'


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def file_record(path):
    raw = path.read_bytes()
    return {'path': str(path.relative_to(ROOT)), 'bytes': len(raw), 'sha256': sha(raw)}


def case(mode):
    name = 'hold60-r58-' + mode
    p = ORACLE / name
    meta = (p / 'meta.txt').read_text()
    assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in meta
    raw = (p / 'mem.bin').read_bytes()
    assert len(raw) == 1048576
    assert len((p / 'screen.bin').read_bytes()) == 307200
    flags = (p / 'flags.txt').read_text().splitlines()
    assert flags[flags.index('-tick') + 1] == '20000'
    assert 'lin:0:1048576:' in flags[flags.index('-dump-mem-at') + 1]
    writes = []
    for line in (p / 'writes.txt').read_text().splitlines():
        parts = line.split()
        assert len(parts) == 5
        step, addr, old, new, ip = parts
        writes.append({'step': int(step), 'linear_hex': '0x' + addr,
                       'before_hex': old, 'after_hex': new,
                       'executed_next_cs_ip': ip})
    return {
        'name': name, 'flags': flags,
        'artifacts': [file_record(p / f) for f in
                      ['meta.txt', 'flags.txt', 'report.txt', 'mem.bin',
                       'machine.state', 'screen.bin', 'screen.png', 'pal.bin', 'writes.txt']],
        'values': {'bottom_admiral_offset': raw[0x5e109],
                   'star_offset': raw[0x5e21c], 'tax': raw[0x5db04],
                   'build_percent': raw[0x5dac2]},
        'writes': writes,
    }


modes = ['base', 'tab', 'star', 'tax']
for stem in ['tab-300000', 'star-300000', 'tax-3000000', 'pct-3000000']:
    modes += [stem + '-a', stem + '-b']
modes += ['tab-300000-move', 'star-300000-move', 'pct-3000000-move', 'pct-3000000-clock']
cases = {mode: case(mode) for mode in modes}
pairs = []
for stem in ['tab-300000', 'star-300000', 'tax-3000000', 'pct-3000000']:
    a, b = ORACLE / ('hold60-r58-' + stem + '-a'), ORACLE / ('hold60-r58-' + stem + '-b')
    comparisons = {}
    for name in ['mem.bin', 'screen.bin', 'screen.png', 'pal.bin']:
        x, y = (a / name).read_bytes(), (b / name).read_bytes()
        assert x == y, (stem, name)
        comparisons[name] = {'bytes': len(x), 'different_bytes': 0, 'sha256': sha(x)}
    pairs.append({'case': stem, 'comparison': comparisons})
assert cases['tab-300000-b']['values']['bottom_admiral_offset'] == 5
assert cases['star-300000-b']['values']['star_offset'] == 2
assert cases['tax-3000000-b']['values']['tax'] == 4
assert cases['pct-3000000-b']['values']['build_percent'] == 27
assert cases['tab-300000-move']['values']['bottom_admiral_offset'] == 5
assert cases['star-300000-move']['values']['star_offset'] == 2
assert cases['pct-3000000-move']['values']['build_percent'] == 26
for name in ['mem.bin', 'screen.bin', 'screen.png', 'pal.bin']:
    assert (ORACLE / 'hold60-r58-pct-3000000-a' / name).read_bytes() == (
        ORACLE / 'hold60-r58-pct-3000000-clock' / name).read_bytes()

exports = []
locations = {}
for name, parent in [('3765', 'data67-r57-parent'), ('0C0A-full', 'data67-r57-parent'),
                     ('0D05', 'data67-r57-parent'), ('0BCE', 'hold60-r58-tax'),
                     ('090A-tax', 'hold60-r58-tax'), ('37A9-tax', 'hold60-r58-tax')]:
    path = IDA / ('hold60-r58-' + name + '.json')
    d = json.loads(path.read_text())
    assert d['kernel_version'] == '9.4' and d['uid'] == 1000
    raw = (ORACLE / parent / 'mem.bin').read_bytes()
    com = IDA / pathlib.Path(d['input']).name
    assert sha(com.read_bytes()) == d['input_sha256']
    segment = int(d['runtime_segment'], 16)
    assert com.read_bytes()[256:] == raw[segment * 16:segment * 16 + len(com.read_bytes()) - 256]
    for row in d['instructions']:
        at = int(row['runtime_linear_hex'], 16)
        data = bytes.fromhex(row['original_bytes_hex'])
        assert raw[at:at + len(data)] == data
        locations[row['original_location']] = row
    exports.append({'artifact': file_record(path), 'source_ram': file_record(ORACLE / parent / 'mem.bin'),
                    'input_com': file_record(com), 'export': d})

semantics = [
    ('3765:0266', '左鍵參數0選擇按下事件碼1', 'c746000100'),
    ('3765:026B', '左鍵參數0選擇放開事件碼4', 'c746020400'),
    ('3765:02B3', '找到按下或放開事件後回傳值設0', 'bf0000'),
    ('0D05:024F', '星系列表位移每次加1', 'fe06ec96'),
    ('0D05:0248', '星系列表上界16', '803eec9610'),
    ('0D05:023D', '星系捲動迴圈先檢查左鍵事件', '9a55026537'),
    ('0C0A:0270', '提督列表位移每次加1', 'fe06a196'),
    ('0C0A:024F', '提督捲動迴圈先檢查左鍵事件', '9a55026537'),
    ('0BCE:02C2', '右下提督列表位移每次加1', 'fe06d995'),
    ('0BCE:02E7', '右下列表更新／重畫後檢查左鍵事件', '9a55026537'),
    ('090A:0394', '稅率檔位每次加1', 'fe87d48f'),
    ('090A:04DB', '造船百分比每次加1', 'fe87928f'),
    ('090A:0317', '稅率重畫後查熱區保持helper', '9a7004a937'),
    ('090A:047D', '百分比重畫後查熱區保持helper', '9a7004a937'),
    ('37A9:047C', '保持helper等待DS:8EFE歸零', '833efe8e00'),
    ('37A9:0485', '下一次等待倒數設8', 'c706fe8e0800'),
    ('37A9:049C', '事件碼4辨識左鍵放開', '83f804'),
    ('37A9:0575', '游標離開熱區時回傳狀態2', 'c6460c02'),
]
annotations = []
for location, meaning, expected in semantics:
    row = locations[location]
    assert row['original_bytes_hex'] == expected, (location, row['original_bytes_hex'], expected)
    annotations.append({'original_location': location, 'original_bytes_hex': expected,
                        'semantic': meaning, 'grade': 'proven',
                        'source': '本收據IDA原始指令與对应RAM；動態更新另見cases'})

data = {
    'schema': 'logh3.left-button-repeat.v1', 'date': '2026-10-03', 'issue': 60,
    'parent_commit': '2c44b828097d9680e689419089809e6cdde36efc',
    'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', 'dosgolem_dirty': False,
    'comparison_level': '原版保持補證；沒有remake same-state聲明',
    'original_executable_sha256': '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918',
    'coordinate_mapping': {'canvas': [640, 480], 'mouse_y_to_screen_y': 40},
    'cases': cases, 'repeat_pairs': pairs, 'ida': exports, 'semantics': annotations,
    'limitations': [
        'tab-a早期監看5E106無更新，已以RAM定位5E109並於tab-b記錄完整更新；a/b全RAM與全圖仍一致',
        '六份IDA掃描不是正式函式邊界，未列附加語意的指令保持unknown',
        '清單速度受重畫成本影響，不把平均指令數當成精確硬體時間',
        '稅務helper倒數8、PIT2048／十IRQ為一倒數沿研究331，屬hardware-spec approximation',
        '游標移出實驗各一跑；四個保持a/b成對雙跑，不擴大決定性聲明',
        '艦隊編成、戰場資料窗及其餘分頁尚未完成保持補驗；089仍DRAFT，正式輸入未改',
    ],
    'sources': [file_record(ROOT / p) for p in
                ['tools/left_hold_oracle.sh', 'tools/scripts/left_hold_ida.py',
                 'tools/scripts/left_hold_receipt.py', 'tools/scripts/left_hold_prototype.html']],
}
OUT.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')
print(json.dumps({'path': str(OUT.relative_to(ROOT)), 'sha256': sha(OUT.read_bytes()),
                  'cases': len(cases), 'instructions': sum(len(x['export']['instructions']) for x in exports),
                  'proven': len(annotations), 'pairs': len(pairs)}, ensure_ascii=False))
