"""147：實際完整版ZIP的自動資料路徑、版本、兩種啟動器診斷抽驗。Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import zipfile

VERSION = 'v.1.5.2-20261024'
p = argparse.ArgumentParser(description=__doc__)
paths = p.add_mutually_exclusive_group(required=True)
paths.add_argument('--package')
paths.add_argument('--archive')
p.add_argument('--out', required=True)
p.add_argument('--windows', action='store_true')
p.add_argument('--version', default=VERSION)
p.add_argument('--with-posters', action='store_true')
a = p.parse_args()
VERSION = a.version
out = Path(a.out)
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
out.mkdir()


def sha(path):
    h = hashlib.sha256()
    with Path(path).open('rb') as f:
        for chunk in iter(lambda: f.read(1024*1024), b''): h.update(chunk)
    return h.hexdigest()


if a.archive:
    parent = out/'unpacked/銀河 完整版'; parent.mkdir(parents=True)
    with zipfile.ZipFile(a.archive) as z:
        assert z.testzip() is None
        for item in z.infolist():
            target = parent/item.filename
            assert target.resolve().is_relative_to(parent.resolve())
            z.extract(item, parent)
            target.chmod((item.external_attr >> 16) & 0o777)
    folder = parent/('LOGH3-'+VERSION+('-windows-amd64' if a.windows else '-linux-amd64'))
else: folder = Path(a.package)
manifest = json.loads((folder/'MANIFEST.json').read_text())
assert manifest['version'] == VERSION
for name, row in manifest['files'].items():
    assert (folder/name).stat().st_size == row['bytes'] and sha(folder/name) == row['sha256'], name
binary = folder/('gin3.exe' if a.windows else 'gin3')
env = dict(os.environ, LP_NUM_THREADS='1', GOMAXPROCS='2', GOMEMLIMIT='1600MiB', GOGC='60')


def path(value): return 'Z:'+str(value).replace('/', '\\') if a.windows else str(value)
base = ['wine', path(binary)] if a.windows else [str(binary)]
assert subprocess.check_output(base+['-version'], text=True, env=env, timeout=60).strip() == VERSION
runs = []
try:
    for name, level in [('direct-intro', None), ('original', -1), ('enhanced', 1)]:
        if level is None: command = base.copy()
        elif a.windows: command = ['wine', 'cmd', '/c', 'call', path(folder/('start-'+name+'.cmd'))]
        else: command = [str(folder/('start-'+name+'.sh'))]
        command += ['-scale', '1.5', '-save', path(out/('save-'+name)), '-shot', path(out/(name+'.png')), '-state-out', path(out/(name+'.txt')), '-frame-trace', path(out/(name+'.jsonl')), '-shot-after', '4']
        if level is not None:
            command += ['-intro=false', '-script', 'click 300 230;type 138;enter;rclick 300 240;rclick 300 240']
        with (out/(name+'.log')).open('x') as log:
            process = subprocess.run(command, cwd='/tmp', env=env, stdout=log, stderr=subprocess.STDOUT, timeout=180)
        assert process.returncode == 0, (out/(name+'.log')).read_text()
        expected = 'StateIntro' if level is None else 'StateCommandPhase'
        state = (out/(name+'.txt')).read_text(); assert 'state='+expected+'\n' in state, state
        rows = [json.loads(line) for line in (out/(name+'.jsonl')).read_text().splitlines()]
        last = rows[-1]
        assert last['graphics'] == 'hd' and last['hd_music_available'] and last['mt32_available']
        assert not last['rules']['non_battleship_flagship']
        if a.with_posters:
            assert last['poster_available'] and not last['poster_mode'] and last['poster_set'] == 'v1' and last['locale'] == 'zh-TW'
        if level is not None:
            assert last['ai_level'] == level
            assert last['music_quality'] == ('original' if level == -1 else 'hd')
        runs.append(dict(name=name, command=command, cwd='/tmp', exit=process.returncode, state=state,
                         final_frame=last, screenshot_sha256=sha(out/(name+'.png')), frame_trace_sha256=sha(out/(name+'.jsonl'))))
        print('PASS', name, expected, 'HD／GM／MT-32／預設關閉', flush=True)
    # 玩家資料在out，安裝包本體必須逐檔保持。
    for name, row in manifest['files'].items(): assert sha(folder/name) == row['sha256'], name
    result = dict(passed=True, version=VERSION, platform='Windows Wine，非真機' if a.windows else 'Linux Xvfb',
                  scope='封包直接執行及兩啟動器的診斷入口；不取代正常鍵鼠驗收或人耳確認',
                  automatic_resources=True, package=str(folder), binary_sha256=sha(binary),
                  archive_sha256=sha(a.archive) if a.archive else None, runs=runs)
    (out/'RESULT.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
finally:
    if a.windows: subprocess.run(['wineserver', '-k'], env=env, check=False, timeout=10)
