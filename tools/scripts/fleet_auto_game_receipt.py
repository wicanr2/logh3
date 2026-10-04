"""第75輪實作驗證索引；不改原版凍結來源或舊輪收據。"""
import hashlib
import json
import os
from pathlib import Path
import re
import sys

root = Path('/src') if Path('/src/internal').is_dir() else Path('/work')
target = root/sys.argv[1]
assert not target.exists() and target.parent.stat().st_uid == os.getuid()
sources = {}

def read(p):
    b = (root/p).read_bytes()
    assert (root/p).stat().st_uid == os.getuid()
    sources[p] = {'sha256': hashlib.sha256(b).hexdigest(), 'size': len(b)}
    return b

native = read('docs/internals/data/fleet-auto-command.json')
assert hashlib.sha256(native).hexdigest() == '23627c23a9d2ef93991c842e00fbf43df75f00d2043c78788454e868d3a19401'
manifest = json.loads(native)
assert manifest['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
assert len(manifest['cases']) == 31
for p, v in manifest['files'].items():
    assert hashlib.sha256(read(p)).hexdigest() == v['sha256'], p

for p in ['internal/rules/fleetauto.go', 'internal/rules/fleetauto_test.go',
          'internal/game/fleetcombatcommand.go', 'internal/game/fleetauto_test.go',
          'tools/combat_player_command_oracle.sh', 'tools/scripts/combat_player_command_plan.py',
          'tools/scripts/combat_player_command_ida.py', 'tools/scripts/combat_auto_receipt.py',
          'tools/scripts/fleet_auto_game_receipt.py', 'docs/spec/101-fleet-auto-command.md']:
    read(p)
log = read('workplace/round75-auto-final.log').decode()
assert '31案，115筆' in log
assert log.count('=== RUN   TestFleetAutoCommandGameAndSave/') == 21
for name in ['TestFleetAutoCommandMatchesNative', 'TestFleetAutoCommandRejectsBeforeRandom',
             'TestFleetAutoCommandGameAndSave', 'TestFleetAutoCommandVersusHumanControl',
             'TestFleetAutoCommandGameRejectsWithoutMutation']:
    assert re.search(r'^--- PASS: '+name+r' \(', log, re.M), name
assert 'FAIL' not in log and 'SKIP' not in log
standard = read('workplace/round75-standard.log').decode()
assert len(re.findall(r'^ok\s', standard, re.M)) == 13
assert len(re.findall(r'^\?\s', standard, re.M)) == 5
assert 'FAIL' not in standard
images = read('workplace/round75-images.txt').decode()
for digest in ['1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac',
               '6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
               '229a2c5bfa27522db7815ea81f9bed70af17ccb9de9fc7ad142b1877b5830d36']:
    assert digest in images
data = {'schema': 'logh3.fleet-auto-game.v1',
        'status': '101 CONFORMED, limited auto-command data; 092 DRAFT',
        'remake_base_commit': '529b61a5deb64e7affbb8b66f43b70b069d1e7a2',
        'dosgolem_head': manifest['dosgolem_head'],
        'native_receipt_sha256': hashlib.sha256(native).hexdigest(),
        'images': images,
        'original_go_image_tag': 'golang:1.24-bookworm',
        'ida_image_id': '6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780',
        'rules_cases': 31, 'native_forecasts': 115, 'game_save_cases': 21,
        'rules_rejection_cases': 12, 'game_rejection_cases': 6,
        'product_packages_passed': 13, 'product_packages_without_tests': 5,
        'limitations': ['資料介面未由正常UI／第④步呼叫',
                        '十案非正式池／編制身分只驗局部原版規則；另八案補真正一般池',
                        '原版無beam的武器殘留及word回繞保持；不擴大手動候選',
                        '061被動CPU、核准零距離及存檔格式保持',
                        '完整092／056、正常玩家／GUI／音效仍未完成'],
        'sources': sources}
target.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
print(len(sources), '份完整來源，SHA-256', hashlib.sha256(target.read_bytes()).hexdigest())
