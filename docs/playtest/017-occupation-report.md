# 017 — 單方面占據報告驗收

日期：2026-09-14  
比較等級：正常玩家路徑加明示狀態注入的局部 `same-state` 畫面比較；
不涵蓋完整逐格記憶體時序。

## 原版

- 原版：`GIN3.COM` SHA-256
  `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`
- 執行器：dosgolem-logh3 `e0b6a03`，工作樹乾淨
- 批次：`workplace/oracle/occupation-report3/`
- 輸入：新遊戲正常路徑 → 第一次「實行」；完整旗標在 `flags.txt`
- 注入：`#217000000` 設最小有效帝國艦隊槽於同盟星域 16；精確位址與
  bytes 見 [`docs/re/306`](../re/306-occupation-report-caller-and-receipt.md)
- 原版輸出：640×480 色號陣列；`screen.bin` SHA-256
  `fa690ac9828134d72e4b3c7e9b6aca7af918943bdfc2f0d0df7154dc0cb0e930`

## remake 比較

`TestOccupationReportTextMatchesOracle` 以產品使用的 `GRAPH.IMG` 字型與
`FlipReportPanel.DrawInline` 重畫「亞斯提星系已被攻入／目前轉移至占領狀態。」；
比較 `(334,178)`–`(534,239)` 共 12,462 像素，結果 **0 差異**。

行為測試另覆蓋：

- 四句選擇；
- 舊 bit4 決定首次／持續；
- bit4 先寫回再顯示；
- 左、右鍵逐則關閉，全部關完 callback 才執行；
- 占據報告先於叛亂與戰鬥掃描。

完整回歸：`internal/screen`、`internal/game`、`internal/rules` 全部通過；
`internal/bootstrap` 可編譯。
