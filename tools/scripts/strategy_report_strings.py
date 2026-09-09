"""把戰略模式報告的訊息表倒成 JSON（`docs/re/247`）。

    tools/py.sh tools/scripts/strategy_report_strings.py <1MB傾印> <輸出json>

表在資料段 `ds:0490`–`ds:09F0` 這一段，`\\x00` 分隔。裡面混著格式字串
（`%s`／`%d`／`\\x1B[色,` 這種顏色控制）與整句的報告文字。

**原版自帶的錯字照原樣保留**：Big5 解不開的位元組寫進 `raw`，
`text` 只在解得開時才有。這是保存的一部分，不是待修的 bug。
"""
import json
import sys

DS = 0x54B30
LO, HI = 0x0490, 0x09F0

mem = open(sys.argv[1], "rb").read()
out = []
i = DS + LO
end = DS + HI
while i < end:
    j = mem.find(b"\x00", i)
    if j < 0 or j > end:
        break
    s = mem[i:j]
    if len(s) >= 2:
        e = {"ds": "%04X" % (i - DS), "raw": s.hex()}
        try:
            e["text"] = s.decode("big5")
        except UnicodeDecodeError:
            e["text"] = None
            e["note"] = "Big5 解不開——原版資料如此"
        out.append(e)
    i = j + 1

json.dump({
    "note": ("戰略模式報告（`戰略模式處理中` 那一頁）的訊息表，資料段 "
             "`ds:0490`–`ds:09F0`。倒法與逐條說明見 docs/re/247。"
             "原版自帶的錯字照原樣保留。"),
    "source": sys.argv[1].rsplit("/", 1)[-1],
    "strings": out,
}, open(sys.argv[2], "w", encoding="utf-8"), ensure_ascii=False, indent=1)
open(sys.argv[2], "a", encoding="utf-8").write("\n")
print("寫出 %s：%d 條" % (sys.argv[2], len(out)))
