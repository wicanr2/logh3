"""以隔離Chromium驗HD試樣排版與可操作性；不連接正式遊戲。"""
import json
import os
from pathlib import Path
import subprocess

ROOT = Path('/work/hd123-s1-20261011/hd')
assert os.getuid() == 1000 and ROOT.stat().st_uid == 1000
(ROOT/'TOOLCHAIN.json').write_text(json.dumps({
    'image':os.environ.get('LOGH3_PREVIEW_IMAGE','logh3-hd-preview:20261011-r1'),
    'base_image':'eob-audio-capture:20261009-r3@sha256:d831720f8dfa388af8ad175fa8adeae8f2814dd4637be1e3db6a43274b26f36c',
    'downloads':json.loads(Path('/opt/debian-downloads-receipt.json').read_text()),
    'installed_packages':Path('/usr/local/share/hd-preview-packages.txt').read_text(),
    'archive_location':'映像內/var/cache/apt/archives保留精確deb，依下載收據SHA-256核對'
},ensure_ascii=False,indent=2)+'\n')
browser = ['chromium', '--headless', '--no-sandbox', '--disable-dev-shm-usage',
           '--disable-gpu', '--no-first-run', '--no-default-browser-check',
           '--hide-scrollbars', '--user-data-dir=/tmp/logh3-hd-browser']
script = r'''<script>
document.fonts.ready.then(()=>{
 const checks=[];const assert=(ok,what)=>{if(!ok)throw Error(what);checks.push(what)};
 assert(document.querySelectorAll('.stepper button').length===4,'預算比例四個按鈕');
 let old=state.tax;document.querySelector('[aria-label="增加徵稅比例"]').click();assert(state.tax===old+5,'徵稅增加');
 document.querySelector('[aria-label="減少徵稅比例"]').click();assert(state.tax===old,'徵稅減少');
 document.getElementById('tab-build').click();assert(document.querySelectorAll('.stepper button').length===14,'七種艦型十四個按鈕');
 names.forEach((n,i)=>{const old=state.pct[i];document.querySelector('[aria-label="增加'+n+'比例"]').click();assert(state.pct[i]===old+1,n+'增加');document.querySelector('[aria-label="減少'+n+'比例"]').click();assert(state.pct[i]===old,n+'減少')});
 step('pct',6,1);save();reset();load();assert(state.pct[6]===1,'保存重設再讀回');
 staff();assert(document.getElementById('status').textContent.startsWith('參謀'),'參謀說明');
 finish();assert(current==='account','結束返回報告');page('build');document.dispatchEvent(new KeyboardEvent('keydown',{key:'Escape'}));assert(current==='account','Esc返回');
 const params=new URLSearchParams(location.search);theme(params.get('theme')==='clear'?'clear':'classic');page(params.get('page')||'budget');
 assert(document.documentElement.scrollWidth<=window.innerWidth,'整頁無水平溢出');
 for(const e of document.querySelectorAll('.row, .sheet-top, .buttons, nav, header'))assert(e.scrollWidth<=e.clientWidth+1,'版面安全：'+e.className);
 for(const b of document.querySelectorAll('.stepper button')){const r=b.getBoundingClientRect();assert(r.width>=44&&r.height>=44,'箭頭44px命中區');assert(r.right<=window.innerWidth,'箭頭在視窗內')}
 const result=document.createElement('pre');result.id='browser-result';result.hidden=true;result.textContent=JSON.stringify({pass:true,checks,width:innerWidth,theme:document.body.dataset.theme,page:current});document.body.appendChild(result);
}).catch(error=>{const result=document.createElement('pre');result.id='browser-result';result.textContent=JSON.stringify({pass:false,error:String(error)});document.body.appendChild(result)})
</script>'''
source = (ROOT/'HD-INTERFACE.html').read_text()
test = ROOT/'BROWSER-CHECK.html'
test.write_text(source.replace('</style>','*{transition:none!important;animation:none!important}</style>').replace('</html>',script+'</html>'))
results=[]
for width, height, theme, page in [(1440,1140,'classic','budget'),(1440,1140,'clear','build'),(800,1100,'classic','build'),(390,1400,'clear','build')]:
    label=f'{width}-{theme}-{page}'
    args=browser+[f'--window-size={width},{height}', '--virtual-time-budget=4000',
                  f'--screenshot={ROOT}/{label}.png','--dump-dom',test.as_uri()+f'?theme={theme}&page={page}']
    run=subprocess.run(args, capture_output=True,text=True,timeout=40,check=True)
    (ROOT/(label+'.log')).write_text(run.stderr)
    import re
    found=re.search(r'<pre id="browser-result"[^>]*>(.*?)</pre>',run.stdout)
    assert found, run.stderr[-1500:]
    result=json.loads(found[1].replace('&quot;','"').replace('&amp;','&'))
    results.append(result)
    print(label,result.get('pass'),result.get('error',''),flush=True)
    assert result['pass'],result
test.unlink()
(ROOT/'BROWSER.json').write_text(json.dumps({'browser':subprocess.check_output(['chromium','--version'],text=True).strip(),'checks':results,'visual_review':'待目視兩款大圖與小視窗'},ensure_ascii=False,indent=2)+'\n')
