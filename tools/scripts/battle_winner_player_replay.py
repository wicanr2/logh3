"""第112輪：目前版本的晚回合勝利，獨立bootstrap及60 Hz公開存讀。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
p = argparse.ArgumentParser()
p.add_argument('revision')
p.add_argument('--input', default='workplace/winner79-r112-latea/normal-1/inputs.json')
p.add_argument('--gui', action='store_true')
a = p.parse_args()
assert a.revision.isascii() and a.revision.isalnum()
source = Path('tools/scripts/battle_occupation_player_replay.py')
parent = json.loads(Path('docs/internals/data/battle-occupation-player.json').read_text())
assert hashlib.sha256(source.read_bytes()).hexdigest() == parent['sources'][str(source)]['sha256']
s = source.read_text().replace('occupy79-r107-replay-', 'winner79-r112-replay-')
s = s.replace('workplace/occupy79-r107-s3planetb/normal-1/inputs.json', a.input)
anchor = "    plan = plan.replace('g.State() != game.StateCommandPhase || len(lifeReports) != 2',\n                        'g.State() != game.StateBattle || len(lifeReports) != 1')"
assert s.count(anchor) == 1
s = s.replace(anchor, '')
anchor = '''    step("click",240,470,0,0)
    mx,my:=r.BattleMenu.BattleMenuAt(240,470)
    step("click",mx+20,my+20+4*r.BattleMenu.Layout.Text.RowPitch,0,0)'''
assert s.count(anchor) == 1
s = s.replace(anchor, '''    step("click",544,125,0,0)
    step("click",560,165,0,0)''')
s = s.replace('snapshot("battle-owned")', 'snapshot("post-victory")')
s = s.replace('公開戰場保存未開', '公開戰後保存未開')
s = s.replace('占領後讀回完整資料不同', '戰後讀回完整資料不同')
s = s.replace("'natural_finish':False,'natural_annihilation':False,'battle_planet_capture':True,'ownership_report':True",
              "'natural_finish':True,'natural_annihilation':True,'battle_planet_capture':False,'ownership_report':False")
s = s.replace('正常主選單、S3、真實X11鍵鼠、行星換手與星域報告；RND test-only，公開存讀',
              '正常主選單、S3、真實X11鍵鼠、第九回合勝利與零占領；RND test-only，公開存讀')
generator = Path('workplace/winner79-r112-replay-' + a.revision + ('-gui-generator.py' if a.gui else '-generator.py'))
with generator.open('x') as f:
    f.write(s)
args = ['python3', str(generator), a.revision, '--input', a.input]
if a.gui:
    args.append('--gui')
subprocess.run(args, check=True)
