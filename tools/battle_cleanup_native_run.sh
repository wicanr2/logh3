#!/usr/bin/env bash
set -euo pipefail
test -f /task/workplace/cleanup79-r88-native.go
test -f /src/go.mod
test -f /orig/GIN3.COM
task_dir="$(mktemp -d /tmp/cleanup79-r88.XXXXXX)"
trap 'rm -rf "$task_dir"' EXIT
cp /task/workplace/cleanup79-r88-native.go "$task_dir/main.go"
cat > "$task_dir/go.mod" <<'EOF'
module github.com/wicanr2/dosgolem/cleanup79r88

go 1.24

require github.com/wicanr2/dosgolem v0.0.0
replace github.com/wicanr2/dosgolem => /src
EOF
cd "$task_dir"
go run . -output "${1:?新輸出路徑}"
