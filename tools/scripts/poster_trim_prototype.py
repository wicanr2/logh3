"""154：由正式AppImage原圖建立可丟棄HTML/SVG外框效果稿，不接正式遊戲。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
from PIL import Image,ImageChops,ImageDraw

ROOT=Path('/src');OUT=ROOT/'workplace/poster154'
BASE=ROOT/'workplace/appimage153/player-linux/poster-map.png'
sha=lambda f:hashlib.sha256(f.read_bytes()).hexdigest()
TITLE='銀河英雄傳說 III SP';LEFT='銀河帝國';RIGHT='自由行星同盟';ENGLISH='LEGEND OF THE GALACTIC HEROES'

def svg(body,height,defs=''):
    return f'<svg xmlns="http://www.w3.org/2000/svg" width="960" height="{height}" viewBox="0 0 960 {height}"><defs>{defs}</defs>{body}</svg>'

def page(header,footer):
    return f'''<!doctype html><html lang="zh-Hant"><meta charset="utf-8"><title>154 prototype</title><style>
*{{box-sizing:border-box}}html,body{{margin:0;width:1280px;height:720px;overflow:hidden;background:#000}}
img{{position:absolute;left:0;top:0;width:1280px;height:720px}}
svg{{display:block;font-family:"Noto Sans CJK TC",sans-serif}}.header,.footer{{position:absolute;left:160px;width:960px;overflow:hidden}}
.header{{top:0;height:60px}}.footer{{top:624px;height:96px}}text{{text-anchor:middle}}
</style><img src="source.png"><div class="header">{header}</div><div class="footer">{footer}</div></html>'''

def prepare():
    assert not OUT.exists() and OUT.parent.stat().st_uid==os.getuid()==1000
    OUT.mkdir();shutil.copyfile(BASE,OUT/'source.png')
    font_path=Path('/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc')
    (OUT/'fonts').mkdir();shutil.copyfile(font_path,OUT/'fonts'/font_path.name)
    (OUT/'fonts.conf').write_text('<?xml version="1.0"?><!DOCTYPE fontconfig SYSTEM "urn:fontconfig:fonts.dtd"><fontconfig><dir>/work/fonts</dir><dir>/usr/share/fonts</dir><cachedir>/tmp/poster154-fontcache</cachedir></fontconfig>')
    gradients='''<linearGradient id="bg" x2="0" y2="1"><stop stop-color="#0d1c2e"/><stop offset="1" stop-color="#040b13"/></linearGradient>
<linearGradient id="gold"><stop stop-color="#526276"/><stop offset=".45" stop-color="#b69a62"/><stop offset=".55" stop-color="#ddc68c"/><stop offset="1" stop-color="#526276"/></linearGradient>'''
    a_header=svg(f'''<rect width="960" height="60" fill="url(#bg)"/>
<path d="M0 7H245L253 13H707L715 7H960M0 56H960" fill="none" stroke="#33485e"/>
<path d="M20 29H288L298 25L308 29L298 33L288 29M652 29L662 25L672 29H940M662 25L672 29L662 33L652 29" fill="none" stroke="url(#gold)"/>
<path d="M0 59H960" stroke="#b89b61"/>
<text x="480" y="28" fill="#ead7a7" font-size="20" letter-spacing="4">{TITLE}</text>
<text x="480" y="46" fill="#8ba1b8" font-size="8" letter-spacing="3">{ENGLISH}</text>''',60,gradients)
    a_footer=svg(f'''<rect width="960" height="96" fill="url(#bg)"/>
<path d="M0 1H960M24 87H936" stroke="#33485e"/>
<path d="M16 22H350L364 36H432M528 36H596L610 22H944M16 72H350L364 58H432M528 58H596L610 72H944" fill="none" stroke="url(#gold)"/>
<path d="M480 24L487 41L504 48L487 55L480 72L473 55L456 48L473 41Z" fill="#111f2e" stroke="#bfa469"/>
<path d="M480 34L484 44L494 48L484 52L480 62L476 52L466 48L476 44Z" fill="#d6c18e"/>
<text x="200" y="54" fill="#91a9c0" font-size="13" letter-spacing="5">{LEFT}</text>
<text x="760" y="54" fill="#cfb47e" font-size="13" letter-spacing="4">{RIGHT}</text>
<circle cx="40" cy="47" r="3" fill="#6d8aab"/><circle cx="920" cy="47" r="3" fill="#b9955b"/>
<text x="480" y="91" fill="#698198" font-size="8" letter-spacing="3">III SP</text>''',96,gradients)
    (OUT/'A.html').write_text(page(a_header,a_footer))
    bronze='''<linearGradient id="plate" x2="0" y2="1"><stop stop-color="#3b352b"/><stop offset=".15" stop-color="#242c35"/><stop offset=".65" stop-color="#101a26"/><stop offset="1" stop-color="#060c13"/></linearGradient>
<linearGradient id="rim"><stop stop-color="#304761"/><stop offset=".5" stop-color="#c8af75"/><stop offset="1" stop-color="#80603e"/></linearGradient>'''
    b_header=svg(f'''<rect width="960" height="60" fill="#080e17"/>
<path d="M0 2H960V42L941 59H19L0 42Z" fill="url(#plate)" stroke="url(#rim)" stroke-width="2"/>
<path d="M24 12H253L270 47H690L707 12H936M6 48H245M715 48H954" fill="none" stroke="#775f3f"/>
<path d="M291 7H669L687 52H273Z" fill="#0b1827" stroke="#b39660"/>
<text x="480" y="29" fill="#f0dca6" font-size="20" letter-spacing="4">{TITLE}</text>
<text x="480" y="45" fill="#8da4bd" font-size="8" letter-spacing="2.8">{ENGLISH}</text>
<path d="M40 27H194M40 33H177M40 39H160" stroke="#638aac" stroke-width="2"/>
<path d="M920 27H766M920 33H783M920 39H800" stroke="#b79a61" stroke-width="2"/>
<path d="M224 20L235 32L224 44M736 20L725 32L736 44" fill="none" stroke="#d3b981" stroke-width="2"/>''',60,bronze)
    b_footer=svg(f'''<rect width="960" height="96" fill="#080e17"/>
<path d="M0 15L22 1H938L960 15V94H0Z" fill="url(#plate)" stroke="url(#rim)" stroke-width="2"/>
<path d="M12 25L30 12H385L411 38V87H12ZM948 25L930 12H575L549 38V87H948Z" fill="#0d1825" stroke="#6f6450"/>
<path d="M26 72H112L124 60H360M934 72H848L836 60H600" fill="none" stroke="#486f94"/>
<path d="M26 78H116L128 66H360M934 78H844L832 66H600" fill="none" stroke="#8e734d"/>
<text x="218" y="46" fill="#b5c8dc" font-size="14" letter-spacing="5">{LEFT}</text>
<text x="742" y="46" fill="#d9bd88" font-size="14" letter-spacing="4">{RIGHT}</text>
<circle cx="480" cy="48" r="34" fill="#0a1625" stroke="#ad925c" stroke-width="2"/>
<circle cx="480" cy="48" r="27" fill="none" stroke="#39536d"/>
<path d="M480 17V28M480 68V79M449 48H460M500 48H511" stroke="#b39761"/>
<ellipse cx="480" cy="48" rx="24" ry="11" fill="none" stroke="#526b83" transform="rotate(-28 480 48)"/>
<path d="M480 29L485 43L499 48L485 53L480 67L475 53L461 48L475 43Z" fill="#d2b777"/>
<path d="M33 22H53M63 22H83M877 22H897M907 22H927" stroke="#baa36b" stroke-width="3"/>''',96,bronze)
    (OUT/'B.html').write_text(page(b_header,b_footer))
    (OUT/'SOURCE.json').write_text(json.dumps(dict(prototype=True,production_changed=False,
        source=str(BASE.relative_to(ROOT)),source_sha256=sha(BASE),
        source_receipt_sha256=sha(ROOT/'workplace/appimage153/player-linux/RESULT.json'),
        font_source=str(font_path),font_source_sha256=sha(font_path),font_copy_sha256=sha(OUT/'fonts'/font_path.name),
        output=[1280,720],allowed_rectangles=[[160,0,1120,60],[160,624,1120,720]],
        note='僅戰略畫面靜態HTML/SVG效果稿；正式UI、工具列、畫面狀態與輸入接線另驗。'),ensure_ascii=False,indent=2)+'\n')
    print('READY A/B HTML prototype',flush=True)

def check():
    base=Image.open(BASE).convert('RGB');rows={}
    for name in ['A','B']:
        f=OUT/(name+'.png');im=Image.open(f).convert('RGB');assert im.size==base.size==(1280,720)
        difference=ImageChops.difference(base,im)
        allowed=Image.new('L',base.size);draw=ImageDraw.Draw(allowed)
        for b in [(160,0,1120,60),(160,624,1120,720)]:draw.rectangle((b[0],b[1],b[2]-1,b[3]-1),fill=255)
        outside=ImageChops.multiply(difference,Image.merge('RGB',(ImageChops.invert(allowed),)*3))
        assert outside.getbbox() is None,(name,outside.getbbox())
        assert difference.getbbox() is not None
        rows[name]=dict(path=str(f.relative_to(ROOT)),sha256=sha(f),outside_allowed_different=0,
            html_sha256=sha(OUT/(name+'.html')))
    (OUT/'CHECK.json').write_text(json.dumps(dict(prototype=True,source_sha256=sha(BASE),variants=rows),indent=2)+'\n')
    print(json.dumps(rows,indent=2),flush=True)

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('mode',choices=['prepare','check']);a=p.parse_args();globals()[a.mode]()
