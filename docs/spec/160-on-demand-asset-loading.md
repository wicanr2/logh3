# 160：素材按需載入與背景預熱

狀態：**第一階段 CONFORMED**（2026-11-19，第174輪）。第二階段（遊戲中首次使用的停頓）為 DRAFT。
驗收：[146](../playtest/146-on-demand-asset-loading.md)。相關：[158](158-scout-formation-lazy-hd.md)（人物圖、字型頁、海報的首用解碼與快取）。

這是 remake 自有的載入架構，不改原版規則、畫面內容或存檔格式。

## 問題

使用者回報 Windows、Linux AppImage、macOS 開啟等待過久，載入視窗、載入完成後、遊戲中切換畫面都有。
另外規劃 Android 版，手機的儲存與 CPU 都比桌機慢，記憶體也少。

v1.6.5 啟動時一次做完的工作（發行版素材，`TestStartupBudget` 量測）：

| 工作 | 讀取 | 說明 |
|---|---:|---|
| 27 張命令階段畫面規格 JSON | 4.7 MB | `LoadCommandScreens` 全部解析，佔 `bootstrap.Load` CPU 約 70% |
| 兩份 HD 字型索引 | 4.4 MB | 明體、黑體各 2.2 MB、約一萬六千字 |
| 戰場投影地板 | 0 | 1872×672 的三角函數運算，只有戰場投影用得到 |
| 28 段音效 OGG | 0.35 MB | 每段完整解碼一遍，只為驗證沒壞 |
| 海報底圖與艦影 | 6.6 MB | 載入完成後第一格才在畫面執行緒同步讀檔、驗雜湊、解 PNG |

## 契約

1. 主選單之前只做一定要用的：原版資料、主選單到月初報告會用到的規格、開場動畫、HD 人物索引與預設字型、存檔索引。
2. 其餘素材首次使用才讀，或在主選單出來之後由一條背景 goroutine 預熱。預熱順序照玩家最先碰到的排。
3. 按需資源同一個鍵只建立一次，之後回同一個指標（畫面程式會改規格裡的列與字碼）。
4. 背景與畫面執行緒並行時不得有資料競爭：共用表上鎖，`sync.Once`，或只交出解好的物件。
   背景 goroutine 不讀 `app` 的欄位，要用的對象在畫面執行緒先取好。
5. 原本在啟動時把關的錯誤不能消失：
   - 規格檔缺檔仍在啟動時報錯（只 `stat`）。JSON 格式改由 `TestCommandScreensAllParse` 逐張解析把關。
   - 音效壞檔改由背景驗證發現。下一次播放回錯誤，`updateEffects` 照原規則關掉音效並印出原因。
   - 字型圖集載入失敗時維持原字型、回報錯誤，之後不再列出那一款。
6. 按需失敗只印一次，不每幀重試，不 panic。
7. 無頭截圖（`-shot`）倒的是遊戲 framebuffer，不受背景預熱影響。

## 第一階段（已完成）

| 項目 | 做法 | 位置 |
|---|---|---|
| 命令階段畫面規格 | `game.ScreenSet`，`Get` 首次才解析；`Warm` 背景補齊 | `internal/game/commandphase.go` |
| 星域地圖 | `game.StarMapSet`，依 `map.colours` 首次才建 | 同上 |
| 投影地板 | `sync.Once`，`Renderer.Warm` 背景補算 | `internal/presentation/projection.go` |
| 黑體字型圖集 | 只登記目錄，`SetFontStyle("sans")` 才載入 | `internal/presentation/typography.go` |
| 音效 | 啟動只確認 28 段存在；背景完整解碼並留 PCM，之後播放不讀檔 | `cmd/gin3/effects.go` |
| 海報底圖與艦影 | 載入尾段開始背景解碼，畫面執行緒只上傳解好的圖；未就緒時塗黑、遊戲照畫 | `cmd/gin3/poster_approved.go`、`warmup.go` |

## 第二階段（DRAFT）：遊戲中首次使用的停頓

「遊戲中切換畫面」的停頓還沒有逐項量測。候選，量到再定案：

- HD 人物圖首次顯示時解碼（規格158，首用解碼、64 MiB LRU）。
- 進戰場時的艦型動畫、地形遮罩與交戰動畫 `COMANI.BMG`。
- 啟動時仍然先載的五張大規格：`screen-admiral-list`（420 KB）、`screen-build-plan`（300 KB）等，
  它們是 `Resources` 的固定欄位，要改成按需得換存取方式。
- 原版資料裡結局 VO5、音效 VOC、3D 艦隊圖塊等解碼，量到的成本都在 10 ms 以下，暫不處理。

量測方法：`-frame-trace` 記每格 `draw_ms`，正常玩家路徑走月初報告、開清單、進戰場，抓超過 50 ms 的格。
Android 版的記憶體上限另訂（快取淘汰、背景預熱是否全做），見[規格161](161-android-port-plan.md)。
