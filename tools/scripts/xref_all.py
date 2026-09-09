"""在整份程式碼傾印裡找 far call 的呼叫端。

    xref_all.py <傾印> <傾印起始線性位址hex> <段:位移> [段:位移 ...]

傾印是 `-dump-mem lin:...` 倒出來的連續記憶體，所以 `9A off16 seg16`
可以直接逐位元組掃。輸出把每個命中換算回「段:位移」——段值取離命中點
最近、且位移落在 0..0xFFFF 的常見段起點沒有意義，這裡改成列出線性
位址，由呼叫端自己決定用哪個段基準去讀。
"""
import sys

dump = open(sys.argv[1], "rb").read()
base = int(sys.argv[2], 16)
for t in sys.argv[3:]:
    seg, off = (int(x, 16) for x in t.split(":"))
    pat = bytes([0x9A, off & 0xFF, off >> 8, seg & 0xFF, seg >> 8])
    hits = []
    i = 0
    while True:
        i = dump.find(pat, i)
        if i < 0:
            break
        hits.append(base + i)
        i += 1
    print("%04X:%04X  %d 處" % (seg, off, len(hits)))
    for h in hits:
        print("    lin %05X" % h)
