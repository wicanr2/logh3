"""原配置A試樣：只寫workplace，不接正式Game、存檔或交付封包。"""
import base64
import hashlib
import html
import io
import json
import os
from pathlib import Path
import struct

from PIL import Image

SRC = Path('/src')
OUT = Path('/work/hd124-original-layout')
DATA = SRC/'docs/internals/data'
assert os.getuid() == 1000 and OUT.parent.stat().st_uid == os.getuid()
OUT.mkdir(exist_ok=True)
assert OUT.stat().st_uid == os.getuid()
sources = {}


def read(path):
    path = SRC/path
    data = path.read_bytes()
    sources[str(path.relative_to(SRC))] = hashlib.sha256(data).hexdigest()
    return data


def doc(name):
    return json.loads(read(Path('docs/internals/data')/(name+'.json')))


def data_url(data, mime='image/png'):
    return 'data:'+mime+';base64,'+base64.b64encode(data).decode()


def png_url(img):
    buf = io.BytesIO()
    img.save(buf, format='PNG')
    return data_url(buf.getvalue())


def runs_path(points):
    """將同列連續像素轉成向量矩形，嚴守原遮罩，不猜補或平滑邊界。"""
    rows = {}
    for x, y in points:
        rows.setdefault(y, []).append(x)
    parts = []
    for y, xs in sorted(rows.items()):
        xs.sort()
        start = prev = xs[0]
        for x in xs[1:]+[xs[-1]+2]:
            if x != prev+1:
                width = prev-start+1
                parts.append(f'M{start},{y}h{width}v1h-{width}z')
                start = x
            prev = x
    return ''.join(parts)


def frame(x, y, w, h, fill='url(#panel)', extra=''):
    return (f'<g {extra}><rect x="{x}" y="{y}" width="{w}" height="{h}" fill="{fill}" '
            f'stroke="#8d7952" stroke-width=".65"/><rect x="{x+2}" y="{y+2}" width="{w-4}" '
            f'height="{h-4}" fill="none" stroke="#314961" stroke-width=".5"/></g>')


def text(x, y, value, size=14, fill='#e9e7db', anchor='start', attrs=''):
    return (f'<text x="{x}" y="{y+12.7}" font-size="{size}" fill="{fill}" '
            f'text-anchor="{anchor}" {attrs}>{html.escape(str(value))}</text>')


def button(x, y, w, h, label, action, ident):
    return (f'<g class="game-button" role="button" tabindex="0" aria-label="{label}" '
            f'data-action="{action}" id="{ident}" data-rect="{x},{y},{w},{h}">'
            + frame(x,y,w,h,fill='url(#button)')
            + text(x+w/2,y+(h-15)/2,label,13,anchor='middle')+'</g>')


defs = '''<defs>
<linearGradient id="panel" x2="0" y2="1"><stop stop-color="#152d47"/><stop offset="1" stop-color="#0c1d31"/></linearGradient>
<linearGradient id="button" x2="0" y2="1"><stop stop-color="#30465b"/><stop offset="1" stop-color="#152b40"/></linearGradient>
<radialGradient id="space"><stop stop-color="#0c2035"/><stop offset="1" stop-color="#030a15"/></radialGradient>
<linearGradient id="sector-1" x2="1" y2="1"><stop stop-color="#214b7b"/><stop offset="1" stop-color="#0a1d39"/></linearGradient>
<linearGradient id="sector-12" x2="1" y2="1"><stop stop-color="#805528"/><stop offset="1" stop-color="#271b18"/></linearGradient>
<linearGradient id="sector-3" x2="1" y2="1"><stop stop-color="#19253a"/><stop offset="1" stop-color="#080f1c"/></linearGradient>
<linearGradient id="sector-6" x2="1" y2="1"><stop stop-color="#81918c"/><stop offset="1" stop-color="#243747"/></linearGradient>
<clipPath id="star-clip"><rect x="4" y="44" width="312" height="292"/></clipPath>
<clipPath id="field-clip"><rect x="16" y="60" width="448" height="360"/></clipPath>
</defs>'''

spec = doc('screen-budget-plan')
command = doc('screen-command-phase')
battle = doc('screen-battle')
regions = doc('strmap-regions')['regions']
stellar = doc('stellar')['versions']['DT8']
budget_rows = doc('budget-panel')
read(Path('internal/game/planscreens.go'))
read(Path('internal/asset/rgn.go'))
read(Path('internal/screen/battle.go'))
read(Path('docs/goal/2026-10-11-original-layout-hd-round-124.md'))
read(Path('tools/scripts/hd_original_layout_check.py'))

reference_paths = {
    'strategy':'dist-all/v.1.0.0-20261010/smoke/gui-0/12-command-x11.png',
    'account':'dist-all/v.1.0.0-20261010/smoke/gui-0/01-account-x11.png',
    'budget':'dist-all/v.1.0.0-20261010/smoke/gui-0/02-budget-x11.png',
    'battle':'workplace/hd123-s1-20261011/record-f/battle.png',
}
references = {key: data_url(read(Path(p))) for key,p in reference_paths.items()}

# STRMAP.BIT原始planar解碼；用既有匯出圖逐個非零色號核對。
bit = read(Path('org_game/extracted/STRMAP.BIT'))
rgn = read(Path('org_game/extracted/STRMAP.RGN'))
assert len(bit) == 48000 and len(regions) == 32
base = Image.new('P', (320,300))
exported = Image.open(io.BytesIO(read(Path('assets/original/strmap/strmap.png'))))
assert exported.mode == 'P' and exported.size == base.size
base.putpalette(exported.getpalette())
indices = []
for y in range(300):
    for x in range(320):
        offset = ((y//20)*20+x//16)*160+(y%20)*8
        # 沿DecodeTiles的byteswap語意，螢幕左側位元在每個word的第一個byte。
        planes = struct.unpack_from('>4H',bit,offset)
        value = sum(((plane>>(15-x%16))&1)<<i for i,plane in enumerate(planes))
        indices.append(value)
base.putdata(indices)
nonzero = [i for i,n in enumerate(indices) if n]
assert all(indices[i] == exported.getdata()[i] for i in nonzero), '原星圖前景需與既有匯出逐色號相同'

region_paths, region_grid, counts = [], [-1]*(320*300), []
for region in regions:
    i = region['index']
    bx,by,bw,bh,offset = struct.unpack_from('<4BH',rgn,i*6)
    assert (bx,by,bw,bh,offset) == tuple(region[k] for k in ['bx','by','bw','bh','offset'])
    points = []
    for y in range(bh*20):
        for x in range(bw*16):
            j = 192+offset+((y//20)*bw+x//16)*40+(y%20)*2+(x%16)//8
            if rgn[j] & (128>>(x%8)):
                gx,gy=bx*16+x,by*20+y
                assert 0<=gx<320 and 0<=gy<300
                if region_grid[gy*320+gx] < 0:
                    region_grid[gy*320+gx] = i
                points.append((gx,gy))
    counts.append(len(points))
    region_paths.append(f'<path d="{runs_path(points)}" fill="url(#sector-{region["colour_796_01"]})" '
                        f'data-region="{i}" class="sector"/>')

# 航路、星名、星球及鷲徽暫留原前景。本試樣只驗區域面材，不宣稱全星圖重繪。
foreground=base.convert('RGBA')
foreground.putdata([(r,g,b,0 if indices[i] in [0,1,3,6,12] else 255)
                    for i,(r,g,b,a) in enumerate(foreground.getdata())])
star = (frame(1,41,317,299,fill='url(#space)')
        +'<g clip-path="url(#star-clip)"><g transform="translate(0 40)">'
        +''.join(region_paths)
        +f'<image width="320" height="300" href="{png_url(foreground)}" pointer-events="none"/>'
        +'<path id="selected-sector" fill="#f1d49a" opacity=".25" pointer-events="none"/>'
        +'</g></g>')


def strategy():
    s = star
    # 只用原board既有大矩形。文字是來源截圖中的固定試樣狀態。
    for b in command['windows']['board']['boxes'][:4]:
        s += frame(b['x'],b['y'],b['w'],b['h'])
    for x,y,v in [(336,60,'帝國軍'),(400,60,'宇宙曆　796 年 1 月 1 日'),
                  (336,80,'軍事費'),(496,80,'總稅收'),(464,80,'46495'),(624,80,'10341')]:
        s += text(x,y,v,13,anchor='end' if str(v).isdigit() else 'start')
    for x,label,act in [(350,'戰略命令','strategy-note'),(446,'一覽表','list-note'),(542,'系統指令','system-note')]:
        s += button(x,120,68,20,label,act,'nav-'+act)
    for x,y,v in [(336,162,'帝國軍星系 瓦爾哈拉'),(336,182,'治安　　82%　行星　4/4'),
                  (336,202,'防衛力　　　27300'),(336,222,'經濟力　　　82302'),
                  (512,222,'成長率　4.5%'),(336,242,'徵稅額　　　2633'),(544,242,'稅率　3.2%')]:
        s += text(x,y,v,13,fill='#b5cadd' if y==162 else '#e9e7db')
    for x,w,label in [(336,50,'艦隊'),(400,50,'行星'),(464,66,'敵艦隊'),(544,50,'提督')]:
        s += frame(x,280,w,20,fill='url(#button)')+text(x+w/2,282,label,13,anchor='middle')
    for y in [300,320,340,360,380]:
        s += frame(336,y,274,20,fill='#12283e')
    return s


def monthly_base(key='budget'):
    # 財政屬月初疊窗，底層與命令畫面不同；沿自己的正常截圖，不借用命令列。
    img=Image.open(io.BytesIO(read(Path(reference_paths[key])))).convert('RGBA')
    pixels=[]
    for r,g,b,a in img.getdata():
        pixels.append((r,g,b,0 if (r,g,b)==(0,81,97) else 255))
    img.putdata(pixels)
    return (frame(1,41,638,378)
            +f'<image width="640" height="480" href="{png_url(img)}"/>'+star.replace('id="selected-sector"',f'id="{key}-selection"'))


def budget():
    w=spec['windows']['report']
    s=frame(w['x'],w['y'],w['w'],w['h'],extra='data-window="budget"')
    for line in w['lines']:
        for cell in line['cells']:
            code=cell['code']
            glyph = chr(code) if code<128 else code.to_bytes(2,'big').decode('big5')
            s += text(cell['x']+cell['w']/2,line['y'],glyph,14,anchor='middle',
                      attrs='class="source-cell" data-width="'+str(cell['w'])+'"')
    # 原字串與字位保持；只用可用HD人物替換同一原矩形。
    face=w['face']
    face_path=Path('assets/hd/face')/f'face-{face["index"]:02d}.png'
    s+=frame(face['x'],face['y'],80,100,fill='#07111e')
    s+=f'<image x="{face["x"]+2}" y="{face["y"]+2}" width="76" height="96" href="{data_url(read(face_path))}"/>'
    s+=text(144,122,'風貝勒坦 上校')
    for y in [239,279]:
        s+=f'<path d="M144 {y}H400" fill="none" stroke="#6b654f" stroke-width=".5"/>'
    values={'budget':0,'ship_paid':0,'fleet_upkeep':0,'movement':0,'next_fleet_upkeep':0,
            'next_movement':0,'tax':10741,'military':60741,'ship_cost':15185,'rest':45556}
    for r in budget_rows['rows']:
        s+=text(budget_rows['right'],r['y'],values[r['field']],14,anchor='end',
                fill='#f1d49a' if r['field']=='rest' else '#f5f1e8',attrs=f'id="value-{r["field"]}"')
    s+=text(448,322,'25%',14,anchor='end',attrs='id="value-pct"')
    actions=[('tax-up','增加稅率檔位','↑'),('tax-down','減少稅率檔位','↓'),
             ('pct-up','增加建造比例','↑'),('pct-down','減少建造比例','↓'),
             ('staff','參謀','參謀'),('end','結束','結束')]
    for box,(act,label,display) in zip(w['boxes'],actions):
        s+=button(box['x'],box['y'],box['w'],box['h'],display,act,act).replace(
            f'aria-label="{display}"',f'aria-label="{label}"')
    return s


def tactical():
    src=Image.open(io.BytesIO(read(Path(reference_paths['battle'])))).convert('RGBA')
    assert src.size==(1280,960)
    # 整個可見戰場與小地圖的內容取正常S1原圖，座標不動。
    field=battle['field'];x,y=field['x'],field['y'];w,h=field['cols']*field['tile_w'],field['rows']*field['tile_h']
    crop=src.crop((x*2,y*2,(x+w)*2,(y+h)*2))
    crop.putdata([(r,g,b,0 if max(r,g,b)<12 else 255) for r,g,b,a in crop.getdata()])
    s=frame(1,41,638,398)+frame(x-2,y-2,w+4,h+4,fill='url(#space)')
    s+=f'<image x="{x}" y="{y}" width="{w}" height="{h}" href="{png_url(crop)}"/>'
    m=battle['minimap']
    s+=frame(465,60,158,158)
    mini=src.crop((m['x']*2,m['y']*2,(m['x']+m['w'])*2,(m['y']+m['h'])*2))
    s+=f'<image x="{m["x"]}" y="{m["y"]}" width="128" height="128" href="{png_url(mini)}"/>'
    f=battle['status']['frame'];s+=frame(f['x'],f['y'],f['w'],f['h'])
    status=battle['status'];rows=status['rows']
    for x,y,v in [(480,rows['phase'],'帝國軍：'),(576,rows['phase'],'輸入'),
                  (480,rows['date'],'宇宙曆：'),(552,rows['date'],'7961010'),
                  (480,rows['turn'],'第１回合'),(480,rows['victory'],'勝利點：'),
                  (480,rows['empire'],'帝國軍：'),(480,rows['alliance'],'同盟軍：'),
                  (480,rows['region'],'亞姆力札')]:
        s+=text(x,y,v,14)
    for y in [rows[k] for k in ['victory','empire','alliance']]:s+=text(608,y,0,14,anchor='end')
    return s


geometry={
 'canvas':[0,0,640,480],
 'budget_window':{k:spec['windows']['report'][k] for k in ['x','y','w','h']},
 'budget_buttons':[{k:b[k] for k in ['x','y','w','h']} for b in spec['windows']['report']['boxes']],
 'map_origin':[spec['map']['x'],spec['map']['y']],
 'regions':32,'region_mask_pixels':counts,'mask_sha256':hashlib.sha256(bytes(n+1 for n in region_grid)).hexdigest(),
 'battle_field':{k:battle['field'][k] for k in ['x','y','cols','rows','tile_w','tile_h']},
 'minimap':{k:battle['minimap'][k] for k in ['x','y','w','h']},
 'comparison':'layout-only；原截圖為正常玩家路徑，試樣為離線呈現，非Game同狀態驗收',
 'remaining':'航路、星名、星球、戰術地形與部隊圖示尚未逐張重繪；本輪示範原配置及面材。',
}
page = r'''<!doctype html><html lang="zh-Hant"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>銀河英雄傳說III SP｜原介面HD加強試樣 A</title><style>
*{box-sizing:border-box}body{margin:0;background:#050c16;color:#e8e5d9;font-family:"Noto Sans CJK TC",sans-serif}
.toolbar{width:min(1280px,100%);margin:20px auto 12px;padding:0 16px;display:flex;align-items:center;gap:10px;flex-wrap:wrap}
.toolbar strong{font-size:17px;margin-right:auto}button{font:inherit;color:inherit;background:#172b40;border:1px solid #655d48;border-radius:2px;padding:9px 14px;cursor:pointer}
button[aria-pressed=true]{border-color:#d4b879;color:#ffe0a0}button:focus-visible,.game-button:focus-visible{outline:2px solid #f5d391;outline-offset:2px}
.screen{width:min(1280px,100%);aspect-ratio:4/3;margin:auto;position:relative;background:#03080e}.screen>svg,.screen>img{display:block;width:100%;height:100%}
svg{font-family:"Noto Sans CJK TC",sans-serif}.game-button{cursor:pointer}.game-button:hover>g>rect:first-child,.game-button:focus>g>rect:first-child{stroke:#f0d394;fill:#344c60}.game-button text{pointer-events:none}
.sector{cursor:pointer}.sector:hover{filter:brightness(1.25)}[hidden]{display:none!important}
footer{width:min(1280px,100%);margin:12px auto 30px;padding:0 16px;font-size:14px;line-height:1.8;color:#b4c2cf}#status{color:#e5cb93;min-height:26px}
@media(max-width:600px){.toolbar{margin-top:12px;gap:6px}.toolbar strong{width:100%;font-size:15px}button{padding:8px 10px;font-size:13px}footer{font-size:12px}}
</style><div class="toolbar"><strong>銀河英雄傳說III SP　原介面HD加強試樣 A</strong>
<button id="view-strategy" data-view="strategy">戰略星圖</button><button id="view-budget" data-view="budget">財政視窗</button><button id="view-battle" data-view="battle">戰場與小地圖</button>
<button id="reference" aria-pressed="false">對照原畫面</button></div>
<main class="screen" aria-label="4:3原配置試樣"><svg id="art" viewBox="0 0 640 480" role="img" aria-label="原配置A風格試樣">__DEFS__<rect width="640" height="480" fill="#03080e"/>
<g id="strategy-layer">__STRATEGY__</g><g id="monthly-layer">__MONTHLY__</g><g id="account-layer">__ACCOUNT__</g><g id="budget-layer">__BUDGET__</g><g id="battle-layer">__BATTLE__</g></svg><img id="original" hidden alt="相同版面的正常遊戲來源截圖"></main>
<footer><div id="status" role="status" aria-live="polite">原配置、原按鈕位置。可操作財政箭頭及結束，也可點星圖選取星域。</div>
可撤回試樣，未接正式遊戲或存檔。星圖區域採原遮罩；航路、星名、星球、戰術地形及部隊圖示仍待重繪。本次示範版面、清晰文字與面材。<br>
原有戰略與戰術座標、疊窗及4:3比例保留。來源畫面分屬開局財政及S1戰場，屬版面對照，沒有宣稱同一局面。</footer>
<script>
const references=__REFERENCES__,geometry=__GEOMETRY__,regionGrid=__REGIONS__,regionNames=__NAMES__;
let view='budget',showReference=false,pct=25,taxLevel=2;
function status(message){document.getElementById('status').textContent=message}
function setView(next){view=next;document.getElementById('strategy-layer').toggleAttribute('hidden',next!=='strategy');document.getElementById('monthly-layer').toggleAttribute('hidden',next!=='budget');document.getElementById('account-layer').toggleAttribute('hidden',next!=='account');document.getElementById('budget-layer').toggleAttribute('hidden',next!=='budget');document.getElementById('battle-layer').toggleAttribute('hidden',next!=='battle');
 for(const b of document.querySelectorAll('[data-view]'))b.setAttribute('aria-pressed',String(b.dataset.view===view));document.getElementById('original').src=references[view];document.getElementById('art').setAttribute('aria-label',{strategy:'戰略星圖',account:'月初會計報告',budget:'財政視窗',battle:'S1戰場與小地圖'}[view]+'，原配置A風格試樣');}
function action(name){if(name==='end'){setView('account');status('預算視窗已關閉，回到月初會計報告。試樣不寫入正式遊戲。');return}
 if(name==='staff'){status('參謀按鈕位置保留；本試樣未連接正式提督資料。');return}if(name.endsWith('-note')){status('原命令列位置保留；本試樣只連接財政視窗。');return}
 if(name.startsWith('pct-')){pct=Math.max(0,Math.min(100,pct+(name.endsWith('up')?1:-1)));document.getElementById('value-pct').textContent=pct+'%';status('建造比例 '+pct+'%。試樣只示範字位與按鈕回饋，財政規則尚未接入。')}
 if(name.startsWith('tax-')){taxLevel=Math.max(0,Math.min(4,taxLevel+(name.endsWith('up')?1:-1)));const labels=['微稅','輕稅','普通','重稅','緊急'];document.querySelectorAll('#budget-layer .source-cell').forEach(e=>{if(+e.getAttribute('y')===294.7){const x=+e.getAttribute('x');if(x===424)e.textContent=labels[taxLevel][0];if(x===440)e.textContent=labels[taxLevel][1]}});status('稅率檔位 '+labels[taxLevel]+'。正式重算與上限沿遊戲規則，本試樣不寫入。')}
}
document.querySelectorAll('[data-view]').forEach(b=>b.addEventListener('click',()=>{setView(b.dataset.view);status('已切換原配置試樣。')}));
document.querySelectorAll('[data-action]').forEach(b=>{b.addEventListener('click',()=>action(b.dataset.action));b.addEventListener('keydown',e=>{if(e.key==='Enter'||e.key===' '){e.preventDefault();action(b.dataset.action)}})});
document.getElementById('reference').addEventListener('click',()=>{showReference=!showReference;document.getElementById('art').toggleAttribute('hidden',showReference);document.getElementById('original').hidden=!showReference;document.getElementById('reference').setAttribute('aria-pressed',String(showReference));});
document.getElementById('strategy-layer').addEventListener('click',e=>{if(view!=='strategy')return;const p=new DOMPoint(e.clientX,e.clientY).matrixTransform(document.getElementById('art').getScreenCTM().inverse());const x=Math.floor(p.x),y=Math.floor(p.y)-40;if(x<0||x>=320||y<0||y>=300)return;const id=regionGrid[y*320+x];if(id<0)return;const source=document.querySelector('[data-region="'+id+'"]');document.getElementById('selected-sector').setAttribute('d',source.getAttribute('d'));status('星域 '+id+'：'+regionNames[id]+'。依原STRMAP.RGN遮罩選取。');});
document.addEventListener('keydown',e=>{if(e.key==='Escape'&&view==='budget')action('end')});setView(new URLSearchParams(location.search).get('view')||'budget');
</script></html>'''
replacements={'__DEFS__':defs,'__STRATEGY__':strategy(),'__MONTHLY__':monthly_base(),'__ACCOUNT__':monthly_base('account'),'__BUDGET__':budget(),'__BATTLE__':tactical(),
              '__REFERENCES__':json.dumps(references),'__GEOMETRY__':json.dumps(geometry),
              '__REGIONS__':json.dumps(region_grid),'__NAMES__':json.dumps([r['name'] for r in stellar],ensure_ascii=False)}
for key,value in replacements.items():page=page.replace(key,value)
(OUT/'HD-ORIGINAL-LAYOUT.html').write_text(page)
(OUT/'GEOMETRY.json').write_text(json.dumps(geometry,ensure_ascii=False,indent=2)+'\n')
sources[str(Path(__file__).relative_to(SRC))]=hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
(OUT/'SOURCES.json').write_text(json.dumps({'input_sha256':sources,'output_sha256':hashlib.sha256(page.encode()).hexdigest(),
 'rights':'原版衍生素材，只留本機；HTML不內嵌字型。','foreground_verified_pixels':len(nonzero)},ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'preview':str(OUT/'HD-ORIGINAL-LAYOUT.html'),'regions':32,'mask_pixels':sum(counts),
                  'foreground_verified_pixels':len(nonzero),'geometry':geometry['budget_window']},ensure_ascii=False))
