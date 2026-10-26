"""160：休伯利安重繪、伯倫希爾下移，以及沿人物海報連續的銀河。"""
import base64
import hashlib
import json
import os
from pathlib import Path
import numpy as np
from PIL import Image

ROOT=Path('/src');OUT=ROOT/'workplace/poster160'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
assert OUT.is_dir() and os.getuid()==os.getgid()==1000
old=ROOT/'workplace/poster159/proposal.html';page=old.read_text()
a=page.index('const DATA=');b=page.index(';\nconst $=',a)
data=json.loads(page[a+11:b]);original=data['heroes'][1]['src']
p=OUT/'hyperion-side.png';im=Image.open(p).convert('RGBA');alpha=np.array(im)[:,:,3];yy,xx=np.where(alpha>=32)
data['heroes'][1]=dict(side=1,width=im.width,height=im.height,hullX=int(xx.min()),hullY=int(yy.min()),hullWidth=int(xx.max()-xx.min()+1),hullHeight=int(yy.max()-yy.min()+1),src='data:image/png;base64,'+base64.b64encode(p.read_bytes()).decode())
page=page[:a]+'const DATA='+json.dumps(data,ensure_ascii=False)+page[b:]
assert original not in page
page=page.replace('上下，各一艘主旗艦。','讓銀河，連成一片。')
page=page.replace('上方只留朝左的休伯利安；下方只放伯倫希爾，依參考圖反方向、艦首朝右，白色船體橫跨整個下緣。四角漸層與人物圖內文字保留。','休伯利安依提供的原作參考重繪，保留方形艦首與兩組圓孔。伯倫希爾再往下移，朝右橫跨下緣。上下銀河連接楊威利與萊因哈特的海報。')
page=page.replace('topY:0,bottomY:0,','topY:0,bottomY:24,')
page=page.replace('topSize:90,bottomSize:100','topSize:65,bottomSize:100')
page=page.replace('angle=side?-12:-10','angle=side?0:-10')
page=page.replace('-h*(side?.45:.65)','-h*(side?.5:.65)')
page=page.replace('x=l.left+28,y=l.topH*.51,ex=l.right-28,ey=l.bottomY+l.bottomH*.49','x=l.left+l.w*(1-c.topSize/100)/2,y=l.topH*.51+c.topY,ex=l.right-28,ey=l.bottomY+l.bottomH*.49+c.bottomY')
page=page.replace('M${x+50} ${y}H${x}','M${x} ${y}H${x-50}').replace('mx=x+50-(p-3)*16','mx=x-(p-3)*16')
page=page.replace('小型分層','單艦特寫')
page=page.replace('四角漸層與人物圖內文字保留','銀河串接與人物圖內文字保留')
page=page.replace('保留仰視船體的視角。','按使用者提供的原作圖重繪，保留方形艦首與上下兩組圓孔。')
page=page.replace('船體橫跨下緣。</td>','船體往下24像素，橫跨下緣。</td>')
page=page.replace('本體固定，只有武器活動。</p>','本體固定，只有武器活動；上下的銀河沿用人物海報星空，串接兩側背景。</p>')
page=page.replace('四角漸層<output','銀河銜接漸層<output')
page=page.replace('調整雙艦特寫','調整銀河與雙艦')
page=page.replace("proposal:'上下各一艘大型旗艦159'","proposal:'銀河串接與原作休伯利安160'")
sky=data['portraits'][0];sky_h=int(sky['height']*.15);tile_w=sky['width']*720/sky['height'];tile_h=sky_h*720/sky['height']
# 原圖上方15%沒有角色。SVG採同一份星空紋理與全域座標，兩條飾框及四角共用。
# 僅在HTML取樣與合成，不改寫任何原PNG或角色圖。
pattern=f'''<pattern id="shared-galaxy" patternUnits="userSpaceOnUse" x="0" y="0" width="{tile_w*2}" height="{tile_h*2}"><svg id="galaxy-cell" x="0" y="0" width="{tile_w}" height="{tile_h}" viewBox="0 0 {sky['width']} {sky_h}" overflow="hidden"><image id="galaxy-source" width="{sky['width']}" height="{sky['height']}"/></svg><use href="#galaxy-cell" transform="translate({tile_w*2} 0) scale(-1 1)"/><use href="#galaxy-cell" transform="translate(0 {tile_h*2}) scale(1 -1)"/><use href="#galaxy-cell" transform="translate({tile_w*2} {tile_h*2}) scale(-1 -1)"/></pattern>'''
page=page.replace('</defs>',pattern+'</defs>',1)
page=page.replace('fill="#07101c"','fill="url(#shared-galaxy)"')
page=page.replace('fill="url(#row-bg)"','fill="url(#shared-galaxy)"').replace("fill:'url(#row-bg)'","fill:'url(#shared-galaxy)'")
page=page.replace("$('background').setAttribute('href',DATA.background);","$('background').setAttribute('href',DATA.background);$('galaxy-source').setAttribute('href',DATA.portraits[0].src);")
page=page.replace('function render(){const c=settings(),l=layout(c);',f"function render(){{const c=settings(),l=layout(c);$('shared-galaxy').setAttribute('x',-({tile_w}-l.left)*.56);")
assert page.count('id="shared-galaxy"')==1 and 'bottomY:24' in page and 'angle=side?0:-10' in page
(OUT/'proposal.html').write_text(page)
record=dict(prototype=True,version='poster160',html_sha256=sha(OUT/'proposal.html'),input_html_sha256=sha(old),generator_sha256=sha(Path(__file__)),browser_script_sha256=sha(ROOT/'tools/scripts/poster_galaxy_browser.mjs'),production_sha256=sha(ROOT/'cmd/gin3/poster_decoration.go'),poster_go_sha256=sha(ROOT/'cmd/gin3/poster.go'),hyperion=dict(asset=p.name,sha256=sha(p),prompts={n:sha(OUT/n) for n in ['hyperion-prompt.txt','hyperion-side-prompt.txt']},references={n:sha(OUT/n) for n in ['hyperion-bow-reference.png','hyperion-body-reference.png']},classification='使用者參考圖重繪；原型藝術重構，不稱動畫逐形完全一致'),galaxy=dict(source='empire-portrait.png',sha256=sha(OUT/'empire-portrait.png'),sample=[0,0,sky['width'],sky_h],tile=[tile_w,tile_h],composition='與人物圖相同尺度，鏡像連續鋪設，四角漸層融合；兩側原人物PNG未改寫'),layout=dict(bottomY=24,gameScale=88,topSize=65,bottomSize=100),layout_reference_sha256=sha(OUT/'layout-reference.png'),research=dict(primary_page='https://www.kotobukiya.co.jp/product/detail/p4934054063635/',image_path_classification='KP771_hyperion_ishiguroban',user_references_take_precedence=True,download='官方縮圖下載403，未冒稱已保存'),quote_status='沿158二手轉錄，未完成尖端紙本逐字／頁碼核對')
(OUT/'SOURCE.json').write_text(json.dumps(record,ensure_ascii=False,indent=2)+'\n')
print('READY galaxy HTML', (OUT/'proposal.html').stat().st_size,flush=True)
