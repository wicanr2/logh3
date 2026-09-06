"""解 PLADAT.DAT（行星資料）並輸出 JSON。

欄位語意大多是**假說**——由數值範圍與遊戲常識推的，沒有對過遊戲畫面。
名稱與記錄長度是 confirmed（名稱解得出中文、82 bytes 整除）。
"""
import json, os
d = open('/orig/PLADAT.DAT', 'rb').read()
HDR, R, NAME = 64, 82, 20
body = d[HDR:]
assert len(body) % R == 0, len(body)

def name(b):
    """名稱是 Big5 全形字串，遇到非 Big5 前導位元組就結束。

    **不能用「跳過填充」的寫法**：填充有 0x0000／0x0040／0xA140 三種，
    而 10 個字的名字填滿整個欄位、後面直接接數值——跳過填充的話那些數值
    會被當成名字的一部分解出亂碼。"""
    s = ''
    for i in range(0, len(b) - 1, 2):
        lead = b[i]
        if not (0xA1 <= lead <= 0xF9):
            break
        try:
            s += b[i:i+2].decode('big5hkscs')
        except Exception:
            break
    return s

def u16(b, o): return int.from_bytes(b[o:o+2], 'little')
def u32(b, o): return int.from_bytes(b[o:o+4], 'little')

recs = []
for r in range(len(body)//R):
    rec = body[r*R:(r+1)*R]
    recs.append(dict(
        index=r, name=name(rec[:NAME]),
        name_bytes=rec[:NAME].hex(),
        w14=u16(rec,0x14), w16=u16(rec,0x16), w18=u16(rec,0x18), w1a=u16(rec,0x1A),
        w1c=u16(rec,0x1C), w1e=u16(rec,0x1E),
        d20=u32(rec,0x20), d24=u32(rec,0x24), d28=u32(rec,0x28), d2c=u32(rec,0x2C),
        d30=u32(rec,0x30), d34=u32(rec,0x34), d38=u32(rec,0x38),
        d3c=u32(rec,0x3C), d40=u32(rec,0x40), d44=u32(rec,0x44),
        rest=rec[0x48:].hex()))

os.makedirs('/work/docs/internals/data', exist_ok=True)
json.dump(dict(
    file='PLADAT.DAT', bytes=len(d),
    header=dict(bytes=HDR, words=[int.from_bytes(d[i:i+2],'little') for i in range(0,HDR,2)],
                note='開機時被單獨讀走的 64 bytes；32 個 word，值都在 1–5，疑似每個星域的行星數或等級（假說）'),
    record=dict(count=len(recs), size=R,
                name=dict(offset=0, bytes=NAME, encoding='Big5（ETen），全形空白補齊'),
                note='名稱與記錄長度 confirmed；其餘欄位的語意是假說，沒有對過遊戲畫面'),
    records=recs), open('/work/docs/internals/data/pladat.json','w'),
    ensure_ascii=False, indent=1)
print(f"{len(recs)} 筆，寫出 docs/internals/data/pladat.json")
print("前 12 個名稱:", [r['name'] for r in recs[:12]])
print("最後 6 個:", [r['name'] for r in recs[-6:]])
