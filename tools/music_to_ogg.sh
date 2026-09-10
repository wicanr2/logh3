#!/usr/bin/env bash
# 把原版的 22 首 `.MDS` 轉成 remake 播得動的 OGG。
#
#   tools/music_to_ogg.sh              # 全部
#   tools/music_to_ogg.sh OPEN_M BGM_0M
#
# 兩段：`.MDS` → MIDI（`tools/scripts/mds_export.py`，走 tools/py.sh）
#       MIDI  → OGG（fluidsynth ＋ ffmpeg，走 tools/Dockerfile.audio）
#
# 產物落在 `workplace/assets/music/ogg/`，**不進版控**——它是從原版素材
# 轉出來的，所以它就是原版素材（AGENTS.md）。remake 用 `-music` 指到那裡；
# 缺檔就不播，而且會在 stderr 說缺哪幾首。
#
# ⚠ **`_F` 那一組不是原音。** `_M` 是寫給 MPU-401／GM 的，用 GM 音色庫
# 合出來接近當年的聲音；`_F` 是寫給 FM（OPL）的，用 GM 音色只會得到
# 「同一首曲子、不同的音源」。要 FM 原音得換一條路：讓 dosgolem 跑原版、
# 錄下 OPL 暫存器的寫入序列（`machine.OPLWrite` 已經連步數一起記了）再合成
# ——那條還沒做，記在 worklist 的 `music-fm-capture`。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${LOGH3_AUDIO_IMAGE:-logh3-audio:1}"
MID="$ROOT/workplace/assets/music"
OGG="$MID/ogg"
SF="${LOGH3_SOUNDFONT:-/usr/share/sounds/sf2/FluidR3_GM.sf2}"
RATE=44100   # 與 internal/audio.SampleRate 同一個數；對不上播放端會抗議

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "先建 image（唯一需要網路的一步）：" >&2
  echo "  docker build -f tools/Dockerfile.audio -t $IMAGE tools/" >&2
  exit 2
fi

# 第一段：.MDS → MIDI。已經有就不重跑——這一段要原版檔案在位。
if [[ ! -d "$MID" || -z "$(ls "$MID"/*.mid 2>/dev/null)" ]]; then
  echo "== .MDS → MIDI =="
  tools/py.sh tools/scripts/mds_export.py
fi
mkdir -p "$OGG"

names=("$@")
if [[ ${#names[@]} -eq 0 ]]; then
  mapfile -t names < <(cd "$MID" && ls *.mid 2>/dev/null | sed 's/\.mid$//')
fi
[[ ${#names[@]} -gt 0 ]] || { echo "$MID 裡沒有 .mid——原版檔案在位嗎？" >&2; exit 2; }

echo "== MIDI → OGG（$RATE Hz）=="
for n in "${names[@]}"; do
  [[ -f "$MID/$n.mid" ]] || { echo "  $n：沒有 .mid，跳過" >&2; continue; }
  # fluidsynth 直接輸出 wav 再交給 ffmpeg 編碼。**不要讓 fluidsynth 自己
  # 輸出 ogg**：它的 libsndfile 不一定編得進 Vorbis，缺的時候它會安靜地
  # 寫出一個 wav 內容卻叫 .ogg 的檔，而播放端要到解碼才會發現。
  docker run --rm --network none \
    --memory 2g --cpus 2 --pids-limit 128 \
    --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" \
    -v "$MID:/mid" -v "$OGG:/ogg" -w /tmp "$IMAGE" \
    sh -c "fluidsynth -ni -F /tmp/o.wav -r $RATE -g 0.8 '$SF' '/mid/$n.mid' >/dev/null 2>&1 &&
           ffmpeg -v error -y -i /tmp/o.wav -ar $RATE -c:a libvorbis -q:a 5 '/ogg/$n.ogg'"
  printf '  %-8s %s\n' "$n" "$(du -h "$OGG/$n.ogg" 2>/dev/null | cut -f1)"
done

echo
echo "轉好 $(ls "$OGG"/*.ogg 2>/dev/null | wc -l) 首 → $OGG"
echo "跑遊戲：tools/gui.sh run ./cmd/gin3 -music $OGG"
