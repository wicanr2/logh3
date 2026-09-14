# SPEC 046 — 逐格叛亂的戰略報告

狀態：**CONFORMED**（本規格限定的玩家可見切片；完整記憶體時序限制見末節）

證據：[`docs/re/292`](../re/292-rebellion-decision-and-effect.md)（`056A:06F3`
反組譯、`rbx292`／`rbx292d` 執行期收據）、
[`docs/re/247`](../re/247-strategy-report-strings.md)（`ds:085D`／
`ds:086F` 字串）、[`docs/spec/034`](034-region-flip-report.md)
（同一個 `3826:0307` 報告 compositor）。

原版：`GIN3.COM`，SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。  
原版位址空間：16-bit DOS segment:offset；畫面收據是 640×480 色號陣列。  
執行器：dosgolem-logh3 `e0b6a03`（`rbx292`）。

## 範圍

本規格只負責 `056A:06F3` 的叛亂報告：

1. `056A:0468` 逐格判定時，新的 `+0x13` bit1 會呼叫
   `056A:06F3(星域記錄)`。
2. 第一行是 `ds:085D`：`〈星域名〉星系`；名字字色取
   `ds:6096[星域擁有者]`，`星系` 固定色 7。
3. 第二行是 `ds:086F`：`發生叛亂。`，固定色 7。
4. 頭像是 `FACE.AMG[ds:8F4D[星域擁有者]]`，落點與裁切沿用
   SPEC 034 的 `(544,180)`、左上 72×93。
5. 報告是模態的；`3765:006C` 只認滑鼠事件碼 1／2，完全不看座標。
   一次只收一則，全部關掉才可進戰鬥掃描與回合尾巴。

不在本規格內：侵略進行中、提督生死、勝敗與費沙事件頁；它們仍由
worklist `strategy-report-screen` 追蹤。

## typed contract

- 新增獨立 `StateRebellionReport`，不可借用易主報告的 transaction state。
- 佇列元素保存星域編號與**叛亂效果前的擁有者**；效果會先把星域改成
  無主，若 Draw 時重讀目前 owner，名字會誤用色 7、頭像也會選錯。
- presentation 共用 `FlipReportPanel` 已證實的幾何與 glyph primitive，
  新增 `rebellion` 訊息 key。
- dismiss callback 只在佇列清空時呼叫一次，接回原本的
  contested-region 戰鬥掃描。
- 開新遊戲與讀檔必須清空佇列及 callback，避免跨局殘留。

## 驗收

- `TestRebellionReportTextMatchesOracle`：`亞爾迪巴朗星系／發生叛亂。`
  的文字矩形對 `rbx292/screen.bin` 逐像素 0 差異。
- `TestRebellionReportsQueueAndResume`：兩則報告需兩次滑鼠事件，
  callback 只在第二次後執行一次；左右鍵各覆蓋一種。
- 完整 `internal/screen`、`internal/game` 回歸通過。

驗收紀錄見 [`docs/playtest/016`](../playtest/016-rebellion-report.md)。

## 證據限制

原版在 `056A:0468` **判到一格叛亂便暫停**，後續星域尚未擲亂數；
現有 remake 的 `rules.DecideRebellions` 已先算完整圈，再顯示報告。
玩家可見的畫面、逐則模態阻塞與後續入口可依本規格驗證，但報告停住當下的
全機記憶體與 RNG 位置仍不同，因此本切片不得宣稱完整 `same-state`
記憶體 parity。要消除此限制，需另把判定器改成可續跑 iterator，不能在本
UI 切片中猜補。
