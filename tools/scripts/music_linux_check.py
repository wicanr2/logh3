"""規格 072 的乾淨容器驗收；只由 tools/py.sh 執行。

來源 /orig 唯讀，PATH 無效；僅呼叫包內原生程式。保存原始命令與實際結果，
FM 仍走 071 全檔閘門，GM 比完整 f32le PCM，不將平台啟動等同人耳驗收。
"""
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess

work = Path('/work/workplace/music69-linux')
stage = work / 'stage'
package = work / '封裝 測試'
if stage.exists():
    assert stage.is_dir() and (stage/'BUILD.json').is_file()
    if package.exists():
        # 前次只有可丟棄探針，不刪已完成正式收據或任意目錄。
        assert not (package/'BUILD.json').exists()
        assert package.stat().st_uid == os.getuid()
        shutil.rmtree(package)
    stage.rename(package)
else:
    # 驗證腳本前置檢查修正後重跑同一包；不覆蓋完成收據或再次匯入。
    assert (package/'BUILD.json').is_file() and not (work/'check-start.json').exists()
build = json.loads((package/'BUILD.json').read_text())
assert not Path('/usr/bin/ffmpeg').exists() and not Path('/usr/bin/fluidsynth').exists()
env = {'PATH': '/none', 'HOME': '/tmp', 'LC_ALL': 'C', 'LD_PRELOAD': '/definitely-missing.so', 'LD_LIBRARY_PATH': '/wrong', 'GLIBC_TUNABLES': 'glibc.cpu.hwcaps=-AVX2'}
tools = package/'tools'
checks = {}


def sha(p):
    with p.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


for name, entry in build['files'].items():
    p = package/name
    assert p.is_file() and not p.is_symlink() and p.stat().st_size == entry['bytes'] and sha(p) == entry['sha256'], name
    assert (p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid()), name
for name, args, expected in [('ffmpeg', ['-version'], 0), ('fluidsynth', ['--version'], 0), ('opl2wav', [], 2)]:
    cmd = [str(tools/name)] + args
    run = subprocess.run(cmd, env=env, capture_output=True, text=True)
    assert run.returncode == expected, (name, run.stderr)
    loader_cmd = [str(tools/'lib/ld-linux-x86-64.so.2'), '--inhibit-cache', '--library-path', '$ORIGIN/../lib', '--list', str(tools/'real'/name)]
    # 載入器解析檢查不用測試注入的 LD_*，避免這些變數先干預載入器本身。
    listing = subprocess.check_output(loader_cmd, env={'PATH': '/none', 'HOME': '/tmp', 'LC_ALL': 'C'}, text=True)
    paths = []
    for line in listing.splitlines():
        resolved = line.split('=>', 1)[-1].split('(', 1)[0].strip()
        if resolved.startswith('/'):
            paths.append(resolved)
    outside = [x for x in paths if not Path(x).resolve().is_relative_to(tools.resolve())]
    assert paths and not outside, (name, outside)
    checks[name] = {'command': cmd, 'exit_code': run.returncode, 'text': run.stdout+run.stderr, 'loader_command': loader_cmd, 'resolved_dependencies': listing}
    print(name, '搬移／原生入口與包內依賴解析通過', flush=True)
lib = tools/'lib/libc.so.6'
parked = lib.with_suffix('.parked')
lib.rename(parked)
try:
    run = subprocess.run([str(tools/'ffmpeg'), '-version'], env=env, capture_output=True, text=True)
    assert run.returncode == 1 and '缺隨包普通檔案' in run.stderr and 'libc.so.6' in run.stderr
    checks['missing-library'] = {'exit_code': run.returncode, 'text': run.stderr}
finally:
    parked.rename(lib)
unknown = tools/'unknown'
shutil.copyfile(tools/'ffmpeg', unknown)
unknown.chmod(0o755)
try:
    run = subprocess.run([str(unknown)], env=env, capture_output=True, text=True)
    assert run.returncode == 1 and '未知隨包工具' in run.stderr
    checks['unknown-tool'] = {'exit_code': run.returncode, 'text': run.stderr}
finally:
    unknown.unlink()
out = work/'imported'
assert not out.exists(), '不覆蓋既有對拍收據'
cmd = [str(package/'music-import'), '-source', '/orig', '-output', str(out), '-tracks', 'BGM_0M,BGM_0F,END_2F,OPEN_F']
(work/'check-start.json').write_text(json.dumps({'status': 'in-progress', 'command': cmd, 'build_sha256': sha(package/'BUILD.json'), 'environment': env}, ensure_ascii=False, indent=2)+'\n')
subprocess.run(cmd, env=env, check=True, timeout=20*60)
fm = json.loads((out/'FM-SOURCES.json').read_text())
gm = json.loads((out/'GM-SOURCES.json').read_text())
assert fm['status'] == 'complete-fm-stage' and len(fm['tracks']) == 3
assert gm['status'] == 'complete-gm-stage' and len(gm['tracks']) == 1
reference = Path('/work/workplace/music69-fm/full')
for track in fm['tracks']:
    name = track['track']
    for suffix in ('.vgm', '.ogg'):
        assert (out/(name+suffix)).read_bytes() == (reference/(name+suffix)).read_bytes(), (name, suffix)
pcm_hashes = {}
for n, source in [('new', out/'BGM_0M.ogg'), ('reference', reference/'BGM_0M.ogg')]:
    pcm = work/(n+'.f32le')
    pcm_cmd = [str(tools/'ffmpeg'), '-v', 'error', '-y', '-i', str(source), '-f', 'f32le', '-acodec', 'pcm_f32le', str(pcm)]
    subprocess.run(pcm_cmd, env=env, check=True)
    pcm_hashes[n] = sha(pcm)
    pcm.unlink()
assert pcm_hashes['new'] == pcm_hashes['reference']
receipt = {'status': 'complete-linux-tools-sampling', 'command': cmd, 'build_sha256': sha(package/'BUILD.json'), 'environment': env, 'launcher_checks': checks, 'fm': fm, 'gm': gm, 'gm_full_pcm_sha256': pcm_hashes, 'raw_outputs': {p.name: sha(p) for p in out.iterdir() if p.is_file()}, 'limits': 'Linux 容器／三種 FM 路徑與一首 GM 抽驗；其他平台、正式來源碼交付與人耳／loop 未驗收。'}
(work/'CHECK.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
print('三種 FM VGM／OGG 全檔一致，GM 完整 PCM 一致；Linux 隨包入口抽驗通過。', flush=True)
