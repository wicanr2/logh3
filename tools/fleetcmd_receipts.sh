#!/usr/bin/env bash
# 戰略艦隊命令那一串畫面的收據（`docs/re/312`、`docs/spec/051`）。
#
#   tools/fleetcmd_receipts.sh            # 全部重錄（約 2 分鐘）
#
# 一條檢查點鏈：冷啟動走 `dec1` 的序列（主選單 → 防拷 → 出擊命令 → 編成 →
# 決定）到戰略地圖介面存一份狀態，之後每一張畫面都從上一張的檢查點展開，
# 每一步只多一個輸入。批次名一律 `fcmd-*`，探索用的 `fcx-*` 不在這裡。
#
# ⚠ 收據綁執行器版本：dosgolem 一改就整條重跑。每個批次的 meta.txt 由
# `tools/oracle.sh` 寫入執行器 HEAD 與旗標；檢查點本身不進版控。
#
# 座標是**滑鼠座標**（640×400，畫面 y ＝ 滑鼠 y ＋ 40，`docs/internals/31`）。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
CK="$GOLEM/workplace/fcmd"          # 檢查點（gitignore）
mkdir -p "$CK"
DS=54B3
PLANE_LO=$(printf '%X' $((0x${DS}*16 + 0xD81E)))   # 字元平面 40×20×6
PLANE_HI=$(printf '%X' $((0x${DS}*16 + 0xD81E + 4800)))
FLEET_LO=$(printf '%X' $((0x${DS}*16 + 0xC074)))   # 艦隊表第 0 筆 122 bytes
FLEET_HI=$(printf '%X' $((0x${DS}*16 + 0xC074 + 122)))
HOLD=3000000
STOCK="-peek lin:5F2FC:28"      # 星域 0 帝國七種存量（第 13、14 段）
ADM="-peek lin:602AC:28,lin:602E4:28,lin:6062C:28"   # 提督記錄 0／2／32（第 15 段）
HOLD5=300000                    # ▲▼ 按住 0.3M 步 ＝ 五列（第 17 段）
HOLD1=100000                    # 一覽表的 ▲▼ 按住 0.1M 步 ＝ 一列（第 20 段）
TABS="-peek lin:5DA4F:1,lin:5E0C2:2,lin:5E0E4:2,lin:5E108:2,lin:61B2C:1,lin:5E106:1"   # 分頁編號與四個分頁的筆數／位移（第 17 段）
LST="-peek lin:5E0B8:1,lin:5E216:1,lin:5E21C:1,lin:5E1F2:2,lin:5DA5D:2,lin:56922:2,lin:5DA61:2,lin:5E1D0:2,lin:63CB2:2"   # 一覽表四張的分頁／位移／筆數（第 20 段）
DIS="-peek lin:60BA4:122,lin:5F2CC:4032,lin:602AC:28,lin:6062C:28,lin:5DA79:2,lin:61AE4:4,lin:668B2:78,lin:66900:78"   # 駐留解散：艦隊 0、星域 32 筆、提督 0／32、ds:8F49、ds:CFB4（668B:0002）、艦艇 0／1（第 21 段）
OCC="-peek lin:60BA4:122,lin:5F2CC:4032,lin:53132:6642"   # 占領與打行星：艦隊 0、星域 32 筆、行星 81 筆（第 19、22 段）
TURN="-peek lin:60BA4:122,lin:5DA82:8,lin:5DA4C:12"   # 回合過場：艦隊 0、ds:8F52 兩陣營軍事費、ds:8F1C–8F27 座位旗標（第 23 段）
EXEC="-clicks 546000000:544:85,549000000:560:85 -click-hold 3000000"   # 系統指令、實行（第 23 段）
S3EXEC="-clicks 301000000:544:85,304000000:560:85 -click-hold 3000000"   # S3 帝國第一回合的系統指令、實行（第 25 段）

# rec <批次> <步數上限> <上一個檢查點|-> <ck|-> <旗標...>
# 每一批都倒：畫面、調色盤、字元平面、艦隊表第 0 筆。`ck` ＝ 在上限的前一道
# 指令存檢查點（存檔點與 -steps 相同時迴圈已經結束，一個檔都不會寫）。
rec() {
  local name=$1 steps=$2 from=$3 save=$4; shift 4
  local extra=()
  [[ "$from" != - ]] && extra+=(-load-state "/src/workplace/fcmd/$from.state")
  [[ "$save" == ck ]] && extra+=(-save-state "$((steps - 1)):/src/workplace/fcmd/$name.state")
  tools/oracle.sh "$name" "${extra[@]}" -steps "$steps" "$@" \
    -dump-mem "$PLANE_LO-$PLANE_HI:/src/workplace/fcmd/$name-plane.bin,$FLEET_LO-$FLEET_HI:/src/workplace/fcmd/$name-fleet0.bin" \
    >/dev/null
  cp "$CK/$name-plane.bin" "$ROOT/workplace/oracle/$name/plane.bin"
  cp "$CK/$name-fleet0.bin" "$ROOT/workplace/oracle/$name/fleet0.bin"
  printf '%s  %s\n' "$name" "$(sed -n 's/^screen.bin SHA-256: //p' "$ROOT/workplace/oracle/$name/meta.txt" | cut -c1-16)"
}

# recmem <批次> <進入步數> <離開步數> <rec 的其餘參數…>：同 rec，另外在兩個步數各倒一份 1 MB 記憶體
# （線性 0 起），複製成批次目錄的 mem-in.bin／mem-out.bin（第 22 段：原版進出 `20B4:1497` 的那一刻）。
recmem() {
  local name=$1 at_in=$2 at_out=$3; shift 3
  rec "$name" "$@" -dump-mem-at "$at_in:lin:0:1048576:/src/workplace/fcmd/$name-in.bin;$at_out:lin:0:1048576:/src/workplace/fcmd/$name-out.bin"
  cp "$CK/$name-in.bin" "$ROOT/workplace/oracle/$name/mem-in.bin"
  cp "$CK/$name-out.bin" "$ROOT/workplace/oracle/$name/mem-out.bin"
}

# mvc／mvr <步數> <x> <y>：先把游標移過去，10M 步之後才按（第 20 段的坑：移動與按鍵同一瞬間會被丟掉）。
mvc() { echo "-mouse-x $2 -mouse-y $3 -mouse-at $1 -clicks $(($1 + 10000000)):$2:$3"; }
mvr() { echo "-mouse-x $2 -mouse-y $3 -mouse-at $1 -rclicks $(($1 + 10000000)):$2:$3"; }
cd "$ROOT"
# 0. 戰略地圖介面，右下角清單有剛編成的第１艦隊（＝ 收據 `dec1` 的序列）。
rec fcmd-base 420000000 - ck \
  -tick 20000 -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -rclicks 130000000:300:200,155000000:300:200 \
  -clicks 40000000:300:170,95000000:300:170,180000000:352:85,210000000:400:85,240000000:400:45,270000000:488:290,320000000:480:310,380000000:510:310 \
  -click-hold $HOLD
# 1. 點清單第 0 列 → 艦隊視窗。
rec fcmd-win     445000000 fcmd-base ck -clicks 421000000:450:268 -click-hold $HOLD
# 2. 點「作戰」→ 五項選單（游標停在第三項「駐留」上，所以它反白）。
rec fcmd-menu    470000000 fcmd-win  ck -clicks 447000000:490:305 -click-hold $HOLD
# 2a. 選單開著，游標移到第一項「移動」（不按）→ 反白跟著游標。
rec fcmd-menu-hover 480000000 fcmd-menu - -mouse-x 497 -mouse-y 272 -mouse-at 472000000
# 3. 點「移動」→ 目的地選取：地圖露出來、依可達性重新著色。
rec fcmd-pick    495000000 fcmd-menu ck -clicks 472000000:497:272 -click-hold $HOLD
# 4. 點亞爾提那（星域 2）→ 確認框「要移動至帝國軍亞爾提那星系嗎？」。
rec fcmd-confirm 520000000 fcmd-pick ck -clicks 497000000:85:88 -click-hold $HOLD
# 5. 點「決定」→ 回地圖，艦隊記錄 +0x07=2、+0x08=0，清單列多一個 ☆（Big5 A1B8、字色 0）。
rec fcmd-done    545000000 fcmd-confirm ck -clicks 522000000:432:223 -click-hold $HOLD
# 5a. 下了命令之後再點清單第 0 列 → 視窗第 5–7 列印移動目標（+0x08 != 5 那一種）。
rec fcmd-reopen  570000000 fcmd-done ck -clicks 547000000:450:268 -click-hold $HOLD
# 5b. 下了命令之後按 系統指令 → 實行：回合處理有沒有出發／到達報告（113A:0B46）。
rec fcmd-exec    640000000 fcmd-done - -clicks 547000000:544:85,572000000:560:85 -click-hold $HOLD
# 5'. 點「中止」→ 回目的地選取（艦隊記錄不變）。
rec fcmd-cancel  545000000 fcmd-confirm - -clicks 522000000:383:223 -click-hold $HOLD
# 4'. 目的地點巴拉特（星域 31，走不到）→「這是無法前往的星系。」。
rec fcmd-far     520000000 fcmd-pick - -clicks 497000000:300:270 -click-hold $HOLD
# 4''. 目的地點自己所在的瓦爾哈拉 →「要待命於…嗎？」（選現在位置 ＝ 取消命令）。
rec fcmd-self    520000000 fcmd-pick ck -clicks 497000000:32:32 -click-hold $HOLD
# 4'''. 那一框按「決定」→ 回地圖：+0x07 仍是 FF、+0x08 = 5，清單列**沒有** ☆。
rec fcmd-self-done 545000000 fcmd-self - -clicks 522000000:432:223 -click-hold $HOLD
# 右鍵：目的地選取 → 回艦隊視窗；選單 → 關選單；艦隊視窗 → 回地圖。
rec fcmd-pick-rclick 520000000 fcmd-pick - -rclicks 497000000:300:200 -click-hold $HOLD
rec fcmd-menu-rclick 495000000 fcmd-menu - -rclicks 472000000:300:200 -click-hold $HOLD
rec fcmd-win-rclick  470000000 fcmd-win  - -rclicks 447000000:300:200 -click-hold $HOLD
# 「結束」→ 回地圖。
rec fcmd-close   470000000 fcmd-win - -clicks 447000000:490:350 -click-hold $HOLD
# 7. 作戰選單第 2 項「駐留」（滑鼠 y 312）與第 3 項「輸送」（y 332）：與移動同一個
#    目的地迴圈，各走 目的地 → 亞爾提那 → 決定，看 +0x08 ＝ 2／3 與動詞。
rec fcmd-garrison-pick    495000000 fcmd-menu ck -clicks 472000000:497:312 -click-hold $HOLD
rec fcmd-garrison-confirm 520000000 fcmd-garrison-pick ck -clicks 497000000:85:88 -click-hold $HOLD
rec fcmd-garrison-done    545000000 fcmd-garrison-confirm ck -clicks 522000000:432:223 -click-hold $HOLD
rec fcmd-transport-pick    495000000 fcmd-menu ck -clicks 472000000:497:332 -click-hold $HOLD
rec fcmd-transport-confirm 520000000 fcmd-transport-pick ck -clicks 497000000:85:88 -click-hold $HOLD
rec fcmd-transport-done    545000000 fcmd-transport-confirm - -clicks 522000000:432:223 -click-hold $HOLD
# 8. 作戰選單第 4 項「待命」（滑鼠 y 352）→「指定待命行星」（docs/re/314）：
#    點第 1 顆（布倫希爾，鈕列 1 ＝ 畫面 y 240–259 → 滑鼠 210）→ 決定 (400,310)／中止 (350,310)／右鍵；
#    決定之後再開視窗看「待命行星」那一列。
rec fcmd-standby        495000000 fcmd-menu ck -clicks 472000000:497:352 -click-hold $HOLD
rec fcmd-standby-pick   520000000 fcmd-standby ck -clicks 497000000:350:210 -click-hold $HOLD
rec fcmd-standby-done   545000000 fcmd-standby-pick ck -clicks 522000000:400:310 -click-hold $HOLD
rec fcmd-standby-cancel 545000000 fcmd-standby-pick - -clicks 522000000:350:310 -click-hold $HOLD
rec fcmd-standby-rclick 545000000 fcmd-standby-pick - -rclicks 522000000:300:200 -click-hold $HOLD
rec fcmd-standby-reopen 570000000 fcmd-standby-done - -clicks 547000000:450:268 -click-hold $HOLD
# 9. 作戰選單第 1 項「占領」（滑鼠 y 292）→ 目的地選取（允許的是敵方與中立）：
#    亞爾提那（自己的）→「無法前往」；亞斯提（同盟，16）→ 確認框 → 決定 → 再開視窗；
#    費沙（中立，14）→ 參謀警告兩頁（1213:0003；每頁「訊息速度×2 ＝ 180」個慢 tick 沒點就
#    自動翻頁，這條鏈量到是 36M 步——探索批次 fcx-occ-phz2：頁 1 在 497.2M、頁 2 在 533.2M；
#    25M 一批、點了之後 23M 就截，來得及）→ 確認框 → 決定。
rec fcmd-occupy-pick    495000000 fcmd-menu ck -clicks 472000000:497:292 -click-hold $HOLD
rec fcmd-occupy-far     520000000 fcmd-occupy-pick - -clicks 497000000:85:88 -click-hold $HOLD
rec fcmd-occupy-confirm 520000000 fcmd-occupy-pick ck -clicks 497000000:193:205 -click-hold $HOLD
rec fcmd-occupy-done    545000000 fcmd-occupy-confirm ck -clicks 522000000:432:223 -click-hold $HOLD
rec fcmd-occupy-reopen  570000000 fcmd-occupy-done - -clicks 547000000:450:268 -click-hold $HOLD
rec fcmd-occupy-phezzan  520000000 fcmd-occupy-pick ck -clicks 497000000:163:96 -click-hold $HOLD
rec fcmd-occupy-phezzan2 545000000 fcmd-occupy-phezzan ck -clicks 522000000:300:200 -click-hold $HOLD
rec fcmd-occupy-phezzan3 570000000 fcmd-occupy-phezzan2 ck -clicks 547000000:300:200 -click-hold $HOLD
rec fcmd-occupy-phezzan-done 595000000 fcmd-occupy-phezzan3 - -clicks 572000000:432:223 -click-hold $HOLD
# 10. 作戰選單五項的停用旗標（docs/re/312 §7.6、docs/re/315）：
#    (a) 首都「有戰鬥」——在 170M（戰略地圖已畫好、還沒點戰略命令）把星域 0 的 +0x13
#        （線性 5F2DF）改成 04，其餘照 fcmd-base 的序列走到地圖：清單那一列 ▼、
#        左下角「處於戰鬥狀態」；開視窗（作戰 戰鬥中）→ 作戰 → 項 1–4 停用（色 0x98）。
#        游標停在第三項（停用）不畫反白、移到第一項才畫；點停用的那一項 → 看收據。
rec fcmd-battle-base 420000000 - ck \
  -tick 20000 -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -rclicks 130000000:300:200,155000000:300:200 \
  -clicks 40000000:300:170,95000000:300:170,180000000:352:85,210000000:400:85,240000000:400:45,270000000:488:290,320000000:480:310,380000000:510:310 \
  -click-hold $HOLD -poke "lin:5F2DF@170000000=04"
rec fcmd-battle-win        445000000 fcmd-battle-base ck -clicks 421000000:450:268 -click-hold $HOLD
rec fcmd-battle-menu       470000000 fcmd-battle-win ck -clicks 447000000:490:305 -click-hold $HOLD
rec fcmd-battle-menu-hover 480000000 fcmd-battle-menu - -mouse-x 497 -mouse-y 272 -mouse-at 472000000
rec fcmd-battle-menu-click 495000000 fcmd-battle-menu - -clicks 472000000:497:312 -click-hold $HOLD
#    (b) 艦隊 +0x00 bit0——從 fcmd-win 在 446M 把艦隊記錄第 0 筆的 +0x00（線性 60BA4）改成 01
#        再點作戰：占領（1）與輸送（3）停用；點停用的「占領」→ 看收據。
rec fcmd-flag1-menu       470000000 fcmd-win ck -poke "lin:60BA4@446000000=01" -clicks 447000000:490:305 -click-hold $HOLD
rec fcmd-flag1-menu-click 495000000 fcmd-flag1-menu - -clicks 472000000:497:292 -click-hold $HOLD
# 11. 艦隊視窗「變更」(490,325) → 副官／編組 兩項小選單（0DEA:17A9，docs/re/316）：
#     副官 (497,332) → 艦隊編成那張視窗的「副官／結束」變體（297B:04EF），結束 (480,300) 回視窗；
#     編組 (497,352) → 艦隊組成（0FF4:000A），中止 (510,280)／決定 (510,300) 回視窗。
rec fcmd-change-menu       470000000 fcmd-win ck -clicks 447000000:490:325 -click-hold $HOLD
rec fcmd-change-menu-hover 480000000 fcmd-change-menu - -mouse-x 497 -mouse-y 352 -mouse-at 472000000
rec fcmd-change-menu-rclick 495000000 fcmd-change-menu - -rclicks 472000000:300:200 -click-hold $HOLD
rec fcmd-change-adj        495000000 fcmd-change-menu ck -clicks 472000000:497:332 -click-hold $HOLD
rec fcmd-change-adj-close  520000000 fcmd-change-adj - -clicks 497000000:480:300 -click-hold $HOLD
rec fcmd-change-org        495000000 fcmd-change-menu ck -clicks 472000000:497:352 -click-hold $HOLD
rec fcmd-change-org-cancel 520000000 fcmd-change-org - -clicks 497000000:510:280 -click-hold $HOLD
rec fcmd-change-org-decide 520000000 fcmd-change-org - -clicks 497000000:510:300 -click-hold $HOLD
rec fcmd-change-adj-rclick 520000000 fcmd-change-adj - -rclicks 497000000:300:200 -click-hold $HOLD
rec fcmd-change-org-rclick 520000000 fcmd-change-org - -rclicks 497000000:300:200 -click-hold $HOLD
# 12. 艦隊組成（出擊命令那條路的最後一張）按「中止」或右鍵（docs/re/316 §四）：
#     中止 → 整支艦隊作廢、回戰略地圖，首都的待命艘數退回 78000（開局值；進艦隊組成時
#     先扣了旗艦那一組 3000）；右鍵 → 什麼都不做。與 fcmd-base 同一條冷啟動序列。
rec fcmd-fok-cancel 400000000 - - \
  -tick 20000 -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -rclicks 130000000:300:200,155000000:300:200 \
  -clicks 40000000:300:170,95000000:300:170,180000000:352:85,210000000:400:85,240000000:400:45,270000000:488:290,320000000:480:310,352000000:510:280 \
  -click-hold $HOLD -peek lin:5F2FC:4
rec fcmd-fok-rclick 380000000 - - \
  -tick 20000 -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -rclicks 130000000:300:200,155000000:300:200,352000000:300:200 \
  -clicks 40000000:300:170,95000000:300:170,180000000:352:85,210000000:400:85,240000000:400:45,270000000:488:290,320000000:480:310 \
  -click-hold $HOLD -peek lin:5F2FC:4
# 13. 艦隊組成編輯器本體（0FF4:000A，docs/re/317）。座標（滑鼠）：格盤一格 16 px，
#     隊形槽 (欄,列) 的中心在 (144＋16欄, 136＋16列)——槽 19＝(2,4)→(176,200)、
#     槽 20 中央＝(4,4)→(208,200)、槽 21＝(6,4)→(240,200)、槽 0＝(0,0)→(144,136)；
#     艦種列 i 在 (360, 130＋20i)；陣形 (510,270)、中止 (510,280)、決定 (510,300)；
#     訊息框「確認」(432,223)；陣形選單第 i 項 (505, 210＋20i)。
#     每一批多倒星域 0 帝國七種存量（$STOCK ＝ -peek lin:5F2FC:28，值在 report.txt）。
#     從 fcmd-change-org（既有艦隊、只有旗艦在中央、應有 8 組、戰艦列反白）展開：
rec fcmd-edit-place1      520000000 fcmd-change-org    ck -clicks 497000000:176:200 -click-hold $HOLD $STOCK
rec fcmd-edit-row2        545000000 fcmd-edit-place1   ck -clicks 522000000:360:170 -click-hold $HOLD $STOCK
rec fcmd-edit-place2      570000000 fcmd-edit-row2     ck -clicks 547000000:240:200 -click-hold $HOLD $STOCK
rec fcmd-edit-replace     595000000 fcmd-edit-place2   -  -clicks 572000000:176:200 -click-hold $HOLD $STOCK
rec fcmd-edit-remove      595000000 fcmd-edit-place2   ck -rclicks 572000000:176:200 -click-hold $HOLD $STOCK
rec fcmd-edit-centre-msg  595000000 fcmd-edit-place2   ck -clicks 572000000:208:200 -click-hold $HOLD $STOCK
rec fcmd-edit-centre-ok   620000000 fcmd-edit-centre-msg - -clicks 597000000:432:223 -click-hold $HOLD $STOCK
rec fcmd-edit-flag-rclick 595000000 fcmd-edit-place2   -  -rclicks 572000000:208:200 -click-hold $HOLD $STOCK
rec fcmd-edit-cancel      595000000 fcmd-edit-place2   -  -clicks 572000000:510:280 -click-hold $HOLD $STOCK
rec fcmd-edit-decide      595000000 fcmd-edit-place2   ck -clicks 572000000:510:300 -click-hold $HOLD $STOCK
rec fcmd-edit-formation   620000000 fcmd-edit-remove   ck -clicks 597000000:510:270 -click-hold $HOLD $STOCK
rec fcmd-edit-form-hover  630000000 fcmd-edit-formation - -mouse-x 505 -mouse-y 350 -mouse-at 622000000
rec fcmd-edit-form-rclick 645000000 fcmd-edit-formation - -rclicks 622000000:300:200 -click-hold $HOLD $STOCK
rec fcmd-edit-form-pick   645000000 fcmd-edit-formation ck -clicks 622000000:505:210 -click-hold $HOLD $STOCK
rec fcmd-edit-full-msg    670000000 fcmd-edit-form-pick -  -clicks 647000000:144:136 -click-hold $HOLD $STOCK
rec fcmd-edit-form-decide 670000000 fcmd-edit-form-pick ck -clicks 647000000:510:300 -click-hold $HOLD $STOCK
#     運補艦存量改成 0 再開編組（docs/spec/014：要狀態就直接改記憶體）：
#     選運補艦列再放 → 「船艦數量不足。」；陣形 → 圓錐時 (5,29) 那一組換成替代表的第一個有貨的艦種。
rec fcmd-edit-poor        495000000 fcmd-change-menu   ck -poke lin:5F310@471000000=00,00,00,00 -clicks 472000000:497:352 -click-hold $HOLD $STOCK
rec fcmd-edit-poor-row5   520000000 fcmd-edit-poor     ck -clicks 497000000:360:230 -click-hold $HOLD $STOCK
rec fcmd-edit-poor-msg    545000000 fcmd-edit-poor-row5 - -clicks 522000000:176:200 -click-hold $HOLD $STOCK
rec fcmd-edit-poor-form   520000000 fcmd-edit-poor     ck -clicks 497000000:510:270 -click-hold $HOLD $STOCK
rec fcmd-edit-poor-pick   545000000 fcmd-edit-poor-form - -clicks 522000000:505:210 -click-hold $HOLD $STOCK
# 14. 出擊命令那條路的同一支編輯器（fcmd-base 的冷啟動序列跑到 350M：艦隊編成按了決定、
#     艦隊組成剛開）。同一組操作，證明兩條路共用同一套規則；決定回戰略地圖。
rec fcmd-fok 350000000 - ck \
  -tick 20000 -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -rclicks 130000000:300:200,155000000:300:200 \
  -clicks 40000000:300:170,95000000:300:170,180000000:352:85,210000000:400:85,240000000:400:45,270000000:488:290,320000000:480:310 \
  -click-hold $HOLD $STOCK
rec fcmd-fok-place1       375000000 fcmd-fok           ck -clicks 352000000:176:200 -click-hold $HOLD $STOCK
rec fcmd-fok-row2         400000000 fcmd-fok-place1    ck -clicks 377000000:360:170 -click-hold $HOLD $STOCK
rec fcmd-fok-place2       425000000 fcmd-fok-row2      ck -clicks 402000000:240:200 -click-hold $HOLD $STOCK
rec fcmd-fok-decide       450000000 fcmd-fok-place2    -  -clicks 427000000:510:300 -click-hold $HOLD $STOCK
rec fcmd-fok-form         425000000 fcmd-fok-row2      ck -clicks 402000000:510:270 -click-hold $HOLD $STOCK
rec fcmd-fok-form-pick    450000000 fcmd-fok-form      -  -clicks 427000000:505:210 -click-hold $HOLD $STOCK
# 15. 「變更 → 副官」那張的「副官」鈕 (430,300) → 變更／卸任 小選單（07F9:0A80，框在格 (25,14)，
#     變更 (430,310)、卸任 (430,330)）：卸任 → 副官那幾列清掉；變更 → 「待命中副官資料」清單
#     （0C0A:000F(星域, 0x65, 等級)，列 i 在 (450, 62＋20i)）→ 點一列 → 提督詳細資料 → 決定／中止。
#     每批多倒提督記錄 0（萊因哈特）、2（米達麥亞）、32（修特萊）各 28 bytes（$ADM）。
rec fcmd-adj-menu          520000000 fcmd-change-adj ck -clicks 497000000:430:300 -click-hold $HOLD $ADM
rec fcmd-adj-menu-hover    530000000 fcmd-adj-menu   -  -mouse-x 430 -mouse-y 330 -mouse-at 522000000
rec fcmd-adj-menu-rclick   545000000 fcmd-adj-menu   -  -rclicks 522000000:300:200 -click-hold $HOLD $ADM
rec fcmd-adj-dismiss       545000000 fcmd-adj-menu   ck -clicks 522000000:430:330 -click-hold $HOLD $ADM
rec fcmd-adj-dismiss-close 570000000 fcmd-adj-dismiss - -clicks 547000000:480:300 -click-hold $HOLD $ADM
rec fcmd-adj-change        545000000 fcmd-adj-menu   ck -clicks 522000000:430:310 -click-hold $HOLD $ADM
rec fcmd-adj-change-hover  555000000 fcmd-adj-change -  -mouse-x 450 -mouse-y 150 -mouse-at 547000000
rec fcmd-adj-change-rclick 570000000 fcmd-adj-change -  -rclicks 547000000:300:200 -click-hold $HOLD $ADM
rec fcmd-adj-change-pick   570000000 fcmd-adj-change ck -clicks 547000000:450:70 -click-hold $HOLD $ADM
#     清單右緣的 ▲ (617,50)／▼ (617,350) 一次捲一頁（十六列），到底就停在最後十六列：
rec fcmd-adj-change-down   570000000 fcmd-adj-change ck -clicks 547000000:617:350 -click-hold $HOLD $ADM
rec fcmd-adj-change-down2  595000000 fcmd-adj-change-down - -clicks 572000000:617:350 -click-hold $HOLD $ADM
rec fcmd-adj-change-up     595000000 fcmd-adj-change-down - -clicks 572000000:617:50 -click-hold $HOLD $ADM
rec fcmd-adj-change-pick2  595000000 fcmd-adj-change-down - -clicks 572000000:450:70 -click-hold $HOLD $ADM
rec fcmd-adj-pick-rclick   595000000 fcmd-adj-change-pick - -rclicks 572000000:300:200 -click-hold $HOLD $ADM
rec fcmd-adj-pick-cancel   595000000 fcmd-adj-change-pick - -clicks 572000000:448:280 -click-hold $HOLD $ADM
rec fcmd-adj-pick-confirm  595000000 fcmd-adj-change-pick ck -clicks 572000000:496:280 -click-hold $HOLD $ADM
rec fcmd-adj-confirm-close 620000000 fcmd-adj-pick-confirm - -clicks 597000000:480:300 -click-hold $HOLD $ADM
# 16. 出擊命令那條路的同一張「副官」小選單（收據 `fadj` 的序列：艦隊編成按副官 (432,310)，360M）：
#     卸任、變更 → 清單 → 詳細資料 → 出擊(決定) → 回艦隊編成 → 確認 → 決定 建出帶新副官的艦隊。
rec fcmd-fadj 360000000 - ck \
  -tick 20000 -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -rclicks 130000000:300:200,155000000:300:200 \
  -clicks 40000000:300:170,95000000:300:170,180000000:352:85,210000000:400:85,240000000:400:45,270000000:488:290,320000000:432:310 \
  -click-hold $HOLD $ADM
rec fcmd-fadj-dismiss      385000000 fcmd-fadj        -  -clicks 362000000:430:330 -click-hold $HOLD $ADM
rec fcmd-fadj-change       385000000 fcmd-fadj        ck -clicks 362000000:430:310 -click-hold $HOLD $ADM
rec fcmd-fadj-change-pick  410000000 fcmd-fadj-change ck -clicks 387000000:450:70 -click-hold $HOLD $ADM
rec fcmd-fadj-pick-confirm 435000000 fcmd-fadj-change-pick ck -clicks 412000000:496:280 -click-hold $HOLD $ADM
rec fcmd-fadj-confirm-fleet 485000000 fcmd-fadj-pick-confirm - -clicks 437000000:480:310,462000000:510:310 -click-hold $HOLD $ADM
# 6. 從 `dec1` 只把游標移到瓦爾哈拉（不按）→ 左下角照游標重寫：剛編成的
#    第１艦隊算不算進「艦隊 a：b」（星域記錄 `+0x28` 建艦隊時就加了）。
rec fcmd-hover-home 440000000 fcmd-base - -mouse-x 32 -mouse-y 32 -mouse-at 421000000
# 17. 右下角清單的四個分頁與捲動（04F8:05ED 換 [ds:957E]／[ds:9582]，docs/re/319）。
#     分頁鈕（滑鼠）：艦隊 (350,250)、行星 (415,250)、敵艦隊 (490,250)、提督 (560,250)；
#     列 k 在 (450, 268＋20k)；▲ (617,268)、▼ (617,348) 是按住一次一列（3765:0255），
#     這裡按 0.3M 步 ＝ 剛好五列（一頁）。每批多倒分頁編號 ds:8F1F 與四個分頁的
#     筆數／捲動位移（$TABS：9592/9593 艦隊、95B4/95B5 敵艦隊、95D8/95D9 提督、CFFC／95D6 行星）。
rec fcmd-tab-planet        445000000 fcmd-base           ck -clicks 421000000:415:250 -click-hold $HOLD $TABS
rec fcmd-tab-planet-row    470000000 fcmd-tab-planet     ck -clicks 446000000:450:268 -click-hold $HOLD $TABS
rec fcmd-tab-planet-ok     495000000 fcmd-tab-planet-row -  -clicks 471000000:462:289 -click-hold $HOLD $TABS
rec fcmd-tab-planet-rclick 495000000 fcmd-tab-planet-row -  -rclicks 471000000:300:200 -click-hold $HOLD $TABS
rec fcmd-tab-planet-scroll 470000000 fcmd-tab-planet     -  -clicks 446000000:617:250 -click-hold $HOLD5 $TABS
rec fcmd-tab-enemy         445000000 fcmd-base           ck -clicks 421000000:490:250 -click-hold $HOLD $TABS
rec fcmd-tab-enemy-row     470000000 fcmd-tab-enemy      -  -clicks 446000000:450:268 -click-hold $HOLD $TABS
rec fcmd-tab-admiral       445000000 fcmd-base           ck -clicks 421000000:560:250 -click-hold $HOLD $TABS
rec fcmd-tab-admiral-down  470000000 fcmd-tab-admiral    ck -clicks 446000000:617:348 -click-hold $HOLD5 $TABS
rec fcmd-tab-admiral-down2 495000000 fcmd-tab-admiral-down - -clicks 471000000:617:348 -click-hold $HOLD5 $TABS
rec fcmd-tab-admiral-up    495000000 fcmd-tab-admiral-down - -clicks 471000000:617:268 -click-hold $HOLD5 $TABS
rec fcmd-tab-admiral-end   470000000 fcmd-tab-admiral    -  -clicks 446000000:617:348 -click-hold $HOLD $TABS
rec fcmd-tab-admiral-row   470000000 fcmd-tab-admiral    ck -clicks 446000000:450:268 -click-hold $HOLD $TABS
rec fcmd-tab-admiral-cancel 495000000 fcmd-tab-admiral-row - -clicks 471000000:448:280 -click-hold $HOLD $TABS
rec fcmd-tab-admiral-rclick 495000000 fcmd-tab-admiral-row - -rclicks 471000000:300:200 -click-hold $HOLD $TABS
rec fcmd-tab-admiral-sortie 495000000 fcmd-tab-admiral-row ck -clicks 471000000:496:280 -click-hold $HOLD $TABS
#     出擊那條路（07F9:00D5）的艦隊編成按右鍵 ＝ 確認：直接進艦隊組成（043F 回來不看結果就 297B:0066 建艦隊）。
#     提督分頁與出擊命令兩條路各一批（後者冷啟動同 fcmd-fok，只是 320M 那一下換成右鍵）。
rec fcmd-tab-sortie-rclick 520000000 fcmd-tab-admiral-sortie - -rclicks 496000000:300:200 -click-hold $HOLD $TABS $STOCK
rec fcmd-forg-rclick 350000000 - - \
  -tick 20000 -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -rclicks 130000000:300:200,155000000:300:200,320000000:300:200 \
  -clicks 40000000:300:170,95000000:300:170,180000000:352:85,210000000:400:85,240000000:400:45,270000000:488:290 \
  -click-hold $HOLD $STOCK
#     編成確認、組成決定 → 第２艦隊，提督分頁還選著、清單少一位；再點艦隊分頁 → 兩支。
rec fcmd-tab-fleet2        545000000 fcmd-tab-admiral-sortie ck -clicks 496000000:480:310,521000000:510:310 -click-hold $HOLD $TABS
rec fcmd-tab-fleet2-list   570000000 fcmd-tab-fleet2     -  -clicks 546000000:350:250 -click-hold $HOLD $TABS
#     再出擊四位（每位：列 0 → 出擊 → 確認 → 決定）→ 六支艦隊；艦隊分頁出現 ▲▼，▼ 捲到第 1–5 列。
rec fcmd-tab-fleet6        945000000 fcmd-tab-fleet2     ck \
  -clicks 546000000:450:268,571000000:496:280,596000000:480:310,621000000:510:310,646000000:450:268,671000000:496:280,696000000:480:310,721000000:510:310,746000000:450:268,771000000:496:280,796000000:480:310,821000000:510:310,846000000:450:268,871000000:496:280,896000000:480:310,921000000:510:310 \
  -click-hold $HOLD $TABS
rec fcmd-tab-fleet6-list   970000000 fcmd-tab-fleet6     ck -clicks 946000000:350:250 -click-hold $HOLD $TABS
rec fcmd-tab-fleet6-down   995000000 fcmd-tab-fleet6-list ck -clicks 971000000:617:348 -click-hold $HOLD5 $TABS
rec fcmd-tab-fleet6-up     1020000000 fcmd-tab-fleet6-down - -clicks 996000000:617:268 -click-hold $HOLD5 $TABS
rec fcmd-tab-fleet6-row    1020000000 fcmd-tab-fleet6-down ck -clicks 996000000:450:348 -click-hold $HOLD $TABS
rec fcmd-tab-fleet6-close  1045000000 fcmd-tab-fleet6-row - -clicks 1021000000:490:350 -click-hold $HOLD $TABS
# 18. 左鍵點星圖（06E8 段 +0A35：3705:00DC 找星域；點的是右側面板看的那一格 ds:9588 → 星系情報視窗
#     073F:05CB → 060A → 06C5／0A39／0AFD；別的格 → 073F:02EE 換選星域；docs/re/320）。
#     右鍵點星圖什麼都不換（fcmd-map-rclick：ds:9588 仍是 0、右側面板與清單不動，只有左下角跟著游標）。
#     瓦爾哈拉 (32,32)、費沙 (163,96)（中立）、巴拉特 (300,270)（同盟）；「結束」在內容格 (15,16) → 滑鼠 (270,350)。
rec fcmd-rinfo             445000000 fcmd-base       ck -clicks 421000000:32:32 -click-hold $HOLD
rec fcmd-rinfo-close       470000000 fcmd-rinfo      -  -clicks 446000000:270:350 -click-hold $HOLD
rec fcmd-rinfo-rclick      470000000 fcmd-rinfo      -  -rclicks 446000000:300:200 -click-hold $HOLD
rec fcmd-rinfo-outside     470000000 fcmd-rinfo      -  -clicks 446000000:100:100 -click-hold $HOLD
rec fcmd-rinfo-fezzan      445000000 fcmd-base       ck -clicks 421000000:163:96 -click-hold $HOLD
rec fcmd-rinfo-fezzan2     470000000 fcmd-rinfo-fezzan - -clicks 446000000:163:96 -click-hold $HOLD
rec fcmd-rinfo-ballat      445000000 fcmd-base       ck -clicks 421000000:300:270 -click-hold $HOLD
rec fcmd-rinfo-ballat2     470000000 fcmd-rinfo-ballat ck -clicks 446000000:300:270 -click-hold $HOLD
rec fcmd-rinfo-ballat2-close 495000000 fcmd-rinfo-ballat2 - -clicks 471000000:270:350 -click-hold $HOLD
rec fcmd-map-rclick        445000000 fcmd-base       -  -rclicks 421000000:300:270 -click-hold $HOLD -peek lin:5E0B8:1

# 19. 占領命令走到之後（113A:0350／047D 逐支走一格、056A:0469 的 bit4 占據與 0374:0556 → 20B4:1497 的
#     戰略打行星；docs/re/322）。一回合的處理（AI 思考、出發前往／已經到達的過場、占據報告與士氣報告的
#     等待）約 90M 步，每批隔 120M；停在報告上的批次先點一下關掉再實行。每批另倒星域表與行星表。
#     (a) 亞斯提（同盟，fcmd-occupy-done 的命令）：0 → 2 → 7 → 8 → 13 → 16，第五回合到達、占據、打行星。
rec fcmd-occ-t1          665000000 fcmd-occupy-done ck -clicks 546000000:544:85,549000000:560:85 -click-hold $HOLD $OCC
rec fcmd-occ-t2          785000000 fcmd-occ-t1      ck -clicks 666000000:544:85,669000000:560:85 -click-hold $HOLD $OCC
rec fcmd-occ-t3          905000000 fcmd-occ-t2      ck -clicks 786000000:544:85,789000000:560:85 -click-hold $HOLD $OCC
rec fcmd-occ-t4         1025000000 fcmd-occ-t3      ck -clicks 906000000:544:85,909000000:560:85 -click-hold $HOLD $OCC
rec fcmd-occ-t5         1145000000 fcmd-occ-t4      ck -clicks 1026000000:544:85,1029000000:560:85 -click-hold $HOLD $OCC
rec fcmd-occ-t5b        1190000000 fcmd-occ-t5      ck -clicks 1146000000:300:200 -click-hold $HOLD $OCC
rec fcmd-occ-t6         1310000000 fcmd-occ-t5b     -  -clicks 1191000000:544:85,1194000000:560:85 -click-hold $HOLD $OCC
#     (b) 費沙（中立）：0 → 2 → 6 → 14，第三回合到達、占據、打行星（之後停在「士氣值低落」報告上）。
rec fcmd-occ-phz-done    595000000 fcmd-occupy-phezzan3 ck -clicks 572000000:432:223 -click-hold $HOLD $OCC
rec fcmd-occ-phz-t1      715000000 fcmd-occ-phz-done ck -clicks 596000000:544:85,599000000:560:85 -click-hold $HOLD $OCC
rec fcmd-occ-phz-t2      835000000 fcmd-occ-phz-t1   ck -clicks 716000000:544:85,719000000:560:85 -click-hold $HOLD $OCC
rec fcmd-occ-phz-t3      955000000 fcmd-occ-phz-t2   ck -clicks 836000000:544:85,839000000:560:85 -click-hold $HOLD $OCC
rec fcmd-occ-phz-t3b    1000000000 fcmd-occ-phz-t3   ck -clicks 956000000:300:200 -click-hold $HOLD $OCC
rec fcmd-occ-phz-t4     1120000000 fcmd-occ-phz-t3b  -  -clicks 1001000000:544:85,1004000000:560:85 -click-hold $HOLD $OCC

# 20. 一覽表四張的列（0ABD 段的四支：星系資料 0D05、艦隊資料 0C7F、提督／敵將 0ABD:00D6 → 0C0A；
#     docs/re/323）。一覽表 (448,85) → 星系資料 (460,85)／艦隊資料 (460,105)／提督資料 (460,125)／
#     敵將資料 (460,145)。列 k 在畫面 y 80＋20k → 滑鼠 (460, 50＋20k)；分頁 (456|520|584, 30)；
#     ▲ (615,50)、▼ (615,350)（按住 0.1M ＝ 一列；0.3M 是兩列）。提督詳細資料（旗標 0）經歷 (448,280)／結束 (496,280)。
#     ⚠ 游標移到會換地圖反白的列（或移出清單）的**同一瞬間**按鍵，原版把那一下丟掉
#     （fcx-lst-star-row0／-r1 沒跑到 0D05:0141；先移再按就收得到，fcx-lst-star-r0b／fcx-lst-adm-rr）
#     ——所以這一段的列與鈕都先 -mouse-at 移過去、10M 步之後才按。
#     另倒：ds:9588 選著的星域、ds:96E6／96EC 星系清單的分頁與位移、ds:96C2／96C3 艦隊清單筆數與位移、
#     ds:8F2D／1DF2 每座位記住的星系分頁與位移、ds:8F31 艦隊清單分頁、ds:96A0／96A1 提督清單筆數與位移、
#     ds:F182／F183 地圖反白的格與開清單前存起來的那一格。
#     (a) 星系資料：點列 ＝ 換選那一格（073F:02EE）＋ 關清單 ＋ 星系情報視窗（073F:05CB）；分頁與位移每座位記住。
rec fcmd-lst-star          470000000 fcmd-base     ck -clicks 421000000:448:85,446000000:460:85 -click-hold $HOLD $LST
rec fcmd-lst-star-row      495000000 fcmd-lst-star ck $(mvc 471000000 460 50) -click-hold $HOLD $LST
rec fcmd-lst-star-row-close 520000000 fcmd-lst-star-row - $(mvc 496000000 270 350) -click-hold $HOLD $LST
rec fcmd-lst-star-row2     495000000 fcmd-lst-star ck $(mvc 471000000 460 90) -click-hold $HOLD $LST
rec fcmd-lst-star-row2-close 520000000 fcmd-lst-star-row2 - $(mvc 496000000 270 350) -click-hold $HOLD $LST
rec fcmd-lst-star-hover    485000000 fcmd-lst-star - -mouse-x 460 -mouse-y 250 -mouse-at 471000000 $LST
rec fcmd-lst-star-hover-out 485000000 fcmd-lst-star - -mouse-x 100 -mouse-y 100 -mouse-at 471000000 $LST
rec fcmd-lst-star-tab-assign 495000000 fcmd-lst-star - $(mvc 471000000 520 30) -click-hold $HOLD $LST
rec fcmd-lst-star-tab-tax  495000000 fcmd-lst-star ck $(mvc 471000000 584 30) -click-hold $HOLD $LST
rec fcmd-lst-star-tax-rclick 520000000 fcmd-lst-star-tab-tax ck $(mvr 496000000 300 200) -click-hold $HOLD $LST
rec fcmd-lst-star-reopen   570000000 fcmd-lst-star-tax-rclick - -clicks 521000000:448:85,546000000:460:85 -click-hold $HOLD $LST
rec fcmd-lst-star-down     495000000 fcmd-lst-star ck $(mvc 471000000 615 350) -click-hold $HOLD1 $LST
rec fcmd-lst-star-down-row 520000000 fcmd-lst-star-down - $(mvc 496000000 460 50) -click-hold $HOLD $LST
rec fcmd-lst-star-up       520000000 fcmd-lst-star-down - $(mvc 496000000 615 50) -click-hold $HOLD1 $LST
rec fcmd-lst-star-rclick   495000000 fcmd-lst-star - $(mvr 471000000 300 200) -click-hold $HOLD $LST
#     (b) 艦隊資料：點列 ＝ 艦隊視窗（0DEA:0003，與右下角清單同一支），關了回清單。
rec fcmd-lst-fleet         470000000 fcmd-base      ck -clicks 421000000:448:85,446000000:460:105 -click-hold $HOLD $LST
rec fcmd-lst-fleet-row     495000000 fcmd-lst-fleet ck $(mvc 471000000 460 50) -click-hold $HOLD $LST
rec fcmd-lst-fleet-row-close 520000000 fcmd-lst-fleet-row - $(mvc 496000000 480 340) -click-hold $HOLD $LST
rec fcmd-lst-fleet-row-rclick 520000000 fcmd-lst-fleet-row - $(mvr 496000000 300 200) -click-hold $HOLD $LST
rec fcmd-lst-fleet-row1    495000000 fcmd-lst-fleet - $(mvc 471000000 460 70) -click-hold $HOLD $LST
rec fcmd-lst-fleet-tab-unit 495000000 fcmd-lst-fleet - $(mvc 471000000 520 30) -click-hold $HOLD $LST
rec fcmd-lst-fleet-tab-ops 495000000 fcmd-lst-fleet - $(mvc 471000000 584 30) -click-hold $HOLD $LST
rec fcmd-lst-fleet-rclick  495000000 fcmd-lst-fleet - $(mvr 471000000 300 200) -click-hold $HOLD $LST
#     (c) 提督資料：點列 ＝ 提督詳細資料（0D92:000D(r, 0, 1, 0)：經歷／結束），關了回清單。
rec fcmd-lst-adm           470000000 fcmd-base    ck -clicks 421000000:448:85,446000000:460:125 -click-hold $HOLD $LST
rec fcmd-lst-adm-row       495000000 fcmd-lst-adm ck $(mvc 471000000 460 50) -click-hold $HOLD $LST
rec fcmd-lst-adm-row-close 520000000 fcmd-lst-adm-row - $(mvc 496000000 496 280) -click-hold $HOLD $LST
rec fcmd-lst-adm-row-bio   520000000 fcmd-lst-adm-row ck $(mvc 496000000 448 280) -click-hold $HOLD $LST
rec fcmd-lst-adm-row-bio-close 545000000 fcmd-lst-adm-row-bio - $(mvc 521000000 496 280) -click-hold $HOLD $LST
rec fcmd-lst-adm-row-rclick 520000000 fcmd-lst-adm-row - $(mvr 496000000 300 200) -click-hold $HOLD $LST
rec fcmd-lst-adm-hover     485000000 fcmd-lst-adm - -mouse-x 460 -mouse-y 210 -mouse-at 471000000 $LST
rec fcmd-lst-adm-down      495000000 fcmd-lst-adm ck $(mvc 471000000 615 350) -click-hold $HOLD1 $LST
rec fcmd-lst-adm-down-row  520000000 fcmd-lst-adm-down - $(mvc 496000000 460 50) -click-hold $HOLD $LST
rec fcmd-lst-adm-rclick    495000000 fcmd-lst-adm - $(mvr 471000000 300 200) -click-hold $HOLD $LST
rec fcmd-lst-adm-outside   495000000 fcmd-lst-adm - $(mvc 471000000 100 100) -click-hold $HOLD $LST
#     (d) 敵將資料：同 (c)，旗標 (r, 0, 0, 0)；沒有列的游標反白（0C0A:0351 的第三個參數 0）。
rec fcmd-lst-enemy         470000000 fcmd-base      ck -clicks 421000000:448:85,446000000:460:145 -click-hold $HOLD $LST
rec fcmd-lst-enemy-row     495000000 fcmd-lst-enemy ck $(mvc 471000000 460 50) -click-hold $HOLD $LST
rec fcmd-lst-enemy-row-close 520000000 fcmd-lst-enemy-row - $(mvc 496000000 496 280) -click-hold $HOLD $LST
rec fcmd-lst-enemy-hover   485000000 fcmd-lst-enemy - -mouse-x 460 -mouse-y 210 -mouse-at 471000000 $LST
rec fcmd-lst-enemy-rclick  495000000 fcmd-lst-enemy - $(mvr 471000000 300 200) -click-hold $HOLD $LST

# 21. 駐留命令走到目的地 → 整支解散（113A:06A6 → 297B:025D：提督與副官 +0x09 |= 80、艦隊 +0x04/+0x05 ← FF、
#     各組艦艇還回所在星域的存量（艦種 7 不還）、星域 +0x28 −1、090A:0E6D、ds:8F49 −1；docs/re/329）。
#     玩家介面沒有「解散」鈕，駐留就是解散；駐留點自己所在的星域 → 下一次實行就地解散。
#     peek 見標頭的 DIS。
rec fcmd-garrison-report  570000000 fcmd-garrison-done -  -clicks 546000000:544:85,549000000:560:85 -click-hold $HOLD $DIS
rec fcmd-garrison-t1      665000000 fcmd-garrison-done -  -clicks 546000000:544:85,549000000:560:85 -click-hold $HOLD $DIS
rec fcmd-garrison-self         520000000 fcmd-garrison-pick ck -clicks 497000000:32:32 -click-hold $HOLD
rec fcmd-garrison-self-done    545000000 fcmd-garrison-self ck -clicks 522000000:432:223 -click-hold $HOLD $DIS
rec fcmd-garrison-self-report  570000000 fcmd-garrison-self-done - -clicks 546000000:544:85,549000000:560:85 -click-hold $HOLD $DIS
rec fcmd-garrison-self-t1      665000000 fcmd-garrison-self-done - -clicks 546000000:544:85,549000000:560:85 -click-hold $HOLD $DIS

# 22. 戰略打行星（20B4:1497，docs/re/330）的同狀態收據：輸入與 fcmd-occ-t5／phz-t3 相同，在進入 1497
#     與它回到 0374:0556 之前（20B4:1507）各倒一份 1 MB。remake 的回合亂數第三回合起就岔開，
#     所以對拍是「拿 mem-in 的狀態驅動 remake 的打行星，比 mem-out」（規格 014：改記憶體，不靠運氣）。
#     步數來自探索批次 fcx-assault-ast／-phz 的 -regs-at（本機證據）。
recmem fcmd-assault-ast 1102006030 1102557471 1145000000 fcmd-occ-t4 - -clicks 1026000000:544:85,1029000000:560:85 -click-hold $HOLD $OCC
recmem fcmd-assault-phz 912160349 912742120 955000000 fcmd-occ-phz-t2 - -clicks 836000000:544:85,839000000:560:85 -click-hold $HOLD $OCC
#     同一條輸入第二次進 1497：同盟攻方在馬爾・亞迪特（24）拿下威爾勃恩（docs/re/335 §六）。1922 的攻占框等 90 tick，1507 在 #932.35M。
recmem fcmd-assault-ai24 912920224 932345563 933000000 fcmd-occ-phz-t2 - -clicks 836000000:544:85,839000000:560:85 -click-hold $HOLD $OCC

# 23. 回合裡的過場視窗（113A:0B46／0797／0A4E／0899，056A:011C「戰略模式處理中」；docs/re/332、規格 059）。
#     每一張錄在 -watch-screen 量到的穩定區間中間（區間與第二個取樣點見 docs/re/332 §六）。
#     錯誤報告的局面用 -poke 造（規格 014）：經費不足把 ds:8F52[0] 寫成 32（< 維持費 22 ＋ 移動費 26），
#     被擋把亞爾提那（星域 2）的擁有者 +0x12 改成同盟——在電腦下完之後（#554.5M）改，AI 看到的局面不變。
#     問話：fcmd-occ-t1（艦隊在亞爾提那、往亞斯提）在實行前把亞爾維斯（星域 7）改成同盟，點「占領」(424,230)。
rec fcmd-turn-think        552600000 fcmd-done - $EXEC $TURN
rec fcmd-turn-busy         555300000 fcmd-done - $EXEC $TURN
rec fcmd-turn-depart       558800000 fcmd-done - $EXEC $TURN
rec fcmd-turn-arrive       570000000 fcmd-done - $EXEC $TURN
rec fcmd-turn-garrison     570000000 fcmd-garrison-done - $EXEC $TURN
rec fcmd-turn-occ-depart   558800000 fcmd-occupy-done - $EXEC $TURN
rec fcmd-turn-occ-onward   570000000 fcmd-occupy-done - $EXEC $TURN
rec fcmd-turn-nofund       565000000 fcmd-done - $EXEC -poke lin:5DA82@548000000=20,00,00,00 $TURN
rec fcmd-turn-block        565000000 fcmd-done - $EXEC -poke lin:5F3DA@554500000=01 $TURN
rec fcmd-turn-block-pre    565000000 fcmd-done - $EXEC -poke lin:5F3DA@548000000=01 $TURN
rec fcmd-turn-ask          680000000 fcmd-occ-t1 - -clicks 666000000:544:85,669000000:560:85,700000000:424:230 -click-hold $HOLD -poke lin:5F650@665500000=01 $TURN
rec fcmd-turn-ask-depart   704000000 fcmd-occ-t1 - -clicks 666000000:544:85,669000000:560:85,700000000:424:230 -click-hold $HOLD -poke lin:5F650@665500000=01 $TURN
rec fcmd-turn-ask-onward   716000000 fcmd-occ-t1 - -clicks 666000000:544:85,669000000:560:85,700000000:424:230 -click-hold $HOLD -poke lin:5F650@665500000=01 $TURN
rec fcmd-turn-clk22        565000000 fcmd-done - -clicks 546000000:544:85,549000000:560:85,558000000:560:85 -click-hold $HOLD $TURN

# 24. 回合裡的報告框（3826:0307；docs/re/333、規格 059）：拿下行星（20B4:1922，人類／電腦）、士氣值低落與
#     已經恢復（1261:0138）、已全軍覆沒（1261:0267）。錄在 90 tick 等待的中間。
#     人類拿下：fcmd-occ-t4 在實行前把巴溫（行星 +0x14 防衛力，lin:53D20）改成 100，第五回合一打就換手。
#     已經恢復：fcmd-base 把第１艦隊 +0x00 ← 01、+0x3C ← 40、艦艇 0 +0x06 ← 40、+0x08 ← 100 艘，實行時瓦爾哈拉補回（1AAB:0FAF）。
#     已全軍覆沒：fcmd-occ-t4 把艦艇 0 +0x08 改 0，走一格之後 113A:03C5 → 1261:0067(f,1) 設 bit3。
#     拿下行星（人類與電腦）、士氣值低落、已全軍覆沒四批用 recmem：mem-in 是那一回合實行之前（poke 之前）的局面、
#     mem-out 是報告框停著的那一刻，分岔時拿來逐欄找差在哪（docs/re/335 §七）。
recmem fcmd-turn-capture 1025000040 1111000000 1111500000 fcmd-occ-t4 - -clicks 1026000000:544:85,1029000000:560:85 -click-hold $HOLD -poke "lin:53D20@1025000050=64 00" $OCC
recmem fcmd-turn-ai-capture 835000040 923000000 923400000 fcmd-occ-phz-t2 - -clicks 836000000:544:85,839000000:560:85 -click-hold $HOLD $OCC
recmem fcmd-turn-morale-low 835000040 961000000 961800000 fcmd-occ-phz-t2 - -clicks 836000000:544:85,839000000:560:85 -click-hold $HOLD $OCC
rec fcmd-turn-recover       442100000 fcmd-base - -clicks 421000000:544:85,424000000:560:85 -click-hold $HOLD -poke "lin:60BA4@420000050=01;lin:60BE0@420000050=28;lin:668B8@420000050=28;lin:668BA@420000050=64 00" $TURN
recmem fcmd-turn-no-ships 1025000040 1068000000 1068200000 fcmd-occ-t4 - -clicks 1026000000:544:85,1029000000:560:85 -click-hold $HOLD -poke "lin:668BA@1025000050=00 00" $OCC

# 25. 同一回合多支人類艦隊的過場（#72、規格 059）：S3「巴米利恩會戰」帝國第一回合（冷啟動序列同
#     `tools/enemytab_receipts.sh` 的 `enm-map`），開局第９艦隊（槽 8，23→25 移動）、第１３艦隊（槽 12，
#     28→26 占領）、第１５艦隊（槽 14，23→25 移動）已經有命令，按實行就連續三張出發前往，不必改記憶體。
#     時間軸（`docs/re/336` §三）：思考中 #305.90M–#307.44M；過場 `113A:0B46` #307.86M／#333.26M／#358.66M，
#     每張第一段 22 tick、第二段 90 tick；每張錄在停頓中間。
rec fcmd-s3-base 300000000 - ck \
  -tick 20000 -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -clicks 40000000:300:230,95000000:300:230,190000000:300:282,240000000:318:168 \
  -click-hold $HOLD
rec fcmd-s3-think 306800000 fcmd-s3-base - $S3EXEC $TURN
rec fcmd-s3-w1a   311500000 fcmd-s3-base - $S3EXEC $TURN
rec fcmd-s3-w1b   323000000 fcmd-s3-base - $S3EXEC $TURN
rec fcmd-s3-w2a   337000000 fcmd-s3-base - $S3EXEC $TURN
rec fcmd-s3-w2b   348000000 fcmd-s3-base - $S3EXEC $TURN
rec fcmd-s3-w3a   362500000 fcmd-s3-base - $S3EXEC $TURN
rec fcmd-s3-w3b   374000000 fcmd-s3-base - $S3EXEC $TURN

# 26. 戰略階段的艦隊全滅（#57，`0374:0608` → `06BC` 耗損 → `073F` 報告 → `297B:034A`；`docs/re/336` §九）：
#     費沙那條第三回合士氣掉到 47（回合尾巴才設 bit0，同回合不耗損），第四回合逐格掃到費沙：
#     #1031.357M `06BC` 扣兵、第 0 組 0 艘 → #1031.363M `3826:0307`「…已經全滅。」→ 90 tick → #1049.345M `034A`
#     （先把 `+0x6C` 的假指標交給 `35D1:0102`，那段迴圈跑到 #1056.12M 才進提督那一步）→ #1057.45M 下一則報告。
#     `-report` 錄在報告等待中間；另一批 mem-in 在第四回合掃描之前、mem-out 在 `034A` 之後。執行器 f7523f9 起不再崩潰（#73）。
rec fcmd-turn-annihilated-report 1040000000 fcmd-occ-phz-t3b - -clicks 1001000000:544:85,1004000000:560:85 -click-hold $HOLD $OCC
recmem fcmd-turn-annihilated 1031000000 1057000000 1057100000 fcmd-occ-phz-t3b - -clicks 1001000000:544:85,1004000000:560:85 -click-hold $HOLD $OCC

# 27. 兩位數艦隊編號（#74，`3982:0038` 的 `%2d` 那條路；`docs/re/143`）。S3 開局帝國槽 9 就是
#     第１０艦隊，不必改記憶體：一覽表（資料 (448,85) → 艦隊 (460,105)）點第 8 列（列 7，滑鼠 y 50＋20×7）
#     → 艦隊視窗；再 作戰 (490,305) → 待命 (497,352) → 指定待命行星。兩張視窗的標題都印半形「10」。
#     ⚠ 點列會換反白，移游標與按鍵要分兩批（$(mvc …)），否則那一下被丟掉——探索時踩過一次。
rec fcmd-s3-n10-list    340000000 fcmd-s3-base ck \
  -mouse-x 448 -mouse-y 85 -mouse-at 301000000 -clicks 311000000:448:85,321000000:460:105 -click-hold $HOLD
rec fcmd-s3-n10-win     365000000 fcmd-s3-n10-list ck $(mvc 341000000 450 190) -click-hold $HOLD
rec fcmd-s3-n10-menu    390000000 fcmd-s3-n10-win  ck $(mvc 366000000 490 305) -click-hold $HOLD
rec fcmd-s3-n10-standby 415000000 fcmd-s3-n10-menu -  $(mvc 391000000 497 352) -click-hold $HOLD

# 28. 逐格開戰之前的「<星域>星系內／發現敵方艦隊。」（#75，`0374:0489` → `04B8` → `3826:0307`；
#     `docs/re/340`）。S3 帝國第一回合：實行之後三張過場各等滿 90 tick，#388M 起這一格
#     （達希利 26）要開戰，報告在 #392–#410M 之間（#410M 整片換掉 ＝ 進戰場）。
#     只對人類座位印；兩個座位都是人的話同一個視窗連開兩次。這一局座位 1 是電腦（`ds:8F22`＝1），
#     所以只印一次。$FOUND 倒的是判斷用的那幾格：目前格、兩個座位是不是電腦、座位的陣營、
#     訊息速度、情報參謀表。
FOUND="-peek lin:5DA47:1,lin:5DA51:2,lin:5DA55:2,lin:5DA68:2,lin:5DA7D:8"
rec fcmd-s3-found 394000000 fcmd-s3-base - $S3EXEC $FOUND
