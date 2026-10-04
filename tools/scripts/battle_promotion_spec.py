#!/usr/bin/env python3
"""由凍結原版證據生成108正式資料，不從UI舊表反推軍階。Docker內執行。"""
import hashlib
import json
from pathlib import Path

root = Path(__file__).resolve().parents[2]
inputs = {
    "battle-promotion.json": "ab96e1d8d0213da5dfc8b34acc1f46bf1b83440eb3d552ed7f37efd8c2b0c661",
    "battle-promotion-review.json": "3d52c61e0e93433297f6240078614c11da38cf27e24347a4cfe926cfc4f451b6",
}
data = {}
for name, digest in inputs.items():
    raw = (root / "docs/internals/data" / name).read_bytes()
    assert hashlib.sha256(raw).hexdigest() == digest, name
    data[name] = json.loads(raw)
p = data["battle-promotion.json"]
r = data["battle-promotion-review.json"]
out = {
    "schema": "logh3.admiral-promotion-spec.v1",
    "sources": inputs,
    "need": p["need"],
    "rank_thresholds": p["rank_thresholds"],
    "ranks": [[{"hex": n["big5_hex"], "text": n["text"]} for n in f["names"]]
              for f in r["rank_names"]],
    "formats": {k: v["hex"] for k, v in p["formats"].items()},
}
(root / "docs/internals/data/admiral-promotion-spec.json").write_text(
    json.dumps(out, ensure_ascii=False, indent=2) + "\n")
