"""原配置試樣的瀏覽器幾何、鍵鼠操作與目視截圖收據。"""
import html
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess

OUT=Path('/work/hd124-original-layout')
assert OUT.stat().st_uid==os.getuid()
source=(OUT/'HD-ORIGINAL-LAYOUT.html').read_text()
check=r'''<script>document.fonts.ready.then(()=>{
 const checks=[];function assert(ok,label){if(!ok)throw Error(label);checks.push(label)};
 const layer=name=>document.getElementById(name+'-layer');
 const click=id=>document.getElementById(id).dispatchEvent(new MouseEvent('click',{bubbles:true}));
 setView('budget');assert(!layer('budget').hasAttribute('hidden'),'預算疊窗顯示');
 const expected=geometry.budget_window,win=document.querySelector('[data-window="budget"]>rect');
 ['x','y','width','height'].forEach((key,i)=>assert(+win.getAttribute(key)===[expected.x,expected.y,expected.w,expected.h][i],'原預算視窗 '+key));
 const keys=['tax-up','tax-down','pct-up','pct-down','staff','end'];
 keys.forEach((id,i)=>{const e=document.getElementById(id),b=geometry.budget_buttons[i],r=e.dataset.rect.split(',').map(Number);assert(r.every((v,j)=>v===[b.x,b.y,b.w,b.h][j]),'原按鈕矩形 '+id)});
 click('pct-up');assert(pct===26,'建造比例增加');click('pct-down');assert(pct===25,'建造比例減少');
 click('tax-up');assert(taxLevel===3,'稅率檔位增加');click('tax-down');assert(taxLevel===2,'稅率檔位減少');
 document.getElementById('staff').dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true}));assert(document.getElementById('status').textContent.startsWith('參謀'),'參謀鍵盤回饋');
 click('end');assert(view==='account'&&layer('budget').hasAttribute('hidden')&&!layer('account').hasAttribute('hidden'),'結束回月初會計');setView('budget');document.dispatchEvent(new KeyboardEvent('keydown',{key:'Escape'}));assert(view==='account','Esc回月初會計');setView('strategy');
 for(let id=0;id<32;id++){
  // MouseEvent的clientX/Y會取整。測像素內部，避免縮小後把邊界點送到相鄰格。
  let k=regionGrid.findIndex((value,index)=>{const x=index%320,y=Math.floor(index/320);if(value!==id||x<8||x>311||y<8||y>291)return false;for(let dy=-3;dy<=3;dy++)for(let dx=-3;dx<=3;dx++)if(regionGrid[index+dy*320+dx]!==id)return false;return true});assert(k>=0,'星域可見 '+id);
  const p=new DOMPoint(k%320+.5,Math.floor(k/320)+40.5).matrixTransform(document.getElementById('art').getScreenCTM());
  layer('strategy').dispatchEvent(new MouseEvent('click',{clientX:p.x,clientY:p.y,bubbles:true}));assert(document.getElementById('status').textContent.startsWith('星域 '+id+'：'),'原遮罩選取 '+id);
 }
 document.getElementById('selected-sector').setAttribute('d','');
 setView('battle');assert(layer('strategy').hasAttribute('hidden')&&!layer('battle').hasAttribute('hidden'),'戰場與戰略分層');
 document.getElementById('reference').click();assert(document.getElementById('art').hasAttribute('hidden')&&!document.getElementById('original').hidden,'原圖切換');document.getElementById('reference').click();assert(!document.getElementById('art').hasAttribute('hidden'),'返回試樣');
 setView('budget');for(const e of document.querySelectorAll('.source-cell')){const b=e.getBBox(),cx=+e.getAttribute('x'),cw=+e.dataset.width;assert(b.x>=cx-cw/2-.25&&b.x+b.width<=cx+cw/2+.25,'字位安全 '+e.textContent)}
 assert(document.querySelector('#monthly-layer').hasAttribute('hidden')===false,'財政使用月初底層');
 const target=new URLSearchParams(location.search).get('view')||'budget';setView(target);status('A 典雅艦橋。保留原配置、按鈕與4:3比例。');
 assert(document.documentElement.scrollWidth<=innerWidth,'整頁無水平溢出');const r=document.querySelector('.screen').getBoundingClientRect();assert(Math.abs(r.width/r.height-4/3)<.001,'原4:3畫布');
 const pre=document.createElement('pre');pre.id='check-result';pre.hidden=true;pre.textContent=JSON.stringify({pass:true,width:innerWidth,view:target,checks});document.body.appendChild(pre);
}).catch(error=>{const pre=document.createElement('pre');pre.id='check-result';pre.textContent=JSON.stringify({pass:false,error:String(error)});document.body.appendChild(pre)})</script>'''
test=OUT/'BROWSER-CHECK.html'
test.write_text(source.replace('</html>',check+'</html>'))
results=[]
try:
    for width,height,view in [(1440,1280,'budget'),(1440,1280,'strategy'),(1440,1280,'battle'),(800,1000,'budget'),(390,800,'budget')]:
        label=f'{width}-{view}'
        command=['chromium','--headless','--no-sandbox','--disable-dev-shm-usage','--disable-gpu',
                 '--no-first-run','--no-default-browser-check','--hide-scrollbars','--user-data-dir=/tmp/logh3-hd124-browser',
                 f'--window-size={width},{height}','--virtual-time-budget=3000',
                 f'--screenshot={OUT}/{label}.png','--dump-dom',test.as_uri()+'?view='+view]
        run=subprocess.run(command,capture_output=True,text=True,timeout=45,check=True)
        (OUT/(label+'.log')).write_text(run.stderr)
        match=re.search(r'<pre id="check-result"[^>]*>(.*?)</pre>',run.stdout)
        assert match,run.stderr[-1200:]
        result=json.loads(html.unescape(match[1]))
        print(label,result.get('pass'),result.get('error',''),flush=True)
        assert result['pass'],result
        results.append(result)
finally:
    test.unlink()
(OUT/'BROWSER.json').write_text(json.dumps({'browser':subprocess.check_output(['chromium','--version'],text=True).strip(),
 'image':'logh3-hd-preview:20261011-r1','html_sha256':hashlib.sha256(source.encode()).hexdigest(),
 'checker_sha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
 'checks':results,'visual_review':'待目視'},ensure_ascii=False,indent=2)+'\n')
