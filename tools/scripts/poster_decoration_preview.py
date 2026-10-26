"""154：從通過的正式AppImage收據複製README截圖及本機循環預覽。"""
import json
import os
from pathlib import Path
import shutil
import subprocess

from poster_decoration_release import ROOT,OUT,VERSION,sha,read,write

assert os.getuid()==os.getgid()==1000
resultdir=ROOT/'workplace/poster154/player-linux';result=read(resultdir/'RESULT.json')
assert result['passed'] and result['appimage']
source=resultdir/'decor-v1.png';event=next(v for v in result['events'] if v['name']=='decor-v1')
assert sha(source)==event['sha256']
target=ROOT/'docs/images/readme-v162-poster.png';index=target.with_suffix('.json')
assert not target.exists() and not index.exists() and target.parent.stat().st_uid==1000
shutil.copyfile(source,target)
write(index,dict(version=VERSION,normal_entry=True,state_injection=False,source=str(source.relative_to(ROOT)),sha256=sha(target),
    player_result=dict(path=str((resultdir/'RESULT.json').relative_to(ROOT)),sha256=sha(resultdir/'RESULT.json')),
    scope='正式AppImage正常同盟入口，v1海報與上下循環裝飾；非DOS parity'))
promo=OUT/'promo';assert not promo.exists();promo.mkdir()
source=resultdir/'loop.mp4';assert sha(source)==result['loop']['sha256']
target=promo/('LOGH3-'+VERSION+'-poster-loop.mp4');shutil.copyfile(source,target)
ff=json.loads(subprocess.check_output(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(target)],text=True))
assert len(ff['streams'])==1 and float(ff['format']['duration'])>=17.9
write(promo/'FFPROBE.json',ff)
write(promo/'PREVIEW.json',dict(version=VERSION,classification='原版及衍生畫面，本機預覽，不公開',
    type='18秒正常視窗無聲循環預覽，不取代v1.6.0完整推廣片',file=target.name,bytes=target.stat().st_size,sha256=sha(target),
    source_result_sha256=sha(resultdir/'RESULT.json'),normal_entry=True,recording='Docker Xvfb軟體繪圖，15fps擷取；不宣稱實機FPS',
    assets=['既有遊戲HD旗艦圖集','assets/hd/poster/fighters.SOURCE.json','既有三組海報與名言來源']))
(promo/'SHA256SUMS.txt').write_text(sha(target)+'  '+target.name+'\n')
print('PASS 正式AppImage原截圖及18秒無聲預覽',flush=True)
