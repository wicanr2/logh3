"""規格140：核對GitHub Release下載身份，發布後才登錄完成訊號。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('mode',choices=['draft','published'])
a=p.parse_args()
ROOT=Path('/src')
BATCH=Path('/work/hero130') if Path('/work/hero130').is_dir() else ROOT/'workplace/hero130'
VERSION='v.1.2.0-20261012'
COMMIT='4096431f43f84a6bfe0588171401c497672f4a98'
TAG='8444344e9e0feaba4efeffb02cee4c2d782af484'
out=ROOT/'dist-all'/VERSION


def read(path):return json.loads(path.read_text())


def sha(path):
    h=hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda:f.read(1024*1024),b''):h.update(chunk)
    return h.hexdigest()


def write(path,value):
    assert path.parent.stat().st_uid==os.getuid()
    path.write_text(json.dumps(value,ensure_ascii=False,indent=2)+'\n')


def record(path):
    assert path.is_file() and path.stat().st_uid==os.getuid()
    return dict(path=str(path.relative_to(ROOT)),bytes=path.stat().st_size,sha256=sha(path))


build=read(out/'BUILD.json')
items=[i for i in build['packages'] if i['path'].startswith('release/')]
expected={Path(i['path']).name:i for i in items}
assert len(items)==4
remote=read(BATCH/('GH-DRAFT.json' if a.mode=='draft' else 'GH-RELEASE.json'))
assert remote['tag_name']==remote['name']==VERSION
assert remote['draft']==(a.mode=='draft') and not remote['prerelease']
assert remote['target_commitish']==COMMIT
assert len(remote['assets'])==5
assert {s['name'] for s in remote['assets']}==set(expected)|{'SHA256SUMS.txt'}
download=BATCH/'gh-download'
assert {p.name for p in download.iterdir()}==set(expected)|{'SHA256SUMS.txt'}
checked=[]
for asset in remote['assets']:
    assert asset['state']=='uploaded'
    path=download/asset['name']
    assert path.stat().st_uid==os.getuid() and path.stat().st_size==asset['size']
    digest=sha(path)
    assert digest==sha(out/'release'/asset['name'])
    if asset['digest']:assert asset['digest']=='sha256:'+digest
    if asset['name'] in expected:
        item=expected[asset['name']]
        assert item['bytes']==asset['size'] and item['sha256']==digest
    checked.append(dict(name=asset['name'],bytes=asset['size'],sha256=digest,id=asset['id']))
proof=dict(version=VERSION,passed=True,release_id=remote['id'],release_commit=COMMIT,
           assets=checked,downloaded_and_hashed=True,rights='四引擎／來源包及校驗檔，無原版／HD美術／動畫',
           authorization='2026-10-13 使用者授權本次主機gh發布')
if a.mode=='draft':
    write(BATCH/'GH-DOWNLOAD-CHECK.json',proof)
    print('PASS GitHub草稿五資產實際下載，逐檔SHA-256與本機相同')
else:
    assert remote['published_at'] and remote['html_url']=='https://github.com/wicanr2/logh3/releases/tag/'+VERSION
    assert remote['body'].strip()==(out/'release/RELEASE-NOTES.md').read_text().strip()
    tag=read(BATCH/'GH-TAG.json')
    assert tag['sha']==TAG and tag['tag']==VERSION
    assert tag['object']['type']=='commit' and tag['object']['sha']==COMMIT
    proof.update(url=remote['html_url'],published_at=remote['published_at'],tag_object=TAG)
    write(BATCH/'GH-PUBLISHED-CHECK.json',proof)
    for name in ['GH-PREFLIGHT.json','GH-DRAFT.json','GH-DOWNLOAD-CHECK.json',
                 'GH-RELEASE.json','GH-TAG.json','GH-PUBLISHED-CHECK.json']:
        shutil.copyfile(BATCH/name,out/'smoke'/name)
    index=ROOT/'docs/internals/data/hero-ai-release-promo.json'
    data=read(index)
    data['published_assets']=[dict(expected[s['name']],name=s['name'],id=s['id'],
        download_url=s['browser_download_url'],remote_digest=s['digest'])
        for s in remote['assets'] if s['name'] in expected]
    data['published_checksum']=next(s for s in remote['assets'] if s['name']=='SHA256SUMS.txt')
    data['remote_release']=dict(url=remote['html_url'],id=remote['id'],tag=VERSION,commit=COMMIT,
        tag_object=TAG,published_at=remote['published_at'],downloaded_and_hashed=True,
        authorization='2026-10-13 使用者：授權gh')
    data['status']='PASS 名將戰術、三平台本機完整版、有聲史詩推廣片與GitHub引擎發行'
    data['sources']=[s for s in data['sources'] if s['path']!='tools/scripts/hero_release_publish_check.py']
    data['sources'].append(record(ROOT/'tools/scripts/hero_release_publish_check.py'))
    names={str((out/'smoke'/n).relative_to(ROOT)) for n in
           ['GH-PREFLIGHT.json','GH-DRAFT.json','GH-DOWNLOAD-CHECK.json','GH-RELEASE.json','GH-TAG.json','GH-PUBLISHED-CHECK.json']}
    data['evidence']=[s for s in data['evidence'] if s['path'] not in names]
    data['evidence'] += [record(ROOT/n) for n in sorted(names)]
    write(index,data)
    print('PASS 已正式發布；遠端tag、五資產、發行說明及下載雜湊相同')
