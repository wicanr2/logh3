"""在固定HD工具image產生啟動畫面的授權字集，非原版遊戲資產。"""
from pathlib import Path
import hashlib
import json
import os
from PIL import Image, ImageDraw, ImageFont

out = Path('/work/modal131/startup-font')
assert out.parent.is_dir() and out.parent.stat().st_uid == os.getuid()
out.mkdir(exist_ok=True)
source = Path('/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc')
face = ImageFont.truetype(str(source), 20, index=3)
text = '銀河英雄傳說III SP遊戲啟動中準備載入原版資料開場動畫存檔初始局HD圖像與字型音樂音效完成正在人物張階段已略過讀取失敗無法找不到請檢查路徑詳細錯誤結束按關閉視窗或返回先取消資料夾權限設定使用者磁碟可用空間重試□正在載入，請稍候'
chars = sorted(set(text + ''.join(chr(i) for i in range(32,127))))
atlas = Image.new('RGBA', (32*24, ((len(chars)+31)//32)*24))
draw = ImageDraw.Draw(atlas)
glyphs = {}
for n, char in enumerate(chars):
    x, y = n % 32*24, n//32*24
    box = face.getbbox(char)
    width = 11 if ord(char) < 128 else 22
    draw.text((x+(width-(box[2]-box[0]))//2-box[0], y+19-face.getmetrics()[0]), char, font=face, fill='white')
    glyphs[char] = [x,y,width,24]
atlas.save(out/'startup_font.png')
(out/'startup_font.json').write_text(json.dumps(glyphs,ensure_ascii=False))
license_path = Path('/usr/share/doc/fonts-noto-cjk/copyright')
(out/'startup_font.LICENSE.txt').write_bytes(license_path.read_bytes())
manifest = {'source':str(source),'source_sha256':hashlib.sha256(source.read_bytes()).hexdigest(),
            'face_index':3,'generator':'tools/scripts/startup_font.py','license':'SIL Open Font License 1.1',
            'generator_sha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
            'files':{name:hashlib.sha256((out/name).read_bytes()).hexdigest() for name in ['startup_font.png','startup_font.json','startup_font.LICENSE.txt']}}
(out/'SOURCE.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
print(out, len(chars))
