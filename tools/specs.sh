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
