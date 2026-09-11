#!/usr/bin/env bash
# 一次性搬移：把 `docs/worklist.json` 裡還沒做完的條目開成 GitHub issue。
#
#   tools/issues_migrate.sh          # 乾跑：只印要開哪幾張
#   tools/issues_migrate.sh --apply  # 真的開
#
# ⚠ **可以重跑，不會重複開。** 開之前先抓快照，本文開頭的
# `<!-- worklist:id=… -->` 已經存在的就跳過——中途斷掉（網路、速率限制）
# 之後直接再跑一次即可。這是對外動作，重複開出來的 issue 要人手一張張關。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO="${LOGH3_REPO:-wicanr2/logh3}"
APPLY=0
[[ "${1:-}" == "--apply" ]] && APPLY=1
cd "$ROOT"

payload="$ROOT/workplace/issues-export.json"
tools/go.sh run ./cmd/worklist export-issues > "$payload"

if (( APPLY )); then
  # label：只建自己用的那幾個，`--force` 讓重跑時更新顏色與說明而不是報錯。
  gh label create worklist       -R "$REPO" --force --color 5319e7 --description "docs/worklist 搬過來的條目（本文尾巴有機器核對區塊）" >/dev/null
  gh label create layer:re       -R "$REPO" --force --color 1d76db --description "逆向：機制還沒讀出來" >/dev/null
  gh label create layer:parity   -R "$REPO" --force --color d93f0b --description "對拍：要與原版逐位元組相同" >/dev/null
  gh label create layer:feature  -R "$REPO" --force --color 0e8a16 --description "功能：玩家會撞到、但 remake 還沒有" >/dev/null
  gh label create layer:decision -R "$REPO" --force --color fbca04 --description "取捨：等使用者裁決" >/dev/null
  gh label create layer:matrix   -R "$REPO" --force --color c5def5 --description "驗證矩陣的一列（完成度）" >/dev/null
  for s in $(python3 -c "import json;print(' '.join(sorted({l for x in json.load(open('$payload')) for l in x['labels'] if l.startswith('section:')})))"); do
    gh label create "$s" -R "$REPO" --force --color ededed --description "驗證矩陣章節 ${s#section:}" >/dev/null
  done
  tools/issues_snapshot.sh >/dev/null
fi

python3 - "$payload" "$ROOT/workplace/issues.json" "$APPLY" "$REPO" <<'PY'
import json, os, subprocess, sys, tempfile
payload, snap, apply, repo = sys.argv[1], sys.argv[2], sys.argv[3] == "1", sys.argv[4]
items = json.load(open(payload, encoding="utf-8"))
have = set()
if os.path.exists(snap):
    for x in json.load(open(snap, encoding="utf-8")):
        b = x.get("body") or ""
        if b.startswith("<!-- worklist:id="):
            have.add(b[len("<!-- worklist:id="):b.index(" -->")])
made = skipped = 0
for it in items:
    if it["id"] in have:
        skipped += 1
        continue
    if not apply:
        print(f"  會開  [{','.join(l for l in it['labels'] if l != 'worklist')}] {it['id']}：{it['title'][:50]}")
        made += 1
        continue
    with tempfile.NamedTemporaryFile("w", suffix=".md", delete=False, encoding="utf-8") as f:
        f.write(it["body"])
        path = f.name
    cmd = ["gh", "issue", "create", "-R", repo, "--title", it["title"], "--body-file", path]
    for l in it["labels"]:
        cmd += ["--label", l]
    url = subprocess.run(cmd, check=True, capture_output=True, text=True).stdout.strip()
    os.unlink(path)
    print(f"  開了  {it['id']} → {url}")
    made += 1
verb = "開了" if apply else "會開"
print(f"\n{verb} {made} 張、已存在跳過 {skipped} 張（共 {len(items)} 條未完成）")
PY

if (( APPLY )); then
  tools/issues_snapshot.sh
fi
