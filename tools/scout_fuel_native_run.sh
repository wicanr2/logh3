#!/usr/bin/env bash
# 在專用dosgolem的tools/go.sh容器內執行；固定模組唯讀，不改工作樹。
set -euo pipefail
test -f /task/workplace/scout79-r83-native-v7.go
test -f /src/go.mod
test -f /orig/GIN3.COM
task_dir="$(mktemp -d /tmp/scout79-r83.XXXXXX)"
trap 'rm -rf "$task_dir"' EXIT
cp /task/workplace/scout79-r83-native-v7.go "$task_dir/main.go"
cat > "$task_dir/go.mod" <<'EOF'
module github.com/wicanr2/dosgolem/scout79r83

go 1.24

require github.com/wicanr2/dosgolem v0.0.0
replace github.com/wicanr2/dosgolem => /src
EOF
cd "$task_dir"
go run . -output "${1:?新的JSON輸出路徑}"
