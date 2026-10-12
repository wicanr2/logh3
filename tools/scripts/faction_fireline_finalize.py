"""凍結114同盟B改選的模型、正常視窗、核准概念、瀏覽器及來源。"""
import hashlib
import json
import os
from pathlib import Path
import shutil

assert Path.cwd()==Path('/src') and os.getuid()==1000
root=Path('workplace/hd126');files=[];windows=pixels=0;attempts=[]
for faction in ('empire','alliance'):
 folder=root/f'theme-x11-{faction}-r7';trace=json.loads((folder/'TRACE.json').read_text())
 assert trace['normal_input'] and not trace['script_flag'] and trace['scale']==1.5
 rows=[f for f in folder.glob('*-x11-comparison.json') if '-attempt-' not in f.name]
 assert len(rows)==34
 for f in rows:
  row=json.loads(f.read_text());assert row['different']==row['max_channel_difference']==0 and row['pixels']==691200
  windows+=1;pixels+=row['pixels']
 for f in folder.glob('*-attempt-*-x11-comparison.json'):
  attempts.append(dict(path=str(f),different=json.loads(f.read_text())['different']))
 files.extend(f for f in folder.rglob('*') if f.is_file())
 files.extend(f for f in (root/f'theme-{faction}-r7').rglob('*') if f.is_file())
approved=json.loads((root/'APPROVED-UI-COMPARISON-r7.json').read_text());assert len(approved['comparisons'])==8 and all(r['different']==0 for r in approved['comparisons'])
browser=json.loads((root/'FACTION-GAME-BROWSER-RESULTS-r7.json').read_text());assert len(browser)==2 and all(r['pass'] and r['checks']==34 for r in browser)
tests=(root/'faction-standard-tests-r7.log').read_text();assert sum(x.startswith('ok') for x in tests.splitlines())==15 and sum(x.startswith('?') for x in tests.splitlines())==4 and 'FAIL' not in tests
files.extend([root/n for n in ('FACTION-HD-CONCEPTS.html','FACTION-HD-GAME.html','FACTION-GAME-BROWSER-RESULTS-r7.json','APPROVED-UI-COMPARISON-r7.json','faction-standard-tests-r7.log')])
files.extend(root.glob('approved-*-difference-r7.png'));files.extend(root.glob('faction-game-*-r7.png'))
source_names=['internal/gfx/framebuffer.go','internal/game/game.go','internal/game/enemyvisibility.go','internal/screen/maingame.go','internal/ui/bevel.go',
 'internal/presentation/artwork.go','internal/presentation/classic.go','internal/presentation/faction.go','internal/presentation/journal.go','internal/presentation/symbols.go',
 'internal/presentation/faction_test.go','internal/bootstrap/hd_journal_test.go','tools/scripts/hd_art_model.go','tools/scripts/hd_art_gui.py',
 'tools/scripts/faction_hd_check.py','tools/scripts/faction_hd_game_browser_check.py','tools/scripts/faction_hd_concepts.py','tools/scripts/faction_fireline_finalize.py',
 'assets/hd/interface-themes.json','assets/hd/interface-alliance-approved.png','assets/hd/interface-empire-approved.png','docs/spec/135-original-style-hd-ui.md',
 'docs/playtest/114-alliance-fireline.md','docs/internals/data/faction-themes.json','tools/audio.sh','tools/gui.sh','tools/s1_record_display.sh']
backup=root/'faction114-source';assert not backup.exists();backup.mkdir()
for name in source_names+['workplace/hd126/FACTION-HD-GAME.html']:
 source=Path(name);destination=backup/name;destination.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(source,destination)
files.extend(map(Path,source_names));files.extend(f for f in backup.rglob('*') if f.is_file())
files.extend(f for f in (root/'faction113-source').rglob('*') if f.is_file())
binary=Path('dist-all/v.1.1.2-20261011/full/LOGH3-v.1.1.2-20261011-linux-amd64/gin3');files.append(binary)
artifacts=[]
for path in sorted(set(files)):
 assert path.stat().st_uid==path.stat().st_gid==1000,path
 artifacts.append(dict(path=str(path),bytes=path.stat().st_size,sha256=hashlib.sha256(path.read_bytes()).hexdigest(),uid=1000,gid=1000))
index=dict(date='2026-10-11',scope='帝國B亮藍艦橋／同盟B赤焰戰線及戰略星圖；取代113同盟A現行外觀',
 git_head='523ced1346358b1d3a9982148c569a69af953a5f',worktree_dirty=True,version='v.1.1.2-20261011',
 x11=dict(windows=windows,pixels=pixels,different=0,early_attempts=attempts),approved_ui=approved,browser=browser,tests=dict(passed=15,no_tests=4),
 images={'logh3-ebiten:2':'sha256:d7f4d3e04c75fc9e9b2b3606d9f015d8f1bcde806c25d6780780450ad40827ab',
 'logh3-hd-preview:20261011-r1':'sha256:932b5ad00e356b4252945c35938d9b30b939109abfaaca8b6673f1a17f532aff'},
 new_dos_oracle_receipts=False,artifacts=artifacts,rights='本機及目前private儲存庫；未公開發行',pending=['新版推廣影片及封包交付驗證，另見115'])
destination=Path('docs/internals/data/alliance-fireline-theme.json');assert not destination.exists();destination.write_text(json.dumps(index,ensure_ascii=False,indent=2)+'\n')
print('PASS114來源',len(artifacts),windows,pixels,hashlib.sha256(destination.read_bytes()).hexdigest())
