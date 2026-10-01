#!/usr/bin/env bash
# #69 隔離的 Windows 音樂工具來源建置；文件入口：docs/re/362-windows-music-tools.md。
# --build 只補既有 MinGW image 缺件；其他命令預設禁止網路。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE=logh3-audio-cross:20261001-r1
PARENT=fd2-build-mingw:latest
EXPECTED=sha256:4ed8c0dc9513b4cf2127ae988aee7a9e238d41ec5ef4f5e0a5c6cb146b09d356
if [[ "${1:-}" == --build ]]; then
    [[ "$(docker image inspect "$PARENT" --format '{{.Id}}')" == "$EXPECTED" ]] || {
        echo 'MinGW 父工具鏈身份改變，先重新核對。' >&2; exit 1;
    }
    exec timeout 20m docker build --network default --memory 4g \
        --cpu-period 100000 --cpu-quota 200000 \
        -f "$ROOT/tools/Dockerfile.music-cross" -t "$IMAGE" "$ROOT/tools"
fi
test -d "$ROOT/workplace/music69-linux/source-cache"
test -f "$ROOT/docs/internals/data/linux-music-source-lock.json"
test -d "$ROOT/tools"
test -d "$ROOT/workplace"
mkdir -p "$ROOT/workplace/music69-windows"
[[ "$(stat -c '%u:%g' "$ROOT/workplace/music69-windows")" == "$(id -u):$(id -g)" ]] || {
    echo 'Windows 工作目錄擁有權不符，先修正確認的目錄。' >&2; exit 1;
}
if [[ "${1:-}" == --probe ]]; then
    SOURCE=/home/anr2/cht/dosgolem-logh3
    test -d "$SOURCE"
    test -d "$ROOT/workplace/gocache"
    test -d "$ROOT/workplace/gomodcache"
    exec timeout 10m docker run --rm --network none --memory 4g --cpus 2 --pids-limit 128 \
        --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
        -v "$SOURCE:/src:ro" \
        -v "$ROOT/workplace/gocache:/gocache" -v "$ROOT/workplace/gomodcache:/gomodcache" \
        -v "$ROOT/workplace/music69-windows:/out" \
        -e HOME=/tmp -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache \
        -e GOOS=windows -e GOARCH=amd64 -e CGO_ENABLED=0 -e GOPROXY=off \
        -w /src golang:1.24-bookworm go build -o /out/dosgolem-probe.exe ./cmd/probe
fi
exec timeout "${LOGH3_TIMEOUT:-25m}" docker run --rm --network none \
    --memory 4g --cpus 2 --pids-limit 256 \
    --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" \
    -v "$ROOT/workplace/music69-linux/source-cache:/source-cache:ro" \
    -v "$ROOT/docs/internals/data:/metadata:ro" \
    -v "$ROOT/tools:/tools:ro" \
    -v "$ROOT/workplace/music69-windows:/out" \
    -e HOME=/tmp -w /out "$IMAGE" "$@"
