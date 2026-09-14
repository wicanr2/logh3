# 015 — 艦隊「移動／戰鬥／索敵」行為收據

狀態：**CONFORMED**  
原版：`GIN3.COM` SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`  
oracle：dosgolem-logh3 `e0b6a03`

三項命令的客觀收據鏈：

- 「移動」：`s3cmdrange`／`s3cmdface` 固定兩段互動與抵達朝向；短程移動
  的同狀態記憶體比較見 [`011`](011-fleet-short-move-strafe.md)。
- 「戰鬥」：`s3fltseq` 把 `+0x48` 行動寫成 1，十一份執行期傾印逐步固定
  回轉、行程額度、受阻落點與抵達朝向；`TestFleetBudgetMatchesReceipt`
  逐值驗證。
- 「索敵」：新收據 `s3cmd-scout-behavior` 從正常玩家路徑建立偵察隊並
  選定目的地，完整 1 MB 傾印及欄位見
  [`docs/re/305`](../re/305-scout-order-runtime-receipt.md)。

共同機器閘門 `TestFleetCommandBehaviorsMatchOriginalReceipts` 逐一執行移動、
戰鬥與索敵的玩家路徑測試；缺原版索敵收據時依專案契約 skip，不製造代用品。

