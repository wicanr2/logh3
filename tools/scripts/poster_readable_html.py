"""158：兩艘主旗艦完整船體、仰視／朝上、四角漸層與舊譯引文來源。"""
import hashlib
import json
import os
from pathlib import Path
import re
import numpy as np
from PIL import Image

import poster_closeup_html as asset

ROOT=Path('/src');OUT=ROOT/'workplace/poster158';sha=asset.sha
assert os.getuid()==os.getgid()==1000 and OUT.is_dir()
old=ROOT/'workplace/poster157/proposal.html'
data=json.loads(old.read_text().split('const DATA=',1)[1].split(';\nconst $=',1)[0])
atlas=Image.open(OUT/'heroes-layout.png').convert('RGBA');assert atlas.size==(1774,887)
heroes=[];regions=[(1200,0,1774,887),(0,0,1200,887)]
for side,region in enumerate(regions):
    cell=atlas.crop(region);alpha=np.array(cell)[:,:,3];yy,xx=np.where(alpha>=32);bounds=(max(0,int(xx.min())-3),max(0,int(yy.min())-3),min(cell.width,int(xx.max())+4),min(cell.height,int(yy.max())+4));cell=cell.crop(bounds)
    target=OUT/('brunhild-up.png' if side==0 else 'hyperion-low.png');cell.save(target)
    heroes.append(dict(side=side,width=cell.width,height=cell.height,src=asset.encoded(cell)))
data['heroes']=heroes
updates={
 'poster.quote.empire':{'zh-TW':'嗯！星星好美啊！','zh-CN':'嗯！星星好美啊！','en':'Hmm! The stars are beautiful!','ja':'ああ、星はいい……','ko':'음! 별이 참 아름답군!'},
 'poster.quote.alliance':{'zh-TW':'我想最好的還是和平吧。','zh-CN':'我想最好的还是和平吧。','en':'I think peace is best.','ja':'いちばんよいのは平和だと思います。','ko':'가장 좋은 것은 평화라고 생각해.'}}
data['quotes'].update(updates)
template=(ROOT/'tools/scripts/poster_flagship_rows.html').read_text()
template=template.replace('上方同盟，下方帝國。','看清楚，兩艘主旗艦。')
template=template.replace('同盟旗艦朝左，帝國旗艦朝右。以遠近大小、錯位與前後遮擋排列景深；兩側名言與姓名一起放進人物圖片下緣。','上方休伯利安朝左、由艦底仰視；下方伯倫希爾艦首朝上。四角以漸層銜接人物海報，兩艘主旗艦保留完整船體空間。')
template=template.replace('上：同盟 ←　／　下：帝國 →','上：休伯利安仰視 ←　／　下：帝國 ↑')
template=template.replace('帝國朝右','帝國朝上').replace('下方帝國艦首朝右','下方帝國艦首朝上')
template=template.replace('id="empire-row" data-facing="right"','id="empire-row" data-facing="up"')
template=template.replace("bottom:'empire-right'", "bottom:'empire-up'")
template=template.replace("proposal:'上下旗艦景深隊列'", "proposal:'主旗艦完整船體與四角漸層158'")
template=template.replace('<image id="background" x="0" y="0" width="1280" height="720"/>','<rect width="1280" height="720" fill="#07101c"/><svg id="game-canvas" x="160" y="0" width="960" height="720" viewBox="160 0 960 720" overflow="hidden"><image id="background" x="0" y="0" width="1280" height="720"/></svg>')
template=template.replace('<g id="portraits"></g>','<g id="portraits"></g>')
template=template.replace('<rect x="160" y="0" width="960" height="62" fill="url(#row-bg)"/>','<rect id="top-bg" x="160" y="0" width="960" height="62" fill="url(#row-bg)"/>')
template=template.replace('<rect x="160" y="627" width="960" height="93" fill="url(#row-bg)"/>','<rect id="bottom-bg" x="160" y="627" width="960" height="93" fill="url(#row-bg)"/>')
template=re.sub(r'<path d="M160 (?:61\.5|627\.5)H1120"[^>]*/>','',template)
template=template.replace('<g id="guides"','<g id="junctions"></g><g id="portrait-labels"></g><g id="guides"')
template=template.replace('</defs>','<linearGradient id="seam-left" gradientUnits="userSpaceOnUse"><stop stop-color="white" stop-opacity="0"/><stop offset=".86" stop-color="white" stop-opacity="1"/></linearGradient><linearGradient id="seam-right" gradientUnits="userSpaceOnUse"><stop stop-color="white" stop-opacity="0"/><stop offset=".86" stop-color="white" stop-opacity="1"/></linearGradient><mask id="seam-mask-left" maskUnits="userSpaceOnUse" x="0" y="0" width="1280" height="720"><rect width="1280" height="720" fill="url(#seam-left)"/></mask><mask id="seam-mask-right" maskUnits="userSpaceOnUse" x="0" y="0" width="1280" height="720"><rect width="1280" height="720" fill="url(#seam-right)"/></mask></defs>')
template=template.replace('<h2>調整隊列與人物文字</h2>','<h2>調整船體與銜接</h2>')
extra='''<div class="actions" style="margin:6px 0 14px;padding:0;border:0"><button id="originalSize">原遊戲大小</button><button id="hullPriority" class="primary">船體優先</button></div><div class="control"><label for="gameScale">遊戲顯示大小（保持4:3）<output id="gameScaleValue"></output></label><input id="gameScale" type="range" min="84" max="100" step="1" value="88"></div><div class="control"><label for="fade">四角漸層<output id="fadeValue"></output></label><input id="fade" type="range" min="0" max="100" step="5" value="70"></div>'''
template=template.replace('<div class="control"><label for="count">',extra+'<div class="control"><label for="count">',1)
template=template.replace('近處旗艦大小','主旗艦顯示倍率').replace('min="350" max="750" step="10" value="520"','min="80" max="140" step="5" value="100"')
template=template.replace("const defaults={count:6,size:520,", "const defaults={gameScale:88,fade:70,count:6,size:100,")
start=template.index('function portraits(c)');end=template.index('for(const[k,v]of Object.entries(defaults))',start)
template=template[:start]+(ROOT/'tools/scripts/poster_readable_js.txt').read_text()+'\n'+template[end:]
template=template.replace("$('background').setAttribute('href',DATA.background);", "$('background').setAttribute('href',DATA.background);$('originalSize').addEventListener('click',()=>{$('gameScale').value=100;render()});$('hullPriority').addEventListener('click',()=>{$('gameScale').value=88;render()});")
template=template.replace('DATA.background,...DATA.ships.map(s=>s.src),...DATA.portraits.map(s=>s.src)','DATA.background,...DATA.ships.map(s=>s.src),...DATA.portraits.map(s=>s.src),...DATA.heroes.map(s=>s.src)')
template=template.replace('遠艦較小、位置稍高，近艦較大並遮住後方部分艦體。每艘沿同一方向固定排列，保留大比例局部特寫。','兩艘主旗艦各有獨立的前景空間，讓主要船體完整可辨識；其他旗艦排列於後景。本體固定，只有武器活動。')
template=template.replace('同盟旗艦，朝左，以遠近大小與遮擋排列。','同盟旗艦朝左，休伯利安採艦底仰視並保留清楚船體。')
template=template.replace('帝國旗艦朝右，採同樣的景深隊列。','帝國旗艦艦首朝上，伯倫希爾在前景保留完整船體。')
quotes_section='''<section class="summary"><h2>名言來源核對</h2><table><tbody><tr><th>萊因哈特</th><td>「嗯！星星好美啊！」<br><a href="https://www.hkweb.org/logh/story/1-1.htm" target="_blank" rel="noreferrer">《黎明篇》第一章〈永恆的夜〉舊版轉錄</a></td></tr><tr><th>楊威利</th><td>「我想最好的還是和平吧。」<br><a href="https://hkweb.org/ginei/yang/" target="_blank" rel="noreferrer">《黎明篇》第五章〈伊謝爾倫攻略戰〉引文節錄</a></td></tr><tr><th>查證狀態</th><td>紅茶館明列其轉錄來源為尖端二十冊舊版；目前已核對網頁用語與章節，尚未取得紙本頁面逐字核對。其他語系為本專案對應譯文，不稱尖端外語版。</td></tr></tbody></table></section>'''
template=template.replace('<footer>',quotes_section+'<footer>')
shared=re.search(r'<style>([\s\S]*?)</style>',(ROOT/'tools/scripts/poster_closeup.html').read_text()).group(1)
# 外框自適應樣式只作用最外層SVG；內層遊戲viewport使用自己的幾何。
shared=shared.replace('.frame svg{','.frame>svg{').replace('.frame:fullscreen svg{','.frame:fullscreen>svg{')
page=template.replace('/*__SHARED_CSS__*/',shared).replace('/*__ROW_DATA__*/',json.dumps(data,ensure_ascii=False))
(OUT/'proposal.html').write_text(page)
sources=dict(prototype=True,version='poster158',html_sha256=sha(OUT/'proposal.html'),input_html_sha256=sha(old),
    generator_sha256=sha(Path(__file__)),js_sha256=sha(ROOT/'tools/scripts/poster_readable_js.txt'),
    browser_script_sha256=sha(ROOT/'tools/scripts/poster_readable_browser.mjs'),
    heroes_atlas_sha256=sha(OUT/'heroes-layout.png'),heroes_raw_sha256=sha(OUT/'heroes-raw.png'),
    hero_assets={p.name:dict(sha256=sha(p),size=list(Image.open(p).size)) for p in [OUT/'brunhild-up.png',OUT/'hyperion-low.png']},
    prompts={p.name:sha(p) for p in [OUT/'hero-prompt.txt',OUT/'layout-prompt.txt']},
    classification='新視角為依既有HD素材生成的提案重構，不稱動畫或DOS逐形證據；僅本機',
    layout='船體優先預設遊戲88%，比例維持4:3；原遊戲大小100%可切換，未核准接正式版',
    production_sha256=sha(ROOT/'cmd/gin3/poster_decoration.go'),poster_go_sha256=sha(ROOT/'cmd/gin3/poster.go'),
    quotes=dict(status='secondary-transcription-checked; print-page-unverified',records=[
      dict(character='萊因哈特',text=updates['poster.quote.empire']['zh-TW'],url='https://www.hkweb.org/logh/story/1-1.htm',chapter='黎明篇第一章永恆的夜'),
      dict(character='楊威利',text=updates['poster.quote.alliance']['zh-TW'],url='https://hkweb.org/ginei/yang/',chapter='黎明篇第五章伊謝爾倫攻略戰',classification='原句中一個子句的節錄')],
      edition_claim='https://www.hkweb.org/logh/story/story.htm',limits='二手轉錄，未取得舊版書頁、版權頁或確定頁碼；不冒稱紙本逐字核對'))
(OUT/'SOURCE.json').write_text(json.dumps(sources,ensure_ascii=False,indent=2)+'\n')
print('READY 主旗艦船體／仰視／朝上／漸層HTML', (OUT/'proposal.html').stat().st_size,flush=True)
