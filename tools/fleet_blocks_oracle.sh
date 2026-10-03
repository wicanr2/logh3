#!/usr/bin/env bash
# 完整下方場景及逐layer邊界；獨立收據，拒絕覆寫。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
side="${1:?a|b|registers-a|registers-b|edges-valid-a|edges-valid-b}"
source=fleet_blocks_native.go
case "$side" in
  a|b) ;;
  registers-a|registers-b) source=fleet_blocks_registers.go ;;
  edges-valid-a|edges-valid-b) source=fleet_blocks_edges.go ;;
  *) exit 1 ;;
esac
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -d "$GOLEM/workplace"
test -f "$ROOT/tools/scripts/$source"
test ! -e "$ROOT/workplace/oracle/blocks48-r70-$side"
test ! -e "$ROOT/workplace/oracle/blocks48-r70-$side.json"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
probe="workplace/blocks48-r70-$side.go"
DOSGOLEM_ORIG="$ROOT/org_game/extracted" DOSGOLEM_EXTRA_MOUNT="$ROOT:/out:ro" DOSGOLEM_GO_CMD=bash \
  "$GOLEM/tools/go.sh" -c 'test ! -e "$1"; cp "/out/tools/scripts/$2" "$1"' _ "$probe" "$source"
trap 'DOSGOLEM_GO_CMD=bash "$GOLEM/tools/go.sh" -c '\''rm -f "$1"'\'' _ "$probe"' EXIT
DOSGOLEM_ORIG="$ROOT/org_game/extracted" DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out" \
  "$GOLEM/tools/go.sh" run "./$probe" "$side"
