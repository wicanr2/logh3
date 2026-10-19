"""提督編輯介面草稿。沿用現行陣營色與素材，只產生離線試樣，不寫遊戲或存檔。

LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=tools/with-xvfb.sh \
tools/gui.sh xrun python3 tools/scripts/admiral_editor_preview.py
"""
import base64
import hashlib
import html
import json
import os
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / 'workplace/admiral-editor-preview'
assert os.getuid() == 1000 and OUT.parent.stat().st_uid == 1000
OUT.mkdir(exist_ok=True)
assert OUT.stat().st_uid == 1000


def uri(path):
    return 'data:image/png;base64,' + base64.b64encode(path.read_bytes()).decode()


names = json.loads((ROOT/'docs/internals/data/keireki.json').read_text())['records']
records = json.loads((ROOT/'docs/internals/data/admiral.json').read_text())['variants']['DTX']
data = []
for i, (name, record) in enumerate(zip(names, records)):
    raw = bytes.fromhex(record['raw'])
    data.append(dict(id=i, name=name['name'], faction=raw[1], level=raw[6],
                     basic=[raw[n] for n in [11, 12, 13, 14, 15, 16, 17, 18, 19, 23]],
                     hidden=[raw[n] for n in [20, 21, 22]],
                     special=[100 if i == 5 else 25, 50 if i in (0, 52) else 0,
                              30 if i in (0, 52) else 0]))

themes = json.loads((ROOT/'assets/hd/interface-themes.json').read_text())['themes']
images = {key: uri(ROOT/path) for key, path in {
    'empire': 'docs/images/readme-settings-frame-empire.png',
    'alliance': 'docs/images/readme-settings-frame-alliance.png',
    'face0': 'assets/hd/face/face-00.png',
    'face52': 'assets/hd/face/face-52.png',
}.items()}

source = r'''<!doctype html><html lang="zh-Hant"><meta charset="utf-8">
<title>提督能力編輯・操作草稿</title>
<style>
*{box-sizing:border-box}body{margin:0;background:#080d16;color:#fff0c9;font:14px sans-serif}
main{width:min(960px,100%);margin:0 auto}header{height:60px;display:flex;align-items:center;gap:16px;padding:0 16px;color:#cbd1db}
header span{font-size:12px}header button{font:inherit;background:#173f70;border:1px solid #82a4cb;color:#fff0c9;padding:5px 14px;cursor:pointer}
#viewport{width:960px;height:720px;position:relative;transform-origin:top left}
#screen{width:640px;height:480px;position:relative;transform:scale(1.5);transform-origin:top left;background-size:640px 480px;color:var(--white);font:12px/1 'Noto Sans Mono CJK TC','Noto Sans CJK TC',sans-serif}
#screen button,#screen input{font:inherit;color:inherit;border-radius:0;outline-offset:1px}
#screen button{cursor:pointer;background:var(--button);border:0;border-top:1px solid var(--metal);border-right:1px solid var(--metal);border-bottom:1px solid #080e18;border-left:1px solid #080e18;box-shadow:inset 0 1px #ffffff18;padding:0 4px}
#screen button:hover{filter:brightness(1.15)}#screen button:focus-visible,#screen input:focus-visible{outline:1px solid var(--white)}
#screen button.active{background:var(--selected);color:var(--selected-ink);box-shadow:inset 1px 1px #0006}
.panel{position:absolute;background:var(--surface);border:1px solid var(--metal);box-shadow:inset 1px -1px #060b16,inset -1px 1px #d2e2ff70,3px 3px 0 #000a}
#editor{left:24px;top:40px;width:592px;height:400px}
.title{height:27px;text-align:center;line-height:26px;font-size:14px;border-bottom:1px solid var(--metal);margin:0 7px}
.sect{position:absolute}.inset{background:#0002;border:1px solid #080e18;border-bottom-color:var(--metal);border-right-color:var(--metal)}
.factions{left:12px;top:36px;display:flex;width:156px;height:23px;gap:4px}.factions button{width:76px}
#list{left:12px;top:67px;width:156px;height:249px;overflow-y:auto;padding:3px;scrollbar-color:var(--metal) var(--surface);scrollbar-width:thin}
#list button{display:block;width:100%;height:26px;margin:0 0 3px;text-align:left;padding-left:8px;white-space:nowrap}
.list-label{left:12px;top:326px;line-height:17px;color:var(--white);opacity:.85}
#portrait{left:180px;top:37px;width:64px;height:80px;object-fit:cover;object-position:top;border:1px solid var(--metal);background:#000}
#identity{left:258px;top:40px;line-height:22px}#name{font-size:16px}#state{font-size:11px;opacity:.8}
.tabs{left:180px;top:128px;height:25px;display:flex;gap:4px}.tabs button{width:124px}
#fields{left:180px;top:165px;width:388px}.basic-grid{display:grid;grid-template-columns:1fr 1fr;gap:10px 16px}
.ability{display:flex;align-items:center;justify-content:space-between;height:23px;white-space:nowrap}
.step{display:flex;gap:3px;align-items:center}.step button{width:21px;height:23px;font-size:15px!important}.step input{width:39px;height:23px;background:#0003;border:1px solid #080e18;border-right-color:var(--metal);border-bottom-color:var(--metal);text-align:center;padding:0;appearance:textfield}
input::-webkit-inner-spin-button{appearance:none}.percent input{width:58px}.special-row{height:49px}.special-row .ability{height:25px}.desc{font-size:10px;opacity:.85;line-height:16px;margin:2px 0 0}
#help{left:180px;top:331px;width:388px;height:25px;line-height:15px;font-size:10px;opacity:.9}
.footer{left:12px;right:12px;bottom:12px;display:flex;gap:8px;align-items:center}.footer button{height:25px;width:93px}.footer .spacer{flex:1}#feedback{left:12px;top:347px;font-size:10px;opacity:.85}
#settings{left:0;top:0;width:640px;height:480px;display:none;border:0;background:transparent;box-shadow:none}#reopen{position:absolute;left:176px;top:378px;width:112px;height:16px}
#numeric{position:absolute;display:none;width:235px;height:138px;left:203px;top:168px;z-index:5;padding:14px;text-align:center}#numeric input{display:block;width:95px;height:28px;margin:13px auto;background:var(--surface);color:var(--white);border:1px solid var(--metal);text-align:center}#numeric button{width:80px;height:24px;margin:0 4px}
#confirm{position:absolute;display:none;width:280px;left:180px;top:186px;z-index:6;padding:17px;line-height:20px;text-align:center}#confirm button{height:25px;margin:13px 3px 0;width:105px}
footer{padding:12px 16px;color:#aebaca;font-size:12px;line-height:1.8}
</style><main><header><b>提督編輯・操作草稿</b><button onclick="setFaction(0)">帝國樣式</button><button onclick="setFaction(1)">同盟樣式</button><span>可點選、輸入、切換分頁</span></header>
<div id="viewport"><div id="screen">
<div id="settings" class="panel"><button id="reopen">編輯提督</button></div>
<div id="editor" class="panel"><div class="title">提督能力編輯</div>
<div class="sect factions"><button id="f0" onclick="setFaction(0)">銀河帝國</button><button id="f1" onclick="setFaction(1)">自由同盟</button></div>
<div id="list" class="sect inset"></div><div class="sect factions" style="top:322px;height:20px"><button aria-label="向上捲動提督清單" onclick="$('list').scrollBy(0,-87)">▲</button><button aria-label="向下捲動提督清單" onclick="$('list').scrollBy(0,87)">▼</button></div>
<img id="portrait" class="sect" alt="提督頭像"><div id="identity" class="sect"><div id="name"></div><div id="affiliation"></div><div id="state">修改暫存於本次草稿</div></div>
<div class="sect tabs"><button id="tab-basic" onclick="setTab('basic')">基本能力</button><button id="tab-hidden" onclick="setTab('hidden')">內部能力</button><button id="tab-special" onclick="setTab('special')">特殊能力</button></div>
<div id="fields" class="sect"></div><div id="help" class="sect"></div><div id="feedback" class="sect">離線操作草稿</div>
<div class="sect footer"><button id="reset" onclick="resetPerson()">還原本提督</button><span class="spacer"></span><button id="cancel" onclick="closeEditor()">取消</button><button id="apply" onclick="applyDraft()">套用變更</button></div></div>
<div id="numeric" class="panel"><span id="numeric-title"></span><input id="numeric-value" type="number"><button id="numeric-cancel" onclick="numeric.style.display='none'">取消</button><button id="numeric-ok">確定</button></div>
<div id="confirm" class="panel"><div id="confirm-message"></div><button id="confirm-no">返回編輯</button><button id="confirm-yes">確定還原</button></div>
</div></div><footer>草稿只在此頁暫存，不會修改遊戲。百分比按鈕每次調整 5%，點數值可直接輸入。<br>楊威利與萊因哈特的全軍攻擊預設增加30%；內部能力按已知用途說明，原版欄名未知。正式功能採全局保存。</footer></main>
<script>
const DATA=__DATA__,THEMES=__THEMES__,IMAGES=__IMAGES__;
const $=id=>document.getElementById(id),numeric=$('numeric'),screen=$('screen');
const basicNames=['統率力','移動','回轉','營運','光線砲','飛彈','戰鬥機','防禦','占領','運用'];
const hiddenNames=['內政／應變參數','稅務／決策權重','目標選擇判斷'];
const hiddenHelp=['影響稅率、治安分段、AI艦隊類別與低完整度時的判斷。','影響內政參謀評分、稅率與戰術AI權重。','數值愈高，愈容易選中評分最高的攻擊目標。'];
const specialNames=['旗艦擊毀換艦率','全軍移動加成','全軍攻擊加成'];
let faction=0,selected=0,tab='special',draft=structuredClone(DATA),saved=structuredClone(DATA),dirty=new Set(),editKey=null;
function status(s){$('feedback').textContent=s}
function theme(f){const t=THEMES[f?'alliance':'empire'];for(const k of ['surface','button','metal','white','selected'])screen.style.setProperty('--'+k,t[k]);screen.style.setProperty('--selected-ink',t.selected_ink||'#fff1d3');screen.style.backgroundImage=`url(${IMAGES[f?'alliance':'empire']})`}
function setFaction(f){faction=f;selected=f?52:0;theme(f);draw()}
function setTab(t){tab=t;draw()}
function changed(){let a=draft[selected],b=saved[selected];JSON.stringify(a)===JSON.stringify(b)?dirty.delete(selected):dirty.add(selected);draw()}
function clampValue(kind,value){return Math.max(kind==='hidden'?1:0,Math.min(kind==='special'?200:200,Math.round(value)))}
function update(kind,i,value){if(Number.isFinite(value)){draft[selected][kind][i]=kind==='special'&&i===0?Math.min(100,clampValue(kind,value)):clampValue(kind,value);changed()}}
function step(kind,i,delta){update(kind,i,(draft[selected][kind][i]??0)+delta)}
function edit(kind,i){editKey=[kind,i];$('numeric-title').textContent=(kind==='basic'?basicNames:kind==='hidden'?hiddenNames:specialNames)[i];$('numeric-value').value=draft[selected][kind][i]??'';$('numeric-value').min=kind==='hidden'?1:0;$('numeric-value').max=kind==='special'&&i===0?100:200;numeric.style.display='block';$('numeric-value').focus()}
$('numeric-ok').onclick=()=>{const input=$('numeric-value');if(input.value!==''&&input.checkValidity()){update(...editKey,Number(input.value));numeric.style.display='none'}};
function field(kind,i,label){let v=draft[selected][kind][i],value=v===null?'待定':kind==='special'?(i===0?'':'+')+v+'%':v;return `<div class="ability"><span>${label}</span><div class="step ${kind==='special'?'percent':''}"><button aria-label="減少${label}" onclick="step('${kind}',${i},${kind==='special'?-5:-1})">−</button><button style="width:${kind==='special'?67:43}px;font-size:12px!important" aria-label="輸入${label}" onclick="edit('${kind}',${i})">${value}</button><button aria-label="增加${label}" onclick="step('${kind}',${i},${kind==='special'?5:1})">＋</button></div></div>`}
function draw(){
 $('f0').classList.toggle('active',faction===0);$('f1').classList.toggle('active',faction===1);
 for(const t of ['basic','hidden','special'])$('tab-'+t).classList.toggle('active',tab===t);
 const scroll=$('list').scrollTop;
 $('list').innerHTML=draft.filter(x=>x.faction===faction).map(x=>`<button class="${x.id===selected?'active':''}" onclick="selectPerson(${x.id})">${dirty.has(x.id)?'＊':'　'}${x.name}</button>`).join('');$('list').scrollTop=scroll;
 const item=$('list').querySelector('button.active');if(item.offsetTop<$('list').scrollTop)$('list').scrollTop=item.offsetTop;else if(item.offsetTop+item.offsetHeight>$('list').scrollTop+$('list').clientHeight)$('list').scrollTop=item.offsetTop+item.offsetHeight-$('list').clientHeight;
 const a=draft[selected];$('name').textContent=a.name;$('affiliation').textContent=(a.faction?'自由行星同盟':'銀河帝國')+'　等級 '+a.level;
 $('portrait').src=IMAGES['face'+selected]||IMAGES['fallback'+selected];$('portrait').style.visibility=$('portrait').src?'visible':'hidden';
 $('state').textContent=dirty.has(selected)?'＊ 此提督有尚未套用的修改':'選擇能力欄位即可調整';
 if(tab==='basic'){$('fields').innerHTML='<div class="basic-grid">'+basicNames.map((n,i)=>field('basic',i,n)).join('')+'</div>';$('help').textContent='十項原版能力。點選數值輸入，或使用 −／＋ 微調。'}
 if(tab==='hidden'){$('fields').innerHTML=hiddenNames.map((n,i)=>'<div class="special-row">'+field('hidden',i,n)+'<p class="desc">'+hiddenHelp[i]+'</p></div>').join('');$('help').textContent='以上為用途說明，原版欄名未知。前兩項並非越高越強。'}
 if(tab==='special'){$('fields').innerHTML=specialNames.map((n,i)=>'<div class="special-row">'+field('special',i,n)+'<p class="desc">'+['旗艦被擊毀時的換艦成功機率。','參戰時，同陣營全部艦隊的移動距離增幅。','參戰時，同陣營全部艦隊的攻擊增幅。'][i]+'</p></div>').join('');$('help').textContent='＋50% 表示 1.5 倍。全軍加成屬強化規則；名將攻擊預設增加30%。'}
 $('apply').classList.toggle('active',dirty.size>0);status(dirty.size?'共 '+dirty.size+' 位提督有修改，尚未套用':'離線操作草稿');
}
function selectPerson(i){selected=i;draw()}
function confirmAction(message,yes,label){$('confirm-message').textContent=message;$('confirm-yes').textContent=label;$('confirm').style.display='block';$('confirm-no').onclick=()=>$('confirm').style.display='none';$('confirm-yes').onclick=()=>{$('confirm').style.display='none';yes()}}
function resetPerson(){confirmAction('將 '+draft[selected].name+' 還原為進入草稿時的數值？',()=>{draft[selected]=structuredClone(DATA[selected]);changed()},'確定還原')}
function hideEditor(){$('editor').style.display='none';$('settings').style.display='block'}
function closeEditor(){if(dirty.size)confirmAction('放棄本次尚未套用的修改？',()=>{draft=structuredClone(saved);dirty.clear();hideEditor()},'放棄修改');else hideEditor()}
function applyDraft(){saved=structuredClone(draft);dirty.clear();draw();status('已套用至草稿，未寫入遊戲')}
$('reopen').onclick=()=>{$('settings').style.display='none';$('editor').style.display='block';draw()};
screen.addEventListener('contextmenu',e=>{e.preventDefault();closeEditor()});
document.addEventListener('keydown',e=>{if(e.key==='Escape'){if(numeric.style.display==='block')numeric.style.display='none';else if($('confirm').style.display==='block')$('confirm').style.display='none';else closeEditor()}if(e.key==='Enter'&&numeric.style.display==='block')$('numeric-ok').click()});
function fit(){const scale=Math.min(1,innerWidth/960);$('viewport').style.transform=`scale(${scale})`;$('viewport').style.marginBottom=(720*scale-720)+'px'}addEventListener('resize',fit);fit();
const params=new URLSearchParams(location.search);setFaction(Number(params.get('faction')||0));setTab(params.get('tab')||'special');
if(params.get('entry')==='settings')hideEditor();
window.previewReady=true;
</script></html>'''

# 除兩位示範名將外，其餘清單仍使用該人物的既有頭像，不借用他人肖像。
for i in range(82):
    from PIL import Image
    import io
    with Image.open(ROOT/f'assets/hd/face/face-{i:02}.png') as face:
        face.thumbnail((128, 160))
        buf = io.BytesIO()
        face.save(buf, format='PNG')
        images['fallback'+str(i)] = 'data:image/png;base64,' + base64.b64encode(buf.getvalue()).decode()

for key, value in [('DATA', data), ('THEMES', themes), ('IMAGES', images)]:
    source = source.replace('__'+key+'__', json.dumps(value, ensure_ascii=False))
(OUT/'index.html').write_text(source)

checks = r'''<script>
addEventListener('load',async()=>{await document.fonts.ready;const checks=[];function ok(v,n){if(!v)throw Error(n);checks.push(n)}
try{
ok(DATA.length===82,'82位提督');setFaction(0);setTab('basic');ok(document.querySelectorAll('#fields .ability').length===10,'十項基本能力');
let old=draft[0].basic[0];document.querySelector('[aria-label="增加統率力"]').click();ok(draft[0].basic[0]===old+1,'點擊微調');
document.querySelector('[aria-label="輸入統率力"]').click();$('numeric-value').value='150';$('numeric-ok').click();ok(draft[0].basic[0]===150,'直接輸入');
setTab('hidden');ok(document.querySelectorAll('#fields .ability').length===3,'三項內部能力');old=draft[0].hidden[0];step('hidden',0,1);ok(draft[0].hidden[0]===old+1,'內部能力可調');
setTab('special');ok(draft[0].special[2]===30,'攻擊預設30%');step('special',1,5);ok(draft[0].special[1]===55,'移動百分比微調');
selectPerson(5);ok(draft[5].special[0]===100,'謬拉現行100%');step('special',0,5);ok(draft[5].special[0]===100,'機率上限');
selectPerson(0);closeEditor();ok($('confirm').style.display==='block','取消未套用修改提示');$('confirm-no').click();
resetPerson();$('confirm-yes').click();ok(draft[0].basic[0]===DATA[0].basic[0]&&draft[0].special[1]===50,'還原本提督');
setFaction(1);ok(selected===52&&$('name').textContent==='楊威利','同盟主題與人物');step('special',2,20);applyDraft();ok(saved[52].special[2]===50&&dirty.size===0,'套用僅在草稿');
let active=$('list').querySelector('button.active').getBoundingClientRect(),list=$('list').getBoundingClientRect();ok(active.top>=list.top&&active.bottom<=list.bottom+1,'選中提督保持可見');
closeEditor();ok($('settings').style.display==='block','返回完整設定頁');$('reopen').click();ok($('editor').style.display==='block','設定頁編輯提督入口');
setFaction(Number(params.get('faction')||0));setTab(params.get('tab')||'special');draft=structuredClone(DATA);saved=structuredClone(DATA);dirty.clear();draw();
if(params.get('entry')==='settings')hideEditor();else $('editor').style.display='block';
for(const e of document.querySelectorAll('#editor button,#fields .desc,#help')){if(e.closest('#list'))continue;const r=e.getBoundingClientRect(),p=$('editor').getBoundingClientRect();ok(r.left>=p.left&&r.right<=p.right+1&&r.top>=p.top&&r.bottom<=p.bottom+1,'控制與說明在面框內')}
const result=document.createElement('pre');result.id='browser-result';result.hidden=true;result.textContent=JSON.stringify({pass:true,checks});document.body.appendChild(result)
}catch(e){let p=document.createElement('pre');p.id='browser-result';p.textContent=JSON.stringify({pass:false,error:String(e)});document.body.appendChild(p)}});
</script>'''
(OUT/'check.html').write_text(source.replace('</html>', checks+'</html>'))
results=[]
for faction, tab, entry in [(0, 'basic', ''), (0, 'hidden', ''), (0, 'special', ''), (1, 'special', ''), (0, 'special', 'settings')]:
    name = ('alliance' if faction else 'empire')+'-'+(entry or tab)
    run = subprocess.run(['chromium', '--headless', '--no-sandbox', '--disable-dev-shm-usage',
        '--disable-gpu', '--no-first-run', '--hide-scrollbars', '--user-data-dir=/tmp/logh3-admiral-preview',
        '--window-size=960,900', '--force-device-scale-factor=1', '--virtual-time-budget=6000',
        '--screenshot='+str(OUT/(name+'.png')), '--dump-dom',
        (OUT/'check.html').as_uri()+f'?faction={faction}&tab={tab}&entry={entry}'], capture_output=True, text=True, timeout=40, check=True)
    match = re.search(r'<pre id="browser-result"[^>]*>(.*?)</pre>',run.stdout)
    assert match, run.stderr[-1000:]
    result=json.loads(html.unescape(match[1]))
    assert result['pass'], result
    results.append(dict(scene=name, **result))
    print(name, 'PASS', len(result['checks']), flush=True)

inputs=['docs/internals/data/admiral.json','docs/internals/data/keireki.json',
        'docs/internals/08-admiral-dt.md','docs/re/287-admiral-remaining-fields.md',
        'assets/hd/interface-themes.json','docs/images/readme-settings-frame-empire.png',
        'docs/images/readme-settings-frame-alliance.png', 'internal/game/herocommand.go',
        'internal/rules/flagshiptransfer.go']
(OUT/'manifest.json').write_text(json.dumps(dict(kind='disposable-ui-prototype',production_changed=False,
    image='logh3-hd-preview:20261011-r1',browser=subprocess.check_output(['chromium','--version'],text=True).strip(),
    inputs={n:hashlib.sha256((ROOT/n).read_bytes()).hexdigest() for n in inputs if (ROOT/n).exists()},
    generator_sha256=hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
    results=results),ensure_ascii=False,indent=2)+'\n')
(OUT/'check.html').unlink()
