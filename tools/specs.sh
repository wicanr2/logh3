#!/usr/bin/env bash
# 從收據重新量出畫面規格。**規格要能重做**——靠翻對話紀錄拼命令列的東西
# 實際上不會被重做（AGENTS.md 收據那一節同理）。
#
#   tools/specs.sh                 # 全部（有先後：後面幾張拿前面那張當底層）
#   tools/specs.sh cmd-menu        # 只做一張
#
# 做完跑 `tools/go.sh test ./internal/screen/`：那幾張都是 0 / 307,200，
# 對不上就是這裡的參數跟當初量的不一樣。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
only="${1:-}"
py() { "$ROOT/tools/py.sh" "$@"; }
want() { [[ -z "$only" || "$1" == "$only" ]]; }

# 一覽表三張共用同一份清單版面：標題列三個框 ＋ 各自的小指標，
# 其餘 16 條清單列每條左右分兩欄。
LIST_BOXES='432:60:481:79,496:60:545:79,560:60:609:79,444:65:450:76,508:65:514:76,572:65:578:76'
for y in 80 100 120 140 160 180 200 220 240 260 280 300 320 340 360 380; do
  LIST_BOXES="$LIST_BOXES,336:$y:609:$((y+19)),610:$y:623:$((y+19))"
done

# 戰略情報報告：上排三個標題格 ＋ 各自的小指標、頭像下的兩個框、
# 左欄五格數值框。
INFO_BOXES='336:158:385:179,400:158:449:179,464:158:513:179,344:163:350:175,408:163:414:175,472:163:478:175,360:145:373:159,525:378:562:399,573:378:610:399,336:242:369:257,336:262:369:277,336:282:369:297,336:302:369:317,336:322:369:337'

# 遊戲設定：那一排音量／速度條與上下按鈕。每個框 ＝ 填色連通分量往外各擴
# 2 像素（左上淺邊 7、右下暗邊 3），用 tools/scripts/comp_rects.py 量。
SETTINGS_BOXES=\
'350:100:385:119,398:100:433:119,'\
'270:120:433:139,270:140:433:159,270:160:433:179,270:180:433:199,270:200:433:219,'\
'270:220:433:239,270:240:433:259,270:260:433:279,270:280:433:299,'\
'400:300:415:319,416:300:431:319,'\
'350:320:385:339,398:320:433:339,'\
'302:340:337:359,350:340:385:359,398:340:433:359,'\
'318:360:385:379,398:360:433:379'

# 疊在戰略底板上那三張（月初會計報告、船艦訂造報告、次月份船艦建造計劃）
# 裡跟局面走的內容——署名、頭像、金額、首都摘要——不進規格，由
# internal/screen 的 ReportPanel／DeliveryPanel／CapitalPanel 現算
# （`docs/spec/011`）。這一步是冪等的，重跑不會壞。
if want computed; then
  for n in screen-main-game screen-ship-order screen-build-plan screen-info \
           screen-admiral-detail screen-admiral-bio \
           screen-tax-plan screen-production-plan screen-budget-plan \
           screen-admiral-list-view screen-standby-admirals screen-star-list; do
    py tools/scripts/maingame_computed.py "$n"
  done
  # 貨幣單位兩個陣營不同（帝國「億帝國馬克」、同盟「億帝納」），
  # 要從兩張收據各讀一次。
  py tools/scripts/faction_text.py screen-tax-plan overlay 142 tax1e tax1eAlly
  # 命令階段那一批只要拿掉底板右側那幾行（其餘本來就沒量進去）。
  py tools/scripts/board_computed.py \
    screen-cmd-menu screen-list-menu screen-sys-menu screen-standby-admirals \
    screen-info screen-star-list screen-fleet-list screen-admiral-list-view \
    screen-enemy-list screen-load-game screen-save-game screen-settings \
    screen-quit-confirm screen-admiral-list screen-budget-plan
fi

# ⚠ **這三張的「量版面」那一步還沒完全腳本化。** 量的命令是
#
#     py tools/scripts/mainscreen_spec.py key1  screen-main-game
#     py tools/scripts/mainscreen_spec.py ship1 screen-ship-order 448 330 144 80 495 399
#     py tools/scripts/mainscreen_spec.py plan1 screen-build-plan  448 310  80 60 559 399
#     py tools/scripts/box_text.py key1 screen-main-game report 0   # 「變更」
#     py tools/scripts/box_text.py key1 screen-main-game report 1   # 「確認」
#
# 但分頁鈕（`tabs`）、輸入框、熱區與 `fields` 是另外從字元平面
# （`tools/scripts/plane_json.py`，`docs/internals/31`）補進去的，那一步
# 沒有腳本。**照上面重跑會把那幾塊弄丟**，補回來的辦法還沒寫下來。

# 提督詳細資料與提督經歷（`docs/spec/012`）。經歷那一層要另外併進詳細
# 資料的規格——`overlay_spec.py` 產出的視窗一律叫 `overlay`，直接用會把
# 底層那個視窗蓋掉。
if want admiral-detail; then
  py tools/scripts/overlay_spec.py adm1e screen-command-phase screen-admiral-detail \
    96 80 543 359 400 45 "382:320:417:339,430:320:465:339,478:320:513:339" "" "0:448:100" \
    "16:352:320:400" "" "6,5"
  py tools/scripts/box_text.py adm1e screen-admiral-detail overlay 0
  py tools/scripts/box_text.py adm1e screen-admiral-detail overlay 1
  py tools/scripts/box_text.py adm1e screen-admiral-detail overlay 2
  py tools/scripts/maingame_computed.py screen-admiral-detail
fi
if want admiral-bio; then
  py tools/scripts/overlay_spec.py keireki1e screen-admiral-detail screen-admiral-bio \
    80 200 559 359 392 290
  py tools/scripts/bio_spec.py
  py tools/scripts/maingame_computed.py screen-admiral-bio
fi

# 艦隊編成（`docs/spec/013`）。視窗比提督詳細資料矮一截、往右挪了兩格：
# (112,80)–(527,379)。
if want fleet-org; then
  py tools/scripts/overlay_spec.py fleet1e screen-command-phase screen-fleet-org \
    112 80 527 379 488 290 "334:340:401:359,414:340:449:359,462:340:497:359" "" \
    "0:432:96,32:432:196" "" "" "6,5"
  py tools/scripts/box_text.py fleet1e screen-fleet-org overlay 0
  py tools/scripts/box_text.py fleet1e screen-fleet-org overlay 1
  # 「確認」那顆的游標正壓在上面：字讀不全，而且游標的像素會被當成
  # 框的角落修正記進去（在 fleet1e 上看不出來——游標最後會蓋回同一批
  # 像素；疊上艦隊編號那張、游標移走之後就露餡）。改用 `fnum`：
  # 同一顆按鈕、游標在別處。
  py tools/scripts/box_text.py fnum screen-fleet-org overlay 2
  py tools/scripts/maingame_computed.py screen-fleet-org
fi

# 艦隊編成的兩個小視窗（`docs/re/67`）。兩張都被游標咬掉了字，所以各配
# 一份「游標停在同一顆熱區另一端」的收據，兩份取文字聯集。
# 反白條不進規格（換一個滑鼠位置就換一列），改由 `hover` 描述版面。
if want fleet-number; then
  py tools/scripts/overlay_spec.py fnum screen-fleet-org screen-fleet-number \
    320 60 415 418 368 310 "" "" "" "" fnum3
  py tools/scripts/stack_spec.py screen-fleet-org screen-fleet-number number
  py tools/scripts/hover_block.py screen-fleet-number number 336 80 64 20 20 16 2
fi
if want fleet-adjutant; then
  py tools/scripts/overlay_spec.py fadj screen-fleet-org screen-fleet-adjutant \
    400 320 463 398 432 310 "" "" "" "" fadj3
  py tools/scripts/stack_spec.py screen-fleet-org screen-fleet-adjutant menu
  py tools/scripts/hover_block.py screen-fleet-adjutant menu 416 340 32 20 20 2 2
fi

# 艦隊組成（`docs/spec/015`）。格盤那一塊要當成小框傳進去才能排除在文字
# 掃描之外，量完再換回純色矩形；七個艦種列的第 0 列內部是**選取**不是底色。
if want fleet-composition; then
  BOXES='112:140:303:339'
  for y in 160 180 200 220 240 260 280; do BOXES="$BOXES,334:$y:401:$((y+19))"; done
  for y in 300 320 340; do BOXES="$BOXES,494:$y:529:$((y+19))"; done
  py tools/scripts/overlay_spec.py fok screen-fleet-org screen-fleet-composition \
    80 100 559 379 480 310 "$BOXES"
  for i in 1 2 3 4 5 6 7 8 9 10; do
    py tools/scripts/box_text.py fok screen-fleet-composition overlay $i
  done
  py tools/scripts/composition_blocks.py
  py tools/scripts/maingame_computed.py screen-fleet-composition
fi

# 資料段的艦種表與開局各星域的艦艇艘數（要一份記憶體 dump，見腳本說明）。
# tools/oracle.sh fokall … -dump-mem lin:0:1048576:/src/workplace/fokall.bin
# py tools/scripts/shiptype_export.py /work/workplace/explore/fleetorg/fokall.bin 0
# py tools/scripts/region_ships_export.py /work/workplace/explore/fleetorg/fokall.bin 0

# 提督詳細資料的「旗艦」對照表（資料段的兩張 far 指標表）。
want flagships && py tools/scripts/flagship_export.py

# 底層：命令階段主畫面。其餘幾張都疊在它上面。
want command-phase && py tools/scripts/cmdscreen_spec.py main1e screen-command-phase

# 戰略命令選單。游標給 -1 ＝ 沿用底層那一份（滑鼠沒動）。
want cmd-menu && py tools/scripts/overlay_spec.py \
  cmd1e screen-command-phase screen-cmd-menu 336 100 431 219 -1 -1

# 戰略情報報告。游標給 -2 ＝ 這一層把它整個蓋掉了。
# 地圖跟著**治安**換色，不是勢力色——那個欄位 overlay_spec 不會寫，補在後面。
if want info; then
  py tools/scripts/overlay_spec.py \
    info1e screen-command-phase screen-info 320 40 639 419 -2 -2 \
    "$INFO_BOXES" '368:160:3:1:3' '13:544:60'
  py -c 'import json,collections
p="/work/docs/internals/data/screen-info.json"
s=json.load(open(p),object_pairs_hook=collections.OrderedDict)
s["map"]["colours"]="colour_796_01_security"
json.dump(s,open(p,"w"),ensure_ascii=False,indent=1)'
fi

# 一覽表：艦隊／提督／敵將。差別只有選中那一列的小指標位置，
# 以及提督資料會重寫左下角那塊（額外文字區）。
want fleet-list && py tools/scripts/overlay_spec.py \
  lst1e screen-command-phase screen-fleet-list 320 40 639 419 -2 -2 \
  "$LIST_BOXES" '468:158:3:1:7,469:159:1:1:7,468:160:3:1:3'

want admiral-list-view && py tools/scripts/overlay_spec.py \
  lst2e screen-command-phase screen-admiral-list-view 320 40 639 419 -2 -2 \
  "$LIST_BOXES" '468:178:3:1:7,469:179:1:1:7,468:180:3:1:3' '' '8:350:320:400'

want enemy-list && py tools/scripts/overlay_spec.py \
  lst3e screen-command-phase screen-enemy-list 320 40 639 419 -2 -2 \
  "$LIST_BOXES" '468:198:3:1:7,469:199:1:1:7,468:200:3:1:3'

# 徵稅計劃 → 次月份預算計劃。視窗版面與 screen-budget-plan 同一個，
# 只是這一次疊在命令階段的底板上。
want tax-plan && py tools/scripts/overlay_spec.py \
  tax1e screen-command-phase screen-tax-plan 128 60 511 399 400 125 \
  '464:280:479:299,480:280:495:299,464:320:479:339,480:320:495:339,398:360:433:379,446:360:481:379' \
  '144:239:256:1:15,144:279:256:1:15' '16:416:80'

# 生產計劃 → 次月份船艦建造計劃。這張的游標壓在標題列上，被蓋住那一格
# 從第二份收據（prod1i，點在選單同一項的左邊）補回來。
want production-plan && py tools/scripts/overlay_spec.py \
  prod1e screen-command-phase screen-production-plan 80 60 559 399 400 145 \
  '208:200:223:219,224:200:239:219,208:220:223:239,224:220:239:239,208:240:223:259,224:240:239:259,208:260:223:279,224:260:239:279,208:280:223:299,224:280:239:299,208:300:223:319,224:300:239:319,208:320:223:339,224:320:239:339,446:360:481:379,494:360:529:379' \
  '' '6:464:80' '' prod1i

# 兩張下拉選單本身（戰略命令那一張是 screen-cmd-menu）。
want sys-menu && py tools/scripts/overlay_spec.py \
  sysmenu1e screen-command-phase screen-sys-menu 528 100 607 239 -1 -1
want list-menu && py tools/scripts/overlay_spec.py \
  lstmenu1e screen-command-phase screen-list-menu 432 100 527 219 -1 -1

# 一覽表 → 星系資料。版面與另外三張清單相同，差別有兩處：
# 標題格裡各有一個 11×11 的小標記（選中的那個填紅色 2），
# 以及**地圖用另一組星域色**——首都不重塗成色 6，選中的星域畫成色 7。
STAR_BOXES='432:60:481:79,496:60:545:79,560:60:609:79,436:65:446:75,500:65:510:75,564:65:574:75'
for y in 80 100 120 140 160 180 200 220 240 260 280 300 320 340 360 380; do
  STAR_BOXES="$STAR_BOXES,336:$y:609:$((y+19)),610:$y:623:$((y+19))"
done
if want star-list; then
  py tools/scripts/overlay_spec.py \
    star1e screen-command-phase screen-star-list 320 40 639 419 460 85 \
    "$STAR_BOXES" '' '' '8:350:320:400' '' 8,6,5
  py -c 'import json,collections
p="/work/docs/internals/data/screen-star-list.json"
s=json.load(open(p),object_pairs_hook=collections.OrderedDict)
s["map"]["colours"]="colour_796_01_star_list"
json.dump(s,open(p,"w"),ensure_ascii=False,indent=1)'
fi

# 出擊命令 → 待命中提督資料。清單十七列 ＋ 右側捲軸（上下箭頭各一格）。
# 左下角那塊會被重寫（星系名、治安、行星、艦隊），走額外文字區；
# 那塊的數字是色 6 與 5，標籤是色 7。
STANDBY_BOXES='610:80:623:99,610:100:623:379,610:380:623:399'
for y in 80 100 120 140 160 180 200 220 240 260 280 300 320 340 360 380; do
  STANDBY_BOXES="$STANDBY_BOXES,336:$y:609:$((y+19))"
done
want standby-admirals && py tools/scripts/overlay_spec.py \
  sortie1e screen-command-phase screen-standby-admirals 320 40 639 419 400 85 \
  "$STANDBY_BOXES" '' '' '8:350:320:400' '' 8,6,5

# 取進度：8 個未使用的欄位（色 8）＋ 3 個劇本。未使用那幾列的顏色與
# 一般文字不同，第 15 個參數就是「除了色 7 還要掃哪些顏色」。
want load-game && py tools/scripts/overlay_spec.py \
  load1e screen-command-phase screen-load-game 32 100 607 359 560 105 \
  '' '' '' '' '' 8

# 存進度：滑鼠停在第二列，那一列被刷成紅色（色 2）的反白條，
# 而刷的那一下把游標蓋掉了——所以反白條要在游標之後再畫一次。
want save-game && py tools/scripts/overlay_spec.py \
  save1e screen-command-phase screen-save-game 32 120 607 319 560 125 \
  '' '48:160:544:20:2' '' '' '' 8

# 結束：確認對話框，兩個按鈕。
want quit-confirm && py tools/scripts/overlay_spec.py \
  quit1e screen-command-phase screen-quit-confirm 128 180 479 299 560 165 \
  '366:260:401:279,414:260:449:279'

# 遊戲設定：底層是戰略命令選單那一份（設定視窗把選單整個蓋住）。
# 游標停在最後一次點擊的位置，要明寫。
want settings && py tools/scripts/overlay_spec.py \
  settings1e screen-cmd-menu screen-settings 160 41 463 399 560 145 "$SETTINGS_BOXES"

# 月底的船艦接收報告。底層是命令階段主畫面，但右側面板的字整塊被清掉、
# 只留「戰略模式處理中」那一行——所以要多量一塊額外文字區。
# 「確認」鈕的字要走 box_text.py：框裡有字的時候 measure_frame 會把字的
# 像素量成一圈一圈的立體邊（十層），而 frame_patches 只補角落，
# 中間那一塊就變成一片交錯的雜訊。
if want ship-delivery; then
  py tools/scripts/overlay_spec.py \
    turnmonth2 screen-command-phase screen-ship-delivery 144 80 495 399 560 85 \
    '430:360:465:379' '' '6:400:100' '320:40:639:79'
  py tools/scripts/box_text.py turnmonth2 screen-ship-delivery overlay 0
fi

exit 0
