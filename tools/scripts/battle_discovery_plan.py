"""第40輪固定索敵呈現分支。計畫只讀已版控 JSON，禁止覆寫收據。"""
import hashlib
import json
from pathlib import Path
import re
import sys


def flags(name):
    match = re.fullmatch(r"battle79-discovery-r40-(detail3|camera|cpu|twohuman|swapped|left|right)-[ab]", name)
    assert match, "不接受探索或既有批次名稱"
    data = json.loads(Path("docs/internals/data/battle-discovery-plan.json").read_text())
    case = next(c for c in data["cases"] if c["name"] == match[1])
    source = Path("workplace/oracle") / data["parent"]
    assert hashlib.sha256(source.read_bytes()).hexdigest() == data["parent_sha256"]
    prefix = "/src/workplace/logh3-" + name
    result = ["-load-state", "/out/oracle/" + data["parent"],
              "-steps", str(case["end"] + 1), "-tick", "20000",
              "-regs-from", "1226630000", "-regs-max", "128", "-regs-at", data["regs_at"]]
    result += case["inputs"]
    result += ["-dump-at", ";".join(f"{p['step']}:{prefix}/{p['name']}.png" for p in case["points"]),
               "-dump-mem-at", ";".join(f"{p['step']}:lin:0:1048576:{prefix}/{p['name']}-mem.bin" for p in case["points"])]
    return result, len(case["points"]) * 3


if __name__ == "__main__":
    name = sys.argv[1]
    result, count = flags(name)
    p = Path("workplace") / (name + "-flags.txt")
    assert not p.exists()
    p.write_text("\n".join(result) + "\n")
    print(count)
