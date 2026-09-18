#!/usr/bin/env bash
# 右下角「敵艦隊」分頁有東西時的收據（`docs/re/325`、規格 052 的補充）。
#
#   tools/enemytab_receipts.sh            # 全部重錄（約 3 分鐘）
#   ENM_FROM=1 tools/enemytab_receipts.sh # 檢查點 enm-map 已在，只跑改記憶體之後那些
#
# 載具是 S3「巴米利恩會戰」帝國座位的第一回合戰略地圖（`docs/re/324` 同一條序列
# 停在第一次「系統指令」之前）。這一頁要有列，得有別的陣營的艦隊在選中的星域
# 而且索敵等級（艦艇 `+0x17` 各組取最大）不是 0——那是打過仗才有的東西，所以照
# 規格 014 直接改記憶體：把同盟第２艦隊（槽 17，四組艦艇 142／144／145／146）的
# `+0x06` 寫成 0（瓦爾哈拉，開局選中的星域）、那四筆艦艇的 `+0x17` 寫成要的等級。
# 位址：艦隊表 ds:C074 ＝ lin 60BA4，一筆 122；艦艇池 far 指標 ds:CFB4（lin 61AE4）
# 在這個檢查點是 668B:0002 ＝ lin 668B2，一筆 78（`$ENM` 每批都倒它，變了就整條重算）。
#
# ⚠ 收據綁執行器版本（AGENTS.md）；檢查點在 dosgolem 副本的 workplace/fcmd/。
# 座標是滑鼠座標（畫面 y ＝ 滑鼠 y ＋ 40）。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
CK="$GOLEM/workplace/fcmd"
mkdir -p "$CK"
PLANE_LO=6234E; PLANE_HI=6360E
FLEET17_LO=613BE; FLEET17_HI=61438      # 槽 17 那一筆 122 bytes
HOLD=3000000
# ds:CFB4 艦艇池指標、ds:9588 選中的星域、ds:8F1F 分頁編號、ds:95B4/95B5 敵艦隊分頁筆數／位移、槽 17 +0x70 索敵等級
ENM="-peek lin:61AE4:4,lin:5E0B8:1,lin:5DA4F:1,lin:5E0E4:2,lin:6142E:1"
AT=300000000                            # 檢查點的下一步：改記憶體都在這一步
# 槽 17 的 +0x06 ← 0；四筆艦艇的 +0x17
P17="lin:613C4@$AT=00"
S17() { echo "lin:6940D@$AT=$1;lin:694A9@$AT=$2;lin:694F7@$AT=$3;lin:69545@$AT=$4"; }
# 槽 16（同盟第１艦隊，六組 102／104／105／106／107／108）
P16="lin:6134A@$AT=00"
S16() { echo "lin:687DD@$AT=$1;lin:68879@$AT=$1;lin:688C7@$AT=$1;lin:68915@$AT=$1;lin:68963@$AT=$1;lin:689B1@$AT=$1"; }

rec() {
  local name=$1 steps=$2 from=$3 save=$4; shift 4
  local extra=()
  [[ "$from" != - ]] && extra+=(-load-state "/src/workplace/fcmd/$from.state")
  [[ "$save" == ck ]] && extra+=(-save-state "$((steps - 1)):/src/workplace/fcmd/$name.state")
  tools/oracle.sh "$name" "${extra[@]}" -steps "$steps" "$@" \
    -dump-mem "$PLANE_LO-$PLANE_HI:/src/workplace/fcmd/$name-plane.bin,$FLEET17_LO-$FLEET17_HI:/src/workplace/fcmd/$name-fleet17.bin" >/dev/null
  cp "$CK/$name-plane.bin" "$ROOT/workplace/oracle/$name/plane.bin"
  cp "$CK/$name-fleet17.bin" "$ROOT/workplace/oracle/$name/fleet17.bin"
  printf '%s  %s\n' "$name" "$(sed -n 's/^screen.bin SHA-256: //p' "$ROOT/workplace/oracle/$name/meta.txt" | cut -c1-16)"
}
# mvc／mvr <步數> <x> <y>：先移過去，10M 步之後才按（`docs/re/318` §六 的坑）。
mvc() { echo "-mouse-x $2 -mouse-y $3 -mouse-at $1 -clicks $(($1 + 10000000)):$2:$3"; }
mvr() { echo "-mouse-x $2 -mouse-y $3 -mouse-at $1 -rclicks $(($1 + 10000000)):$2:$3"; }

cd "$ROOT"
if [[ "${ENM_FROM:-0}" -le 0 ]]; then
# 0. S3 帝國第一回合的戰略地圖（`docs/re/324` 的序列到 300M，第一次「系統指令」之前）。
rec enm-map 300000000 - ck \
  -tick 20000 -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -clicks 40000000:300:230,95000000:300:230,190000000:300:282,240000000:318:168 \
  -click-hold $HOLD $ENM
fi
# 1. 等級 3：列印編號、提督、士氣；點列開視窗（編號、提督、部隊種類）；確認 (488,245) 或右鍵關。
rec enm-l3        335000000 enm-map    ck -poke "$P17;$(S17 03 03 03 03)" $(mvc 310000000 490 250) -click-hold $HOLD $ENM
rec enm-l3-row    365000000 enm-l3     ck $(mvc 340000000 450 268) -click-hold $HOLD $ENM
rec enm-l3-ok     395000000 enm-l3-row -  $(mvc 370000000 488 245) -click-hold $HOLD $ENM
rec enm-l3-rclick 395000000 enm-l3-row -  $(mvr 370000000 300 200) -click-hold $HOLD $ENM
# 2. 等級 1：列是「第？艦隊 ？？？？？？？？？ ??」；點列 → 「情報資料來源不足」（模態：
#    點在框內別處不動、右鍵不動、確認 (432,230) 才關）。
rec enm-l1          335000000 enm-map    ck -poke "$P17;$(S17 01 01 01 01)" $(mvc 310000000 490 250) -click-hold $HOLD $ENM
rec enm-l1-row      365000000 enm-l1     ck $(mvc 340000000 450 268) -click-hold $HOLD $ENM
rec enm-l1-row-miss 395000000 enm-l1-row -  $(mvc 370000000 300 200) -click-hold $HOLD $ENM
rec enm-l1-row-rclk 395000000 enm-l1-row -  $(mvr 370000000 300 200) -click-hold $HOLD $ENM
rec enm-l1-row-ok   395000000 enm-l1-row -  $(mvc 370000000 432 230) -click-hold $HOLD $ENM
# 3. 混合：槽 17 四組 3／2／1／0（列照最大值 3；視窗的部隊種類只列 ≥2 的兩組）、士氣 30（<50 換色）；
#    槽 16 也搬進來、六組全 2（列是「第？艦隊」）——兩支的先後看 297B:131C 的順序。
rec enm-mix       335000000 enm-map    ck -poke "$P17;$(S17 03 02 01 00);lin:613FA@$AT=1E;$P16;$(S16 02)" $(mvc 310000000 490 250) -click-hold $HOLD $ENM
rec enm-mix-row0  365000000 enm-mix    -  $(mvc 340000000 450 268) -click-hold $HOLD $ENM
rec enm-mix-row1  365000000 enm-mix    -  $(mvc 340000000 450 288) -click-hold $HOLD $ENM
# 4. 等級 0 的不列（槽 17 搬進來但 +0x17 全 0）。
rec enm-l0        335000000 enm-map    -  -poke "$P17;$(S17 00 00 00 00)" $(mvc 310000000 490 250) -click-hold $HOLD $ENM
# 5. 兩位數編號（#74）：S3 同盟槽 28 是第１３艦隊（九組艦艇 109／111–118），同樣把 `+0x06`
#    寫成 0 搬進瓦爾哈拉、九筆 `+0x17` 寫成 3；列與視窗的編號都印半形「13」。
P28="lin:61902@$AT=00"
S28() { local a; for a in 689FF 68B37 68A9B 68AE9 68B85 68BD3 68C21 68C6F 68CBD; do printf 'lin:%s@%s=%s;' "$a" "$AT" "$1"; done; }
rec enm-n13     335000000 enm-map ck -poke "$P28;$(S28 03)" $(mvc 310000000 490 250) -click-hold $HOLD $ENM
rec enm-n13-row 365000000 enm-n13 -  $(mvc 340000000 450 268) -click-hold $HOLD $ENM
