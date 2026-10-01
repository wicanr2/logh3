#!/bin/sh
# 規格 072：在既有 audio image 內收集真實工具／完整依賴／著作權。
# 入口：bash tools/audio.sh sh /src/tools/scripts/music_linux_collect.sh <新輸出目錄>
set -eu
out=${1:?必須提供新的 /work/ 輸出目錄}
case "$out" in /work/*) ;; *) echo '只准寫入明確研究輸出' >&2; exit 2;; esac
test ! -e "$out" || { echo '輸出已存在，拒絕覆蓋' >&2; exit 2; }
test "$(stat -c %u "$(dirname "$out")")" = "$(id -u)"
umask 022
mkdir -p "$out/tools/real" "$out/tools/lib" "$out/licenses/debian"
ldd /usr/bin/ffmpeg /usr/bin/fluidsynth /usr/local/bin/opl2wav > "$out/ldd.txt"
if grep -q 'not found' "$out/ldd.txt"; then echo 'image 缺共享依賴' >&2; exit 2; fi
awk '/=> \// {print $3} /^[[:space:]]*\// && /\(0x/ {print $1}' "$out/ldd.txt" | sort -u > "$out/libraries.txt"
test -s "$out/libraries.txt"
: > "$out/owners.tsv"
: > "$out/package-names.txt"
for name in ffmpeg fluidsynth; do
  cp "/usr/bin/$name" "$out/tools/real/$name"
  printf '%s\n' "$name" >> "$out/package-names.txt"
done
cp /usr/local/bin/opl2wav "$out/tools/real/opl2wav"
while IFS= read -r source; do
  test -f "$source"
  dest="$out/tools/lib/$(basename "$source")"
  test ! -e "$dest" || { echo '依賴檔名撞號' >&2; exit 2; }
  cp -L "$source" "$dest"
  resolved=$(readlink -f "$source")
  owner=$(dpkg-query -S "$resolved" 2>/dev/null || dpkg-query -S "$source" 2>/dev/null)
  printf '%s\t%s\n' "$source" "$owner" >> "$out/owners.tsv"
  printf '%s\n' "${owner%%: *}" >> "$out/package-names.txt"
done < "$out/libraries.txt"
sort -u "$out/package-names.txt" -o "$out/package-names.txt"
while IFS= read -r package; do
  doc=${package%%:*}
  test -f "/usr/share/doc/$doc/copyright"
  cp -L "/usr/share/doc/$doc/copyright" "$out/licenses/debian/$package.txt"
done < "$out/package-names.txt"
cp -RL /usr/share/common-licenses "$out/licenses/common-licenses"
cp /var/lib/dpkg/status "$out/dpkg-status"
printf '完整 Linux 工具與依賴已收集；尚須原生入口、音色庫與驗收。\n'
