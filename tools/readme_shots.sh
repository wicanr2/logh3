#!/usr/bin/env bash
# 產生 README 的截圖。**每一張都先與原版收據逐位元組比過才寫出去**——
# 沒比過的圖只是「看起來像」，那不算數（AGENTS.md 的完成定義）。
#
#   tools/readme_shots.sh          # 全部
#   tools/readme_shots.sh 06       # 只做一張（用檔名前綴）
#
# 顏色用收據自己的 pal.bin：遊戲本體的調色盤與主選單不同，用錯的那份
# 色號陣列還是全中，但 PNG 會整個偏色。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
only="${1:-}"

MENU='move 300 210'
GAME="$MENU; click; type 138; enter"
CMD="$GAME; rclick 300 240; rclick 300 240; move 600 340"

# 檔名前綴|收據|腳本
shots=(
  "01-main-menu|r0|$MENU"
  "02-copy-protection|cb1|$MENU; click"
  "03-month-report|key1|$GAME"
  "04-command-phase|main1e|$CMD"
  "05-load-game|load1e|$CMD; click 544 125; click 560 145"
  "06-info|info1e|$CMD; click 352 125; click 400 145"
  "07-admiral-list|lst2e|$CMD; click 448 125; click 460 165"
  "08-settings|settings1e|$CMD; click 544 125; click 560 185"
  "09-admiral-detail|adm1e|$CMD; click 352 125; click 400 125; click 400 85"
  "10-biography|keireki1e|$CMD; click 352 125; click 400 125; click 400 85; click 392 330"
  "11-fleet-org|fleet1e|$CMD; click 352 125; click 400 125; click 400 85; click 488 330"
)

# 戰術畫面沒有同狀態收據可比——remake 的行軍速度還沒對到原版（原版要按
# 六次實行才會合，remake 一次就到），所以它不走上面那條「比過才寫出去」
# 的迴圈，另外拍。畫面組成本身由 TestBattleMatchesOracle 與
# TestMinimapMatchesOracle 逐位元組驗過。
BATTLE="$CMD; click 544 125; click 560 145; click 300 322; click 318 208; click 544 125; click 560 125"

out=workplace/readme-shots
rm -rf "$out"; mkdir -p "$out"
fail=0
for s in "${shots[@]}"; do
  IFS='|' read -r name batch script <<<"$s"
  [[ -n "$only" && "$name" != "$only"* ]] && continue
  oracle="workplace/oracle/$batch"
  if [[ ! -f "$oracle/screen.bin" ]]; then
    echo "  $name 缺收據 $batch，跳過"
    continue
  fi
  rm -rf "$out/$name"; mkdir -p "$out/$name"
  tools/go.sh run ./cmd/play -orig /orig -pal "$oracle/pal.bin" \
    -out "$out/$name" -script "$script" >/dev/null
  last=$(ls "$out/$name"/*.bin | sort | tail -1)
  if cmp -s "$last" "$oracle/screen.bin"; then
    cp "${last%.bin}.png" "docs/images/$name.png"
    echo "  $name 逐位元組相同（收據 $batch）"
  else
    echo "  $name **與收據 $batch 不同**，不寫出去"
    fail=1
  fi
done

# 12 戰術畫面：沒有同狀態收據可比（remake 的行軍速度還沒對到原版——原版
# 要按六次「實行」才會合，remake 一次就到），所以不走上面「比過才寫出去」
# 的迴圈。畫面組成本身由 `TestBattleMatchesOracle` 與
# `TestMinimapMatchesOracle` 逐位元組驗過。
if [[ -z "$only" || "12" == "$only"* ]] && [[ -f workplace/oracle/s3fight/pal.bin ]]; then
  rm -rf "$out/12-battle"; mkdir -p "$out/12-battle"
  tools/go.sh run ./cmd/play -orig /orig -pal workplace/oracle/s3fight/pal.bin \
    -out "$out/12-battle" -script "$CMD; click 544 125; click 560 145; \
      click 300 322; click 318 208; click 544 125; click 560 125" >/dev/null
  last=$(ls "$out/12-battle"/*.png | sort | tail -1)
  cp "$last" docs/images/12-battle.png
  echo "  12-battle 拍好了（無收據可比，見上面的註）"
fi

exit $fail
