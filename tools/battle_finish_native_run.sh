#!/usr/bin/env bash
set -euo pipefail
test -f /task/workplace/finish79-r89-native.go
test -f /src/go.mod
test -f /orig/GIN3.COM
task_dir="$(mktemp -d /tmp/finish79-r89.XXXXXX)"
trap 'rm -rf "$task_dir"' EXIT
cp /task/workplace/finish79-r89-native.go "$task_dir/main.go"
cat > "$task_dir/go.mod" <<'EOF'
module github.com/wicanr2/dosgolem/finish79r89

go 1.24

require github.com/wicanr2/dosgolem v0.0.0
replace github.com/wicanr2/dosgolem => /src
EOF
cd "$task_dir"
go run . -output "${1:?新輸出路徑}"
