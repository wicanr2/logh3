#!/usr/bin/env bash
# 第69輪完整艦隊交戰scene，專用dosgolem，原版只讀，不覆寫收據。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
kind="${1:?laser|missile|same-target|death}"; side="${2:?a|b}"
[[ "$kind" == laser || "$kind" == missile || "$kind" == same-target || "$kind" == death ]] || exit 1
[[ "$side" == a || "$side" == b ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -d "$GOLEM/workplace"
test -f "$ROOT/tools/scripts/fleet_scene_native.go"
test ! -e "$ROOT/workplace/oracle/scene48-r69-v2-$kind-$side"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
probe="workplace/scene48-r69-$kind-$side.go"
DOSGOLEM_ORIG="$ROOT/org_game/extracted" DOSGOLEM_EXTRA_MOUNT="$ROOT:/out:ro" DOSGOLEM_GO_CMD=bash \
  "$GOLEM/tools/go.sh" -c 'test ! -e "$1"; cp /out/tools/scripts/fleet_scene_native.go "$1"' _ "$probe"
trap 'DOSGOLEM_GO_CMD=bash "$GOLEM/tools/go.sh" -c '\''rm -f "$1"'\'' _ "$probe"' EXIT
DOSGOLEM_ORIG="$ROOT/org_game/extracted" DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out" \
  "$GOLEM/tools/go.sh" run "./$probe" "$kind" "$side"
