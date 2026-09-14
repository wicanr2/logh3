#!/usr/bin/env bash
# 從**原版自己的 OPL（ＦＭ）輸出**錄配樂，不是拿 MIDI 去套 General MIDI 音色庫。
#
#   tools/music_fm_capture.sh                 # 11 首 `_F` 全錄
#   tools/music_fm_capture.sh BGM_0F END_2F   # 只錄這幾首
#
# 三段：
#   1. **觸發**：讓原版自己去播指定那一首（作法見下，逐首確定性、可重跑）
#   2. **錄**：dosgolem 跑原版，`-dump-vgm` 把 OPL 暫存器寫入序列倒成 VGM
#      （dosgolem `docs/spec/190-opl-vgm-dump`）
#   3. **合成**：AdPlug（`tools/opl2wav.cpp`）算成 WAV，ffmpeg 編成 OGG
#
# 產物落在 gitignore 的 `workplace/assets/music/`，**不進版控**——
# 它是從原版素材轉出來的，所以它就是原版素材（CLAUDE.md）。
#
# ## 三條觸發路徑（`docs/re/274`、`docs/re/275`）
#
# | 曲目 | 執行檔 | 怎麼指定 |
# |---|---|---|
# | `BGM_0F`–`BGM_7F` | `GIN3.COM` → `GIN3PS.EXE` | `-poke` 寫 `ds:8F34`（平時配樂曲號）|
# | `END_1F`／`END_2F` | `ENDING.EXE` | 命令列 `9 1`／`9 2`，曲號就是第二個數字 |
# | `OPEN_F` | `OPEN.EXE` | **原版永遠不會開它**（見下），只能換檔案內容 |
#
# ⚠ **`ds:8F34` 的 poke 有一個很窄的時間窗。** 音樂模組初始化會把它清成 0
# （`docs/re/274` §4），而**清成 0 的寫入在 `-watch` 上看不見**——值本來
# 就是 0，沒有變化就不會印出來。所以窗口不是用監看找到的，是掃出來的：
# 太早（#116M、#117.000M）poke 會被清掉，太晚（#117.023M）檔名已經
# `sprintf` 完了，兩種都安靜地回到 `BGM_0F`。**改了輸入腳本就要重掃。**
#
# ⚠ **`END_1F`／`END_2F` 走 Sound Blaster 相容埠。** `ENDING.EXE` 先掃
# `210h`–`260h` 的 DSP reset／ready 埠，在標準基底 `220h` 找到裝置後，
# 從 `base+8/+9`（`228h/229h`）送 OPL。它因此沒有 `0388h` 字面值；
# 這不是沒走 FM。該分支必須傳 `-sound-blaster`，不能只傳 `-adlib`。
#
# ⚠ **`OPEN_F.MDS` 在出貨的遊戲裡是死素材。** `OPEN.EXE` 的 ＦＭ 與 MIDI
# 兩個外皮都寫死 `'m'`，所以不論偵測到什麼音源，開頭動畫一律讀
# `OPEN_M.MDS`（`docs/re/274` §8，457 份收據裡 689 次 `open_m.mds`、
# 0 次 `open_f.mds`）。這裡用**換檔案內容**的方式讓那支 OPL 驅動去播
# `OPEN_F` 的資料——那是「原版的 OPL 驅動演奏這份資料」會有的聲音，
# **不是**「玩家當年聽到的開頭音樂」。收據裡標成 `FM(OPL,substituted)`。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
IMAGE="${LOGH3_AUDIO_IMAGE:-logh3-audio:2}"
ORIG="$ROOT/org_game/extracted"
WORK="$ROOT/workplace/fmcap"
GAME="$WORK/game"                      # 換檔觸發用的暫存遊戲目錄（原版唯讀不動）
OUT="$ROOT/workplace/assets/music"
RATE=44100                             # 與 internal/audio.SampleRate 同一個數
TICK="${FM_TICK:-20000}"
POKE_AT="${FM_POKE_AT:-117022200}"     # 上面說的那個窗口
# 錄多久（指令數）。`-tick 20000` ＋ 遊戲把 PIT 分頻設成 2048 之下，
# 一秒約 11.65 M 道指令。曲子會循環，所以這是「錄幾秒」不是「曲子多長」。
BGM_STEPS="${FM_BGM_STEPS:-1600000000}"
END_STEPS="${FM_END_STEPS:-1600000000}"
OPEN_STEPS="${FM_OPEN_STEPS:-1200000000}"

command -v docker >/dev/null || { echo "要 docker" >&2; exit 2; }
[[ -d "$ORIG" ]] || { echo "找不到原版目錄 $ORIG" >&2; exit 2; }
if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "先建 image（唯一需要網路的一步）：" >&2
  echo "  docker build -f tools/Dockerfile.audio -t $IMAGE tools/" >&2
  exit 2
fi

ALL=(BGM_0F BGM_1F BGM_2F BGM_3F BGM_4F BGM_5F BGM_6F BGM_7F END_1F END_2F OPEN_F)
names=("$@"); [[ ${#names[@]} -eq 0 ]] && names=("${ALL[@]}")

mkdir -p "$OUT/vgm" "$OUT/ogg" "$WORK/tmp"
head=$(git -C "$GOLEM" rev-parse --short HEAD)
dirty=$(git -C "$GOLEM" status --porcelain | wc -l)
[[ $dirty -eq 0 ]] || echo "⚠ dosgolem 工作樹有 $dirty 個未提交檔案——收據綁不住版本" >&2
echo "dosgolem HEAD $head"

# 進遊戲本體的輸入腳本（收據 `key1` 那一組，`docs/playtest/` 有說明）。
# 兩次點擊跳過開頭動畫並開新遊戲，`-press 1,3,8,enter` 答防拷問答。
PROLOGUE=(-clicks 40000000:300:170,95000000:300:170
          -press 1,3,8,enter -press-at 105000000 -press-every 2000000
          -click-hold 3000000)

# run_probe <標籤> <exe> <steps> <掛載目錄> <框曲子的檔名> [額外旗標...]
#
# `-vgm-clear-on-open` 是把 OPL 輸出框到某一首上的錨：開機會連放好幾首
# （開頭動畫的 `OPEN_M` 就在前面），暫存器串是接起來的，事後從序列裡
# 找分界只能用猜的。**沒開到那個檔 probe 就不寫 VGM**，所以「poke 沒生效」
# 會變成缺檔，不會變成一份內容是別首、檔名卻寫著這一首的音檔。
run_probe() {
  local tag="$1" exe="$2" steps="$3" root="$4" anchor="$5"; shift 5
  ( cd "$GOLEM" && DOSGOLEM_ORIG="$root" DOSGOLEM_CPUS="${DOSGOLEM_CPUS:-2}" \
      DOSGOLEM_MEM="${DOSGOLEM_MEM:-2g}" timeout "${FM_TIMEOUT:-580}" \
      tools/go.sh run ./cmd/probe -exe "$exe" -root /orig -empty-env \
        -tick "$TICK" -steps "$steps" -vgm-clear-on-open "$anchor" "$@" \
        -dump-vgm "/src/workplace/fmcap-$tag.vgm" ) > "$WORK/tmp/$tag.log" 2>&1 || true
}

# `FM_RENDER_ONLY=1` 跳過錄製，只拿現有的 VGM 重算音檔——換 `FM_NORM`
# 不必重跑一小時的執行器。**VGM 才是收據，OGG 是它的算出來的樣子。**
[[ "${FM_RENDER_ONLY:-}" == 1 ]] && names=()
for n in "${names[@]}"; do
  src="$ORIG/$n.MDS"
  [[ -f "$src" ]] || { echo "  $n：原版沒有這個檔，跳過" >&2; continue; }
  route=FM
  case "$n" in
    BGM_[0-7]F)
      num=$(printf '%02d' "${n:4:1}")
      run_probe "$n" /orig/GIN3.COM "$BGM_STEPS" "$ORIG" "$n.MDS" \
        "${PROLOGUE[@]}" -poke "lin:5DA64@$POKE_AT=$num"
      ;;
    END_1F|END_2F)
      # 曲號就是命令列的第二個數字（`docs/re/274` §9）。
      run_probe "$n" /orig/ENDING.EXE "$END_STEPS" "$ORIG" "$n.MDS" \
        -sound-blaster -args " 9 ${n:4:1}"
      ;;
    OPEN_F)
      # 換檔內容觸發。暫存遊戲目錄建一次就好，之後只換那一個檔。
      if [[ ! -f "$GAME/GIN3.COM" ]]; then
        # **不要拿 `-v` 掛單一檔案去蓋**：來源路徑打錯一個字，dockerd 會用
        # root 建一個目錄頂替，而容器裡讀到的是空目錄、不會報錯（CLAUDE.md）。
        rm -rf "$GAME"; cp -r "$ORIG" "$GAME"; chmod -R u+w "$GAME"
      fi
      cp "$src" "$GAME/OPEN_M.MDS"
      route="FM(substituted)"
      # 開頭動畫不送任何輸入——一按下去動畫被跳過，音樂跟著斷。
      # 錨仍然是 `OPEN_M.MDS`：那是原版**實際開的**檔名，內容才是 `OPEN_F`。
      run_probe "$n" /orig/GIN3.COM "$OPEN_STEPS" "$GAME" OPEN_M.MDS
      ;;
    *) echo "  $n：沒有這一首的觸發路徑（`_M` 那一組是 MPU-401/GM，走 music_to_ogg.sh）" >&2; continue ;;
  esac

  vgm="$GOLEM/workplace/fmcap-$n.vgm"
  if [[ ! -s "$vgm" ]]; then
    # probe 在「沒開到那個檔」時刻意不寫 VGM，所以缺檔＝觸發沒生效
    # （poke 的時間窗跑掉了、步數不夠、或那條路根本沒走到）。
    echo "  $n：沒錄到——看 $WORK/tmp/$n.log" >&2
    continue
  fi
  mv "$vgm" "$OUT/vgm/$n.vgm"
  # 每首的來源資訊寫成 sidecar，合成那一段再據此重建收據。
  # **沒有這幾行的音檔不算數**——dosgolem 的行為一改（即使是改對），
  # 同一組輸入跑出來的東西就可能不同（CLAUDE.md「收據綁執行器版本」）。
  printf 'route=%s\nsrc=%s\ngolem=%s%s\n' "$route" \
    "$(sha256sum "$src" | cut -d' ' -f1)" \
    "$head" "$([[ $dirty -gt 0 ]] && echo '+dirty')" > "$OUT/vgm/$n.meta"
  printf '  %-8s 錄到 %s\n' "$n" "$(grep -o '（YM.*' "$WORK/tmp/$n.log" | tr -d '（）' || echo '?')"
done

# ── 合成：一次做完整批，套**同一個**增益 ────────────────────────────
#
# ⚠ **音量是一個取捨，這裡不替使用者決定。** 量到的峰值：開頭動畫那一首
# 在 −0.4 dBFS，遊戲本體八首在 −11 至 −21 dBFS——**那個 20 dB 的落差是
# 原版自己寫進 OPL 音量暫存器的**（同一支驅動、同一顆晶片），不是錄壞。
#
#   FM_NORM=batch（預設）全批共用一個增益。**保留曲子之間的音量關係**，
#                        代價是遊戲本體那幾首聽起來比 `_M` 那組小很多。
#   FM_NORM=track        每首各自把峰值推到 −1 dBFS。聽感一致，
#                        **但原版的音量關係被抹掉了**。
#   FM_NORM=off          不動，就是 AdPlug 算出來的樣子。
#
# 不論哪一種，套用的增益都記進 `SOURCES.tsv`，換一種重跑就好。
# 每次都重算整批（不只這次錄的那幾首），補錄一首不會讓舊的停在舊增益上。
NORM="${FM_NORM:-batch}"
mapfile -t have < <(cd "$OUT/vgm" && ls *.vgm 2>/dev/null | sed 's/\.vgm$//')
if [[ ${#have[@]} -eq 0 ]]; then
  echo "沒有任何 VGM 可以合成" >&2; exit 1
fi
echo "== 合成 ${#have[@]} 首（共用增益）=="
docker run --rm --network none \
  --memory 2g --cpus 2 --pids-limit 128 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -v "$OUT/vgm:/vgm:ro" -v "$OUT/ogg:/ogg" -w /tmp "$IMAGE" \
  sh -c '
    set -e
    names="$1"; rate="$2"; mode="$3"; peak=-999
    : > /ogg/.gain
    for n in $names; do
      opl2wav "/vgm/$n.vgm" "/tmp/$n.wav" "$rate" 2>/dev/null
      p=$(ffmpeg -hide_banner -i "/tmp/$n.wav" -af volumedetect -f null /dev/null 2>&1 |
          sed -n "s/.*max_volume: \(-*[0-9.]*\) dB.*/\1/p")
      echo "$n $p" >> /tmp/peaks
      peak=$(awk -v a="$peak" -v b="$p" "BEGIN{print (b>a)?b:a}")
    done
    batch=$(awk -v p="$peak" "BEGIN{printf \"%.2f\", -1.0 - p}")
    echo "全批峰值 ${peak} dBFS（FM_NORM=$mode）"
    for n in $names; do
      case "$mode" in
        track) g=$(awk -v n="$n" "\$1==n{printf \"%.2f\", -1.0 - \$2}" /tmp/peaks) ;;
        off)   g=0.00 ;;
        *)     g="$batch" ;;
      esac
      printf "%s\t%s\n" "$n" "$g" >> /ogg/.gain
      ffmpeg -v error -y -i "/tmp/$n.wav" -af "volume=${g}dB" \
        -ar "$rate" -c:a libvorbis -q:a 5 "/ogg/$n.ogg"
      printf "  %-8s 峰值 %6s dBFS  增益 %6s dB\n" \
        "$n" "$(awk -v n="$n" "\$1==n{print \$2}" /tmp/peaks)" "$g"
      rm -f "/tmp/$n.wav"
    done
  ' sh "${have[*]}" "$RATE" "$NORM"

{
  printf '# FM_NORM=%s；曲名\t來源\t增益(dB)\t原始 .MDS SHA-256\tVGM SHA-256\tOGG SHA-256\tdosgolem HEAD\n' "$NORM"
  for n in "${have[@]}"; do
    r=$(sed -n 's/^route=//p' "$OUT/vgm/$n.meta" 2>/dev/null || echo '?')
    s=$(sed -n 's/^src=//p'   "$OUT/vgm/$n.meta" 2>/dev/null || echo '?')
    gh=$(sed -n 's/^golem=//p' "$OUT/vgm/$n.meta" 2>/dev/null || echo '?')
    g=$(awk -v n="$n" -F'\t' '$1==n{print $2}' "$OUT/ogg/.gain")
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$n" "$r" "$g" "$s" \
      "$(sha256sum "$OUT/vgm/$n.vgm" | cut -d' ' -f1)" \
      "$(sha256sum "$OUT/ogg/$n.ogg" | cut -d' ' -f1)" "$gh"
  done
} > "$OUT/ogg/SOURCES.tsv"

echo
column -t -s$'\t' "$OUT/ogg/SOURCES.tsv" | cut -c1-60
echo
echo "VGM → $OUT/vgm   OGG → $OUT/ogg   收據 → $OUT/ogg/SOURCES.tsv"
echo "跑遊戲：tools/gui.sh run ./cmd/gin3 -music $OUT/ogg"
