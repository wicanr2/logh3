"""163：以原海報自然留邊向外融合，三組共用同一銜接機制。"""
import base64
import hashlib
import json
import os
from pathlib import Path
import re
from PIL import Image

ROOT=Path('/src');OUT=ROOT/'workplace/poster163';sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
assert OUT.is_dir() and os.getuid()==os.getgid()==1000
old=ROOT/'workplace/poster162/proposal.html';page=old.read_text();a=page.index('const DATA=');b=page.index(';\nconst $=',a);data=json.loads(page[a+11:b])
sets={'v1':data['portraits']};inputs=[]
for version in ['v2','v3']:
    pair=[]
    for side in ['empire','alliance']:
        p=ROOT/'assets/hd/poster'/f'{version}-{side}-narrow.png';im=Image.open(p);pair.append(dict(width=im.width,height=im.height,src='data:image/png;base64,'+base64.b64encode(p.read_bytes()).decode()));inputs.append(dict(file=str(p.relative_to(ROOT)),sha256=sha(p),size=list(im.size)))
    sets[version]=pair
data['posterSets']=sets
page=page[:a]+'const DATA='+json.dumps(data,ensure_ascii=False)+page[b:]
a=page.index('function portraits(c)');b=page.index('function layout(c)',a);page=page[:a]+(ROOT/'tools/scripts/poster_sets_blend_js.txt').read_text()+'\n'+page[b:]
# 遊戲沿原本繪製路徑；星圖底框移到人物下面，天然留邊只在上下獨立層延伸。
backgrounds=re.findall(r'<rect id="(?:top-bg|bottom-bg)"[^>]*/>',page);assert len(backgrounds)==2
for background in backgrounds:page=page.replace(background,'',1)
page=page.replace('<g id="portraits"></g>','<g id="galaxy-frame-bg">'+''.join(backgrounds)+'</g><g id="portraits"></g>',1)
page=page.replace('</defs>','<clipPath id="left-extension-clip"></clipPath><clipPath id="right-extension-clip"></clipPath></defs>',1)
page=page.replace('gameScale:88,fade:28,','gameScale:88,fade:100,').replace("locale:'zh-TW',","locale:'zh-TW',posterSet:'v1',")
page=page.replace('id="fade" type="range" min="0" max="100" step="2" value="28"','id="fade" type="range" min="0" max="100" step="2" value="100"')
page=page.replace('銀河銜接漸層','原圖留邊融合').replace("['topSize','bottomSize','gameScale'].includes(k)","['topSize','bottomSize','gameScale','fade'].includes(k)")
selector='<div class="control"><label for="posterSet">海報組別</label><select id="posterSet"><option value="v1">v1 萊因哈特／楊威利</option><option value="v2">v2 帝國／同盟將領</option><option value="v3">v3 兩位主角的親近者</option></select></div>'
page=page.replace('<div class="control"><label for="locale">',selector+'<div class="control"><label for="locale">',1)
page=page.replace('...DATA.portraits.map(s=>s.src),...DATA.heroes.map(s=>s.src)','...Object.values(DATA.posterSets).flatMap(pair=>pair.map(s=>s.src)),...DATA.heroes.map(s=>s.src)')
page=page.replace("proposal:'人物腳邊銀河切面修正162'","proposal:'三組海報原圖留邊銜接163'")
assert 'function portraits(c)' in page and 'posterSet' in page and 'left-extension-clip' in page and 'game-body-clip' not in page
(OUT/'proposal.html').write_text(page)
record=dict(prototype=True,version='poster163',html_sha256=sha(OUT/'proposal.html'),input_html_sha256=sha(old),generator_sha256=sha(Path(__file__)),js_sha256=sha(ROOT/'tools/scripts/poster_sets_blend_js.txt'),browser_script_sha256=sha(ROOT/'tools/scripts/poster_sets_blend_browser.mjs'),production_sha256=sha(ROOT/'cmd/gin3/poster_decoration.go'),poster_go_sha256=sha(ROOT/'cmd/gin3/poster.go'),inputs=inputs,method='原PNG依原比例延伸到天然留邊，在原裁切線外融合；上下星空獨立延伸層，遊戲保留原繪製路徑；三組共用',labels='v1保留五語名言姓名，v2/v3沿既有規格不加v1署名',classification='本機HTML原型；未改寫任何原PNG，沒有生成新人物')
(OUT/'SOURCE.json').write_text(json.dumps(record,ensure_ascii=False,indent=2)+'\n')
print('READY three sets blend HTML', (OUT/'proposal.html').stat().st_size,flush=True)
