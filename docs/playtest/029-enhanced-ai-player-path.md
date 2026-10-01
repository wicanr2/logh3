# 029 — 增強策略正常玩家路徑與存讀

工作項目：[Issue #87](https://github.com/wicanr2/logh3/issues/87)。
規格：[068](../spec/068-enhanced-ai-player-path.md)。來源調查：[357](../re/357-enhanced-ai-player-path.md)。
完整來源／輸出雜湊與重生命令摘要：
[`enhanced-ai-player-path.json`](../internals/data/enhanced-ai-player-path.json)。

## 正式 CLI 重現

入口：`bash tools/enhanced_ai_check.sh`。既有 `logh3-ebiten:2` 容器、正式 `cmd/gin3`、
正常 GUI、相同玩家輸入與執行前固定 seed `0xD3DC167E`。
每個模式（`-ai -1`／`-ai 1`）各跑兩次：主選單 → 防拷 → 關月初報告 → 玩家編隊 →
實行 → 保存第 0 格 → 讀第 0 格 → 保存第 1 格 → 下一回合 → 保存第 2 格。
未注入狀態、未以 direct-entry 進入畫面。載入後沿用正常載入器的既有種子設定，
不新增 RNG 存檔格式，不要求兩種策略消耗相同骰序。

兩次每模式的六檔、三槽完全一致。第 0／1 格的 FLEET、SHIP、STELLAR、PLANET
全表相同；兩側首回合均玩家 1／敵方 3 支，續回合均玩家 1／敵方 4 支。
強化版首回合保留一支於首都，兩支朝同一集結目標移動；原版則派向不同目標。
載入後續回合的命令仍有差異。這證明正式入口的增強策略生效，
不是將增強模式稱為原版同狀態一致，也不是 AI 必勝或全局最優保證。

收據 `workplace/playtest/ai87/receipt.json` SHA-256
`91d3ebdccdd9ffab7a6ca736ca444c722db396f998921cd5d7dd79c66696853b`；
每次輸入在 `<1|2>/<original|enhanced>/script.txt`，原始 PNG 與 state 同目錄。
正式 CLI 所有原始表、圖與腳本雜湊均記於版控 JSON，不需從對話重組命令。

## 測試範圍與限制

`TestEnhancedAINormalPlayerPath` 抽兩陣營新局與正常存讀，驗敵方實際建隊、命令不同，
載入後保留模式／難度並重接建隊來源。直接 `startNewGame` 只作額外策略暫態清理測試，
不代替上述正常入口。偏好 0–3／-1、不合法難度拒絕、現況及零戰力另有回歸。
`TestFocusedAttackUsesProductionMoveGate` 讓雙向攻勢經正式 `PlanStep`，
確認占領命令可走進指定邊界目標，未繞過正式規則。

最初測試誤比較含函式欄位的 `OriginalStrategy` 值，及使用漏載權重的舊助手，
兩次 panic 均分類為驗證環境／腳本錯誤；補齊與正式載入器相同資料後同一命令重跑。
沒有調整策略常數、正式規則或原版黃金值。

被動戰場裁決仍在：正常單人戰場的電腦不下令、不造成傷害。
本項完成只證明策略接線、實際決策與模式生命週期；不聲稱正常戰術攻擊已增強。
