# SPEC 044 — 400 列結局 VO5 的垂直位置

狀態：**CONFORMED**（2026-09-14；解碼層與產品顯示層 plane 2／3 均為 0 差異）

證據：[`docs/re/301`](../re/301-ending-image-y-is-40.md)、
[`docs/internals/18`](../internals/18-vo5.md)

## 範圍

只裁決現有 `ShowEnding(ENDT30.VO5)` 測試涵蓋的 640×400 圖片位置，以及
新版原版收據可獨立驗證的 plane 2／3。不聲明完整結局播放、文字、轉場、
palette 或 plane 0／1 合成已完成。

## 契約

- 邏輯畫布是 640×480。
- 640×400 VO5 的左上角是 `(0,40)`，即上下各留 40 列。
- 繪製、測試取樣與任何命中這張圖的幾何都使用同一個
  `screen.EndingImageY`；不得在測試另留舊座標 80。
- 圖片像素的四個位元平面不在本項重排；`asset.DecodeVO5` 的輸出原樣寫入。

## 驗收

- `internal/asset.TestVO5MatchesOriginalScreen`：`ENDT30.VO5` 解碼後的
  plane 2／3 在原版 `ending-empire` 的 `(0,40)` 逐位元組相同。
- `internal/game.TestEndingScreenMatchesOriginalPlanes`：產品
  `ShowEnding → Frame` 路徑的 plane 2／3 在同一區域逐位元組相同。
- 兩項都缺原版素材／收據即 skip，不製造替代 oracle。

## 實作與結果

`screen.EndingImageY` 是唯一座標來源，已由 80 改為 40；`screen.Ending`
以它繪製，`internal/game` 測試也以同一常數取樣。`internal/asset` 因反向
匯入 `screen` 會形成循環，只在測試內明寫由本規格固定的 40。

兩項驗收測試均通過；完整 `go test ./internal/asset ./internal/game` 也通過。
