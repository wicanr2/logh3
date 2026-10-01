"""規格 072 的隨包工具收據合併器；由 tools/py.sh 執行，入口見驗證 032。"""
import hashlib
import json
import os
from pathlib import Path


def sha(p):
    with p.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


work = Path('/work/workplace/music69-linux')
package = work/'封裝 測試'
build = json.loads((package/'BUILD.json').read_text())
check = json.loads((work/'CHECK.json').read_text())
assert check['status'] == 'complete-linux-tools-sampling'
assert check['build_sha256'] == sha(package/'BUILD.json')
assert len(check['fm']['tracks']) == 3 and len(check['gm']['tracks']) == 1
assert check['gm_full_pcm_sha256']['new'] == check['gm_full_pcm_sha256']['reference']
for name, entry in build['files'].items():
    p = package/name
    assert p.stat().st_size == entry['bytes'] and sha(p) == entry['sha256'], name
    assert not p.is_symlink() and (p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid()), name
for name, expected in check['raw_outputs'].items():
    assert sha(work/'imported'/name) == expected
test_log = work/'player-test.log'
assert '--- PASS: TestBundledLinuxMusicImportPlayerDecode' in test_log.read_text()
source_probe = json.loads((work/'source-index/probe.json').read_text())
for entry in source_probe.values():
    assert sha(work/'source-index'/Path(entry['url']).name.replace('%2B', '+')) == entry['sha256']
doc = {'schema': 'logh3.music-tools.linux.validation.v1', 'date': '2026-10-01', 'status': 'Linux 隨包工具技術抽驗通過；正式來源碼交付／其他平台／人耳待驗', 'build': build, 'check': check, 'validation': {'clean_image': 'sha256:229a2c5bfa27522db7815ea81f9bed70af17ccb9de9fc7ad142b1877b5830d36', 'gui_image': 'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8', 'check_sha256': sha(work/'CHECK.json'), 'player_test_log_sha256': sha(test_log), 'prototype_result_sha256': sha(work/'probe-result.json'), 'source_descriptions': source_probe, 'source_packages_total': len({(x['source_package'], x['source_version']) for x in build['packages'].values()}), 'package_bytes': sum(x['bytes'] for x in build['files'].values()), 'full_source_archives_delivered': False, 'limits': '只驗當輪 Linux amd64／三種 FM 路徑與一首 GM。沒有發佈正式封包；原版來源／生成曲庫不在工具包。來源身份與完整授權已保存，157 個來源套件的完整 archive 交付仍須補齊。'}}
doc['validation']['go_image'] = 'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac'
doc['validation']['launcher_buildinfo'] = (work/'launcher-buildinfo.txt').read_text()
doc['validation']['verification_scripts_sha256'] = {n: sha(Path('/work/tools/scripts')/n) for n in ['music_linux_check.py', 'music_linux_receipt.py', 'music_linux_source_probe.py']}
target = Path('/work/docs/internals/data/linux-music-tools.json')
assert target.parent.stat().st_uid == os.getuid()
if target.exists():
    assert not target.is_symlink() and (target.stat().st_uid, target.stat().st_gid) == (os.getuid(), os.getgid())
target.write_text(json.dumps(doc, ensure_ascii=False, indent=2)+'\n')
print('Linux 隨包工具收據 SHA-256：', sha(target))
