# 146 — 素材按需載入：量測與驗收

日期：2026-11-19。第174輪。規格 [160](../spec/160-on-demand-asset-loading.md)。

## 比較等級

remake 自身的效能量測，不涉及原版對拍。改的是載入時機，畫面內容、規則與存檔格式不變。

## 環境

- 素材：v.1.6.5-20261102 Linux 發行版解開的 `usr/bin`（`workplace/release172/extracted/squashfs-root`），唯讀掛載。
- 基準：main `b0b3dbd7` 的 git worktree，加上同一支量測測試的舊版本（量舊流程：音效完整驗證、第一格同步解海報）。
- 候選：本輪工作樹。
- 容器：`logh3-ebiten:20261030-r169`，4 CPU、6 GB。主機同時有其他工作，load average 10 到 24，
  所以用同一容器內**交替**跑基準與候選，看相對差距，不看絕對秒數。

## 量測一：主選單前擋住玩家的工作（`TestStartupBudget`，A/B 交替各 5 次）

「擋住」＝載入視窗期間加上載入完成後第一格同步做的事。候選另列背景預熱，不擋畫面。

| 次 | 基準 CPU | 候選 CPU | 候選背景 CPU |
|---:|---:|---:|---:|
| 1 | 843 ms | 106 ms | 612 ms |
| 2 | 783 ms | 102 ms | 643 ms |
| 3 | 969 ms | 134 ms | 810 ms |
| 4 | 904 ms | 137 ms | 693 ms |
| 5 | 837 ms | 105 ms | 634 ms |

讀取位元組（決定性，五次相同到個位數）：基準 22,485,689，候選 8,548,497，少 62%。
擋住的 CPU 少 84% 到 88%。原始輸出 `workplace/perf174/ab-2.txt`，SHA-256 `a7da5364…93f4e0a9`。
第一輪 A/B（`ab-1.txt`，音效檢查與黑體圖集還在啟動時）擋住的 CPU 少 46% 到 70%。

候選版剩下的擋住工作：原版資料 64 ms（3.9 MB）、HD 人物索引與明體圖集 88 ms（2.8 MB）、開場動畫 3 ms（1.8 MB）。

## 量測二：實際程式的載入時間軸

`gin3 -graphics hd -poster -poster-set v1`，Xvfb 加 PulseAudio null sink（讓程式認得音效裝置），stderr 逐行加時間戳。

| 階段 | v1.6.5 | 候選 |
|---|---:|---:|
| 原版資料 | 133 ms | 40 ms |
| HD 圖像與字型 | 232 ms | 44 ms |
| 音樂與音效 | 203 ms | 9 ms |
| 程序啟動到「載入完成」| 793 ms | 278 ms |

日誌 `workplace/perf174/tl-base/log.txt`（SHA-256 `d8e031fe…`）、`tl-cand/log.txt`（`c28f44da…`）。

## 量測三：載入完成後的畫面

舊版第一格同步解海報底圖（6 MB PNG）與艦影，在慢的情況下量到 0.5 到 1.3 秒的停頓。候選版改到背景解碼。

逐秒截圖中，候選版有一張剛切換時的雜訊畫面。底圖未就緒時加上塗黑之後，以約 0.05 秒間隔重截，
兩版各 40 張都沒有雜訊，海報在載入完成後很快出現。那張雜訊是否全由未塗黑造成沒有定論
（Xvfb 軟體繪圖下，視窗剛放大的內容本來就未定義），塗黑沒有副作用所以保留。
最終候選執行檔 `workplace/perf174/gin3-cand`，SHA-256 `5c36f760…19393e1a`（時間軸量測用的是加塗黑之前的建置，載入程式碼相同）。

## 自動測試

| 測試 | 內容 |
|---|---|
| `TestCommandScreensAllParse` | 27 張命令階段規格逐張解析、星域地圖逐組建立；同名回同一指標；不存在的名字回 nil |
| `TestWarmResourcesWhilePlaying` | 背景預熱與畫面操作並行，`-race` 無競爭，預熱後規格指標不變 |
| `TestHDLoadDefersGridAndSansAtlas` | HD 載入不算投影地板、不解析黑體；`Warm` 補算；選黑體才載入且生效 |
| `TestEffectWarmCachesPCMAndFlagsBroken` | 背景解好的 PCM 與串流解碼逐位元組相同；截斷一段時背景抓到，`PlayEffect` 回錯誤（`-race`）|
| `TestApprovedPosterPrefetch` | 底圖背景解碼、只交出一次、換組回來能重解（`-race`）|
| `TestStartupBudget` | 量測用，設 `LOGH3_STARTUP_BUDGET` 才跑 |

標準 `tools/gui.sh test -p 2 -count=1 ./cmd/... ./internal/...`：16 PASS、5 無測試、0 失敗。
塗黑修改之後 `cmd/gin3`、`internal/presentation` 另跑一次通過。

## 還沒驗的

- Windows 與 macOS 實機沒有量；AppImage 冷啟動（squashfs 解壓）沒有量，這台主機無法清頁快取。
- 遊戲中首次開畫面、進戰場的停頓沒有逐項量（規格160第二階段）。
- 正常玩家路徑只看了開機到主選單的截圖；遊戲中的畫面由既有測試涵蓋，本輪沒有重新目視。
