#!/usr/bin/env bash
# 第68輪原版量表局部對照；不改dosgolem版控檔案或前輪收據。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
side="${1:?a|b}"
[[ "$side" == a || "$side" == b ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -d "$GOLEM/workplace"
test -f "$ROOT/tools/scripts/engagement_gauges_native.go"
test ! -e "$ROOT/workplace/oracle/gauges48-r68-$side"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
# 只為internal套件可見性把受版控來源複製至已gitignore的工作區；編譯及複製都在Docker。
DOSGOLEM_ORIG="$ROOT/org_game/extracted" DOSGOLEM_EXTRA_MOUNT="$ROOT:/out:ro" DOSGOLEM_GO_CMD=bash \
  "$GOLEM/tools/go.sh" -c 'test ! -e workplace/gauge48-r68-probe.go; cp /out/tools/scripts/engagement_gauges_native.go workplace/gauge48-r68-probe.go'
trap 'DOSGOLEM_GO_CMD=bash "$GOLEM/tools/go.sh" -c "rm -f workplace/gauge48-r68-probe.go"' EXIT
DOSGOLEM_ORIG="$ROOT/org_game/extracted" DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out" \
  "$GOLEM/tools/go.sh" run ./workplace/gauge48-r68-probe.go "$side"
