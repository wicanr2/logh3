#!/usr/bin/env bash
# #79 來源與人類結果的原版重播；入口見 docs/re/369。
# 只控制既有 Docker 包裝器，不在主機執行分析。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
mode="${1:?用法: bash tools/battle_preparation_lifecycle_oracle.sh <模式> <新批次名>}"
name="${2:?必須指定新批次名}"
[[ "$name" =~ ^battle79-life-[a-z0-9-]+$ ]] || { echo "批次名稱不符" >&2; exit 1; }
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -d "$GOLEM"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
if [[ "$mode" == boot ]]; then
  bash "$ROOT/tools/oracle.sh" "$name" -steps 60000000 -tick 20000 \
    -click-x 320 -click-y 200 -click-at 15000000 \
    -mouse-x 620 -mouse-y 460 -mouse-at 30000000 \
    -peek lin:5EF5E:4,lin:5EF70:4 \
    -dump-mem-at "59999999:lin:0:1048576:/out/oracle/$name/mem.bin"
  exit
fi
test -f "$GOLEM/workplace/fcmd/eng-t5.state"
[[ "$(sha256sum "$GOLEM/workplace/fcmd/eng-t5.state" | cut -d' ' -f1)" == c6ce5af3a8d52acc415ac1029a29a0905e9c95e02ac0e0ba547e71300fd947c1 ]] || exit 1
args=(-load-state /src/workplace/fcmd/eng-t5.state
  -clicks 1140000000:240:200,1195000000:240:200,1210000000:240:200
  -click-hold 3000000 -regs-from 1211800000 -regs-max 4
  -regs-at 181B:003F,1930:019D,1930:0AA2,1930:0D99,341B:02E5,1AAB:006D,1AAB:08FB,1AAB:0A2A,1261:000B,1261:0138,3826:0405,3765:006C,181B:0083,181B:00FA)
seed='lin:5E00C@1211893242=78 56 34 12'
case "$mode" in
  new-game|load-s3|load-existing|battle-entry)
    args=(-tick 20000 -regs-max 12 -regs-at 12AA:000E,1930:019D,1AAB:0190,1AAB:006D
      -watch 5EF5E-5EF73)
    if [[ "$mode" == load-existing ]]; then
      test -f "$GOLEM/workplace/fcmd/fcmd-base.state"
      args+=(-load-state /src/workplace/fcmd/fcmd-base.state
        -clicks 421000000:544:85,441000000:560:105,471000000:300:282,511000000:318:168
        -click-hold 3000000)
      start=420000001; end=550000000
    else
      args+=(-press 1,3,8,enter -press-at 105000000 -press-every 2000000 -click-hold 3000000)
      start=100000000; end=300000000
      if [[ "$mode" == new-game ]]; then
        args+=(-clicks 40000000:300:170,95000000:300:170 -rclicks 130000000:300:200,155000000:300:200)
      else
        clicks=40000000:300:230,95000000:300:230,190000000:300:282,240000000:318:168
        if [[ "$mode" == battle-entry ]]; then
          clicks+=,300000000:544:85,330000000:560:85,405000000:544:85,435000000:560:85,510000000:544:85,540000000:560:85,615000000:544:85,645000000:560:85,720000000:544:85,750000000:560:85,825000000:544:85,855000000:560:85
          end=900000000
        fi
        args+=(-clicks "$clicks")
      fi
    fi
    poke="lin:5E00C@$start=78 56 34 12;lin:5EF5E@$start=EC 00 8B 66;lin:5EF70@$start=3A 01 8B 66"
    bash "$ROOT/tools/oracle.sh" "$name" "${args[@]}" -steps "$end" -poke "$poke" \
      -dump-mem-at "$((start+1)):lin:0:1048576:/out/oracle/$name/mem-in.bin;$((end-1)):lin:0:1048576:/out/oracle/$name/mem-out.bin" \
      -peek lin:5EF5E:4,lin:5EF70:4,lin:61AE4:4,lin:5E00C:4
    exit
    ;;
esac
case "$mode" in
  supply-panel|supply-end)
    poke="$seed;lin:68386@1211893242=05;lin:6118A@1211893242=05;lin:68390@1211893242=D0 07 00 00 D0 07 00 00;lin:68120@1211893242=00 00 00 00 00 00 00 00"
    end=1212045378
    [[ "$mode" == supply-end ]] && end=1213547865
    ;;
  repair-trace|repair-panel|repair-end)
    poke="$seed;lin:68386@1211893242=06;lin:6118A@1211893242=06;lin:68390@1211893242=14 00 00 00 00 00 00 00;lin:6811A@1211893242=32 00;lin:6811E@1211893242=4B 00;lin:68120@1211893242=00 00 00 00 00 00 00 00"
    end=1216000000
    [[ "$mode" == repair-panel ]] && end=1212415407
    [[ "$mode" == repair-end ]] && end=1213521898
    ;;
  morale-trace|morale-panel|morale-end)
    poke="$seed;lin:6115C@1211893242=01"
    end=1240000000
    [[ "$mode" == morale-panel ]] && end=1213517111
    [[ "$mode" == morale-end ]] && end=1231561135
    ;;
  *) echo "未知模式 $mode" >&2; exit 1 ;;
esac
bash "$ROOT/tools/oracle.sh" "$name" "${args[@]}" -steps "$end" -poke "$poke" \
  -dump-mem-at "1211893243:lin:0:1048576:/out/oracle/$name/mem-in.bin;$((end-1)):lin:0:1048576:/out/oracle/$name/mem-out.bin" \
  -peek lin:6811A:14,lin:68390:8,lin:5EF5E:4,lin:5EF70:4
