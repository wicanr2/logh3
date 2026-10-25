"""已核准網路語錄的五語輪替預覽；不修改凍結HTML或正式素材。Docker內執行。"""
import json
from pathlib import Path

root = Path('/src')
data = json.loads((root/'tools/scripts/poster_quote_candidates.json').read_text())
original = (root/'docs/spec/157-poster-galaxy-frame.html').read_text()
code = r"""
<script>
(()=>{
const quoteDraft=__DATA__;
const pairSelector=document.createElement('select');pairSelector.id='quote-pair';
quoteDraft.pairs.forEach((p,i)=>{const o=document.createElement('option');o.value=i;o.textContent='第'+(i+1)+'組：'+p.reinhard['zh-TW']+' ／ '+p.yang['zh-TW'];pairSelector.append(o)});
const panel=document.createElement('section');panel.style.cssText='padding:18px;border:1px solid #d6bc86;margin:20px 0';
const label=document.createElement('p');label.textContent='名言輪替預覽 · 五組、每組60秒、五語。已同意附來源網路語錄，尚未逐字核對尖端舊版。';panel.append(label,pairSelector);
const remaining=document.createElement('p');panel.append(remaining);
const button=document.createElement('button');button.textContent='暫停名言輪替';panel.append(button);
const list=document.createElement('details'),summary=document.createElement('summary');summary.textContent='五組原文、來源與核對限制';list.append(summary);
quoteDraft.pairs.forEach((p,i)=>{const para=document.createElement('p');para.textContent=(i+1)+'　萊因哈特：'+p.reinhard['zh-TW']+'　楊威利：'+p.yang['zh-TW']+'　';for(const url of p.sources){const link=document.createElement('a');link.href=url.startsWith('http')?url:'../../'+url;link.textContent='［來源］';para.append(link)}const note=document.createElement('small');note.textContent=' '+p.verification;para.append(note);list.append(para)});panel.append(list);
document.querySelector('.wrap').prepend(panel);
let pair=0,began=performance.now(),running=true;
function fitQuotes(){for(const side of ['empire','alliance']){const quote=document.getElementById('quote-'+side),name=document.getElementById('name-'+side);if(!quote||!name)continue;const spans=[...quote.querySelectorAll('tspan')];if(!spans.length)continue;const shift=Math.max(0,Number(spans.at(-1).getAttribute('y'))-(Number(name.getAttribute('y'))-21));if(shift)spans.forEach(s=>s.setAttribute('y',Number(s.getAttribute('y'))-shift))}}
new MutationObserver(fitQuotes).observe(document.getElementById('portrait-labels'),{childList:true,subtree:true});
function applyPair(){const p=quoteDraft.pairs[pair];DATA.quotes['poster.quote.empire']=p.reinhard;DATA.quotes['poster.quote.alliance']=p.yang;document.getElementById('posterSet').value='v1';proposalControl.render();fitQuotes();pairSelector.value=String(pair)}
pairSelector.onchange=()=>{pair=Number(pairSelector.value);began=performance.now();applyPair()};
button.onclick=()=>{running=!running;began=performance.now();button.textContent=running?'暫停名言輪替':'繼續名言輪替'};
setInterval(()=>{if(!window.proposalReady)return;if(running&&performance.now()-began>=60000){pair=(pair+1)%5;began=performance.now();applyPair()}remaining.textContent=running?'第'+(pair+1)+'／5組，'+Math.max(0,60-Math.floor((performance.now()-began)/1000))+'秒後換下一組':'輪替已暫停'},200);
window.quotePreview={setPair:i=>{pair=i;began=performance.now();applyPair()},draft:quoteDraft};
})();
</script>
""".replace('__DATA__', json.dumps(data, ensure_ascii=False).replace('</', '<\/'))
out = root/'workplace/poster165/quotes-preview.html'
assert out.parent.stat().st_uid == (root/'tools/scripts').stat().st_uid
out.write_text(original.replace('</body>',code+'</body>'))
print(out, out.stat().st_size)
