#!/usr/bin/env python3
"""第67輪正式純核心驗證索引。由tools/py.sh執行，不重寫既有原版收據。"""
import hashlib
import json
import os
from pathlib import Path
import re
import struct
import sys


ROOT = Path("/ref")
OUTPUT = Path(sys.argv[1])
if OUTPUT.exists():
    raise SystemExit("拒絕覆寫既有交戰核心收據")
assert OUTPUT.parent.is_dir()
assert OUTPUT.parent.stat().st_uid == os.getuid()
sources = {}


def record(path, expected=None):
    data = (ROOT / path).read_bytes()
    digest = hashlib.sha256(data).hexdigest()
    if expected is not None:
        assert digest == expected, path
    sources[path] = {"sha256": digest, "bytes": len(data)}
    return data


manifests = {}
for name, digest in {
    "fleet-combat-damage.json": "360277c8dcbf6214da429baead75d4a7bf20b09b03f1ca2af3c0313df4171b30",
    "weapon-candidates.json": "59a5ad2bd68a54c888179be3d85a6b99c15f16a1892b7d2b05c4de0667ad601f",
    "combat-sequence.json": "199f328b4463d49a817e54c2ade49dd148a24d45e94e94d135488d699c3205a8",
}.items():
    manifests[name] = json.loads(record("docs/internals/data/" + name, digest))
    assert manifests[name]["dosgolem_head"] == "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c"

record("org_game/extracted/GIN3.COM", "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918")


def native(manifest, path):
    index = manifest.get("sources", manifest.get("files"))
    return record(path, index[path]["sha256"])


damage = manifests["fleet-combat-damage.json"]
sequence = manifests["combat-sequence.json"]
weapon = manifests["weapon-candidates.json"]
singles = [(damage, "damage79-r64-" + c["case"] + "-a") for c in damage["cases"]]
near = [c["case"] for c in sequence["pairs"] if c["case"].startswith("curve") and c["case"].endswith("-near")]
zero = [c["case"] for c in sequence["pairs"] if c["case"].endswith("-zero") or c["case"] == "curve0-stack-first"]
assert len(singles) == 31 and len(near) == 11 and len(zero) == 12
singles += [(sequence, "sequence79-r66-" + name + "-a") for name in near]
singles += [(weapon, "weapon79-r65-fighter-end-a")]
for manifest, name in singles:
    for filename in ("input-mem.bin", "end-mem.bin", "report.txt", "flags.txt", "meta.txt"):
        native(manifest, "workplace/oracle/" + name + "/" + filename)
    native(manifest, "workplace/" + name + "-plan.json")
for name in zero:
    native(sequence, "workplace/oracle/sequence79-r66-" + name + "-a/input-mem.bin")
for name in ("multi-ranked", "same-target-measured"):
    for filename in ("input-mem.bin", "end-mem.bin", "keys-mem.bin", "sorted-mem.bin",
                     "damage1-mem.bin", "damage2-mem.bin", "report.txt", "flags.txt", "meta.txt"):
        native(sequence, "workplace/oracle/sequence79-r66-" + name + "-a/" + filename)

ram = native(sequence, "workplace/oracle/normal79-r61-a/end-mem.bin")
assert len(ram) == 1048576
ds = 0x54b30
profiles = []
for type_id in range(12):
    power, major, minor = struct.unpack_from("<3H", ram, ds + 0x4fda + type_id * 6)
    off, seg = struct.unpack_from("<2H", ram, ds + 0x477c + type_id * 4)
    curve = [list(struct.unpack_from("<iH", ram, seg * 16 + off + i * 6)) for i in range(6)]
    profiles.append({"type": type_id, "power": power, "major": major, "minor": minor, "curve": curve})

core_log = record("workplace/round67-fleet-core-tests.log").decode()
assert not any(x in core_log for x in ("--- FAIL", "--- SKIP"))
parents = re.findall(r"^--- PASS: (TestFleetCombat\w+)", core_log, re.M)
assert len(parents) == 6
standard = record("workplace/round67-standard-tests.log").decode()
assert not re.search(r"^FAIL(?:\s|$)", standard, re.M)
passed = re.findall(r"^ok\s+(github.com/wicanr2/logh3/\S+)", standard, re.M)
empty = re.findall(r"^\?\s+(github.com/wicanr2/logh3/\S+)", standard, re.M)
assert len(passed) == 13 and len(empty) == 5
for path in ("internal/rules/fleetcombat.go", "internal/rules/fleetcombat_test.go",
             "docs/spec/093-fleet-combat-core.md", "tools/scripts/fleet_combat_core_receipt.py"):
    record(path)
images = record("workplace/round67-images.txt").decode().splitlines()
assert len(images) == 3 and all("sha256:" in image for image in images)
receipt = {
    "schema": "logh3.fleet-combat-core.v1", "date": "2026-10-04",
    "remake_parent": "627c73bae07c29b4d3d8662cf31d75147a43110e", "remake_worktree_dirty": True,
    "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": False,
    "status": "093 READY; pure core implemented and verified; 092 DRAFT",
    "images": images,
    "coverage": {"damage_cases": 31, "near_types": 11, "legal_fighter": 1,
                 "multi_sequences": 2, "approved_zero_distance_cases": 12,
                 "core_test_groups": parents, "product_passed": passed, "product_no_tests": empty},
    "weapon_profiles": profiles,
    "order_bases": list(struct.unpack_from("<5H", ram, ds + 0x4426)),
    "mobility_coefficients": list(ram[ds + 0x4e2e:ds + 0x4e2e + 11]),
    "sources": sources,
    "limits": ["純核心原版資料對拍，不替代正常Game/UI/保存或畫面same-state",
               "零距離取首點為已核准差異，不比原版殘留stack戰損",
               "合法戰鬥機收據只核呼叫與末端資料，不虛稱其中途係數已直接觀測",
               "原版序列化state／死亡／其他取消及完整演出仍由092驗收",
               "061單人CPU被動政策保持；本核心不選命令或啟用CPU攻擊"],
}
OUTPUT.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n")
assert OUTPUT.stat().st_uid == os.getuid()
print("sources", len(sources), "SHA-256", hashlib.sha256(OUTPUT.read_bytes()).hexdigest())
