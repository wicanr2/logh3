#!/usr/bin/env bash
# 正常X11玩家交戰；準備及範圍見 docs/playtest/068-fleet-player-gui.md。
# bash tools/fleet_player_gui.sh <新輸出批次> <陣營0或1> <倍率1或2>
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
[[ $# == 3 && "$1" =~ ^[A-Za-z0-9][A-Za-z0-9_-]*$ && "$2" =~ ^[01]$ && "$3" =~ ^[12]$ ]] || exit 2
[[ -d "$ROOT/workplace" && ! -e "$ROOT/workplace/$1" ]] || exit 2
for file in round79-gin3-v3 round79-input-v2 round79-capture round79-native-check "round79-plan-final-$2/plan.json"; do
  [[ -f "$ROOT/workplace/$file" ]] || { echo "缺少已準備工具：$file" >&2; exit 2; }
done
[[ "$(stat -c %u "$ROOT/workplace")" == "$(id -u)" ]] || exit 2
cd "$ROOT"
exec bash tools/gui.sh xrun bash tools/fleet_player_gui_xrun.sh "$1" "$2" "$3"
