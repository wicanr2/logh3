#!/usr/bin/env bash
# 描述檔驅動的差異測試：產生原版向量、用 C 參考實作比對、跑突變對照。
#
#   decomp/tools/diff_run.sh <函式id> <新批次名>
# 讀 decomp/specs/<id>.json。輸出 workplace/decomp/diff/<id>/<批次名>/（vectors.txt、base.bin）。
# 收據綁執行器版本：dosgolem HEAD 不符或工作樹不乾淨就不跑。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
HEAD_EXPECT=f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c
id="${1:?用法: diff_run.sh <函式id> <新批次名>}"; name="${2:?新批次名}"
[[ "$id" =~ ^[0-9A-F]{4}_[0-9A-F]{4}$ ]] || { echo "函式 id 格式是 SEG_OFF（大寫hex）" >&2; exit 1; }
[[ "$name" =~ ^[a-z0-9][a-z0-9._-]*$ ]] || { echo "批次名只能用小寫英數與 . _ -" >&2; exit 1; }
spec="$ROOT/decomp/specs/$id.json"
test -f "$spec"
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/decomp/ledger/insns.tsv"
head="$(git -C "$GOLEM" rev-parse HEAD)"
[[ "$head" == "$HEAD_EXPECT" ]] || { echo "dosgolem HEAD 不符：$head" >&2; exit 1; }
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || { echo "dosgolem 工作樹不乾淨" >&2; exit 1; }

out="$ROOT/workplace/decomp/diff/$id/$name"
test ! -e "$out" || { echo "批次已存在：$out" >&2; exit 1; }
mkdir -p "$out/src"
cp "$ROOT/decomp/harness/diffprobe.go" "$out/src/main.go"
cat > "$out/run_in_container.sh" <<'INNER'
#!/usr/bin/env bash
set -euo pipefail
test -f /src/go.mod
test -f /orig/GIN3.COM
work="$(mktemp -d /tmp/diffprobe.XXXXXX)"
trap 'rm -rf "$work"' EXIT
cp /out/src/main.go "$work/main.go"
cat > "$work/go.mod" <<'MOD'
module github.com/wicanr2/dosgolem/diffprobe

go 1.24

require github.com/wicanr2/dosgolem v0.0.0
replace github.com/wicanr2/dosgolem => /src
MOD
cd "$work"
go run . -spec "/out/specs/$1.json" -out /out/vec -insns /out/ledger/insns.tsv -dosgolem-head "$2"
INNER
chmod +x "$out/run_in_container.sh"

export DOSGOLEM_ORIG="$ROOT/org_game/extracted"
export DOSGOLEM_EXTRA_MOUNT="$out:/out"
export DOSGOLEM_GO_CMD=bash DOSGOLEM_TIMEOUT=20m
# 描述檔與台帳以唯讀複本放進輸出目錄，容器只看得到 /out（單一 -v）
mkdir -p "$out/specs" "$out/ledger"
cp "$spec" "$out/specs/$id.json"
cp "$ROOT/decomp/ledger/insns.tsv" "$out/ledger/insns.tsv"
( cd "$GOLEM" && bash tools/go.sh /out/run_in_container.sh "$id" "$head" )

echo "--- C 參考實作比對"
bash "$ROOT/decomp/tools/check_diff.sh" "$out/vec/vectors.txt"
if [[ -f "$ROOT/decomp/specs/$id.mutants.json" ]]; then
  echo "--- 突變對照"
  mbuild="$(mktemp -d "$ROOT/workplace/decomp-build.XXXXXX")"
  trap 'rm -rf "$mbuild"' EXIT
  docker run --rm --network none --memory 1g --cpus 2 --pids-limit 64 \
    --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
    -v "$ROOT/decomp:/d:ro" -v "$out/vec:/v:ro" -v "$mbuild:/build" \
    golang:1.24-bookworm python3 -I /d/tools/mutants.py "$id" /v vectors.txt
fi
sha256sum "$out/vec/vectors.txt"
