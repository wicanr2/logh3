"""第139輪：發布後以實際下載附件及Git tag核對公開Release。Docker內執行。"""
import hashlib
import json
from pathlib import Path
import subprocess
import argparse
import re

root=Path('/src')
version='v.1.4.0-20261015'
batch=root/'workplace/promo139'
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--version',default=version)
p.add_argument('--batch',default=str(batch))
p.add_argument('--index',default='docs/internals/data/v140-release-promo.json')
args=p.parse_args()
version=args.version
assert re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}',version)
batch=Path(args.batch)
assert batch.resolve().is_relative_to((root/'workplace').resolve())
delivery=root/'dist-all'/version


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):return json.loads(path.read_text())


remote=read(batch/'GH-RELEASE.json')
assert remote['tagName']==remote['name']==version and not remote['isDraft']
assert remote['url']=='https://github.com/wicanr2/logh3/releases/tag/'+version
expected={Path(row['path']).name:row for row in read(delivery/'SHA256SUMS-platforms.json')['files'] if row['path'].startswith('release/')}
expected['SHA256SUMS.txt']=dict(bytes=(delivery/'release/SHA256SUMS.txt').stat().st_size,sha256=sha(delivery/'release/SHA256SUMS.txt'))
assert set(expected)=={row['name'] for row in remote['assets']}
download=batch/'release-download'
assert {p.name for p in download.iterdir()}==set(expected)
checks=[]
for name,want in expected.items():
    p=download/name
    assert p.stat().st_size==want['bytes'] and sha(p)==want['sha256'],name
    checks.append(dict(name=name,bytes=p.stat().st_size,sha256=sha(p)))
for line in (download/'SHA256SUMS.txt').read_text().splitlines():
    digest,name=line.split('  ',1)
    assert digest==sha(download/name)
tag_commit=subprocess.check_output(['git','rev-parse',version+'^{commit}'],text=True).strip()
frozen=read(delivery/'BUILD.json')['source_files']
count=0
for name,row in frozen.items():
    if name.startswith('vendor/'):continue
    content=subprocess.check_output(['git','show',version+':'+name])
    assert hashlib.sha256(content).hexdigest()==row['sha256'],name
    count+=1
result=dict(version=version,passed=True,url=remote['url'],tag_commit=tag_commit,
            tag_nonvendor_source_files=count,assets=checks,release_json_sha256=sha(batch/'GH-RELEASE.json'),
            public_release=True,full_packages_uploaded=False,promo_uploaded=False)
(batch/'RELEASE-CHECK.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
index=root/args.index
assert index.resolve().is_relative_to((root/'docs/internals/data').resolve())
data=read(index);data['release']=result
index.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
print('PASS GitHub Release實際五附件下載／SHA-256、tag與',count,'項非vendor凍結來源一致')
