#!/usr/bin/env bash
# 交戰演出的收據（`docs/re/324`、規格 056）。
#
#   tools/engagement_receipts.sh          # 全部重錄（約 5 分鐘）
#
# 載具是 S3「巴米利恩會戰」：主選單第 3 列「取進度」×2 → 防拷 → S3 (300,282) →
# 帝國 (318,168) → 系統指令／實行 ×6 進戰場（`docs/internals/34`），之後每個戰術
# 回合是 (240,200) 兩下（開選單、實行）。第五回合的處理裡同盟艦隊打帝國行星
# 雷奧米爾（`20B4:0276` → `1DD6:0152`），演出從 #1140.09M 到 #1169.07M。
#
# ⚠ 收據綁執行器版本（AGENTS.md）；檢查點在 dosgolem 副本的 workplace/fcmd/。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
CK="$GOLEM/workplace/fcmd"
mkdir -p "$CK"
PLANE_LO=6234E; PLANE_HI=6360E
HOLD=3000000
ENG="-peek lin:5DA66:2,lin:5E0B8:1"   # ds:8F36 動畫效果／8F37 動畫速度、ds:9588

rec() {
  local name=$1 steps=$2 from=$3 save=$4; shift 4
  local extra=()
  [[ "$from" != - ]] && extra+=(-load-state "/src/workplace/fcmd/$from.state")
  [[ "$save" == ck ]] && extra+=(-save-state "$((steps - 1)):/src/workplace/fcmd/$name.state")
  tools/oracle.sh "$name" "${extra[@]}" -steps "$steps" "$@" \
    -dump-mem "$PLANE_LO-$PLANE_HI:/src/workplace/fcmd/$name-plane.bin" >/dev/null
  cp "$CK/$name-plane.bin" "$ROOT/workplace/oracle/$name/plane.bin"
  printf '%s  %s\n' "$name" "$(sed -n 's/^screen.bin SHA-256: //p' "$ROOT/workplace/oracle/$name/meta.txt" | cut -c1-16)"
}

cd "$ROOT"
# ENG_FROM=2 只跑第 2 段起（檢查點 eng-t5 已在）。
if [[ "${ENG_FROM:-0}" -le 1 ]]; then
# 0. 進戰場（達希利，第 1 回合）。
rec eng-entry 900000000 - ck \
  -tick 20000 -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -clicks 40000000:300:230,95000000:300:230,190000000:300:282,240000000:318:168,300000000:544:85,330000000:560:85,405000000:544:85,435000000:560:85,510000000:544:85,540000000:560:85,615000000:544:85,645000000:560:85,720000000:544:85,750000000:560:85,825000000:544:85,855000000:560:85 \
  -click-hold $HOLD $ENG
# 1. 四個回合的「實行」，停在第五回合處理開始前（演出在 #1140.09M）。
rec eng-t5 1139000000 eng-entry ck \
  -clicks 920000000:240:200,935000000:240:200,1000000000:240:200,1015000000:240:200,1080000000:240:200,1095000000:240:200,1110000000:240:200,1125000000:240:200 \
  -click-hold $HOLD $ENG
fi
# 2. 第五回合的「實行」(240,200) 在 #1140M，演出 #1140.09M 起：開場（面板、三個視窗、
#    兩張圖）、中段（方塊在動）、尾聲、結束回戰場。
rec eng-open  1141500000 eng-t5 -  -clicks 1140000000:240:200 -click-hold $HOLD $ENG
rec eng-text  1144000000 eng-t5 ck -clicks 1140000000:240:200 -click-hold $HOLD $ENG
rec eng-mid   1150000000 eng-text - $ENG
rec eng-mid2  1156000000 eng-text - $ENG
rec eng-end   1169000000 eng-text - $ENG
rec eng-after 1172000000 eng-text - $ENG
# 3. 動畫效果關閉（ds:8F36 ← 0）：同一段時間裡有沒有演出、有沒有換曲。
rec eng-skip  1172000000 eng-t5 - -poke lin:5DA66@1139500000=00 -clicks 1140000000:240:200 -click-hold $HOLD -watch-screen 1000000 -screen-delta 3000 $ENG
rec eng-skip-mid 1150000000 eng-t5 - -poke lin:5DA66@1139500000=00 -clicks 1140000000:240:200 -click-hold $HOLD $ENG
# 4. 演出中按左鍵／右鍵會不會提早結束（#1146M 按、#1152M 看）。
rec eng-click 1152000000 eng-text - -clicks 1146000000:300:200 -click-hold $HOLD -watch-screen 1000000 -screen-delta 3000 $ENG
rec eng-rclick 1152000000 eng-text - -rclicks 1146000000:300:200 -click-hold $HOLD -watch-screen 1000000 -screen-delta 3000 $ENG
# 5. 戰鬥中的士氣報告（`3826:0405` 臨時視窗，#72、規格 059、`docs/re/336` §八）：戰場結算最後一步 `181B:03F0` →
#    `185C:001D` 對沒被消滅的艦隊叫 `1261:0104`。前五回合沒有自然翻轉，照規格 014 在 #1170M 把第１３艦隊
#    （槽 12，lin 6115C）`+0x00` 寫成 05（bit0 士氣低落＋bit2 翻轉）、`+0x3C`（lin 61198）寫成 40。
#    #1187.96M 開窗、#1188.57M 起等 90 tick、#1206.55M 關；錄在中間。
rec eng-morale-low 1197000000 eng-t5 - -poke "lin:6115C@1170000000=05;lin:61198@1170000000=28" -clicks 1140000000:240:200 -click-hold $HOLD $ENG
# 6. 兩位數編號（#74）：打雷奧米爾的是同盟第１６艦隊（槽 31，艦艇 135／137–141／147），但索敵
#    等級 0 時演出印的是「第？艦隊　？？？…」。照規格 014 把那七筆艦艇的 `+0x17` 寫成 3，
#    演出的右框就印「第16艦隊　曹雪」——半形兩位數（`3982:0038` 的 `%2d`）。
KNOWN=$(for a in 691EB 69287 692D5 69323 69371 693BF 69593; do printf 'lin:%s@1139500000=03;' "$a"; done)
rec eng-known 1144000000 eng-t5 - -poke "$KNOWN" -clicks 1140000000:240:200 -click-hold $HOLD $ENG
# 7. 戰場那一條的艦隊全滅（#57，規格 062 的子步 8-1）：`185C:001D` 只看第 0 組艘數，
#    0 就走 `028E` 報告（`3826:0405` 臨時視窗「<陣營>第Ｎ艦隊<提督>\n已經全滅。」）
#    ＋ `00F0` 擦小地圖點、作廢艦艇、`297B:034A` 提督失蹤。前五回合沒有自然全滅，
#    照規格 014 造局面：帝國槽 12（第１３艦隊 舒坦梅茲，`lin:6115C`）第 0 組的艦艇池
#    索引是 80（`f[+0x10]`），該筆記錄 `lin:668B2 + 80×78 = 68112`，艘數 `+0x08`
#    ＝ `lin:6811A`。在第 ⑧ 步之前（#1187.90M，步 ⑦ 已經跑完）寫成 0。
#    時間線：#1187.92M 進第 ⑧ 步 → #1188.1M 報告還在畫、#1188.4M 才畫完（收據錄在 1188.4M）→ 等 90 tick。
#    mem-in 在寫完之後、第 ⑧ 步之前；mem-out 留到 `35D1:0102` 那段長迴圈跑完（第七輪的坑）。
ANN="lin:6811A@1187900000=00,00"
rec eng-annihilate 1188400000 eng-t5 - -poke "$ANN" -clicks 1140000000:240:200 -click-hold $HOLD $ENG
rec eng-annihilate-mem 1189900000 eng-t5 - -poke "$ANN" -clicks 1140000000:240:200 -click-hold $HOLD $ENG \
  -dump-mem-at "1187910000:lin:0:1048576:/src/workplace/fcmd/eng-annihilate-in.bin;1189800000:lin:0:1048576:/src/workplace/fcmd/eng-annihilate-out.bin"
cp "$CK/eng-annihilate-in.bin" "$ROOT/workplace/oracle/eng-annihilate-mem/mem-in.bin"
cp "$CK/eng-annihilate-out.bin" "$ROOT/workplace/oracle/eng-annihilate-mem/mem-out.bin"
