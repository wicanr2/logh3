// 156：沿既有Chromium，以原生CDP檢查本機HTML，不需npm套件。
import fs from 'node:fs/promises';
import {spawn} from 'node:child_process';
import {createHash} from 'node:crypto';
const out='/work',delay=ms=>new Promise(r=>setTimeout(r,ms));
const errors=[],requests=[];let ws,seq=0;const pending=new Map();
const log=await fs.open(out+'/browser.log','w');
const browser=spawn('/ms-playwright/chromium-1194/chrome-linux/chrome',['--headless=new','--no-sandbox','--disable-gpu','--disable-dev-shm-usage','--no-first-run','--hide-scrollbars','--force-color-profile=srgb','--force-device-scale-factor=1','--remote-debugging-port=9222','--remote-allow-origins=*','--user-data-dir=/tmp/poster156-chrome','about:blank'],{stdio:['ignore',log.fd,log.fd]});
const assert=(v,m)=>{if(!v)throw Error(m)};
async function call(method,params={}){const id=++seq;const result=new Promise((resolve,reject)=>pending.set(id,{resolve,reject}));ws.send(JSON.stringify({id,method,params}));return result}
async function evaluate(expression){const r=await call('Runtime.evaluate',{expression,returnByValue:true,awaitPromise:true});if(r.exceptionDetails)throw Error(JSON.stringify(r.exceptionDetails));return r.result.value}
async function shot(name,clip){const r=await call('Page.captureScreenshot',{format:'png',captureBeyondViewport:true,...(clip?{clip}:{})});await fs.writeFile(out+'/'+name,Buffer.from(r.data,'base64'))}
const observations=[];
try{
 let target;
 for(let n=0;n<100;n++){try{const r=await fetch('http://127.0.0.1:9222/json/new?about:blank',{method:'PUT'});target=await r.json();break}catch{await delay(100)}}
 assert(target,'Chromium未啟動');ws=new WebSocket(target.webSocketDebuggerUrl);
 await new Promise((resolve,reject)=>{ws.onopen=resolve;ws.onerror=reject});
 ws.onmessage=event=>{const m=JSON.parse(event.data);if(m.id){const p=pending.get(m.id);if(p){pending.delete(m.id);m.error?p.reject(Error(JSON.stringify(m.error))):p.resolve(m.result)}}else if(m.method==='Runtime.exceptionThrown')errors.push(m.params);else if(m.method==='Network.requestWillBeSent')requests.push(m.params.request.url)};
 await call('Page.enable');await call('Runtime.enable');await call('Network.enable');
 await call('Emulation.setDeviceMetricsOverride',{width:1600,height:1100,deviceScaleFactor:1,mobile:false});
 await call('Page.navigate',{url:'file:///work/proposal.html'});
 let ready=false;for(let i=0;i<150;i++){if(await evaluate('Boolean(window.proposalReady)')){ready=true;break}await delay(100)}assert(ready,'圖像未就緒');
 await evaluate('proposalControl.setRunning(false);proposalControl.setTime(1.5)');
 const initial=await evaluate(`({settings:proposalControl.settings(),images:[...document.querySelectorAll('#scene image')].map(n=>({x:n.getAttribute('x'),y:n.getAttribute('y'),w:n.getAttribute('width'),h:n.getAttribute('height'),transform:n.getAttribute('transform')})),svg:(()=>{const r=document.getElementById('scene').getBoundingClientRect();return {x:r.x,y:r.y,width:r.width,height:r.height}})(),pageWidth:document.documentElement.scrollWidth,viewport:innerWidth})`);
 assert(initial.settings.shipSize===720&&initial.settings.planeSize===360,'初始尺寸不符');
 assert(initial.images.filter(n=>Number(n.w)>=480).length>=7,'旗艦沒有維持大型圖像');
 assert(initial.images.filter(n=>Number(n.w)===360).length===2,'戰機尺寸不符');
 assert(initial.pageWidth===initial.viewport,'桌面頁面水平溢出');
 const staticBefore=JSON.stringify(initial.images);
 await evaluate('proposalControl.setTime(5.8)');
 const second=await evaluate(`({images:[...document.querySelectorAll('#scene image')].map(n=>({x:n.getAttribute('x'),y:n.getAttribute('y'),w:n.getAttribute('width'),h:n.getAttribute('height'),transform:n.getAttribute('transform')})),top:document.getElementById('top-weapons').childElementCount,bottom:document.getElementById('bottom-weapons').childElementCount})`);
 assert(JSON.stringify(second.images)===staticBefore,'本體在相位切換時移動');assert(second.bottom===2&&second.top===0,'光束位置未切換');
 observations.push({check:'大型圖像與固定本體',passed:true,images:initial.images});
 await evaluate(`document.getElementById('shipSize').value='950';document.getElementById('planeSize').value='520';proposalControl.render()`);
 assert(await evaluate(`document.querySelectorAll('#scene image[width="950"]').length===2&&document.querySelectorAll('#scene image[width="520"]').length===2`),'縮放控制無效');
 await evaluate(`document.getElementById('escorts').checked=false;proposalControl.render()`);
 assert(await evaluate(`document.querySelectorAll('#ships-left image,#ships-right image').length===2`),'後景開關無效');
 await evaluate(`document.getElementById('weapons').value='missile';proposalControl.render();proposalControl.setTime(3.4)`);
 assert(await evaluate(`document.getElementById('top-weapons').childElementCount===2`),'飛彈開關無效');
 await evaluate(`document.getElementById('weapons').value='none';proposalControl.render()`);
 assert(await evaluate(`!document.getElementById('top-weapons').childElementCount&&!document.getElementById('bottom-weapons').childElementCount`),'武器關閉無效');
 await evaluate(`document.getElementById('guidesToggle').checked=true;proposalControl.render()`);
 assert(await evaluate(`document.getElementById('scene').classList.contains('show-guides')`),'裁切示意無效');
 const downloadDir=out+'/downloads-'+Date.now();await fs.mkdir(downloadDir);await call('Browser.setDownloadBehavior',{behavior:'allow',downloadPath:downloadDir});
 await evaluate(`document.getElementById('save').click()`);
 let downloaded;for(let i=0;i<100;i++){downloaded=(await fs.readdir(downloadDir)).find(n=>n.endsWith('.json'));if(downloaded)break;await delay(100)}
 assert(downloaded,'設定下載失敗');const saved=JSON.parse(await fs.readFile(downloadDir+'/'+downloaded,'utf8'));
 assert(saved.shipSize===950&&saved.planeSize===520&&saved.bodies==='fixed','下載設定不符');
 observations.push({check:'縮放、後景、光束／飛彈／關閉、裁切示意及設定下載',passed:true,saved});
 await evaluate(`document.getElementById('reset').click();proposalControl.setRunning(false);proposalControl.setTime(1.5)`);
 const layout=await call('Page.getLayoutMetrics');const size=layout.cssContentSize||layout.contentSize;
 await shot('proposal-page.png',{x:0,y:0,width:1600,height:Math.ceil(size.height),scale:1});
 const rect=await evaluate(`(()=>{const r=document.getElementById('scene').getBoundingClientRect();return {x:r.x,y:r.y,width:r.width,height:r.height,scale:1}})()`);
 await shot('proposal-frame.png',rect);
 for(const y of [-80,-40,0,40]){await evaluate(`document.getElementById('shipY').value='${y}';proposalControl.render();proposalControl.setTime(1.5)`);await shot('crop-'+y+'.png',rect)}
 await evaluate(`document.getElementById('reset').click();proposalControl.setRunning(false)`);
 await evaluate('proposalControl.setTime(5.8)');await shot('proposal-frame-weapons.png',rect);
 await call('Emulation.setDeviceMetricsOverride',{width:390,height:844,deviceScaleFactor:1,mobile:true});
 await delay(200);assert(await evaluate('document.documentElement.scrollWidth<=innerWidth'),'手機頁面水平溢出');await shot('proposal-mobile.png');
 assert(errors.length===0,'JavaScript例外');assert(!requests.some(x=>/^https?:/.test(x)),'頁面依賴外網');
 const sha=bytes=>createHash('sha256').update(bytes).digest('hex');
 await fs.writeFile(out+'/BROWSER.json',JSON.stringify({passed:true,prototype:true,html_sha256:sha(await fs.readFile(out+'/proposal.html')),desktop:[1600,1100],mobile:[390,844],pageHeight:size.height,observations,errors,externalRequests:0},null,2)+'\n');
 process.stdout.write('PASS Chromium離線HTML、放大裁切、固定本體、武器切換、下載與手機版\n');
}catch(error){
 await fs.writeFile(out+'/BROWSER-FAILED.json',JSON.stringify({error:String(error),errors},null,2)+'\n');
 if(ws?.readyState===1){try{await shot('browser-failed.png');process.stderr.write(JSON.stringify(await evaluate(`({ready:window.proposalReady,status:document.getElementById('status')?.textContent,url:location.href})`))+'\n')}catch{}}
 throw error;
}finally{
 ws?.close();browser.kill('SIGTERM');await Promise.race([new Promise(r=>browser.once('exit',r)),delay(2000)]);if(browser.exitCode===null)browser.kill('SIGKILL');await log.close();
}
