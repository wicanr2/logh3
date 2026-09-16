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
# 6. 從 `dec1` 只把游標移到瓦爾哈拉（不按）→ 左下角照游標重寫：剛編成的
#    第１艦隊算不算進「艦隊 a：b」（星域記錄 `+0x28` 建艦隊時就加了）。
rec fcmd-hover-home 440000000 fcmd-base - -mouse-x 32 -mouse-y 32 -mouse-at 421000000
