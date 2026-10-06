#!/usr/bin/env bash
# #94：由已凍結的vendor來源建立兩架構與通用Mach-O。
set -euo pipefail
version="${1:?需版本}"
source_relative="${2:?需來源相對目錄}"
[[ "$version" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
[[ "$source_relative" =~ ^workplace/package92-v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}(-[a-z0-9]+)?/source$ ]]
source_dir="/src/$source_relative"
test -f "$source_dir/go.mod"
test -d "$source_dir/vendor"
test "$(stat -c '%u:%g' /out)" = "$(id -u):$(id -g)"
test ! -e /out/gin3-universal
go version > /out/TOOLCHAIN.txt
o64-clang --version >> /out/TOOLCHAIN.txt
oa64-clang --version >> /out/TOOLCHAIN.txt
cd "$source_dir"
for arch in amd64 arm64; do
    compiler=o64-clang
    if [[ "$arch" == arm64 ]]; then compiler=oa64-clang; fi
    test ! -e "/out/gin3-$arch"
    GOOS=darwin GOARCH="$arch" CGO_ENABLED=1 CC="$compiler" \
      CGO_CFLAGS='-mmacosx-version-min=11.0' CGO_LDFLAGS='-mmacosx-version-min=11.0' \
      go build -mod=vendor -trimpath -buildvcs=false \
      -ldflags="-X main.buildVersion=$version" -o "/out/gin3-$arch" ./cmd/gin3
    go version -m "/out/gin3-$arch" > "/out/BUILDINFO-$arch.txt"
    x86_64-apple-darwin24.5-otool -L "/out/gin3-$arch" > "/out/DEPENDENCIES-$arch.txt"
    x86_64-apple-darwin24.5-otool -l "/out/gin3-$arch" > "/out/LOADCOMMANDS-$arch.txt"
done
lipo -create /out/gin3-amd64 /out/gin3-arm64 -output /out/gin3-universal
lipo -info /out/gin3-universal > /out/ARCHITECTURES.txt
chmod 755 /out/gin3-universal
printf '完成macOS通用程式：%s\n' "$version"
