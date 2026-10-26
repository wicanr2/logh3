// 157：驗上下艦隊方向、景深與圖片內文字，使用既有Chromium及原生CDP。
import fs from 'node:fs/promises';
import {spawn} from 'node:child_process';
import {createHash} from 'node:crypto';
const out='/work',delay=ms=>new Promise(r=>setTimeout(r,ms));
const errors=[],requests=[],pending=new Map();let ws,seq=0;
const log=await fs.open(out+'/browser.log','w');
const browser=spawn('/ms-playwright/chromium-1194/chrome-linux/chrome',['--headless=new','--no-sandbox','--disable-gpu','--disable-dev-shm-usage','--no-first-run','--hide-scrollbars','--force-color-profile=srgb','--force-device-scale-factor=1','--remote-debugging-port=9222','--remote-allow-origins=*','--user-data-dir=/tmp/poster157-chrome','about:blank'],{stdio:['ignore',log.fd,log.fd]});
const assert=(v,m)=>{if(!v)throw Error(m)};
async function call(method,params={}){const id=++seq;const promise=new Promise((resolve,reject)=>pending.set(id,{resolve,reject}));ws.send(JSON.stringify({id,method,params}));return promise}
async function evaluate(expression){const r=await call('Runtime.evaluate',{expression,returnByValue:true,awaitPromise:true});if(r.exceptionDetails)throw Error(JSON.stringify(r.exceptionDetails));return r.result.value}
async function shot(name,clip){const r=await call('Page.captureScreenshot',{format:'png',captureBeyondViewport:true,...(clip?{clip}:{})});await fs.writeFile(out+'/'+name,Buffer.from(r.data,'base64'))}
const observations=[];
try{
 let target;for(let i=0;i<100;i++){try{target=await(await fetch('http://127.0.0.1:9222/json/new?about:blank',{method:'PUT'})).json();break}catch{await delay(100)}}assert(target,'瀏覽器未啟動');
 ws=new WebSocket(target.webSocketDebuggerUrl);await new Promise((resolve,reject)=>{ws.onopen=resolve;ws.onerror=reject});
 ws.onmessage=e=>{const m=JSON.parse(e.data);if(m.id){const p=pending.get(m.id);if(p){pending.delete(m.id);m.error?p.reject(Error(JSON.stringify(m.error))):p.resolve(m.result)}}else if(m.method==='Runtime.exceptionThrown')errors.push(m.params);else if(m.method==='Network.requestWillBeSent')requests.push(m.params.request.url)};
 await call('Page.enable');await call('Runtime.enable');await call('Network.enable');
 await call('Emulation.setDeviceMetricsOverride',{width:1600,height:1100,deviceScaleFactor:1,mobile:false});await call('Page.navigate',{url:'file:///work/proposal.html'});
 let ready=false;for(let i=0;i<150;i++){if(await evaluate('Boolean(window.proposalReady)')){ready=true;break}await delay(100)}assert(ready,'HTML圖像未就緒');
 await evaluate('proposalControl.setRunning(false);proposalControl.setTime(1.5)');
 const geometry=await evaluate(`['alliance-row','empire-row'].map(id=>({id,facing:document.getElementById(id).dataset.facing,ships:[...document.querySelectorAll('#'+id+' .flagship')].map(g=>({key:g.dataset.key,near:Number(g.dataset.depth),w:Number(g.firstElementChild.getAttribute('width')),x:Number(g.firstElementChild.getAttribute('x')),y:Number(g.firstElementChild.getAttribute('y')),transform:g.firstElementChild.getAttribute('transform')}))}))`);
 for(const row of geometry){assert(row.ships.length===6,'旗艦數量錯誤');for(let i=1;i<row.ships.length;i++)assert(row.ships[i].near>row.ships[i-1].near&&row.ships[i].w>row.ships[i-1].w,'景深沒有由遠到近展開');assert(row.ships.every(s=>row.id==='alliance-row'?s.key.startsWith('46')&&s.transform?.includes('scale(-1 1)'):s.key.startsWith('45')&&!s.transform),'艦隊或朝向錯誤')}
 assert(await evaluate(`!('fighters' in DATA)&&!document.querySelector('#scene [data-kind="fighter"]')`),'戰機未移除');
 const pose=await evaluate(`JSON.stringify([...document.querySelectorAll('.flagship image,.portrait-panel image')].map(n=>[n.getAttribute('x'),n.getAttribute('y'),n.getAttribute('width'),n.getAttribute('transform')]))`);
 await evaluate('proposalControl.setTime(6.5)');assert(await evaluate(`JSON.stringify([...document.querySelectorAll('.flagship image,.portrait-panel image')].map(n=>[n.getAttribute('x'),n.getAttribute('y'),n.getAttribute('width'),n.getAttribute('transform')]))`)===pose,'武器動畫移動了本體');
 const locales=[];
 for(const locale of ['zh-TW','zh-CN','ja','en','ko']){
  await evaluate(`document.getElementById('locale').value='${locale}';proposalControl.render()`);
  const labels=await evaluate(`['empire','alliance'].map(key=>({key,quote:document.getElementById('quote-'+key).textContent,name:document.getElementById('name-'+key).textContent,expected:DATA.quotes['poster.name.'+key][proposalControl.settings().locale],bounds:['quote-','name-'].map(prefix=>{const b=document.getElementById(prefix+key).getBBox();return {x:b.x,y:b.y,w:b.width,h:b.height}})}))`);
  for(const label of labels){assert(label.name===label.expected&&label.quote.length,'名言或姓名語系錯誤');const left=label.key==='empire'?0:1120;for(const b of label.bounds)assert(b.x>=left+3&&b.x+b.w<=left+157&&b.y>=490&&b.y+b.h<670,'文字沒有完整放進人物圖片')}
  locales.push({locale,labels});
 }
 observations.push({check:'上同盟朝左、下帝國朝右、六艘景深與本體固定',passed:true,geometry});observations.push({check:'五語名言與姓名在人物圖片內',passed:true,locales});
 await evaluate(`document.getElementById('count').value=4;document.getElementById('size').value=620;document.getElementById('depth').value=60;document.getElementById('quoteY').value=610;proposalControl.render()`);
 assert(await evaluate(`document.querySelectorAll('.flagship').length===8&&document.getElementById('name-empire').getAttribute('y')==='640'`),'隊列或文字位置控制無效');
 await evaluate(`document.getElementById('weapons').value='missile';proposalControl.render();proposalControl.setTime(4)`);assert(await evaluate(`document.getElementById('top-weapons').childElementCount===1`),'飛彈切換無效');
 await evaluate(`document.getElementById('weapons').value='none';proposalControl.render()`);assert(await evaluate(`!document.getElementById('top-weapons').childElementCount&&!document.getElementById('bottom-weapons').childElementCount`),'關閉武器無效');
 const downloadDir=out+'/downloads-'+Date.now();await fs.mkdir(downloadDir);await call('Browser.setDownloadBehavior',{behavior:'allow',downloadPath:downloadDir});await evaluate(`document.getElementById('save').click()`);
 let file;for(let i=0;i<100;i++){file=(await fs.readdir(downloadDir)).find(n=>n.endsWith('.json'));if(file)break;await delay(100)}assert(file,'設定下載失敗');const saved=JSON.parse(await fs.readFile(downloadDir+'/'+file,'utf8'));assert(saved.count===4&&saved.quoteY===610&&saved.top==='alliance-left'&&saved.bottom==='empire-right'&&saved.portraitText==='inside-image','下載設定錯誤');
 observations.push({check:'隊列、文字位置、武器及下載',passed:true,saved});
 await evaluate(`document.getElementById('reset').click();proposalControl.setRunning(false);proposalControl.setTime(1.5)`);
 const layout=await call('Page.getLayoutMetrics'),size=layout.cssContentSize||layout.contentSize;await shot('proposal-page.png',{x:0,y:0,width:1600,height:Math.ceil(size.height),scale:1});
 const rect=await evaluate(`(()=>{const r=document.getElementById('scene').getBoundingClientRect();return {x:r.x,y:r.y,width:r.width,height:r.height,scale:1}})()`);await shot('proposal-frame.png',rect);
 await call('Emulation.setDeviceMetricsOverride',{width:390,height:844,deviceScaleFactor:1,mobile:true});await delay(150);assert(await evaluate('document.documentElement.scrollWidth<=innerWidth'),'手機頁面溢出');await shot('proposal-mobile.png');
 assert(!errors.length,'JavaScript例外');assert(!requests.some(x=>/^https?:/.test(x)),'外網依賴');
 const sha=bytes=>createHash('sha256').update(bytes).digest('hex');await fs.writeFile(out+'/BROWSER.json',JSON.stringify({prototype:true,passed:true,html_sha256:sha(await fs.readFile(out+'/proposal.html')),desktop:[1600,1100],mobile:[390,844],observations,errors,externalRequests:0},null,2)+'\n');
 console.log('PASS 旗艦方向／景深、固定本體、五語圖內名言姓名、控制與設定下載');
}catch(error){await fs.writeFile(out+'/BROWSER-FAILED.json',JSON.stringify({error:String(error),errors},null,2)+'\n');if(ws?.readyState===1)try{await shot('browser-failed.png')}catch{}throw error
}finally{ws?.close();browser.kill('SIGTERM');await Promise.race([new Promise(r=>browser.once('exit',r)),delay(2000)]);if(browser.exitCode===null)browser.kill('SIGKILL');await log.close()}
