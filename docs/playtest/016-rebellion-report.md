# 016 — 叛亂戰略報告的正常玩家路徑

日期：2026-09-14  
狀態：**CONFORMED**（SPEC 046 的玩家可見範圍）  
原版：`GIN3.COM` SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`  
oracle：dosgolem-logh3 `e0b6a03`

## 原版收據

`rbx292` 從新遊戲正常路徑按一次「實行」，只注入
`lin:60426@211000000=01`，把帝國情報專任參謀的營運改成 1，使
星域 1 的叛亂不依賴運氣。第一則報告停在：

- 星域：亞爾迪巴朗（星域 1，事件前擁有者帝國）；
- 訊息：`發生叛亂。`；
- 頭像：`FACE.AMG` 第 13 張，左上 72×93；
- `screen.bin` SHA-256：
  `8a2a5b749cf42be754f706e4e4594878a668b9a88edafa8dcaf58dc4c4e05d73`。

## remake 對拍

`TestRebellionReportPlayerPathMatchesOriginal` 走主選單 → 防拷 →
月初報告 → 系統指令 → 實行，確認先進 `StateRebellionReport`，並比較：

- 文字區 `(334,178)–(509,219)`：0 像素差；
- 頭像 `(544,180)` 起 72×93：0 像素差。

`TestRebellionReportsQueueAndResume` 另驗兩格叛亂必須按兩次，左鍵與右鍵
各收一則，callback 只在佇列清空後執行一次。完整
`internal/screen`、`internal/game`、`internal/rules` 回歸均通過。

## 限制

原版判到第一格叛亂便暫停，後續星域尚未擲亂數；remake 目前先算完整圈，
再逐則顯示。玩家可見畫面與後續入口已對上，但停在報告時的完整記憶體與
RNG 位置不同，不能把本收據升格為全機 `same-state`。
