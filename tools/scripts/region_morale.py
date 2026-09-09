"""把 796/1 量到的民心（星域記錄 `+0x7C`）補進月底結算的資料檔。

    tools/py.sh tools/scripts/region_morale.py <1MB傾印> <json>

傾印是 `turn260mem` 的 `t260_pre.bin`（AI 跑之前的開局狀態）。
擁有者那一欄已經在檔案裡了，這裡只加民心，並順便核對兩者一致。
"""
import json
import sys

DS = 0x54B30
REGION = DS + 0xA79C

mem = open(sys.argv[1], "rb").read()
path = sys.argv[2]
d = json.load(open(path, encoding="utf-8"))

for r in d["regions"]:
    rec = REGION + r["index"] * 126
    owner = mem[rec + 0x12]
    if owner != r["owner"]:
        raise SystemExit("星域 %d 擁有者 %d 與檔案的 %d 不同"
                         % (r["index"], owner, r["owner"]))
    r["morale_796_01"] = mem[rec + 0x7C]

with open(path, "w", encoding="utf-8") as f:
    json.dump(d, f, ensure_ascii=False, indent=1)
    f.write("\n")
print("補了 %d 個星域的民心" % len(d["regions"]))
