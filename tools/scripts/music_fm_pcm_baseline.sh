#!/usr/bin/env bash
# 規格 074 的 Linux 無損基準；由 bash tools/audio.sh bash /src/tools/scripts/music_fm_pcm_baseline.sh 執行。
# 原版 VGM 在 music_fm_lock.py 再驗證；此處不宣稱跨平台已通過。
set -euo pipefail
input=/work/music69-fm/full
output=/work/music69-windows/linux-pcm-baseline
test -d "$input"
test ! -e "$output"
mkdir "$output"
for track in BGM_{0..7}F END_1F END_2F OPEN_F; do
    opl2wav "$input/$track.vgm" "$output/$track.wav" 44100
    ffmpeg -v error -y -i "$output/$track.wav" -af volume=-1.00dB \
        -ar 44100 -c:a pcm_f32le -f f32le "$output/$track.f32"
done
sha256sum /usr/local/bin/opl2wav /usr/bin/ffmpeg > "$output/tools.sha256"
ffmpeg -version > "$output/ffmpeg-version.txt"
sha256sum "$output"/*.wav "$output"/*.f32 > "$output/files.sha256"
