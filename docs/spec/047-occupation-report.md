# SPEC 047 — 單方面占據的戰略報告

狀態：**CONFORMED**（本規格限定的單人玩家可見切片；逐格與對戰限制見末節）

證據：[`docs/re/306`](../re/306-occupation-report-caller-and-receipt.md)、
[`docs/re/245`](../re/245-region-flag-bits-one-and-four.md)、
[`docs/re/250`](../re/250-flip-report-routine.md)。

## 範圍與順序

1. 艦隊移動後，依 `rules.StarGraph.Occupations` 重算 32 格 bit2／bit4。
2. 新 bit4 成立的每一格排一則 `StateOccupationReport`；舊 bit4 決定是
   首次或持續，占據者相對目前座位決定我方或敵方，共四句。
3. 報告逐則模態顯示，滑鼠左右鍵都能關、座標不參與；全部關完才執行
   叛亂判定、叛亂報告與戰鬥掃描。
4. 星系名字色依事件時的星域擁有者；頭像依觀看座位的情報專任參謀。

## typed contract

- `OccupationReport` 保存 `Region`、`Owner`、`Occupier`、`Seat`、`Continuing`。
- scene owner、pending queue 與 callback 不與易主或叛亂報告混用。
- `prepareOccupationReports` 先以舊 flags 選 `Continuing`，再用
  `rules.ApplyOccupation` 寫回 bit2／bit4。
- presentation 共用 `FlipReportPanel` 幾何，但用 `DrawInline`：訊息第一行
  緊接星系名，第二行才從 `(336,202)` 起。
- 新遊戲與讀檔清空報告、pending queue 與 callback。

## 驗收

- 四種 `MessageKey` 全部有單元測試。
- 首次占據會先設 bit4；下一輪同局面改選持續句。
- 兩則報告需兩次滑鼠事件，callback 只在第二次後執行一次；左右鍵各覆蓋。
- `occupation-report3` 的首次我方占據文字矩形對原版逐像素 0 差異。
- 完整 `internal/screen`、`internal/game`、`internal/rules` 回歸通過。

驗收紀錄見 [`docs/playtest/017`](../playtest/017-occupation-report.md)。

## 證據限制

- remake 先批次重算 32 格再顯示；原版在逐格迴圈內顯示，因此停住當下的
  完整記憶體與後續叛亂 RNG 不宣稱 same-state。
- 原版對兩個非電腦座位各呼叫一次；目前只為當前 `g.faction` 顯示，
  對戰雙人座位切換另待專門收據與 scene lifecycle 規格。
