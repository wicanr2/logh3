#!/usr/bin/env bash
set -euo pipefail
test -f /task/workplace/promotion79-r86-global.go
test -f /src/go.mod
test -f /orig/GIN3.COM
task_dir="$(mktemp -d /tmp/promotion79-r86.XXXXXX)"
trap 'rm -rf "$task_dir"' EXIT
cp /task/workplace/promotion79-r86-global.go "$task_dir/main.go"
cat > "$task_dir/go.mod" <<'EOF'
module github.com/wicanr2/dosgolem/promotion79r86

go 1.24

require github.com/wicanr2/dosgolem v0.0.0
replace github.com/wicanr2/dosgolem => /src
EOF
cd "$task_dir"
go run . -output "${1:?新的輸出路徑}"
