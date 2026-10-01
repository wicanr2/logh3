#!/bin/sh
# #69 原生 FM 前置拒絕驗證；只在 tools/audio.sh 的容器內執行。
set -eu
base=/work/music69-fm
program=$base/music-import
check=$base/check
mkdir -p "$check"

reject() {
    name=$1
    shift
    if "$program" "$@" > "$check/$name.log" 2>&1; then
        echo "$name：錯誤輸入被誤報成功" >&2
        exit 1
    fi
    grep -q '音樂匯入失敗' "$check/$name.log"
    test ! -e "$base/rejected-$name"
    echo "$name：正確拒絕，未建立輸出"
}

reject exclusive -gm-only -fm-only -source /orig -output "$base/rejected-exclusive"
reject track -fm-only -tracks BGM_8F -source /orig -output "$base/rejected-track"
reject tools -fm-only -tracks BGM_0F -source /orig -output "$base/rejected-tools" -tools "$check/no-tools"
reject original-output -fm-only -tracks BGM_0F -source /orig -output /orig/forbidden
test ! -e /orig/forbidden

# 錯 probe 是當輪匯入器本身；這只驗身份拒絕，不將它當原版執行器。
wrong=$check/wrong-probe
mkdir -p "$wrong"
ln -sf "$program" "$wrong/dosgolem-probe"
ln -sf /usr/local/bin/opl2wav "$wrong/opl2wav"
ln -sf /usr/bin/ffmpeg "$wrong/ffmpeg"
reject probe -fm-only -tracks BGM_0F -source /orig -output "$base/rejected-probe" -tools "$wrong"
grep -q 'dosgolem 必須為 clean' "$check/probe.log"

# 空目錄只驗缺原版檔案的拒絕；沒有執行、合成或替代原版收據。
mkdir -p "$check/missing-original"
reject source -fm-only -tracks BGM_0F -source "$check/missing-original" -output "$base/rejected-source" -tools "$base/tools"
grep -q 'SHA-256 不符或缺檔' "$check/source.log"
