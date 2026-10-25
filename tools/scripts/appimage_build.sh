#!/bin/sh
set -eu
test -x /appdir/AppRun
LOGH3_VERSION="${LOGH3_APPIMAGE_VERSION:-v.1.6.1-20261025}"
printf '%s\n' "$LOGH3_VERSION" | grep -Eq '^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$'
LOGH3_OUTPUT="/out/LOGH3-$LOGH3_VERSION-linux-x86_64.AppImage"
test ! -e "$LOGH3_OUTPUT"
printf '%s  %s\n' 1cc49bcf1e2ccd593c379adb17c9f85a36d619088296504de95b1d06215aebbf /opt/runtime-x86_64 | sha256sum -c -
test "$(/opt/runtime-x86_64 --appimage-offset)" = 944632
LOGH3_EPOCH=$(cat /out/SOURCE_DATE_EPOCH)
case "$LOGH3_EPOCH" in ''|*[!0-9]*) exit 2 ;; esac
trap 'rm -f /out/filesystem.squashfs' EXIT HUP INT TERM
mksquashfs /appdir /out/filesystem.squashfs -noappend -comp zstd -Xcompression-level 15 \
  -b 131072 -processors 2 -no-progress -no-xattrs -force-uid 1000 -force-gid 1000 \
  -all-time "$LOGH3_EPOCH" -mkfs-time "$LOGH3_EPOCH"
cat /opt/runtime-x86_64 /out/filesystem.squashfs > "$LOGH3_OUTPUT"
chmod 755 "$LOGH3_OUTPUT"
mksquashfs -version > /out/SQUASHFS-TOOLCHAIN.txt
printf '%s\n' 'READY AppImage，後續須獨立核對與正常啟動'
