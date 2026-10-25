// 164：從凍結提案匯出正式藝術層及原生尺寸參考圖，不修改定版HTML。
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
 await call('Emulation.setDeviceMetricsOverride',{width:1600,height:1100,deviceScaleFactor:1,mobile:false});await call('Page.navigate',{url:'file:///work/quotes-preview.html'});
 let ready=false;for(let i=0;i<150;i++){if(await evaluate('Boolean(window.proposalReady)')){ready=true;break}await delay(100)}assert(ready,'HTML圖像未就緒');
 await evaluate('proposalControl.setRunning(false);proposalControl.setTime(1.5)');

 const rows=[];
 for(let pair=0;pair<5;pair++)for(const locale of ["zh-TW","zh-CN","ja","en","ko"]){
  await evaluate(`document.getElementById("locale").value="${locale}";quotePreview.setPair(${pair})`);await delay(30);
  const boxes=await evaluate(`(()=>["empire","alliance"].map(side=>{const q=document.getElementById("quote-"+side).getBBox(),n=document.getElementById("name-"+side).getBBox();return {side,x:q.x,y:q.y,w:q.width,h:q.height,nameY:n.y,gap:n.y-q.y-q.height}}))()`);
  for(const b of boxes)assert(b.y>0&&b.y+b.h<720&&b.gap>=1,"名言與署名重疊："+JSON.stringify({pair,locale,boxes}));rows.push({pair,locale,boxes});
 }
 await evaluate(`document.getElementById("locale").value="zh-TW";quotePreview.setPair(0)`);
 await delay(30500);await delay(30500);
 const selected=await evaluate(`document.getElementById("quote-pair").value`);assert(selected==="1","60秒未換到第二組："+selected);
 assert(!errors.length,"JavaScript例外");assert(!requests.some(x=>/^https?:/.test(x)),"外網依賴");
 await fs.writeFile(out+"/QUOTES-PREVIEW.json",JSON.stringify({passed:true,status:"approved-preview",rows,timer:{wall_wait_ms:61000,selected},errors,external_requests:0},null,2)+"\n");console.log("PASS quote preview 25 pairs/locales; 60s timer");
}catch(error){await fs.writeFile(out+"/QUOTES-FAILED.json",JSON.stringify({error:String(error),errors},null,2)+"\n");throw error}
finally{ws?.close();browser.kill("SIGTERM");await Promise.race([new Promise(r=>browser.once("exit",r)),delay(2000)]);if(browser.exitCode===null)browser.kill("SIGKILL");await log.close()}
