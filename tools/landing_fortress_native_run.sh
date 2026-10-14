#!/usr/bin/env bash
set -euo pipefail
task_dir="$(mktemp -d /tmp/landing48-r135.XXXXXX)"
trap 'rm -rf "$task_dir"' EXIT
cp /task/tools/scripts/landing_fortress_native.go "$task_dir/main.go"
cat > "$task_dir/go.mod" <<'EOF'
module github.com/wicanr2/dosgolem/landing48r135

go 1.24

require github.com/wicanr2/dosgolem v0.0.0
replace github.com/wicanr2/dosgolem => /src
EOF
cd "$task_dir"
go run . "${2:?failure|capture}" "${3:?a|b}" "${1:?fortress|iserlohn}"
