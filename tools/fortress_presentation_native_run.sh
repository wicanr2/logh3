#!/usr/bin/env bash
set -euo pipefail
task_dir="$(mktemp -d /tmp/fortress48-r134.XXXXXX)"
trap 'rm -rf "$task_dir"' EXIT
cp /task/tools/scripts/fortress_presentation_native.go "$task_dir/main.go"
cat > "$task_dir/go.mod" <<'EOF'
module github.com/wicanr2/dosgolem/fortress48r133

go 1.24

require github.com/wicanr2/dosgolem v0.0.0
replace github.com/wicanr2/dosgolem => /src
EOF
cd "$task_dir"
go run . "${1:?kind}" "${2:?side}"
