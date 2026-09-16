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
TABS="-peek lin:5DA4F:1,lin:5E0C2:2,lin:5E0E4:2,lin:5E108:2,lin:61B2C:1,lin:5E106:1"   # 分頁編號與四個分頁的筆數／位移（第 17 段）

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
rec fcmd-garrison-done    545000000 fcmd-garrison-confirm - -clicks 522000000:432:223 -click-hold $HOLD
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
