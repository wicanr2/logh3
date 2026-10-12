#!/usr/bin/env python3
"""第126輪：以正式原配置HD畫面製作四款離線HTML概念，不修改遊戲素材。"""
import argparse
import base64
import hashlib
import io
import json
import os
from pathlib import Path

from PIL import Image, ImageChops, ImageDraw, ImageFilter

ROOT = Path('/src')
SCENES = {
    'budget': ('財政預算', '02-budget', 'screen-budget-plan'),
    'build': ('艦艇建造', '07-build', 'screen-build-plan'),
    'strategy': ('戰略星圖', '12-command', 'screen-command-phase'),
    'settings': ('遊戲設定', 'system-selected', 'screen-settings'),
}
THEMES = {
    'empire-a': {
        'faction': 0, 'label': '帝國 A', 'name': '皇室鎏金',
        'lead': '深藍底色，香檳金立體框。',
        'detail': '黃金用在原版金屬框與主要操作，內文保留暖白。深藍面板襯托人物，保留原版浮雕層次。',
        'swatches': ['#0d203c', '#b99b5d', '#f4eddc'],
        'surface': '#0d203c', 'button': '#293e58', 'metal': '#b99b5d', 'white': '#f4eddc',
        'selected': '#8a692e', 'primary': '#344863', 'gradient': 0,
    },
    'empire-b': {
        'faction': 0, 'label': '帝國 B', 'name': '亮藍艦橋',
        'lead': '依你的選擇，B 改為亮藍與高雅黃金。',
        'detail': '保留 B 的藍色框體與明暗層次，提高面板亮度。黃金集中在文字、亮邊與選取狀態，沿用原版框線、字形與按鈕位置。',
        'swatches': ['#1c5393', '#88b4e2', '#fff0c9'],
        'surface': '#1c5393', 'button': '#2f67ab', 'metal': '#88b4e2', 'white': '#fff0c9',
        'selected': '#bb8f36', 'selected_ink': '#182a43', 'primary': '#32629a', 'gradient': 10,
    },
    'alliance-a': {
        'faction': 1, 'label': '同盟 A', 'name': '淡金軍務',
        'lead': '依你的選擇，A 改為淡黃與淡赤紅。',
        'detail': '淡黃色面板搭配柔和赤紅的主要指令與選取狀態。亮底內文使用深褐色，保留原版筆畫、立體框線與按鈕位置。',
        'swatches': ['#e7d7a3', '#eddcb1', '#d89386'],
        'surface': '#e7d7a3', 'button': '#d6bd82', 'metal': '#eddcb1', 'white': '#fff4dc',
        'ink': '#3c2e24', 'selected': '#d89386', 'primary': '#daa398', 'gradient': 0,
    },
    'alliance-b': {
        'faction': 1, 'label': '同盟 B', 'name': '赤焰戰線',
        'lead': '深土黃底色，鮮明熱情紅。',
        'detail': '土黃金屬框搭配更深的褐色面板。高彩度紅色強調操作與選取，讓同盟的畫面更有行動感。',
        'swatches': ['#4b3924', '#b99050', '#ed533b'],
        'surface': '#4b3924', 'button': '#7d5f35', 'metal': '#b99050', 'white': '#ffedca',
        'selected': '#ed533b', 'primary': '#b73d2e', 'gradient': 7,
    },
}


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def rgb(s):
    return tuple(bytes.fromhex(s.lstrip('#')))


def exact_mask(im, value):
    channels = im.split()
    masks = [channel.point([255 if i == c else 0 for i in range(256)]) for channel, c in zip(channels, value)]
    return ImageChops.multiply(ImageChops.multiply(masks[0], masks[1]), masks[2])


def rect(w):
    return [w[k] for k in ('x', 'y', 'w', 'h')]


def fill_rect(draw, r, fill):
    x, y, w, h = r
    draw.rectangle((x*3, y*3, (x+w)*3-1, (y+h)*3-1), fill=fill)


def ui_geometry(spec):
    board = spec['windows']['board']
    mask = Image.new('L', (1920, 1440))
    draw = ImageDraw.Draw(mask)
    fill_rect(draw, rect(board), 255)
    fill_rect(draw, [spec['map']['x'], spec['map']['y'], 320, 300], 0)
    windows, buttons = [], []
    for name in spec.get('order', ['report']):
        if name in spec['windows']:
            fill_rect(draw, rect(spec['windows'][name]), 255)
    for name, w in spec['windows'].items():
        windows.append({'name': name, 'rect': rect(w)})
        if w.get('face'):
            f = w['face']
            fill_rect(draw, [f['x'], f['y'], 80, 100], 0)
        for box in w.get('boxes', []):
            if box['h'] <= 24 and box['w'] <= 140 and (box.get('text') or box['w'] == 16):
                buttons.append({'rect': rect(box), 'label': box.get('text') or '調整箭頭'})
    return mask, windows, buttons


def recolour(source, mask, theme, buttons, windows):
    out = source.copy()
    # 只置換UI遮罩內的已知原色；星圖、人物與未識別色保留來源。
    replacements = {
        (0, 81, 97): theme['surface'],
        (65, 81, 97): theme['button'],
        (178, 178, 195): theme['metal'],
        (243, 243, 243): theme['white'],
        (227, 32, 32): theme['selected'],
    }
    for old, new in replacements.items():
        selected = ImageChops.multiply(exact_mask(source, old), mask)
        layer = Image.new('RGB', source.size, rgb(new))
        if old == (0, 81, 97) and theme['gradient']:
            painter = ImageDraw.Draw(layer)
            for y in range(source.height):
                light = round(theme['gradient'] * (1-y/source.height))
                painter.line((0, y, source.width, y), fill=tuple(min(255, c+light) for c in rgb(new)))
        out.paste(layer, (0, 0), selected)
    primary = Image.new('L', source.size)
    painter = ImageDraw.Draw(primary)
    for b in buttons:
        if b['label'] in ('結束', '參謀', '決定', '實行'):
            fill_rect(painter, b['rect'], 255)
    selected = ImageChops.multiply(exact_mask(source, (65, 81, 97)), primary)
    out.paste(Image.new('RGB', source.size, rgb(theme['primary'])), (0, 0), selected)
    if theme.get('ink') or theme.get('selected_ink'):
        # 淡色面板的字用深色；框的亮邊仍保留，黑底資訊也保留亮字。
        fill = exact_mask(source, (227, 32, 32))
        if theme.get('ink'):
            fill = ImageChops.lighter(fill, exact_mask(source, (0, 81, 97)))
            fill = ImageChops.lighter(fill, exact_mask(source, (65, 81, 97)))
        near_fill = fill.filter(ImageFilter.MaxFilter(13))
        edges = Image.new('L', source.size)
        painter = ImageDraw.Draw(edges)
        for item in windows + buttons:
            x, y, w, h = item['rect']
            painter.rectangle((x*3, y*3, (x+w)*3-1, (y+h)*3-1), outline=255, width=6)
        ink = ImageChops.multiply(exact_mask(source, (243, 243, 243)), mask)
        ink = ImageChops.multiply(ink, ImageChops.subtract(near_fill, edges))
        out.paste(Image.new('RGB', source.size, rgb(theme.get('ink') or theme['selected_ink'])), (0, 0), ink)
    return out


def data_url(image):
    stream = io.BytesIO()
    image.save(stream, format='PNG', compress_level=6)
    return 'data:image/png;base64,'+base64.b64encode(stream.getvalue()).decode()


HTML = r'''<!doctype html>
<html lang="zh-Hant"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>銀河英雄傳說 III SP｜陣營 HD 概念確認</title>
<style>
:root{color-scheme:dark;--ink:#efe9dd;--muted:#a4a6ad;--edge:#343943;--accent:#cfb578}
*{box-sizing:border-box}[hidden]{display:none!important}body{margin:0;background:#101216;color:var(--ink);font:15px/1.6 system-ui,'Noto Sans TC',sans-serif}
main{max-width:1480px;margin:auto;padding:32px 32px 44px}header{display:flex;gap:24px;align-items:flex-end;justify-content:space-between;margin-bottom:28px}
.eyebrow{font-size:12px;letter-spacing:.16em;color:var(--accent)}h1{margin:4px 0 8px;font-size:clamp(25px,3vw,38px);font-weight:600;letter-spacing:.04em}p{margin:0;color:var(--muted)}
button,select{font:inherit}button{border:1px solid var(--edge);border-radius:5px;color:var(--ink);background:#20242b;padding:8px 14px;cursor:pointer}button:hover{border-color:var(--accent)}button:focus-visible,select:focus-visible,input:focus-visible{outline:2px solid #ffe1a2;outline-offset:3px}button[aria-pressed=true]{background:#3d3424;border-color:var(--accent);color:#ffe8b5}
.tag{font-size:12px;white-space:nowrap;border:1px solid var(--edge);padding:5px 12px;border-radius:20px;color:#b6b9c1}
.cards{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:14px;margin:20px 0 28px}.card{padding:0;overflow:hidden;text-align:left;background:#171b21;border-radius:7px}.card img{display:block;width:100%;aspect-ratio:4/3;object-fit:contain;background:#000}.card .caption{padding:12px 14px}.card strong{font-size:17px;font-weight:600}.card small{display:block;color:#aeb1b9;line-height:1.5;margin-top:4px}.swatches{display:flex;gap:5px;margin-top:10px}.swatches i{width:24px;height:6px;border-radius:2px}
.toolbar{display:flex;align-items:center;flex-wrap:wrap;gap:9px;padding:14px 0;border-top:1px solid var(--edge)}.toolbar .push{margin-left:auto}label{display:inline-flex;align-items:center;gap:7px;cursor:pointer}input{accent-color:var(--accent)}
.workspace{display:grid;grid-template-columns:minmax(0,1fr) 258px;gap:24px;align-items:start}.screen-scroll{min-width:0;overflow:auto;background:#000;border:1px solid #58606e;border-radius:3px}.screen{position:relative;width:100%;aspect-ratio:4/3}.screen img{display:block;width:100%;height:100%}.screen.zoom{width:1920px;height:1440px}.screen svg,.hit-layer{position:absolute;inset:0;width:100%;height:100%}.screen svg{pointer-events:none}svg rect,svg line{vector-effect:non-scaling-stroke;stroke:#7de7db;stroke-width:1;fill:none;stroke-dasharray:5 3}svg .bottom{stroke:#ffb76b;stroke-dasharray:none}svg text{fill:#7de7db;font-size:7px;paint-order:stroke;stroke:#050c14;stroke-width:2px}
.hit{position:absolute;background:transparent;border:0;border-radius:0;padding:0;min-width:0;min-height:0;color:transparent}.hit:hover,.hit:focus-visible{background:#ffe59a20;outline:1px solid #ffe1a2;outline-offset:-1px}.hit.clicked{box-shadow:inset 0 0 0 2px #ffdf8b;background:#ffdf8b44}
.sidebar{padding:18px;background:#181c22;border:1px solid var(--edge);border-radius:6px}.sidebar h2{font-size:22px;line-height:1.3;margin:3px 0 12px}.sidebar p{font-size:14px;line-height:1.8;color:#c0c2ca}.sidebar .facts{margin:20px 0;padding:16px 0;border-top:1px solid var(--edge);border-bottom:1px solid var(--edge)}.facts div{display:flex;justify-content:space-between;margin:6px 0;font-size:13px}.facts span{color:#979da8}.sidebar button{width:100%;margin-top:10px}.choice{font-size:13px!important;margin-top:12px!important;color:#d6bb7e!important}.status{min-height:26px;margin:12px 0;color:#d6c59e;font-size:13px}footer{border-top:1px solid var(--edge);margin-top:25px;padding-top:17px;color:#999fa9;font-size:13px}.decision{color:#dfc78b;font-size:15px;margin:14px 0 5px}.original .tag{color:#99cfc9}
@media(max-width:1050px){.cards{grid-template-columns:repeat(2,minmax(0,1fr))}.workspace{grid-template-columns:minmax(0,1fr) 225px}main{padding:24px}.sidebar{padding:14px}}
@media(max-width:700px){main{padding:17px 14px}.workspace{display:block}.sidebar{margin-top:16px}header{align-items:flex-start;gap:10px;flex-direction:column}.toolbar .push{margin-left:0}.cards{gap:9px;margin-bottom:20px}.card .caption{padding:9px}.card strong{font-size:15px}.card small{font-size:12px}.toolbar button{padding:6px 10px}}
</style>
<main>
<header><div><div class="eyebrow">銀河英雄傳說 III SP · 操作介面概念</div><h1>同一個銀河，兩種陣營氣質。</h1><p>已依你的選擇調整：帝國 B 亮藍與黃金，同盟 A 淡黃與淡赤紅。</p></div><span class="tag">離線概念圖 · 配色修訂</span></header>
<div class="cards" id="cards" aria-label="四套陣營外觀"></div>
<div class="toolbar" aria-label="畫面切換" id="scene-nav"></div>
<div class="toolbar"><label><input id="guides" type="checkbox">顯示框線與按鈕對齊線</label><label><input id="original" type="checkbox">對照原配色</label><button class="push" id="zoom">檢查原尺寸</button><button id="fullscreen">全螢幕</button></div>
<div class="workspace"><div><div class="screen-scroll"><div class="screen" id="screen"><img id="preview" alt="陣營介面概念圖"><svg viewBox="0 0 640 480" id="guide-layer" hidden aria-hidden="true"></svg><div class="hit-layer" id="hit-layer"></div></div></div><div class="status" id="status" role="status">可試點畫面中的按鈕，檢查按鈕與外框的位置。</div></div>
<aside class="sidebar"><div class="eyebrow" id="faction-label"></div><h2 id="theme-name"></h2><p id="theme-detail"></p><div class="swatches" id="large-swatches"></div><div class="facts"><div><span>視窗與按鈕</span>沿原版位置</div><div><span>外框</span>保留立體層次</div><div><span>字形</span>沿原版筆畫</div><div><span>畫面比例</span>4 : 3</div><div><span>概念原尺寸</span>1920 × 1440</div></div><button id="choose">選擇這套概念</button><p class="choice" id="selection">帝國：尚未選擇<br>同盟：尚未選擇</p></aside></div>
<footer><p class="decision">目前方向：帝國 B 亮藍與黃金，同盟 A 淡黃與淡赤紅。</p><p>可切換四個操作畫面，檢查調亮後的字色、框線與按鈕底緣。另兩款保留比較。概念用正常玩家路徑的實際 HD 畫面製作；此頁用於確認外觀與按鈕位置，遊戲及新版影片另行製作。</p></footer>
</main>
<script id="concept-data" type="application/json">__DATA__</script>
<script>
const data=JSON.parse(document.getElementById('concept-data').textContent);
const themes=data.themes, scenes=data.scenes, images=data.images, originalImages=data.original;
let active='empire-b',view='budget';const chosen={0:'empire-b',1:'alliance-a'};
const $=id=>document.getElementById(id);
function swatches(t){return t.swatches.map(c=>'<i style="background:'+c+'" title="'+c+'"></i>').join('')}
for(const [id,t] of Object.entries(themes)){
 const b=document.createElement('button');b.className='card';b.dataset.theme=id;b.setAttribute('aria-pressed','false');
 b.innerHTML='<img src="'+images[id].budget+'" alt="'+t.label+' '+t.name+' 財政畫面"><div class="caption"><strong>'+t.label+' · '+t.name+'</strong><small>'+t.lead+'</small><div class="swatches">'+swatches(t)+'</div></div>';
 b.onclick=()=>setTheme(id);$('cards').appendChild(b);
}
for(const [id,s] of Object.entries(scenes)){
 const b=document.createElement('button');b.dataset.scene=id;b.textContent=s.label;b.onclick=()=>setScene(id);$('scene-nav').appendChild(b);
}
function setTheme(id){active=id;render()}
function setScene(id){view=id;render()}
function render(){
 const t=themes[active],s=scenes[view],o=$('original').checked;
 $('preview').src=o?originalImages[t.faction][view]:images[active][view];
 $('preview').alt=t.label+' '+t.name+' '+s.label+(o?' 原配色對照':'');
 $('faction-label').textContent=active==='empire-b'?'銀河帝國 · 亮藍與高雅黃金':active==='alliance-a'?'自由行星同盟 · 淡黃與淡赤紅':t.faction===0?'銀河帝國 · 深藍與高雅黃金':'自由行星同盟 · 土黃與熱情紅';
 $('theme-name').textContent=t.label+' · '+t.name;$('theme-detail').textContent=t.detail;$('large-swatches').innerHTML=swatches(t);
 $('selection').innerHTML=[0,1].map(f=>(f===0?'帝國':'同盟')+'：'+(chosen[f]?themes[chosen[f]].label+' '+themes[chosen[f]].name:'尚未選擇')).join('<br>');
 document.querySelectorAll('[data-theme]').forEach(b=>b.setAttribute('aria-pressed',String(b.dataset.theme===active)));
 document.querySelectorAll('[data-scene]').forEach(b=>b.setAttribute('aria-pressed',String(b.dataset.scene===view)));
 $('guide-layer').innerHTML=s.windows.map(w=>{const [x,y,wid,hei]=w.rect;return '<rect x="'+x+'" y="'+y+'" width="'+wid+'" height="'+hei+'"/><line class="bottom" x1="'+x+'" y1="'+(y+hei)+'" x2="'+(x+wid)+'" y2="'+(y+hei)+'"/>'}).join('');
 $('hit-layer').replaceChildren();
 for(const b of s.buttons){
  const [x,y,w,h]=b.rect;const hit=document.createElement('button');hit.className='hit';hit.setAttribute('aria-label',b.label+' 位置試點');hit.dataset.rect=b.rect.join(',');
  hit.style.cssText='left:'+x/640*100+'%;top:'+y/480*100+'%;width:'+w/640*100+'%;height:'+h/480*100+'%';
  hit.onclick=()=>{hit.classList.add('clicked');$('status').textContent='已試點「'+b.label+'」：按鈕與外框共用 ('+b.rect.join(', ')+')。';setTimeout(()=>hit.classList.remove('clicked'),400)};
  $('hit-layer').appendChild(hit);$('guide-layer').insertAdjacentHTML('beforeend','<rect x="'+x+'" y="'+y+'" width="'+w+'" height="'+h+'"/>');
 }
 $('guide-layer').toggleAttribute('hidden',!$('guides').checked);
 $('status').textContent=t.label+' · '+s.label+'。可開啟對齊線，或試點畫面中的按鈕。';
}
$('guides').onchange=()=>{$('guide-layer').toggleAttribute('hidden',!$('guides').checked)};$('original').onchange=render;
$('zoom').onclick=()=>{const on=$('screen').classList.toggle('zoom');$('zoom').textContent=on?'縮回適合視窗':'檢查原尺寸'};
$('fullscreen').onclick=()=>{if(!document.fullscreenElement)$('screen').parentElement.requestFullscreen?.();else document.exitFullscreen?.()};
$('choose').onclick=()=>{chosen[themes[active].faction]=active;$('selection').innerHTML=[0,1].map(f=>(f===0?'帝國':'同盟')+'：'+(chosen[f]?themes[chosen[f]].label+' '+themes[chosen[f]].name:'尚未選擇')).join('<br>');$('status').textContent='已暫選 '+themes[active].label+'。請把帝國與同盟的選擇回覆給我。'};
document.addEventListener('keydown',e=>{if(e.key==='Escape'){$('screen').classList.remove('zoom');$('zoom').textContent='檢查原尺寸'}});
const params=new URLSearchParams(location.search);if(themes[params.get('theme')])active=params.get('theme');if(scenes[params.get('scene')])view=params.get('scene');render();
</script></html>'''


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--root', type=Path, default=Path('/work/hd126'))
    args = p.parse_args()
    root = args.root
    assert os.getuid() == 1000 and root.stat().st_uid == 1000
    sources, originals, scenes, geometry = [], {}, {}, {}
    for key, (label, stem, spec_name) in SCENES.items():
        path = ROOT/'docs/internals/data'/(spec_name+'.json')
        spec = json.loads(path.read_text())
        mask, windows, buttons = ui_geometry(spec)
        if key == 'settings':
            buttons.append({'rect': [176, 360, 128, 20], 'label': '顯示敵軍 OFF F7'})
        geometry[key] = (mask, buttons, windows)
        scenes[key] = {'label': label, 'windows': windows, 'buttons': buttons}
        sources.append({'path': str(path.relative_to(ROOT)), 'sha256': sha(path)})
    for faction, folder in [(0, 'classic-empire'), (1, 'classic-alliance')]:
        originals[faction] = {}
        for key, (_, stem, _) in SCENES.items():
            path = root/folder/(stem+'.png')
            im = Image.open(path).convert('RGB')
            assert im.size == (1920, 1440)
            originals[faction][key] = data_url(im)
            sources.append({'path': str(path), 'sha256': sha(path)})
    images = {}
    for ident, theme in THEMES.items():
        images[ident] = {}
        for key, (_, stem, _) in SCENES.items():
            folder = 'classic-empire' if theme['faction'] == 0 else 'classic-alliance'
            source = Image.open(root/folder/(stem+'.png')).convert('RGB')
            mask, buttons, windows = geometry[key]
            image = recolour(source, mask, theme, buttons, windows)
            images[ident][key] = data_url(image)
            if key == 'budget':
                image.save(root/(ident+'.png'))
    payload = {'themes': THEMES, 'scenes': scenes, 'images': images, 'original': originals}
    html = root/'FACTION-HD-CONCEPTS.html'
    html.write_text(HTML.replace('__DATA__', json.dumps(payload, ensure_ascii=False, separators=(',', ':'))))
    manifest = {'scope': '離線概念，四套外觀尚未接正式遊戲', 'revision': 2,
                'user_decisions': {'empire': 'B為基礎，改亮藍與高雅黃金', 'alliance': 'A為基礎，改淡黃與淡赤紅'},
                'generator_sha256': sha(Path(__file__)),
                'sources': sources, 'html_sha256': sha(html), 'html_bytes': html.stat().st_size,
                'themes': THEMES, 'scenes': scenes, 'external_requests': False}
    (root/'CONCEPT-SOURCES.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({'html': str(html), 'bytes': html.stat().st_size, 'themes': len(THEMES), 'scenes': len(SCENES)}, ensure_ascii=False))


if __name__ == '__main__':
    main()
