# 第174輪：素材按需載入與 Android 規劃

使用者要求做素材載入效能改善，需要時再讀取，避免開啟等待過久；並一併規劃 Android 版。
起點main `b0b3dbd7`，工作樹乾淨。

使用者回答（2026-11-19）：等待發生在 Windows、Linux AppImage、macOS 三個平台，載入視窗、載入完成後、
遊戲中切換畫面都有；範圍選「先量再全面按需」。

## 目前狀態表

| 項目 | 狀態與契約 |
|---|---|
| 量測 | PASS，`TestStartupBudget` 逐段量耗時、CPU 與讀取位元組；A/B 在同一容器交替跑 |
| 畫面規格與星圖按需 | PASS，`ScreenSet`／`StarMapSet`，缺檔仍在啟動時報錯，格式由測試逐張把關 |
| 投影地板、黑體圖集 | PASS，首次使用才算或載入 |
| 音效 | PASS，啟動只確認檔案在；背景完整解碼並快取 PCM，壞檔照原規則關音效 |
| 海報第一格停頓 | PASS，背景解碼；未就緒時塗黑 |
| 結果 | 擋住主選單的 CPU 少 84%–88%，讀取 22.5 MB → 8.5 MB；實際程式到「載入完成」793 → 278 ms |
| 遊戲中停頓 | 未做，規格160第二階段 DRAFT |
| Android | 規劃 DRAFT（規格161）。發行已裁決：完整版內含原版檔供自己玩、只留本機；release 版 App 內匯入。觸控對應待裁決 |
| 提交 | 使用者同意 commit 與 push main |

## 文件

- 規格：[160](../spec/160-on-demand-asset-loading.md)、[161](../spec/161-android-port-plan.md)
- 驗收：[146](../playtest/146-on-demand-asset-loading.md)
