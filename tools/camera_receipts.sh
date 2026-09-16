#!/usr/bin/env bash
# 戰場視野的滑鼠捲動收據（`docs/re/326`、規格 023 的補充）。
#
#   tools/camera_receipts.sh              # 全部重錄（約 2 分鐘；檢查點 eng-entry 要先在）
#
# 載具是 `tools/engagement_receipts.sh` 的 `eng-entry`（S3 帝國，達希利戰場第 1 回合，
# #900M）。原版在戰場視窗上**按住右鍵**就捲動視野：每一圈重讀游標、算它相對於
# 滑鼠座標 (224,180) 的方位（`2BE1:011D` → `3982:0022` 八個方向）、視野往那邊走一格
# （`1588:03A9`，格 0–100 × 0–84），放開才停；游標停在戰場邊緣不按什麼都不動。
# 每批倒鏡頭 ds:9BF4／9BF6（lin 5E724）與格座標 ds:9BA3／9BA4（lin 5E6D3）。
#
# ⚠ 收據綁執行器版本（AGENTS.md）；座標是滑鼠座標（畫面 y ＝ 滑鼠 y ＋ 40）。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
CK="$GOLEM/workplace/fcmd"
CAM="-peek lin:5E724:4,lin:5E6D3:2"
T0=900000000

rec() {
  local name=$1 steps=$2; shift 2
  tools/oracle.sh "$name" -load-state /src/workplace/fcmd/eng-entry.state -steps "$steps" "$@" >/dev/null
  printf '%s  %s  %s\n' "$name" "$(sed -n 's/^screen.bin SHA-256: //p' "$ROOT/workplace/oracle/$name/meta.txt" | cut -c1-16)" \
    "$(grep -h 'lin:5E724' "$ROOT/workplace/oracle/$name/report.txt" | sed 's/.*）//')"
}
# hold <批次> <x> <y> <按住步數> [看的步數]：游標先搬到 (x,y)，#901M 按下右鍵按住 N 步，#906M 看
# （按得比那久的要把看的步數往後挪，不然看到的是捲到一半的畫面）。
hold() { rec "$1" "${5:-906000000}" -mouse-x "$2" -mouse-y "$3" -mouse-at $((T0 + 500000)) -rclicks $((T0 + 1000000)):"$2":"$3" -click-hold "$4" $CAM; }

cd "$ROOT"
[[ -f "$CK/eng-entry.state" ]] || { echo "缺 eng-entry.state：先跑 tools/engagement_receipts.sh"; exit 1; }
# 0. 基準：只把游標搬到 (400,180)。
rec cam-base 906000000 -mouse-x 400 -mouse-y 180 -mouse-at $((T0 + 500000)) $CAM
# 1. 游標停在戰場四個邊緣不按：畫面不動（-watch-screen 一行都不該印）。
rec cam-edge-e 906000000 -mouse-x 462 -mouse-y 200 -mouse-at $((T0 + 500000)) -watch-screen 500000 -screen-delta 100 $CAM
rec cam-edge-w 906000000 -mouse-x 17  -mouse-y 200 -mouse-at $((T0 + 500000)) -watch-screen 500000 -screen-delta 100 $CAM
rec cam-edge-n 906000000 -mouse-x 240 -mouse-y 21  -mouse-at $((T0 + 500000)) -watch-screen 500000 -screen-delta 100 $CAM
rec cam-edge-s 906000000 -mouse-x 240 -mouse-y 378 -mouse-at $((T0 + 500000)) -watch-screen 500000 -screen-delta 100 $CAM
# 2. 按住右鍵 0.5M 步：八個方向（相對於 (224,180)）。
hold cam-e  400 180 500000
hold cam-w  100 180 500000
hold cam-s  224 350 500000
hold cam-n  224  30 500000
hold cam-se 400 350 500000
hold cam-sw 100 350 500000
hold cam-ne 400  30 500000
hold cam-nw 100  30 500000
# 3. 極短（0.1M）與很長（12M，撞到右邊界 100）；正好在 (224,180) 上按。
hold cam-tap  400 180 100000
hold cam-long 400 180 12000000 920000000
hold cam-zero 224 180 500000
# 方位的原點是 (224,180) 不是戰場正中央 (240,200)：(324,230) 對前者是 26.6°（東南），對後者 19.7°（東）。
hold cam-anchor 324 230 500000
# 4. 戰術選單開著時按住右鍵：只關選單、不捲（左鍵 (240,200) 先開選單）。
rec cam-menu 906000000 -clicks $((T0 + 200000)):240:200 -mouse-x 400 -mouse-y 180 -mouse-at $((T0 + 500000)) -rclicks $((T0 + 1000000)):400:180 -click-hold 500000 $CAM
