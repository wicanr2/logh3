"""159：上下各一艘大型旗艦，依使用者參考圖反向伯倫希爾。"""
import base64
import hashlib
import json
import os
from pathlib import Path
import re
import numpy as np
from PIL import Image

ROOT=Path('/src');OUT=ROOT/'workplace/poster159'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
assert OUT.is_dir() and os.getuid()==os.getgid()==1000
old=ROOT/'workplace/poster158/proposal.html';page=old.read_text()
start=page.index('const DATA=');end=page.index(';\nconst $=',start)
data=json.loads(page[start+11:end]);data['ships']=[]
assets=[]
for side,p in enumerate([OUT/'brunhild-right.png',ROOT/'workplace/poster158/hyperion-low.png']):
    im=Image.open(p).convert('RGBA');alpha=np.array(im)[:,:,3];yy,xx=np.where(alpha>=32)
    assets.append(dict(side=side,width=im.width,height=im.height,hullX=int(xx.min()),hullY=int(yy.min()),hullWidth=int(xx.max()-xx.min()+1),hullHeight=int(yy.max()-yy.min()+1),src='data:image/png;base64,'+base64.b64encode(p.read_bytes()).decode()))
data['heroes']=assets
page=page[:start]+'const DATA='+json.dumps(data,ensure_ascii=False)+page[end:]
page=page.replace('旗艦景深隊列與人物名言','上下各一艘大型旗艦').replace('海報提案 · 旗艦景深隊列','海報提案 · 雙主旗艦')
page=page.replace('看清楚，兩艘主旗艦。','上下，各一艘主旗艦。')
page=page.replace('上方休伯利安朝左、由艦底仰視；下方伯倫希爾艦首朝上。四角以漸層銜接人物海報，兩艘主旗艦保留完整船體空間。','上方只留朝左的休伯利安；下方只放伯倫希爾，依參考圖反方向、艦首朝右，白色船體橫跨整個下緣。四角漸層與人物圖內文字保留。')
page=page.replace('上：休伯利安仰視 ←　／　下：帝國 ↑','上：休伯利安 ←　／　下：伯倫希爾 →')
page=page.replace('上方同盟艦首朝左，下方帝國艦首朝上，遠近排列；名言姓名在人物圖內','上方一艘休伯利安朝左，下方一艘伯倫希爾朝右，橫跨下緣；名言姓名在人物圖內')
page=page.replace('id="empire-row" data-facing="up"','id="empire-row" data-facing="right"').replace('帝國朝上','伯倫希爾朝右')
page=page.replace('<small>隊列</small><h2>由深處到近處，逐艘展開</h2><p>兩艘主旗艦各有獨立的前景空間，讓主要船體完整可辨識；其他旗艦排列於後景。本體固定，只有武器活動。</p>','<small>單艦特寫</small><h2>船體橫跨飾框</h2><p>上下各一艘大型旗艦。依比例放大，允許上下局部出框；保留可辨識的主要船體，不拉伸圖像。本體固定，只有武器活動。</p>')
page=page.replace('調整船體與銜接','調整雙艦特寫').replace('可調整景深、間距及文字位置，下載設定供定稿。','可調整上下船體大小、取景與漸層，下載設定供定稿。')
for key in ['count','size','depth','spread']:
    page,n=re.subn(r'<div class="control"><label for="'+key+r'">.*?</div>','',page,count=1);assert n==1,key
controls=''.join('<div class="control"><label for="'+key+'">'+label+'<output id="'+key+'Value"></output></label><input id="'+key+'" type="range" min="60" max="140" step="5" value="'+str(default)+'"></div>' for key,label,default in [('topSize','休伯利安寬度',90),('bottomSize','伯倫希爾寬度',100)])
page=page.replace('<div class="control"><label for="topY">',controls+'<div class="control"><label for="topY">',1)
page=page.replace('帝國旗艦艦首朝上，伯倫希爾在前景保留完整船體。','只放一艘伯倫希爾，依參考圖反方向朝右，船體橫跨下緣。')
page=page.replace('全部改成同盟旗艦朝左，休伯利安採艦底仰視並保留清楚船體。','只放一艘休伯利安，朝左，保留仰視船體的視角。')
page=page.replace('上下兩軍旗艦隊列','上下各一艘旗艦').replace('顯示隊列裁切範圍','顯示上下取景範圍')
page=page.replace('const defaults={gameScale:88,fade:70,count:6,size:100,depth:48,spread:830,','const defaults={gameScale:88,fade:70,topSize:90,bottomSize:100,')
js=(ROOT/'tools/scripts/poster_single_js.txt').read_text()
# 沿用人物與漸層函式，只替換原艦列、render與武器。
page=page[:page.index('function fleet(c,side)')]+page[page.index('function junctions(c)'):]
a=page.index('function layout(c)');b=page.index('function portraits(c)');page=page[:a]+page[b:]
a=page.index('function render()');b=page.index('for(const[k,v]of Object.entries(defaults))',a);page=page[:a]+js+'\n'+page[b:]
page=page.replace("proposal:'主旗艦完整船體與四角漸層158'","proposal:'上下各一艘大型旗艦159'").replace("bottom:'empire-up'","bottom:'empire-right'").replace("a.download='旗艦景深與人物文字設定.json'","a.download='上下單旗艦設定.json'")
assert all(s not in page for s in ['countValue','depthValue','empire-up','艦首朝上','其他旗艦排列於後景'])
(OUT/'proposal.html').write_text(page)
record=dict(prototype=True,version='poster159',html_sha256=sha(OUT/'proposal.html'),input_html_sha256=sha(old),generator_sha256=sha(Path(__file__)),js_sha256=sha(ROOT/'tools/scripts/poster_single_js.txt'),browser_script_sha256=sha(ROOT/'tools/scripts/poster_single_browser.mjs'),reference_sha256=sha(OUT/'brunhild-reference.png'),generated_asset_sha256=sha(OUT/'brunhild-right.png'),prompt_sha256=sha(OUT/'brunhild-prompt.txt'),hyperion_sha256=sha(ROOT/'workplace/poster158/hyperion-low.png'),production_sha256=sha(ROOT/'cmd/gin3/poster_decoration.go'),poster_go_sha256=sha(ROOT/'cmd/gin3/poster.go'),classification='使用者參考圖衍生的本機HTML提案；imagegen內建工具；非動畫或DOS逐形證據',layout='單艦大型特寫，下方朝右；預設遊戲88%，上下寬度90%／100%，可調；不拉伸圖像',quote_status='沿158二手文字核對，尖端紙本與確定頁碼未核對',assets=[{k:v for k,v in a.items() if k!='src'} for a in assets])
(OUT/'SOURCE.json').write_text(json.dumps(record,ensure_ascii=False,indent=2)+'\n')
print('READY single flagships HTML', (OUT/'proposal.html').stat().st_size,flush=True)
