"""#79 排序收據核對與欄位匯出；原始記錄只留本機，入口見研究368。"""
import argparse
import hashlib
import json
from pathlib import Path


def digest(data):
    return hashlib.sha256(data).hexdigest()


def comb_model(order, units, donor):
    """可丟棄的診斷模型；不接正式遊戲或充當原版輸出。"""
    values = list(order)
    gap = len(values)
    if gap < 2:
        return values

    def compare(a, b):
        a, b = units[a], units[b]
        if a['fleet'] == b['fleet']:
            return (a['squadron'] > b['squadron']) - (a['squadron'] < b['squadron'])
        if a['fleet'] == donor:
            return -1
        if b['fleet'] == donor:
            return 1
        return 0

    while True:
        gap = gap * 10 // 13
        gap = 1 if gap == 0 else 11 if gap in (9, 10) else gap
        swaps = 0
        for i in range(len(values) - gap):
            if compare(values[i], values[i + gap]) > 0:
                values[i], values[i + gap] = values[i + gap], values[i]
                swaps += 1
        if gap <= 1 and swaps == 0:
            return values


def decode_unit(value):
    b = bytes.fromhex(value)
    assert len(b) == 78
    return {'faction': b[0], 'fleet': b[1], 'squadron': b[3], 'kind': b[4],
            'ships': int.from_bytes(b[8:10], 'little'),
            'full_ships': int.from_bytes(b[10:12], 'little'),
            'repair_cap': int.from_bytes(b[12:14], 'little'),
            'missiles': int.from_bytes(b[14:18], 'little'),
            'fighters': int.from_bytes(b[18:22], 'little')}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--output', required=True)
    parser.add_argument('--golem-head', required=True)
    parser.add_argument('--golem-dirty', required=True)
    args = parser.parse_args()
    assert args.golem_head == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
    assert args.golem_dirty == ''
    directory = Path('workplace/oracle/battle79-sort')
    a = (directory / 'run-a.json').read_bytes()
    assert a == (directory / 'run-b.json').read_bytes(), '原版雙跑不一致'
    assert digest(a) == '370cfe8b8c7ca2efc1d23e8db6f4dab646be29cda33b2b3123b737a3fc49cb99'
    probe = Path('tools/scripts/battle_preparation_sort_probe.go').read_bytes()
    assert probe == Path('workplace/battle_preparation_sort_probe.go').read_bytes(), '實際探針來源不符'
    original = json.loads(a)
    assert original['schema'] == 'logh3.battle-preparation-sort-oracle.v1'
    assert len(original['rows']) == 75
    assert all(row['rng_after_hex'] == '78563412' for row in original['rows'])
    cases = []
    for row in original['rows']:
        if 'before' in row:
            assert row['after'] == comb_model(row['before'], row['units'], row['donor_fleet']), row['name']
            cases.append({key: row[key] for key in ('name', 'entry', 'comparator', 'units', 'donor_fleet', 'before', 'after', 'rng_after_hex')})
        else:
            cases.append({'name': row['name'], 'entry': row['entry'],
                          'units_before': [decode_unit(u) for u in row['units_before_hex']],
                          'units_after': [decode_unit(u) for u in row['units_after_hex']],
                          'source_pointers_after_hex': row['source_pointers_after_hex'],
                          'rng_after_hex': row['rng_after_hex']})
    for name, pointer, index, field, value in (
        ('planet-supply-last-unit', '9e008b66/ec008b66', 1, 'missiles', 100),
        ('planet-repair-retained-source', 'ec008b66/ec008b66', 0, 'ships', 70),
        ('planet-repair-last-donor', 'ec008b66/9e008b66', 1, 'ships', 70),
    ):
        row = next(r for r in cases if r['name'] == name)
        assert row['source_pointers_after_hex'] == pointer
        assert row['units_after'][index][field] == value
    backlinks = [
        {'original_key': 'DOS/GIN3.COM/1930:0510', 'older_file': 'docs/re/349-1930-supply-family.md',
         'required_marker': '2026-10-02排序勘誤', 'state': 'corrected'},
        {'original_key': 'DOS/GIN3.COM/1AAB:04BD', 'older_file': 'docs/re/367-battle-preparation.md',
         'required_marker': '2026-10-02追加：排序與指標未知縮小', 'state': 'corrected'},
        {'original_key': 'DOS/GIN3.COM/3C6A:000D', 'older_file': 'docs/spec/079-battle-preparation.md',
         'required_marker': 'READY：自動候選排序', 'state': 'corrected'},
        {'original_key': 'DOS/GIN3.COM/3C6A:000D', 'older_file': 'docs/spec/029-strategic-ai-turn.md',
         'required_marker': '研究368', 'state': 'unaffected',
         'reason': '平手次序沒有保證仍成立；本輪沒有驗AI配額比較器或改AI正式路徑。'},
    ]
    evidence = Path('docs/re/368-battle-preparation-sort.md').read_text()
    for row in backlinks:
        assert row['original_key'].split('/')[-1] in evidence, '證據缺原始定位'
        assert row['required_marker'] in Path(row['older_file']).read_text(), '舊文件缺勘誤入口'
        row['evidence_file'] = 'docs/re/368-battle-preparation-sort.md'
        row['program_sha256'] = original['program_sha256']
    data = {'schema': 'logh3.battle-preparation-sort.v1', 'grade': 'proven',
            'scope': original['scope'], 'seed': original['seed'], 'seed_linear': original['seed_linear'],
            'program_sha256': original['program_sha256'], 'memory_sha256': original['input_sha256'],
            'golem_head': args.golem_head, 'golem_worktree': 'clean',
            'remake_parent_head': '1a036681d4c2ebf136c9ebeb481bb14ceb60e3ea',
            'remake_worktree': '第二十六輪研究／探針未提交；正式規則未修改',
            'go': {'version': '1.24.13', 'image': 'golang:1.24-bookworm',
                   'image_id': 'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac'},
            'probe_path': 'tools/scripts/battle_preparation_sort_probe.go', 'probe_sha256': digest(probe),
            'generator_path': 'tools/scripts/battle_preparation_sort_receipt.py',
            'generator_sha256': digest(Path(__file__).read_bytes()),
            'local_replays': [{'path': str(directory / f'run-{run}.json'), 'sha256': digest(a),
                              'command': 'DOSGOLEM_ORIG=/home/anr2/cht/logh3/org_game/extracted DOSGOLEM_EXTRA_MOUNT=/home/anr2/cht/logh3/workplace:/out bash /home/anr2/cht/dosgolem-logh3/tools/go.sh run /out/battle_preparation_sort_probe.go -output /out/oracle/battle79-sort/run-'+run+'.json'} for run in ('a', 'b')],
            'diagnostic_model_matches': 72, 'resolution_backlinks': backlinks, 'cases': cases}
    out = Path(args.output)
    assert out.parent.stat().st_uid == 1000
    encoded = (json.dumps(data, ensure_ascii=False, indent=2) + '\n').encode()
    out.open('xb').write(encoded)
    print('原版75案雙跑相同；72案診斷模型相同；SHA-256', digest(encoded))


if __name__ == '__main__':
    main()
