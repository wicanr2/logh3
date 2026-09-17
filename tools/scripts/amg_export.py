# FACE.AMG → 83 張 PNG（4bpp 調色盤圖）＋ 格式 JSON。
# PNG 落在 workplace/assets/face/，進版控的那份複製到 assets/original/face/
# （容器看不到 repo 根目錄）；JSON 只放版面、索引與雜湊。
#
# 第 n 張 ＝ 提督索引 n（`ADMIRAL.DT*` 第 n 筆、`KEIREKI.DAT` 第 n 段），
# 證據見 docs/internals/09 與 docs/re/271。提督姓名與陣營從
# docs/internals/data/keireki.json 取（那份由 keireki_export.py 產生）。
import hashlib, json, os, struct, sys, zlib
sys.path.insert(0, '/work/tools/scripts')
from amg_decode import decode

def write_png(path, w, h, px, pal):
    raw = b''.join(b'\x00' + px[y*w:(y+1)*w] for y in range(h))
    def chunk(t, d):
        c = t + d
        return struct.pack('>I', len(d)) + c + struct.pack('>I', zlib.crc32(c))
    body = b'\x89PNG\r\n\x1a\n'
    body += chunk(b'IHDR', struct.pack('>IIBBBBB', w, h, 8, 3, 0, 0, 0))
    body += chunk(b'PLTE', bytes(pal))
    body += chunk(b'IDAT', zlib.compress(raw, 9))
    body += chunk(b'IEND', b'')
    with open(path, 'wb') as fp:
        fp.write(body)
    return body

# 調色盤取自 oracle 的畫面 PNG：色號 → 屬性暫存器 → DAC 這條鏈由執行器
# 走完了，`pal.bin` 只有最後一段（DAC），直接拿會得到錯的顏色。
png = open('/work/workplace/oracle/face-fresh/screen.png', 'rb').read()
i = 8
pal = None
while i < len(png):
    n = struct.unpack_from('>I', png, i)[0]
    t = png[i+4:i+8]
    if t == b'PLTE':
        pal = list(png[i+8:i+8+n])
        break
    i += 12 + n
pal = (pal + [0]*768)[:48]

# 提督姓名與陣營（缺檔就留空，不自製代用品）
KEIREKI = '/work/docs/internals/data/keireki.json'
who = {}
if os.path.exists(KEIREKI):
    for r in json.load(open(KEIREKI))['records']:
        who[r['index']] = (r.get('name') or '', r.get('side'))
else:
    print('⚠ 缺 %s，姓名與陣營欄會是空的（先跑 keireki_export.py）' % KEIREKI)

GREEN = 4          # #306100，同盟制服的綠；帝國制服不用這個色號

f = open('/orig/FACE.AMG', 'rb').read()
idx = [struct.unpack_from('<HI', f, k*6) for k in range(83)]
out = '/work/workplace/assets/face'
os.makedirs(out, exist_ok=True)
meta = []
for k, (s, o) in enumerate(idx):
    img = f[o:o+s]
    W, H, px, used = decode(img)
    body = write_png(f'{out}/face-{k:02d}.png', W, H, px, pal)
    a, b = struct.unpack_from('<II', img, 0)
    colors = sorted(set(px))
    green = px.count(GREEN)
    name, side = who.get(k, ('', None))
    m = {'index': k, 'size': s, 'offset': o, 'w': W, 'h': H,
         'used': used, 'exact': used == s,
         'hdr0': a, 'hdr4': b,
         # +12 是壓縮資料的第一個位元組，不是旗標——先前寫成 flag12／
         # 「0xFF ＝ 空白」是誤讀，見 docs/re/271。留著是為了讓那個更正
         # 可以重算：值為 0xFF 的八張只是外框頂列的色號與其他張不同。
         'data0': img[12],
         'colors': colors,
         'solid': len(colors) == 1,
         'green_px': green,
         'uniform': None if len(colors) == 1 else ('alliance' if green > 100 else 'empire'),
         'sha256_pixels': hashlib.sha256(px).hexdigest(),
         'sha256_png': hashlib.sha256(body).hexdigest()}
    if k < 82:
        m['admiral'] = k
        m['name'] = name
        m['side'] = side
    else:
        m['admiral'] = None
        m['name'] = ''
        m['side'] = None
        m['note'] = '第 83 格：整張單色 9（外框色），沒有人物；KEIREKI.DAT 的第 83 段同樣是空的'
    meta.append(m)

json.dump({'source': 'FACE.AMG',
           'sha256': hashlib.sha256(f).hexdigest(),
           'count': len(idx),
           'index_entry': {'size': 'u16 @+0', 'offset': 'u32 @+2'},
           'image_header': {'+0': 'u32 固定 1', '+4': 'u32 size-6',
                            '+8': 'u16 直行數（每行 8 像素）',
                            '+10': 'u16 每直行的 word 數（每 word 2 列）',
                            '+12': '壓縮資料的第一個位元組（**不是旗標**，見 docs/re/271）'},
           'mapping': {
               'rule': '第 n 張 ＝ 提督索引 n（0–81）；第 82 張是沒有人物的單色格',
               'level': 'confirmed',
               'evidence': [
                   '逐位元組收據：0（防拷照片 face-fresh）、1／17（提督詳細資料 adm2e／adm15e）、'
                   '6（月底船艦接收報告 turnmonth2）、12（戰術艦隊詳細資料 s3detail）、'
                   '13（戰略情報報告 info1e、易主報告 chain3）、16（月初會計報告 key1）、'
                   '60（同盟月初會計報告 keyAlly）',
                   '制服分界：0–39 帝國制服、40–79 同盟制服（色 4）、80–81 帝國制服；'
                   '與 ADMIRAL.DTX +1 的陣營欄一致，兩個例外 80／81 正是投奔同盟的'
                   '梅爾卡茲與舒奈德（KEIREKI.DAT 第 80／81 段自己寫著）',
                   '內容對應：0 金髮、1 火紅髮、2 蜂蜜色髮、3 右眼黑左眼藍、'
                   '34 鬍鬚、49 淺黑膚色、74 淡綠眼——都與同編號的人物傳字面相符',
                   'ADMIRAL.DT* 的 28 個位移沒有任何一欄等於記錄索引，所以頭像編號'
                   '不是提督記錄裡的欄位，就是索引本身',
               ],
               'not_yet': '原版繪圖端「索引怎麼算出來」的那段程式碼還沒從反組譯讀出來',
           },
           'hash_note': 'sha256_pixels 是解碼後 w×h 個色號位元組的雜湊，與 PNG 編碼器無關，'
                        '引用時以它為準；sha256_png 只在同一個 zlib 版本下才重現。',
           'images': meta},
          open('/work/docs/internals/data/face-amg.json', 'w'),
          ensure_ascii=False, indent=1)
print('導出', len(meta), '張 →', out)
print('全部剛好用完:', all(m['exact'] for m in meta))
print('單色（沒有人物）的格:', [m['index'] for m in meta if m['solid']])
print('制服分界:', [(m['index'], m['uniform']) for m in meta
                    if m['index'] in (0, 39, 40, 79, 80, 81)])
