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
  "vs5|move 300 250; click; type 126; enter; click 432 263; rclick 300 240; rclick 300 240; click 544 125; click 560 125; tick 18"
  # 戰略艦隊命令（docs/spec/051）：dec1 的序列再點右下角清單第 0 列開艦隊視窗；
  # fcmd-done 走完「作戰 → 移動 → 亞爾提那 → 決定」；fcmd-exec 再按 系統指令 → 實行
  # （艦隊走到亞爾提那、那一格變黑、首都經濟力 82339；中間那張過場點兩下關掉）。
  "fcmd-win|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308"
  "fcmd-done|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 312; click 85 128; click 432 263"
  "fcmd-exec|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 312; click 85 128; click 432 263; click 544 125; click 560 125; tick 18; click 560 125; click 560 125"
  "supply84-gui|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 544 125; click 560 125; tick 18"
  # 實行之後停在「出發前往」那張過場（規格 059），點一下換「已經到達」（docs/re/332）。
  # 實行之後、回合處理之前：電腦座位的「戰略模式思考中」（規格 059，停 18 tick）。
  "fcmd-turn-think|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 312; click 85 128; click 432 263; click 544 125; click 560 125"
  "fcmd-turn-depart|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 312; click 85 128; click 432 263; click 544 125; click 560 125; tick 18"
  "fcmd-turn-arrive|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 312; click 85 128; click 432 263; click 544 125; click 560 125; tick 18; click 560 125"
  # 待命 → 「指定待命行星」再點布連娜（docs/re/314）。排除小圖最後一列
  # （x 152–279、y 347）：原版那一列讀到地圖緩衝區外（規格 051 已知差異）。
  "fcmd-standby-pick|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 392; click 350 250|152,347,128,1"
  # 占領 → 費沙 → 參謀提醒兩頁 → 「要占領中立軍費沙星系嗎？」（docs/re/315）。
  # fcmd-occ-t2 走完「占領 → 亞斯提 → 決定」再按兩次 系統指令 → 實行（第二回合在星域 7；每回合的過場點兩下）
  "fcmd-occ-t2|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 332; click 193 245; click 432 263; click 544 125; click 560 125; tick 18; click 560 125; click 560 125; click 544 125; click 560 125; tick 18; click 560 125; click 560 125"
  # 作戰 → 駐留 → 亞爾提那 → 決定 → 實行 → 過場點兩下：走到目的地整支解散（docs/re/329）。
  "fcmd-garrison-t1|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 352; click 85 128; click 432 263; click 544 125; click 560 125; tick 18; click 560 125; click 560 125"
  # 占領費沙第三回合：電腦拿下威爾勃恩的訊息框（過場與報告點掉四下）、再兩下是士氣值低落（docs/re/333、335）。
  "fcmd-turn-ai-capture|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 332; click 163 136; click 300 240; click 300 240; click 432 263; click 544 125; click 560 125; tick 18; click 560 125; click 560 125; click 544 125; click 560 125; tick 18; click 560 125; click 560 125; click 544 125; click 560 125; tick 18; click 560 125; click 560 125; click 560 125; click 560 125"
  "fcmd-turn-morale-low|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 332; click 163 136; click 300 240; click 300 240; click 432 263; click 544 125; click 560 125; tick 18; click 560 125; click 560 125; click 544 125; click 560 125; tick 18; click 560 125; click 560 125; click 544 125; click 560 125; tick 18; click 560 125; click 560 125; click 560 125; click 560 125; click 560 125; click 560 125"
  # 第三回合的士氣報告照收據點在地圖 (300,240)、第四回合點掉占據報告 → 全滅報告（報告視窗是刻意差異：原版畫成花紋，#57 裁決修成正常文字）。
  "fcmd-turn-annihilated-report|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 332; click 163 136; click 300 240; click 300 240; click 432 263; click 544 125; click 560 125; tick 18; click 560 125; click 560 125; click 544 125; click 560 125; tick 18; click 560 125; click 560 125; click 544 125; click 560 125; tick 18; click 560 125; click 560 125; click 560 125; click 560 125; click 560 125; click 560 125; click 300 240; click 544 125; click 560 125; tick 18; click 560 125|320,160,320,160"
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
  # S3「巴米利恩會戰」：取進度 → S3 → 帝國 → 一覽表 → 艦隊資料 → 第 8 列（第１０艦隊）
  # → 艦隊視窗；再 作戰 → 待命 → 指定待命行星。兩張的編號都是半形兩位數（#74、docs/re/338）。
  "fcmd-s3-n10-win|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 544 125; click 560 145; click 300 322; click 318 208; click 448 125; click 460 145; click 450 230"
  "fcmd-s3-n10-standby|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 544 125; click 560 145; click 300 322; click 318 208; click 448 125; click 460 145; click 450 230; click 490 345; click 497 392|152,347,128,1"
  # 逐格開戰之前的「達希利星系內／發現敵方艦隊。」（#75、docs/re/340）：S3 帝國第一回合
  # 實行之後三張過場各等滿 90 tick（報告在第 354 tick 出現，等 90 tick），掃到達希利那一格先印這一張再進戰場。
  "fcmd-s3-found|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 544 125; click 560 145; click 300 322; click 318 208; click 544 125; click 560 125; tick 400"
  # 艦隊編號清單（#65）：艦隊編成按「艦隊編號」開十六列，點第 3 列 → 標題變第３艦隊。
  "fcmd-fnum-row|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 368 350; click 368 130"
  # 系統指令 → 結束 →「您確定要結束這個遊戲嗎？」（1 號訊息框，中止／決定）。
  "fcmd-quit|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 544 125; click 560 205"
  # 情報收集 → 勢力分頁（#63）：標題與五行圖例整段換掉、地圖照各格自軍編制組數著色。
  "fcmd-info-power|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 352 125; click 400 145; click 424 168"
  "fcmd-staff-decide|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 352 125; click 400 145; click 543 388; click 450 92; click 497 330"
)
# #80：完整正常玩家路徑到徵稅面板，八份都不排除箭頭或游標。
plan_path="move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 352 125; click 400 165"
cases+=(
  "planarrow-tax-up|$plan_path; click 471 289"
  "planarrow-tax-down|$plan_path; click 487 289"
  "planarrow-pct-up|$plan_path; click 471 329"
  "planarrow-pct-down|$plan_path; click 487 329"
  "planarrow-hover|$plan_path; move 471 289"
  "planarrow-moved|$plan_path; click 471 289; move 600 380"
  "planarrow-updown|$plan_path; click 471 289; click 487 289"
  "planarrow-reopen|$plan_path; click 471 289; click 463 369; click 352 125; click 400 165"
)
budget_path="move 300 210; click; type 138; enter; click 432 340"
cases+=("planarrow-budget-base|$budget_path")
for entry in 'tax-up:471:289' 'tax-down:487:289' 'pct-up:471:329' 'pct-down:487:329'; do
  IFS=: read -r key x y <<<"$entry"
  cases+=("planarrow-budget-$key|$budget_path; click $x $y")
  cases+=("planarrow-budget-$key-post|$budget_path; click $x $y; click 463 369; rclick 300 240; rclick 300 240")
done
cases+=("planarrow-budget-unchanged-post|$budget_path; click 463 369; rclick 300 240; rclick 300 240")
fail=0
for c in "${cases[@]}"; do
  IFS='|' read -r batch script except <<<"$c"
  # #85 沿用同一條正常實行輸入，使用獨立重生的原版批次。
  if [[ "$batch" == supply84-gui && "${LOGH3_GUI_PREFIX:-}" == runtime85-normal ]]; then
    batch=runtime85-normal
  fi
  [[ -z "${LOGH3_GUI_PREFIX:-}" || "$batch" == "$LOGH3_GUI_PREFIX"* ]] || continue
  oracle="workplace/oracle/$batch/screen.bin"
  if [[ ! -f "$oracle" ]]; then
    echo "  $batch 缺收據，跳過"
    continue
  fi
  shots=(-shot-bin /src/workplace/gui-check.bin)
  if [[ "$batch" == planarrow-* || "$batch" == supply84-gui || "$batch" == runtime85-normal ]]; then
    shots=(-shot-bin "/src/workplace/oracle/$batch/remake.bin" -shot "/src/workplace/oracle/$batch/remake.png")
    # 記錄實際輸入；不在主機執行資料分析。
    printf '%s\n' "$script" > "workplace/oracle/$batch/remake-script.txt"
    LOGH3_EXTRA_RO="$ROOT" tools/py.sh -c '
import hashlib,json,pathlib,sys
root=pathlib.Path("/ref"); out=pathlib.Path("workplace/oracle",sys.argv[1])
sources=["internal/game/game.go","internal/game/planscreens.go","internal/game/planscreens_test.go","tools/gui_check.sh"]
is_supply=sys.argv[1] in ("supply84-gui","runtime85-normal")
if is_supply: sources += ["internal/game/regionturn.go","internal/rules/regionresupply.go","internal/game/originalbrain.go"]
if sys.argv[1]=="runtime85-normal": sources += ["internal/ai/original_turn.go","internal/game/battle.go","internal/game/runtime_fields_test.go"]
out.joinpath("remake-meta.json").write_text(json.dumps({"git_head":sys.argv[2],"working_tree":sys.argv[3],"image":sys.argv[4],"command":"tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data -script <remake-script.txt> -shot-bin /src/workplace/oracle/<批次>/remake.bin -shot /src/workplace/oracle/<批次>/remake.png -shot-after 2","source_sha256":{p:hashlib.sha256(root.joinpath(p).read_bytes()).hexdigest() for p in sources},"rng":"正常劇本 1 帝國；執行前原版檢查點及 remake strategyRandSeed 固定 0xD3DC167E；比較第一回合完整畫面，不宣稱整段骰序一致" if is_supply else "正常劇本 1 帝國，箭頭不含亂數判定","scenario":"S1 帝國 796-01-04" if is_supply else "S1 帝國 796-01-01","input":"remake-script.txt 完整正常玩家路徑；未注入"},ensure_ascii=False,indent=2)+"\n")
' "$batch" "$(git rev-parse HEAD)" "$(git status --porcelain)" "${LOGH3_GUI_IMAGE:-logh3-ebiten:2}"
  fi
  tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data \
    -script "$script" "${shots[@]}" -shot-after 2 \
    >/dev/null 2>&1
  actual="${shots[1]#/src/}"
  if [[ -n "$except" ]]; then
    # 有排除矩形的用 python 比（x,y,w,h 之外要逐位元組相同）。
    if tools/py.sh tools/scripts/cmp_except.py "$actual" "$oracle" "$except"; then
      echo "  $batch 排除 ($except) 之外逐位元組相同"
    else
      echo "  $batch **不同**"
      fail=1
    fi
  elif tools/py.sh tools/scripts/cmp_except.py "$actual" "$oracle" 0,0,0,0; then
    echo "  $batch 逐位元組相同"
  else
    echo "  $batch **不同**"
    fail=1
  fi
done
exit $fail
