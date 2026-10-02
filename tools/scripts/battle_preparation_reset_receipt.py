"""研究370：核對初始化、讀檔、戰場入口雙跑；保存解出的生命週期事實。"""
import hashlib
import json
import re
from pathlib import Path

root = Path('/work')
out = root / 'docs/internals/data/battle-preparation-reset.json'
assert out.parent.stat().st_uid == 1000
assert not out.exists() or out.stat().st_uid == 1000
sha = lambda b: hashlib.sha256(b).hexdigest()
rows = []
for mode in ('newgame', 'loadexisting', 'battleentry'):
    runs = []
    for suffix in ('a', 'b'):
        p = root / 'workplace/oracle' / f'battle79-life-{mode}-{suffix}'
        files = {n: (p/n).read_bytes() for n in
                 ('mem-in.bin', 'mem-out.bin', 'screen.bin', 'pal.bin', 'flags.txt', 'meta.txt', 'report.txt')}
        assert len(files['mem-in.bin']) == len(files['mem-out.bin']) == 1048576
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in files['meta.txt'].decode()
        assert '前面丟掉 0 筆' in files['report.txt'].decode()
        before, after = files['mem-in.bin'], files['mem-out.bin']
        for addr, expected in ((0x5EF5E, 'ec008b66'), (0x5EF70, '3a018b66'), (0x61AE4, '02008b66')):
            assert before[addr:addr+4].hex() == after[addr:addr+4].hex() == expected
        assert before[0x5E00C:0x5E010].hex() == '78563412'
        report = files['report.txt'].decode()
        if mode == 'newgame':
            assert '#58339340  5EF5E: 09→00  ip=0349:01EA' in report
            assert 'ADMIRAL.DTX' in report and 'BGM_0F.MDS' in report
        if mode == 'loadexisting':
            assert all(n in report for n in ('ADMIRAL.DTA', 'FLEET.DTA', 'SHIP.DTA', 'WORK.DTA', 'BGM_3F.MDS'))
        if mode == 'battleentry':
            assert '12AA:000E 執行時的暫存器（前 1 次）' in report
            assert '#430919835 AX=' in report
            assert 'ANI_PNL0.A98' in report
        writes = re.findall(r'^  #(\d+)  (5EF[0-9A-F]+): ([0-9A-F]{2})→([0-9A-F]{2})  ip=([^\n]+)', report, re.M)
        start = 420000001 if mode == 'loadexisting' else 100000000
        # 監看只記值改變；不把「未記錄」說成完全沒有同值寫入。
        assert not [w for w in writes if int(w[0]) > start]
        runs.append({'path': str(p.relative_to(root)), 'sha256': {n: sha(b) for n, b in files.items()},
                     'flags': files['flags.txt'].decode().splitlines(),
                     'seed_before_hex': before[0x5E00C:0x5E010].hex(),
                     'seed_after_hex': after[0x5E00C:0x5E010].hex()})
    assert all(runs[0]['sha256'][n] == runs[1]['sha256'][n] for n in
               ('mem-in.bin', 'mem-out.bin', 'screen.bin', 'pal.bin'))
    rows.append({'mode': mode, 'grade': 'proven', 'runs': runs,
                 'supply_source_hex': 'ec008b66', 'repair_source_hex': '3a018b66',
                 'pool_base_hex': '02008b66', 'changed_writes_after_injection': 0})
checkpoint = Path('/ref/workplace/fcmd/fcmd-base.state').read_bytes()
resolutions = []
for old in ('docs/re/367-battle-preparation.md', 'docs/re/369-battle-preparation-lifecycle.md',
            'docs/re/143-supply-panel.md', 'docs/spec/062-battle-turn-eight-steps.md',
            'docs/spec/079-battle-preparation.md', 'docs/internals/65-battle-preparation.md'):
    marker = '2026-10-02／研究370續證'
    assert marker in (root / old).read_text(), '缺少受影響舊文件的續證／勘誤入口：' + old
    resolutions.append({'old_document': old, 'required_marker': marker,
                        'program_sha256': '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918',
                        'original_locations': ['1930:019D', '1AAB:006D', '1629:023B'],
                        'evidence': 'docs/re/370-battle-preparation-reset.md',
                        'scope': '指定來源生命週期與正式第一階段，不宣稱八步一致。'})
value = {'schema': 'logh3.battle-preparation-reset.v1', 'issue': 79, 'date': '2026-10-02', 'resolutions': resolutions,
         'golem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', 'golem_dirty': False,
         'program': 'GIN3.COM', 'program_sha256': '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918',
         'bundle_sha256': 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
         'remake_parent': '58e02df905cec946eb7be40f53bfe0ceed54a6cf',
         'tick': 20000, 'seed': '0x12345678', 'seed_method': '執行前以lin:5E00C固定；入口流程可能正常消費亂數，不以首尾相同作此項門檻。',
         'go_image': 'golang:1.24-bookworm',
         'go_image_id': 'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac',
         'checkpoint': 'dosgolem-logh3/workplace/fcmd/fcmd-base.state', 'checkpoint_sha256': sha(checkpoint),
         'address_space': 'dosgolem實體線性位址；原版CS:IP獨立列示。',
         'scope': '正常入口加明示來源哨兵的診斷；新遊戲初始化、正常S3讀檔、達希利戰場入口。未宣稱所有存檔、第二戰場或所有DOS環境。',
         'bootstrap_clear': {'grade': 'proven', 'original_location': '0349:01EA', 'step': 58339340,
                             'source': 'newgame兩次原版值改變監看；不是以零值猜測初始化。'},
         'rows': rows,
         'source_sha256': {n: sha((root/n).read_bytes()) for n in
                           ('tools/battle_preparation_lifecycle_oracle.sh', 'tools/scripts/battle_preparation_reset_receipt.py')}}
out.write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')
print('三組入口各雙跑完整記憶體／畫面／色盤相同；來源及池基底保留。')
print('欄位收據 SHA-256', sha(out.read_bytes()))
