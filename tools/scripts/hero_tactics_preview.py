"""第130輪可撤回戰術示意。只產生HTML，不接正式遊戲或存檔。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--out', default='/src/workplace/hero130')
parser.add_argument('--check', action='store_true')
args = parser.parse_args()

ROOT = Path('/src')
OUT = Path(args.out)
assert os.getuid() == 1000 and OUT.parent.stat().st_uid == os.getuid()
OUT.mkdir(exist_ok=True)
assert OUT.stat().st_uid == os.getuid()
names = json.loads((ROOT/'docs/internals/data/keireki.json').read_text())['records']
assert names[0]['name'] == '萊因哈特' and names[52]['name'] == '楊威利'
html = r'''<!doctype html><html lang="zh-Hant"><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>名將戰術與移動加成試樣</title>
<style>
*{box-sizing:border-box}body{margin:0;background:#080f1e;color:#e9dfc2;font:17px/1.65 system-ui,sans-serif}
main{max-width:1350px;margin:auto;padding:24px}h1{font-size:30px;margin:0 0 12px}p{margin:8px 0}
.note{color:#b8c7d7}.controls{display:flex;flex-wrap:wrap;gap:16px;padding:20px 0}
select,button{font:inherit;background:#173f70;color:#fff0c9;border:1px solid #82a4cb;padding:8px;cursor:pointer}
label{display:flex;align-items:center;gap:9px;max-width:100%;flex-wrap:wrap}select{max-width:100%}.panels{display:grid;grid-template-columns:1fr 1fr;gap:16px}
section{border:1px solid #6a839f;background:#101e31;padding:12px}h2{font-size:20px;margin:0 0 8px}
canvas{width:100%;height:auto;display:block;background:#060c18}.legend{display:flex;gap:24px}
.legend b{color:#74b4f5}.legend i{color:#ee8564;font-style:normal}output{display:block;min-height:54px}
@media(max-width:850px){.panels{grid-template-columns:1fr}main{padding:14px}}
</style><main><h1>名將參戰，戰局如何改變？</h1>
<p class="note">可撤回規則示意，尚未套入遊戲。這頁不是遊戲錄影，圖示與路線不作正式驗收。</p>
<div class="controls">
<label>敵方指揮官<select id="hero"><option value="yang">楊威利：牽制與兩翼包抄</option><option value="reinhard">萊因哈特：集中與側後封鎖</option><option value="none">沒有名將</option></select></label>
<label>加速範圍<select id="scope"><option value="hero">英雄陣營</option><option value="both">戰場雙方</option></select></label>
<label>移動提案<select id="speed"><option value="1.25">1.25倍</option><option value="1.5" selected>1.5倍</option><option value="2">2倍</option></select></label>
<button id="restart">重播</button><button id="pause">暫停</button></div>
<div class="legend"><b>藍色：玩家艦隊</b><i>赤紅：敵方艦隊</i><span>虛線：預定航路</span></div>
<div class="panels"><section><h2>現行強化：通用包圍</h2><canvas id="baseline" width="640" height="480"></canvas><output id="baseText"></output></section>
<section><h2>提案：名將分翼協同</h2><canvas id="proposal" width="640" height="480"></canvas><output id="proposalText"></output></section></div>
<p>加速指每回合可移動距離增加。戰鬥動畫、音效與結果等待沿普通速度，不加快影片掩蓋差異。</p>
<p>提案只套用強化模式。依存活在場的帶隊名將觸發，離場即取消；原版模式沿既有規則。</p>
<script>
'use strict';
const $=id=>document.getElementById(id), controls=['hero','scope','speed'];
let start=performance.now(), frozen=null, pauseAt=0;
const red=[[510,118],[540,178],[520,242],[550,306],[510,370]],blue=[[164,196],[210,237],[155,284]];
function targets(proposal){
 const mode=$('hero').value;
 if(!proposal||mode==='none')return red.map((_,i)=>[210+Math.cos(i*2*Math.PI/5)*95,240+Math.sin(i*2*Math.PI/5)*95]);
 if(mode==='yang')return [[310,240],[218,102],[122,128],[206,378],[120,350]];
 return [[280,215],[280,266],[132,150],[130,330],[120,240]];
}
function ship(c,x,y,colour,angle,n){c.save();c.translate(x,y);c.rotate(angle);c.fillStyle=colour;c.beginPath();c.moveTo(11,0);c.lineTo(-7,-7);c.lineTo(-4,0);c.lineTo(-7,7);c.closePath();c.fill();c.restore();c.fillStyle=colour;c.font='15px sans-serif';c.fillText(n,x-5,y+24);}
function draw(id,proposal,t){
 const c=$(id).getContext('2d'),hero=$('hero').value, enabled=proposal&&hero!=='none', multiplier=enabled?Number($('speed').value):1;
 c.clearRect(0,0,640,480);c.strokeStyle='#15263c';c.lineWidth=1;
 for(let x=0;x<640;x+=40){c.beginPath();c.moveTo(x,0);c.lineTo(x,480);c.stroke()}
 for(let y=0;y<480;y+=40){c.beginPath();c.moveTo(0,y);c.lineTo(640,y);c.stroke()}
 c.fillStyle='#4c6489';for(let i=0;i<70;i++){const x=(i*131+31)%640,y=(i*79+17)%480;c.fillRect(x,y,1,1)}
 c.strokeStyle='#625849';c.beginPath();c.ellipse(345,420,88,32,-.2,0,2*Math.PI);c.stroke();
 c.fillStyle='#bdc9d9';c.font='16px sans-serif';c.fillText('戰術示意 · 非正式地圖',16,26);
 const ts=targets(proposal), back=enabled&&$('scope').value==='both'?multiplier:1;
 blue.forEach((p,i)=>ship(c,p[0]-Math.min(45,t*5*back),p[1], '#74b4f5',0,i+1));
 red.forEach((p,i)=>{
  let way=ts[i],route=[p,way];
  if(enabled&&hero==='yang'&&i>0)route=[p,[340,i<3?78:404],way];
  if(enabled&&hero==='reinhard'&&i>1)route=[p,[350,i===2?116:364],way];
  c.save();c.setLineDash([5,7]);c.strokeStyle='#b5694f';c.beginPath();route.forEach((v,k)=>k?c.lineTo(...v):c.moveTo(...v));c.stroke();c.restore();
  let left=t*22*multiplier,pos=p,ang=Math.PI;
  for(let j=1;j<route.length;j++){let a=route[j-1],b=route[j],d=Math.hypot(b[0]-a[0],b[1]-a[1]);ang=Math.atan2(b[1]-a[1],b[0]-a[0]);if(left<=d){let u=d?left/d:1;pos=[a[0]+(b[0]-a[0])*u,a[1]+(b[1]-a[1])*u];break}left-=d;pos=b;}
  ship(c,...pos,'#ee8564',ang,i+1);
 });
 if(enabled){c.fillStyle='#e8c791';c.font='16px sans-serif';const label=hero==='yang'?'牽制主力 · 左右翼迂迴':'集中突擊 · 側後封鎖';c.fillText(label,18,456)}
}
function render(now){let t=((frozen??now)-start)/1000%23;draw('baseline',false,t);draw('proposal',true,t);
 $('baseText').textContent='通用八方位包圍；移動倍率 1.0。';
 $('proposalText').textContent=$('hero').value==='none'?'名將不在場：保持現行策略與移動額度。':`名將在場：${$('scope').value==='hero'?'英雄陣營':'戰場雙方'} ${$('speed').value} 倍。分翼路線是待定案提案。`;
 requestAnimationFrame(render)}
controls.forEach(id=>$(id).addEventListener('change',()=>{start=performance.now();if(frozen!==null)frozen=start}));
$('restart').onclick=()=>{start=performance.now();if(frozen!==null)frozen=start};
$('pause').onclick=()=>{if(frozen===null){frozen=performance.now();pauseAt=frozen;$('pause').textContent='繼續'}else{start+=performance.now()-pauseAt;frozen=null;$('pause').textContent='暫停'}};
requestAnimationFrame(render);
</script></main></html>'''
path = OUT/'HERO-TACTICS-PREVIEW.html'
path.write_text(html)
sources = {}
for name in ('docs/internals/data/keireki.json','internal/ai/tactical.go',
             'internal/game/battle.go','internal/rules/battlemovement.go'):
    sources[name] = hashlib.sha256((ROOT/name).read_bytes()).hexdigest()
(OUT/'PREVIEW.json').write_text(json.dumps(dict(classification='可撤回規則示意，未進正式Game',
    heroes={'0':'萊因哈特','52':'楊威利'},sources=sources,
    html_sha256=hashlib.sha256(path.read_bytes()).hexdigest()),ensure_ascii=False,indent=2)+'\n')
print(path)

if args.check:
    check = r'''<script>
    const checks=[];function expect(ok,name){if(!ok)throw Error(name);checks.push(name)};
    try{
      expect(document.querySelectorAll('canvas').length===2,'兩欄比較');
      expect($('scope').value==='hero','預設英雄陣營');
      for(const h of ['yang','reinhard','none']){
        $('hero').value=h;$('hero').dispatchEvent(new Event('change'));
        for(const s of ['hero','both'])for(const m of ['1.25','1.5','2']){
          $('scope').value=s;$('speed').value=m;draw('proposal',true,10);
          expect(targets(true).length===5,'五艦隊目標 '+h+s+m);
        }
      }
      $('pause').click();expect($('pause').textContent==='繼續','暫停');
      $('pause').click();expect($('pause').textContent==='暫停','繼續');
      $('restart').click();
      $('hero').value='yang';$('scope').value='hero';$('speed').value='1.5';
      draw('baseline',false,10);draw('proposal',true,10);
      expect(document.documentElement.scrollWidth<=innerWidth,'整頁沒有水平溢出');
      document.body.insertAdjacentHTML('beforeend','<pre id="check-result">'+JSON.stringify({pass:true,width:innerWidth,checks:checks.length})+'</pre>');
    }catch(e){document.body.insertAdjacentHTML('beforeend','<pre id="check-result">'+JSON.stringify({pass:false,error:String(e)})+'</pre>')}
    </script>'''
    test = OUT/'BROWSER-CHECK.html'
    test.write_text(html.replace('</html>',check+'</html>'))
    results=[]
    for width,height in ((1440,1100),(390,1400)):
        command=['chromium','--headless','--no-sandbox','--disable-dev-shm-usage',
                 '--hide-scrollbars','--user-data-dir=/tmp/hero130-chrome-'+str(width),
                 '--window-size='+str(width)+','+str(height),'--virtual-time-budget=1200',
                 '--screenshot='+str(OUT/('preview-'+str(width)+'.png')),'--dump-dom',test.as_uri()]
        run=subprocess.run(command,text=True,capture_output=True,timeout=50,check=True)
        (OUT/('browser-'+str(width)+'.log')).write_text(run.stderr)
        found=re.search(r'<pre id="check-result">(\{[^<]*\})</pre>',run.stdout)
        assert found,run.stdout[-1000:]
        result=json.loads(found.group(1));assert result['pass'],result
        results.append(result);print('BROWSER',result,flush=True)
    (OUT/'BROWSER.json').write_text(json.dumps(dict(
        browser=subprocess.check_output(['chromium','--version'],text=True).strip(),results=results),
        ensure_ascii=False,indent=2)+'\n')
