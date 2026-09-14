"""從費沙事件權威 JSON 產生可直接繪製的雙頁報告資料。"""
import json
import sys

src = json.load(open(sys.argv[1], encoding="utf-8"))

reports = {}
key_by_ds = {"3260": "owner_0", "3268": "owner_1",
             "3278": "opponent_0", "3270": "opponent_1"}
for item in src["messages"]:
    if item["ds"] not in key_by_ds:
        continue
    reports[key_by_ds[item["ds"]]] = {
        "ds": item["ds"],
        "pages": [{"raw": page.encode("big5").hex(), "text": page}
                  for page in item["pages"]],
    }

out = {
    "note": "費沙事件雙頁報告（1213:01B8；docs/spec/049、re/308）。",
    "cell_x": 7, "cell_y": 6, "cols": 24, "rows": 5,
    "style": 1, "colour": 0x97,
    "face": {"x": 432, "y": 180, "w": 80, "h": 100,
             "index_from": "WORK +0x37 military staff for viewer faction"},
    "reports": reports,
}
json.dump(out, open(sys.argv[2], "w", encoding="utf-8"),
          ensure_ascii=False, indent=1)
open(sys.argv[2], "a", encoding="utf-8").write("\n")
print("寫出 %s" % sys.argv[2])
