"""規格 077：用來源重建程式及已驗 Linux 隨包工具，正常匯入一首 FM。

入口：tools/py.sh /work/tools/scripts/music_source_import_check.py。
不覆寫既有收據；原版 /orig 唯讀，容器無網路，沒有系統音樂工具。
"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

import music_source_kit as kit

ROOT = Path('/work/workplace/music69-delivery')
BASE = Path('/work/workplace/music69-linux/封裝 測試')
OUT = ROOT/'source-import-vcs-root'
BUILD = ROOT/'rebuild-vcs-root'


def sha(p):
    kit.owned(p)
    with p.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


kit.parent_owned(OUT)
assert not OUT.exists() and not OUT.is_symlink(), '拒絕覆寫收據'
assert not Path('/usr/bin/ffmpeg').exists()
base_manifest = json.loads((BASE/'BUILD.json').read_text())
for name, row in base_manifest['files'].items():
    p = BASE/name
    assert sha(p) == row['sha256'] and p.stat().st_size == row['bytes'], name
for platform in ('linux', 'windows'):
    name = 'dosgolem-probe-'+platform
    assert sha(BUILD/(name+'-source')) == sha(ROOT/(name+'-vcs-reference')), name
OUT.mkdir()
package = OUT/'重建 工具'
package.mkdir()
shutil.copytree(BASE/'tools', package/'tools')
for name in ('ffmpeg', 'fluidsynth', 'opl2wav'):
    shutil.copyfile(BUILD/'music-tool-linux-source', package/'tools'/name)
    (package/'tools'/name).chmod(0o755)
for source, target in [(BUILD/'music-import-linux-source', package/'music-import'),
                       (BUILD/'dosgolem-probe-linux-source', package/'tools/dosgolem-probe')]:
    shutil.copyfile(source, target)
    target.chmod(0o755)
env = {'PATH': '/none', 'HOME': '/tmp', 'LC_ALL': 'C'}
destination = OUT/'匯入 曲庫'
command = [str(package/'music-import'), '-source', '/orig', '-output', str(destination),
           '-fm-only', '-tracks', 'BGM_0F']
files = {p.relative_to(package).as_posix(): sha(p) for p in sorted(package.rglob('*')) if p.is_file()}
start = {'command': command, 'environment': env, 'base_build_sha256': sha(BASE/'BUILD.json'),
         'tools_sha256': files, 'source_archive_sha256': sha(ROOT/'source-kit-vcs-root/music-source-kit.tar')}
(OUT/'START.json').write_text(json.dumps(start, ensure_ascii=False, indent=2)+'\n')
with (OUT/'import.log').open('w') as log:
    subprocess.run(command, env=env, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=20*60)
fm = json.loads((destination/'FM-SOURCES.json').read_text())
assert fm['status'] == 'complete-fm-stage' and len(fm['tracks']) == 1
track = fm['tracks'][0]
assert track['track'] == 'BGM_0F'
locks = json.loads((Path('/work/docs/internals/data/fm-import-lock.json')).read_text())
lock = locks['tracks']['BGM_0F']
for key in ('vgm_sha256', 'wav_sha256', 'pcm_sha256'):
    assert track[key] == lock[key], key
assert sha(destination/'BGM_0F.vgm') == lock['vgm_sha256']
assert sha(destination/'BGM_0F.ogg') == lock['ogg_sha256']
record = {**start, 'status': 'complete-source-built-fm-sample', 'fm': fm,
          'raw_outputs_sha256': {p.name: sha(p) for p in sorted(destination.iterdir()) if p.is_file()},
          'import_log_sha256': sha(OUT/'import.log'),
          'limits': '來源重建正式 CLI 的一首 FM 技術抽驗；非人耳、實機或完整正式發行驗收。'}
(OUT/'CHECK.json').write_text(json.dumps(record, ensure_ascii=False, indent=2)+'\n')
print('來源重建正式 FM 匯入通過；VGM／WAV／增益後 PCM 相同，Linux OGG 全檔亦符合既有鎖。')
