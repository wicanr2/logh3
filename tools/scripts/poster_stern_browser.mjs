// 161：核對實際透明像素的左右輪廓，消除艦尾裁切及漸層覆蓋。
import fs from 'node:fs/promises';
import {spawn} from 'node:child_process';
import {createHash} from 'node:crypto';
const out='/work',delay=ms=>new Promise(r=>setTimeout(r,ms));
const errors=[],requests=[],pending=new Map(),observations=[];let ws,seq=0;
const log=await fs.open(out+'/browser.log','w');
const browser=spawn('/ms-playwright/chromium-1194/chrome-linux/chrome',['--headless=new','--no-sandbox','--disable-gpu','--disable-dev-shm-usage','--no-first-run','--hide-scrollbars','--force-color-profile=srgb','--force-device-scale-factor=1','--remote-debugging-port=9222','--remote-allow-origins=*','--user-data-dir=/tmp/poster161-chrome','about:blank'],{stdio:['ignore',log.fd,log.fd]});
const assert=(v,m)=>{if(!v)throw Error(m)};
async function call(method,params={}){const id=++seq,p=new Promise((resolve,reject)=>pending.set(id,{resolve,reject}));ws.send(JSON.stringify({id,method,params}));return p}
async function evaluate(expression){const r=await call('Runtime.evaluate',{expression,returnByValue:true,awaitPromise:true});if(r.exceptionDetails)throw Error(JSON.stringify(r.exceptionDetails));return r.result.value}
async function shot(name,clip){const r=await call('Page.captureScreenshot',{format:'png',captureBeyondViewport:true,...(clip?{clip}:{})});await fs.writeFile(out+'/'+name,Buffer.from(r.data,'base64'))}
try{
 let target;for(let i=0;i<100;i++){try{target=await(await fetch('http://127.0.0.1:9222/json/new?about:blank',{method:'PUT'})).json();break}catch{await delay(100)}}assert(target,'瀏覽器未啟動');
 ws=new WebSocket(target.webSocketDebuggerUrl);await new Promise((resolve,reject)=>{ws.onopen=resolve;ws.onerror=reject});
 ws.onmessage=e=>{const m=JSON.parse(e.data);if(m.id){const p=pending.get(m.id);if(p){pending.delete(m.id);m.error?p.reject(Error(JSON.stringify(m.error))):p.resolve(m.result)}}else if(m.method==='Runtime.exceptionThrown')errors.push(m.params);else if(m.method==='Network.requestWillBeSent')requests.push(m.params.request.url)};
 await call('Page.enable');await call('Runtime.enable');await call('Network.enable');
 await call('Emulation.setDeviceMetricsOverride',{width:1600,height:1100,deviceScaleFactor:1,mobile:false});await call('Page.navigate',{url:'file:///work/proposal.html'});
 let ready=false;for(let i=0;i<150;i++){if(await evaluate('Boolean(window.proposalReady)')){ready=true;break}await delay(100)}assert(ready,'HTML圖像未就緒');
 await evaluate('proposalControl.setRunning(false);proposalControl.setTime(1.5)');
 const geometry=await evaluate(`['alliance-row','empire-row'].map(id=>({id,facing:document.getElementById(id).dataset.facing,ships:[...document.querySelectorAll('#'+id+' .flagship')].map(g=>{const b=g.getBBox();return {key:g.dataset.key,near:Number(g.dataset.depth),facing:g.dataset.facing,hero:g.classList.contains('hero'),camera:g.dataset.camera,bandY:Number(g.dataset.bandY),bandH:Number(g.dataset.bandH),transform:g.getAttribute('transform'),bounds:{x:b.x,y:b.y,w:b.width,h:b.height}}})}))`);
 const l=await evaluate('proposalLayout');
 const viewport=await evaluate(`(()=>{const g=getComputedStyle(document.getElementById('game-canvas'));return {w:parseFloat(g.width),h:parseFloat(g.height)}})()`);assert(Math.abs(viewport.w-l.w)<.1&&Math.abs(viewport.h-l.h)<.1,'樣式覆蓋遊戲viewport幾何');
 for(const row of geometry){assert(row.ships.length===1,'上下必須各只有一艘');const s=row.ships[0],upper=row.id==='alliance-row';assert(row.facing===(upper?'left':'right')&&s.facing===row.facing,'單艦方向錯誤');assert(s.key===(upper?'46BD':'45B1'),'主旗艦錯誤');assert(s.hero,'主艦遺漏')}
 const widths=await evaluate(`[...document.querySelectorAll('.flagship')].map(g=>({key:g.dataset.key,width:Number(g.dataset.width),transform:g.getAttribute('transform')}))`);assert(widths.find(s=>s.key==='45B1').width>=l.w*.9,'伯倫希爾縮小過多');
 assert(await evaluate(`document.querySelectorAll('#junctions rect').length===4&&!document.querySelector('.escort')&&DATA.ships.length===0`),'漸層或隊列移除錯誤');
 const hull=await evaluate(`(async()=>{const g=document.querySelector('#empire-row .hero'),n=g.querySelector('image'),a=DATA.heroes[0],img=new Image();img.src=a.src;await img.decode();const canvas=document.createElement('canvas');canvas.width=img.naturalWidth;canvas.height=img.naturalHeight;const ctx=canvas.getContext('2d');ctx.drawImage(img,0,0);const pixels=ctx.getImageData(0,0,canvas.width,canvas.height).data,m=g.transform.baseVal.consolidate().matrix,scale=Number(n.getAttribute('width'))/canvas.width,x0=Number(n.getAttribute('x')),y0=Number(n.getAttribute('y'));let min=Infinity,max=-Infinity,count=0;for(let y=0;y<canvas.height;y++)for(let x=0;x<canvas.width;x++){if(pixels[(y*canvas.width+x)*4+3]<32)continue;const px=x0+x*scale,py=y0+y*scale,worldX=m.a*px+m.c*py+m.e;min=Math.min(min,worldX);max=Math.max(max,worldX);count++}const l=proposalLayout;return {min,max,count,safeLeft:l.left+18,safeRight:l.right-18,size:proposalControl.settings().bottomSize}})()`);assert(hull.size===92&&hull.count>0&&hull.min>=hull.safeLeft&&hull.max<=hull.safeRight,'船體仍受左右裁切或漸層覆蓋');observations.push({check:'實際alpha船體左右輪廓均避開裁切及漸層',passed:true,hull});
 const galaxy=await evaluate(`({fills:['top-bg','bottom-bg'].map(id=>document.getElementById(id).getAttribute('fill')),seams:[...document.querySelectorAll('#junctions rect')].map(n=>n.getAttribute('fill')),same:document.getElementById('galaxy-source').getAttribute('href')===DATA.portraits[0].src,shift:proposalControl.settings().bottomY})`);assert(galaxy.fills.every(x=>x==='url(#shared-galaxy)')&&galaxy.seams.every(x=>x==='url(#shared-galaxy)')&&galaxy.same&&galaxy.shift===24,'銀河共用或下移設定錯誤');observations.push({check:'上下與四角共用海報銀河，伯倫希爾下移24像素',passed:true,galaxy});
 const pose=await evaluate(`JSON.stringify([...document.querySelectorAll('.flagship,.flagship image,.portrait-panel image')].map(n=>[n.getAttribute('x'),n.getAttribute('y'),n.getAttribute('width'),n.getAttribute('transform')]))`);
 await evaluate('proposalControl.setTime(6.5)');assert(await evaluate(`JSON.stringify([...document.querySelectorAll('.flagship,.flagship image,.portrait-panel image')].map(n=>[n.getAttribute('x'),n.getAttribute('y'),n.getAttribute('width'),n.getAttribute('transform')]))`)===pose,'武器動畫移動本體');
 const locales=[];
 for(const locale of ['zh-TW','zh-CN','ja','en','ko']){
  await evaluate(`document.getElementById('locale').value='${locale}';proposalControl.render()`);
  const labels=await evaluate(`['empire','alliance'].map(key=>({key,quote:document.getElementById('quote-'+key).textContent,name:document.getElementById('name-'+key).textContent,expected:DATA.quotes['poster.name.'+key][proposalControl.settings().locale],bounds:['quote-','name-'].map(prefix=>{const b=document.getElementById(prefix+key).getBBox();return {x:b.x,y:b.y,w:b.width,h:b.height}})}))`);
  for(const label of labels){assert(label.name===label.expected&&label.quote.length,'五語文字錯誤');const x=label.key==='empire'?0:l.right;for(const b of label.bounds)assert(b.x>=x+3&&b.x+b.w<=x+l.left-3&&b.y>=490&&b.y+b.h<670,'文字超出人物圖片')}
  locales.push({locale,labels});
 }
 for(const [button,expected]of [['originalSize',100],['hullPriority',88]]){await evaluate(`document.getElementById('${button}').click()`);const g=await evaluate('({l:proposalLayout,scale:proposalControl.settings().gameScale})');assert(g.scale===expected&&Math.abs(g.l.w/g.l.h-4/3)<1e-9,'遊戲大小或4:3比例錯誤')}
 await evaluate(`document.getElementById('fade').value=0;proposalControl.render()`);assert(await evaluate(`!document.querySelector('#junctions rect')`),'漸層關閉無效');
 await evaluate(`document.getElementById('fade').value=70;document.getElementById('topSize').value=80;document.getElementById('bottomSize').value=110;document.getElementById('quoteY').value=610;proposalControl.render()`);assert(await evaluate(`document.querySelectorAll('.flagship').length===2&&document.querySelectorAll('#junctions rect').length===4&&document.getElementById('name-empire').getAttribute('y')==='640'`),'隊列、文字或漸層控制無效');
 await evaluate(`document.getElementById('weapons').value='missile';proposalControl.render();proposalControl.setTime(4)`);assert(await evaluate(`document.getElementById('top-weapons').childElementCount===1`),'飛彈切換無效');
 await evaluate(`document.getElementById('weapons').value='none';proposalControl.render()`);assert(await evaluate(`!document.getElementById('top-weapons').childElementCount&&!document.getElementById('bottom-weapons').childElementCount`),'關閉武器無效');
 const downloadDir=out+'/downloads-'+Date.now();await fs.mkdir(downloadDir);await call('Browser.setDownloadBehavior',{behavior:'allow',downloadPath:downloadDir});await evaluate(`document.getElementById('save').click()`);
 let file;for(let i=0;i<100;i++){file=(await fs.readdir(downloadDir)).find(n=>n.endsWith('.json'));if(file)break;await delay(100)}assert(file,'設定下載失敗');const saved=JSON.parse(await fs.readFile(downloadDir+'/'+file,'utf8'));assert(saved.topSize===80&&saved.bottomSize===110&&saved.bottomY===24&&saved.quoteY===610&&saved.gameScale===88&&saved.fade===70&&saved.bottom==='empire-right'&&saved.top==='alliance-left','下載設定錯誤');
 observations.push({check:'上下各一艘大型主旗艦、上左下右及本體固定',passed:true,geometry},{check:'五語圖內文字',passed:true,locales},{check:'四角漸層、4:3大小、單艦大小、武器與設定下載',passed:true,saved});
 await evaluate(`document.getElementById('reset').click();proposalControl.setRunning(false);proposalControl.setTime(1.5)`);
 const metrics=await call('Page.getLayoutMetrics'),size=metrics.cssContentSize||metrics.contentSize;await shot('proposal-page.png',{x:0,y:0,width:1600,height:Math.ceil(size.height),scale:1});
 const rect=await evaluate(`(()=>{const r=document.getElementById('scene').getBoundingClientRect();return {x:r.x,y:r.y,width:r.width,height:r.height,scale:1}})()`);await shot('proposal-frame.png',rect);
 await call('Emulation.setDeviceMetricsOverride',{width:390,height:844,deviceScaleFactor:1,mobile:true});await delay(150);assert(await evaluate('document.documentElement.scrollWidth<=innerWidth'),'手機頁面溢出');await shot('proposal-mobile.png');
 assert(!errors.length,'JavaScript例外');assert(!requests.some(x=>/^https?:/.test(x)),'外網依賴');
 const sha=bytes=>createHash('sha256').update(bytes).digest('hex');await fs.writeFile(out+'/BROWSER.json',JSON.stringify({prototype:true,passed:true,html_sha256:sha(await fs.readFile(out+'/proposal.html')),desktop:[1600,1100],mobile:[390,844],observations,errors,externalRequests:0},null,2)+'\n');
 console.log('PASS 艦尾與艦首實際輪廓安全／原有銀河、下移、單艦、五語、控制與手機');
}catch(error){await fs.writeFile(out+'/BROWSER-FAILED.json',JSON.stringify({error:String(error),errors},null,2)+'\n');if(ws?.readyState===1)try{await shot('browser-failed.png')}catch{}throw error
}finally{ws?.close();browser.kill('SIGTERM');await Promise.race([new Promise(r=>browser.once('exit',r)),delay(2000)]);if(browser.exitCode===null)browser.kill('SIGKILL');await log.close()}
