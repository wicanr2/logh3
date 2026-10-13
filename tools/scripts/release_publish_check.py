"""核對新GitHub發行的實際下載、固定tag及資產權利範圍。"""
import argparse
import hashlib
import json
import os
from pathlib import Path

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('mode', choices=['draft', 'published'])
p.add_argument('--version', required=True)
p.add_argument('--commit', required=True)
p.add_argument('--tag', required=True)
p.add_argument('--batch', required=True)
p.add_argument('--index', required=True)
a = p.parse_args()
ROOT = Path('/src')
BATCH = ROOT / a.batch
OUT = ROOT / 'dist-all' / a.version


def read(path):
    return json.loads(path.read_text())


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b''):
            h.update(chunk)
    return h.hexdigest()


def write(path, value):
    assert path.parent.stat().st_uid == os.getuid()
    if path.exists():
        assert path.stat().st_uid == os.getuid()
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


items = [i for i in read(OUT / 'BUILD.json')['packages'] if i['path'].startswith('release/')]
assert read(BATCH / 'GH-PREFLIGHT.json')['isPrivate'] is True
source = read(BATCH / 'COMMIT-SOURCE-CHECK.json')
assert source['passed'] and source['commit'] == a.commit and source['files_checked'] == 575
expected = {Path(i['path']).name: i for i in items}
assert len(items) == 4
remote = read(BATCH / ('GH-DRAFT.json' if a.mode == 'draft' else 'GH-RELEASE.json'))
assert remote['tag_name'] == remote['name'] == a.version
assert remote['draft'] == (a.mode == 'draft') and not remote['prerelease']
assert remote['target_commitish'] == a.commit
assert {s['name'] for s in remote['assets']} == set(expected) | {'SHA256SUMS.txt'}
assert len(remote['assets']) == 5
download = BATCH / 'gh-download'
assert {i.name for i in download.iterdir()} == set(expected) | {'SHA256SUMS.txt'}
checked = []
for asset in remote['assets']:
    path = download / asset['name']
    assert asset['state'] == 'uploaded' and path.stat().st_uid == os.getuid()
    assert path.stat().st_size == asset['size']
    digest = sha(path)
    assert digest == sha(OUT / 'release' / asset['name'])
    if asset.get('digest'):
        assert asset['digest'] == 'sha256:' + digest
    if asset['name'] in expected:
        item = expected[asset['name']]
        assert item['bytes'] == asset['size'] and item['sha256'] == digest
    checked.append(dict(name=asset['name'], bytes=asset['size'], sha256=digest, id=asset['id']))
proof = dict(version=a.version, passed=True, release_id=remote['id'], release_commit=a.commit,
             assets=checked, downloaded_and_hashed=True,
             rights='四引擎／來源包與校驗檔；不含原版、HD遊戲素材或配樂；啟動字集附授權',
             authorization='使用者授權gh、發布release及相關Issue；完整版只留本機')
if a.mode == 'draft':
    write(BATCH / 'GH-DOWNLOAD-CHECK.json', proof)
else:
    assert remote['published_at'] and remote['html_url'] == 'https://github.com/wicanr2/logh3/releases/tag/' + a.version
    assert remote['body'].strip() == (OUT / 'release/RELEASE-NOTES.md').read_text().strip()
    tag = read(BATCH / 'GH-TAG.json')
    assert tag['sha'] == a.tag and tag['tag'] == a.version
    assert tag['object']['type'] == 'commit' and tag['object']['sha'] == a.commit
    proof.update(url=remote['html_url'], published_at=remote['published_at'], tag_object=a.tag)
    write(BATCH / 'GH-PUBLISHED-CHECK.json', proof)
    index = ROOT / a.index
    data = read(index)
    data['published_assets'] = [dict(expected[s['name']], name=s['name'], id=s['id'],
        download_url=s['browser_download_url'], remote_digest=s.get('digest'))
        for s in remote['assets'] if s['name'] in expected]
    data['published_checksum'] = next(s for s in remote['assets'] if s['name'] == 'SHA256SUMS.txt')
    data['remote_release'] = dict(url=remote['html_url'], id=remote['id'], tag=a.version,
        commit=a.commit, tag_object=a.tag, published_at=remote['published_at'], downloaded_and_hashed=True)
    write(index, data)
print('PASS GitHub ' + a.mode + ' 五資產實際下載、SHA-256及發行身份')
