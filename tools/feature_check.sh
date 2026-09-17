#!/usr/bin/env bash
# 功能選項驗證：每個選項都要有作用、怎麼點都不能 crash（使用者 2026-09-17）。
#
#   tools/feature_check.sh              # 巡覽（命令列 12 項、艦隊視窗與兩個選單、戰場）＋ 存讀一圈 ＋ 隨機點擊
#   STEPS=20000 tools/feature_check.sh  # 隨機點擊每組步數（預設 3000）；20000 約 16 分鐘
#   LOGH3_TOUR_ALL=1 tools/feature_check.sh  # 巡覽跑帝國／同盟／對戰三個起點、戰場全格點
#
# 走正式入口（bootstrap.Load，與 cmd/gin3 同一份資源），需要原版素材 org_game/extracted；沒有就 skip。
# 存進度寫到測試暫存目錄，原版目錄一律唯讀。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
LOGH3_TOUR=1 LOGH3_MONKEY_STEPS="${STEPS:-3000}" LOGH3_TIMEOUT="${LOGH3_TIMEOUT:-120m}" \
  tools/go.sh test ./internal/bootstrap -run 'TestEveryCommandItemWorks|TestDeepScreensNoPanic|TestSaveThenLoadRoundTrip|TestMonkeyNoPanic' \
  -v -timeout 120m 2>&1 | grep -E '^(---|\s+---|ok|FAIL|panic)|tour_test|monkey_test|saveload_test' 
