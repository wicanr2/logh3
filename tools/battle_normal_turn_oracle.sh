#!/usr/bin/env bash
# 第117輪正常原版續行。沿既有Go image，單獨唯讀掛來源與父保存。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
revision="${1:?新的批次名}"
[[ "$revision" =~ ^[a-z0-9]+$ ]] || exit 1
parent="$ROOT/workplace/oracle/normal89-r55-work-a"
output="$ROOT/workplace/oracle/normal79-r117-$revision"
for path in "$GOLEM" "$ROOT/tools" "$ROOT/org_game/extracted" "$parent" \
  "$ROOT/workplace/oracle" "$ROOT/workplace/gocache" "$ROOT/workplace/gomodcache"; do
  test -d "$path"
done
test -f "$parent/machine.state"
test -f "$ROOT/tools/scripts/battle_normal_turn_native.go"
test -f "$GOLEM/go.mod"
test ! -e "$output"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
uid="$(id -u)"; gid="$(id -g)"
[[ "$(stat -c %u:%g "$ROOT/workplace/oracle")" == "$uid:$gid" ]] || exit 1
[[ "$(stat -c %u:%g "$parent/machine.state")" == "$uid:$gid" ]] || exit 1
mkdir "$output"
exec timeout 5m docker run --rm --network none --memory 4g --cpus 4 --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 --user "$uid:$gid" \
  -v "$GOLEM:/golem:ro" -v "$ROOT/tools:/tools:ro" \
  -v "$ROOT/org_game/extracted:/orig:ro" -v "$parent:/parent:ro" -v "$output:/out" \
  -v "$ROOT/workplace/gocache:/gocache" -v "$ROOT/workplace/gomodcache:/gomodcache" \
  -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -e HOME=/tmp -e GOFLAGS=-mod=mod \
  golang:1.24-bookworm bash /tools/battle_normal_turn_native_run.sh
