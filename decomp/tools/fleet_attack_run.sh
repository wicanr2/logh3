#!/usr/bin/env bash
# 203A:0424 戰略艦隊攻擊力：產生原版向量，再用 C 重寫版比對。
#
#   decomp/tools/fleet_attack_run.sh <新批次名> [案數]
#
# 輸出：workplace/decomp/fleet_attack/<批次名>/vectors.txt（含執行器 HEAD、保存狀態與函式雜湊）。
# 收據綁執行器版本：HEAD 不符或工作樹不乾淨就不跑。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
HEAD_EXPECT=f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c
name="${1:?用法: fleet_attack_run.sh <新批次名> [案數]}"
count="${2:-5000}"
[[ "$name" =~ ^[a-z0-9][a-z0-9._-]*$ ]] || { echo "批次名只能用小寫英數與 . _ -" >&2; exit 1; }

test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-base.state"
head="$(git -C "$GOLEM" rev-parse HEAD)"
[[ "$head" == "$HEAD_EXPECT" ]] || { echo "dosgolem HEAD 不符：$head" >&2; exit 1; }
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || { echo "dosgolem 工作樹不乾淨" >&2; exit 1; }

out="$ROOT/workplace/decomp/fleet_attack/$name"
test ! -e "$out" || { echo "批次已存在：$out" >&2; exit 1; }
mkdir -p "$out/src"
cp "$ROOT/decomp/harness/fleet_attack_probe.go" "$out/src/main.go"
cat > "$out/run_in_container.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
test -f /src/go.mod
test -f /orig/GIN3.COM
work="$(mktemp -d /tmp/fleet-attack.XXXXXX)"
trap 'rm -rf "$work"' EXIT
cp /out/src/main.go "$work/main.go"
cat > "$work/go.mod" <<'MOD'
module github.com/wicanr2/dosgolem/fleetattack

go 1.24

require github.com/wicanr2/dosgolem v0.0.0
replace github.com/wicanr2/dosgolem => /src
MOD
cd "$work"
go run . -output /out/vectors.txt -n "$1" -dosgolem-head "$2"
EOF
chmod +x "$out/run_in_container.sh"

export DOSGOLEM_ORIG="$ROOT/org_game/extracted"
export DOSGOLEM_EXTRA_MOUNT="$out:/out"
export DOSGOLEM_GO_CMD=bash DOSGOLEM_TIMEOUT=10m
( cd "$GOLEM" && bash tools/go.sh /out/run_in_container.sh "$count" "$head" )

echo "--- 重寫版比對"
bash "$ROOT/decomp/tools/check_vectors.sh" "$out/vectors.txt"
sha256sum "$out/vectors.txt"
