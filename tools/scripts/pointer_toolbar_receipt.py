"""152：獨立核對更新ZIP、凍結來源、GUI截圖及存檔。Docker內執行。"""
import argparse
import ast
import hashlib
import json
import os
from pathlib import Path
import struct
import zipfile
from PIL import Image, ImageChops

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--check', action='store_true')
a = p.parse_args()
root = Path('/src')
version = 'v.1.6.1-20261025'
delivery = root/'dist-all'/version
assert os.getuid() == os.getgid() == 1000
sha = lambda f: hashlib.sha256(f.read_bytes()).hexdigest()
read = lambda f: json.loads(f.read_text())
relative = lambda f: str(f.relative_to(root))
build_file = delivery/'BUILD.json'; build = read(build_file)
assert build['version'] == version
linux_build_file=root/'dist-all/v.1.6.1-20261024/BUILD.json'
linux_build=read(linux_build_file)
assert linux_build['source_files']==build['source_files']
source = root/build['source']
for name, digest in build['source_files'].items():
    assert sha(source/name) == digest, name
    if not name.startswith('vendor/'): assert sha(root/name) == digest, name
archive = delivery/build['archive']['path']
assert archive.stat().st_size == build['archive']['bytes'] and sha(archive) == build['archive']['sha256']
with zipfile.ZipFile(archive) as z:
    assert z.testzip() is None
    stem = archive.stem
    manifest = read(delivery/'release'/stem/'MANIFEST.json')
    assert manifest['version'] == version and manifest['files'] == build['files']
    assert set(z.namelist()) == {stem+'/'+n for n in [*manifest['files'],'MANIFEST.json']}
    for name,row in manifest['files'].items():
        content = z.read(stem+'/'+name)
        assert len(content) == row['bytes'] and hashlib.sha256(content).hexdigest() == row['sha256']
        assert not name.startswith(('original/','art/','music/','soundfont/'))
    for info in z.infolist():
        if not info.filename.isascii(): assert info.flag_bits & 0x800
    text = z.read(stem+'/README.txt')
    assert text.startswith(b'\xef\xbb\xbf') and b'\n' not in text.replace(b'\r\n',b'')
    exe = z.read(stem+'/gin3.exe')
    pe = struct.unpack_from('<I',exe,0x3c)[0]
    assert exe[:2] == b'MZ' and exe[pe:pe+4] == b'PE\0\0' and struct.unpack_from('<H',exe,pe+4)[0] == 0x8664
exe_file = root/'workplace/music69-windows/pointer152'/version/'installed'/stem/'gin3.exe'
assert sha(exe_file) == hashlib.sha256(exe).hexdigest() == build['commands'][0]['sha256']
version_file = root/'workplace/music69-windows/pointer152'/version/'VERSION.txt'
assert version_file.read_text().strip() == version
standard_file = root/'workplace/pointer152/standard-final.log'
standard = standard_file.read_text().splitlines()
assert standard and not any('FAIL' in v for v in standard)
passed = sum(v.startswith('ok ') for v in standard)
untested = sum(v.startswith('?') for v in standard)
assert passed == 16 and untested == 4
driver = root/'workplace/pointer152/driver-final.py'
driver2 = root/'workplace/pointer152/driver-final-v2.py'
ast.parse(driver.read_text()); ast.parse((root/'tools/scripts/pointer_toolbar_gui.py').read_text())
runs = {}
for name, directory in [
    ('linux-alliance',root/'workplace/pointer152/final-linux'),
    ('windows-alliance',root/'workplace/music69-windows/pointer152'/version/'final-windows'),
    ('linux-empire',root/'workplace/pointer152/final-empire-fixed')]:
    result_file = directory/'RESULT.json'; row = read(result_file)
    assert row['passed'] and row['normal_entry']
    assert not any(row[k] for k in ['state_injection','seed_injection','script_flag'])
    assert row['driver_sha256'] == sha(driver2 if name=='linux-empire' else driver)
    expected = build['commands'][0]['sha256'] if name.startswith('windows') else linux_build['commands'][1]['sha256']
    assert row['binary_sha256'] == expected
    perf = row['performance']
    assert perf['samples'] == 24 and perf['uploads_before'] == perf['uploads_after']
    assert len(row['events']) == 31 and len(row['saved']) == 6
    screenshots = []
    for event in row['events']:
        f=directory/event['screenshot']
        assert sha(f) == event['sha256']
        with Image.open(f) as im:
            assert im.size == tuple(event['window']); im.verify()
        screenshots.append(dict(name=event['name'],path=relative(f),sha256=sha(f)))
    for code in ['zh-TW','zh-CN','ja','en','ko']:
        for prefix in ['font','ai']:
            event=next(v for v in row['events'] if v['name']==prefix+'-'+code)
            assert event['observed']['locale']==code and event['observed']['toolbar_menu']==prefix
    for n,h in row['saved'].items(): assert sha(directory/'save'/n)==h
    runs[name]=dict(path=relative(directory),result_sha256=sha(result_file),binary_sha256=expected,
        performance=perf,saved=row['saved'],screenshots=screenshots)
baseline_dir=root/'workplace/pointer152/baseline-linux-r2'
baseline=read(baseline_dir/'RESULT.json'); assert baseline['passed']
assert baseline['saved']==runs['linux-alliance']['saved']==runs['windows-alliance']['saved']
# 不含游標的固定小區域，使用未縮放完整視窗原圖計算。
rois={'button':(565,122,607,138),'menu':(565,162,590,178)}
hover={}
for name,directory in [('baseline',baseline_dir)]+[(n,root/v['path']) for n,v in runs.items()]:
    counts={}
    for mode in ['hd'] if name=='baseline' else ['hd','original']:
        for control,logical in rois.items():
            box=tuple(int(v*1.5) for v in logical)
            off=('map-off' if control=='button' else 'menu-off') if mode=='hd' else ('original-off' if control=='button' else 'original-menu-off')
            on=control+'-hover' if mode=='hd' else 'original-'+control+'-hover'
            i1=Image.open(directory/(off+'.png')).convert('RGB').crop(box)
            i2=Image.open(directory/(on+'.png')).convert('RGB').crop(box)
            counts[mode+'-'+control]=sum(v!=(0,0,0) for v in ImageChops.difference(i1,i2).getdata())
    assert all(v==0 for v in counts.values()) if name=='baseline' else all(v>0 for v in counts.values())
    hover[name]=counts
centers={}
for name,run in runs.items():
    directory=root/run['path']
    normal=Image.open(directory/'map-sans.png').convert('RGB')
    poster=Image.open(directory/'poster-map.png').convert('RGB')
    assert normal.size==(960,720) and poster.size==(1280,720)
    delta=ImageChops.difference(normal,poster.crop((160,0,1120,720)))
    assert delta.getbbox() is None
    centers[name]=dict(pixels=960*720,different=0,poster_crop=[160,0,1120,720])
data=dict(version=version,scope='滑鼠顯示回應、hover及工具列字型／AI。remake正常玩家路徑，不新增DOS parity。',
    archive=build['archive'],build_sha256=sha(build_file),binary_sha256=sha(exe_file),
    source_files=len(build['source_files']),source_nonvendor=sum(not n.startswith('vendor/') for n in build['source_files']),
    source=build['source'],source_head=build['head'],
    linux_validation_build=dict(path=relative(linux_build_file),sha256=sha(linux_build_file),version=linux_build['version'],source_identical=True),
    standards=dict(path=relative(standard_file),sha256=sha(standard_file),passed=passed,no_tests=untested),
    drivers=[dict(path=relative(f),sha256=sha(f)) for f in [driver,driver2]],
    version_output=dict(path=relative(version_file),sha256=sha(version_file)),
    gui=runs,baseline=dict(path=relative(baseline_dir),result_sha256=sha(baseline_dir/'RESULT.json'),performance=baseline['performance']),
    hover=dict(logical_rois={k:list(v) for k,v in rois.items()},window_scale=1.5,changed_pixels=hover),poster_centers=centers,
    limits=['Windows採Wine，未作Windows真機或硬體端到端延遲驗收。',
            'GUI收據latency_ms從Python啟動xdotool前計時，包含程序啟動與排程。原scope文字過強，不能作輸入延遲改善證據。',
            'Draw耗時是隔離環境抽樣，受排程與負載影響；正式樣本以本索引gui及baseline欄位為準。',
            '字型與AI沿既有本次執行設定，不新增永久偏好；原版顯示固定原字形。',
            '設定頁沿專用黑體，工具列字型影響其他HD遊戲文字。',
            '未發布GitHub Release，先前v1.6.0三平台包及影片保持。'])
target=root/'docs/internals/data/pointer-toolbar152.json'
if a.check:
    assert read(target)==data
else:
    assert target.parent.stat().st_uid==1000
    target.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(dict(passed=True,archive=data['archive'],standards=data['standards'],gui={n:len(v['screenshots']) for n,v in runs.items()},hover=hover),ensure_ascii=False,indent=2))
