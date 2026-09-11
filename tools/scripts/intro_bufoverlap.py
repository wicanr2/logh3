"""驗證開頭動畫的兩個讀取緩衝區重疊 40 個位元組。

    tools/py.sh tools/scripts/intro_bufoverlap.py

輸入是收據 `openmem` 的 1 MB 記憶體傾印（`workplace/openmem.bin`，
`OPEN.EXE` 載入段為 0，所以線性位址 ＝ 段 × 16 ＋ 位移）與原版的
`IBM16.350`。只讀不寫。
"""
import sys

dump = open("/work/workplace/openmem.bin", "rb").read()
film = open("/orig/IBM16.350", "rb").read()

def word(lin):
    return dump[lin] | dump[lin + 1] << 8

DS = 0x070A * 16                      # 載入器的資料段（`docs/re/116`）
a_off, a_seg = word(DS + 0x4F6), word(DS + 0x4F8)
b_off, b_seg = word(DS + 0x4FC), word(DS + 0x4FE)
a_lin, b_lin = a_seg * 16 + a_off, b_seg * 16 + b_off
print("緩衝區 A = %04X:%04X → 線性 %05X" % (a_seg, a_off, a_lin))
print("緩衝區 B = %04X:%04X → 線性 %05X" % (b_seg, b_off, b_lin))
print("兩者相差 %d 個位元組" % (b_lin - a_lin))

lens = [word(DS + 0x3C0 + 2 * i) for i in range(30)]
frames = [word(DS + 2 * i) for i in range(30)]
print("分塊長度表 30 筆總和 = %d，IBM16.350 = %d，%s"
      % (sum(lens), len(film), "相符" if sum(lens) == len(film) else "不符"))
print("每塊格數表 30 筆總和 = %d（第 0 塊 %d 格）" % (sum(frames), frames[0]))

gap = b_lin - a_lin
ok_head = dump[a_lin:a_lin + gap] == film[:gap]
c1 = lens[0]                                   # 第 1 塊在檔案裡的起點
n = lens[1]
ok_rest = dump[b_lin:b_lin + n] == film[c1:c1 + n]
print("記憶體 %05X 起 %d bytes ＝ 檔案[0:%d]：%s" % (a_lin, gap, gap, ok_head))
print("記憶體 %05X 起 %d bytes ＝ 檔案[%d:%d]：%s"
      % (b_lin, n, c1, c1 + n, ok_rest))

# 解碼器從緩衝區 A 讀到的整條資料流
stream = dump[a_lin:a_lin + gap + n]
expect = film[:gap] + film[c1:c1 + n]
print("解碼器看到的資料流 ＝ 檔案[0:%d] ＋ 檔案[%d:]：%s"
      % (gap, c1, stream == expect))
if gap < lens[0]:
    print("因此第 0 塊的 %d 格裡，只有第 0 格的前 %d 個位元組被解到，"
          "其餘 %d 格永遠不會出現" % (frames[0], gap, frames[0] - 1))
else:
    # `docs/re/283`：兩個緩衝區在修好的執行器上相差一整塊，不重疊。
    # 舊傾印會走上面那一路——**先確認 openmem 是哪一版執行器錄的**。
    print("兩塊不重疊（相差 %d ≥ 第 0 塊的 %d bytes），沒有覆蓋問題"
          % (gap, lens[0]))

# 用 `IBMFLS16.350` 的逐格長度回算每一塊裝了幾格
fls = open("/orig/IBMFLS16.350", "rb").read()
flen = [fls[2 * i] | fls[2 * i + 1] << 8 for i in range(len(fls) // 2)]
pos = fi = 0
calc = []
for n in lens:
    end, c = pos + n, 0
    while pos < end:
        pos += flen[fi]; fi += 1; c += 1
    calc.append(c)
print("逐格長度回算的每塊格數 %s 表上的值；用掉 %d 格（檔案共 %d 格）"
      % ("＝" if calc == frames else "≠", fi, len(flen)))
if calc != frames:
    print("  回算:", calc)
    print("  表  :", frames)

if "--emit" in sys.argv:
    import json
    off = 0
    chunks = []
    for i, (n, f) in enumerate(zip(lens, frames)):
        chunks.append({"index": i, "offset": off, "length": n, "frames": f})
        off += n
    doc = {
        "檔案": "IBM16.350",
        "說明": "OPEN.EXE 開頭動畫的分塊載入表；長度表在載入器資料段的 +0x03C0，"
                "每塊格數表在 +0x0000，兩者都是 30 個 uint16。",
        "分塊數": len(chunks),
        "總長度": sum(lens),
        "總格數": sum(frames) + 1,
        "緩衝區": {"A": "%04X:%04X" % (a_seg, a_off),
                   "B": "%04X:%04X" % (b_seg, b_off),
                   "相差位元組": b_lin - a_lin},
        "分塊": chunks,
    }
    out = "/work/docs/internals/data/intro-chunks.json"
    with open(out, "w", encoding="utf-8") as fh:
        json.dump(doc, fh, ensure_ascii=False, indent=2)
        fh.write("\n")
    print("寫出", out)
sys.exit(0)
