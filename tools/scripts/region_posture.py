"""把 `2703:0933` 的同狀態收據倒成 JSON。

    tools/py.sh tools/scripts/region_posture.py <1MB傾印> <輸出json>

收據內容是 32 個星域的擁有者、在用艦隊的（所在星域、陣營、`+0x00`），
以及原版當下的 `+0x24` 與 `+0x13`。remake 的回歸測試拿它重算
（`docs/re/244`、`docs/spec/032`）。
"""
import json
import sys

mem = open(sys.argv[1], "rb").read()
DS = 0x54B30
REG, FLEET = DS + 0xA79C, DS + 0xC074

owner = [mem[REG + i * 126 + 0x12] for i in range(32)]
fleets = []
for s in range(32):
    f = mem[FLEET + s * 122:FLEET + s * 122 + 122]
    if f[0x04] == 0xFF:            # 這一筆沒在用
        continue
    fleets.append({"slot": s, "region": f[0x06], "faction": f[0x01],
                   "flags": f[0x00]})
regions = [{"owner": owner[i],
            "capital_cost": int.from_bytes(mem[REG + i * 126 + 0x24:][:2], "little"),
            "flags": mem[REG + i * 126 + 0x13]} for i in range(32)]

out = {
    "note": ("`2703:0933` 的同狀態收據：星域 `+0x24`（到自己首都的路徑成本）"
             "與 `+0x13`（bit0 遠、bit2 雙方都有艦隊、bit3 前線）。"
             "算法見 docs/re/244，邊權表在 star-graph.json 的 edge_cost。"),
    "source": sys.argv[1].rsplit("/", 1)[-1],
    "owners": owner,
    "fleets": fleets,
    "regions": regions,
}
open(sys.argv[2], "w", encoding="utf-8").write(
    json.dumps(out, ensure_ascii=False, indent=1) + "\n")
print("寫出 %s：%d 支在用艦隊" % (sys.argv[2], len(fleets)))
