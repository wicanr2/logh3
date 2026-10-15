"""從實際 ZIP 解壓驗證 Windows 完整／引擎包及中文空格路徑。"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import zipfile
import argparse

VERSION='v.1.4.0-20261015'
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--root',default='/out/windows138-delivery')
ROOT=Path(p.parse_args().root)
OUTPUT=ROOT/'unpacked/銀河 修正版'
assert ROOT.is_dir() and not OUTPUT.exists() and os.getuid()==os.getgid()==1000
OUTPUT.mkdir(parents=True)
os.environ['LP_NUM_THREADS']='1'


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def win(path):
    return 'Z:'+str(path).replace('/','\\')


packages={}
for kind in ['full','engine']:
    with zipfile.ZipFile(ROOT/(kind+'.zip')) as z:
        assert z.testzip() is None
        z.extractall(OUTPUT/kind)
    folder=OUTPUT/kind/f'LOGH3-{VERSION}-windows-amd64'
    manifest=json.loads((folder/'MANIFEST.json').read_text())
    assert manifest['version']==VERSION
    for name,info in manifest['files'].items():
        assert (folder/name).stat().st_size==info['bytes'] and sha(folder/name)==info['sha256'],name
    query=subprocess.check_output(['wine',win(folder/'gin3.exe'),'-version'],text=True,timeout=45)
    assert query.strip()==VERSION
    packages[kind]=folder


runs=[]
def run(name,kind,args,state=None,launcher=None,missing=False):
    folder=packages[kind]
    prefix=ROOT/name
    command=['wine',win(folder/'gin3.exe')] if launcher is None else ['wine','cmd','/c','call',win(folder/launcher)]
    command+=['-save',win(ROOT/'save'),'-shot',win(prefix.with_suffix('.png')),
              '-state-out',win(prefix.with_suffix('.txt')),'-shot-after','4']+args
    with prefix.with_suffix('.log').open('x') as log:
        r=subprocess.run(command,cwd='/tmp',stdout=log,stderr=subprocess.STDOUT,timeout=90)
    text=prefix.with_suffix('.log').read_text()
    if missing:
        assert r.returncode==1 and 'original/GRAPH.IMG' in text.replace('\\','/'),text
        observed=None
    else:
        assert r.returncode==0,text
        observed=prefix.with_suffix('.txt').read_text()
        assert 'state='+state+'\n' in observed,observed
    runs.append(dict(name=name,command=command,cwd='/tmp',exit=r.returncode,state=observed,
                     binary_sha256=sha(folder/'gin3.exe'),
                     screenshot_sha256=None if missing else sha(prefix.with_suffix('.png'))))
    print('PASS',name,flush=True)


run('full-direct-intro','full',[],state='StateIntro')
run('full-direct-menu','full',['-script','move 300 240'],state='StateMainMenu')
for mode in ['original','enhanced']:
    run('full-'+mode,'full',['-script','click 300 210;type 138;click 432 270;click 448 350'],state='StateBudgetPlan',launcher='start-'+mode+'.cmd')
run('engine-missing-original','engine',['-script','move 300 240'],missing=True)
external=['-orig',win(packages['full']/'original')]
run('engine-direct-import','engine',external+['-script','move 300 240'],state='StateMainMenu')
for mode in ['original','enhanced']:
    run('engine-'+mode,'engine',external+['-script','click 300 210;type 138;click 432 270;click 448 350'],state='StateBudgetPlan',launcher='start-'+mode+'.cmd')
(ROOT/'RESULT.json').write_text(json.dumps(dict(version=VERSION,platform='Windows Wine，非真機',runs=runs,
    archives={kind:dict(sha256=sha(ROOT/(kind+'.zip')),files=len(json.loads((folder/'MANIFEST.json').read_text())['files'])) for kind,folder in packages.items()}),ensure_ascii=False,indent=2)+'\n')
