#!/bin/sh
# #69 原生 GM 真實工具驗收；僅由 bash tools/audio.sh sh /src/tools/scripts/music_native_check.sh 執行。
# 原始遊戲／保存 MIDI 唯讀；只寫 /work/music69-native。
set -eu
BIN=/work/music69-native/music-import
ROOT=/work/music69-native
FONT=/usr/share/sounds/sf2/FluidR3_GM.sf2
mkdir -p "$ROOT/reference" "$ROOT/check"

# 以前一個完整原生批次為待驗輸出，不重新把 FM 套 GM。
for name in BGM_0M END_2M OPEN_M; do
  fluidsynth -ni -F "$ROOT/reference/$name.wav" -r 44100 -g 0.8 \
    "$FONT" "/src/assets/original/music/mid/$name.mid" >"$ROOT/check/$name-synth.log" 2>&1
  ffmpeg -v error -y -i "$ROOT/reference/$name.wav" -ar 44100 \
    -c:a libvorbis -q:a 5 "$ROOT/reference/$name.ogg"
  ffmpeg -v error -y -i "$ROOT/reference/$name.ogg" -f f32le "$ROOT/check/$name-reference.pcm"
  ffmpeg -v error -y -i "$ROOT/曲庫 測試/$name.ogg" -f f32le "$ROOT/check/$name-native.pcm"
  cmp "$ROOT/check/$name-reference.pcm" "$ROOT/check/$name-native.pcm"
  sha256sum "$ROOT/check/$name-reference.pcm" > "$ROOT/check/$name-pcm.sha256"
  rm "$ROOT/reference/$name.wav" "$ROOT/check/$name-reference.pcm" "$ROOT/check/$name-native.pcm"
  echo "$name：舊命令／原生入口解碼 PCM 完全一致"
done

reject() {
  label=$1; shift
  if "$BIN" "$@" >"$ROOT/check/$label.log" 2>&1; then
    echo "$label 未拒絕" >&2; exit 1
  fi
  echo "$label：已拒絕"
}
reject no-gm -source /orig -output "$ROOT/rejected"
reject fm -gm-only -source /orig -output "$ROOT/rejected" -tools /usr/bin -soundfont "$FONT" -tracks BGM_0F
reject original-output -gm-only -source /orig -output /orig/converted -tools /usr/bin -soundfont "$FONT"
reject missing-tools -gm-only -source /orig -output "$ROOT/rejected" -tools "$ROOT/missing-tools" -soundfont "$FONT"
reject wrong-font -gm-only -source /orig -output "$ROOT/rejected" -tools /usr/bin -soundfont /src/go.mod
test ! -e "$ROOT/rejected"
echo '原生 GM 三曲 PCM 比較與前置拒絕通過。'
