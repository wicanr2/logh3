import sys, os
sys.path.insert(0, '/work/tools/scripts')
import png
img = open('/orig/GRAPH.IMG', 'rb').read()

def idx(code):
    lead, trail = code >> 8, code & 0xFF
    t = trail - 0x40 if trail <= 0x7E else trail - 0xA1 + 63
    return (lead - 0xA1) * 157 + t

def strip(name, off, n, per=16):
    nrow = (n + per - 1)//per
    s = [[0]*(per*16) for _ in range(nrow*15)]
    for g in range(n):
        gx, gy = (g % per)*16, (g//per)*15
        for y in range(15):
            hi, lo = img[off+g*30+y*2], img[off+g*30+y*2+1]
            v = hi << 8 | lo
            for x in range(16):
                s[gy+y][gx+x] = (v >> (15-x)) & 1
    png.write_mono(f'/work/workplace/assets/font/{name}.png', per*16, nrow*15, s)
    print(f"{name}: 偏移 0x{off:X} 起 {n} 個")

# 假說一：符號區以 12900 為 base、157/lead
BASE_A = 12900
print("假說一 base=12900：ㄅ(A374) 預測偏移", hex(BASE_A + idx(0xA374)*30))
strip('sym-hypA-bopomofo', BASE_A + idx(0xA374)*30, 48)
strip('sym-hypA-fullA',    BASE_A + idx(0xA2CF)*30, 48)   # 全形Ａ
# 假說二：符號區從 3840 起、157/lead
BASE_B = 3840
print("假說二 base=3840：ㄅ(A374) 預測偏移", hex(BASE_B + idx(0xA374)*30))
strip('sym-hypB-bopomofo', BASE_B + idx(0xA374)*30, 48)
strip('sym-hypB-fullA',    BASE_B + idx(0xA2CF)*30, 48)

# 符號塊（base 3840、471 格）之後到漢字塊（27030）之間的落差
strip('gap-17970-27030', 17970, 302, 32)

# 假說三：0xC6A1 起的假名/西里爾塊接在 17970
def bi(c): return idx(c)
b = 17970 - bi(0xC6A1)*30
for name, code in (('kana-ji', 0xC7B4), ('kana-start', 0xC6A1)):
    print(f"假說三：{name} {code:04X} → 0x{b + bi(code)*30:X}")
    strip('hypC-' + name, b + bi(code)*30, 32)
