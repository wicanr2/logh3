"""規格 077：重新核對來源包、授權、五種建置與正式匯入後寫唯一完成收據。

LOGH3_EXTRA_RO 指向專案，tools/py.sh /work/tools/scripts/music_source_receipt.py。
"""
import hashlib
import json
import os
from pathlib import Path
import re
import tarfile

import music_source_kit as kit

ROOT = Path('/work/workplace/music69-delivery')
REPO = Path('/ref')


def sha(p):
    kit.owned(p)
    with p.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def load(p):
    return json.loads(p.read_text())


first = load(ROOT/'source-kit/BUNDLE.json')
bundle = load(ROOT/'source-kit-vcs-root/BUNDLE.json')
repeat = load(ROOT/'source-kit-vcs-root-repeat/BUNDLE.json')
assert bundle['archive_sha256'] == repeat['archive_sha256']
manifest = bundle['manifest']
assert manifest['builder_sha256'] == sha(REPO/'tools/scripts/music_source_kit.py')
assert manifest['license_sha256'] == sha(REPO/'LICENSE')
assert manifest['license_substitutions'] == kit.license_mapping((ROOT/'RRSAL-TEMPLATE.txt').read_text(), (REPO/'LICENSE').read_text())
assert manifest['license_template_sha256'] == sha(ROOT/'RRSAL-TEMPLATE.txt')
for project, row in manifest['inputs'].items():
    assert row['head'] == kit.HEADS[project]
    p = ROOT/('logh3-source-reviewed.tar' if project == 'logh3' else 'dosgolem-source-full.tar')
    assert sha(p) == row['archive_sha256']
    kit.read_git(p, row['head'], project)
for name, row in manifest['files'].items():
    if name.startswith('logh3/') and not name.startswith('logh3/vendor/'):
        assert sha(REPO/name.removeprefix('logh3/')) == row['sha256'], name
for directory in ['source-kit-vcs-root', 'source-kit-vcs-root-repeat']:
    base = ROOT/directory
    assert sha(base/'music-source-kit.tar') == bundle['archive_sha256']
    assert (base/'music-source-kit.tar').stat().st_size == bundle['bytes']
    with tarfile.open(base/'music-source-kit.tar', 'r:') as archive:
        entries = archive.getmembers()
        assert len(entries) == bundle['file_count'] == len(manifest['files'])+1
        assert {p.name for p in entries} == set(manifest['files']) | {'MANIFEST.json'}
        for entry in entries:
            assert entry.isfile() and entry.name == kit.safe_name(entry.name)
            row = manifest['files'].get(entry.name, {'sha256': bundle['manifest_sha256'], 'bytes': entry.size, 'mode': 0o644})
            assert (entry.size, entry.mode, entry.uid, entry.gid, entry.mtime) == (row['bytes'], row['mode'], 1000, 1000, 0)
            assert hashlib.sha256(archive.extractfile(entry).read()).hexdigest() == row['sha256']
            assert sha(base/'extracted'/entry.name) == row['sha256']
    assert sha(base/'extracted/MANIFEST.json') == bundle['manifest_sha256']

out = ROOT/'rebuild-vcs-root'
assert (out/'vendor-diff.txt').read_bytes() == b''
versions = [(out/'go-version.txt').read_text(), (ROOT/'dosgolem-reference-go-version.txt').read_text()]
assert versions[0] == versions[1] == 'go version go1.24.13 linux/amd64\n'
builds = {}
for name in ['music-import-linux', 'music-import-windows', 'music-tool-linux', 'dosgolem-probe-windows', 'dosgolem-probe-linux']:
    reference = (ROOT/(name+'-vcs-reference')) if name.startswith('dosgolem') else out/(name+'-reference')
    source = out/(name+'-source')
    assert sha(reference) == sha(source), name
    builds[name] = {'sha256': sha(source), 'bytes': source.stat().st_size, 'reference': str(reference.relative_to(ROOT)), 'source': str(source.relative_to(ROOT))}
events = [json.loads(line) for line in (out/'cli-tests.jsonl').read_text().splitlines()]
assert not any(e['Action'] == 'fail' for e in events)
assert {e['Package'] for e in events if e['Action'] == 'pass' and 'Test' not in e} == {'github.com/wicanr2/logh3/cmd/music-import', 'github.com/wicanr2/logh3/cmd/music-tool'}
skips = [e for e in events if e['Action'] == 'skip']
assert not skips
assert any(e['Action'] == 'pass' and e.get('Test') == 'TestFMProbeIdentity' for e in events)
git_proof = (out/'dosgolem-git-proof.txt').read_text()
assert git_proof == '/src/workplace/music69-delivery/source-kit-vcs-root/extracted/dosgolem\n'+kit.HEADS['dosgolem']+'\n'
for platform in ['linux', 'windows']:
    info = (out/('dosgolem-'+platform+'-buildinfo.txt')).read_text()
    assert '\tvcs.revision='+kit.HEADS['dosgolem']+'\n' in info
    assert '\tvcs.modified=false\n' in info
text = (ROOT/'source-rebuild-vcs-root.log').read_text()
assert text.endswith('來源側空模組快取及無網路條件下的三組對照建置、dosgolem 真實 Git 身份建置與 CLI 測試通過。\n')
boundary = (ROOT/'source-boundary-vcs-root-tests.log').read_text()
assert re.search(r'Ran 4 tests in .+\n\nOK\n', boundary)
product = (ROOT/'source-product-tests.log').read_text()
expected = set((ROOT/'source-product-packages.txt').read_text().splitlines())
actual = {m[1] for m in re.findall(r'^(ok|\?)\s+(\S+)', product, re.M)}
assert actual == expected and not re.search(r'^FAIL', product, re.M)
failure_events = [json.loads(line) for line in (ROOT/'rebuild-2/cli-tests.jsonl').read_text().splitlines()]
assert any(e['Action'] == 'fail' and e.get('Test') == 'TestFMImportLocksAndRoutes' for e in failure_events)
assert 'differ:' in (ROOT/'source-rebuild.log').read_text()
assert sha(ROOT/'source-kit/music-source-kit.tar') == first['archive_sha256']
assert sha(ROOT/'source-kit-repeat/music-source-kit.tar') == first['archive_sha256']
sample = load(ROOT/'source-import-vcs-root/CHECK.json')
assert sample['status'] == 'complete-source-built-fm-sample'
assert sample['source_archive_sha256'] == bundle['archive_sha256']
for file, build in [('music-import', 'music-import-linux'), ('tools/dosgolem-probe', 'dosgolem-probe-linux'), ('tools/ffmpeg', 'music-tool-linux')]:
    assert sample['tools_sha256'][file] == builds[build]['sha256']
track = sample['fm']['tracks'][0]
lock = load(REPO/'docs/internals/data/fm-import-lock.json')['tracks']['BGM_0F']
assert track['track'] == 'BGM_0F'
for key in ['vgm_sha256', 'wav_sha256', 'pcm_sha256']:
    assert track[key] == lock[key]
assert sha(ROOT/'source-import-vcs-root/import.log') == sample['import_log_sha256']
for name, value in sample['raw_outputs_sha256'].items():
    assert sha(ROOT/'source-import-vcs-root/匯入 曲庫'/name) == value
logs = ['source-rebuild.log', 'source-rebuild-2.log', 'source-rebuild-reviewed.log',
        'source-rebuild-vcs-2.log', 'source-rebuild-vcs-root.log',
        'source-boundary-tests.log', 'source-boundary-reviewed-tests.log', 'source-boundary-vcs-root-tests.log', 'source-product-tests.log',
        'source-product-packages.txt', 'rebuild/reference-buildinfo.txt', 'rebuild/source-buildinfo.txt',
        'rebuild-2/cli-tests.jsonl', 'rebuild-reviewed/cli-tests.jsonl', 'rebuild-vcs-root/cli-tests.jsonl',
        'rebuild-vcs-root/dosgolem-git-proof.txt', 'rebuild-vcs-root/dosgolem-linux-buildinfo.txt',
        'rebuild-vcs-root/dosgolem-windows-buildinfo.txt', 'dosgolem-reference-vcs-buildinfo.txt',
        'source-import-vcs-root/CHECK.json', 'source-import-vcs-root/import.log']
scripts = ['music_source_kit.py', 'music_source_kit_test.py', 'music_source_rebuild.sh', 'music_source_receipt.py', 'music_source_import_check.py']
record = {'schema': 'logh3.music-source-delivery.v1', 'date': '2026-10-02',
          'status': '本作品 RRSAL 全文與音樂 Go 來源套件技術驗收完成；正式發行未完成',
          'source_parent_head': kit.HEADS['logh3'], 'validation_tree_dirty': True,
          'bundle': bundle, 'repeat_archive_sha256': repeat['archive_sha256'],
          'builds': builds, 'go_version': versions[0].strip(),
          'git_root_proof': git_proof, 'formal_fm_sample': sample,
          'before_vcs_receipt_sha256': sha(ROOT/'music-source-before-vcs.json'),
          'failed_vcs_archive_sha256': sha(ROOT/'source-kit-vcs/music-source-kit.tar'),
          'cli_packages': {e['Package']: e.get('Elapsed') for e in events if e['Action'] == 'pass' and 'Test' not in e},
          'cli_test_pass_events': sum(e['Action'] == 'pass' and 'Test' in e for e in events),
          'cli_skips': [{'test': e['Test'], 'reason': ''.join(x.get('Output', '') for x in events if x.get('Test') == e['Test'] and x['Action'] == 'output')} for e in skips],
          'boundary_tests': 4, 'product_packages': sorted(actual),
          'first_archive': {'sha256': first['archive_sha256'], 'file_count': first['file_count'], 'reason': '首次 h1 建置資訊差異；統一 vendor 後二進位相同，但來源測試揭露 FM 固定鎖漏項'},
          'scripts_sha256': {n: sha(REPO/'tools/scripts'/n) for n in scripts},
          'logs_sha256': {n: sha(ROOT/n) for n in logs},
          'gui_image': 'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
          'formal_delivery_complete': False,
          'limits': '限指定 Go 入口的來源重建；第三方合成器／編碼器完整來源另依 033／035 配送。不等於 macOS、實機、聆聽／循環或正式版號／dist-all 已通過；私有原版素材與歷史仍不可因 LICENSE 補齊而公開。'}
target = Path('/work/docs/internals/data/music-source-delivery.json')
kit.parent_owned(ROOT)
assert not target.is_symlink() and (target.parent.stat().st_uid, target.parent.stat().st_gid) == (os.getuid(), os.getgid())
if target.exists():
    kit.owned(target)
target.write_text(json.dumps(record, ensure_ascii=False, indent=2)+'\n')
print('來源重建完成收據已保存，五種二進位相同、2361 檔／兩次 tar 相同、CLI 無跳過及正式 FM 匯入通過。')
print('SHA-256', sha(target))
