"""171：沿166正常前線整軍輸入補錄，縮短拍攝停留，保持遊戲原速。"""
import hashlib
from pathlib import Path
import subprocess
import sys

source=Path('tools/scripts/poster_frame_garrison_gui.py')
assert hashlib.sha256(source.read_bytes()).hexdigest()=='76b40657f62ee9cba9f09d40d0cac719ffe63f3b9c30f822a7b4088ec2831bbe'
text=source.read_text();old='time.sleep(1.2 if a.video else .2)';assert text.count(old)==1
text=text.replace(old,'time.sleep(.6 if a.video else .2)')
output=Path(sys.argv[1]);assert not output.exists()and output.parent.stat().st_uid==1000
output.write_text(text)
subprocess.run([sys.executable,str(output),*sys.argv[2:]],check=True)
