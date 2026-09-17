#!/usr/bin/env bash
# Ebiten 顯示層的同狀態驗收：走玩家路徑，把 Ebiten 畫出來的色號陣列
# 與原版收據逐位元組比。
#
#   tools/gui_check.sh
#
# **這是顯示層唯一算數的驗收**：視窗開得起來、看起來像，都不算。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

cases=(
  "r0|move 300 210"
  "cb1|move 300 210; click"
  "key1|move 300 210; click; type 138; enter"
  "cbAlly|move 300 230; click"
  "keyAlly|move 300 230; click; type 138; enter"
  "main1e|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340"
  "load1e|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 544 125; click 560 145"
  # vs3 收據在交棒視窗出現後還點了 (400,262)；那點在確認框內、熱區外，
  # 狀態不變但游標會搬過去（docs/re/280）。漏掉就只差兩個游標位置 274 px。
  "vs3|move 300 250; click; type 126; enter; click 400 262"
  "vs4|move 300 250; click; type 126; enter; click 432 263"
  "vs5|move 300 250; click; type 126; enter; click 432 263; rclick 300 240; rclick 300 240; click 544 125; click 560 125"
  # 戰略艦隊命令（docs/spec/051）：dec1 的序列再點右下角清單第 0 列開艦隊視窗；
  # fcmd-done 走完「作戰 → 移動 → 亞爾提那 → 決定」；fcmd-exec 再按 系統指令 → 實行
  # （艦隊走到亞爾提那、那一格變黑、首都經濟力 82339；中間那張過場點兩下關掉）。
  "fcmd-win|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308"
  "fcmd-done|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 312; click 85 128; click 432 263"
  "fcmd-exec|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 312; click 85 128; click 432 263; click 544 125; click 560 125; click 560 125; click 560 125"
  # 實行之後停在「出發前往」那張過場（規格 059），點一下換「已經到達」（docs/re/332）。
  "fcmd-turn-depart|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 312; click 85 128; click 432 263; click 544 125; click 560 125"
  "fcmd-turn-arrive|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 312; click 85 128; click 432 263; click 544 125; click 560 125; click 560 125"
  # 待命 → 「指定待命行星」再點布連娜（docs/re/314）。排除小圖最後一列
  # （x 152–279、y 347）：原版那一列讀到地圖緩衝區外（規格 051 已知差異）。
  "fcmd-standby-pick|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 392; click 350 250|152,347,128,1"
  # 占領 → 費沙 → 參謀提醒兩頁 → 「要占領中立軍費沙星系嗎？」（docs/re/315）。
  # fcmd-occ-t2 走完「占領 → 亞斯提 → 決定」再按兩次 系統指令 → 實行（第二回合在星域 7；每回合的過場點兩下）
  "fcmd-occ-t2|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 332; click 193 245; click 432 263; click 544 125; click 560 125; click 560 125; click 560 125; click 544 125; click 560 125; click 560 125; click 560 125"
  # 作戰 → 駐留 → 亞爾提那 → 決定 → 實行 → 過場點兩下：走到目的地整支解散（docs/re/329）。
  "fcmd-garrison-t1|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 352; click 85 128; click 432 263; click 544 125; click 560 125; click 560 125; click 560 125"
  # 占領費沙第三回合：電腦拿下威爾勃恩的訊息框（過場與報告點掉四下）、再兩下是士氣值低落（docs/re/333、335）。
  "fcmd-turn-ai-capture|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 332; click 163 136; click 300 240; click 300 240; click 432 263; click 544 125; click 560 125; click 560 125; click 560 125; click 544 125; click 560 125; click 560 125; click 560 125; click 544 125; click 560 125; click 560 125; click 560 125; click 560 125; click 560 125"
  "fcmd-turn-morale-low|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 332; click 163 136; click 300 240; click 300 240; click 432 263; click 544 125; click 560 125; click 560 125; click 560 125; click 544 125; click 560 125; click 560 125; click 560 125; click 544 125; click 560 125; click 560 125; click 560 125; click 560 125; click 560 125; click 560 125; click 560 125"
  "fcmd-occupy-phezzan3|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 332; click 163 136; click 300 240; click 300 240"
  # 一覽表 → 提督資料 → 點列 0 → 萊因哈特的詳細資料（第１艦隊／出擊中／經歷／結束）疊在清單上（docs/re/323）。
  "fcmd-lst-adm-row|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 448 125; click 460 165; click 460 90"
  # 一覽表 → 艦隊資料 → 點列 0 → 艦隊視窗疊在清單上，那一列反白。
  "fcmd-lst-fleet-row|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 448 125; click 460 145; click 460 90"
  # 一覽表 → 星系資料 → 稅金分頁：自己的格「徵稅額 ( 稅率%)」、別家 --- (-----)（docs/re/327）。
  "fcmd-lst-star-tab-tax|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 448 125; click 460 125; click 584 70"
  # 變更 → 編組：這一支的艦隊組成疊在艦隊視窗上（docs/re/316）。
  "fcmd-change-org|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 365; click 497 392"
  # 編組的編輯器（docs/re/317）：槽 19 放戰艦 → 換巡航艦列 → 槽 21 放 → 右鍵拆槽 19
  # → 陣形 → 圓錐（8/8、機動力 81％）→ 再點槽 0 跳「無法超過…」訊息框。
  "fcmd-edit-full-msg|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 365; click 497 392; click 176 240; click 360 210; click 240 240; rclick 176 240; click 510 310; click 505 250; click 144 176"
  # 出擊命令那條路的同一支編輯器：槽 19 放戰艦、換列、槽 21 放巡航艦、決定 → 地圖。
  "fcmd-fok-decide|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 176 240; click 360 210; click 240 240; click 510 340"
  # 變更 → 副官 → 副官 → 變更 → 待命中副官資料第 1 列（米達麥亞）→ 出擊(決定) → 結束：
  # 艦隊視窗的副官換成米達麥亞、維持費 29（docs/re/318）。
  # 右下角清單：行星分頁點第 0 列 → 行星情報視窗；提督分頁 ▼ 一頁；提督分頁第 0 列 → 出擊 →
  # 確認 → 決定 → 第２艦隊、提督分頁還選著（docs/re/319）。
  "fcmd-tab-planet-row|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 415 290; click 450 308"
  "fcmd-tab-admiral-down|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 560 290; click 617 388"
  "fcmd-tab-fleet2|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 560 290; click 450 308; click 496 320; click 480 350; click 510 350"
  # 左鍵點瓦爾哈拉 → 星系情報視窗（排除小圖最後一列 128 個像素，docs/re/320）。
  "fcmd-rinfo|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 32 72|24,212,128,1"
  "fcmd-adj-confirm-close|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 365; click 497 372; click 430 340; click 430 350; click 450 110; click 496 320; click 480 340"
)
fail=0
for c in "${cases[@]}"; do
  IFS='|' read -r batch script except <<<"$c"
  oracle="workplace/oracle/$batch/screen.bin"
  if [[ ! -f "$oracle" ]]; then
    echo "  $batch 缺收據，跳過"
    continue
  fi
  tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data \
    -script "$script" -shot-bin /src/workplace/gui-check.bin -shot-after 2 \
    >/dev/null 2>&1
  if [[ -n "$except" ]]; then
    # 有排除矩形的用 python 比（x,y,w,h 之外要逐位元組相同）。
    if tools/py.sh tools/scripts/cmp_except.py workplace/gui-check.bin "$oracle" "$except"; then
      echo "  $batch 排除 ($except) 之外逐位元組相同"
    else
      echo "  $batch **不同**"
      fail=1
    fi
  elif cmp -s workplace/gui-check.bin "$oracle"; then
    echo "  $batch 逐位元組相同"
  else
    echo "  $batch **不同**"
    fail=1
  fi
done
exit $fail
