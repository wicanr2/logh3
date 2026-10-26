"""161：只修伯倫希爾艦尾的左側截斷，保留原素材與其他版面。"""
import hashlib
import json
import os
from pathlib import Path

ROOT=Path('/src');OUT=ROOT/'workplace/poster161'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
assert OUT.is_dir() and os.getuid()==os.getgid()==1000
old=ROOT/'workplace/poster160/proposal.html';page=old.read_text()
assert page.count('topSize:65,bottomSize:100,')==1
assert page.count('l.left+l.w/2-(side?0:16)')==1
page=page.replace('topSize:65,bottomSize:100,','topSize:65,bottomSize:92,')
page=page.replace('l.left+l.w/2-(side?0:16)','l.left+l.w/2')
# range原為5%一步；92%必須以1%一步保留，避免瀏覽器自動捨入至90%。
page=page.replace('id="bottomSize" type="range" min="60" max="140" step="5" value="100"','id="bottomSize" type="range" min="60" max="140" step="1" value="92"')
page=page.replace("proposal:'銀河串接與原作休伯利安160'","proposal:'伯倫希爾艦尾取景161'")
page=page.replace('船體往下24像素，橫跨下緣。','船體往下24像素，左右保留邊距，讓艦尾輪廓避開裁切及漸層。')
assert 'id="bottomSize" type="range" min="60" max="140" step="1"' in page
(OUT/'proposal.html').write_text(page)
source=json.loads((ROOT/'workplace/poster160/SOURCE.json').read_text())
record=dict(prototype=True,version='poster161',html_sha256=sha(OUT/'proposal.html'),input_html_sha256=sha(old),input_source_sha256=sha(ROOT/'workplace/poster160/SOURCE.json'),generator_sha256=sha(Path(__file__)),browser_script_sha256=sha(ROOT/'tools/scripts/poster_stern_browser.mjs'),production_sha256=sha(ROOT/'cmd/gin3/poster_decoration.go'),poster_go_sha256=sha(ROOT/'cmd/gin3/poster.go'),change=dict(bottomSize=92,horizontal_center_offset=0,bottomY=24,scope='只調HTML等比大小及水平位置，原素材與其他版面保持'),inherited_art_source=source['hyperion']['sha256'],classification='本機HTML原型，不是正式程式或發行包')
(OUT/'SOURCE.json').write_text(json.dumps(record,ensure_ascii=False,indent=2)+'\n')
print('READY stern HTML', (OUT/'proposal.html').stat().st_size,flush=True)
