#!/usr/bin/env bash
set -euo pipefail
task_dir="$(mktemp -d /tmp/landing48-r109-branch.XXXXXX)"
trap 'rm -rf "$task_dir"' EXIT
native=/task/tools/scripts/landing_frame_native.go
[[ "${3:-r1}" != r2 ]] || native=/task/tools/scripts/landing_frame_native_r2.go
cp "$native" "$task_dir/main.go"
cat > "$task_dir/go.mod" <<'EOF'
module github.com/wicanr2/dosgolem/landing48r109

go 1.24

require github.com/wicanr2/dosgolem v0.0.0
replace github.com/wicanr2/dosgolem => /src
EOF
cd "$task_dir"
go run . "${1:?kind}" "${2:?side}"
