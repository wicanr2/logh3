#!/usr/bin/env bash
# 僅在tools/gui.sh容器內：二倍視窗用可容納1280x960的獨立Xvfb。
set -euo pipefail
gui_x_pid=''
trap 'if [[ -n "$gui_x_pid" ]]; then kill "$gui_x_pid" 2>/dev/null || true; wait "$gui_x_pid" 2>/dev/null || true; fi' EXIT
[[ -d /src && -d /orig && $# == 3 ]] || exit 2
if [[ "$3" == 2 ]]; then
  Xvfb :100 -screen 0 1600x1200x24 -nolisten tcp > /tmp/logh3-fleet-xvfb.log 2>&1 &
  gui_x_pid=$!
  export DISPLAY=:100
  for _ in {1..50}; do if xset q >/dev/null 2>&1; then break; fi; sleep 0.1; done
  xset q >/dev/null
fi
python3 tools/scripts/fleet_player_gui_driver.py \
  --binary /src/workplace/round79-gin3-v3 \
  --plan "/src/workplace/round79-plan-final-$2/plan.json" \
  --out "/src/workplace/$1" --scale "$3"
/src/workplace/round79-native-check -dir "/src/workplace/$1"
