"""第74輪候選與原池命令驗證索引。既有原版來源唯讀，不改歷史收據。"""
from pathlib import Path
import hashlib
import json
import re

ROOT = Path("/src")
DEST = ROOT / "docs/internals/data/fleet-command-data.json"


def fingerprint(path):
    p = ROOT / path
    b = p.read_bytes()
    st = p.stat()
    if (st.st_uid, st.st_gid) != (1000, 1000):
        raise RuntimeError(f"來源擁有權不符：{path}")
    return {"sha256": hashlib.sha256(b).hexdigest(), "bytes": len(b)}


def checked(path, digest):
    actual = fingerprint(path)
    if actual["sha256"] != digest:
        raise RuntimeError(f"來源SHA不符：{path}")
    return json.loads((ROOT / path).read_text())


candidate_path = "docs/internals/data/weapon-candidates.json"
command_path = "docs/internals/data/fleet-combat-command.json"
candidate = checked(candidate_path, "59a5ad2bd68a54c888179be3d85a6b99c15f16a1892b7d2b05c4de0667ad601f")
command = checked(command_path, "e3b39582bfda07105d27837747d46bca19eaf2f1fb5d6ff383500870e4a6870d")
head = "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c"
if candidate["dosgolem_head"] != head or command["dosgolem_head"] != head:
    raise RuntimeError("原版執行器版本不符")

sources = {p: fingerprint(p) for p in [candidate_path, command_path]}
cases = []
for pair in candidate["pairs"]:
    if pair["stage"] != "entry":
        continue
    if not all(c["equal"] for c in pair["comparisons"]):
        raise RuntimeError("原版雙跑不相同")
    files = []
    for side in "ab":
        for name in ["input-mem.bin", "report.txt"]:
            path = f"workplace/oracle/weapon79-r65-{pair['case']}-{side}/{name}"
            source = fingerprint(path)
            if source["sha256"] != candidate["sources"][path]["sha256"]:
                raise RuntimeError(f"原版來源改變：{path}")
            sources[path] = source
            files.append(path)
    cases.append({"case": pair["case"], "sources": files})
if len(cases) != 48:
    raise RuntimeError("原版候選案例數不符")

target_log = "workplace/round74-command-final.log"
standard_log = "workplace/round74-standard.log"
log = (ROOT / target_log).read_text()
standard = (ROOT / standard_log).read_text()
if "--- FAIL" in log or "--- SKIP" in log or "FAIL" in standard:
    raise RuntimeError("測試包含失敗或限定驗收跳過")
required = ["TestFleetCombatCandidatesNativeQueries", "TestFleetCombatCandidatesInvalidInput",
            "TestFleetCombatCommandDataAndSave", "TestFleetCombatCommandRejectedWithoutMutation",
            "TestNormalSpottingInformationAndSave"]
for name in required:
    if f"--- PASS: {name} " not in log:
        raise RuntimeError(f"缺限定測試通過：{name}")
queries = re.findall(r"query\d+ pool\d+ count\d+ writes(\d+)", log)
if len(queries) != 111 or "48案、111次逐查詢原版AX／mask符合" not in log:
    raise RuntimeError("原版逐查詢數不符")
if "48個一般分隊候選查詢不改寫原池、畫面及RND" not in log:
    raise RuntimeError("缺正常bootstrap資料域驗收")
passes = len(re.findall(r"^ok\s+github.com/wicanr2/logh3/", standard, re.M))
empty = len(re.findall(r"^\?\s+github.com/wicanr2/logh3/.*\[no test files\]", standard, re.M))
if (passes, empty) != (13, 5):
    raise RuntimeError("標準18套件結果不符")

for path in [target_log, standard_log, "internal/rules/fleetcandidates.go",
             "internal/rules/fleetcandidates_test.go", "internal/game/fleetcombatcommand.go",
             "internal/game/fleetcombatcommand_test.go", "internal/game/normalspotting_test.go"]:
    sources[path] = fingerprint(path)

result = {
    "schema": "logh3.fleet-command-data.v1",
    "date": "2026-10-04",
    "remake_parent": "dfccda34441a6271b18fd3594bfa340f177fa6ea",
    "dosgolem_head": head,
    "images": candidate["images"],
    "sources": sources,
    "cases": cases,
    "comparison": {"native_cases": 48, "native_queries": 111,
                   "native_mask_writes": sum(map(int, queries)),
                   "game_submission_and_save_cases": 5,
                   "game_rejected_cases": 10,
                   "normal_bootstrap_read_only_units": 48,
                   "standard_packages": 18, "standard_pass": 13, "standard_no_tests": 5},
    "limits": ["Game資料介面已實作，玩家武器UI及第④步尚未呼叫",
               "正常首戰只驗資料域／唯讀性，不稱完整交戰same-state",
               "沿原版候選收據雙跑與版本，未修改dosgolem或歷史索引",
               "100限定資料鏈，不升格092／056，不啟用單人CPU攻擊"],
}
out = (json.dumps(result, ensure_ascii=False, indent=2) + "\n").encode()
if DEST.exists() and DEST.read_bytes() != out:
    raise RuntimeError("拒絕覆寫不同的第74輪凍結索引")
DEST.write_bytes(out)
print(json.dumps({"sha256": hashlib.sha256(out).hexdigest(), "sources": len(sources),
                  "comparison": result["comparison"]}, ensure_ascii=False))
