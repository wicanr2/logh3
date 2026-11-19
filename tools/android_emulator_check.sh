#!/usr/bin/env bash
# 在 Docker 內的 API35 模擬器（KVM 加速）安裝 APK、啟動、送輸入並截圖（規格161）。
#
#   tools/android_emulator_check.sh <APK> <輸出目錄> [輸入腳本]
#
# 迴圈裡的 adb 一律接 </dev/null：adb shell 會吃掉 while read 的標準輸入，後面的指令就消失了。
# 預設 20:9 的 1280×576、240 dpi：軟體 OpenGL 在 2400×1080 會讓系統介面未回應、對話框蓋住畫面。
# 解析度可用 LOGH3_EMU_W／LOGH3_EMU_H（直向寬高）、LOGH3_EMU_DPI 改。
# 輸入腳本一行一個指令：`wait 秒`、`tap x y`（裝置像素，預設橫向 1280×576）、`press x y`（按住 300 ms）、`back`、`shot 名字`、
# `waitlog 字串`（等 logcat 出現該字串，最多 120 秒；用來等「載入完成」再送輸入）、
# `restart`（強制停止 App、清 logcat、重新啟動；驗證存檔跨啟動保留）。
# 不給腳本時只開機、安裝、啟動，每 10 秒截一張共 9 張。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${LOGH3_EMULATOR_IMAGE:-logh3-android-emulator:api35-20261119-r1}"
apk="$(cd "$(dirname "${1:?APK}")" && pwd)/$(basename "$1")"
out="${2:?輸出目錄}"; mkdir -p "$out"; out="$(cd "$out" && pwd)"
script="${3:-}"
[[ -f "$apk" ]] || { echo "缺 $apk" >&2; exit 2; }
[[ -z "$(ls -A "$out")" ]] || { echo "$out 不是空目錄，拒絕覆寫" >&2; exit 2; }
extra=()
if [[ -n "$script" ]]; then script="$(cd "$(dirname "$script")" && pwd)/$(basename "$script")"; extra+=(-v "$script:/input.txt:ro"); fi
kvm_gid="$(getent group kvm | cut -d: -f3)"
exec timeout 1800 docker run --rm --network none --memory 6g --cpus "${LOGH3_EMU_CPUS:-6}" --pids-limit 2048 \
  --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
  --device /dev/kvm --group-add "$kvm_gid" \
  -v "$apk:/app.apk:ro" -v "$out:/out" "${extra[@]}" "$IMAGE" bash -c '
set -euo pipefail
mkdir -p /out/home/.android /out/avd /out/shots
echo no | avdmanager create avd -n logh3 -k "system-images;android-35;default;x86_64" -d pixel_7 --force >/out/avdmanager.log 2>&1
cfg=/out/avd/logh3.avd/config.ini
W=${LOGH3_EMU_W:-576}; H=${LOGH3_EMU_H:-1280}; D=${LOGH3_EMU_DPI:-240}
sed -i "s/^hw.lcd.width=.*/hw.lcd.width=$W/; s/^hw.lcd.height=.*/hw.lcd.height=$H/; s/^hw.lcd.density=.*/hw.lcd.density=$D/; s/^hw.ramSize=.*/hw.ramSize=4096/" "$cfg"
grep -q "^hw.ramSize" "$cfg" || echo "hw.ramSize=4096" >> "$cfg"
emulator -avd logh3 -no-window -no-audio -no-snapshot -no-boot-anim -gpu swiftshader_indirect \
  -accel on -memory 4096 -cores 4 >/out/emulator.log 2>&1 &
t0=$(date +%s)
adb wait-for-device
until [ "$(adb shell getprop sys.boot_completed 2>/dev/null | tr -d "\r")" = 1 ]; do sleep 2; done
echo "開機 $(( $(date +%s) - t0 )) 秒" | tee /out/timeline.txt
adb shell settings put system accelerometer_rotation 0
adb shell settings put system user_rotation 1
adb shell settings put secure immersive_mode_confirmations confirmed
for k in window_animation_scale transition_animation_scale animator_duration_scale; do adb shell settings put global $k 0; done
adb install -r /app.apk >/out/install.log 2>&1
echo "安裝完成 $(( $(date +%s) - t0 )) 秒" | tee -a /out/timeline.txt
adb logcat -c
adb shell am start -n tw.logh3.app/.MainActivity >/dev/null
s0=$(date +%s)
shot() { adb exec-out screencap -p > "/out/shots/$1.png" </dev/null; echo "$(( $(date +%s) - s0 )) 秒 截圖 $1" | tee -a /out/timeline.txt; }
if [ -f /input.txt ]; then
  while read -r cmd a b; do
    case "$cmd" in
      wait) sleep "$a" ;;
      waitlog) for i in $(seq 1 120); do adb logcat -d </dev/null | grep -q "$a" && break; sleep 1; done
               echo "$(( $(date +%s) - s0 )) 秒 看到 $a" | tee -a /out/timeline.txt ;;
      tap) adb shell input tap "$a" "$b" </dev/null; sleep 1 ;;
      press) adb shell input swipe "$a" "$b" "$a" "$b" 300 </dev/null; sleep 1 ;;
      restart) adb shell am force-stop tw.logh3.app </dev/null; adb logcat -c </dev/null
               adb shell am start -n tw.logh3.app/.MainActivity </dev/null >/dev/null; echo "$(( $(date +%s) - s0 )) 秒 重新啟動" | tee -a /out/timeline.txt ;;
      back) adb shell input keyevent KEYCODE_BACK </dev/null; sleep 1 ;;
      shot) shot "$a" ;;
      ""|\#*) ;;
    esac
  done < /input.txt
else
  for i in 1 2 3 4 5 6 7 8 9; do sleep 10; shot "t$i"; done
fi
adb logcat -d > /out/logcat.txt 2>&1 || true
adb emu kill >/dev/null 2>&1 || true
wait || true
'
