#!/usr/bin/env bash
# 固定原版場景的等待輸入診斷；每次獨立輸出。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
side="${1:?a|b}"
[[ "$side" == a || "$side" == b ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -d "$GOLEM/workplace"
test -f "$ROOT/tools/scripts/fleet_wait_native.go"
test ! -e "$ROOT/workplace/oracle/wait48-r72-v2-$side"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
probe="workplace/wait48-r72-v2-$side.go"
DOSGOLEM_ORIG="$ROOT/org_game/extracted" DOSGOLEM_EXTRA_MOUNT="$ROOT:/out:ro" DOSGOLEM_GO_CMD=bash \
  "$GOLEM/tools/go.sh" -c 'test ! -e "$1"; cp /out/tools/scripts/fleet_wait_native.go "$1"' _ "$probe"
trap 'DOSGOLEM_GO_CMD=bash "$GOLEM/tools/go.sh" -c '\''rm -f "$1"'\'' _ "$probe"' EXIT
DOSGOLEM_ORIG="$ROOT/org_game/extracted" DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out" \
  "$GOLEM/tools/go.sh" run "./$probe" "$side"
