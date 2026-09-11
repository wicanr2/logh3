#!/usr/bin/env bash
# Go 走 docker，不裝到主機。
#
#   tools/go.sh build ./...
#   tools/go.sh run ./cmd/logh3 -out workplace/remake
#   tools/go.sh gofmt -l internal/rules/fleetmove.go   # 唯讀檢查格式
#
# ⚠ **檢查格式用 `gofmt -l`，不要用 `go fmt ./...`。** 後者會直接改檔，
# 而這個 image 的 gofmt 對齊規則與部分檔案當初寫入時不同——跑一次整個
# repo 會順手改掉十幾個與任務無關的檔。之前沒有唯讀的入口，於是兩個
# subagent 都手打了 `docker run … gofmt -l`（違反「一律走包裝器」）；
# 缺口在工具，所以補在這裡。
#
# 原版素材（字型等）唯讀掛在 /orig。remake **不內嵌原版素材**：
# 需要字型的路徑一律由旗標傳入，缺檔就明講缺檔，不用代用品。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${LOGH3_GO_IMAGE:-golang:1.24-bookworm}"
mkdir -p "$ROOT/workplace/gocache" "$ROOT/workplace/gomodcache"
MOUNTS=()
[[ -d "$ROOT/org_game/extracted" ]] && MOUNTS+=(-v "$ROOT/org_game/extracted:/orig:ro")
# LOGH3_* 一律傳進容器，測試才能靠環境變數開診斷輸出。
ENVS=()
for v in ${!LOGH3_@}; do ENVS+=(-e "$v"); done
BIN=go
if [[ "${1:-}" == "gofmt" ]]; then BIN=gofmt; shift; fi
exec timeout "${LOGH3_TIMEOUT:-15m}" docker run --rm --network none \
  --memory "${LOGH3_MEM:-4g}" --cpus "${LOGH3_CPUS:-4}" --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -v "$ROOT:/src" \
  -v "$ROOT/workplace/gocache:/gocache" -v "$ROOT/workplace/gomodcache:/gomodcache" \
  -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -e HOME=/tmp -e GOFLAGS=-mod=mod \
  "${ENVS[@]}" "${MOUNTS[@]}" -w /src "$IMAGE" "$BIN" "$@"
