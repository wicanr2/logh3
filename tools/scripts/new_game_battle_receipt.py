"""172：把新遊戲第一場的兩平台前後對照整理成機器收據。

用法：tools/py.sh tools/scripts/new_game_battle_receipt.py
sources.txt 由主機先寫：sha256sum 相關來源 > workplace/newgame172/sources.txt（容器看不到 internal/）。
"""
import hashlib
import json
from pathlib import Path

ROOT = Path("/work/workplace/newgame172")


def sha(path):
    with Path(path).open("rb") as f:
        return hashlib.file_digest(f, "sha256").hexdigest()


def run(name, binary):
    d = json.loads((ROOT / name / "RESULT.json").read_text())
    return {
        "name": name,
        "binary": binary,
        "binary_sha256": d["binary_sha256"],
        "driver_sha256": d["driver_sha256"],
        "command": d["command"],
        "state_injection": d["state_injection"],
        "seed_injection": d["seed_injection"],
        "time_injection": d["time_injection"],
        "script_flag": d["script_flag"],
        "passed": d["passed"],
        "panic": d.get("panic", False),
        "panic_head": d.get("panic_head"),
        "returncode": d.get("pre_cleanup_returncode"),
        "battle_region": d.get("battle_region"),
        "battle_after_load": d.get("battle_after_load"),
        "saved_sha256": d.get("saved_sha256"),
        "last_state": d.get("last_state"),
        "trace_sha256": d.get("trace_sha256"),
        "log_sha256": d.get("log_sha256"),
        "captures": [
            {"name": e["name"], "state": e["state"], "turn": (e.get("tactical") or {}).get("turn"), "sha256": e["sha256"]}
            for e in d["events"]
            if e["kind"] == "capture"
        ],
        "inputs": sum(1 for e in d["events"] if e["kind"] in ("click", "key")),
    }


sources = {}
for line in (ROOT / "sources.txt").read_text().splitlines():
    digest, name = line.split(None, 1)
    sources[name] = digest

receipt = {
    "schema": "logh3.new-game-first-battle.v1",
    "date": "2026-11-02",
    "base_head": (ROOT / "base-head.txt").read_text().strip(),
    "worktree_dirty": True,
    "status": "limited; 新遊戲第一場不崩潰且有配置，不宣稱與原版同態",
    "comparison_level": "nearby-state",
    "sources_sha256": sources,
    "original_evidence": {
        "fcmd-base-fleet0.bin": {
            "location": "~/cht/dosgolem-workplace-archive/logh3/fcmd/",
            "sha256": "3a077e0847f611d3db1cfd18a88c6bfb768fc4e24e7822e40e5251b6708704d7",
            "fleet0": {"region": 0, "x": 336, "y": 280, "facing": 34},
        },
        "t260_post.bin": {
            "location": "~/cht/dosgolem-workplace-archive/logh3/",
            "sha256": "4b4e6083be68a2ad4b86f9d93685642d99d7a1b33ad022f2c382ca45b03bb1da",
            "fleets": [
                {"index": 28, "region": 31, "x": 1696, "y": 1740, "facing": 162},
                {"index": 31, "region": 31, "x": 1728, "y": 1708, "facing": 160},
                {"index": 23, "region": 31, "x": 1728, "y": 1740, "facing": 161},
            ],
        },
    },
    "toolchain": {
        "gui_image": "logh3-ebiten:20261030-r169",
        "go_image": "golang:1.24-bookworm",
        "wine_image": "yuan-wine-verify:noble-r2（wine-9.0，專案預設kol-wine-verify映像本機已不存在）",
    },
    "standard_tests": (ROOT / "standard-tests.txt").read_text().splitlines(),
    "standard_tests_sha256": sha(ROOT / "standard-tests.txt"),
    "runs": [
        run("before-linux-empire", "v.1.6.4-20261031 AppImage解出的gin3"),
        run("before-linux-alliance", "v.1.6.4-20261031 AppImage解出的gin3"),
        run("after-linux-empire", "工作樹建置的gin3-fixed"),
        run("after-linux-alliance", "工作樹建置的gin3-fixed"),
        run("before-windows-empire", "v.1.6.4-20261031 Windows封包的gin3.exe，Wine"),
        run("before-windows-alliance", "v.1.6.4-20261031 Windows封包的gin3.exe，Wine"),
        run("after-windows-empire", "工作樹交叉編譯的gin3-fixed.exe，Wine"),
        run("after-windows-alliance", "工作樹交叉編譯的gin3-fixed.exe，Wine"),
    ],
    "saves_identical_across_platforms": {
        f: json.loads((ROOT / f"after-linux-{f}" / "RESULT.json").read_text())["saved_sha256"]
        == json.loads((ROOT / f"after-windows-{f}" / "RESULT.json").read_text())["saved_sha256"]
        for f in ("empire", "alliance")
    },
    "limits": [
        "Windows只在Wine驗證，沒有實機",
        "同一星域多支新編艦隊的避讓順序依表位，與原版不同，未對拍",
        "非首都新編艦隊的待命行星分支只有靜態證據",
        "使用者確認第169輪的閃退發生在進戰場，與本輪同一個崩潰",
    ],
}
Path("/work/docs/internals/data/new-game-battle172.json").write_text(json.dumps(receipt, ensure_ascii=False, indent=1) + "\n")
print("ok", len(receipt["runs"]))
