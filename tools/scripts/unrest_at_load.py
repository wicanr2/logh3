"""比對存檔裡的未安定度與載入後記憶體裡的值。

    unrest_at_load.py <輸出.json> <槽位>:<STELLAR.DT?>:<載入後傾印>:<劇本名>:<回合> ...

載入後的傾印是 `-dump-mem lin:5F2CC:4032`（星域表 32 × 126）。
未安定度是記錄的最後一個 byte（`+0x7D`）——存檔裡有一份，載入之後
變成另一份，中間那一步就是還沒解出來的算式（`docs/re/41`）。
"""
import json
import sys

REC = 126


def big5(b):
    return b.split(b"\x00")[0].decode("big5", "replace")


slots = []
for arg in sys.argv[2:]:
    slot, stellar, dump, scenario, turn, receipt = arg.split(":")
    f = open(stellar, "rb").read()
    m = open(dump, "rb").read()
    regions = []
    for i in range(32):
        a, b = f[i * REC:(i + 1) * REC], m[i * REC:(i + 1) * REC]
        regions.append({
            "index": i, "name": big5(a[:12]), "owner": b[0x12],
            "planets": list(b[0x16:0x19]),
            "economy": int.from_bytes(b[0x1A:0x1E], "little"),
            "defence": int.from_bytes(b[0x20:0x22], "little"),
            "growth_per_mille": b[0x1E], "tax_rate_per_mille": b[0x26],
            "own_fleets": b[0x28], "enemy_fleets": b[0x2B],
            "stored": a[0x7D], "at_load": b[0x7D],
            # 載入後的整筆記錄。算式吃的就是這一份——新遊戲那條路徑的
            # 經濟力、成長率與稅率都是載入時重算的，檔案裡沒有。
            "raw_at_load": b.hex(),
        })
    slots.append({"slot": slot, "scenario": scenario, "turn": int(turn),
                  "receipt": receipt, "regions": regions})

out = {
    "note": "存檔裡的未安定度與載入之後記憶體裡的值。存檔的其餘欄位載入時原封不動"
            "（`STELLAR.DT?` 與執行期星域表只差 +0x7D），只有這一欄會變——"
            "載入時怎麼算出來的還沒解（docs/re/24、docs/re/34、docs/re/41）。",
    "record": {"offset": "0x7D", "width": 1, "table": "星域表 32 × 126"},
    "slots": slots,
}
with open(sys.argv[1], "w") as fh:
    json.dump(out, fh, ensure_ascii=False, indent=1)
    fh.write("\n")
print("寫出 %d 格 × %d 星域" % (len(slots), 32))
