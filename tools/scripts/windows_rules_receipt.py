"""核對第138輪最終來源、實際ZIP、正常GUI及本機交付，產生唯一收據索引。"""
import hashlib
import json
from pathlib import Path
import shutil
import subprocess

ROOT=Path('/src')
VERSION='v.1.4.0-20261015'
OUTPUT=ROOT/'dist-all'/VERSION


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def record(path):
    return dict(path=str(path.relative_to(ROOT)),bytes=path.stat().st_size,sha256=sha(path))


def read(path):
    return json.loads(path.read_text())


build=read(OUTPUT/'BUILD.json')
assert build['version']==VERSION
verified_source=[]
for name,expected in build['source_files'].items():
    if not name.startswith('vendor/'):
        assert sha(ROOT/name)==expected['sha256'],name
        verified_source.append(record(ROOT/name))
subprocess.run(['git','diff','--exit-code','HEAD','--','cmd','internal'],check=True)
packages=read(OUTPUT/'SHA256SUMS.json')['files']
for row in packages:
    p=OUTPUT/row['path']
    assert p.stat().st_size==row['bytes'] and sha(p)==row['sha256'],p
tests=ROOT/'workplace/rules138-tests-final.log'
lines=tests.read_text().splitlines()
assert sum(l.startswith('ok ') for l in lines)==15
assert sum('[no test files]' in l for l in lines)==4
assert not any('FAIL' in l for l in lines)
gui=[]
paths=[ROOT/'workplace/rules138-gui-f0',ROOT/'workplace/rules138-gui-f1',ROOT/'workplace/music69-windows/rules138-wine-f0-r5']
for folder in paths:
    d=read(folder/'RESULT.json')
    assert d['passed'] and d['normal_entry'] and not d['script_flag'] and not d['state_injection'] and not d['seed_injection']
    pictures=[]
    for event in d['events']:
        if 'screenshot' in event:
            p=folder/event['screenshot']
            assert sha(p)==event['sha256']
            pictures.append(dict(file=record(p),observed=event['observed']))
    assert len(pictures)==5
    gui.append(dict(receipt=record(folder/'RESULT.json'),platform=d['platform'],faction=d['faction'],
                    binary_sha256=d['binary_sha256'],screenshots=pictures,
                    checks=[e['check'] for e in d['events'] if 'check' in e]))
delivery=ROOT/'workplace/music69-windows/windows138-delivery/RESULT.json'
windows=read(delivery)
assert len(windows['runs'])==8 and windows['version']==VERSION
assert windows['archives']['full']['sha256']==packages[0]['sha256']
assert windows['archives']['engine']['sha256']==packages[1]['sha256']
binary=OUTPUT/'full'/f'LOGH3-{VERSION}-windows-amd64/gin3.exe'
assert gui[2]['binary_sha256']==sha(binary)
target=ROOT/'docs/images/readme-v140-windows-settings.png'
shutil.copyfile(paths[2]/'01-default-on.png',target)
smoke=OUTPUT/'smoke'
for name,p in [('WINDOWS-ZIP.json',delivery),('WINDOWS-SETTINGS.json',paths[2]/'RESULT.json'),('TESTS.log',tests)]:
    shutil.copyfile(p,smoke/name)
    assert sha(p)==sha(smoke/name)
for p in paths[2].glob('*.png'):
    shutil.copyfile(p,smoke/('settings-'+p.name))
extra=[ROOT/'workplace/music69-windows/windows138-paths/RESULT.json',
       ROOT/'workplace/music69-windows/windows138-geometry/RESULT.json',tests,delivery,
       OUTPUT/'BUILD.json',OUTPUT/'SHA256SUMS.json']
for p in [OUTPUT,smoke,target,binary]+[OUTPUT/row['path'] for row in packages]:
    assert p.stat().st_uid==p.stat().st_gid==1000,p
inputs=[ROOT/p for p in ['cmd/gin3/resourcepaths.go','cmd/gin3/rule_settings.go','cmd/gin3/rule_settings_test.go',
    'internal/game/flagshiptransfer.go','internal/game/flagshiptransfer_test.go','internal/game/ruleoptions.go',
    'internal/rules/flagshiptransfer.go','internal/rules/flagshiptransfer_test.go','tools/scripts/windows_hotfix_package.py',
    'tools/scripts/windows_delivery_smoke.py','tools/scripts/windows_resource_paths.py','tools/scripts/rule_settings_gui.py',
    'tools/scripts/windows_rules_receipt.py','docs/spec/146-flagship-transfer-and-rule-settings.md']]
result=dict(schema='logh3.windows-rule-settings.v1',date='2026-10-15',version=VERSION,
    verified_product_source_commit=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
    build_head=build['head'],build_working_tree=build['working_tree'],
    frozen_product_sources_verified=len(verified_source),product_sources=verified_source,
    tools=dict(build_image='logh3-hd-preview:20261011-r1',go_version=build['go_version'],
               zip_wine_image='kol-wine-verify:ubuntu-noble-20261007-r1',
               gui_wine_image='eob-remake-wine-verify:ubuntu-noble-20261008-r2'),
    packages=packages,binary=record(binary),gui=gui,evidence=[record(p) for p in extra],
    inputs=[record(p) for p in inputs],readme_image=record(target),
    limits=['使用者指定的remake擴充，不是DOS原版parity','完整傷害為受控S3測試，未另驗自然遊玩觸發',
            'Windows Wine通過；Windows真機未驗','本輪僅Windows二進位；Linux GUI已驗，未重包Linux／macOS',
            '#48的正式戰術讀回重建與DOS整段同狀態仍OPEN'],
    rights='完整包、原版與衍生音訊及HD保留本機；引擎／來源包排除原版素材；未發布tag或Release')
index=ROOT/'docs/internals/data/windows-rule-settings.json'
index.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('PASS 最終來源、三份正常GUI、8案ZIP與交付',record(index))
