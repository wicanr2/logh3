"""#85 完整表收據摘要；入口 docs/playtest/027-post-turn-runtime-fields.md。

tools/py.sh tools/scripts/runtime_fields_summary.py
原版重生：tools/runtime_fields_receipts.sh；語意索引另經人工審查，不由本程式升格。
"""
import hashlib
import json
from pathlib import Path


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


root = Path('workplace/oracle')
report = {'schema': 1, 'comparison': 'same-state（指定 S1 首回合全表及未注入 GUI）',
          'source': 'docs/playtest/027-post-turn-runtime-fields.md',
          'semantic_index': 'docs/internals/data/post-turn-runtime-fields.json',
          'seed': {'original': '0xD3DC167E', 'remake': '0xD3DC167E',
                   'setup': '執行前固定檢查點與 strategyRandSeed；初始 dump 實讀 DS:94DC'},
          'limits': ['不宣稱整段骰序或全遊戲一致', '完全平手的原版 qsort 次序未裁決',
                     '#86 現況艦隊／艦艇保存缺口未完成；本輪只驗全表寫出及新增距離／座標讀回'],
          'batches': []}
for batch in ('runtime85-normal', 'runtime85-stationed'):
    directory = root / batch
    inputs = json.loads((directory / 'inputs.json').read_text())
    # 保留包裝器原有 metadata，不假定診斷分支與正常分支的欄位形狀相同。
    memory = []
    for path in sorted(directory.glob('*.bin')):
        if path.name.startswith('remake'):
            continue
        original = path.read_bytes()
        assert original == (root / (batch + '-r2') / path.name).read_bytes(), path
        memory.append({'name': path.name, 'bytes': len(original), 'sha256': sha(original)})
    for name in ('screen.png', 'pal.bin'):
        assert (directory / name).read_bytes() == (root / (batch + '-r2') / name).read_bytes()
    original = (directory / 'command.bin').read_bytes()
    assert len(original) == 1048576
    tables = []
    for name, address, size in [('stellar', 0x5F2CC, 4032), ('planet', 0x53132, 6642)]:
        a, b = original[address:address+size], (directory / ('remake-' + name + '.bin')).read_bytes()
        assert len(a) == len(b) == size and a == b
        tables.append({'name': name, 'original_address_space': 'dosgolem 執行期線性位址',
                       'original_linear_hex': hex(address), 'bytes': size, 'different': 0,
                       'original_sha256': sha(a), 'remake_sha256': sha(b)})
    record = {'name': batch, 'repeated_original_equal': True, 'inputs': inputs,
              'memory': memory, 'tables': tables,
              'remake_test': 'internal/game.TestRuntimeFieldsNormalTurnMatchesOriginal',
              'remake_metadata': json.loads((directory / 'remake-tables.json').read_text()),
              'metadata_sha256': {name: sha((directory / name).read_bytes())
                                  for name in ('inputs.json', 'flags.txt', 'meta.txt')}}
    if batch == 'runtime85-normal':
        record['gui'] = json.loads((directory / 'stats.json').read_text())
        record['remake_gui_metadata'] = json.loads((directory / 'remake-meta.json').read_text())
        assert record['gui']['index_different'] == 0
        assert all(r['different'] == 0 for r in record['gui']['regions'].values())
    report['batches'].append(record)
report['ida_exports'] = {p.name: sha(p.read_bytes())
                         for p in sorted(Path('workplace/ida').glob('re85-*-reviewed.json'))}
assert len(report['ida_exports']) == 3
report['resolution_backlinks'] = []
for old in ('docs/re/89-minimap-unit-markers.md', 'docs/re/320-region-info-window.md',
            'docs/re/353-region-supply-turn-timing.md', 'docs/spec/043-ai-capital-reach.md',
            'docs/playtest/026-region-supply-turn.md', 'docs/internals/22-stellar-dt.md',
            'docs/internals/29-planets-and-economy.md', 'docs/internals/58-strategic-fleet-window.md'):
    text = Path(old).read_text()
    assert '354-post-turn-runtime-fields.md' in text, ('缺勘誤回指', old)
    report['resolution_backlinks'].append({'original_key': 'DOS GIN3.COM 固定雜湊；星域 +0x78／行星 +0x46,+0x48',
        'evidence': 'docs/re/354-post-turn-runtime-fields.md', 'older_document': old,
        'required_marker': '354-post-turn-runtime-fields.md'})
output = Path('docs/internals/data/post-turn-runtime-receipts.json')
output.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
print('兩組原版重跑、完整 4032／6642 bytes 與正常 GUI 零差異；摘要：', output)
