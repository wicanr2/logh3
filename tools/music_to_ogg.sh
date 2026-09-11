#!/usr/bin/env bash
# 把原版的 22 首 `.MDS` 轉成 remake 播得動的 OGG。
#
#   tools/music_to_ogg.sh              # 全部（`_F` 走 ＦＭ、`_M` 走 GM）
#   tools/music_to_ogg.sh OPEN_M BGM_0F
#   tools/music_to_ogg.sh -gm          # 整批都用 GM 音色庫（要拿來對照時）
#
# **兩組曲子走兩條不同的路，因為原版本來就用兩種音源**
# （`docs/re/274`：`_F` ＝ ＦＭ／OPL，`_M` ＝ MPU-401／GM）：
#
#   `_M` → MIDI → OGG     `.MDS` 轉標準 MIDI（`tools/scripts/mds_export.py`），
#                         再用 fluidsynth ＋ FluidR3_GM 合成。`_M` 本來就是
#                         寫給 GM 音源的，所以 GM 音色庫合出來接近當年。
#   `_F` → VGM → OGG      讓 dosgolem 跑原版、錄下原版驅動送給 OPL 的
#                         暫存器寫入，再離線合成（`tools/music_fm_capture.sh`）。
#                         **這一組是原版自己的 ＦＭ 輸出，不是 GM 音色庫。**
#
# 這支腳本照名字分流，`-gm` 可以強制整批走 GM 那條（要拿來對照時用）。
#
# 產物落在 `workplace/assets/music/ogg/`，**不進版控**——它是從原版素材
# 轉出來的，所以它就是原版素材（AGENTS.md）。remake 用 `-music` 指到那裡；
# 缺檔就不播，而且會在 stderr 說缺哪幾首。
#
# ⚠ **不要拿檔頭曲名去判哪一組是 ＦＭ，會判反**（`OPEN_F.MDS` 的曲名寫
# `OPEN MIDI For…`、`OPEN_M.MDS` 寫 `OPEN_OPS…`）。依據是程式碼：
# 兩個外皮一個送 `'f'` 接 OPL 播放常式、一個送 `'m'` 接 MPU-401 播放常式。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${LOGH3_AUDIO_IMAGE:-logh3-audio:2}"
MID="$ROOT/workplace/assets/music"
OGG="$MID/ogg"
SF="${LOGH3_SOUNDFONT:-/usr/share/sounds/sf2/FluidR3_GM.sf2}"
RATE=44100   # 與 internal/audio.SampleRate 同一個數；對不上播放端會抗議

GM_ONLY=0
if [[ "${1:-}" == -gm ]]; then GM_ONLY=1; shift; fi

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "先建 image（唯一需要網路的一步）：" >&2
  echo "  docker build -f tools/Dockerfile.audio -t $IMAGE tools/" >&2
  exit 2
fi

# ── 分流：`_F` 交給 ＦＭ 那條路 ────────────────────────────────────
# 沒有指定曲名時 `_F` 那 11 首全部走過去；指定了就只送指定的那幾首。
if [[ $GM_ONLY -eq 0 ]]; then
  fm=()
  if [[ $# -eq 0 ]]; then
    fm=(BGM_0F BGM_1F BGM_2F BGM_3F BGM_4F BGM_5F BGM_6F BGM_7F END_1F END_2F OPEN_F)
  else
    for a in "$@"; do [[ "$a" == *_[0-9]F || "$a" == OPEN_F ]] && fm+=("$a"); done
  fi
  if [[ ${#fm[@]} -gt 0 ]]; then
    echo "== ＦＭ（原版 OPL 輸出）：${fm[*]} =="
    "$ROOT/tools/music_fm_capture.sh" "${fm[@]}"
  fi
  # `_F` 已經處理掉了，剩下的走 GM。
  if [[ $# -gt 0 ]]; then
    rest=()
    for a in "$@"; do [[ "$a" == *_[0-9]F || "$a" == OPEN_F ]] || rest+=("$a"); done
    set -- ${rest[@]+"${rest[@]}"}
    [[ $# -gt 0 ]] || exit 0
  fi
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
  # 沒有 `-gm` 的時候 `_F` 已經走 ＦＭ 那條路了，別再用 GM 音色蓋掉。
  if [[ $GM_ONLY -eq 0 ]]; then
    keep=(); for a in "${names[@]}"; do
      [[ "$a" == *_[0-9]F || "$a" == OPEN_F ]] || keep+=("$a")
    done
    names=(${keep[@]+"${keep[@]}"})
  fi
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
