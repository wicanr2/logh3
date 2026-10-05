"""116完整來源、離線二進位、正常標題及FM技術收據。"""
import argparse
import ast
import hashlib
import json
import os
from pathlib import Path
import tarfile

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('--freeze', action='store_true')
p.add_argument('--check', action='store_true')
a = p.parse_args()
assert a.freeze != a.check
target = Path('docs/internals/data/current-source.json')


def sha(path):
    with Path(path).open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


if a.freeze:
    assert not target.exists()
    first = Path('workplace/source69-r103-d')
    second = Path('workplace/source69-r103-e')
    bundle = json.loads((first/'BUNDLE.json').read_text())
    repeat = json.loads((second/'BUNDLE.json').read_text())
    assert bundle == repeat and sha(first/'source.tar') == sha(second/'source.tar') == bundle['archive_sha256']
    assert bundle['file_count'] == 2810 and bundle['all_members_verified']
    manifest = bundle['manifest']
    assert manifest['schema'] == 'logh3.music-source-kit.v1'
    assert manifest['inputs']['logh3']['head'] == '9e1d31830f3d5555cfc313e06f83b0d159186ff1'
    assert manifest['inputs']['dosgolem']['head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
    assert manifest['builder_sha256'] == sha('tools/scripts/remake_source_kit.py')
    for name, data in manifest['attachments'].items():
        assert sha(name) == data['sha256']
    for root in (first, second):
        with tarfile.open(root/'source.tar', 'r:') as archive:
            entries = archive.getmembers()
            assert len(entries) == bundle['file_count'] and len({e.name for e in entries}) == len(entries)
            assert all(e.isfile() and (e.uid, e.gid, e.mtime) == (1000, 1000, 0) for e in entries)
            for e in entries:
                if e.name == 'MANIFEST.json':
                    assert json.loads(archive.extractfile(e).read()) == manifest
                    continue
                row = manifest['files'][e.name]
                assert e.size == row['bytes'] and e.mode == row['mode']
                assert hashlib.sha256(archive.extractfile(e).read()).hexdigest() == row['sha256']
    rebuild = json.loads((first/'rebuild/REBUILD.json').read_text())
    preliminary = json.loads(Path('workplace/source69-r103-b/rebuild/REBUILD.json').read_text())
    assert rebuild['binaries'] == preliminary['binaries'] and len(rebuild['binaries']) == 7
    assert rebuild['cli_pass'] == 6 and rebuild['cli_skip'] == 0
    assert rebuild['offline_empty_module_cache'] and rebuild['vendor_equal']
    title = json.loads(Path('workplace/source69-r103-b/runtime-title-second/CHECK.json').read_text())
    assert title['state'] == 'StateTitle' and title['pixels'] == 307200 and title['different'] == 0
    assert title['source_binary_sha256'] == rebuild['binaries'][0]['sha256']
    assert title['updates'] == 4998 and not title['script_used'] and not title['state_injection']
    fm = json.loads(Path('workplace/source69-r103-b/runtime-fm/CHECK.json').read_text())
    assert fm['status'] == 'complete-source-built-fm-sample'
    assert fm['tools']['music-import'] == next(r['sha256'] for r in rebuild['binaries'] if r['name'] == 'music-import-linux')
    windows = json.loads(Path('workplace/music69-windows/source103-player/CHECK.json').read_text())
    assert windows['exit'] == 0 and windows['sha256'] == rebuild['binaries'][1]['sha256']
    boundary = json.loads(Path('workplace/source69-r103-boundary.json').read_text())
    assert boundary['existing_preserved'] and boundary['invalid_head_no_output']
    tests = Path('workplace/round103-standard-gui.log').read_text().splitlines()
    assert sum(t.split()[:1] == ['ok'] for t in tests) == 13
    assert sum('[no test files]' in t for t in tests) == 5 and not any(t.startswith('FAIL') for t in tests)
    for f in Path('tools/scripts').glob('remake_source_*.py'):
        ast.parse(f.read_text())
    paths = {'docs/goal/2026-10-05-current-source-round-103.md', 'docs/spec/116-current-source.md',
             'docs/playtest/091-current-source.md', 'docs/spec/077-music-source-delivery.md',
             'docs/internals/data/music-source-delivery.json', 'docs/internals/data/fm-import-lock.json',
             'tools/scripts/music_source_kit.py', 'tools/gui.sh', 'tools/py.sh', 'tools/music_wine.sh',
             'tools/music_wine_entry.sh', 'workplace/round103-standard-gui.log',
             'workplace/music69-delivery/source-kit-vcs-root/music-source-kit.tar'}
    paths.update(str(f) for f in Path('tools/scripts').glob('remake_source_*.py'))
    for f in Path('workplace').glob('source69-r103-*'):
        if f.is_file():
            paths.add(str(f))
        else:
            paths.update(str(v) for v in f.rglob('*') if v.is_file())
    paths.update(str(v) for v in Path('workplace/music69-windows/source103-player').rglob('*') if v.is_file())
    sources = {}
    for name in sorted(paths):
        f = Path(name)
        assert f.stat().st_uid == f.stat().st_gid == 1000 and not f.is_symlink(), name
        sources[name] = {'bytes': f.stat().st_size, 'sha256': sha(f)}
    result = {'schema': 'logh3.current-source-receipts.v1', 'date': '2026-10-05', 'issue': 69,
              'parent_commit': manifest['inputs']['logh3']['head'], 'source_bundle': {
                  'paths': [str(first/'source.tar'), str(second/'source.tar')],
                  'bytes': bundle['bytes'], 'files': bundle['file_count'], 'sha256': bundle['archive_sha256']},
              'images': {'build_title': 'logh3-ebiten:2', 'fm': 'python:3.12-slim',
                         'windows_help': 'colonization-windows-verification:20261001-r1'},
              'rebuild': rebuild, 'title': title, 'fm': fm, 'windows_help': windows,
              'standard': {'packages': 18, 'pass': 13, 'no_tests': 5},
              'spec116': 'CONFORMED，限定來源封裝及指定Go入口重建／技術抽驗',
              'limitations': ['來源不含原版媒體／生成曲庫，也不包含全部合成器／編碼器來源',
                              'Windows玩家程式只驗建置及Wine help，未驗遊戲GUI或實機',
                              '正常標題比目前程式，未新增原版正常整回合同態',
                              '正式版號／dist-all、macOS、實機、人耳及循環仍缺，#69 OPEN'],
              'sources': sources}
    with target.open('x') as stream:
        stream.write(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
record = json.loads(target.read_text())
for name, row in record['sources'].items():
    f = Path(name)
    assert f.stat().st_size == row['bytes'] and sha(f) == row['sha256'], name
print(len(record['sources']), '來源身份通過，索引SHA-256', sha(target))
