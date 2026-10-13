#!/bin/sh
# 在tools/audio.sh容器內合成已保留事件的原曲MIDI，輸出只供本機完整版。
set -eu
source_dir="${1:?需MIDI來源目錄}"
output_dir="${2:?需OGG輸出目錄}"
soundfont="${3:?需已核對音色庫}"
test -d "$source_dir"
test -f "$soundfont"
test -d "$(dirname "$output_dir")"
test "$(stat -c %u "$(dirname "$output_dir")")" = "$(id -u)"
mkdir -p "$output_dir"
test "$(stat -c %u "$output_dir")" = "$(id -u)"
for midi in "$source_dir"/*.mid; do
    name=$(basename "$midi" .mid)
    timeout 90s fluidsynth -ni -F /tmp/hd.wav -T wav -O float -r 44100 -g 0.5 -R 1 -C 0 \
        -o synth.polyphony=256 "$soundfont" "$midi" >"$output_dir/$name.synth.log" 2>&1
    timeout 90s ffmpeg -nostdin -v error -y -i /tmp/hd.wav \
        -af 'highpass=f=35,equalizer=f=300:t=q:w=0.7:g=-1,afade=t=in:d=0.008,areverse,afade=t=in:d=0.03,areverse,loudnorm=I=-18:TP=-1.5:LRA=11' \
        -ar 44100 -ac 2 -c:a libvorbis -q:a 7 "$output_dir/$name.ogg"
    printf '%s\n' "$name 完成"
done
