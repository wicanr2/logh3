// 165：匯出已同意的四組新增名言五語透明層；原藝術及首組不改。
import fs from 'node:fs/promises';
import {spawn} from 'node:child_process';
import {createHash} from 'node:crypto';
const out='/work',delay=ms=>new Promise(r=>setTimeout(r,ms));
const errors=[],requests=[],pending=new Map(),observations=[];let ws,seq=0;
const log=await fs.open(out+'/browser.log','w');
const browser=spawn('/ms-playwright/chromium-1194/chrome-linux/chrome',['--headless=new','--no-sandbox','--disable-gpu','--disable-dev-shm-usage','--no-first-run','--hide-scrollbars','--force-color-profile=srgb','--force-device-scale-factor=1','--remote-debugging-port=9222','--remote-allow-origins=*','--user-data-dir=/tmp/poster164-export','about:blank'],{stdio:['ignore',log.fd,log.fd]});
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
 await fs.mkdir(out+'/art',{recursive:true});
 await call('Emulation.setDeviceMetricsOverride',{width:1280,height:720,deviceScaleFactor:2,mobile:false});
 async function exportLayer(kind){await evaluate(`(()=>{document.getElementById('export-scene')?.remove();const source=document.getElementById('scene'),svg=source.cloneNode(true);svg.id='export-scene';svg.setAttribute('width','1280');svg.setAttribute('height','720');svg.style.cssText='position:fixed;left:0;top:0;width:1280px;height:720px;display:block';const find=id=>svg.querySelector('#'+id);if('${kind}'==='full'){}else if('${kind}'==='base'){find('game-canvas')?.remove();svg.querySelectorAll('.flagship').forEach(n=>n.remove());find('portrait-labels').replaceChildren();find('top-weapons').replaceChildren();find('bottom-weapons').replaceChildren()}else{const defs=find('scene-defs')||svg.querySelector('defs');if('${kind}'==='hulls'){const rows=['alliance-row','empire-row'].map((id,i)=>{const g=document.createElementNS(NS,'g');g.setAttribute('clip-path','url(#'+(i?'bottom':'top')+'-clip)');g.append(find(id).cloneNode(true));return g});svg.replaceChildren(defs,...rows)}else svg.replaceChildren(defs,find('portrait-labels'))}find('guides')?.remove();const ids=new Map([...svg.querySelectorAll('[id]')].map(n=>[n.id,'bake-'+n.id]));for(const n of svg.querySelectorAll('*')){for(const attr of [...n.attributes]){if(attr.name==='id')continue;let v=attr.value;if(v.includes('url(#'))for(const [oldId,newId]of ids)v=v.split('url(#'+oldId+')').join('url(#'+newId+')');if(v.startsWith('#')&&ids.has(v.slice(1)))v='#'+ids.get(v.slice(1));if(v!==attr.value)n.setAttribute(attr.name,v)}if(n.id)n.id=ids.get(n.id)||n.id}document.querySelector('.wrap').style.display='none';document.body.style.setProperty('background','transparent','important');document.body.style.margin='0';document.documentElement.style.background='transparent';document.body.append(svg)})()`);await delay(80)}
 const source=await fs.readFile(out+'/quote-data.json'),data=JSON.parse(source);
 assert(data.status==='user-approved-2026-10-26'&&data.pairs.length===5&&data.interval_seconds===60,'名言未定案');
 await call('Emulation.setDefaultBackgroundColorOverride',{color:{r:0,g:0,b:0,a:0}});
 const sets=[],geometry=[];
 for(let pair=0;pair<5;pair++){
  const files={};
  for(const locale of ['zh-TW','zh-CN','ja','en','ko']){
   const item=data.pairs[pair];
   await evaluate(`(()=>{DATA.quotes['poster.quote.empire']=${JSON.stringify(item.reinhard)};DATA.quotes['poster.quote.alliance']=${JSON.stringify(item.yang)};document.getElementById('locale').value='${locale}';document.getElementById('posterSet').value='v1';proposalControl.render();for(const side of ['empire','alliance']){const q=document.getElementById('quote-'+side),n=document.getElementById('name-'+side),spans=[...q.querySelectorAll('tspan')],shift=Math.max(0,Number(spans.at(-1).getAttribute('y'))-(Number(n.getAttribute('y'))-21));if(shift)spans.forEach(s=>s.setAttribute('y',Number(s.getAttribute('y'))-shift))}})()`);
   await exportLayer('labels');
   const boxes=await evaluate(`['empire','alliance'].map(side=>{const q=document.getElementById('bake-quote-'+side).getBBox(),n=document.getElementById('bake-name-'+side).getBBox();return {side,x:q.x,y:q.y,w:q.width,h:q.height,gap:n.y-q.y-q.height,nameBottom:n.y+n.height}})`);
   for(const b of boxes)assert(b.y>=442&&b.y+b.h<638&&b.nameBottom<638&&b.gap>=1,'文字超過裁切或重疊：'+JSON.stringify({pair,locale,boxes}));
   geometry.push({pair:pair+1,locale,boxes});
   if(pair===0)continue;
   const file=`quote-${pair+1}-${locale}.png`;
   await shot('art/'+file,{x:0,y:440,width:1280,height:200,scale:1});
   const bytes=await fs.readFile(out+'/art/'+file);
   assert(bytes.readUInt32BE(16)===2560&&bytes.readUInt32BE(20)===400,'透明層尺寸錯誤');
   files[locale]={file,sha256:createHash('sha256').update(bytes).digest('hex')};
  }
  if(pair>0)sets.push({id:data.pairs[pair].id,files});
 }
 const sha=bytes=>createHash('sha256').update(bytes).digest('hex');
 assert(!errors.length&&!requests.some(x=>/^https?:/.test(x)),'瀏覽器例外或外網依賴');
 await fs.writeFile(out+'/art/quotes.json',JSON.stringify({format:'logh3-poster-quotes-v1',proposal_sha256:sha(await fs.readFile(out+'/proposal.html')),source_sha256:sha(source),interval_seconds:60,rect:[0,880,2560,1280],sets},null,2)+'\n');
 await fs.writeFile(out+'/EXPORT.json',JSON.stringify({passed:true,geometry,external_requests:0,source_sha256:sha(source),files:20},null,2)+'\n');
 console.log('PASS quote export: five groups/five locales; 20 cropped layers');
}catch(error){await fs.writeFile(out+'/EXPORT-FAILED.json',JSON.stringify({error:String(error),errors},null,2)+'\n');throw error}
finally{ws?.close();browser.kill('SIGTERM');await Promise.race([new Promise(r=>browser.once('exit',r)),delay(2000)]);if(browser.exitCode===null)browser.kill('SIGKILL');await log.close()}
