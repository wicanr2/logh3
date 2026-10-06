#!/usr/bin/env bash
set -euo pipefail
test -f /tools/scripts/battle_normal_turn_native.go
test -f /parent/machine.state
test -f /golem/go.mod
test -f /orig/GIN3.COM
test -d /out
task_dir="$(mktemp -d /tmp/normal79-r117.XXXXXX)"
trap 'rm -rf "$task_dir"' EXIT
cp /tools/scripts/battle_normal_turn_native.go "$task_dir/main.go"
cat > "$task_dir/go.mod" <<'EOF'
module github.com/wicanr2/dosgolem/normal79r117

go 1.24

require github.com/wicanr2/dosgolem v0.0.0
replace github.com/wicanr2/dosgolem => /golem
EOF
cd "$task_dir"
go run main.go
