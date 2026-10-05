#!/usr/bin/env bash
set -euo pipefail
task_dir="$(mktemp -d /tmp/assault56-r100.XXXXXX)"
trap 'rm -rf "$task_dir"' EXIT
cp /task/workplace/assault56-r100-lifecycle-v6.go "$task_dir/main.go"
cat > "$task_dir/go.mod" <<'EOF'
module github.com/wicanr2/dosgolem/assault56r100

go 1.24

require github.com/wicanr2/dosgolem v0.0.0
replace github.com/wicanr2/dosgolem => /src
EOF
cd "$task_dir"
go run . -output "${1:?新輸出路徑}"
