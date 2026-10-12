"""只產生供畫布方向討論的HTML試樣，不改正式UI或素材。"""
import base64
from pathlib import Path

root = Path('/src/dist-all/v.1.0.0-20261010')
screen = root / 'smoke/gui-0/02-budget-x11.png'
image = 'data:image/png;base64,' + base64.b64encode(screen.read_bytes()).decode()
html = '''<!doctype html><html lang="zh-Hant"><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1"><title>HD畫布方向試樣</title>
<style>
*{box-sizing:border-box}body{margin:0;background:#071820;color:#f3f3f3;font:18px/1.6 system-ui,sans-serif;padding:24px}
main{max-width:1500px;margin:auto}h1{font-size:28px}button{font:inherit;padding:10px 20px;margin:0 12px 16px 0;background:#163b47;color:#fff;border:1px solid #61d3f3;cursor:pointer}
button[aria-pressed=true]{background:#61d3f3;color:#071820}.note{color:#b2c2cf}.view{border:1px solid #365566;padding:20px;background:#102b35}.original img{display:block;max-width:100%;max-height:75vh;margin:auto}
.wide{display:none;aspect-ratio:16/9;padding:28px;grid-template-columns:30% 1fr;gap:26px}.wide header{grid-column:1/-1;border-bottom:1px solid #61d3f3;font-size:26px}.map{overflow:hidden;position:relative;border:1px solid #365566}.map img{height:100%;position:absolute;left:0;max-width:none}.budget{background:#045361;border:2px solid #aeb3c9;padding:28px}.budget h2{margin:0 0 8px}.budget table{width:100%;border-collapse:collapse;font-size:clamp(14px,1.5vw,24px)}td{padding:8px;border-bottom:1px solid #62818b}td:last-child{text-align:right}.actions{text-align:right;margin-top:18px}.chip{display:inline-block;padding:5px 15px;border:1px solid #61d3f3;margin-left:10px}footer{grid-column:1/-1;color:#b2c2cf}
@media(max-width:850px){.wide{aspect-ratio:auto;grid-template-columns:1fr}.map{display:none}.budget{padding:15px}}
</style><main><h1>下一階段HD：畫布方向試樣</h1>
<p class="note">A為v1.0.0實際財政畫面。B為可丟棄的寬螢幕版面草圖，尚未接入遊戲。此頁只切換樣式，不能變更預算。</p>
<button id="a" aria-pressed="true" onclick="choose(false)">A　保留4:3</button><button id="b" aria-pressed="false" onclick="choose(true)">B　16:9重排草圖</button>
<section id="original" class="view original"><img src="IMAGE" alt="目前4:3財政預算實際畫面"></section>
<section id="wide" class="view wide"><header>銀河帝國　／　次月份預算計劃</header>
<aside class="map"><img src="IMAGE" alt="原版星圖局部作版面占位"></aside>
<article class="budget"><h2>軍事預算</h2><div class="note">資料取自同一張財政畫面，中文字僅作排版比較。</div>
<table><tr><td>本月份原定軍事費</td><td>0</td></tr><tr><td>艦隊維持費</td><td>0</td></tr><tr><td>移動經費</td><td>0</td></tr><tr><td>下月份預估徵稅額</td><td>10,741　普通　↑ ↓</td></tr><tr><td>預估軍事費</td><td>60,741</td></tr><tr><td>船艦建造費</td><td>15,185　25%　↑ ↓</td></tr><tr><td>剩餘軍事費</td><td>45,556</td></tr></table>
<div class="actions"><span class="chip">參謀</span><span class="chip">結束</span></div></article>
<footer>試樣：寬螢幕重排需要重新驗證命中座標、視窗、文字及地圖範圍。</footer></section>
<p id="explain">A沿現有構圖與操作位置。新字型、艦艇與地圖可逐批替換。</p>
<p class="note">這份草圖不代表使用者已核准版面，也不算HD完成證據。含原版衍生圖像，只留本機。</p></main>
<script>function choose(w){document.getElementById('original').style.display=w?'none':'block';document.getElementById('wide').style.display=w?'grid':'none';document.getElementById('a').setAttribute('aria-pressed',!w);document.getElementById('b').setAttribute('aria-pressed',w);document.getElementById('explain').textContent=w?'B把星圖與財政表分欄，字級可提高；需要重做介面幾何與操作驗證。':'A沿現有構圖與操作位置。新字型、艦艇與地圖可逐批替換。';}</script></html>'''
target = root / 'HD-PREVIEW.html'
assert not target.exists()
target.write_text(html.replace('IMAGE', image))
print('已產生可丟棄HD畫布試樣', target)
