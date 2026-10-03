"""凍結兩次原版量表診斷的像素與來源；不從 remake 算期望值。"""
import hashlib
import json
from pathlib import Path


def digest(data):
    return hashlib.sha256(data).hexdigest()


root = Path("/work")
a = root / "workplace/oracle/gauges48-r68-a"
b = root / "workplace/oracle/gauges48-r68-b"
reports = [json.loads((p / "report.json").read_text()) for p in (a, b)]
normalized = [{k: v for k, v in d.items() if k != "command"} for d in reports]
assert normalized[0] == normalized[1]
sources = []
for directory in (a, b):
    for p in sorted(directory.iterdir()):
        data = p.read_bytes()
        if p.name != "report.json":
            assert data == (b / p.name).read_bytes() == (a / p.name).read_bytes()
        sources.append({"path": str(p.relative_to(root)), "sha256": digest(data), "size": len(data)})
cases = []
for c in reports[0]["cases"]:
    raw = (a / (c["case"] + "-screen.bin")).read_bytes()
    assert len(raw) == 640 * 480
    side = c["fleet_faction"] if c["kind"] == "fleet" else 1 - c["fleet_faction"]
    x = (298, 333)[side]
    crop = b"".join(raw[y * 640 + x:y * 640 + x + 11] for y in range(63, 129))
    cases.append({**c, "rect": [x, 63, 11, 66], "native_pixels_hex": crop.hex(),
                  "native_pixels_sha256": digest(crop)})
ida = []
for name in ("321F", "20B4-scenes"):
    p = root / f"workplace/ida/gauge48-r68-{name}.json"
    data = p.read_bytes()
    sources.append({"path": str(p.relative_to(root)), "sha256": digest(data), "size": len(data)})
    ida.append(json.loads(data))
result = {"schema": "logh3.engagement-gauges.v1", "grade": "proven",
          "scope": "原版量表局部常式；不代表正常玩家路徑或完整演出同狀態",
          "metadata": {k: v for k, v in normalized[0].items() if k != "cases"},
          "base_palette": json.loads((a / "base-palette.json").read_text()),
          "cases": cases, "ida": ida, "sources": sources}
data = (json.dumps(result, ensure_ascii=False, indent=2) + "\n").encode()
out = root / "docs/internals/data/engagement-gauges.json"
if out.exists():
    assert out.read_bytes() == data, "拒絕覆寫不相同的凍結證據"
else:
    assert out.parent.stat().st_uid == 1000
    out.write_bytes(data)
print("cases", len(cases), "sources", len(sources), "sha256", digest(data))
