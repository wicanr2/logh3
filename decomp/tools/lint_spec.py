"""描述檔與突變清單的快速檢查（在 Docker 內以 python3 -I 執行），避免跑完整個探針才發現格式或 find 錯誤。

    lint_spec.py <函式id> /d
檢查：
  - specs/<id>.json 與 specs/<id>.mutants.json 是合法 JSON
  - src/f_<id>.c 存在
  - 每個突變的 find 在 C 檔裡剛好出現一次，replace 與 find 不同
  - 突變至少 3 個
任何一項不成立以非 0 結束，並列出全部問題。
"""
import json
import os
import sys

fid, d = sys.argv[1], sys.argv[2]
problems = []


def load(path):
    try:
        with open(path, encoding="utf-8") as f:
            return json.load(f)
    except Exception as e:  # noqa: BLE001
        problems.append(f"{path}：不是合法 JSON（{e}）")
        return None


spec = load(f"{d}/specs/{fid}.json")
muts_path = f"{d}/specs/{fid}.mutants.json"
muts = load(muts_path) if os.path.exists(muts_path) else None
src_path = f"{d}/src/f_{fid}.c"
src = open(src_path, encoding="utf-8").read() if os.path.exists(src_path) else None
if src is None:
    problems.append(f"找不到 {src_path}")
if muts is None and os.path.exists(muts_path):
    pass
elif muts is None:
    problems.append(f"找不到 {muts_path}")
elif src is not None:
    if len(muts) < 3:
        problems.append(f"突變只有 {len(muts)} 個（至少 3 個）")
    for m in muts:
        n = src.count(m.get("find", ""))
        if n != 1:
            problems.append(f"突變「{m.get('name')}」：find 在 C 檔出現 {n} 次（必須剛好 1 次）")
        if m.get("find") == m.get("replace"):
            problems.append(f"突變「{m.get('name')}」：replace 與 find 相同")
if spec is not None and spec.get("id") != fid:
    problems.append(f"描述檔的 id 是 {spec.get('id')}，不是 {fid}")

if problems:
    print("描述檔檢查失敗：")
    for p in problems:
        print("  - " + p)
    sys.exit(1)
print(f"描述檔檢查通過：{fid}（突變 {len(muts)} 個）")
