#!/usr/bin/env python3
"""在既有隔離Chromium中核對四套概念的切換、圖片、點擊框與對齊線。"""
import base64
import json
import os
from pathlib import Path
import socket
import struct
import subprocess
import time
import urllib.request
import urllib.parse

ROOT = Path('/work/hd126')
assert os.getuid() == 1000 and ROOT.stat().st_uid == 1000
source = ROOT/'FACTION-HD-CONCEPTS.html'
test = ROOT/'CONCEPT-BROWSER-CHECK.html'
check = r'''<script>
(async()=>{
 document.body.dataset.checkStage='start';
 const checks=[];const assert=(ok,what)=>{if(!ok)throw Error(what);checks.push(what)};
 assert(document.querySelectorAll('.card').length===4,'四套陣營概念');
 assert(document.querySelectorAll('[data-scene]').length===4,'四個操作畫面');
 assert(chosen[0]==='empire-b'&&chosen[1]==='alliance-a','已確認的帝國B與同盟A');
 assert(themes['empire-b'].surface==='#1c5393'&&themes['alliance-a'].surface==='#e7d7a3','亮藍與淡黃修訂');
 for(const id of Object.keys(themes))for(const scene of Object.keys(scenes)){
  document.querySelector('[data-theme="'+id+'"]').click();
  document.querySelector('[data-scene="'+scene+'"]').click();
  await $('preview').decode();
  document.body.dataset.checkStage=id+'-'+scene;
  assert(active===id&&view===scene,'切換 '+id+' '+scene);
  assert($('preview').naturalWidth===1920&&$('preview').naturalHeight===1440,'HD圖片 '+id+' '+scene);
  const r=$('screen').getBoundingClientRect();assert(Math.abs(r.width/r.height-4/3)<.0001,'4:3 '+id+' '+scene);
  for(const b of document.querySelectorAll('.hit')){
   const [x,y,w,h]=b.dataset.rect.split(',').map(Number),p=b.getBoundingClientRect();
   assert(x>=0&&y>=0&&x+w<=640&&y+h<=480,'按鈕畫布安全');
   assert(Math.abs(p.left-r.left-x*r.width/640)<.1&&Math.abs(p.top-r.top-y*r.height/480)<.1,'按鈕位置對齊');
   assert(Math.abs(p.width-w*r.width/640)<.1&&Math.abs(p.height-h*r.height/480)<.1,'按鈕外框尺寸');
  }
 }
 setTheme('empire-a');setScene('budget');$('guides').click();
 assert(getComputedStyle($('guide-layer')).display!=='none','對齊線可顯示');
 $('guides').click();assert(getComputedStyle($('guide-layer')).display==='none','對齊線可關閉');
 const old=$('preview').src;$('original').click();await $('preview').decode();
 assert($('preview').src!==old,'原配色對照');$('original').click();
 const hit=document.querySelector('.hit');assert(!!hit,'財政可試點');hit.click();assert($('status').textContent.startsWith('已試點'),'按鈕點擊回饋');
 $('zoom').click();assert($('screen').classList.contains('zoom'),'原尺寸檢查');
 document.dispatchEvent(new KeyboardEvent('keydown',{key:'Escape'}));assert(!$('screen').classList.contains('zoom'),'Esc返回適合視窗');
 $('choose').click();setTheme('alliance-b');$('choose').click();
 assert(chosen[0]==='empire-a'&&chosen[1]==='alliance-b','陣營可分別選擇');
 assert(document.documentElement.scrollWidth<=innerWidth,'整頁無水平溢出');
 chosen[0]='empire-b';chosen[1]='alliance-a';
 const p=new URLSearchParams(location.search);setTheme(p.get('theme')||'empire-b');setScene(p.get('scene')||'budget');
 document.body.dataset.checkStage='complete';
 const pre=document.createElement('pre');pre.id='check-result';pre.hidden=true;
 pre.textContent=JSON.stringify({pass:true,width:innerWidth,checks:checks.length,theme:active,scene:view});document.body.appendChild(pre);
})().catch(error=>{const pre=document.createElement('pre');pre.id='check-result';pre.textContent=JSON.stringify({pass:false,error:String(error)});document.body.appendChild(pre)})
</script>'''
test.write_text(source.read_text().replace('</html>', check+'</html>'))

# data URL的圖片解碼不服從headless虛擬時間；--dump-dom可能比完成訊號早。
# 用Chrome原生偵錯協定等實際頁面完成，僅連容器內loopback，不加未鎖版套件。
class CDP:
    def __init__(self, address):
        u = urllib.parse.urlparse(address)
        self.sock = socket.create_connection((u.hostname, u.port), timeout=20)
        key = base64.b64encode(os.urandom(16)).decode()
        self.sock.sendall((f'GET {u.path} HTTP/1.1\r\nHost: localhost:{u.port}\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: {key}\r\nSec-WebSocket-Version: 13\r\n\r\n').encode())
        self.stream = self.sock.makefile('rb')
        assert b'101' in self.stream.readline()
        while self.stream.readline() != b'\r\n':
            pass
        self.ident = 0

    def call(self, method, params=None):
        self.ident += 1
        payload = json.dumps({'id': self.ident, 'method': method, 'params': params or {}}).encode()
        n = len(payload)
        prefix = bytes([0x81, 0x80 | n]) if n < 126 else bytes([0x81, 0xFE])+struct.pack('>H', n)
        mask = os.urandom(4)
        self.sock.sendall(prefix+mask+bytes(c ^ mask[i % 4] for i, c in enumerate(payload)))
        while True:
            first, second = self.stream.read(2)
            length = second & 127
            if length == 126:
                length = struct.unpack('>H', self.stream.read(2))[0]
            elif length == 127:
                length = struct.unpack('>Q', self.stream.read(8))[0]
            value = self.stream.read(length)
            assert first & 0x80 and first & 15 == 1, '只接受完整CDP文字訊息'
            message = json.loads(value)
            if message.get('id') == self.ident:
                assert 'error' not in message, message
                return message.get('result', {})


command = ['chromium', '--headless', '--no-sandbox', '--disable-dev-shm-usage', '--disable-gpu',
           '--renderer-process-limit=2', '--disable-extensions', '--no-first-run', '--no-default-browser-check',
           '--hide-scrollbars', '--user-data-dir=/tmp/logh3-concepts-real', '--remote-debugging-port=9222', 'about:blank']
results = []
with (ROOT/'concept-browser.log').open('w') as log:
    browser = subprocess.Popen(command, stdout=log, stderr=subprocess.STDOUT)
    try:
        deadline = time.monotonic()+15
        while True:
            try:
                with urllib.request.urlopen('http://127.0.0.1:9222/json', timeout=1) as response:
                    targets = json.load(response)
                break
            except OSError:
                assert time.monotonic() < deadline, '瀏覽器未啟動'
                time.sleep(.1)
        cdp = CDP(next(t['webSocketDebuggerUrl'] for t in targets if t['type'] == 'page'))
        cdp.call('Page.enable')
        for width, height, theme, scene in [(1440, 1700, 'empire-b', 'budget'), (1440, 1700, 'alliance-a', 'budget'), (1440, 1700, 'alliance-a', 'settings'), (390, 1600, 'alliance-a', 'budget')]:
            name = f'concept-{width}-{theme}-{scene}'
            cdp.call('Emulation.setDeviceMetricsOverride', {'width': width, 'height': height, 'deviceScaleFactor': 1, 'mobile': False})
            cdp.call('Page.navigate', {'url': test.as_uri()+f'?theme={theme}&scene={scene}'})
            deadline = time.monotonic()+45
            while True:
                reply = cdp.call('Runtime.evaluate', {'expression': 'document.querySelector("#check-result")?.textContent || ""', 'returnByValue': True})
                value = reply.get('result', {}).get('value')
                if value:
                    result = json.loads(value)
                    break
                assert time.monotonic() < deadline, '頁面未回報完成'
                time.sleep(.1)
            assert result['pass'], result
            # 圖片已逐一解碼通過；最後的切換再等兩個正常渲染frame。
            cdp.call('Runtime.evaluate', {'expression': 'new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r)))', 'awaitPromise': True})
            shot = cdp.call('Page.captureScreenshot', {'format': 'png', 'captureBeyondViewport': False})
            (ROOT/(name+'.png')).write_bytes(base64.b64decode(shot['data']))
            print(name, result, flush=True)
            results.append(result)
        cdp.sock.close()
    finally:
        browser.terminate()
        try:
            browser.wait(timeout=10)
        except subprocess.TimeoutExpired:
            browser.kill()
            browser.wait(timeout=5)
(ROOT/'CONCEPT-BROWSER-RESULTS.json').write_text(json.dumps(results, ensure_ascii=False, indent=2)+'\n')
