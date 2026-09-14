"""從戰略報告與陣營名權威 JSON 產生勝負報告資料。"""
import json
import sys

reports = json.load(open(sys.argv[1], encoding="utf-8"))
names = json.load(open(sys.argv[2], encoding="utf-8"))
raw = {item["ds"]: item["raw"] for item in reports["strings"]}

out = {
    "note": "勝負確定報告（04C8:00F3；docs/spec/050、re/309）。",
    "cell_x": 10, "cell_y": 5, "cols": 20, "rows": 10,
    "style": 1, "colour": 0x97,
    "face": {"x": 416, "y": 160, "w": 80, "h": 100},
    "button": {"cell_x": 17, "cell_y": 9, "cols": 2,
               "colour": 0x87, "codes": ["BD54", "BB7B"], "text": "確認"},
    "formats": {
        "date": raw["05DE"], "win_intro": raw["05F9"],
        "win_name": raw["0615"], "win_regions": raw["0624"],
        "loss_intro": raw["063D"], "loss_name": raw["0659"],
        "victory": raw["0669"], "score": raw["067A"],
    },
    "faction_names": [item["bytes"] for item in names["factions"][:2]],
    "plain_names": ["abd2b0eaad78", "a650b7f9ad78"],
}
json.dump(out, open(sys.argv[3], "w", encoding="utf-8"), ensure_ascii=False, indent=1)
open(sys.argv[3], "a", encoding="utf-8").write("\n")
print("寫出 %s" % sys.argv[3])
