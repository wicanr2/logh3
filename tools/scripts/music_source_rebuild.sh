#!/usr/bin/env bash
# 規格 077：離線來源重建；入口 bash tools/gui.sh xrun bash tools/scripts/music_source_rebuild.sh。
set -euo pipefail
ROOT=/src/workplace/music69-delivery
SOURCE="$ROOT/source-kit-vcs-root/extracted"
OUT="$ROOT/rebuild-vcs-root"
REFERENCE="$OUT/reference-tree"
test -f "$SOURCE/MANIFEST.json"
test -f "$SOURCE/logh3/LICENSE"
test ! -e "$OUT"
test ! -L "$OUT"
test "$(stat -c '%u:%g' "$ROOT")" = "$(id -u):$(id -g)"
mkdir "$OUT"
TASK_CACHE=$(mktemp -d /tmp/logh3-source-modules.XXXXXX)
trap 'rm -rf "$TASK_CACHE"' EXIT
export GOMODCACHE="$TASK_CACHE" GOPROXY=off GOSUMDB=off GOWORK=off CGO_ENABLED=0 GOARCH=amd64
go version > "$OUT/go-version.txt"
test -z "$(find "$TASK_CACHE" -mindepth 1 -print -quit)"
# 從目前程式另建對照樹，重新離線 vendor；編譯兩側均用空快取。
# 不能以一般模組模式的 h1 建置註記與 vendor 比整份二進位。
mkdir "$REFERENCE"
cp -a /src/cmd /src/internal /src/tools /src/go.mod /src/go.sum "$REFERENCE/"
(cd /src && GOMODCACHE=/gomodcache go mod vendor -o "$REFERENCE/vendor")
diff -qr "$REFERENCE/vendor" "$SOURCE/logh3/vendor" > "$OUT/vendor-diff.txt"
build_pair() {
    local NAME="$1" PLATFORM="$2" TARGET="$3"
    (cd "$REFERENCE" && GOMODCACHE="$TASK_CACHE" GOOS="$PLATFORM" go build -mod=vendor -trimpath -buildvcs=false -o "$OUT/$NAME-reference" "$TARGET")
    (cd "$SOURCE/logh3" && GOMODCACHE="$TASK_CACHE" GOOS="$PLATFORM" go build -mod=vendor -trimpath -buildvcs=false -o "$OUT/$NAME-source" "$TARGET")
    cmp "$OUT/$NAME-reference" "$OUT/$NAME-source"
    sha256sum "$OUT/$NAME-reference" "$OUT/$NAME-source"
}
build_pair music-import-linux linux ./cmd/music-import
build_pair music-import-windows windows ./cmd/music-import
build_pair music-tool-linux linux ./cmd/music-tool
(cd "$SOURCE/dosgolem" &&
    export GIT_DIR="$SOURCE/dosgolem/.git" GIT_WORK_TREE="$SOURCE/dosgolem" &&
    test "$(git rev-parse --show-toplevel)" = "$PWD" &&
    git fsck --full && git reset --mixed HEAD &&
    test "$(git rev-list --count HEAD)" = 1 &&
    test "$(git rev-parse HEAD)" = f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c &&
    test -z "$(git status --porcelain)" &&
    git rev-parse --show-toplevel HEAD && git status --porcelain) > "$OUT/dosgolem-git-proof.txt"
(cd "$SOURCE/dosgolem" && GOOS=windows go build -mod=readonly -trimpath -buildvcs=true -o "$OUT/dosgolem-probe-windows-source" ./cmd/probe)
(cd "$SOURCE/dosgolem" && GOOS=linux go build -mod=readonly -trimpath -buildvcs=true -o "$OUT/dosgolem-probe-linux-source" ./cmd/probe)
go version -m "$OUT/dosgolem-probe-windows-source" > "$OUT/dosgolem-windows-buildinfo.txt"
go version -m "$OUT/dosgolem-probe-linux-source" > "$OUT/dosgolem-linux-buildinfo.txt"
mkdir -p "$SOURCE/logh3/workplace/music69-fm/tools"
cp "$OUT/dosgolem-probe-linux-source" "$SOURCE/logh3/workplace/music69-fm/tools/dosgolem-probe"
chmod 0755 "$SOURCE/logh3/workplace/music69-fm/tools/dosgolem-probe"
(cd "$SOURCE/logh3" && GOOS=linux go test -mod=vendor -count=1 -json ./cmd/music-import ./cmd/music-tool) > "$OUT/cli-tests.jsonl"
test -z "$(find "$TASK_CACHE" -mindepth 1 -print -quit)"
printf '%s\n' '來源側空模組快取及無網路條件下的三組對照建置、dosgolem 真實 Git 身份建置與 CLI 測試通過。'
