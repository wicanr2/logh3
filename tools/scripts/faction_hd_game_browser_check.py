"""兩陣營正式畫面HTML的真實瀏覽器切換與窄螢幕核對。"""
import ast
import base64
import json
import os
from pathlib import Path
import socket
import struct
import subprocess
import time
import urllib.parse
import urllib.request

root=Path('/work/hd126')
assert os.getuid()==1000 and root.stat().st_uid==1000
source=Path('/src/tools/scripts/faction_hd_concept_check.py')
tree=ast.parse(source.read_text())
definition=next(n for n in tree.body if isinstance(n,ast.ClassDef) and n.name=='CDP')
exec(compile(ast.Module(body=[definition],type_ignores=[]),str(source),'exec'))
check=r'''(async()=>{
 const checks=[];const assert=(ok,msg)=>{if(!ok)throw Error(msg);checks.push(msg)};
 assert(document.querySelectorAll('[data-faction]').length===2,'兩個陣營');
 assert(document.querySelectorAll('[data-scene]').length===4,'四個畫面');
 for(const faction of ['empire','alliance'])for(const scene of ['strategy','budget','build','settings']){
  document.querySelector('[data-faction="'+faction+'"]').click();
  document.querySelector('[data-scene="'+scene+'"]').click();
  const im=document.getElementById('preview');await im.decode();
  assert(im.naturalWidth===1920&&im.naturalHeight===1440,'HD尺寸 '+faction+scene);
  assert(im.alt.includes(faction==='empire'?'亮藍':'赤焰戰線'),'最新配色 '+faction+scene);
  assert(document.documentElement.scrollWidth<=innerWidth,'沒有水平溢出 '+faction+scene);
  assert(document.querySelector('[data-faction="'+faction+'"]').getAttribute('aria-pressed')==='true','按鈕切換 '+faction+scene);
 }
 return {pass:true,width:innerWidth,checks:checks.length};
})()'''
command=['chromium','--headless','--no-sandbox','--disable-dev-shm-usage','--disable-gpu',
 '--renderer-process-limit=2','--disable-extensions','--no-first-run','--no-default-browser-check',
 '--hide-scrollbars','--user-data-dir=/tmp/logh3-faction-r7','--remote-debugging-port=9222','about:blank']
results=[]
with (root/'faction-game-browser-r7.log').open('x') as log:
 browser=subprocess.Popen(command,stdout=log,stderr=subprocess.STDOUT)
 try:
  deadline=time.monotonic()+20
  while True:
   try:
    with urllib.request.urlopen('http://127.0.0.1:9222/json',timeout=1) as response:targets=json.load(response)
    break
   except OSError:
    assert time.monotonic()<deadline
    time.sleep(.1)
  cdp=CDP(next(t['webSocketDebuggerUrl'] for t in targets if t['type']=='page'))
  cdp.call('Page.enable')
  for width,height in [(1440,1700),(390,1600)]:
   cdp.call('Emulation.setDeviceMetricsOverride',dict(width=width,height=height,deviceScaleFactor=1,mobile=False))
   cdp.call('Page.navigate',{'url':(root/'FACTION-HD-GAME.html').as_uri()})
   deadline=time.monotonic()+30
   while True:
    reply=cdp.call('Runtime.evaluate',{'expression':'!!document.querySelector("[data-faction]")','returnByValue':True})
    if reply.get('result',{}).get('value'):break
    assert time.monotonic()<deadline;time.sleep(.1)
   reply=cdp.call('Runtime.evaluate',{'expression':check,'awaitPromise':True,'returnByValue':True})
   assert 'exceptionDetails' not in reply,reply
   result=reply['result']['value'];assert result['pass'] and result['checks']==34
   results.append(result)
   for faction in ('empire','alliance'):
    cdp.call('Runtime.evaluate',{'expression':f'document.querySelector("[data-faction={faction}]").click();document.querySelector("[data-scene=strategy]").click()'})
    cdp.call('Runtime.evaluate',{'expression':'document.getElementById("preview").decode().then(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))))','awaitPromise':True})
    shot=cdp.call('Page.captureScreenshot',{'format':'png','captureBeyondViewport':False})
    (root/f'faction-game-{width}-{faction}-r7.png').write_bytes(base64.b64decode(shot['data']))
   print(result,flush=True)
  cdp.sock.close()
 finally:
  browser.terminate()
  try:browser.wait(timeout=10)
  except subprocess.TimeoutExpired:browser.kill();browser.wait(timeout=5)
(root/'FACTION-GAME-BROWSER-RESULTS-r7.json').write_text(json.dumps(results,ensure_ascii=False,indent=2)+'\n')
