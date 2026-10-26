"""162：以全高海報邊緣羽化取代四角矩形覆蓋，去除腳邊銀河切面。"""
import hashlib
import json
import os
from pathlib import Path

ROOT=Path('/src');OUT=ROOT/'workplace/poster162'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
assert OUT.is_dir() and os.getuid()==os.getgid()==1000
old=ROOT/'workplace/poster161/proposal.html';page=old.read_text()
a=page.index('function junctions(c)');b=page.index('function layout(c)',a)
page=page[:a]+'''function junctions(c){const l=layout(c),fade=c.fade;for(const [side,x1,x2,key]of[['left',l.left-fade,l.left,'empire'],['right',l.right,l.right+fade,'alliance']]){const gradient=$('portrait-feather-'+side);gradient.setAttribute('x1',x1);gradient.setAttribute('x2',x2);const panel=document.querySelector('.portrait-panel[data-side="'+key+'"]');if(fade>0)panel.setAttribute('mask','url(#portrait-mask-'+side+')');else panel.removeAttribute('mask')}$('junctions').replaceChildren()}
'''+page[b:]
defs='''<linearGradient id="portrait-feather-left" gradientUnits="userSpaceOnUse" y1="0" y2="0"><stop stop-color="white"/><stop offset="1" stop-color="white" stop-opacity="0"/></linearGradient><linearGradient id="portrait-feather-right" gradientUnits="userSpaceOnUse" y1="0" y2="0"><stop stop-color="white" stop-opacity="0"/><stop offset="1" stop-color="white"/></linearGradient><mask id="portrait-mask-left" maskUnits="userSpaceOnUse" x="0" y="0" width="1280" height="720"><rect width="1280" height="720" fill="url(#portrait-feather-left)"/></mask><mask id="portrait-mask-right" maskUnits="userSpaceOnUse" x="0" y="0" width="1280" height="720"><rect width="1280" height="720" fill="url(#portrait-feather-right)"/></mask>'''
page=page.replace('</defs>',defs+'</defs>',1)
assert page.count('gameScale:88,fade:70,')==1
page=page.replace('gameScale:88,fade:70,','gameScale:88,fade:28,')
page=page.replace('id="fade" type="range" min="0" max="100" step="5" value="70"','id="fade" type="range" min="0" max="100" step="2" value="28"')
page=page.replace("proposal:'伯倫希爾艦尾取景161'","proposal:'人物腳邊銀河切面修正162'")
assert 'id="fade" type="range" min="0" max="100" step="2"' in page
(OUT/'proposal.html').write_text(page)
record=dict(prototype=True,version='poster162',html_sha256=sha(OUT/'proposal.html'),input_html_sha256=sha(old),generator_sha256=sha(Path(__file__)),browser_script_sha256=sha(ROOT/'tools/scripts/poster_seam_browser.mjs'),production_sha256=sha(ROOT/'cmd/gin3/poster_decoration.go'),poster_go_sha256=sha(ROOT/'cmd/gin3/poster.go'),change=dict(method='全高左右海報邊緣透明遮罩，移除四角矩形覆蓋',fade=28,scope='只修原型星空銜接，人物姓名名言為獨立上層，艦艇位置與素材保持'),classification='本機HTML原型，不是正式遊戲或發行包')
(OUT/'SOURCE.json').write_text(json.dumps(record,ensure_ascii=False,indent=2)+'\n')
print('READY seam HTML', (OUT/'proposal.html').stat().st_size,flush=True)
